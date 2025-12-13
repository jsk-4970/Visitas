package middleware

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiterConfig holds configuration for rate limiting
type RateLimiterConfig struct {
	// RequestsPerHour is the maximum number of requests per hour per client
	RequestsPerHour int
	// BurstSize is the maximum burst size allowed
	BurstSize int
	// CleanupInterval is how often to clean up old entries
	CleanupInterval time.Duration
	// EntryTTL is how long to keep entries after last access
	EntryTTL time.Duration
}

// DefaultRateLimiterConfig returns the default configuration
// Based on OpenAPI spec: 1000 requests/hour
func DefaultRateLimiterConfig() *RateLimiterConfig {
	return &RateLimiterConfig{
		RequestsPerHour: 1000,
		BurstSize:       50, // Allow burst of 50 requests
		CleanupInterval: 10 * time.Minute,
		EntryTTL:        1 * time.Hour,
	}
}

// clientLimiter holds rate limiter and last access time for a client
type clientLimiter struct {
	limiter    *rate.Limiter
	lastAccess time.Time
}

// RateLimiterMiddleware implements per-client rate limiting
type RateLimiterMiddleware struct {
	config   *RateLimiterConfig
	clients  map[string]*clientLimiter
	mu       sync.RWMutex
	stopChan chan struct{}
}

// NewRateLimiterMiddleware creates a new rate limiter middleware
func NewRateLimiterMiddleware(config *RateLimiterConfig) *RateLimiterMiddleware {
	if config == nil {
		config = DefaultRateLimiterConfig()
	}

	rl := &RateLimiterMiddleware{
		config:   config,
		clients:  make(map[string]*clientLimiter),
		stopChan: make(chan struct{}),
	}

	// Start cleanup goroutine
	go rl.cleanupRoutine()

	return rl
}

// Stop stops the cleanup goroutine
func (rl *RateLimiterMiddleware) Stop() {
	close(rl.stopChan)
}

// cleanupRoutine periodically removes stale entries
func (rl *RateLimiterMiddleware) cleanupRoutine() {
	ticker := time.NewTicker(rl.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-rl.stopChan:
			return
		}
	}
}

// cleanup removes entries that haven't been accessed recently
func (rl *RateLimiterMiddleware) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-rl.config.EntryTTL)
	for key, client := range rl.clients {
		if client.lastAccess.Before(cutoff) {
			delete(rl.clients, key)
		}
	}
}

// getLimiter gets or creates a rate limiter for a client
func (rl *RateLimiterMiddleware) getLimiter(clientID string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if client, exists := rl.clients[clientID]; exists {
		client.lastAccess = time.Now()
		return client.limiter
	}

	// Calculate rate: requests per second from requests per hour
	ratePerSecond := rate.Limit(float64(rl.config.RequestsPerHour) / 3600.0)
	limiter := rate.NewLimiter(ratePerSecond, rl.config.BurstSize)

	rl.clients[clientID] = &clientLimiter{
		limiter:    limiter,
		lastAccess: time.Now(),
	}

	return limiter
}

// getClientID extracts a unique identifier for the client
// Priority: User ID (from auth) > X-Forwarded-For > X-Real-IP > RemoteAddr
func (rl *RateLimiterMiddleware) getClientID(r *http.Request) string {
	// First, try to use authenticated user ID
	if userID, ok := GetUserIDFromContext(r.Context()); ok && userID != "" {
		return "user:" + userID
	}

	// Fall back to IP address
	return "ip:" + getClientIP(r)
}

// Limit is the middleware handler that enforces rate limiting
func (rl *RateLimiterMiddleware) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientID := rl.getClientID(r)
		limiter := rl.getLimiter(clientID)

		if !limiter.Allow() {
			// Rate limit exceeded
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "60") // Suggest retry after 60 seconds
			w.Header().Set("X-RateLimit-Limit", formatInt(rl.config.RequestsPerHour))
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"rate_limit_exceeded","message":"Too many requests. Please try again later."}`))
			return
		}

		// Add rate limit headers to response
		w.Header().Set("X-RateLimit-Limit", formatInt(rl.config.RequestsPerHour))

		next.ServeHTTP(w, r)
	})
}

// formatInt converts an int to string without importing strconv
func formatInt(n int) string {
	if n == 0 {
		return "0"
	}

	negative := n < 0
	if negative {
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
