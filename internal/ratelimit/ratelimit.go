package ratelimit

import (
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	mu    sync.RWMutex
	buckets map[string]*bucket
}

// bucket represents a token bucket
type bucket struct {
	tokens     int
	maxTokens  int
	refillRate int           // tokens per second
	lastRefill time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		buckets: make(map[string]*bucket),
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// Allow checks if a request is allowed for the given key
func (rl *RateLimiter) Allow(key string, requestsPerMinute int) bool {
	if requestsPerMinute <= 0 {
		return true // No rate limiting
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Get or create bucket
	b, exists := rl.buckets[key]
	if !exists {
		b = &bucket{
			tokens:     requestsPerMinute,
			maxTokens:  requestsPerMinute,
			refillRate: requestsPerMinute / 60, // Convert to tokens per second
			lastRefill: now,
		}
		rl.buckets[key] = b
	}

	// Refill tokens
	elapsed := now.Sub(b.lastRefill).Seconds()
	tokensToAdd := int(elapsed * float64(b.refillRate))

	if tokensToAdd > 0 {
		b.tokens += tokensToAdd
		if b.tokens > b.maxTokens {
			b.tokens = b.maxTokens
		}
		b.lastRefill = now
	}

	// Check if we have tokens available
	if b.tokens > 0 {
		b.tokens--
		return true
	}

	return false
}

// cleanup removes old buckets periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()

		// Remove buckets that haven't been used in the last 10 minutes
		for key, bucket := range rl.buckets {
			if now.Sub(bucket.lastRefill) > 10*time.Minute {
				delete(rl.buckets, key)
			}
		}

		rl.mu.Unlock()
	}
}

// GetStats returns statistics about the rate limiter
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["total_buckets"] = len(rl.buckets)

	return stats
}

// Reset resets the rate limiter for a specific key
func (rl *RateLimiter) Reset(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	delete(rl.buckets, key)
}

// ResetAll clears all buckets
func (rl *RateLimiter) ResetAll() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.buckets = make(map[string]*bucket)
}
