package services

import (
	"time"

	"github.com/garnizeh/luckyfive/internal/store"
)

// SystemService handles system-level operations
type SystemService struct {
	db        *store.DB
	startTime time.Time
}

// NewSystemService creates a new system service
func NewSystemService(db *store.DB, startTime time.Time) *SystemService {
	return &SystemService{
		db:        db,
		startTime: startTime,
	}
}

// HealthStatus represents the health status of the system
type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Version   string            `json:"version"`
	Uptime    string            `json:"uptime"`
	Services  map[string]string `json:"services"`
}

// CheckHealth performs a comprehensive health check
func (s *SystemService) CheckHealth() (*HealthStatus, error) {
	services := make(map[string]string)

	// Check database connectivity
	if err := s.checkDatabaseHealth(); err != nil {
		services["database"] = "unhealthy"
	} else {
		services["database"] = "healthy"
	}

	// Check other services here (e.g., external APIs, caches, etc.)
	services["api"] = "healthy"

	status := "healthy"
	for _, serviceStatus := range services {
		if serviceStatus == "unhealthy" {
			status = "unhealthy"
			break
		}
	}

	return &HealthStatus{
		Status:    status,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0.0", // TODO: Get from build info or config
		Uptime:    time.Since(s.startTime).String(),
		Services:  services,
	}, nil
}

// checkDatabaseHealth verifies connectivity to all databases
func (s *SystemService) checkDatabaseHealth() error {
	// Test Results database
	if err := s.db.ResultsDB.Ping(); err != nil {
		return err
	}

	// Test Simulations database
	if err := s.db.SimulationsDB.Ping(); err != nil {
		return err
	}

	// Test Configs database
	if err := s.db.ConfigsDB.Ping(); err != nil {
		return err
	}

	// Test Finances database
	if err := s.db.FinancesDB.Ping(); err != nil {
		return err
	}

	return nil
}
