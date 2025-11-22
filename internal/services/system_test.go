package services

import (
	"testing"
	"time"	
)

func TestNewSystemService(t *testing.T) {
	db := createTestDB(t)

	startTime := time.Now()
	service := NewSystemService(db, startTime)

	if service == nil {
		t.Fatal("NewSystemService returned nil")
	}
	if service.db != db {
		t.Error("Expected db to match")
	}
	if !service.startTime.Equal(startTime) {
		t.Errorf("Expected startTime to be %v, got %v", startTime, service.startTime)
	}
}

func TestSystemService_CheckHealth_AllHealthy(t *testing.T) {
	db := createTestDB(t)

	startTime := time.Now().Add(-time.Hour) // 1 hour ago
	service := NewSystemService(db, startTime)

	status, err := service.CheckHealth()

	if err != nil {
		t.Fatalf("CheckHealth returned error: %v", err)
	}
	if status == nil {
		t.Fatal("CheckHealth returned nil status")
	}
	if status.Status != "healthy" {
		t.Errorf("Expected status to be 'healthy', got '%s'", status.Status)
	}
	if status.Version != "1.0.0" {
		t.Errorf("Expected version to be '1.0.0', got '%s'", status.Version)
	}
	if status.Uptime == "" {
		t.Error("Uptime should not be empty")
	}
	if status.Timestamp == "" {
		t.Error("Timestamp should not be empty")
	}

	// Check services
	if status.Services["database"] != "healthy" {
		t.Errorf("Expected database service to be 'healthy', got '%s'", status.Services["database"])
	}
	if status.Services["api"] != "healthy" {
		t.Errorf("Expected api service to be 'healthy', got '%s'", status.Services["api"])
	}
}

func TestSystemService_CheckHealth_DatabaseUnhealthy(t *testing.T) {
	db := createTestDB(t)

	// Close one database to simulate unhealthy state
	if err := db.ResultsDB.Close(); err != nil {
		t.Fatalf("Failed to close results DB: %v", err)
	}

	startTime := time.Now().Add(-time.Minute * 30)
	service := NewSystemService(db, startTime)

	status, err := service.CheckHealth()

	if err != nil {
		t.Fatalf("CheckHealth returned error: %v", err)
	}
	if status == nil {
		t.Fatal("CheckHealth returned nil status")
	}
	if status.Status != "unhealthy" {
		t.Errorf("Expected status to be 'unhealthy', got '%s'", status.Status)
	}
	if status.Version != "1.0.0" {
		t.Errorf("Expected version to be '1.0.0', got '%s'", status.Version)
	}
	if status.Uptime == "" {
		t.Error("Uptime should not be empty")
	}

	// Check services
	if status.Services["database"] != "unhealthy" {
		t.Errorf("Expected database service to be 'unhealthy', got '%s'", status.Services["database"])
	}
	if status.Services["api"] != "healthy" {
		t.Errorf("Expected api service to be 'healthy', got '%s'", status.Services["api"])
	}
}

func TestSystemService_CheckHealth_MultipleDatabasesUnhealthy(t *testing.T) {
	db := createTestDB(t)

	// Close multiple databases to simulate unhealthy state
	if err := db.ResultsDB.Close(); err != nil {
		t.Fatalf("Failed to close results DB: %v", err)
	}
	if err := db.SimulationsDB.Close(); err != nil {
		t.Fatalf("Failed to close simulations DB: %v", err)
	}

	startTime := time.Now().Add(-time.Second * 45)
	service := NewSystemService(db, startTime)

	status, err := service.CheckHealth()

	if err != nil {
		t.Fatalf("CheckHealth returned error: %v", err)
	}
	if status == nil {
		t.Fatal("CheckHealth returned nil status")
	}
	if status.Status != "unhealthy" {
		t.Errorf("Expected status to be 'unhealthy', got '%s'", status.Status)
	}
	if status.Services["database"] != "unhealthy" {
		t.Errorf("Expected database service to be 'unhealthy', got '%s'", status.Services["database"])
	}
	if status.Services["api"] != "healthy" {
		t.Errorf("Expected api service to be 'healthy', got '%s'", status.Services["api"])
	}
}

func TestSystemService_CheckHealth_AllDatabasesUnhealthy(t *testing.T) {
	db := createTestDB(t)

	// Close all databases to simulate unhealthy state
	if err := db.ResultsDB.Close(); err != nil {
		t.Fatalf("Failed to close results DB: %v", err)
	}
	if err := db.SimulationsDB.Close(); err != nil {
		t.Fatalf("Failed to close simulations DB: %v", err)
	}
	if err := db.ConfigsDB.Close(); err != nil {
		t.Fatalf("Failed to close configs DB: %v", err)
	}
	if err := db.FinancesDB.Close(); err != nil {
		t.Fatalf("Failed to close finances DB: %v", err)
	}

	startTime := time.Now()
	service := NewSystemService(db, startTime)

	status, err := service.CheckHealth()

	if err != nil {
		t.Fatalf("CheckHealth returned error: %v", err)
	}
	if status == nil {
		t.Fatal("CheckHealth returned nil status")
	}
	if status.Status != "unhealthy" {
		t.Errorf("Expected status to be 'unhealthy', got '%s'", status.Status)
	}
	if status.Services["database"] != "unhealthy" {
		t.Errorf("Expected database service to be 'unhealthy', got '%s'", status.Services["database"])
	}
	if status.Services["api"] != "healthy" {
		t.Errorf("Expected api service to be 'healthy', got '%s'", status.Services["api"])
	}
}

