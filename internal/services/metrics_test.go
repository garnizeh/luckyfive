package services

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func TestNewMetricsService(t *testing.T) {
	db := createTestDB(t)

	startTime := time.Now()
	service := NewMetricsService(db, startTime)

	if service == nil {
		t.Fatal("NewMetricsService returned nil")
	}
	if service.db != db {
		t.Error("Expected db to match")
	}
	if !service.startTime.Equal(startTime) {
		t.Errorf("Expected startTime to be %v, got %v", startTime, service.startTime)
	}
	if service.queryStats == nil {
		t.Error("Expected queryStats to be initialized")
	}
	if service.httpStats == nil {
		t.Error("Expected httpStats to be initialized")
	}
}

func TestMetricsService_RecordQueryExecution(t *testing.T) {
	db := createTestDB(t)

	startTime := time.Now()
	service := NewMetricsService(db, startTime)

	// Record first execution
	duration1 := 100 * time.Millisecond
	service.RecordQueryExecution("test_query", duration1, nil)

	// Record second execution with error
	duration2 := 200 * time.Millisecond
	testErr := sql.ErrNoRows
	service.RecordQueryExecution("test_query", duration2, testErr)

	// Record execution for different query
	duration3 := 50 * time.Millisecond
	service.RecordQueryExecution("other_query", duration3, nil)

	// Check stats
	service.mu.RLock()
	defer service.mu.RUnlock()

	// Check test_query stats
	testStats, exists := service.queryStats["test_query"]
	if !exists {
		t.Fatal("Expected test_query stats to exist")
	}
	if testStats.TotalCalls != 2 {
		t.Errorf("Expected 2 total calls, got %d", testStats.TotalCalls)
	}
	if testStats.TotalTime != duration1+duration2 {
		t.Errorf("Expected total time %v, got %v", duration1+duration2, testStats.TotalTime)
	}
	if testStats.ErrorCount != 1 {
		t.Errorf("Expected 1 error count, got %d", testStats.ErrorCount)
	}
	if testStats.MaxTime != duration2 {
		t.Errorf("Expected max time %v, got %v", duration2, testStats.MaxTime)
	}
	if testStats.MinTime != duration1 {
		t.Errorf("Expected min time %v, got %v", duration1, testStats.MinTime)
	}
	expectedAvg := (duration1 + duration2) / 2
	if testStats.AvgTime != expectedAvg {
		t.Errorf("Expected avg time %v, got %v", expectedAvg, testStats.AvgTime)
	}

	// Check other_query stats
	otherStats, exists := service.queryStats["other_query"]
	if !exists {
		t.Fatal("Expected other_query stats to exist")
	}
	if otherStats.TotalCalls != 1 {
		t.Errorf("Expected 1 total call, got %d", otherStats.TotalCalls)
	}
	if otherStats.ErrorCount != 0 {
		t.Errorf("Expected 0 error count, got %d", otherStats.ErrorCount)
	}
}

func TestMetricsService_RecordHTTPRequest(t *testing.T) {
	db := createTestDB(t)

	startTime := time.Now()
	service := NewMetricsService(db, startTime)

	// Record first request
	duration1 := 50 * time.Millisecond
	service.RecordHTTPRequest("GET", 200, duration1)

	// Record second request
	duration2 := 75 * time.Millisecond
	service.RecordHTTPRequest("POST", 201, duration2)

	// Record third request with error status
	duration3 := 100 * time.Millisecond
	service.RecordHTTPRequest("GET", 500, duration3)

	// Check stats
	service.mu.RLock()
	defer service.mu.RUnlock()

	if service.httpStats.TotalRequests != 3 {
		t.Errorf("Expected 3 total requests, got %d", service.httpStats.TotalRequests)
	}

	// Check status codes
	if service.httpStats.StatusCodes[200] != 1 {
		t.Errorf("Expected 1 request with status 200, got %d", service.httpStats.StatusCodes[200])
	}
	if service.httpStats.StatusCodes[201] != 1 {
		t.Errorf("Expected 1 request with status 201, got %d", service.httpStats.StatusCodes[201])
	}
	if service.httpStats.StatusCodes[500] != 1 {
		t.Errorf("Expected 1 request with status 500, got %d", service.httpStats.StatusCodes[500])
	}

	// Check methods
	if service.httpStats.MethodStats["GET"] != 2 {
		t.Errorf("Expected 2 GET requests, got %d", service.httpStats.MethodStats["GET"])
	}
	if service.httpStats.MethodStats["POST"] != 1 {
		t.Errorf("Expected 1 POST request, got %d", service.httpStats.MethodStats["POST"])
	}

	// Check response time calculations
	expectedAvg := (duration1 + duration2 + duration3) / 3
	if service.httpStats.AvgResponseTime != expectedAvg {
		t.Errorf("Expected avg response time %v, got %v", expectedAvg, service.httpStats.AvgResponseTime)
	}
	if service.httpStats.MaxResponseTime != duration3 {
		t.Errorf("Expected max response time %v, got %v", duration3, service.httpStats.MaxResponseTime)
	}
}

