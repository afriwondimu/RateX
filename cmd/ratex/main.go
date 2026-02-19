package main

import (
    "time"
    "github.com/gin-gonic/gin"
    "github.com/afriwondimu/RateX/limiter"
    "github.com/afriwondimu/RateX/limiter/middleware"
)

func main() {
    // Create rate limiter
    l, _ := limiter.New(&limiter.Config{
        Algorithm:   "token-bucket",
        MaxRequests: 10,
        WindowSize:  time.Minute,
        BurstSize:   5,
    })
    defer l.Close()

    // Use with Gin
    r := gin.Default()
    r.Use(middleware.RateLimitByIP(l))

    r.GET("/", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "Hello!"})
    })

    r.Run()
}