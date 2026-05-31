package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTimeoutCompletesNormally(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	timeout := Timeout(100 * time.Millisecond)
	wrapped := timeout(handler)

	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()

	wrapped.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("want 200, got %d", res.Code)
	}
}

func TestTimeoutExceeds(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	timeout := Timeout(10 * time.Millisecond)
	wrapped := timeout(handler)

	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()

	wrapped.ServeHTTP(res, req)

	if res.Code != http.StatusGatewayTimeout {
		t.Errorf("want 504, got %d", res.Code)
	}
}

func TestTimeoutConfigurable(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	short := Timeout(20 * time.Millisecond)
	wrappedShort := short(handler)

	req := httptest.NewRequest("GET", "/", nil)
	resShort := httptest.NewRecorder()
	wrappedShort.ServeHTTP(resShort, req)

	if resShort.Code != http.StatusGatewayTimeout {
		t.Errorf("short timeout: want 504, got %d", resShort.Code)
	}

	long := Timeout(200 * time.Millisecond)
	wrappedLong := long(handler)

	resLong := httptest.NewRecorder()
	wrappedLong.ServeHTTP(resLong, req)

	if resLong.Code != http.StatusOK {
		t.Errorf("long timeout: want 200, got %d", resLong.Code)
	}
}

func TestTimeoutRespectsExistingDeadline(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	parentCtx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	timeout := Timeout(100 * time.Millisecond)
	wrapped := timeout(handler)

	req := httptest.NewRequest("GET", "/", nil).WithContext(parentCtx)
	res := httptest.NewRecorder()

	wrapped.ServeHTTP(res, req)

	// should still work, just respects parent deadline
	if res.Code != http.StatusOK {
		t.Errorf("want 200, got %d", res.Code)
	}
}

func TestTimeoutZeroDuration(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	timeout := Timeout(0)
	wrapped := timeout(handler)

	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()

	wrapped.ServeHTTP(res, req)

	if res.Code != http.StatusGatewayTimeout {
		t.Errorf("want 504, got %d", res.Code)
	}
}
