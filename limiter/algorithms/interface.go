package algorithms

import (
	"fmt"
	"time"
)

// RateLimiter defines the interface for all rate limiting algorithms
type RateLimiter interface {
    // Allow checks if a request is allowed
    Allow(key string) (bool, *RateLimitInfo, error)
    
    // AllowN checks if N requests are allowed
    AllowN(key string, n int64) (bool, *RateLimitInfo, error)
}

// RateLimitInfo contains metadata about the rate limit
type RateLimitInfo struct {
    Limit      int64         // Max requests allowed
    Remaining  int64         // Remaining requests
    Reset      time.Time     // When the limit resets
    RetryAfter time.Duration // Time to wait before retrying
}

// Config holds common configuration for rate limiters
type Config struct {
    MaxRequests int64         // Maximum requests per time window
    WindowSize  time.Duration // Time window size
    BurstSize   int64         // Max burst size (for token bucket)
}

// String returns string representation of RateLimitInfo
func (r *RateLimitInfo) String() string {
    return fmt.Sprintf("Limit: %d, Remaining: %d, Reset: %v, RetryAfter: %v", 
        r.Limit, r.Remaining, r.Reset, r.RetryAfter)
}