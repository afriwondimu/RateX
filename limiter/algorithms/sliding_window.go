package algorithms

import (
    "sync"
    "time"
)

// SlidingWindow implements the sliding window log algorithm
type SlidingWindow struct {
    limit    int64
    window   time.Duration
    requests map[string][]time.Time
    mu       sync.RWMutex
}

// NewSlidingWindow creates a new sliding window rate limiter
func NewSlidingWindow(limit int64, window time.Duration) *SlidingWindow {
    return &SlidingWindow{
        limit:    limit,
        window:   window,
        requests: make(map[string][]time.Time),
    }
}

// Allow checks if a request is allowed
func (sw *SlidingWindow) Allow(key string) (bool, *RateLimitInfo, error) {
    return sw.AllowN(key, 1)
}

// AllowN checks if N requests are allowed
func (sw *SlidingWindow) AllowN(key string, n int64) (bool, *RateLimitInfo, error) {
    sw.mu.Lock()
    defer sw.mu.Unlock()

    now := time.Now()
    cutoff := now.Add(-sw.window)

    // Get existing requests
    requests, exists := sw.requests[key]
    if !exists {
        requests = []time.Time{}
    }

    // Remove old requests
    valid := make([]time.Time, 0, len(requests))
    for _, t := range requests {
        if t.After(cutoff) {
            valid = append(valid, t)
        }
    }

    // Check if adding n requests would exceed limit
    currentCount := int64(len(valid))
    if currentCount+n > sw.limit {
        // Calculate when next request can be made
        var retryAfter time.Duration
        if len(valid) > 0 {
            // The oldest request will expire at: valid[0].Add(sw.window)
            retryAfter = valid[0].Add(sw.window).Sub(now)
            if retryAfter < 0 {
                retryAfter = 0
            }
        }
        
        info := &RateLimitInfo{
            Limit:      sw.limit,
            Remaining:  0,
            RetryAfter: retryAfter,
            Reset:      now.Add(retryAfter),
        }
        return false, info, nil
    }

    // Add new requests
    for i := int64(0); i < n; i++ {
        valid = append(valid, now)
    }
    sw.requests[key] = valid

    info := &RateLimitInfo{
        Limit:     sw.limit,
        Remaining: sw.limit - (currentCount + n),
        Reset:     now.Add(sw.window),
    }
    return true, info, nil
}