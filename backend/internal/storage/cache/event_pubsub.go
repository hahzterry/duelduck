package cache

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type EventPubSub struct {
	client *redis.Client
}

func NewEventPubSub(client *redis.Client) *EventPubSub {
	return &EventPubSub{client: client}
}

func (e *EventPubSub) Subscribe(ctx context.Context, userID uuid.UUID) *redis.PubSub {
	channel := fmt.Sprintf("events:%s", userID.String())
	return e.client.Subscribe(ctx, channel)
}

func (e *EventPubSub) Publish(ctx context.Context, userID uuid.UUID, payload string) error {
	channel := fmt.Sprintf("events:%s", userID.String())
	return e.client.Publish(ctx, channel, payload).Err()
}
