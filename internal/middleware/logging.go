package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// MetricsRecorder defines the interface for recording metrics
type MetricsRecorder interface {
	RecordHTTPRequest(method string, statusCode int, duration time.Duration)
}

// Logging logs HTTP requests with structured logging and metrics recording.
func Logging(logger *slog.Logger, metricsRecorder MetricsRecorder) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap ResponseWriter to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)

			// Record metrics if recorder is provided
			if metricsRecorder != nil {
				metricsRecorder.RecordHTTPRequest(r.Method, wrapped.statusCode, duration)
			}

			logger.Info("HTTP request",
				"method", r.Method,
				"url", r.URL.Path,
				"status", wrapped.statusCode,
				"duration", duration,
				"user_agent", r.UserAgent(),
				"remote_addr", r.RemoteAddr,
			)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
