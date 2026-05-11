package click

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"dd-prediction-api/internal/model"
)

type Config struct {
	BufferSize    int
	FlushInterval time.Duration
}

type ActivityBuffer struct {
	repo        *UserActivityRepository
	ch          chan model.UserActivity
	flushSize   int
	flushTicker *time.Ticker
	ctx         context.Context
	cancel      context.CancelFunc
	flushMu     sync.Mutex
}

var (
	once   sync.Once
	buffer *ActivityBuffer
)

func InitActivityBuffer(repo *UserActivityRepository, cfg Config) {
	once.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		buffer = &ActivityBuffer{
			repo:        repo,
			ch:          make(chan model.UserActivity, cfg.BufferSize),
			flushSize:   cfg.BufferSize,
			flushTicker: time.NewTicker(cfg.FlushInterval),
			ctx:         ctx,
			cancel:      cancel,
		}
		go buffer.run()
	})
}

func GetActivityBuffer() *ActivityBuffer {
	return buffer
}

func (b *ActivityBuffer) Stop() {
	b.cancel()
	b.flushTicker.Stop()
}

func (b *ActivityBuffer) Add(activity model.UserActivity) {
	select {
	case b.ch <- activity:
	default:
		go b.flush()
		b.ch <- activity
	}
}

func (b *ActivityBuffer) run() {
	for {
		select {
		case <-b.ctx.Done():
			b.flush()
			return
		case <-b.flushTicker.C:
			b.flush()
		}
	}
}

func (b *ActivityBuffer) flush() {
	log.Printf("Flushing activity buffer")

	b.flushMu.Lock()
	defer b.flushMu.Unlock()

	rawActivities := make([]model.UserActivity, 0, b.flushSize)

loop:
	for len(rawActivities) < b.flushSize {
		select {
		case act := <-b.ch:
			rawActivities = append(rawActivities, act)
		default:
			break loop
		}
	}

	if len(rawActivities) == 0 {
		return
	}

	aggMap := make(map[string]model.UserActivity)

	for _, act := range rawActivities {
		key := fmt.Sprintf("%s|%s|%s", act.UserID, act.ActivityType, act.Metadata)

		if act.ActivityType == model.ActivityPageView ||
			act.ActivityType == model.ActivityTokenRefresh {

			if existing, ok := aggMap[key]; ok {
				existing.ActivityCount += act.ActivityCount
				existing.CreatedAt = act.CreatedAt
				aggMap[key] = existing
				continue
			}
		} else {
			act.ActivityCount = 1
		}

		if act.ActivityCount == 0 {
			act.ActivityCount = 1
		}
		aggMap[key] = act
	}

	activities := make([]model.UserActivity, 0, len(aggMap))
	for _, act := range aggMap {
		activities = append(activities, act)
	}

	if err := b.repo.BulkInsert(b.ctx, activities); err != nil {
		log.Printf("failed to bulk insert user activities: %v", err)
	}
}
