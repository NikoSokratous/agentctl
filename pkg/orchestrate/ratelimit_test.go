package orchestrate

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiterLocal(t *testing.T) {
	config := RateLimitConfig{
		Enabled:        true,
		RequestsPerMin: 10,
		BurstSize:      5,
	}

	limiter, err := NewRateLimiter(config)
	if err != nil {
		t.Fatalf("NewRateLimiter failed: %v", err)
	}
	defer limiter.Close()

	ctx := context.Background()
	key := "test-key"

	// Should allow first 10 requests
	for i := 0; i < 10; i++ {
		allowed, err := limiter.Allow(ctx, key, 10, time.Minute)
		if err != nil {
			t.Fatalf("Allow failed: %v", err)
		}
		if !allowed {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 11th request should be denied
	allowed, err := limiter.Allow(ctx, key, 10, time.Minute)
	if err != nil {
		t.Fatalf("Allow failed: %v", err)
	}
	if allowed {
		t.Error("Request 11 should be denied")
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	config := RateLimitConfig{
		Enabled:        true,
		RequestsPerMin: 5,
	}

	limiter, err := NewRateLimiter(config)
	if err != nil {
		t.Fatalf("NewRateLimiter failed: %v", err)
	}
	defer limiter.Close()

	middleware := NewRateLimitMiddleware(limiter, config)

	handler := middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Make requests
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: expected 200, got %d", i+1, w.Code)
		}

		// Check rate limit headers
		if w.Header().Get("X-RateLimit-Limit") == "" {
			t.Error("Missing X-RateLimit-Limit header")
		}
	}

	// 6th request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected 429, got %d", w.Code)
	}

	// Check rate limit headers on error
	if w.Header().Get("Retry-After") == "" {
		t.Error("Missing Retry-After header")
	}
}

func TestRateLimitReset(t *testing.T) {
	config := RateLimitConfig{
		Enabled:        true,
		RequestsPerMin: 5,
	}

	limiter, err := NewRateLimiter(config)
	if err != nil {
		t.Fatalf("NewRateLimiter failed: %v", err)
	}
	defer limiter.Close()

	ctx := context.Background()
	key := "test-reset"

	// Exhaust rate limit
	for i := 0; i < 5; i++ {
		limiter.Allow(ctx, key, 5, time.Minute)
	}

	// Should be denied
	allowed, _ := limiter.Allow(ctx, key, 5, time.Minute)
	if allowed {
		t.Error("Should be rate limited")
	}

	// Reset
	if err := limiter.Reset(ctx, key); err != nil {
		t.Fatalf("Reset failed: %v", err)
	}

	// Should be allowed again
	allowed, _ = limiter.Allow(ctx, key, 5, time.Minute)
	if !allowed {
		t.Error("Should be allowed after reset")
	}
}

func TestRateLimitTokenBucketRefill(t *testing.T) {
	config := RateLimitConfig{
		Enabled:        true,
		RequestsPerMin: 60, // 1 per second
	}

	limiter, err := NewRateLimiter(config)
	if err != nil {
		t.Fatalf("NewRateLimiter failed: %v", err)
	}
	defer limiter.Close()

	ctx := context.Background()
	key := "test-refill"

	// Use up 5 tokens
	for i := 0; i < 5; i++ {
		allowed, _ := limiter.Allow(ctx, key, 60, time.Minute)
		if !allowed {
			t.Fatalf("Token %d should be available", i+1)
		}
	}

	// Wait for tokens to refill (60 req/min = 1 req/sec)
	time.Sleep(2 * time.Second)

	// Should have at least 2 tokens refilled
	for i := 0; i < 2; i++ {
		allowed, _ := limiter.Allow(ctx, key, 60, time.Minute)
		if !allowed {
			t.Errorf("Token should have refilled after waiting (attempt %d)", i+1)
		}
	}
}

func TestRateLimitDifferentKeys(t *testing.T) {
	config := RateLimitConfig{
		Enabled:        true,
		RequestsPerMin: 3,
	}

	limiter, err := NewRateLimiter(config)
	if err != nil {
		t.Fatalf("NewRateLimiter failed: %v", err)
	}
	defer limiter.Close()

	ctx := context.Background()

	// Exhaust key1
	for i := 0; i < 3; i++ {
		limiter.Allow(ctx, "key1", 3, time.Minute)
	}

	// key1 should be rate limited
	allowed, _ := limiter.Allow(ctx, "key1", 3, time.Minute)
	if allowed {
		t.Error("key1 should be rate limited")
	}

	// key2 should still work (different bucket)
	for i := 0; i < 3; i++ {
		allowed, _ := limiter.Allow(ctx, "key2", 3, time.Minute)
		if !allowed {
			t.Errorf("key2 request %d should be allowed", i+1)
		}
	}
}

func TestRateLimitMiddlewareDisabled(t *testing.T) {
	config := RateLimitConfig{
		Enabled:        false, // Disabled
		RequestsPerMin: 1,
	}

	limiter, err := NewRateLimiter(config)
	if err != nil {
		t.Fatalf("NewRateLimiter failed: %v", err)
	}
	defer limiter.Close()

	middleware := NewRateLimitMiddleware(limiter, config)

	handler := middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Make many requests - all should pass since rate limiting is disabled
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: expected 200, got %d (rate limiting should be disabled)", i+1, w.Code)
		}
	}
}

func TestRateLimitKeyExtraction(t *testing.T) {
	config := RateLimitConfig{
		Enabled:        true,
		RequestsPerMin: 5,
	}

	limiter, err := NewRateLimiter(config)
	if err != nil {
		t.Fatalf("NewRateLimiter failed: %v", err)
	}
	defer limiter.Close()

	middleware := NewRateLimitMiddleware(limiter, config)

	tests := []struct {
		name      string
		headers   map[string]string
		expectKey string
	}{
		{
			name: "API key",
			headers: map[string]string{
				"X-API-Key": "test-key-123",
			},
			expectKey: "apikey:test-key-123",
		},
		{
			name: "IP address",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.1",
			},
			expectKey: "ip:192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			extractedKey := middleware.extractKey(req)
			if extractedKey != tt.expectKey {
				t.Errorf("Expected key %s, got %s", tt.expectKey, extractedKey)
			}
		})
	}
}
