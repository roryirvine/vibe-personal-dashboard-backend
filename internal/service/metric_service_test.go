package service

import (
	"context"
	"errors"
	"testing"

	"github.com/roryirvine/vibe-personal-dashboard-backend/internal/models"
)

// mockRepository is a test double that implements repository.Repository
type mockRepository struct {
	singleValueResult interface{}
	singleValueErr    error
	multiRowResult    []map[string]interface{}
	multiRowErr       error
	queryCalls        int
}

func (m *mockRepository) QuerySingleValue(ctx context.Context, query string, args ...interface{}) (interface{}, error) {
	m.queryCalls++
	return m.singleValueResult, m.singleValueErr
}

func (m *mockRepository) QueryMultiRow(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	m.queryCalls++
	return m.multiRowResult, m.multiRowErr
}

func (m *mockRepository) Close() error {
	return nil
}

func TestNewMetricService(t *testing.T) {
	metrics := []models.Metric{
		{
			Name:     "active_users",
			Query:    "SELECT COUNT(*) FROM users",
			MultiRow: false,
		},
		{
			Name:     "user_signups",
			Query:    "SELECT date, count FROM signups",
			MultiRow: true,
		},
	}

	repo := &mockRepository{}
	service := NewMetricService(repo, metrics, nil)

	if service == nil {
		t.Error("NewMetricService returned nil")
	}
}

func TestMetricService_GetMetricNames(t *testing.T) {
	metrics := []models.Metric{
		{Name: "active_users", Query: "SELECT 1", MultiRow: false},
		{Name: "user_signups", Query: "SELECT 1", MultiRow: true},
		{Name: "revenue", Query: "SELECT 1", MultiRow: false},
	}

	repo := &mockRepository{}
	service := NewMetricService(repo, metrics, nil)

	names := service.GetMetricNames()

	if len(names) != 3 {
		t.Errorf("GetMetricNames() returned %d names, want 3", len(names))
	}

	// Check all metric names are present
	nameMap := make(map[string]bool)
	for _, name := range names {
		nameMap[name] = true
	}

	expectedNames := []string{"active_users", "user_signups", "revenue"}
	for _, expected := range expectedNames {
		if !nameMap[expected] {
			t.Errorf("GetMetricNames() missing %s", expected)
		}
	}
}

func TestMetricService_GetMetric_SingleValue(t *testing.T) {
	metrics := []models.Metric{
		{
			Name:     "active_users",
			Query:    "SELECT COUNT(*) FROM users",
			MultiRow: false,
		},
	}

	repo := &mockRepository{
		singleValueResult: int64(1523),
	}
	service := NewMetricService(repo, metrics, nil)

	results, err := service.GetMetric(context.Background(), "active_users", nil)

	if err != nil {
		t.Errorf("GetMetric() error = %v, want nil", err)
	}

	if len(results) != 1 {
		t.Errorf("GetMetric() returned %d results, want 1", len(results))
	}

	if results[0].Name != "active_users" {
		t.Errorf("GetMetric() Name = %s, want active_users", results[0].Name)
	}

	if results[0].Value != int64(1523) {
		t.Errorf("GetMetric() Value = %v, want 1523", results[0].Value)
	}
}

func TestMetricService_GetMetric_MultiRow(t *testing.T) {
	metrics := []models.Metric{
		{
			Name:     "signups_by_day",
			Query:    "SELECT date, count FROM signups",
			MultiRow: true,
		},
	}

	multiRowData := []map[string]interface{}{
		{"date": "2025-01-01", "count": int64(45)},
		{"date": "2025-01-02", "count": int64(52)},
	}

	repo := &mockRepository{
		multiRowResult: multiRowData,
	}
	service := NewMetricService(repo, metrics, nil)

	results, err := service.GetMetric(context.Background(), "signups_by_day", nil)

	if err != nil {
		t.Errorf("GetMetric() error = %v, want nil", err)
	}

	if len(results) != 1 {
		t.Errorf("GetMetric() returned %d results, want 1", len(results))
	}

	if results[0].Name != "signups_by_day" {
		t.Errorf("GetMetric() Name = %s, want signups_by_day", results[0].Name)
	}

	value, ok := results[0].Value.([]map[string]interface{})
	if !ok {
		t.Errorf("GetMetric() Value is not []map[string]interface{}, got %T", results[0].Value)
	}

	if len(value) != 2 {
		t.Errorf("GetMetric() returned %d rows, want 2", len(value))
	}
}

func TestMetricService_GetMetric_WithParameters(t *testing.T) {
	metrics := []models.Metric{
		{
			Name:     "signups_by_date",
			Query:    "SELECT COUNT(*) FROM signups WHERE date >= ?",
			MultiRow: false,
			Params: []models.ParamDefinition{
				{Name: "start_date", Type: models.ParamTypeString, Required: true},
			},
		},
	}

	repo := &mockRepository{
		singleValueResult: int64(150),
	}
	service := NewMetricService(repo, metrics, nil)

	params := map[string]string{
		"start_date": "2025-01-01",
	}

	results, err := service.GetMetric(context.Background(), "signups_by_date", params)

	if err != nil {
		t.Errorf("GetMetric() error = %v, want nil", err)
	}

	if len(results) != 1 {
		t.Errorf("GetMetric() returned %d results, want 1", len(results))
	}

	if results[0].Value != int64(150) {
		t.Errorf("GetMetric() Value = %v, want 150", results[0].Value)
	}
}

