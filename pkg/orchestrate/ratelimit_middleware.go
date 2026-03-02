package orchestrate

import (
	"fmt"
	"net/http"
	"time"
)

// RateLimitMiddleware wraps HTTP handlers with rate limiting.
type RateLimitMiddleware struct {
	limiter *RateLimiter
	config  RateLimitConfig
}

// NewRateLimitMiddleware creates a new rate limit middleware.
func NewRateLimitMiddleware(limiter *RateLimiter, config RateLimitConfig) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiter: limiter,
		config:  config,
	}
}

// Middleware applies rate limiting to HTTP requests.
func (m *RateLimitMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.config.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Extract rate limit key from request
		// Can be IP address, API key, user ID, etc.
		key := m.extractKey(r)

		// Check rate limit
		window := time.Minute
		allowed, err := m.limiter.Allow(r.Context(), key, m.config.RequestsPerMin, window)

		if err != nil {
			http.Error(w, "Rate limit error", http.StatusInternalServerError)
			return
		}

		if !allowed {
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", m.config.RequestsPerMin))
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("Retry-After", "60")
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		// Set rate limit headers
		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", m.config.RequestsPerMin))

		next.ServeHTTP(w, r)
	})
}

// extractKey extracts the rate limit key from the request.
func (m *RateLimitMiddleware) extractKey(r *http.Request) string {
	// Try API key first
	if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
		return "apikey:" + apiKey
	}

	// Fall back to IP address
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}

	return "ip:" + ip
}
