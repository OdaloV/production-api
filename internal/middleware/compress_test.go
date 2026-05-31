package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// gzipReader decompresses response body for testing
func gzipReader(t *testing.T, data []byte) string {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer r.Close()
	result, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read gzip data: %v", err)
	}
	return string(result)
}

func TestCompressWithGzipSupport(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(strings.Repeat("hello world ", 100)))
	})

	compress := Compress(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	res := httptest.NewRecorder()

	compress.ServeHTTP(res, req)

	if res.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("want 'gzip', got '%s'", res.Header().Get("Content-Encoding"))
	}

	if res.Header().Get("Vary") != "Accept-Encoding" {
		t.Errorf("want 'Accept-Encoding', got '%s'", res.Header().Get("Vary"))
	}

	// decompress and verify
	decompressed := gzipReader(t, res.Body.Bytes())
	if !strings.Contains(decompressed, "hello world") {
		t.Error("decompressed content missing expected text")
	}
}

func TestCompressSkipsWhenNoGzipSupport(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("hello"))
	})

	compress := Compress(handler)

	req := httptest.NewRequest("GET", "/", nil)
	// no accept-encoding header
	res := httptest.NewRecorder()

	compress.ServeHTTP(res, req)

	if res.Header().Get("Content-Encoding") == "gzip" {
		t.Error("should not have gzip encoding when not accepted")
	}

	if res.Body.String() != "hello" {
		t.Errorf("want 'hello', got '%s'", res.Body.String())
	}
}

func TestCompressSkipsSmallResponses(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("small"))
	})

	compress := Compress(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	res := httptest.NewRecorder()

	compress.ServeHTTP(res, req)

	// small response (<512 bytes) should not be compressed
	if res.Header().Get("Content-Encoding") == "gzip" {
		t.Error("small response should not be compressed")
	}
}

func TestCompressSkipsAlreadyCompressedTypes(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		w.Write([]byte(strings.Repeat("x", 1000)))
	})

	compress := Compress(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	res := httptest.NewRecorder()

	compress.ServeHTTP(res, req)

	// jpeg already compressed, should skip
	if res.Header().Get("Content-Encoding") == "gzip" {
		t.Error("jpeg should not be compressed again")
	}
}

func TestCompressCompressesJSON(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data := strings.Repeat(`{"key":"value"}`, 50)
		w.Write([]byte(data))
	})

	compress := Compress(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	res := httptest.NewRecorder()

	compress.ServeHTTP(res, req)

	if res.Header().Get("Content-Encoding") != "gzip" {
		t.Error("json should be compressed")
	}
}

func TestCompressSkipsNoContent(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	compress := Compress(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	res := httptest.NewRecorder()

	compress.ServeHTTP(res, req)

	if res.Header().Get("Content-Encoding") == "gzip" {
		t.Error("204 no content should not have encoding header")
	}
}

func TestCompressSetsVaryHeader(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(strings.Repeat("x", 1000)))
	})

	compress := Compress(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	res := httptest.NewRecorder()

	compress.ServeHTTP(res, req)

	if res.Header().Get("Vary") != "Accept-Encoding" {
		t.Errorf("want 'Accept-Encoding', got '%s'", res.Header().Get("Vary"))
	}
}

func TestCompressPreservesOtherHeaders(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom", "myvalue")
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(strings.Repeat("hello ", 200)))
	})

	compress := Compress(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	res := httptest.NewRecorder()

	compress.ServeHTTP(res, req)

	if res.Header().Get("X-Custom") != "myvalue" {
		t.Errorf("want 'myvalue', got '%s'", res.Header().Get("X-Custom"))
	}
}
