package main

import (
    "log"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/afriwondimu/RateX/limiter"
    "github.com/afriwondimu/RateX/limiter/middleware"
)

func main() {
    // Create rate limiter with Redis
    l, err := limiter.New(&limiter.Config{
        Algorithm:   "token-bucket",
        MaxRequests: 1000,            // 1000 requests
        WindowSize:  time.Hour,       // per hour
        BurstSize:   50,              // burst of 50
        RedisAddr:   "localhost:6379", // Redis address
        RedisDB:     0,
    })
    if err != nil {
        log.Fatal("Failed to create limiter:", err)
    }
    defer l.Close()

    // Create Gin router
    r := gin.Default()

    // Public routes - rate limit by IP
    public := r.Group("/")
    public.Use(middleware.RateLimitByIP(l))
    {
        public.GET("/", func(c *gin.Context) {
            c.JSON(http.StatusOK, gin.H{
                "message": "Public endpoint - limited by IP",
            })
        })

        public.GET("/status", func(c *gin.Context) {
            c.JSON(http.StatusOK, gin.H{
                "status": "online",
            })
        })
    }

    // API routes - rate limit by API key with custom handler
    api := r.Group("/api")
    api.Use(middleware.RateLimiter(l, middleware.Config{
        KeyFunc: func(c *gin.Context) string {
            apiKey := c.GetHeader("X-API-Key")
            if apiKey == "" {
                apiKey = "anonymous"
            }
            return "api:" + apiKey
        },
        OnLimitReached: func(c *gin.Context) {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error":   "API rate limit exceeded",
                "message": "Please upgrade your plan or try again later",
                "docs":    "https://docs.ratex.com/rate-limits",
            })
            c.Abort()
        },
    }))
    {
        api.GET("/users", func(c *gin.Context) {
            c.JSON(http.StatusOK, gin.H{"users": []string{"alice", "bob"}})
        })

        api.POST("/data", func(c *gin.Context) {
            c.JSON(http.StatusOK, gin.H{"status": "created"})
        })
    }

    // Admin routes - different rate limits
    admin := r.Group("/admin")
    admin.Use(middleware.RateLimiter(l, middleware.Config{
        KeyFunc: func(c *gin.Context) string {
            // More complex key: combination of IP and role
            role := c.GetHeader("X-Admin-Role")
            return "admin:" + c.ClientIP() + ":" + role
        },
    }))
    {
        admin.GET("/stats", func(c *gin.Context) {
            c.JSON(http.StatusOK, gin.H{"stats": "admin statistics"})
        })
    }

    log.Println("Server starting on http://localhost:8080")
    log.Println("Make sure Redis is running on localhost:6379")
    r.Run(":8080")
}