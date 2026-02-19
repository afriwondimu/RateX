package algorithms

import (

    "sync"
    "time"
)

// TokenBucket implements the token bucket algorithm
type TokenBucket struct {
    rate       float64       // Tokens added per second
    capacity   int64         // Maximum tokens in bucket
    buckets    map[string]*bucket
    mu         sync.RWMutex
}

type bucket struct {
    tokens     float64
    lastRefill time.Time
}

// NewTokenBucket creates a new token bucket rate limiter
func NewTokenBucket(rate float64, capacity int64) *TokenBucket {
    return &TokenBucket{
        rate:     rate,
        capacity: capacity,
        buckets:  make(map[string]*bucket),
    }
}

// Allow checks if a request is allowed
func (tb *TokenBucket) Allow(key string) (bool, *RateLimitInfo, error) {
    return tb.AllowN(key, 1)
}

// AllowN checks if N requests are allowed
func (tb *TokenBucket) AllowN(key string, n int64) (bool, *RateLimitInfo, error) {
    tb.mu.Lock()
    defer tb.mu.Unlock()

    now := time.Now()
    
    // Get or create bucket
    b, exists := tb.buckets[key]
    if !exists {
        b = &bucket{
            tokens:     float64(tb.capacity),
            lastRefill: now,
        }
        tb.buckets[key] = b
    }

    // Refill tokens based on elapsed time
    elapsed := now.Sub(b.lastRefill).Seconds()
    b.tokens = min(float64(tb.capacity), b.tokens+elapsed*tb.rate)
    b.lastRefill = now

    // Check if enough tokens
    if b.tokens >= float64(n) {
        b.tokens -= float64(n)
        
        // Calculate reset time (when bucket will be full again)
        tokensNeeded := float64(tb.capacity) - b.tokens
        resetSeconds := tokensNeeded / tb.rate
        resetTime := now.Add(time.Duration(resetSeconds * float64(time.Second)))
        
        info := &RateLimitInfo{
            Limit:     tb.capacity,
            Remaining: int64(b.tokens),
            Reset:     resetTime,
        }
        return true, info, nil
    }

    // Not enough tokens
    tokensNeeded := float64(n) - b.tokens
    retrySeconds := tokensNeeded / tb.rate
    
    info := &RateLimitInfo{
        Limit:      tb.capacity,
        Remaining:  0,
        RetryAfter: time.Duration(retrySeconds * float64(time.Second)),
        Reset:      now.Add(time.Duration(retrySeconds * float64(time.Second))),
    }
    return false, info, nil
}

func min(a, b float64) float64 {
    if a < b {
        return a
    }
    return b
}