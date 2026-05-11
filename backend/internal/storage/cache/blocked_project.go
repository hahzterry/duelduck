package cache

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type IBlockedProjectCache interface {
	Block(ctx context.Context, projectID uuid.UUID) error
	Unblock(ctx context.Context, projectID uuid.UUID) error
	IsBlocked(ctx context.Context, projectID uuid.UUID) (bool, error)
}

type BlockedProjectCache struct {
	client *redis.Client
}

func NewBlockedProjectCache(client *redis.Client) *BlockedProjectCache {
	return &BlockedProjectCache{client: client}
}

func (c *BlockedProjectCache) Block(ctx context.Context, projectID uuid.UUID) error {
	return c.client.Set(ctx, blockedProjectKey(projectID), 1, 0).Err()
}

func (c *BlockedProjectCache) Unblock(ctx context.Context, projectID uuid.UUID) error {
	return c.client.Del(ctx, blockedProjectKey(projectID)).Err()
}

func (c *BlockedProjectCache) IsBlocked(ctx context.Context, projectID uuid.UUID) (bool, error) {
	n, err := c.client.Exists(ctx, blockedProjectKey(projectID)).Result()
	return n > 0, err
}

func blockedProjectKey(id uuid.UUID) string {
	return fmt.Sprintf("project:blocked:%s", id.String())
}
