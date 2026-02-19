package test

import (
    "testing"
    "time"

    "github.com/afriwondimu/RateX/limiter"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestRedisLimiter(t *testing.T) {
    // Skip if not running integration tests
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    l, err := limiter.New(&limiter.Config{
        Algorithm:   "token-bucket",
        MaxRequests: 10,
        WindowSize:  time.Minute,
        BurstSize:   5,
        RedisAddr:   "localhost:6379",
        RedisDB:     0,
    })
    
    // Skip if Redis is not available
    if err != nil {
        t.Skip("Redis not available:", err)
    }
    defer l.Close()

    t.Run("Basic Rate Limiting", func(t *testing.T) {
        key := "redis-test-basic"
        
        // Should allow 5 requests
        for i := 0; i < 5; i++ {
            allowed, info, err := l.Allow(key)
            assert.NoError(t, err)
            assert.True(t, allowed)
            assert.NotNil(t, info)
        }
    })

    t.Run("Limit Exceeded", func(t *testing.T) {
        key := "redis-test-limit"
        
        // Use all tokens
        for i := 0; i < 5; i++ {
            allowed, _, err := l.Allow(key)
            require.NoError(t, err)
            require.True(t, allowed)
        }
        
        // Next should be denied
        allowed, info, err := l.Allow(key)
        assert.NoError(t, err)
        assert.False(t, allowed)
        assert.Equal(t, int64(0), info.Remaining)
    })

    t.Run("Different Keys", func(t *testing.T) {
        key1 := "redis-test-key1"
        key2 := "redis-test-key2"
        
        // Use key1
        for i := 0; i < 3; i++ {
            allowed, _, err := l.Allow(key1)
            assert.NoError(t, err)
            assert.True(t, allowed)
        }
        
        // key2 should have full bucket
        allowed, info, err := l.Allow(key2)
        assert.NoError(t, err)
        assert.True(t, allowed)
        assert.Equal(t, int64(4), info.Remaining)
    })

    t.Run("AllowN", func(t *testing.T) {
        key := "redis-test-allown"
        
        allowed, info, err := l.AllowN(key, 3)
        assert.NoError(t, err)
        assert.True(t, allowed)
        assert.Equal(t, int64(2), info.Remaining)
        
        allowed, info, err = l.AllowN(key, 3)
        assert.NoError(t, err)
        assert.False(t, allowed)
        assert.Equal(t, int64(0), info.Remaining)
    })
}

func TestRedisWithSlidingWindow(t *testing.T) {
    // Skip if not running integration tests
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    // Sliding window with Redis should return error
    l, err := limiter.New(&limiter.Config{
        Algorithm:   "sliding-window",
        MaxRequests: 10,
        WindowSize:  time.Minute,
        RedisAddr:   "localhost:6379",
    })
    
    if err == nil {
        defer l.Close()
        
        // Should fail because sliding window not supported with Redis
        _, _, err = l.Allow("test")
        assert.Error(t, err)
    }
}