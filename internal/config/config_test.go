package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		// Create a temporary config file
		content := `
[[metrics]]
name = "test_metric"
query = "SELECT COUNT(*) FROM users"
multi_row = false

[[metrics]]
name = "users_list"
query = "SELECT id, name FROM users WHERE created > ?"
multi_row = true

[[metrics.params]]
name = "start_date"
type = "string"
required = true
`
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "metrics.toml")
		if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test config: %v", err)
		}

		// Load config
		metrics, err := LoadConfig(configPath)
		if err != nil {
			t.Fatalf("LoadConfig() error = %v", err)
		}

		// Verify we got 2 metrics
		if len(metrics) != 2 {
			t.Errorf("got %d metrics, want 2", len(metrics))
		}

		// Verify first metric
		if metrics[0].Name != "test_metric" {
			t.Errorf("first metric name = %s, want test_metric", metrics[0].Name)
		}
		if metrics[0].MultiRow {
			t.Error("first metric should be single-row")
		}

		// Verify second metric
		if metrics[1].Name != "users_list" {
			t.Errorf("second metric name = %s, want users_list", metrics[1].Name)
		}
		if !metrics[1].MultiRow {
			t.Error("second metric should be multi-row")
		}
		if len(metrics[1].Params) != 1 {
			t.Errorf("second metric has %d params, want 1", len(metrics[1].Params))
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := LoadConfig("/nonexistent/path.toml")
		if err == nil {
			t.Error("expected error for nonexistent file")
		}
	})

	t.Run("invalid toml", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "bad.toml")
		if err := os.WriteFile(configPath, []byte("invalid { toml"), 0644); err != nil {
			t.Fatalf("failed to write test config: %v", err)
		}

		_, err := LoadConfig(configPath)
		if err == nil {
			t.Error("expected error for invalid TOML")
		}
	})

	t.Run("duplicate metric names", func(t *testing.T) {
		content := `
[[metrics]]
name = "duplicate"
query = "SELECT 1"

[[metrics]]
name = "duplicate"
query = "SELECT 2"
`
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "dup.toml")
		if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test config: %v", err)
		}

		_, err := LoadConfig(configPath)
		if err == nil {
			t.Error("expected error for duplicate metric names")
		}
	})

	t.Run("empty metrics array", func(t *testing.T) {
		content := `# Valid TOML but no metrics defined
`
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "empty.toml")
		if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test config: %v", err)
		}

		_, err := LoadConfig(configPath)
		if err == nil {
			t.Error("expected error for config with no metrics")
		}
	})
}
