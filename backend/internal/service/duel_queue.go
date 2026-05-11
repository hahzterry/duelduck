package service

import (
	"context"
	"dd-prediction-api/config"
	"dd-prediction-api/internal/model"
	"dd-prediction-api/pkg/apperrors"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type CommandResult struct {
	Value any
	Err   error
}

type RequestKey struct {
	DuelID uuid.UUID
	UserID uuid.UUID
	Method string
}

type DuelCommand struct {
	DuelID   uuid.UUID
	UserID   uuid.UUID
	Name     string
	Ctx      context.Context
	Func     func(context.Context) CommandResult
	ResultCh chan CommandResult
}

type duelWorker struct {
	duelID    uuid.UUID
	ch        chan *DuelCommand
	busy      bool
	lastRunAt time.Time
}

type DuelQueue struct {
	DuelService *DuelService

	workers         map[uuid.UUID]*duelWorker
	pendingRequests map[RequestKey]struct{}
	queueSize       uint32
	mu              sync.Mutex
}

func NewDuelQueue(
	duelService *DuelService,
	c *config.Config,
) *DuelQueue {
	return &DuelQueue{
		DuelService:     duelService,
		workers:         make(map[uuid.UUID]*duelWorker),
		pendingRequests: make(map[RequestKey]struct{}),
		queueSize:       c.App.DuelQueueSize,
	}
}

func (q *DuelQueue) Enqueue(cmd *DuelCommand) error {
	reqKey := RequestKey{
		DuelID: cmd.DuelID,
		UserID: cmd.UserID,
		Method: cmd.Name,
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	if _, exists := q.pendingRequests[reqKey]; exists {
		return apperrors.Forbidden("request already exists")
	}

	q.pendingRequests[reqKey] = struct{}{}

	w, ok := q.workers[cmd.DuelID]
	if !ok {
		w = &duelWorker{
			duelID:    cmd.DuelID,
			ch:        make(chan *DuelCommand, q.queueSize),
			lastRunAt: time.Now(),
		}
		q.workers[cmd.DuelID] = w
		go q.runWorker(w)
	}

	select {
	case w.ch <- cmd:
		return nil
	default:
		delete(q.pendingRequests, reqKey)

		return apperrors.TooManyRequests("queue is full")
	}
}

func (q *DuelQueue) CleanupWorkers(ttl time.Duration) int {
	now := time.Now()
	removed := 0

	q.mu.Lock()
	defer q.mu.Unlock()

	for duelID, w := range q.workers {
		if now.Sub(w.lastRunAt) < ttl {
			continue
		}

		if len(w.ch) > 0 || w.busy {
			continue
		}

		delete(q.workers, duelID)
		close(w.ch)
		removed++
	}

	return removed
}

func enqueueAndWait[T any](
	ctx context.Context,
	q *DuelQueue,
	cmd *DuelCommand,
) (*T, error) {
	if err := q.Enqueue(cmd); err != nil {
		return nil, err
	}

	select {
	case res := <-cmd.ResultCh:
		if res.Err != nil {
			return nil, res.Err
		}

		resp, ok := res.Value.(*T)
		if !ok {
			return nil, apperrors.Internal("failed to get response")
		}

		return resp, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (q *DuelQueue) runWorker(w *duelWorker) {
	for cmd := range w.ch {
		func() {
			reqKey := RequestKey{
				DuelID: cmd.DuelID,
				UserID: cmd.UserID,
				Method: cmd.Name,
			}

			q.mu.Lock()
			w.busy = true
			q.mu.Unlock()

			defer func() {
				q.mu.Lock()
				delete(q.pendingRequests, reqKey)
				w.busy = false
				w.lastRunAt = time.Now()
				q.mu.Unlock()

				if r := recover(); r != nil {
					zap.L().Error("duel queue panic occured", zap.Any("panic", r))
					select {
					case cmd.ResultCh <- CommandResult{
						Err: apperrors.Internal("something went wrong"),
					}:
					default:
					}
				}
			}()

			res := cmd.Func(cmd.Ctx)

			select {
			case cmd.ResultCh <- res:
			case <-cmd.Ctx.Done():
			}
		}()
	}
}

func (q *DuelQueue) JoinCryptoDuel(
	ctx context.Context,
	userID uuid.UUID,
	req *model.JoinCryptoDuelReq,
) (*model.JoinCryptoDuelResp, error) {
	resultCh := make(chan CommandResult, 1)

	cmd := &DuelCommand{
		DuelID: req.JoinDuelReq.DuelID,
		UserID: userID,
		Name:   "join-crypto-external",
		Ctx:    ctx,
		Func: func(ctx context.Context) CommandResult {
			resp, err := q.DuelService.JoinCryptoDuel(ctx, userID, req)
			return CommandResult{
				Value: resp,
				Err:   err,
			}
		},
		ResultCh: resultCh,
	}

	return enqueueAndWait[model.JoinCryptoDuelResp](ctx, q, cmd)
}

func (q *DuelQueue) SignJoinCryptoDuelTransaction(
	ctx context.Context,
	userID uuid.UUID,
	req *model.JoinDuelReq,
) (*model.TxHashResp, error) {
	resultCh := make(chan CommandResult, 1)

	cmd := &DuelCommand{
		DuelID: req.DuelID,
		UserID: userID,
		Name:   "sign-join-crypto-external",
		Ctx:    ctx,
		Func: func(ctx context.Context) CommandResult {
			resp, err := q.DuelService.SignJoinCryptoDuelTransaction(ctx, userID, req)
			return CommandResult{
				Value: resp,
				Err:   err,
			}
		},
		ResultCh: resultCh,
	}

	return enqueueAndWait[model.TxHashResp](ctx, q, cmd)
}

func (q *DuelQueue) ResolveCryptoDuelByOwner(
	ctx context.Context,
	userID uuid.UUID,
	req *model.DuelResolveReq,
) (*model.ResolveCryptoDuelResp, error) {
	resultCh := make(chan CommandResult, 1)

	cmd := &DuelCommand{
		DuelID: req.DuelID,
		UserID: userID,
		Name:   "resolve-crypto-owner",
		Ctx:    ctx,
		Func: func(ctx context.Context) CommandResult {
			resp, err := q.DuelService.ResolveCryptoDuelByOwner(ctx, userID, req)
			return CommandResult{
				Value: resp,
				Err:   err,
			}
		},
		ResultCh: resultCh,
	}

	return enqueueAndWait[model.ResolveCryptoDuelResp](ctx, q, cmd)
}

func (q *DuelQueue) ApproveCryptoDuel(
	ctx context.Context,
	userID uuid.UUID,
	req *model.DuelApproveReq,
) (*model.JoinCryptoDuelResp, error) {
	resultCh := make(chan CommandResult, 1)

	cmd := &DuelCommand{
		DuelID:   req.DuelID,
		UserID:   userID,
		Name:     "approve-crypto-admin",
		Ctx:      ctx,
		ResultCh: resultCh,
		Func: func(ctx context.Context) CommandResult {
			resp, err := q.DuelService.ApproveCryptoDuel(ctx, userID, req)
			return CommandResult{
				Value: resp,
				Err:   err,
			}
		},
	}

	return enqueueAndWait[model.JoinCryptoDuelResp](ctx, q, cmd)
}

func (q *DuelQueue) ResolveCryptoDuel(
	ctx context.Context,
	userID uuid.UUID,
	req *model.DuelResolveReq,
) (*model.ResolveCryptoDuelResp, error) {
	resultCh := make(chan CommandResult, 1)

	cmd := &DuelCommand{
		DuelID: req.DuelID,
		UserID: userID,
		Name:   "resolve-crypto-admin",
		Ctx:    ctx,
		Func: func(ctx context.Context) CommandResult {
			resp, err := q.DuelService.ResolveCryptoDuelByAdmin(ctx, userID, req)
			return CommandResult{
				Value: resp,
				Err:   err,
			}
		},
		ResultCh: resultCh,
	}

	return enqueueAndWait[model.ResolveCryptoDuelResp](ctx, q, cmd)
}

func (q *DuelQueue) CancelCryptoDuel(
	ctx context.Context,
	userID uuid.UUID,
	req *model.DuelCancelReq,
) (*model.CancelCryptoDuelResp, error) {
	resultCh := make(chan CommandResult, 1)

	cmd := &DuelCommand{
		DuelID: req.DuelID,
		UserID: userID,
		Name:   "cancel-crypto-admin",
		Ctx:    ctx,
		Func: func(ctx context.Context) CommandResult {
			resp, err := q.DuelService.CancelCryptoDuelByAdmin(ctx, userID, req)
			return CommandResult{
				Value: resp,
				Err:   err,
			}
		},
		ResultCh: resultCh,
	}

	return enqueueAndWait[model.CancelCryptoDuelResp](ctx, q, cmd)
}
