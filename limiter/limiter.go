package limiter

import (
    "errors"
    "fmt"
    "time"
    
    "github.com/afriwondimu/RateX/limiter/algorithms"
    "github.com/afriwondimu/RateX/limiter/storage"
)

// Limiter is the main rate limiter struct
type Limiter struct {
    algorithm algorithms.RateLimiter
    storage   storage.Storage
    config    *Config
}

// Config holds the rate limiter configuration
type Config struct {
    Algorithm   string        // "token-bucket" or "sliding-window"
    MaxRequests int64         // Maximum requests allowed
    WindowSize  time.Duration // Time window
    BurstSize   int64         // For token bucket
    RedisAddr   string        // Optional, if using Redis
    RedisPass   string
    RedisDB     int
}

// RateLimitInfo contains rate limit information
type RateLimitInfo struct {
    Limit      int64
    Remaining  int64
    Reset      time.Time
    RetryAfter time.Duration
}

// New creates a new rate limiter
func New(config *Config) (*Limiter, error) {
    l := &Limiter{
        config: config,
    }

    // Initialize storage
    if config.RedisAddr != "" {
        redis, err := storage.NewRedisStorage(config.RedisAddr, config.RedisPass, config.RedisDB)
        if err != nil {
            return nil, fmt.Errorf("failed to connect to Redis: %w", err)
        }
        l.storage = redis
    } else {
        l.storage = storage.NewMemoryStorage()
    }

    // Initialize algorithm if not using Redis (Redis handles algorithm internally)
    if config.RedisAddr == "" {
        var alg algorithms.RateLimiter
        
        switch config.Algorithm {
        case "token-bucket":
            rate := float64(config.MaxRequests) / config.WindowSize.Seconds()
            alg = algorithms.NewTokenBucket(rate, config.BurstSize)
        case "sliding-window":
            alg = algorithms.NewSlidingWindow(config.MaxRequests, config.WindowSize)
        default:
            return nil, errors.New("unknown algorithm: use 'token-bucket' or 'sliding-window'")
        }
        
        l.algorithm = alg
    }

    return l, nil
}

// Allow checks if a request is allowed
func (l *Limiter) Allow(key string) (bool, *RateLimitInfo, error) {
    return l.AllowN(key, 1)
}

// AllowN checks if N requests are allowed
func (l *Limiter) AllowN(key string, n int64) (bool, *RateLimitInfo, error) {
    // Use Redis if available
    if redis, ok := l.storage.(*storage.RedisStorage); ok {
        return l.redisAllowN(redis, key, n)
    }
    
    // Use in-memory algorithm
    if l.algorithm == nil {
        return false, nil, errors.New("no algorithm initialized")
    }
    
    allowed, info, err := l.algorithm.AllowN(key, n)
    if err != nil {
        return false, nil, err
    }
    
    return allowed, convertInfo(info), nil
}

func (l *Limiter) redisAllowN(redis *storage.RedisStorage, key string, n int64) (bool, *RateLimitInfo, error) {
    switch l.config.Algorithm {
    case "token-bucket":
        rate := float64(l.config.MaxRequests) / l.config.WindowSize.Seconds()
        allowed, info, err := redis.TokenBucketAllow(key, rate, l.config.BurstSize, n)
        if err != nil {
            return false, nil, err
        }
        return allowed, &RateLimitInfo{
            Limit:     info.Limit,
            Remaining: info.Remaining,
            Reset:     info.Reset,
        }, nil
    default:
        return false, nil, fmt.Errorf("algorithm '%s' not supported with Redis", l.config.Algorithm)
    }
}

// Reset clears all rate limit data for a key
func (l *Limiter) Reset(key string) error {
    l.storage.Delete(key)
    return nil
}

// Close closes the limiter and its storage
func (l *Limiter) Close() error {
    if l.storage != nil {
        return l.storage.Close()
    }
    return nil
}

func convertInfo(info *algorithms.RateLimitInfo) *RateLimitInfo {
    if info == nil {
        return nil
    }
    
    return &RateLimitInfo{
        Limit:      info.Limit,
        Remaining:  info.Remaining,
        Reset:      info.Reset,
        RetryAfter: info.RetryAfter,
    }
}