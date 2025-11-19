// internal/ratelimit/tokenbucket_test.go
package ratelimit

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) (*redis.Client, func()) {
	t.Helper()

	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	cleanup := func() {
		client.Close()
		mr.Close()
	}

	return client, cleanup
}

func TestTokenBucket_BasicFunctionality(t *testing.T) {
	client, cleanup := setupTest(t)
	defer cleanup()

	bucket := NewTokenBucket(client, "test:basic", 3, 60)
	ctx := context.Background()

	// Новый bucket должен позволить 3 запроса
	for i := 0; i < 3; i++ {
		allowed, err := bucket.Allow(ctx)
		assert.NoError(t, err)
		assert.True(t, allowed)
	}

	// 4-й должен быть отклонен
	allowed, err := bucket.Allow(ctx)
	assert.NoError(t, err)
	assert.False(t, allowed)
}

func TestTokenBucket_RefillLogic(t *testing.T) {
	client, cleanup := setupTest(t)
	defer cleanup()

	bucket := NewTokenBucket(client, "test:refill", 10, 10)
	ctx := context.Background()

	// Симулируем: 0 токенов 5 секунд назад
	initialTime := time.Now().Unix() - 5
	err := bucket.saveAllowance(ctx, 0, initialTime)
	require.NoError(t, err)

	// Должно быть 5 токенов (5 * 10 / 10)
	allowed, err := bucket.Allow(ctx)
	assert.NoError(t, err)
	assert.True(t, allowed)

	// Проверяем что осталось ~4 токена
	result, err := client.HGetAll(ctx, "test:refill").Result()
	require.NoError(t, err)

	var allowance float64
	if val, exists := result["allowance"]; exists {
		fmt.Sscanf(val, "%f", &allowance)
	}
	assert.InDelta(t, 4.0, allowance, 0.1)
}

func TestTokenBucket_IndependentBuckets(t *testing.T) {
	client, cleanup := setupTest(t)
	defer cleanup()

	bucket1 := NewTokenBucket(client, "test:bucket1", 2, 60)
	bucket2 := NewTokenBucket(client, "test:bucket2", 2, 60)
	ctx := context.Background()

	// Исчерпываем bucket1 с проверкой ошибок
	allowed, err := bucket1.Allow(ctx)
	require.NoError(t, err)
	assert.True(t, allowed)

	allowed, err = bucket1.Allow(ctx)
	require.NoError(t, err)
	assert.True(t, allowed)

	// Третий запрос должен быть отклонен
	allowed, err = bucket1.Allow(ctx)
	require.NoError(t, err)
	assert.False(t, allowed, "Bucket1 should be exhausted")

	// Bucket2 все еще работает
	allowed, err = bucket2.Allow(ctx)
	assert.NoError(t, err)
	assert.True(t, allowed, "Bucket2 should still work")
}
