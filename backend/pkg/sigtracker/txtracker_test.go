package sigtracker

import (
	"context"
	"errors"
	"testing"
	"time"

	solanalib "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

type fakeRPCClient struct {
	statusBySignature map[solanalib.Signature]*rpc.SignatureStatusesResult
	err               error
	latency           time.Duration
}

func newFakeRPCClient() *fakeRPCClient {
	return &fakeRPCClient{statusBySignature: make(map[solanalib.Signature]*rpc.SignatureStatusesResult)}
}
func (f *fakeRPCClient) setStatus(signature solanalib.Signature, status *rpc.SignatureStatusesResult) {
	f.statusBySignature[signature] = status
}
func (f *fakeRPCClient) GetSignatureStatuses(ctx context.Context, _ bool, signatures ...solanalib.Signature) (*rpc.GetSignatureStatusesResult, error) {
	if f.latency > 0 {
		select {
		case <-time.After(f.latency):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if f.err != nil {
		return nil, f.err
	}
	out := &rpc.GetSignatureStatusesResult{Value: make([]*rpc.SignatureStatusesResult, len(signatures))}
	for i, sig := range signatures {
		out.Value[i] = f.statusBySignature[sig]
	}
	return out, nil
}

type fakeSubscription struct {
	resultToDeliver *ws.SignatureResult
	errToDeliver    error
	deliverDelay    time.Duration
	unsubscribed    bool
}

func (s *fakeSubscription) Recv(ctx context.Context) (*ws.SignatureResult, error) {
	if s.deliverDelay > 0 {
		select {
		case <-time.After(s.deliverDelay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if s.errToDeliver != nil {
		return nil, s.errToDeliver
	}
	return s.resultToDeliver, nil
}
func (s *fakeSubscription) Unsubscribe() { s.unsubscribed = true }

type fakeWSClient struct {
	subscriptions map[solanalib.Signature]*fakeSubscription
	subscribeErr  error
}

func newFakeWSClient() *fakeWSClient {
	return &fakeWSClient{subscriptions: make(map[solanalib.Signature]*fakeSubscription)}
}
func (c *fakeWSClient) SignatureSubscribe(signature solanalib.Signature, _ rpc.CommitmentType) (signatureSubscriptionInterface, error) {
	if c.subscribeErr != nil {
		return nil, c.subscribeErr
	}
	sub := c.subscriptions[signature]
	if sub == nil {
		sub = &fakeSubscription{}
		c.subscriptions[signature] = sub
	}
	return sub, nil
}
func (c *fakeWSClient) Close() {}

func makeSignature(s string) solanalib.Signature {
	var sig solanalib.Signature
	copy(sig[:], []byte(s))
	return sig
}

func statusConfirmed() *rpc.SignatureStatusesResult {
	return &rpc.SignatureStatusesResult{
		ConfirmationStatus: rpc.ConfirmationStatusConfirmed,
		Err:                nil,
	}
}
func statusFailed() *rpc.SignatureStatusesResult {
	return &rpc.SignatureStatusesResult{
		Err: map[string]any{"InstructionError": []any{0, map[string]any{"Custom": 1}}},
	}
}

func newTestTracker(t *testing.T, rpcFake *fakeRPCClient, wsFake *fakeWSClient) *TxTracker {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	tracker := &TxTracker{
		rpcClient:           rpcFake,
		wsURL:               "wss://fake",
		wsClient:            wsFake,
		entries:             make(map[solanalib.Signature]*signatureEntry),
		pendingSignaturesCh: make(chan solanalib.Signature, 1024),
		minCommitment:       rpc.CommitmentConfirmed,
		maxWSRetries:        1,
		ctx:                 ctx,
		cancel:              cancel,
		// Must be large enough that per-call timeouts can fire before WS gives up.
		wsRecvTimeout: 500 * time.Millisecond,
		rpcTimeout:    200 * time.Millisecond,
		minTimeout:    5 * time.Millisecond,
		maxTimeout:    50 * time.Millisecond,
	}

	go tracker.runSubscriptionManager()

	t.Cleanup(func() { _ = tracker.Close() })
	return tracker
}

func TestSubscribe_ReturnsTrue_WhenWSNotifiesAndRPCConfirms(t *testing.T) {
	rpcFake := newFakeRPCClient()
	wsFake := newFakeWSClient()
	tracker := newTestTracker(t, rpcFake, wsFake)

	signature := makeSignature("sig-happy-path")
	rpcFake.setStatus(signature, statusConfirmed())

	wsFake.subscriptions[signature] = &fakeSubscription{
		resultToDeliver: &ws.SignatureResult{},
		deliverDelay:    10 * time.Millisecond,
	}

	confirmed, err := tracker.SubscribeForSignatureStatus(signature, 2*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !confirmed {
		t.Fatalf("expected confirmed=true, got false")
	}
}

func TestSubscribe_TimesOut_WhenNoWebsocketEvent(t *testing.T) {
	rpcFake := newFakeRPCClient()
	wsFake := newFakeWSClient()
	tracker := newTestTracker(t, rpcFake, wsFake)

	signature := makeSignature("sig-timeout")

	wsFake.subscriptions[signature] = &fakeSubscription{
		deliverDelay: 2 * time.Second,
	}

	confirmed, err := tracker.SubscribeForSignatureStatus(signature, 80*time.Millisecond)
	if err == nil {
		t.Fatalf("expected timeout error, got nil")
	}
	if confirmed {
		t.Fatalf("expected confirmed=false on timeout")
	}
}

func TestSubscribe_FallsBackToRPC_OnBrokenPipe(t *testing.T) {
	rpcFake := newFakeRPCClient()
	wsFake := newFakeWSClient()
	tracker := newTestTracker(t, rpcFake, wsFake)

	signature := makeSignature("sig-broken-pipe")
	rpcFake.setStatus(signature, statusConfirmed())

	wsFake.subscriptions[signature] = &fakeSubscription{
		errToDeliver: errors.New("broken pipe"),
		deliverDelay: 5 * time.Millisecond,
	}

	confirmed, err := tracker.SubscribeForSignatureStatus(signature, 2*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !confirmed {
		t.Fatalf("expected confirmed=true after RPC fallback")
	}
}

func TestSubscribe_NotifiesAllWaiters_ForSameSignature(t *testing.T) {
	rpcFake := newFakeRPCClient()
	wsFake := newFakeWSClient()
	tracker := newTestTracker(t, rpcFake, wsFake)

	signature := makeSignature("sig-multi-waiters")
	rpcFake.setStatus(signature, statusFailed())

	wsFake.subscriptions[signature] = &fakeSubscription{
		errToDeliver: errors.New("transaction failed"),
		deliverDelay: 5 * time.Millisecond,
	}

	type result struct {
		ok  bool
		err error
	}
	done1 := make(chan result, 1)
	done2 := make(chan result, 1)

	go func() {
		ok, err := tracker.SubscribeForSignatureStatus(signature, 2*time.Second)
		done1 <- result{ok, err}
	}()
	go func() {
		ok, err := tracker.SubscribeForSignatureStatus(signature, 2*time.Second)
		done2 <- result{ok, err}
	}()

	r1 := <-done1
	r2 := <-done2

	if r1.err != nil || r2.err != nil {
		t.Fatalf("unexpected errors: %v / %v", r1.err, r2.err)
	}
	if r1.ok || r2.ok {
		t.Fatalf("both should be false due to on-chain error")
	}
}

func TestClose_CancelsPendingWaiters(t *testing.T) {
	rpcFake := newFakeRPCClient()
	wsFake := newFakeWSClient()
	tracker := newTestTracker(t, rpcFake, wsFake)

	sig := makeSignature("sig-close-cancels")
	done := make(chan struct{})

	go func() {
		_, _ = tracker.SubscribeForSignatureStatus(sig, 2*time.Second)
		close(done)
	}()

	time.Sleep(20 * time.Millisecond)
	_ = tracker.Close()

	select {
	case <-done:
	case <-time.After(300 * time.Millisecond):
		t.Fatal("expected waiter to be released on Close()")
	}
}