func TestSystemService_CheckHealth_SimulationsDBUnhealthy(t *testing.T) {
	db := createTestDB(t)

	// Close only simulations database
	if err := db.SimulationsDB.Close(); err != nil {
		t.Fatalf("Failed to close simulations DB: %v", err)
	}

	startTime := time.Now()
	service := NewSystemService(db, startTime)

	status, err := service.CheckHealth()

	if err != nil {
		t.Fatalf("CheckHealth returned error: %v", err)
	}
	if status == nil {
		t.Fatal("CheckHealth returned nil status")
	}
	if status.Status != "unhealthy" {
		t.Errorf("Expected status to be 'unhealthy', got '%s'", status.Status)
	}
	if status.Services["database"] != "unhealthy" {
		t.Errorf("Expected database service to be 'unhealthy', got '%s'", status.Services["database"])
	}
	if status.Services["api"] != "healthy" {
		t.Errorf("Expected api service to be 'healthy', got '%s'", status.Services["api"])
	}
}

func TestSystemService_CheckHealth_ConfigsDBUnhealthy(t *testing.T) {
	db := createTestDB(t)

	// Close only configs database
	if err := db.ConfigsDB.Close(); err != nil {
		t.Fatalf("Failed to close configs DB: %v", err)
	}

	startTime := time.Now()
	service := NewSystemService(db, startTime)

	status, err := service.CheckHealth()

	if err != nil {
		t.Fatalf("CheckHealth returned error: %v", err)
	}
	if status == nil {
		t.Fatal("CheckHealth returned nil status")
	}
	if status.Status != "unhealthy" {
		t.Errorf("Expected status to be 'unhealthy', got '%s'", status.Status)
	}
	if status.Services["database"] != "unhealthy" {
		t.Errorf("Expected database service to be 'unhealthy', got '%s'", status.Services["database"])
	}
	if status.Services["api"] != "healthy" {
		t.Errorf("Expected api service to be 'healthy', got '%s'", status.Services["api"])
	}
}

func TestSystemService_CheckHealth_FinancesDBUnhealthy(t *testing.T) {
	db := createTestDB(t)

	// Close only finances database
	if err := db.FinancesDB.Close(); err != nil {
		t.Fatalf("Failed to close finances DB: %v", err)
	}

	startTime := time.Now()
	service := NewSystemService(db, startTime)

	status, err := service.CheckHealth()

	if err != nil {
		t.Fatalf("CheckHealth returned error: %v", err)
	}
	if status == nil {
		t.Fatal("CheckHealth returned nil status")
	}
	if status.Status != "unhealthy" {
		t.Errorf("Expected status to be 'unhealthy', got '%s'", status.Status)
	}
	if status.Services["database"] != "unhealthy" {
		t.Errorf("Expected database service to be 'unhealthy', got '%s'", status.Services["database"])
	}
	if status.Services["api"] != "healthy" {
		t.Errorf("Expected api service to be 'healthy', got '%s'", status.Services["api"])
	}
}

func TestSystemService_CheckHealth_UptimeCalculation(t *testing.T) {
	db := createTestDB(t)

	startTime := time.Now().Add(-time.Hour*2 - time.Minute*30) // 2.5 hours ago
	service := NewSystemService(db, startTime)

	status, err := service.CheckHealth()

	if err != nil {
		t.Fatalf("CheckHealth returned error: %v", err)
	}
	if status == nil {
		t.Fatal("CheckHealth returned nil status")
	}
	if status.Uptime == "" {
		t.Error("Uptime should not be empty")
	}
}

func TestSystemService_CheckHealth_TimestampFormat(t *testing.T) {
	db := createTestDB(t)

	startTime := time.Now()
	service := NewSystemService(db, startTime)

	status, err := service.CheckHealth()

	if err != nil {
		t.Fatalf("CheckHealth returned error: %v", err)
	}
	if status == nil {
		t.Fatal("CheckHealth returned nil status")
	}

	// Verify timestamp is in RFC3339 format
	_, err = time.Parse(time.RFC3339, status.Timestamp)
	if err != nil {
		t.Errorf("Timestamp should be in RFC3339 format, got '%s', error: %v", status.Timestamp, err)
	}
}
