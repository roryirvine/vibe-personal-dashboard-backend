// Executes metric queries with parameter validation and concurrent execution.
package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/roryirvine/vibe-personal-dashboard-backend/internal/models"
	"github.com/roryirvine/vibe-personal-dashboard-backend/internal/repository"
	"golang.org/x/sync/errgroup"
)

// MetricService orchestrates metric queries between HTTP handlers and the repository.
type MetricService struct {
	repo    repository.Repository
	metrics map[string]models.Metric
	logger  *slog.Logger
}

// NewMetricService creates a new MetricService with the given repository and metrics.
// It builds a map for efficient O(1) metric lookup by name.
func NewMetricService(repo repository.Repository, metricsList []models.Metric, logger *slog.Logger) *MetricService {
	metricsMap := make(map[string]models.Metric)
	for _, m := range metricsList {
		metricsMap[m.Name] = m
	}

	return &MetricService{
		repo:    repo,
		metrics: metricsMap,
		logger:  logger,
	}
}

// GetMetricNames returns a slice of all available metric names.
func (ms *MetricService) GetMetricNames() []string {
	names := make([]string, 0, len(ms.metrics))
	for name := range ms.metrics {
		names = append(names, name)
	}
	return names
}

// GetMetric executes a single metric query with optional parameters.
// Returns a slice containing one MetricResult, or an error.
func (ms *MetricService) GetMetric(ctx context.Context, name string, params map[string]string) ([]models.MetricResult, error) {
	metric, exists := ms.metrics[name]
	if !exists {
		return nil, fmt.Errorf("metric %q not found", name)
	}

	// Prepare and validate parameters
	args, err := ms.prepareParams(metric, params)
	if err != nil {
		return nil, err
	}

	var value interface{}

	if metric.MultiRow {
		// Execute multi-row query
		rows, err := ms.repo.QueryMultiRow(ctx, metric.Query, args...)
		if err != nil {
			return nil, fmt.Errorf("metric %q failed: %w", metric.Name, err)
		}
		value = rows
	} else {
		// Execute single-value query
		result, err := ms.repo.QuerySingleValue(ctx, metric.Query, args...)
		if err != nil {
			return nil, fmt.Errorf("metric %q failed: %w", metric.Name, err)
		}
		value = result
	}

	return []models.MetricResult{
		{
			Name:  metric.Name,
			Value: value,
		},
	}, nil
}

// GetMetrics executes multiple metrics concurrently using errgroup.
// If any metric fails, returns error immediately (fail-fast).
// Returns a slice of MetricResult, one per requested metric (if successful).
func (ms *MetricService) GetMetrics(ctx context.Context, names []string, params map[string]string) ([]models.MetricResult, error) {
	results := make([]models.MetricResult, len(names))
	eg, egCtx := errgroup.WithContext(ctx)

	for i, name := range names {
		// Capture loop variables for goroutine
		i, name := i, name

		eg.Go(func() error {
			metricResults, err := ms.GetMetric(egCtx, name, params)
			if err != nil {
				return err
			}
			if len(metricResults) > 0 {
				results[i] = metricResults[0]
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return results, nil
}

// prepareParams validates required parameters and converts string values to typed values.
// Returns a slice of interface{} that can be passed directly to repository query methods.
func (ms *MetricService) prepareParams(metric models.Metric, params map[string]string) ([]interface{}, error) {
	if len(metric.Params) == 0 {
		return nil, nil
	}

	args := make([]interface{}, len(metric.Params))

	for i, paramDef := range metric.Params {
		value, exists := params[paramDef.Name]

		// Check if required parameter is present
		if paramDef.Required && !exists {
			return nil, fmt.Errorf("metric %q: required parameter %q is missing", metric.Name, paramDef.Name)
		}

		// If optional and missing, use empty string (caller will decide if this is valid)
		if !exists {
			value = ""
		}

		// Convert string value to typed value
		convertedValue, err := convertParamValue(value, paramDef.Type)
		if err != nil {
			return nil, fmt.Errorf("metric %q: parameter %q: %w", metric.Name, paramDef.Name, err)
		}

		args[i] = convertedValue
	}

	return args, nil
}
