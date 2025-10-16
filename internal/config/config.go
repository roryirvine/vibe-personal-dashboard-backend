// Loads and validates TOML metric configuration files.
package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/roryirvine/vibe-personal-dashboard-backend/internal/models"
)

type Config struct {
	Metrics []models.Metric `toml:"metrics"`
}

func LoadConfig(path string) ([]models.Metric, error) {
	var config Config

	// Parse TOML file
	if _, err := toml.DecodeFile(path, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate all metrics
	if err := validateMetrics(config.Metrics); err != nil {
		return nil, err
	}

	return config.Metrics, nil
}

func validateMetrics(metrics []models.Metric) error {
	if len(metrics) == 0 {
		return fmt.Errorf("no metrics defined in config")
	}

	// Check for duplicate names
	names := make(map[string]bool)
	for _, metric := range metrics {
		if names[metric.Name] {
			return fmt.Errorf("duplicate metric name: %s", metric.Name)
		}
		names[metric.Name] = true

		// Validate each metric
		if err := metric.Validate(); err != nil {
			return fmt.Errorf("invalid metric %s: %w", metric.Name, err)
		}
	}

	return nil
}
