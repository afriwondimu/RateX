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
    // Create rate limiter
    l, err := limiter.New(&limiter.Config{
        Algorithm:   "token-bucket",  // Use token bucket algorithm
        MaxRequests: 10,              // 10 requests
        WindowSize:  time.Minute,     // per minute
        BurstSize:   5,               // burst of 5 requests
    })
    if err != nil {
        log.Fatal("Failed to create limiter:", err)
    }
    defer l.Close()

    // Create Gin router
    r := gin.Default()

    // Apply rate limiting middleware to all routes
    r.Use(middleware.RateLimiter(l, middleware.Config{
        KeyFunc: func(c *gin.Context) string {
            // Rate limit by IP
            return "ip:" + c.ClientIP()
        },
    }))

    // Public routes
    r.GET("/", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "message": "Welcome to RateX!",
            "time":    time.Now().Format(time.RFC3339),
        })
    })

    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "status": "healthy",
        })
    })

    // API routes
    api := r.Group("/api")
    {
        api.GET("/users", func(c *gin.Context) {
            c.JSON(http.StatusOK, gin.H{
                "users": []string{"alice", "bob", "charlie"},
            })
        })

        api.GET("/data", func(c *gin.Context) {
            c.JSON(http.StatusOK, gin.H{
                "data": "Sensitive information",
            })
        })
    }

    log.Println("Server starting on http://localhost:8080")
    log.Println("Try: curl -i http://localhost:8080/")
    r.Run(":8080")
}