func TestMetricService_GetMetric_MissingRequiredParam(t *testing.T) {
	metrics := []models.Metric{
		{
			Name:     "signups_by_date",
			Query:    "SELECT COUNT(*) FROM signups WHERE date >= ?",
			MultiRow: false,
			Params: []models.ParamDefinition{
				{Name: "start_date", Type: models.ParamTypeString, Required: true},
			},
		},
	}

	repo := &mockRepository{}
	service := NewMetricService(repo, metrics, nil)

	// Call with empty params (missing required start_date)
	results, err := service.GetMetric(context.Background(), "signups_by_date", nil)

	if err == nil {
		t.Error("GetMetric() error = nil, want error for missing required param")
	}

	if len(results) != 0 {
		t.Errorf("GetMetric() returned %d results on error, want 0", len(results))
	}
}

func TestMetricService_GetMetric_InvalidParamType(t *testing.T) {
	metrics := []models.Metric{
		{
			Name:     "users_with_limit",
			Query:    "SELECT * FROM users LIMIT ?",
			MultiRow: true,
			Params: []models.ParamDefinition{
				{Name: "limit", Type: models.ParamTypeInt, Required: true},
			},
		},
	}

	repo := &mockRepository{}
	service := NewMetricService(repo, metrics, nil)

	params := map[string]string{
		"limit": "not_a_number",
	}

	results, err := service.GetMetric(context.Background(), "users_with_limit", params)

	if err == nil {
		t.Error("GetMetric() error = nil, want error for invalid int param")
	}

	if len(results) != 0 {
		t.Errorf("GetMetric() returned %d results on error, want 0", len(results))
	}
}

func TestMetricService_GetMetric_MetricNotFound(t *testing.T) {
	metrics := []models.Metric{
		{Name: "active_users", Query: "SELECT 1", MultiRow: false},
	}

	repo := &mockRepository{}
	service := NewMetricService(repo, metrics, nil)

	results, err := service.GetMetric(context.Background(), "nonexistent", nil)

	if err == nil {
		t.Error("GetMetric() error = nil, want error for nonexistent metric")
	}

	if len(results) != 0 {
		t.Errorf("GetMetric() returned %d results on error, want 0", len(results))
	}
}

func TestMetricService_GetMetrics_Concurrent(t *testing.T) {
	metrics := []models.Metric{
		{
			Name:     "active_users",
			Query:    "SELECT COUNT(*) FROM users",
			MultiRow: false,
		},
		{
			Name:     "signups",
			Query:    "SELECT COUNT(*) FROM signups",
			MultiRow: false,
		},
		{
			Name:     "revenue",
			Query:    "SELECT SUM(amount) FROM transactions",
			MultiRow: false,
		},
	}

	repo := &mockRepository{
		singleValueResult: int64(100),
	}
	service := NewMetricService(repo, metrics, nil)

	results, err := service.GetMetrics(context.Background(), []string{"active_users", "signups", "revenue"}, nil)

	if err != nil {
		t.Errorf("GetMetrics() error = %v, want nil", err)
	}

	if len(results) != 3 {
		t.Errorf("GetMetrics() returned %d results, want 3", len(results))
	}

	// Verify all metrics are present
	resultMap := make(map[string]interface{})
	for _, r := range results {
		resultMap[r.Name] = r.Value
	}

	expectedMetrics := []string{"active_users", "signups", "revenue"}
	for _, expected := range expectedMetrics {
		if _, ok := resultMap[expected]; !ok {
			t.Errorf("GetMetrics() missing result for %s", expected)
		}
	}
}

func TestMetricService_GetMetrics_ErrorHandling(t *testing.T) {
	metrics := []models.Metric{
		{
			Name:     "active_users",
			Query:    "SELECT COUNT(*) FROM users",
			MultiRow: false,
		},
		{
			Name:     "signups",
			Query:    "SELECT COUNT(*) FROM signups",
			MultiRow: false,
		},
	}

	// Create a repo that fails on the second call
	failingRepo := &testRepositoryWithFailure{
		successCount: 1,
	}

	service := NewMetricService(failingRepo, metrics, nil)

	_, err := service.GetMetrics(context.Background(), []string{"active_users", "signups"}, nil)

	// Should return error (fail-fast)
	if err == nil {
		t.Error("GetMetrics() error = nil, want error when one metric fails")
	}
}

// testRepositoryWithFailure fails on nth query
type testRepositoryWithFailure struct {
	count        int
	successCount int
}

var errQueryFailed = errors.New("query failed")

func (t *testRepositoryWithFailure) QuerySingleValue(ctx context.Context, query string, args ...interface{}) (interface{}, error) {
	t.count++
	if t.count > t.successCount {
		return nil, errQueryFailed
	}
	return int64(100), nil
}

func (t *testRepositoryWithFailure) QueryMultiRow(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	return nil, nil
}

func (t *testRepositoryWithFailure) Close() error {
	return nil
}
