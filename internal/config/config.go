package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Worker   WorkerConfig
	LogLevel string
}

type ServerConfig struct {
	Host string
	Port int
}

type DatabaseConfig struct {
	ResultsPath     string
	SimulationsPath string
	ConfigsPath     string
	FinancesPath    string
	SweepsPath      string
}

type WorkerConfig struct {
	Concurrency  int
	PollInterval time.Duration
}

// getEnv returns the value for key or defaultVal if not present.
func getEnv(key, defaultVal string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return defaultVal
}

// Load reads configuration from environment variables and .env file, returns a Config.
func Load(envFilePath string) (*Config, error) {
	// Load .env file if path is provided and file exists
	if envFilePath != "" {
		_ = godotenv.Load(envFilePath)
	}

	portStr := getEnv("SERVER_PORT", "8080")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}

	concurrencyStr := getEnv("WORKER_CONCURRENCY", "4")
	conc, err := strconv.Atoi(concurrencyStr)
	if err != nil {
		return nil, err
	}

	pollIntervalStr := getEnv("WORKER_POLL_INTERVAL_SECONDS", "5")
	pollIntervalSec, err := strconv.Atoi(pollIntervalStr)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "localhost"),
			Port: port,
		},
		Database: DatabaseConfig{
			ResultsPath:     getEnv("DB_RESULTS_PATH", "data/results.db"),
			SimulationsPath: getEnv("DB_SIMULATIONS_PATH", "data/simulations.db"),
			ConfigsPath:     getEnv("DB_CONFIGS_PATH", "data/configs.db"),
			FinancesPath:    getEnv("DB_FINANCES_PATH", "data/finances.db"),
			SweepsPath:      getEnv("DB_SWEEPS_PATH", "data/sweeps.db"),
		},
		Worker: WorkerConfig{
			Concurrency:  conc,
			PollInterval: time.Duration(pollIntervalSec) * time.Second,
		},
		LogLevel: strings.ToUpper(getEnv("LOG_LEVEL", "INFO")),
	}

	return cfg, nil
}
