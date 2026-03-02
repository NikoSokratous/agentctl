package orchestrate

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter provides rate limiting capabilities.
type RateLimiter struct {
	redis      *redis.Client
	localStore *localRateLimitStore
	useRedis   bool
}

// RateLimitConfig holds rate limiting configuration.
type RateLimitConfig struct {
	Enabled        bool   `yaml:"enabled" json:"enabled"`
	RequestsPerMin int    `yaml:"requests_per_minute" json:"requests_per_minute"`
	BurstSize      int    `yaml:"burst_size" json:"burst_size"`
	RedisURL       string `yaml:"redis_url" json:"redis_url,omitempty"`
}

// localRateLimitStore is an in-memory rate limiter.
type localRateLimitStore struct {
	mu      sync.RWMutex
	buckets map[string]*tokenBucket
}

type tokenBucket struct {
	tokens     int
	maxTokens  int
	refillRate time.Duration
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(config RateLimitConfig) (*RateLimiter, error) {
	rl := &RateLimiter{
		localStore: &localRateLimitStore{
			buckets: make(map[string]*tokenBucket),
		},
	}

	// If Redis URL is provided, use Redis-backed rate limiting
	if config.RedisURL != "" {
		opts, err := redis.ParseURL(config.RedisURL)
		if err != nil {
			return nil, fmt.Errorf("parse Redis URL: %w", err)
		}

		rl.redis = redis.NewClient(opts)
		rl.useRedis = true

		// Test connection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := rl.redis.Ping(ctx).Err(); err != nil {
			return nil, fmt.Errorf("connect to Redis: %w", err)
		}
	}

	return rl, nil
}

// Allow checks if a request is allowed for the given key.
func (rl *RateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	if rl.useRedis {
		return rl.allowRedis(ctx, key, limit, window)
	}
	return rl.allowLocal(key, limit, window), nil
}

// allowRedis implements Redis-backed rate limiting using sliding window.
func (rl *RateLimiter) allowRedis(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	now := time.Now()
	windowStart := now.Add(-window)

	redisKey := fmt.Sprintf("ratelimit:%s", key)

	// Use Lua script for atomic operation
	script := `
		local key = KEYS[1]
		local now = tonumber(ARGV[1])
		local window_start = tonumber(ARGV[2])
		local limit = tonumber(ARGV[3])
		
		redis.call('ZREMRANGEBYSCORE', key, '-inf', window_start)
		
		local current = redis.call('ZCARD', key)
		
		if current < limit then
			redis.call('ZADD', key, now, now)
			redis.call('EXPIRE', key, 60)
			return 1
		else
			return 0
		end
	`

	result, err := rl.redis.Eval(ctx, script, []string{redisKey},
		now.UnixNano(),
		windowStart.UnixNano(),
		limit,
	).Int()

	if err != nil {
		return false, err
	}

	return result == 1, nil
}

// allowLocal implements in-memory token bucket rate limiting.
func (rl *RateLimiter) allowLocal(key string, limit int, window time.Duration) bool {
	rl.localStore.mu.Lock()
	bucket, exists := rl.localStore.buckets[key]
	if !exists {
		bucket = &tokenBucket{
			tokens:     limit,
			maxTokens:  limit,
			refillRate: window / time.Duration(limit),
			lastRefill: time.Now(),
		}
		rl.localStore.buckets[key] = bucket
	}
	rl.localStore.mu.Unlock()

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	// Refill tokens
	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill)
	tokensToAdd := int(elapsed / bucket.refillRate)

	if tokensToAdd > 0 {
		bucket.tokens = min(bucket.tokens+tokensToAdd, bucket.maxTokens)
		bucket.lastRefill = now
	}

	// Check if request is allowed
	if bucket.tokens > 0 {
		bucket.tokens--
		return true
	}

	return false
}

// Reset clears rate limit for a key.
func (rl *RateLimiter) Reset(ctx context.Context, key string) error {
	if rl.useRedis {
		redisKey := fmt.Sprintf("ratelimit:%s", key)
		return rl.redis.Del(ctx, redisKey).Err()
	}

	rl.localStore.mu.Lock()
	delete(rl.localStore.buckets, key)
	rl.localStore.mu.Unlock()

	return nil
}

// Close closes the rate limiter connections.
func (rl *RateLimiter) Close() error {
	if rl.redis != nil {
		return rl.redis.Close()
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
