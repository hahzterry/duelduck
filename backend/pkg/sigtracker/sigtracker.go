package sigtracker

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	solanalib "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	solana "github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"go.uber.org/zap"
)

// TxTracker tracks Solana transaction confirmations using both WS + RPC fallback
type TxTracker struct {
	rpcClient rpcClientInterface
	wsURL     string

	wsClient wsClientInterface
	wsMutex  sync.Mutex

	entries      map[solanalib.Signature]*signatureEntry
	entriesMutex sync.Mutex

	pendingSignaturesCh chan solanalib.Signature

	minCommitment rpc.CommitmentType
	maxWSRetries  int

	maxRPCRetries int
	minTimeout    time.Duration
	maxTimeout    time.Duration
	rpcTimeout    time.Duration
	wsRecvTimeout time.Duration

	ctx    context.Context
	cancel context.CancelFunc
}

type signatureEntry struct {
	waiters []chan bool // all goroutines waiting for this signature
	started bool        // whether WS subscription already started
}

// Initialize a new transaction tracker instance
func NewTransactionTracker(rpcClient *solana.Client, wsURL string) *TxTracker {
	ctx, cancel := context.WithCancel(context.Background())
	return &TxTracker{
		rpcClient:           rpcClientAdapter{rpcClient},
		wsURL:               wsURL,
		entries:             make(map[solanalib.Signature]*signatureEntry),
		pendingSignaturesCh: make(chan solanalib.Signature, 1024),
		minCommitment:       rpc.CommitmentConfirmed,
		maxWSRetries:        5,
		maxRPCRetries:       3,
		minTimeout:          300 * time.Millisecond,
		maxTimeout:          1200 * time.Millisecond,
		wsRecvTimeout:       30 * time.Second,
		rpcTimeout:          10 * time.Second,
		ctx:                 ctx,
		cancel:              cancel,
	}
}

// Start async manager that processes pending signatures
func (t *TxTracker) Start() {
	zap.L().Info("TxTracker started", zap.String("wsURL", t.wsURL))
	go t.runSubscriptionManager()
}

// Gracefully stop all subscriptions and release resources
func (t *TxTracker) Close() error {
	t.cancel()

	t.wsMutex.Lock()
	if t.wsClient != nil {
		t.wsClient.Close()
		t.wsClient = nil
	}
	t.wsMutex.Unlock()

	// Notify all pending waiters that tracker is shutting down
	t.entriesMutex.Lock()
	for signature, entry := range t.entries {
		for _, waiterCh := range entry.waiters {
			select {
			case waiterCh <- false:
			default:
			}
		}
		delete(t.entries, signature)
	}
	t.entriesMutex.Unlock()

	zap.L().Info("TxTracker closed")
	return nil
}

// Subscribe for a specific transaction signature confirmation
// Returns true when confirmed, false otherwise (timeout or failed)
func (t *TxTracker) SubscribeForSignatureStatus(signature solanalib.Signature, timeout time.Duration) (bool, error) {
	waiterCh := make(chan bool, 1) // buffer prevents goroutine leak

	shouldEnqueue := false

	// Register a waiter; if it's first for this signature → enqueue
	t.entriesMutex.Lock()
	entry, exists := t.entries[signature]
	if !exists {
		entry = &signatureEntry{waiters: make([]chan bool, 0, 1)}
		t.entries[signature] = entry
	}
	entry.waiters = append(entry.waiters, waiterCh)
	if !entry.started {
		entry.started = true
		shouldEnqueue = true
	}
	t.entriesMutex.Unlock()

	if shouldEnqueue {
		select {
		case t.pendingSignaturesCh <- signature:
		default:
			go t.subscribeOne(signature) // fallback if channel is full
		}
	}

	zap.L().Debug("waiter registered",
		zap.String("signature", signature.String()),
		zap.Duration("timeout", timeout),
	)

	// Wait until either confirmation or timeout
	ctx, cancel := context.WithTimeout(t.ctx, timeout)
	defer cancel()

	select {
	case confirmed := <-waiterCh:
		return confirmed, nil
	case <-ctx.Done():
		// Remove this waiter if timed out
		t.entriesMutex.Lock()
		if current := t.entries[signature]; current != nil {
			remaining := current.waiters[:0]
			for _, ch := range current.waiters {
				if ch != waiterCh {
					remaining = append(remaining, ch)
				}
			}
			current.waiters = remaining
		}
		t.entriesMutex.Unlock()
		return false, fmt.Errorf("signature %s: %w", signature, rpc.ErrNotConfirmed)
	}
}

