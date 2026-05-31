package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRealIPWithXRealIP(t *testing.T) {
	var gotIP string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotIP = GetRealIP(r)
		w.WriteHeader(http.StatusOK)
	})

	wrapped := RealIP(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Real-IP", "192.168.1.100")
	res := httptest.NewRecorder()

	wrapped.ServeHTTP(res, req)

	if gotIP != "192.168.1.100" {
		t.Errorf("want 192.168.1.100, got %s", gotIP)
	}
}

func TestRealIPWithXForwardedFor(t *testing.T) {
	var gotIP string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotIP = GetRealIP(r)
		w.WriteHeader(http.StatusOK)
	})

	wrapped := RealIP(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.5, 10.0.0.1, 192.168.1.1")
	res := httptest.NewRecorder()

	wrapped.ServeHTTP(res, req)

	if gotIP != "203.0.113.5" {
		t.Errorf("want 203.0.113.5, got %s", gotIP)
	}
}

func TestRealIPFallsBackToRemoteAddr(t *testing.T) {
	var gotIP string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotIP = GetRealIP(r)
		w.WriteHeader(http.StatusOK)
	})

	wrapped := RealIP(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.50:54321"
	res := httptest.NewRecorder()

	wrapped.ServeHTTP(res, req)

	if gotIP != "192.168.1.50" {
		t.Errorf("want 192.168.1.50, got %s", gotIP)
	}
}

func TestRealIPPreferXRealIPOverXForwardedFor(t *testing.T) {
	var gotIP string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotIP = GetRealIP(r)
		w.WriteHeader(http.StatusOK)
	})

	wrapped := RealIP(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Real-IP", "10.0.0.1")
	req.Header.Set("X-Forwarded-For", "203.0.113.5, 10.0.0.1")
	req.RemoteAddr = "192.168.1.1:12345"
	res := httptest.NewRecorder()

	wrapped.ServeHTTP(res, req)

	if gotIP != "10.0.0.1" {
		t.Errorf("want 10.0.0.1 (X-Real-IP), got %s", gotIP)
	}
}

func TestRealIPWithEmptyHeaders(t *testing.T) {
	var gotIP string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotIP = GetRealIP(r)
		w.WriteHeader(http.StatusOK)
	})

	wrapped := RealIP(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:8081"
	res := httptest.NewRecorder()

	wrapped.ServeHTTP(res, req)

	if gotIP != "10.0.0.1" {
		t.Errorf("want 10.0.0.1, got %s", gotIP)
	}
}

func TestRealIPWithIPv6(t *testing.T) {
	var gotIP string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotIP = GetRealIP(r)
		w.WriteHeader(http.StatusOK)
	})

	wrapped := RealIP(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Real-IP", "2001:db8::1")
	res := httptest.NewRecorder()

	wrapped.ServeHTTP(res, req)

	if gotIP != "2001:db8::1" {
		t.Errorf("want 2001:db8::1, got %s", gotIP)
	}
}

func TestRealIPSingleXForwardedFor(t *testing.T) {
	var gotIP string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotIP = GetRealIP(r)
		w.WriteHeader(http.StatusOK)
	})

	wrapped := RealIP(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.5")
	res := httptest.NewRecorder()

	wrapped.ServeHTTP(res, req)

	if gotIP != "203.0.113.5" {
		t.Errorf("want 203.0.113.5, got %s", gotIP)
	}
}

func TestGetRealIPWithNilRequest(t *testing.T) {
	if ip := GetRealIP(nil); ip != "" {
		t.Errorf("want empty, got %s", ip)
	}
}

func TestGetRealIPWithNoContext(t *testing.T) {
	req := &http.Request{}
	if ip := GetRealIP(req); ip != "" {
		t.Errorf("want empty, got %s", ip)
	}
}
