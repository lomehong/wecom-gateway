package ratelimit

import (
	"sync"
	"testing"
	"time"
)

func TestNewRateLimiter(t *testing.T) {
	limiter := NewRateLimiter()
	if limiter == nil {
		t.Fatal("NewRateLimiter returned nil")
	}

	if limiter.buckets == nil {
		t.Error("buckets map should not be nil")
	}
}

func TestAllow_Basic(t *testing.T) {
	limiter := NewRateLimiter()
	apiKey := "test-key"

	// First request should be allowed
	if !limiter.Allow(apiKey, 10) {
		t.Error("First request should be allowed")
	}

	// Second request should also be allowed
	if !limiter.Allow(apiKey, 10) {
		t.Error("Second request should be allowed")
	}
}

func TestAllow_RateLimit(t *testing.T) {
	limiter := NewRateLimiter()
	apiKey := "rate-limit-key"

	// Allow 5 requests per minute
	requestsPerMinute := 5

	// First 5 requests should be allowed
	for i := 0; i < requestsPerMinute; i++ {
		if !limiter.Allow(apiKey, requestsPerMinute) {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 6th request should be denied
	if limiter.Allow(apiKey, requestsPerMinute) {
		t.Error("Request exceeding rate limit should be denied")
	}
}

func TestAllow_DifferentKeys(t *testing.T) {
	limiter := NewRateLimiter()

	// Different API keys should have independent rate limits
	key1 := "key-1"
	key2 := "key-2"

	// Use up key1's limit
	for i := 0; i < 5; i++ {
		limiter.Allow(key1, 5)
	}

	// key1 should be rate limited
	if limiter.Allow(key1, 5) {
		t.Error("key1 should be rate limited")
	}

	// key2 should still be allowed
	if !limiter.Allow(key2, 5) {
		t.Error("key2 should be allowed (independent rate limit)")
	}
}

func TestAllow_SlidingWindow(t *testing.T) {
	limiter := NewRateLimiter()
	apiKey := "sliding-window-key"

	requestsPerMinute := 5

	// Make 5 requests
	for i := 0; i < requestsPerMinute; i++ {
		if !limiter.Allow(apiKey, requestsPerMinute) {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// Verify rate limit is hit
	if limiter.Allow(apiKey, requestsPerMinute) {
		t.Error("Should be rate limited")
	}

	// Wait for window to slide (this is a simplified test)
	// In a real implementation, you'd need to mock time or use a shorter window
	time.Sleep(100 * time.Millisecond)

	// The implementation might still limit, so this is more of a smoke test
	_ = limiter.Allow(apiKey, requestsPerMinute)
}

func TestAllow_Concurrent(t *testing.T) {
	limiter := NewRateLimiter()
	apiKey := "concurrent-key"

	requestsPerMinute := 100
	numGoroutines := 10
	requestsPerGoroutine := 10

	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerGoroutine; j++ {
				if limiter.Allow(apiKey, requestsPerMinute) {
					mu.Lock()
					successCount++
					mu.Unlock()
				}
			}
		}()
	}

	wg.Wait()

	// All requests should be allowed since we're under the limit
	expectedSuccess := numGoroutines * requestsPerGoroutine
	if successCount != expectedSuccess {
		t.Errorf("Expected %d successful requests, got %d", expectedSuccess, successCount)
	}
}

func TestAllow_ZeroLimit(t *testing.T) {
	limiter := NewRateLimiter()
	apiKey := "zero-limit-key"

	// With zero limit, rate limiting is disabled, all requests should be allowed
	if !limiter.Allow(apiKey, 0) {
		t.Error("Request should be allowed when limit is 0 (no rate limiting)")
	}

	if !limiter.Allow(apiKey, 0) {
		t.Error("Request should still be allowed when limit is 0 (no rate limiting)")
	}
}

func TestAllow_HighLimit(t *testing.T) {
	limiter := NewRateLimiter()
	apiKey := "high-limit-key"

	// Test with a very high limit
	highLimit := 10000

	for i := 0; i < 100; i++ {
		if !limiter.Allow(apiKey, highLimit) {
			t.Errorf("Request %d should be allowed with high limit", i+1)
		}
	}
}

func TestReset(t *testing.T) {
	limiter := NewRateLimiter()
	apiKey := "reset-key"

	// Use up the limit
	for i := 0; i < 5; i++ {
		limiter.Allow(apiKey, 5)
	}

	// Should be rate limited
	if limiter.Allow(apiKey, 5) {
		t.Error("Should be rate limited before reset")
	}

	// Reset the limiter for this key
	limiter.Reset(apiKey)

	// Should now be allowed again
	if !limiter.Allow(apiKey, 5) {
		t.Error("Request should be allowed after reset")
	}
}

func TestResetAll(t *testing.T) {
	limiter := NewRateLimiter()

	// Add multiple keys
	keys := []string{"key-1", "key-2", "key-3"}
	for _, key := range keys {
		for i := 0; i < 5; i++ {
			limiter.Allow(key, 5)
		}
	}

	// All should be rate limited
	for _, key := range keys {
		if limiter.Allow(key, 5) {
			t.Errorf("Key %s should be rate limited", key)
		}
	}

	// Reset all
	limiter.ResetAll()

	// All should now be allowed
	for _, key := range keys {
		if !limiter.Allow(key, 5) {
			t.Errorf("Key %s should be allowed after reset all", key)
		}
	}
}
