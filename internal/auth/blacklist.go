package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type TokenBlacklist interface {
	Revoke(ctx context.Context, tokenID string, remainingTTL time.Duration) error
	IsRevoked(ctx context.Context, tokenID string) (bool, error)
}

type redisBlacklist struct {
	client *redis.Client
}

func NewTokenBlacklist(client *redis.Client) TokenBlacklist {
	return &redisBlacklist{client: client}
}

func (b *redisBlacklist) Revoke(ctx context.Context, tokenID string, remainingTTL time.Duration) error {
	if b.client == nil || tokenID == "" {
		return nil
	}
	key := fmt.Sprintf("blacklist:token:%s", tokenID)
	return b.client.Set(ctx, key, "revoked", remainingTTL).Err()
}

func (b *redisBlacklist) IsRevoked(ctx context.Context, tokenID string) (bool, error) {
	if b.client == nil || tokenID == "" {
		return false, nil
	}
	key := fmt.Sprintf("blacklist:token:%s", tokenID)
	exists, err := b.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}