// Background worker — starts WS subscriptions for pending signatures
func (t *TxTracker) runSubscriptionManager() {
	for {
		select {
		case <-t.ctx.Done():
			zap.L().Info("TxTracker subscription manager stopped")
			return

		case signature := <-t.pendingSignaturesCh:
			// Ensure WS connection exists
			if err := t.ensureWS(); err != nil {
				zap.L().Error("ensureWS failed", zap.Error(err))
				// Retry after small delay
				time.AfterFunc(300*time.Millisecond, func() {
					select {
					case t.pendingSignaturesCh <- signature:
					default:
						go t.subscribeOne(signature)
					}
				})
				continue
			}
			go t.subscribeOne(signature)
		}
	}
}

// Subscribe to WS updates for a single signature
// Fallback to RPC if WS connection breaks
func (t *TxTracker) subscribeOne(signature solanalib.Signature) {
	sub, err := t.wsSubscribe(signature)
	if err != nil {
		t.finish(signature, false)
		return
	}
	defer sub.Unsubscribe()

	_, err = t.awaitWS(t.ctx, sub)
	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			zap.L().Warn("WS recv timed out; falling back to RPC",
				zap.String("signature", signature.String()),
				zap.Duration("ws_recv_timeout", t.wsRecvTimeout),
				zap.Duration("rpc_fallback_window", t.rpcTimeout),
			)
		case isBrokenConnectionError(err):
			zap.L().Warn("WS connection issue; falling back to RPC",
				zap.String("signature", signature.String()),
				zap.Error(err),
				zap.Duration("rpc_fallback_window", t.rpcTimeout),
			)
		default:
			zap.L().Info("WS recv failed; falling back to RPC",
				zap.String("signature", signature.String()),
				zap.Error(err),
				zap.Duration("rpc_fallback_window", t.rpcTimeout),
			)
		}

		if ok, err := t.waitForRPCConfirmation(signature, t.rpcTimeout); ok && err == nil {
			t.finish(signature, true)
			return
		}
		t.finish(signature, false)
		return
	}

	const rpcGraceAfterWS = 5 * time.Second
	if t.rpcConfirmWithGrace(signature, rpcGraceAfterWS) {
		t.finish(signature, true)
		return
	}
	zap.L().Info("WS confirmed but RPC not yet updated",
		zap.String("signature", signature.String()),
		zap.Duration("grace_window", rpcGraceAfterWS),
	)
	t.finish(signature, true)
}

func (t *TxTracker) wsSubscribe(signature solanalib.Signature) (signatureSubscriptionInterface, error) {
	var sub signatureSubscriptionInterface
	var err error

	for attempt := 0; attempt <= t.maxWSRetries; attempt++ {
		if e := t.ensureWS(); e != nil {
			zap.L().Error("ensureWS failed in wsSubscribe",
				zap.String("signature", signature.String()),
				zap.Int("attempt", attempt),
				zap.Error(e),
			)
			time.Sleep(exponentialBackoff(attempt, t.minTimeout, t.maxTimeout))
			continue
		}

		t.wsMutex.Lock()
		sub, err = t.wsClient.SignatureSubscribe(signature, t.minCommitment)
		t.wsMutex.Unlock()
		if err == nil {
			return sub, nil
		}

		zap.L().Error("SignatureSubscribe failed",
			zap.String("signature", signature.String()),
			zap.Int("attempt", attempt),
			zap.Error(err),
		)
		if isBrokenConnectionError(err) {
			_ = t.restartWS()
		}
		time.Sleep(exponentialBackoff(attempt, t.minTimeout, t.maxTimeout))
	}
	return nil, err
}