func TestMetricsService_GetSystemMetrics(t *testing.T) {
	db := createTestDB(t)

	startTime := time.Now().Add(-time.Hour)
	service := NewMetricsService(db, startTime)

	// Add some test data
	service.RecordQueryExecution("test_query", 100*time.Millisecond, nil)
	service.RecordHTTPRequest("GET", 200, 50*time.Millisecond)

	metrics, err := service.GetSystemMetrics(context.Background())
	if err != nil {
		t.Fatalf("GetSystemMetrics failed: %v", err)
	}

	if metrics == nil {
		t.Fatal("GetSystemMetrics returned nil")
	}

	// Check uptime
	if metrics.Uptime == "" {
		t.Error("Expected uptime to be set")
	}

	// Check timestamp
	if metrics.Timestamp == "" {
		t.Error("Expected timestamp to be set")
	}

	// Check database stats
	if len(metrics.DatabaseStats) != 5 {
		t.Errorf("Expected 5 database stats, got %d", len(metrics.DatabaseStats))
	}

	expectedDBs := []string{"results", "simulations", "configs", "finances", "sweeps"}
	for _, dbName := range expectedDBs {
		if _, exists := metrics.DatabaseStats[dbName]; !exists {
			t.Errorf("Expected database stats for %s", dbName)
		}
	}

	// Check query stats
	if len(metrics.QueryStats) != 1 {
		t.Errorf("Expected 1 query stat, got %d", len(metrics.QueryStats))
	}

	if _, exists := metrics.QueryStats["test_query"]; !exists {
		t.Error("Expected test_query in query stats")
	}

	// Check HTTP stats
	if metrics.HTTPStats.TotalRequests != 1 {
		t.Errorf("Expected 1 HTTP request, got %d", metrics.HTTPStats.TotalRequests)
	}

	// Check memory stats (placeholder)
	if metrics.MemoryStats == nil {
		t.Error("Expected memory stats to be initialized")
	}
}

func TestMetricsService_Reset(t *testing.T) {
	db := createTestDB(t)

	startTime := time.Now()
	service := NewMetricsService(db, startTime)

	// Add some test data
	service.RecordQueryExecution("test_query", 100*time.Millisecond, nil)
	service.RecordHTTPRequest("GET", 200, 50*time.Millisecond)

	// Verify data exists
	service.mu.RLock()
	initialQueryCount := len(service.queryStats)
	initialRequestCount := service.httpStats.TotalRequests
	service.mu.RUnlock()

	if initialQueryCount == 0 {
		t.Error("Expected query stats to exist before reset")
	}
	if initialRequestCount == 0 {
		t.Error("Expected HTTP stats to exist before reset")
	}

	// Reset
	service.Reset()

	// Verify data is cleared
	service.mu.RLock()
	defer service.mu.RUnlock()

	if len(service.queryStats) != 0 {
		t.Errorf("Expected query stats to be empty after reset, got %d", len(service.queryStats))
	}
	if service.httpStats.TotalRequests != 0 {
		t.Errorf("Expected HTTP requests to be 0 after reset, got %d", service.httpStats.TotalRequests)
	}
	if len(service.httpStats.StatusCodes) != 0 {
		t.Errorf("Expected status codes to be empty after reset, got %d", len(service.httpStats.StatusCodes))
	}
	if len(service.httpStats.MethodStats) != 0 {
		t.Errorf("Expected method stats to be empty after reset, got %d", len(service.httpStats.MethodStats))
	}
}

func TestMetricsService_RecordQueryExecution_ThreadSafety(t *testing.T) {
	db := createTestDB(t)

	startTime := time.Now()
	service := NewMetricsService(db, startTime)

	// Run concurrent recordings
	done := make(chan bool, 10)
	for i := range 10 {
		go func(id int) {
			for j := range 100 {
				queryName := "query_" + string(rune(id))
				duration := time.Duration(j) * time.Millisecond
				service.RecordQueryExecution(queryName, duration, nil)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for range 10 {
		<-done
	}

	// Verify total calls
	service.mu.RLock()
	defer service.mu.RUnlock()

	totalCalls := int64(0)
	for _, stats := range service.queryStats {
		totalCalls += stats.TotalCalls
	}

	if totalCalls != 1000 { // 10 goroutines * 100 calls each
		t.Errorf("Expected 1000 total calls, got %d", totalCalls)
	}
}

func TestMetricsService_RecordHTTPRequest_ThreadSafety(t *testing.T) {
	db := createTestDB(t)

	startTime := time.Now()
	service := NewMetricsService(db, startTime)

	// Run concurrent recordings
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 200; j++ {
				service.RecordHTTPRequest("GET", 200, time.Duration(j)*time.Millisecond)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	// Verify total requests
	service.mu.RLock()
	defer service.mu.RUnlock()

	if service.httpStats.TotalRequests != 1000 { // 5 goroutines * 200 requests each
		t.Errorf("Expected 1000 total requests, got %d", service.httpStats.TotalRequests)
	}

	if service.httpStats.StatusCodes[200] != 1000 {
		t.Errorf("Expected 1000 status 200 codes, got %d", service.httpStats.StatusCodes[200])
	}

	if service.httpStats.MethodStats["GET"] != 1000 {
		t.Errorf("Expected 1000 GET methods, got %d", service.httpStats.MethodStats["GET"])
	}
}
