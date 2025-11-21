package config

import (
	"os"
	"strconv"
	"strings"
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
}

type WorkerConfig struct {
	Concurrency int
}

// getEnv returns the value for key or defaultVal if not present.
func getEnv(key, defaultVal string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return defaultVal
}

// Load reads configuration from environment variables and returns a Config.
func Load() (*Config, error) {
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
		},
		Worker: WorkerConfig{
			Concurrency: conc,
		},
		LogLevel: strings.ToUpper(getEnv("LOG_LEVEL", "INFO")),
	}

	return cfg, nil
}
