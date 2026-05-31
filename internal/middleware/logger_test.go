package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoggerLogsRequest(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := Logger(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()

	// should not panic
	wrapped.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("want 200, got %d", res.Code)
	}
}

func TestLoggerCapturesStatusCode(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	wrapped := Logger(handler)

	req := httptest.NewRequest("GET", "/missing", nil)
	res := httptest.NewRecorder()

	wrapped.ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Errorf("want 404, got %d", res.Code)
	}
}

func TestLoggerWithRequestID(t *testing.T) {
	var called bool

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := GetRequestID(r.Context())
		if id == "" {
			t.Error("request id should be present")
		}
		called = true
		w.WriteHeader(http.StatusOK)
	})

	// add request id middleware first
	withID := RequestID(handler)
	wrapped := Logger(withID)

	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()

	wrapped.ServeHTTP(res, req)

	if !called {
		t.Error("handler was not called")
	}
}

func TestLoggerWithRealIP(t *testing.T) {
	var called bool

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := GetRealIP(r)
		if ip == "" {
			t.Error("real ip should be present")
		}
		called = true
		w.WriteHeader(http.StatusOK)
	})

	// add real ip middleware first
	withIP := RealIP(handler)
	wrapped := Logger(withIP)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Real-IP", "192.168.1.100")
	res := httptest.NewRecorder()

	wrapped.ServeHTTP(res, req)

	if !called {
		t.Error("handler was not called")
	}
}

func TestLoggerWithDifferentMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			wrapped := Logger(handler)

			req := httptest.NewRequest(method, "/", nil)
			res := httptest.NewRecorder()

			wrapped.ServeHTTP(res, req)

			if res.Code != http.StatusOK {
				t.Errorf("want 200, got %d", res.Code)
			}
		})
	}
}

func TestLoggerWithErrorStatus(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	wrapped := Logger(handler)

	req := httptest.NewRequest("GET", "/error", nil)
	res := httptest.NewRecorder()

	wrapped.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Errorf("want 500, got %d", res.Code)
	}
}
