package store

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/garnizeh/luckyfive/internal/store/configs"
	"github.com/garnizeh/luckyfive/internal/store/finances"
	"github.com/garnizeh/luckyfive/internal/store/results"
	"github.com/garnizeh/luckyfive/internal/store/simulations"
)

// TestOpenClose verifies that Open creates all four SQLite DBs and initializes queriers.
// It tests basic operations and proper resource cleanup on Close.
func TestOpenClose(t *testing.T) {
	tmp := t.TempDir()

	cfg := Config{
		ResultsPath:     filepath.Join(tmp, "results.db"),
		SimulationsPath: filepath.Join(tmp, "simulations.db"),
		ConfigsPath:     filepath.Join(tmp, "configs.db"),
		FinancesPath:    filepath.Join(tmp, "finances.db"),
	}

	db, err := Open(cfg)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	if db == nil {
		t.Fatalf("Open returned nil DB")
	}
	if db.ResultsDB == nil || db.SimulationsDB == nil || db.ConfigsDB == nil || db.FinancesDB == nil {
		t.Fatalf("one or more DB connections are nil")
	}
	if db.Results == nil || db.Simulations == nil || db.Configs == nil || db.Finances == nil {
		t.Fatalf("one or more queriers are nil")
	}

	// Test basic operations on each DB
	ctx := context.Background()

	// Results DB
	if _, err := db.ResultsDB.ExecContext(ctx, "CREATE TABLE test_results (id INTEGER PRIMARY KEY)"); err != nil {
		t.Fatalf("create table on results DB failed: %v", err)
	}

	// Simulations DB
	if _, err := db.SimulationsDB.ExecContext(ctx, "CREATE TABLE test_simulations (id INTEGER PRIMARY KEY)"); err != nil {
		t.Fatalf("create table on simulations DB failed: %v", err)
	}

	// Configs DB
	if _, err := db.ConfigsDB.ExecContext(ctx, "CREATE TABLE test_configs (id INTEGER PRIMARY KEY)"); err != nil {
		t.Fatalf("create table on configs DB failed: %v", err)
	}

	// Finances DB
	if _, err := db.FinancesDB.ExecContext(ctx, "CREATE TABLE test_finances (id INTEGER PRIMARY KEY)"); err != nil {
		t.Fatalf("create table on finances DB failed: %v", err)
	}

	// Test transaction helpers
	err = db.WithResultsTx(ctx, func(q results.Querier) error {
		// This would be results.Querier, but for test we just return nil
		return nil
	})
	if err != nil {
		t.Fatalf("WithResultsTx failed: %v", err)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// After Close, Ping should return errors
	if err := db.ResultsDB.Ping(); err == nil {
		t.Fatalf("expected Ping on results DB after Close to return error")
	}
	if err := db.SimulationsDB.Ping(); err == nil {
		t.Fatalf("expected Ping on simulations DB after Close to return error")
	}
	if err := db.ConfigsDB.Ping(); err == nil {
		t.Fatalf("expected Ping on configs DB after Close to return error")
	}
	if err := db.FinancesDB.Ping(); err == nil {
		t.Fatalf("expected Ping on finances DB after Close to return error")
	}
}

// TestOpenFailures verifies that Open fails gracefully when DB connections cannot be established.
func TestOpenFailures(t *testing.T) {
	// Test with invalid path that should cause sql.Open to fail
	cfg := Config{
		ResultsPath:     "/invalid/path/results.db",
		SimulationsPath: ":memory:", // This should work
		ConfigsPath:     ":memory:",
		FinancesPath:    ":memory:",
	}

	_, err := Open(cfg)
	if err == nil {
		t.Fatalf("expected Open to fail with invalid path")
	}

	// Test with valid paths but simulate ping failure (hard to simulate, but we can test partial failures)
	// For now, just ensure error handling works
}

// TestWithTxHelpers verifies that transaction helpers work correctly.
func TestWithTxHelpers(t *testing.T) {
	cfg := Config{
		ResultsPath:     ":memory:",
		SimulationsPath: ":memory:",
		ConfigsPath:     ":memory:",
		FinancesPath:    ":memory:",
	}

	db, err := Open(cfg)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Test WithResultsTx
	err = db.WithResultsTx(ctx, func(q results.Querier) error {
		// In real usage, q would be results.Querier
		// For this test, we just verify the transaction executes
		return nil
	})
	if err != nil {
		t.Fatalf("WithResultsTx failed: %v", err)
	}

	// Test WithSimulationsTx
	err = db.WithSimulationsTx(ctx, func(q simulations.Querier) error {
		return nil
	})
	if err != nil {
		t.Fatalf("WithSimulationsTx failed: %v", err)
	}

	// Test WithConfigsTx
	err = db.WithConfigsTx(ctx, func(q configs.Querier) error {
		return nil
	})
	if err != nil {
		t.Fatalf("WithConfigsTx failed: %v", err)
	}

	// Test WithFinancesTx
	err = db.WithFinancesTx(ctx, func(q finances.Querier) error {
		return nil
	})
	if err != nil {
		t.Fatalf("WithFinancesTx failed: %v", err)
	}
}

// TestWithTxRollback verifies that transactions roll back on error.
func TestWithTxRollback(t *testing.T) {
	cfg := Config{
		ResultsPath:     ":memory:",
		SimulationsPath: ":memory:",
		ConfigsPath:     ":memory:",
		FinancesPath:    ":memory:",
	}

	db, err := Open(cfg)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Create a table in results DB
	_, err = db.ResultsDB.ExecContext(ctx, "CREATE TABLE test_rollback (id INTEGER PRIMARY KEY, value TEXT)")
	if err != nil {
		t.Fatalf("failed to create test table: %v", err)
	}

	// Test that transaction rolls back on error
	err = db.WithResultsTx(ctx, func(q results.Querier) error {
		// Insert something
		_, err := db.ResultsDB.ExecContext(ctx, "INSERT INTO test_rollback (value) VALUES ('should_rollback')")
		if err != nil {
			return err
		}
		// Return an error to force rollback
		return fmt.Errorf("forced error for rollback test")
	})
	if err == nil {
		t.Fatalf("expected transaction to fail")
	}

	// Verify the insert was rolled back
	var count int
	err = db.ResultsDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM test_rollback").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query count: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 rows after rollback, got %d", count)
	}
}

