// Copyright 2026 The OpenTrusty Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package http

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter manages rate limiting for IPs
type RateLimiter struct {
	ips             map[string]*rate.Limiter
	mu              sync.RWMutex
	rps             rate.Limit
	burst           int
	cleanupInterval time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		ips:             make(map[string]*rate.Limiter),
		rps:             rate.Limit(rps),
		burst:           burst,
		cleanupInterval: 10 * time.Minute,
	}

	// Start background cleanup (simplified for now, avoiding goroutine leak in tests)
	go rl.cleanup()

	return rl
}

// GetLimiter returns a limiter for an IP
func (rl *RateLimiter) GetLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.ips[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.rps, rl.burst)
		rl.ips[ip] = limiter
	}

	return limiter
}

// cleanup removes old entries (simplified: just clear all every interval for now to prevent memory leak)
// In production, we'd track last access time per IP
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cleanupInterval)
	for range ticker.C {
		rl.mu.Lock()
		// Simple strategy: reset map to free memory from drive-by IPs
		// Active users will get new limiter on next request
		rl.ips = make(map[string]*rate.Limiter)
		rl.mu.Unlock()
	}
}

// RateLimitMiddleware creates a middleware for rate limiting
func RateLimitMiddleware(rl *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)

			limiter := rl.GetLimiter(ip)
			if !limiter.Allow() {
				respondError(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts IP from request (handling proxies)
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}
	// Fallback to RemoteAddr
	return r.RemoteAddr
}
