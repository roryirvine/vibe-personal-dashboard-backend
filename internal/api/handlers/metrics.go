// HTTP handlers for metrics API endpoints.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/roryirvine/vibe-personal-dashboard-backend/internal/models"
)

// MetricService defines the interface that handlers depend on.
type MetricService interface {
	GetMetricNames() []string
	GetMetrics(ctx context.Context, names []string, params map[string]string) ([]models.MetricResult, error)
}

// MetricsHandler handles HTTP requests for metrics.
type MetricsHandler struct {
	service MetricService
	logger  *slog.Logger
}

// NewMetricsHandler creates a new metrics handler.
func NewMetricsHandler(service MetricService, logger *slog.Logger) *MetricsHandler {
	return &MetricsHandler{
		service: service,
		logger:  logger,
	}
}

// ListMetrics handles GET /metrics (with no ?names parameter).
func (h *MetricsHandler) ListMetrics(w http.ResponseWriter, r *http.Request) {
	names := h.service.GetMetricNames()

	results := make([]models.MetricResult, len(names))
	for i, name := range names {
		results[i] = models.MetricResult{
			Name:  name,
			Value: name,
		}
	}

	h.respondJSON(w, http.StatusOK, results)
}

// GetMetric handles GET /metrics/{name}.
func (h *MetricsHandler) GetMetric(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		h.respondError(w, http.StatusBadRequest, "metric name required")
		return
	}

	// Extract query parameters (excluding standard HTTP params)
	params := extractQueryParams(r)

	results, err := h.service.GetMetrics(r.Context(), []string{name}, params)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	if len(results) == 0 {
		h.respondError(w, http.StatusNotFound, fmt.Sprintf("metric %q not found", name))
		return
	}

	h.respondJSON(w, http.StatusOK, results)
}

// GetMetrics handles GET /metrics?names=metric1,metric2.
func (h *MetricsHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	namesParam := r.URL.Query().Get("names")

	// If no names parameter, return all metrics
	if namesParam == "" {
		h.ListMetrics(w, r)
		return
	}

	// Parse comma-separated metric names, handling whitespace
	namesRaw := strings.Split(namesParam, ",")
	names := make([]string, 0, len(namesRaw))
	for _, name := range namesRaw {
		trimmed := strings.TrimSpace(name)
		if trimmed != "" {
			names = append(names, trimmed)
		}
	}

	if len(names) == 0 {
		h.respondError(w, http.StatusBadRequest, "no valid metric names provided")
		return
	}

	// Extract query parameters (excluding 'names')
	params := extractQueryParams(r)

	results, err := h.service.GetMetrics(r.Context(), names, params)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, results)
}

// extractQueryParams extracts all query parameters except 'names'.
func extractQueryParams(r *http.Request) map[string]string {
	params := make(map[string]string)
	for key, values := range r.URL.Query() {
		if key != "names" && len(values) > 0 {
			params[key] = values[0]
		}
	}
	return params
}

// respondJSON writes a JSON response.
func (h *MetricsHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode JSON response", "error", err)
	}
}

// respondError writes a JSON error response.
func (h *MetricsHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}

// handleServiceError converts service layer errors to HTTP responses.
func (h *MetricsHandler) handleServiceError(w http.ResponseWriter, err error) {
	h.logger.Error("service error", "error", err)

	errMsg := err.Error()

	// Determine status code based on error message
	if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "unknown metric") {
		h.respondError(w, http.StatusNotFound, errMsg)
	} else if strings.Contains(errMsg, "invalid") || strings.Contains(errMsg, "required") {
		h.respondError(w, http.StatusBadRequest, errMsg)
	} else {
		h.respondError(w, http.StatusInternalServerError, "internal server error")
	}
}