func (t *TxTracker) awaitWS(parent context.Context, sub signatureSubscriptionInterface) (*ws.SignatureResult, error) {
	recvCtx, cancel := context.WithTimeout(parent, t.wsRecvTimeout)
	defer cancel()
	return sub.Recv(recvCtx)
}

func (t *TxTracker) rpcConfirmWithGrace(sig solanalib.Signature, window time.Duration) bool {
	ok, err := t.waitForRPCConfirmation(sig, window)
	if err != nil {
		zap.L().Warn("RPC confirmation check during grace failed",
			zap.String("signature", sig.String()),
			zap.Duration("window", window),
			zap.Error(err),
		)
	}
	return ok
}

// Notify all waiters and clear entry for this signature
func (t *TxTracker) finish(signature solanalib.Signature, confirmed bool) {
	var waiterChans []chan bool

	t.entriesMutex.Lock()
	if entry := t.entries[signature]; entry != nil {
		waiterChans = entry.waiters
		delete(t.entries, signature)
	}
	t.entriesMutex.Unlock()

	for _, ch := range waiterChans {
		select {
		case ch <- confirmed:
		default:
		}
	}
	zap.L().Info("signature resolved",
		zap.String("signature", signature.String()),
		zap.Bool("ok", confirmed),
	)
}

func (t *TxTracker) checkByRPCOnce(signature solanalib.Signature) (bool, error) {
	ctx, cancel := context.WithTimeout(t.ctx, t.rpcTimeout)
	defer cancel()

	statuses, err := t.rpcClient.GetSignatureStatuses(ctx, true, signature)
	if err != nil {
		return false, err
	}

	if len(statuses.Value) == 0 || statuses.Value[0] == nil {
		return false, rpc.ErrNotConfirmed
	}
	st := statuses.Value[0]

	if st.Err != nil {
		return false, nil
	}

	switch st.ConfirmationStatus {
	case rpc.ConfirmationStatusFinalized, rpc.ConfirmationStatusConfirmed:
		return true, nil
	default:
		return false, rpc.ErrNotConfirmed
	}
}

func (t *TxTracker) waitForRPCConfirmation(signature solanalib.Signature, timeout time.Duration) (bool, error) {
	ctx, cancel := context.WithTimeout(t.ctx, timeout)
	defer cancel()

	for attempt := 0; ; attempt++ {
		select {
		case <-ctx.Done():
			return false, rpc.ErrNotConfirmed
		default:
			ok, err := t.checkByRPCOnce(signature)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
			// Confirmed failed
			return false, nil
		}
	}
}

// Unified connect/reconnect logic for WS client
func (t *TxTracker) connectWS(force bool) error {
	t.wsMutex.Lock()
	defer t.wsMutex.Unlock()

	if t.wsClient != nil && !force {
		return nil
	}

	if t.wsClient != nil {
		t.wsClient.Close()
		t.wsClient = nil
	}

	wsConn, err := ws.Connect(context.Background(), t.wsURL)
	if err != nil {
		return err
	}

	t.wsClient = wsClientAdapter{wsConn}
	if force {
		zap.L().Info("ws reconnected", zap.String("url", t.wsURL))
	} else {
		zap.L().Info("ws connected", zap.String("url", t.wsURL))
	}
	return nil
}

func (t *TxTracker) ensureWS() error  { return t.connectWS(false) }
func (t *TxTracker) restartWS() error { return t.connectWS(true) }

// Detects if WS/RPC connection was broken
func isBrokenConnectionError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	message := err.Error()
	return strings.Contains(message, "broken pipe") ||
		strings.Contains(message, "unexpected EOF") ||
		strings.Contains(message, "use of closed network connection") ||
		strings.Contains(message, "websocket: close") ||
		strings.Contains(message, "reset by peer") ||
		strings.Contains(message, "aborted") ||
		strings.Contains(message, "i/o timeout")
}

// Simple exponential backoff for retry logic
func exponentialBackoff(attempt int, baseDelay, maxDelay time.Duration) time.Duration {
	delayMS := float64(baseDelay.Milliseconds()) * math.Pow(2, float64(attempt))
	delay := time.Duration(delayMS) * time.Millisecond
	if delay > maxDelay {
		return maxDelay
	}
	return delay
}
