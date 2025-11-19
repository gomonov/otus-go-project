package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type TokenBucket struct {
	client *redis.Client
	key    string
	limit  int
	window int
}

func NewTokenBucket(client *redis.Client, key string, limit, window int) *TokenBucket {
	return &TokenBucket{
		client: client,
		key:    key,
		limit:  limit,
		window: window,
	}
}

func (tb *TokenBucket) Allow(ctx context.Context) (bool, error) {
	allowance, timestamp, err := tb.loadAllowance(ctx)
	if err != nil {
		return false, err
	}

	current := time.Now().Unix()

	timePassed := current - timestamp
	allowance += float64(timePassed) * float64(tb.limit) / float64(tb.window)
	if allowance > float64(tb.limit) {
		allowance = float64(tb.limit)
	}

	if allowance < 1 {
		if err := tb.saveAllowance(ctx, 0, current); err != nil {
			return false, err
		}
		return false, nil
	}

	if err := tb.saveAllowance(ctx, allowance-1, current); err != nil {
		return false, err
	}

	return true, nil
}

func (tb *TokenBucket) loadAllowance(ctx context.Context) (float64, int64, error) {
	result, err := tb.client.HGetAll(ctx, tb.key).Result()
	if err != nil {
		return 0, 0, err
	}

	if len(result) == 0 {
		return float64(tb.limit), time.Now().Unix(), nil
	}

	var allowance float64
	var timestamp int64

	if val, exists := result["allowance"]; exists {
		fmt.Sscanf(val, "%f", &allowance)
	}
	if val, exists := result["timestamp"]; exists {
		fmt.Sscanf(val, "%d", &timestamp)
	}

	return allowance, timestamp, nil
}

func (tb *TokenBucket) saveAllowance(ctx context.Context, allowance float64, timestamp int64) error {
	data := map[string]interface{}{
		"allowance": fmt.Sprintf("%.6f", allowance),
		"timestamp": timestamp,
	}

	ttl := time.Duration(tb.window+1) * time.Second

	_, err := tb.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		if err := pipe.HSet(ctx, tb.key, data).Err(); err != nil {
			return err
		}
		return pipe.Expire(ctx, tb.key, ttl).Err()
	})

	return err
}
