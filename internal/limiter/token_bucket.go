package limiter

import (
	"sync"
	"time"
)

type TokenBucket struct {
	rate       float64
	capacity   float64
	tokens     float64
	lastUpdate time.Time
	mu         sync.Mutex
}

func NewTokenBucket(rate float64, capacity float64) *TokenBucket {
	return &TokenBucket{
		rate:       rate,
		capacity:   capacity,
		tokens:     capacity,
		lastUpdate: time.Now(),
	}
}

func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	// Add tokens based on time elapsed
	tb.tokens += now.Sub(tb.lastUpdate).Seconds() * tb.rate
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}

	// Check if we have at least 1 token
	if tb.tokens < 1 {
		return false
	}

	// Consume 1 token
	tb.tokens -= 1
	tb.lastUpdate = now
	return true
}

func (tb *TokenBucket) GetTokens() float64 {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	
	now := time.Now()
	tokens := tb.tokens + now.Sub(tb.lastUpdate).Seconds()*tb.rate
	if tokens > tb.capacity {
		tokens = tb.capacity
	}
	return tokens
}