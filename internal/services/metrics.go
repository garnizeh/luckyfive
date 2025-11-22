package services

import (
	"context"
	"sync"
	"time"

	"github.com/garnizeh/luckyfive/internal/store"
)

// MetricsService handles performance monitoring and metrics collection
type MetricsService struct {
	db         *store.DB
	startTime  time.Time
	mu         sync.RWMutex
	queryStats map[string]*QueryStats
	httpStats  *HTTPStats
}

// QueryStats holds statistics for database queries
type QueryStats struct {
	QueryName    string        `json:"query_name"`
	TotalCalls   int64         `json:"total_calls"`
	TotalTime    time.Duration `json:"total_time"`
	AvgTime      time.Duration `json:"avg_time"`
	MaxTime      time.Duration `json:"max_time"`
	MinTime      time.Duration `json:"min_time"`
	ErrorCount   int64         `json:"error_count"`
	LastExecuted time.Time     `json:"last_executed"`
}

// HTTPStats holds statistics for HTTP requests
type HTTPStats struct {
	TotalRequests   int64            `json:"total_requests"`
	StatusCodes     map[int]int64    `json:"status_codes"`
	MethodStats     map[string]int64 `json:"method_stats"`
	AvgResponseTime time.Duration    `json:"avg_response_time"`
	MaxResponseTime time.Duration    `json:"max_response_time"`
}

// SystemMetrics represents comprehensive system performance metrics
type SystemMetrics struct {
	Uptime        string                    `json:"uptime"`
	DatabaseStats map[string]*DatabaseStats `json:"database_stats"`
	QueryStats    map[string]*QueryStats    `json:"query_stats"`
	HTTPStats     *HTTPStats                `json:"http_stats"`
	MemoryStats   *MemoryStats              `json:"memory_stats"`
	Timestamp     string                    `json:"timestamp"`
}

// DatabaseStats holds statistics for individual databases
type DatabaseStats struct {
	Name            string `json:"name"`
	OpenConnections int    `json:"open_connections"`
	IdleConnections int    `json:"idle_connections"`
	InUse           int    `json:"in_use"`
}

// MemoryStats holds basic memory statistics
type MemoryStats struct {
	// Placeholder for future memory monitoring
	AllocatedBytes uint64 `json:"allocated_bytes"`
	SystemBytes    uint64 `json:"system_bytes"`
}

// NewMetricsService creates a new metrics service
func NewMetricsService(db *store.DB, startTime time.Time) *MetricsService {
	return &MetricsService{
		db:         db,
		startTime:  startTime,
		queryStats: make(map[string]*QueryStats),
		httpStats: &HTTPStats{
			StatusCodes: make(map[int]int64),
			MethodStats: make(map[string]int64),
		},
	}
}

// RecordQueryExecution records statistics for a database query execution
func (m *MetricsService) RecordQueryExecution(queryName string, duration time.Duration, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	stats, exists := m.queryStats[queryName]
	if !exists {
		stats = &QueryStats{
			QueryName: queryName,
			MinTime:   time.Hour, // Initialize with large value
		}
		m.queryStats[queryName] = stats
	}

	stats.TotalCalls++
	stats.TotalTime += duration
	stats.LastExecuted = time.Now()

	if duration > stats.MaxTime {
		stats.MaxTime = duration
	}
	if duration < stats.MinTime {
		stats.MinTime = duration
	}
	if err != nil {
		stats.ErrorCount++
	}

	// Calculate average time
	if stats.TotalCalls > 0 {
		stats.AvgTime = stats.TotalTime / time.Duration(stats.TotalCalls)
	}
}

// RecordHTTPRequest records statistics for an HTTP request
func (m *MetricsService) RecordHTTPRequest(method string, statusCode int, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.httpStats.TotalRequests++
	m.httpStats.StatusCodes[statusCode]++
	m.httpStats.MethodStats[method]++

	// Update response time statistics
	if m.httpStats.TotalRequests == 1 {
		m.httpStats.AvgResponseTime = duration
		m.httpStats.MaxResponseTime = duration
	} else {
		// Running average calculation
		m.httpStats.AvgResponseTime = (m.httpStats.AvgResponseTime*time.Duration(m.httpStats.TotalRequests-1) + duration) / time.Duration(m.httpStats.TotalRequests)
		if duration > m.httpStats.MaxResponseTime {
			m.httpStats.MaxResponseTime = duration
		}
	}
}

// GetSystemMetrics returns comprehensive system performance metrics
func (m *MetricsService) GetSystemMetrics(ctx context.Context) (*SystemMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Get database statistics
	dbStats := make(map[string]*DatabaseStats)

	// Results DB stats
	if stats := m.db.ResultsDB.Stats(); true {
		dbStats["results"] = &DatabaseStats{
			Name:            "results",
			OpenConnections: stats.OpenConnections,
			IdleConnections: stats.Idle,
			InUse:           stats.InUse,
		}
	}

	// Simulations DB stats
	if stats := m.db.SimulationsDB.Stats(); true {
		dbStats["simulations"] = &DatabaseStats{
			Name:            "simulations",
			OpenConnections: stats.OpenConnections,
			IdleConnections: stats.Idle,
			InUse:           stats.InUse,
		}
	}

	// Configs DB stats
	if stats := m.db.ConfigsDB.Stats(); true {
		dbStats["configs"] = &DatabaseStats{
			Name:            "configs",
			OpenConnections: stats.OpenConnections,
			IdleConnections: stats.Idle,
			InUse:           stats.InUse,
		}
	}

	// Finances DB stats
	if stats := m.db.FinancesDB.Stats(); true {
		dbStats["finances"] = &DatabaseStats{
			Name:            "finances",
			OpenConnections: stats.OpenConnections,
			IdleConnections: stats.Idle,
			InUse:           stats.InUse,
		}
	}

	// Sweeps DB stats
	if stats := m.db.SweepsDB.Stats(); true {
		dbStats["sweeps"] = &DatabaseStats{
			Name:            "sweeps",
			OpenConnections: stats.OpenConnections,
			IdleConnections: stats.Idle,
			InUse:           stats.InUse,
		}
	}

	// Copy query stats to avoid race conditions
	queryStats := make(map[string]*QueryStats)
	for k, v := range m.queryStats {
		queryStats[k] = &QueryStats{
			QueryName:    v.QueryName,
			TotalCalls:   v.TotalCalls,
			TotalTime:    v.TotalTime,
			AvgTime:      v.AvgTime,
			MaxTime:      v.MaxTime,
			MinTime:      v.MinTime,
			ErrorCount:   v.ErrorCount,
			LastExecuted: v.LastExecuted,
		}
	}

	return &SystemMetrics{
		Uptime:        time.Since(m.startTime).String(),
		DatabaseStats: dbStats,
		QueryStats:    queryStats,
		HTTPStats:     m.httpStats,
		MemoryStats: &MemoryStats{
			AllocatedBytes: 0, // TODO: Implement actual memory monitoring
			SystemBytes:    0,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// Reset resets all metrics (useful for testing or manual resets)
func (m *MetricsService) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.queryStats = make(map[string]*QueryStats)
	m.httpStats = &HTTPStats{
		StatusCodes: make(map[int]int64),
		MethodStats: make(map[string]int64),
	}
}
