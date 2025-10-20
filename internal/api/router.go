// Configures HTTP routes and middleware for the metrics API.
package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/roryirvine/vibe-personal-dashboard-backend/internal/api/handlers"
)

// NewRouter creates and configures the HTTP router with middleware.
func NewRouter(handler *handlers.MetricsHandler, logger *slog.Logger) *chi.Mux {
	r := chi.NewRouter()

	// Middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(requestLoggerMiddleware(logger))
	r.Use(middleware.Timeout(30 * time.Second))

	// Routes
	r.Get("/metrics", handler.GetMetrics)
	r.Get("/metrics/{name}", handler.GetMetric)

	return r
}

// requestLoggerMiddleware logs HTTP requests with timing information.
func requestLoggerMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Wrap response writer to capture status and size
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			start := time.Now()
			next.ServeHTTP(wrapped, r)
			duration := time.Since(start)

			logger.Info(
				"request",
				"method", r.Method,
				"path", r.RequestURI,
				"status", wrapped.statusCode,
				"duration_ms", duration.Milliseconds(),
				"request_id", middleware.GetReqID(r.Context()),
			)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
