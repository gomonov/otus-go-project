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
