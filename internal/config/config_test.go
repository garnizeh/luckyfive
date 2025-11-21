package config_test

import (
	"os"
	"testing"

	"github.com/garnizeh/luckyfive/internal/config"
)

func TestLoadDefaults(t *testing.T) {
	// Clear relevant env vars
	keys := []string{"SERVER_HOST", "SERVER_PORT", "DB_RESULTS_PATH", "DB_SIMULATIONS_PATH", "DB_CONFIGS_PATH", "DB_FINANCES_PATH", "LOG_LEVEL", "WORKER_CONCURRENCY"}
	saved := map[string]string{}
	for _, k := range keys {
		if v, ok := os.LookupEnv(k); ok {
			saved[k] = v
			os.Unsetenv(k)
		}
	}
	// restore at end
	defer func() {
		for k, v := range saved {
			os.Setenv(k, v)
		}
	}()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Server.Host != "localhost" {
		t.Errorf("expected default host localhost, got %s", cfg.Server.Host)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Database.ResultsPath != "data/results.db" {
		t.Errorf("expected default results path data/results.db, got %s", cfg.Database.ResultsPath)
	}
	if cfg.LogLevel != "INFO" {
		t.Errorf("expected default log level INFO, got %s", cfg.LogLevel)
	}
	if cfg.Worker.Concurrency != 4 {
		t.Errorf("expected default worker concurrency 4, got %d", cfg.Worker.Concurrency)
	}
}

func TestLoadEnvOverrides(t *testing.T) {
	os.Setenv("SERVER_HOST", "127.0.0.1")
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("DB_RESULTS_PATH", "/tmp/results.db")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("WORKER_CONCURRENCY", "10")
	defer func() {
		os.Unsetenv("SERVER_HOST")
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("DB_RESULTS_PATH")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("WORKER_CONCURRENCY")
	}()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("expected host 127.0.0.1, got %s", cfg.Server.Host)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Server.Port)
	}
	if cfg.Database.ResultsPath != "/tmp/results.db" {
		t.Errorf("expected results path /tmp/results.db, got %s", cfg.Database.ResultsPath)
	}
	if cfg.LogLevel != "DEBUG" && cfg.LogLevel != "debug" {
		// Load uppercases the level
		if cfg.LogLevel != "DEBUG" {
			t.Errorf("expected log level DEBUG, got %s", cfg.LogLevel)
		}
	}
	if cfg.Worker.Concurrency != 10 {
		t.Errorf("expected worker concurrency 10, got %d", cfg.Worker.Concurrency)
	}
}

func TestLoadInvalidPort(t *testing.T) {
	os.Setenv("SERVER_PORT", "notanint")
	defer os.Unsetenv("SERVER_PORT")

	_, err := config.Load()
	if err == nil {
		t.Fatalf("expected error when SERVER_PORT is invalid, got nil")
	}
}

func TestLoadInvalidConcurrency(t *testing.T) {
	os.Setenv("WORKER_CONCURRENCY", "notanint")
	defer os.Unsetenv("WORKER_CONCURRENCY")

	_, err := config.Load()
	if err == nil {
		t.Fatalf("expected error when WORKER_CONCURRENCY is invalid, got nil")
	}
}
