package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
)

const minSize = 512

var skipCompressTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"video/mp4":  true,
	"font/woff2": true,
}

func Compress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// check if response should be compressed
		gzw := &gzipResponseWriter{
			ResponseWriter: w,
			writer:         gzip.NewWriter(w),
			status:         http.StatusOK,
		}
		defer gzw.Close()

		next.ServeHTTP(gzw, r)
	})
}

type gzipResponseWriter struct {
	http.ResponseWriter
	writer *gzip.Writer
	status int
	wrote  bool
	buf    []byte
}

func (w *gzipResponseWriter) WriteHeader(code int) {
	if w.wrote {
		return
	}
	w.status = code
	w.wrote = true
}

func (w *gzipResponseWriter) Write(data []byte) (int, error) {
	if !w.wrote {
		w.WriteHeader(w.status)
	}

	// skip compression for certain status codes
	if w.status != http.StatusOK || w.status == http.StatusNoContent {
		return w.ResponseWriter.Write(data)
	}

	// skip for already compressed types
	ctype := w.Header().Get("Content-Type")
	for skip := range skipCompressTypes {
		if strings.Contains(ctype, skip) {
			return w.ResponseWriter.Write(data)
		}
	}

	// skip small responses
	if len(data) < minSize && len(w.buf) == 0 {
		w.Header().Del("Content-Encoding")
		return w.ResponseWriter.Write(data)
	}

	// compress
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Vary", "Accept-Encoding")

	if _, err := w.writer.Write(data); err != nil {
		return 0, err
	}
	if err := w.writer.Flush(); err != nil {
		return 0, err
	}
	return len(data), nil
}

func (w *gzipResponseWriter) Close() {
	if w.writer != nil {
		w.writer.Close()
	}
}
