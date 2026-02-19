// cmd/server/main.go
package main

import (
	"log"
	"time"

	"github.com/afriwondimu/RateX/internal/middleware"
	"github.com/afriwondimu/RateX/internal/storage"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	redisStore := storage.NewRedisStore("localhost:6379")

	// API endpoint with VERY strict rate limiting
	api := r.Group("/api")
	api.Use(middleware.RateLimit(redisStore, middleware.Config{
		Rate:     0.1,    // 1 token every 10 seconds
		Capacity: 5,      // only 5 requests allowed
		ByIP:     true,
		UseRedis: false,
	}))

	api.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	log.Fatal(r.Run(":8080"))
}