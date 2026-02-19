package middleware

import (
    "net/http"
    "strconv"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/afriwondimu/RateX/limiter"
)

// Config holds middleware configuration
type Config struct {
    // KeyFunc generates a unique key for each request (IP, API key, user ID, etc.)
    KeyFunc func(*gin.Context) string
    
    // SkipFunc determines if rate limiting should be skipped
    SkipFunc func(*gin.Context) bool
    
    // OnLimitReached is called when rate limit is exceeded
    OnLimitReached func(*gin.Context)
    
    // Headers configuration
    Headers struct {
        Limit     string // X-RateLimit-Limit
        Remaining string // X-RateLimit-Remaining
        Reset     string // X-RateLimit-Reset
        RetryAfter string // Retry-After
    }
}

// RateLimiter creates a Gin middleware for rate limiting
func RateLimiter(l *limiter.Limiter, cfg Config) gin.HandlerFunc {
    // Set default values
    if cfg.KeyFunc == nil {
        cfg.KeyFunc = defaultKeyFunc
    }

    if cfg.SkipFunc == nil {
        cfg.SkipFunc = func(c *gin.Context) bool { return false }
    }

    if cfg.OnLimitReached == nil {
        cfg.OnLimitReached = defaultOnLimitReached
    }

    // Set default headers
    if cfg.Headers.Limit == "" {
        cfg.Headers.Limit = "X-RateLimit-Limit"
    }
    if cfg.Headers.Remaining == "" {
        cfg.Headers.Remaining = "X-RateLimit-Remaining"
    }
    if cfg.Headers.Reset == "" {
        cfg.Headers.Reset = "X-RateLimit-Reset"
    }
    if cfg.Headers.RetryAfter == "" {
        cfg.Headers.RetryAfter = "Retry-After"
    }

    return func(c *gin.Context) {
        // Skip rate limiting if needed
        if cfg.SkipFunc(c) {
            c.Next()
            return
        }

        // Get key for this request
        key := cfg.KeyFunc(c)

        // Check rate limit
        allowed, info, err := l.Allow(key)
        if err != nil {
            // Log error but don't block request in production
            c.Next()
            return
        }

        // Set rate limit headers
        if info != nil {
            c.Header(cfg.Headers.Limit, strconv.FormatInt(info.Limit, 10))
            c.Header(cfg.Headers.Remaining, strconv.FormatInt(info.Remaining, 10))
            
            if !info.Reset.IsZero() {
                c.Header(cfg.Headers.Reset, strconv.FormatInt(info.Reset.Unix(), 10))
            }
            
            if info.RetryAfter > 0 {
                c.Header(cfg.Headers.RetryAfter, strconv.FormatInt(int64(info.RetryAfter.Seconds()), 10))
            }
        }

        // Block if not allowed
        if !allowed {
            cfg.OnLimitReached(c)
            return
        }

        c.Next()
    }
}

// defaultKeyFunc generates a key based on API key, token, IP, or UUID
func defaultKeyFunc(c *gin.Context) string {
    // Try API key first
    apiKey := c.GetHeader("X-API-Key")
    if apiKey != "" {
        return "api:" + apiKey
    }

    // Try authorization header
    auth := c.GetHeader("Authorization")
    if auth != "" {
        if strings.HasPrefix(auth, "Bearer ") {
            return "token:" + auth[7:]
        }
        return "auth:" + auth
    }

    // Try API token
    token := c.Query("token")
    if token != "" {
        return "token:" + token
    }

    // Fall back to IP
    ip := c.ClientIP()
    if ip != "" && ip != "::1" {
        return "ip:" + ip
    }

    // Last resort: generate UUID for this request
    return "uuid:" + uuid.New().String()
}

// defaultOnLimitReached is the default handler for rate limit exceeded
func defaultOnLimitReached(c *gin.Context) {
    c.JSON(http.StatusTooManyRequests, gin.H{
        "error": "Too Many Requests",
        "message": "Rate limit exceeded. Please try again later.",
    })
    c.Abort()
}

// RateLimitByIP returns a middleware that rate limits by IP address
func RateLimitByIP(l *limiter.Limiter) gin.HandlerFunc {
    return RateLimiter(l, Config{
        KeyFunc: func(c *gin.Context) string {
            return "ip:" + c.ClientIP()
        },
    })
}

// RateLimitByAPIKey returns a middleware that rate limits by API key
func RateLimitByAPIKey(l *limiter.Limiter) gin.HandlerFunc {
    return RateLimiter(l, Config{
        KeyFunc: func(c *gin.Context) string {
            apiKey := c.GetHeader("X-API-Key")
            if apiKey == "" {
                apiKey = "anonymous"
            }
            return "api:" + apiKey
        },
    })
}