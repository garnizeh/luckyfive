package services

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/garnizeh/luckyfive/internal/store"
)

// createTestDB creates an in-memory SQLite database for testing
func createTestDB(t *testing.T) *store.DB {
	t.Helper()

	// Create in-memory databases
	resultsDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create results DB: %v", err)
	}

	simulationsDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create simulations DB: %v", err)
	}

	configsDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create configs DB: %v", err)
	}

	financesDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create finances DB: %v", err)
	}

	sweepsDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create sweeps DB: %v", err)
	}

	// Create a mock DB struct
	db := &store.DB{
		ResultsDB:     resultsDB,
		SimulationsDB: simulationsDB,
		ConfigsDB:     configsDB,
		FinancesDB:    financesDB,
		SweepsDB:      sweepsDB,
	}

	// Configure connection pools
	for _, sqlDB := range []*sql.DB{resultsDB, simulationsDB, configsDB, financesDB, sweepsDB} {
		sqlDB.SetMaxOpenConns(1)
		sqlDB.SetMaxIdleConns(1)
	}

	t.Cleanup(func() {
		if db == nil {
			return
		}

		if err := db.Close(); err != nil {
			t.Errorf("Failed to close test DB: %v", err)
		}
	})
	return db
}
