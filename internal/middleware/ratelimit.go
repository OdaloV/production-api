package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type client struct {
	count     int
	expiresAt time.Time
}

type rateLimiter struct {
	mu       sync.RWMutex
	clients  map[string]*client
	limit    int
	window   time.Duration
	stopChan chan struct{}
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		clients:  make(map[string]*client),
		limit:    limit,
		window:   window,
		stopChan: make(chan struct{}),
	}
	go rl.cleanup()
	return rl
}

func (rl *rateLimiter) allow(ip string) (bool, time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	c, exists := rl.clients[ip]

	if !exists || now.After(c.expiresAt) {
		rl.clients[ip] = &client{
			count:     1,
			expiresAt: now.Add(rl.window),
		}
		return true, 0
	}

	if c.count >= rl.limit {
		retryAfter := time.Until(c.expiresAt)
		if retryAfter < 0 {
			retryAfter = 0
		}
		return false, retryAfter
	}

	c.count++
	return true, 0
}

func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			now := time.Now()
			for ip, c := range rl.clients {
				if now.After(c.expiresAt) {
					delete(rl.clients, ip)
				}
			}
			rl.mu.Unlock()
		case <-rl.stopChan:
			return
		}
	}
}

func (rl *rateLimiter) stop() {
	close(rl.stopChan)
}

func RateLimit(limit int, window time.Duration) func(http.Handler) http.Handler {
	limiter := newRateLimiter(limit, window)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := GetRealIP(r)
			if ip == "" {
				parts := strings.Split(r.RemoteAddr, ":")
				if len(parts) > 0 {
					ip = parts[0]
				}
			}

			allowed, retryAfter := limiter.allow(ip)
			if !allowed {
				seconds := int(retryAfter.Seconds())
				if seconds < 1 {
					seconds = 1
				}
				w.Header().Set("Retry-After", strconv.Itoa(seconds))
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte("Too Many Requests\n"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
