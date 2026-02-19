// internal/middleware/rate_limit.go
package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/afriwondimu/RateX/internal/limiter"
	"github.com/afriwondimu/RateX/internal/storage"
	"github.com/gin-gonic/gin"
)

type Config struct {
	Rate       float64
	Capacity   float64
	ByIP       bool
	ByAPIKey   bool
	UseRedis   bool
}

var (
	buckets = make(map[string]*limiter.TokenBucket)
	mu      sync.RWMutex
)

func getBucket(key string, rate, capacity float64) *limiter.TokenBucket {
	mu.RLock()
	bucket, exists := buckets[key]
	mu.RUnlock()

	if exists {
		return bucket
	}

	mu.Lock()
	defer mu.Unlock()
	if bucket, exists = buckets[key]; exists {
		return bucket
	}
	bucket = limiter.NewTokenBucket(rate, capacity)
	buckets[key] = bucket
	return bucket
}

func RateLimit(redisStore *storage.RedisStore, cfg Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var key string

		switch {
		case cfg.ByIP:
			key = "rate:ip:" + c.ClientIP()
		case cfg.ByAPIKey:
			key = "rate:apikey:" + c.GetHeader("X-API-Key")
		default:
			key = "rate:global"
		}

		// Use Redis if configured
		if cfg.UseRedis && redisStore != nil {
			count, err := redisStore.Increment(c.Request.Context(), key, time.Minute)
			if err != nil {
				c.Next()
				return
			}
			
			c.Header("X-RateLimit-Limit", strconv.Itoa(int(cfg.Capacity)))
			remaining := int(cfg.Capacity) - count
			if remaining < 0 {
				remaining = 0
			}
			c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
			
			if count > int(cfg.Capacity) {
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
					"error": "rate limit exceeded",
				})
				return
			}
		} else {
			// Use in-memory token bucket
			bucket := getBucket(key, cfg.Rate, cfg.Capacity)
			
			c.Header("X-RateLimit-Limit", strconv.Itoa(int(cfg.Capacity)))
			remaining := int(bucket.GetTokens())
			c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
			
			if !bucket.Allow() {
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
					"error": "rate limit exceeded",
				})
				return
			}
		}

		c.Next()
	}
}