// TestWithTxErrorHandling verifies error handling in transaction helpers.
func TestWithTxErrorHandling(t *testing.T) {
	cfg := Config{
		ResultsPath:     ":memory:",
		SimulationsPath: ":memory:",
		ConfigsPath:     ":memory:",
		FinancesPath:    ":memory:",
	}

	db, err := Open(cfg)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Test WithResultsTx with function error
	err = db.WithResultsTx(ctx, func(q results.Querier) error {
		return fmt.Errorf("function error")
	})
	if err == nil {
		t.Fatalf("expected WithResultsTx to return error")
	}

	// Test WithSimulationsTx with function error
	err = db.WithSimulationsTx(ctx, func(q simulations.Querier) error {
		return fmt.Errorf("function error")
	})
	if err == nil {
		t.Fatalf("expected WithSimulationsTx to return error")
	}

	// Test WithConfigsTx with function error
	err = db.WithConfigsTx(ctx, func(q configs.Querier) error {
		return fmt.Errorf("function error")
	})
	if err == nil {
		t.Fatalf("expected WithConfigsTx to return error")
	}

	// Test WithFinancesTx with function error
	err = db.WithFinancesTx(ctx, func(q finances.Querier) error {
		return fmt.Errorf("function error")
	})
	if err == nil {
		t.Fatalf("expected WithFinancesTx to return error")
	}
}

// TestCloseErrors verifies that Close handles errors gracefully.
func TestCloseErrors(t *testing.T) {
	cfg := Config{
		ResultsPath:     ":memory:",
		SimulationsPath: ":memory:",
		ConfigsPath:     ":memory:",
		FinancesPath:    ":memory:",
	}

	db, err := Open(cfg)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	// Close once successfully
	if err := db.Close(); err != nil {
		t.Fatalf("first Close failed: %v", err)
	}

	// Close again should not panic or error
	if err := db.Close(); err != nil {
		t.Fatalf("second Close failed: %v", err)
	}

	// Test Close on nil DB
	var nilDB *DB
	if err := nilDB.Close(); err != nil {
		t.Fatalf("Close on nil DB failed: %v", err)
	}
}

// TestConnectionPooling verifies that connection pool settings are applied.
func TestConnectionPooling(t *testing.T) {
	cfg := Config{
		ResultsPath:     ":memory:",
		SimulationsPath: ":memory:",
		ConfigsPath:     ":memory:",
		FinancesPath:    ":memory:",
	}

	db, err := Open(cfg)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	// Verify connection pool settings
	if db.ResultsDB.Stats().MaxOpenConnections != 25 {
		t.Errorf("expected MaxOpenConnections=25, got %d", db.ResultsDB.Stats().MaxOpenConnections)
	}
	if db.SimulationsDB.Stats().MaxOpenConnections != 25 {
		t.Errorf("expected MaxOpenConnections=25, got %d", db.SimulationsDB.Stats().MaxOpenConnections)
	}
	if db.ConfigsDB.Stats().MaxOpenConnections != 25 {
		t.Errorf("expected MaxOpenConnections=25, got %d", db.ConfigsDB.Stats().MaxOpenConnections)
	}
	if db.FinancesDB.Stats().MaxOpenConnections != 25 {
		t.Errorf("expected MaxOpenConnections=25, got %d", db.FinancesDB.Stats().MaxOpenConnections)
	}
}
