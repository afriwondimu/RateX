package storage

import "time"

// Storage defines the interface for rate limiter storage backends
type Storage interface {
    // Get retrieves a value by key
    Get(key string) (interface{}, bool)
    
    // Set stores a value with TTL
    Set(key string, value interface{}, ttl time.Duration)
    
    // Delete removes a key
    Delete(key string)
    
    // Clear removes all keys
    Clear() error
    
    // Close closes the storage connection
    Close() error
}

// RateLimitInfo contains rate limit information from storage
type RateLimitInfo struct {
    Limit     int64
    Remaining int64
    Reset     time.Time
}