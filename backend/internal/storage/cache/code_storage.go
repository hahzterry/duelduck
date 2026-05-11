package cache

import (
	"context"
	"dd-prediction-api/config"
	"dd-prediction-api/pkg/apperrors"
	"dd-prediction-api/pkg/mtype"

	"github.com/redis/go-redis/v9"
)

type CodeStorage struct {
	client *redis.Client
	cfg    *config.AuthConfig
}

func NewCodeCacheStorage(c *config.Config, client *redis.Client) *CodeStorage {
	return &CodeStorage{
		client: client,
		cfg:    &c.Auth,
	}
}

func (s *CodeStorage) Save(ctx context.Context, email mtype.Email, code string) error {
	err := s.client.Set(ctx, email.String(), code, s.cfg.CodeTTL).Err()
	if err != nil {
		return apperrors.Unauthorized("failed to save user's token", err)
	}

	return nil
}

func (s *CodeStorage) GetCodeByEmail(ctx context.Context, email mtype.Email) (string, error) {
	result, err := s.client.Get(ctx, email.String()).Result()
	if err != nil {
		return "", apperrors.Unauthorized("failed to find code by user's email", err)
	}

	if result == "" {
		return "", apperrors.Unauthorized("token not found")
	}

	return result, nil
}

func (s *CodeStorage) Delete(ctx context.Context, key string) error {
	err := s.client.Del(ctx, key).Err()
	if err != nil {
		return apperrors.Internal("failed to delete code", err)
	}

	return nil
}

func (s *CodeStorage) SaveWithKey(ctx context.Context, key string, code string) error {
	err := s.client.Set(ctx, key, code, s.cfg.CodeTTL).Err()
	if err != nil {
		return apperrors.Unauthorized("failed to save code", err)
	}

	return nil
}

func (s *CodeStorage) GetCodeByKey(ctx context.Context, key string) (string, error) {
	result, err := s.client.Get(ctx, key).Result()
	if err != nil {
		return "", apperrors.Unauthorized("failed to find code by key", err)
	}

	if result == "" {
		return "", apperrors.Unauthorized("code not found")
	}

	return result, nil
}
