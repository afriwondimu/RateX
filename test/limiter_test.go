package test

import (
    "testing"
    "time"

    "github.com/afriwondimu/RateX/limiter"
    "github.com/stretchr/testify/assert"
)

func TestTokenBucketLimiter(t *testing.T) {
    l, err := limiter.New(&limiter.Config{
        Algorithm:   "token-bucket",
        MaxRequests: 10,
        WindowSize:  time.Minute,
        BurstSize:   5,
    })
    assert.NoError(t, err)
    defer l.Close()

    // Test basic allow
    allowed, info, err := l.Allow("test-key")
    assert.NoError(t, err)
    assert.True(t, allowed)
    assert.Equal(t, int64(10), info.Limit)
    assert.Equal(t, int64(9), info.Remaining)

    // Test multiple requests
    for i := 0; i < 5; i++ {
        allowed, info, err = l.Allow("test-key")
        assert.NoError(t, err)
        assert.True(t, allowed)
    }
    assert.Equal(t, int64(4), info.Remaining)

    // Test different key
    allowed, info, err = l.Allow("different-key")
    assert.NoError(t, err)
    assert.True(t, allowed)
    assert.Equal(t, int64(10), info.Limit)
    assert.Equal(t, int64(9), info.Remaining)
}

func TestSlidingWindowLimiter(t *testing.T) {
    l, err := limiter.New(&limiter.Config{
        Algorithm:   "sliding-window",
        MaxRequests: 5,
        WindowSize:  time.Second * 10,
    })
    assert.NoError(t, err)
    defer l.Close()

    // Test allows up to limit
    for i := 0; i < 5; i++ {
        allowed, _, err := l.Allow("test-key")
        assert.NoError(t, err)
        assert.True(t, allowed)
    }

    // Next should be denied
    allowed, info, err := l.Allow("test-key")
    assert.NoError(t, err)
    assert.False(t, allowed)
    assert.Equal(t, int64(0), info.Remaining)
    assert.True(t, info.RetryAfter > 0)
}

func TestAllowN(t *testing.T) {
    l, err := limiter.New(&limiter.Config{
        Algorithm:   "token-bucket",
        MaxRequests: 10,
        WindowSize:  time.Minute,
        BurstSize:   10,
    })
    assert.NoError(t, err)
    defer l.Close()

    // Allow 5 requests at once
    allowed, info, err := l.AllowN("test-key", 5)
    assert.NoError(t, err)
    assert.True(t, allowed)
    assert.Equal(t, int64(5), info.Remaining)

    // Try to allow 6 more (should fail)
    allowed, info, err = l.AllowN("test-key", 6)
    assert.NoError(t, err)
    assert.False(t, allowed)
    assert.Equal(t, int64(0), info.Remaining)
}

func TestReset(t *testing.T) {
    l, err := limiter.New(&limiter.Config{
        Algorithm:   "token-bucket",
        MaxRequests: 5,
        WindowSize:  time.Minute,
        BurstSize:   5,
    })
    assert.NoError(t, err)
    defer l.Close()

    // Use some tokens
    for i := 0; i < 3; i++ {
        l.Allow("test-key")
    }

    // Reset the key
    err = l.Reset("test-key")
    assert.NoError(t, err)

    // Should have full bucket again
    allowed, info, err := l.Allow("test-key")
    assert.NoError(t, err)
    assert.True(t, allowed)
    assert.Equal(t, int64(4), info.Remaining)
}