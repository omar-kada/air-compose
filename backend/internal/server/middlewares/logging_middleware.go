package middlewares

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/felixge/httpsnoop"
)

// LoggingMiddleware logs each HTTP request using slog with method, path, status, remote addr, duration and bytes.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		metrics := httpsnoop.CaptureMetrics(next, w, r)
		status := metrics.Code
		if status == 0 {
			status = http.StatusOK
		}
		dur := time.Since(start)
		slog.Debug("[HTTP] request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", status,
			"remote", r.RemoteAddr,
			"duration", dur,
			"bytes", metrics.Written,
		)
	})
}
