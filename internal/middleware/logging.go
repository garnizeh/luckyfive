package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// Logging logs HTTP requests with structured logging.
func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap ResponseWriter to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			logger.Info("HTTP request",
				"method", r.Method,
				"url", r.URL.Path,
				"status", wrapped.statusCode,
				"duration", time.Since(start),
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
