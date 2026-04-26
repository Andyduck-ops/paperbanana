package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimitConfig holds rate limiting configuration.
type RateLimitConfig struct {
	RequestsPerMinute int
	Burst             int
	CleanupInterval   time.Duration
}

// tokenBucket represents a token bucket for rate limiting.
type tokenBucket struct {
	tokens     int
	lastUpdate time.Time
	mu         sync.Mutex
}

// rateLimiter manages rate limiting for multiple keys.
type rateLimiter struct {
	buckets     map[string]*tokenBucket
	mu          sync.RWMutex
	config      RateLimitConfig
	stopCleanup chan struct{}
}

// newRateLimiter creates a new rate limiter.
func newRateLimiter(config RateLimitConfig) *rateLimiter {
	if config.RequestsPerMinute <= 0 {
		config.RequestsPerMinute = 60
	}
	if config.Burst <= 0 {
		config.Burst = 10
	}
	if config.CleanupInterval <= 0 {
		config.CleanupInterval = 5 * time.Minute
	}

	rl := &rateLimiter{
		buckets:     make(map[string]*tokenBucket),
		config:      config,
		stopCleanup: make(chan struct{}),
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// Stop stops the cleanup goroutine.
func (rl *rateLimiter) Stop() {
	close(rl.stopCleanup)
}

// cleanup removes old buckets periodically.
func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(rl.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			now := time.Now()
			for key, bucket := range rl.buckets {
				bucket.mu.Lock()
				if now.Sub(bucket.lastUpdate) > rl.config.CleanupInterval*2 {
					delete(rl.buckets, key)
				}
				bucket.mu.Unlock()
			}
			rl.mu.Unlock()
		case <-rl.stopCleanup:
			return
		}
	}
}

// allow checks if a request is allowed for the given key.
func (rl *rateLimiter) allow(key string) (allowed bool, remaining int, resetAt time.Time) {
	rl.mu.Lock()
	bucket, exists := rl.buckets[key]
	if !exists {
		bucket = &tokenBucket{
			tokens:     rl.config.Burst,
			lastUpdate: time.Now(),
		}
		rl.buckets[key] = bucket
	}
	rl.mu.Unlock()

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(bucket.lastUpdate)

	// Refill tokens based on elapsed time
	tokensToAdd := int(elapsed.Minutes() * float64(rl.config.RequestsPerMinute))
	if tokensToAdd > 0 {
		bucket.tokens += tokensToAdd
		if bucket.tokens > rl.config.Burst {
			bucket.tokens = rl.config.Burst
		}
		bucket.lastUpdate = now
	}

	if bucket.tokens > 0 {
		bucket.tokens--
		return true, bucket.tokens, now.Add(time.Minute / time.Duration(rl.config.RequestsPerMinute))
	}

	// Calculate when tokens will be available
	resetAt = now.Add(time.Minute / time.Duration(rl.config.RequestsPerMinute))
	return false, 0, resetAt
}

// RateLimit returns a middleware that implements rate limiting.
func RateLimit(config RateLimitConfig) gin.HandlerFunc {
	limiter := newRateLimiter(config)

	return func(c *gin.Context) {
		// Use API key ID or IP address as the key
		key := GetAPIKeyID(c)
		if key == "anonymous" {
			key = c.ClientIP()
		}

		allowed, remaining, resetAt := limiter.allow(key)

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", intToStr(config.RequestsPerMinute))
		c.Header("X-RateLimit-Remaining", intToStr(remaining))
		c.Header("X-RateLimit-Reset", resetAt.Format(time.RFC3339))

		if !allowed {
			c.Header("Retry-After", "60")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"code":        "rate_limit_exceeded",
				"retry_after": 60,
			})
			return
		}

		c.Next()
	}
}

// RateLimitByIP returns a middleware that rate limits by IP address.
func RateLimitByIP(config RateLimitConfig) gin.HandlerFunc {
	limiter := newRateLimiter(config)

	return func(c *gin.Context) {
		key := c.ClientIP()

		allowed, remaining, resetAt := limiter.allow(key)

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", intToStr(config.RequestsPerMinute))
		c.Header("X-RateLimit-Remaining", intToStr(remaining))
		c.Header("X-RateLimit-Reset", resetAt.Format(time.RFC3339))

		if !allowed {
			c.Header("Retry-After", "60")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"code":        "rate_limit_exceeded",
				"retry_after": 60,
			})
			return
		}

		c.Next()
	}
}

// intToStr converts an int to a string without importing strconv.
func intToStr(n int) string {
	if n == 0 {
		return "0"
	}

	var negative bool
	if n < 0 {
		negative = true
		n = -n
	}

	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}

	if negative {
		digits = append([]byte{'-'}, digits...)
	}

	return string(digits)
}
