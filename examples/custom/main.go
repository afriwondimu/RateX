package main

import (
    "log"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/afriwondimu/RateX/limiter"
    "github.com/afriwondimu/RateX/limiter/middleware"
)

// Custom key function for user-based rate limiting
func userKeyFunc(c *gin.Context) string {
    // Get user ID from JWT token or session
    userID := c.GetHeader("X-User-ID")
    if userID == "" {
        // Try to get from query parameter
        userID = c.Query("user_id")
    }
    if userID == "" {
        // Fall back to IP
        return "ip:" + c.ClientIP()
    }
    return "user:" + userID
}

// Custom rate limit exceeded handler
func customLimitHandler(c *gin.Context) {
    c.Header("X-RateX-Error", "rate_limit_exceeded")
    c.JSON(http.StatusTooManyRequests, gin.H{
        "code":    429,
        "error":   "Rate limit exceeded",
        "message": "You have made too many requests. Please wait before trying again.",
        "retry_in": c.GetHeader("Retry-After"),
    })
    c.Abort()
}

func main() {
    // Create limiter with sliding window algorithm
    l, err := limiter.New(&limiter.Config{
        Algorithm:   "sliding-window",
        MaxRequests: 5,               // 5 requests
        WindowSize:  10 * time.Second, // per 10 seconds
        BurstSize:   5,                // same as max for sliding window
    })
    if err != nil {
        log.Fatal("Failed to create limiter:", err)
    }
    defer l.Close()

    r := gin.Default()

    // Apply custom middleware
    r.Use(middleware.RateLimiter(l, middleware.Config{
        KeyFunc:        userKeyFunc,
        OnLimitReached: customLimitHandler,
        Headers: struct {
            Limit     string
            Remaining string
            Reset     string
            RetryAfter string
        }{
            Limit:      "X-RateX-Limit",
            Remaining:  "X-RateX-Remaining",
            Reset:      "X-RateX-Reset",
            RetryAfter: "X-RateX-Retry-After",
        },
    }))

    // Routes
    r.GET("/profile", func(c *gin.Context) {
        userID := c.GetHeader("X-User-ID")
        if userID == "" {
            userID = "anonymous"
        }
        c.JSON(http.StatusOK, gin.H{
            "user_id": userID,
            "profile": "User profile data",
        })
    })

    r.POST("/action", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "status": "action completed",
        })
    })

    log.Println("Server starting on http://localhost:8080")
    r.Run(":8080")
}