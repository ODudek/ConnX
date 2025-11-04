package ratelimit

import (
	"sync"
	"time"
)

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	mu            sync.Mutex
	rate          int           // requests per second
	burst         int           // maximum burst size
	tokens        float64       // current tokens
	lastRefill    time.Time
	perIPLimiters map[string]*ipLimiter
	enabled       bool
}

// ipLimiter tracks rate limiting for a single IP
type ipLimiter struct {
	tokens     float64
	lastRefill time.Time
}

// Config holds rate limiter configuration
type Config struct {
	Enabled        bool
	RequestsPerSec int
	Burst          int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config Config) *RateLimiter {
	if config.RequestsPerSec <= 0 {
		config.RequestsPerSec = 100
	}
	if config.Burst <= 0 {
		config.Burst = config.RequestsPerSec * 2
	}

	return &RateLimiter{
		rate:          config.RequestsPerSec,
		burst:         config.Burst,
		tokens:        float64(config.Burst),
		lastRefill:    time.Now(),
		perIPLimiters: make(map[string]*ipLimiter),
		enabled:       config.Enabled,
	}
}

// AllowGlobal checks if a request is allowed globally
func (rl *RateLimiter) AllowGlobal() bool {
	if !rl.enabled {
		return true
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refill()

	if rl.tokens >= 1.0 {
		rl.tokens -= 1.0
		return true
	}

	return false
}

// AllowIP checks if a request from a specific IP is allowed
func (rl *RateLimiter) AllowIP(ip string) bool {
	if !rl.enabled {
		return true
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.perIPLimiters[ip]
	if !exists {
		limiter = &ipLimiter{
			tokens:     float64(rl.burst),
			lastRefill: time.Now(),
		}
		rl.perIPLimiters[ip] = limiter
	}

	// Refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(limiter.lastRefill).Seconds()
	limiter.tokens += elapsed * float64(rl.rate)

	if limiter.tokens > float64(rl.burst) {
		limiter.tokens = float64(rl.burst)
	}

	limiter.lastRefill = now

	// Check if request can be allowed
	if limiter.tokens >= 1.0 {
		limiter.tokens -= 1.0
		return true
	}

	return false
}

// refill refills tokens for global rate limiter
func (rl *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()

	rl.tokens += elapsed * float64(rl.rate)
	if rl.tokens > float64(rl.burst) {
		rl.tokens = float64(rl.burst)
	}

	rl.lastRefill = now
}

// Cleanup removes old IP limiters to prevent memory leaks
func (rl *RateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip, limiter := range rl.perIPLimiters {
		// Remove limiters that haven't been used in the last 5 minutes
		if now.Sub(limiter.lastRefill) > 5*time.Minute {
			delete(rl.perIPLimiters, ip)
		}
	}
}

// StartCleanup starts a background goroutine to periodically cleanup old limiters
func (rl *RateLimiter) StartCleanup() {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			rl.Cleanup()
		}
	}()
}

// GetStats returns current rate limiter statistics
func (rl *RateLimiter) GetStats() (tokens float64, trackedIPs int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	return rl.tokens, len(rl.perIPLimiters)
}
