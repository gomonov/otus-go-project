package ratelimit

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client *redis.Client
	config Config
}

type Config struct {
	LoginLimit    int
	PasswordLimit int
	IPLimit       int
	Window        int
}

func NewRateLimiter(client *redis.Client, config Config) *RateLimiter {
	return &RateLimiter{
		client: client,
		config: config,
	}
}

func (r *RateLimiter) Check(ctx context.Context, login, password, ip string) error {
	loginBucket := NewTokenBucket(r.client, "ratelimit:login:"+login, r.config.LoginLimit, r.config.Window)
	allowed, err := loginBucket.Allow(ctx)
	if err != nil {
		return fmt.Errorf("login limit check failed: %w", err)
	}
	if !allowed {
		return fmt.Errorf("login limit exceeded: %s", login)
	}

	passwordBucket := NewTokenBucket(r.client, "ratelimit:password:"+password, r.config.PasswordLimit, r.config.Window)
	allowed, err = passwordBucket.Allow(ctx)
	if err != nil {
		return fmt.Errorf("password limit check failed: %w", err)
	}
	if !allowed {
		return fmt.Errorf("password limit exceeded")
	}

	ipBucket := NewTokenBucket(r.client, "ratelimit:ip:"+ip, r.config.IPLimit, r.config.Window)
	allowed, err = ipBucket.Allow(ctx)
	if err != nil {
		return fmt.Errorf("ip limit check failed: %w", err)
	}
	if !allowed {
		return fmt.Errorf("ip limit exceeded: %s", ip)
	}

	return nil
}

func (r *RateLimiter) ResetBuckets(ctx context.Context, login, ip string) error {
	keys := []string{
		fmt.Sprintf("ratelimit:login:%s", login),
		fmt.Sprintf("ratelimit:ip:%s", ip),
	}

	_, err := r.client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for _, key := range keys {
			if err := pipe.Del(ctx, key).Err(); err != nil {
				return fmt.Errorf("failed to delete key %s: %w", key, err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to reset buckets: %w", err)
	}

	return nil
}
