package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/roryirvine/vibe-personal-dashboard-backend/internal/models"
)

// Mock service for testing
type mockMetricService struct {
	metricsFunc func(ctx context.Context, names []string, params map[string]string) ([]models.MetricResult, error)
	namesFunc   func() []string
}

func (m *mockMetricService) GetMetricNames() []string {
	if m.namesFunc != nil {
		return m.namesFunc()
	}
	return []string{}
}

func (m *mockMetricService) GetMetrics(ctx context.Context, names []string, params map[string]string) ([]models.MetricResult, error) {
	if m.metricsFunc != nil {
		return m.metricsFunc(ctx, names, params)
	}
	return nil, nil
}

func TestListMetrics(t *testing.T) {
	tests := []struct {
		name           string
		mockMetrics    []string
		expectedStatus int
	}{
		{
			name:           "list all metrics",
			mockMetrics:    []string{"active_users", "revenue_total", "user_signups"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "empty metrics list",
			mockMetrics:    []string{},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockMetricService{
				namesFunc: func() []string {
					return tt.mockMetrics
				},
			}

			handler := &MetricsHandler{
				service: svc,
				logger:  slog.New(slog.NewJSONHandler(os.Stderr, nil)),
			}

			req := httptest.NewRequest("GET", "/metrics", nil)
			w := httptest.NewRecorder()

			handler.ListMetrics(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var result []string
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if len(result) != len(tt.mockMetrics) {
				t.Errorf("expected %d metrics, got %d", len(tt.mockMetrics), len(result))
			}

			// Verify order and content
			for i, name := range result {
				if name != tt.mockMetrics[i] {
					t.Errorf("metric %d: expected %q, got %q", i, tt.mockMetrics[i], name)
				}
			}
		})
	}
}

func TestGetSingleMetric(t *testing.T) {
	tests := []struct {
		name            string
		metricName      string
		queryParams     string
		mockResult      []models.MetricResult
		mockError       error
		expectedStatus  int
		expectedHasBody bool
	}{
		{
			name:       "get single metric",
			metricName: "active_users",
			mockResult: []models.MetricResult{
				{Name: "active_users", Value: int64(1523)},
			},
			expectedStatus:  http.StatusOK,
			expectedHasBody: true,
		},
		{
			name:            "metric not found",
			metricName:      "nonexistent",
			mockError:       fmt.Errorf("metric not found"),
			expectedStatus:  http.StatusNotFound,
			expectedHasBody: true,
		},
		{
			name:           "get metric with query params",
			metricName:     "user_signups",
			queryParams:    "?start_date=2025-01-01",
			mockResult:     []models.MetricResult{{Name: "user_signups", Value: []map[string]interface{}{}}},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockMetricService{
				metricsFunc: func(ctx context.Context, names []string, params map[string]string) ([]models.MetricResult, error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return tt.mockResult, nil
				},
			}

			handler := &MetricsHandler{
				service: svc,
				logger:  slog.New(slog.NewJSONHandler(os.Stderr, nil)),
			}

			url := fmt.Sprintf("/metrics/%s%s", tt.metricName, tt.queryParams)
			req := httptest.NewRequest("GET", url, nil)

			// Set up chi URL context
			ctx := chi.NewRouteContext()
			ctx.URLParams.Add("name", tt.metricName)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

			w := httptest.NewRecorder()
			handler.GetMetric(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedHasBody {
				if w.Body.Len() == 0 {
					t.Error("expected response body, got empty")
				}
			}
		})
	}
}

func TestGetMultipleMetrics(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockResults    []models.MetricResult
		mockError      error
		expectedStatus int
		expectedCount  int
	}{
		{
			name:        "get multiple metrics",
			queryParams: "?names=active_users,revenue_total",
			mockResults: []models.MetricResult{
				{Name: "active_users", Value: int64(1523)},
				{Name: "revenue_total", Value: float64(15230.50)},
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "missing names parameter",
			queryParams:    "",
			mockResults:    []models.MetricResult{},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "with whitespace in names",
			queryParams: "?names=active_users,%20revenue_total",
			mockResults: []models.MetricResult{
				{Name: "active_users", Value: int64(1523)},
				{Name: "revenue_total", Value: float64(15230.50)},
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "query error",
			queryParams:    "?names=active_users,revenue_total",
			mockError:      fmt.Errorf("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockMetricService{
				metricsFunc: func(ctx context.Context, names []string, params map[string]string) ([]models.MetricResult, error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return tt.mockResults, nil
				},
			}

			handler := &MetricsHandler{
				service: svc,
				logger:  slog.New(slog.NewJSONHandler(os.Stderr, nil)),
			}

			url := "/metrics" + tt.queryParams
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			handler.GetMetrics(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK && tt.expectedCount > 0 {
				var result []models.MetricResult
				if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if len(result) != tt.expectedCount {
					t.Errorf("expected %d results, got %d", tt.expectedCount, len(result))
				}
			}
		})
	}
}

func TestErrorResponse(t *testing.T) {
	handler := &MetricsHandler{
		service: nil,
		logger:  slog.New(slog.NewJSONHandler(os.Stderr, nil)),
	}

	w := httptest.NewRecorder()
	handler.respondError(w, http.StatusBadRequest, "invalid input")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}

	if result["error"] == nil {
		t.Error("expected error field in response")
	}
}
