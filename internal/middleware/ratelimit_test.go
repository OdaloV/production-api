package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimitWithinLimit(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rateLimit := RateLimit(3, time.Second)
	wrapped := rateLimit(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	for i := 0; i < 3; i++ {
		res := httptest.NewRecorder()
		wrapped.ServeHTTP(res, req)
		if res.Code != http.StatusOK {
			t.Errorf("request %d: want 200, got %d", i+1, res.Code)
		}
	}
}

func TestRateLimitExceedsLimit(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rateLimit := RateLimit(2, time.Second)
	wrapped := rateLimit(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	// first 2 requests should succeed
	for i := 0; i < 2; i++ {
		res := httptest.NewRecorder()
		wrapped.ServeHTTP(res, req)
		if res.Code != http.StatusOK {
			t.Errorf("request %d: want 200, got %d", i+1, res.Code)
		}
	}

	// third request should fail with 429
	res := httptest.NewRecorder()
	wrapped.ServeHTTP(res, req)
	if res.Code != http.StatusTooManyRequests {
		t.Errorf("want 429, got %d", res.Code)
	}

	// check retry-after header
	if res.Header().Get("Retry-After") == "" {
		t.Error("expected Retry-After header")
	}
}

func TestRateLimitDifferentIPs(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rateLimit := RateLimit(2, time.Second)
	wrapped := rateLimit(handler)

	ips := []string{"192.168.1.1", "192.168.1.2", "192.168.1.3"}

	for _, ip := range ips {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = ip + ":12345"

		// each ip can make 2 requests
		for i := 0; i < 2; i++ {
			res := httptest.NewRecorder()
			wrapped.ServeHTTP(res, req)
			if res.Code != http.StatusOK {
				t.Errorf("ip %s request %d: want 200, got %d", ip, i+1, res.Code)
			}
		}

		// third request should fail
		res := httptest.NewRecorder()
		wrapped.ServeHTTP(res, req)
		if res.Code != http.StatusTooManyRequests {
			t.Errorf("ip %s third request: want 429, got %d", ip, res.Code)
		}
	}
}

func TestRateLimitWindowExpiration(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rateLimit := RateLimit(1, 500*time.Millisecond)
	wrapped := rateLimit(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	// first request succeeds
	res1 := httptest.NewRecorder()
	wrapped.ServeHTTP(res1, req)
	if res1.Code != http.StatusOK {
		t.Errorf("first request: want 200, got %d", res1.Code)
	}

	// second request fails (within window)
	res2 := httptest.NewRecorder()
	wrapped.ServeHTTP(res2, req)
	if res2.Code != http.StatusTooManyRequests {
		t.Errorf("second request: want 429, got %d", res2.Code)
	}

	// wait for window to expire
	time.Sleep(600 * time.Millisecond)

	// third request should succeed again
	res3 := httptest.NewRecorder()
	wrapped.ServeHTTP(res3, req)
	if res3.Code != http.StatusOK {
		t.Errorf("third request after window: want 200, got %d", res3.Code)
	}
}

func TestRateLimitWithRealIP(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rateLimit := RateLimit(1, time.Second)
	withRealIP := RealIP(rateLimit(handler))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Real-IP", "10.0.0.1")
	req.RemoteAddr = "192.168.1.1:12345"

	// first request succeeds
	res1 := httptest.NewRecorder()
	withRealIP.ServeHTTP(res1, req)
	if res1.Code != http.StatusOK {
		t.Errorf("first: want 200, got %d", res1.Code)
	}

	// second request fails (same x-real-ip)
	res2 := httptest.NewRecorder()
	withRealIP.ServeHTTP(res2, req)
	if res2.Code != http.StatusTooManyRequests {
		t.Errorf("second: want 429, got %d", res2.Code)
	}

	// change x-real-ip
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.Header.Set("X-Real-IP", "10.0.0.2")
	req2.RemoteAddr = "192.168.1.1:12345"

	res3 := httptest.NewRecorder()
	withRealIP.ServeHTTP(res3, req2)
	if res3.Code != http.StatusOK {
		t.Errorf("different ip: want 200, got %d", res3.Code)
	}
}

func TestRateLimitRetryAfterHeader(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rateLimit := RateLimit(1, 2*time.Second)
	wrapped := rateLimit(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	// first succeeds
	res1 := httptest.NewRecorder()
	wrapped.ServeHTTP(res1, req)

	// second fails with retry-after
	res2 := httptest.NewRecorder()
	wrapped.ServeHTTP(res2, req)

	retryAfter := res2.Header().Get("Retry-After")
	if retryAfter == "" {
		t.Error("expected Retry-After header")
	}
}
