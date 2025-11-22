package store

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"

	"github.com/garnizeh/luckyfive/internal/store/configs"
	"github.com/garnizeh/luckyfive/internal/store/finances"
	"github.com/garnizeh/luckyfive/internal/store/results"
	"github.com/garnizeh/luckyfive/internal/store/simulations"
	"github.com/garnizeh/luckyfive/internal/store/sweeps"
)

// DB holds connections to all four SQLite databases and their corresponding sqlc queriers.
type DB struct {
	ResultsDB     *sql.DB
	SimulationsDB *sql.DB
	ConfigsDB     *sql.DB
	FinancesDB    *sql.DB
	SweepsDB      *sql.DB

	// Querier interfaces (mockable)
	Results     results.Querier
	Simulations simulations.Querier
	Configs     configs.Querier
	Finances    finances.Querier
	Sweeps      sweeps.Querier
}

// Config holds the paths for each database.
type Config struct {
	ResultsPath     string
	SimulationsPath string
	ConfigsPath     string
	FinancesPath    string
	SweepsPath      string
}

// Open opens all four SQLite databases and initializes the queriers.
func Open(cfg Config) (*DB, error) {
	db := &DB{}

	// Open Results DB
	resultsDB, err := sql.Open("sqlite", cfg.ResultsPath)
	if err != nil {
		return nil, fmt.Errorf("open results db: %w", err)
	}
	if err := resultsDB.Ping(); err != nil {
		resultsDB.Close()
		return nil, fmt.Errorf("ping results db: %w", err)
	}
	db.ResultsDB = resultsDB
	db.Results = results.New(resultsDB)

	// Open Simulations DB
	simulationsDB, err := sql.Open("sqlite", cfg.SimulationsPath)
	if err != nil {
		db.ResultsDB.Close()
		return nil, fmt.Errorf("open simulations db: %w", err)
	}
	if err := simulationsDB.Ping(); err != nil {
		db.ResultsDB.Close()
		simulationsDB.Close()
		return nil, fmt.Errorf("ping simulations db: %w", err)
	}
	db.SimulationsDB = simulationsDB
	db.Simulations = simulations.New(simulationsDB)

	// Open Configs DB
	configsDB, err := sql.Open("sqlite", cfg.ConfigsPath)
	if err != nil {
		db.ResultsDB.Close()
		db.SimulationsDB.Close()
		return nil, fmt.Errorf("open configs db: %w", err)
	}
	if err := configsDB.Ping(); err != nil {
		db.ResultsDB.Close()
		db.SimulationsDB.Close()
		configsDB.Close()
		return nil, fmt.Errorf("ping configs db: %w", err)
	}
	db.ConfigsDB = configsDB
	db.Configs = configs.New(configsDB)

	// Open Finances DB
	financesDB, err := sql.Open("sqlite", cfg.FinancesPath)
	if err != nil {
		db.ResultsDB.Close()
		db.SimulationsDB.Close()
		db.ConfigsDB.Close()
		return nil, fmt.Errorf("open finances db: %w", err)
	}
	if err := financesDB.Ping(); err != nil {
		db.ResultsDB.Close()
		db.SimulationsDB.Close()
		db.ConfigsDB.Close()
		financesDB.Close()
		return nil, fmt.Errorf("ping finances db: %w", err)
	}
	db.FinancesDB = financesDB
	db.Finances = finances.New(financesDB)

	// Open Sweeps DB
	sweepsDB, err := sql.Open("sqlite", cfg.SweepsPath)
	if err != nil {
		db.ResultsDB.Close()
		db.SimulationsDB.Close()
		db.ConfigsDB.Close()
		db.FinancesDB.Close()
		return nil, fmt.Errorf("open sweeps db: %w", err)
	}
	if err := sweepsDB.Ping(); err != nil {
		db.ResultsDB.Close()
		db.SimulationsDB.Close()
		db.ConfigsDB.Close()
		db.FinancesDB.Close()
		sweepsDB.Close()
		return nil, fmt.Errorf("ping sweeps db: %w", err)
	}
	db.SweepsDB = sweepsDB
	db.Sweeps = sweeps.New(sweepsDB)

	// Configure connection pools
	for _, sqlDB := range []*sql.DB{resultsDB, simulationsDB, configsDB, financesDB, sweepsDB} {
		sqlDB.SetMaxOpenConns(25)
		sqlDB.SetMaxIdleConns(5)
		sqlDB.SetConnMaxLifetime(0)
	}

	return db, nil
}

// Close closes all database connections.
func (db *DB) Close() error {
	if db == nil {
		return nil
	}
	var errs []error
	if db.ResultsDB != nil {
		if err := db.ResultsDB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close results db: %w", err))
		}
	}
	if db.SimulationsDB != nil {
		if err := db.SimulationsDB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close simulations db: %w", err))
		}
	}
	if db.ConfigsDB != nil {
		if err := db.ConfigsDB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close configs db: %w", err))
		}
	}
	if db.FinancesDB != nil {
		if err := db.FinancesDB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close finances db: %w", err))
		}
	}
	if db.SweepsDB != nil {
		if err := db.SweepsDB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close sweeps db: %w", err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("close errors: %v", errs)
	}
	return nil
}

// WithResultsTx executes a function within a results DB transaction using a querier.
func (db *DB) WithResultsTx(ctx context.Context, fn func(results.Querier) error) error {
	tx, err := db.ResultsDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin results tx: %w", err)
	}
	defer tx.Rollback()

	q := results.New(tx)
	if err := fn(q); err != nil {
		return err
	}

	return tx.Commit()
}

// WithSimulationsTx executes a function within a simulations DB transaction using a querier.
func (db *DB) WithSimulationsTx(ctx context.Context, fn func(simulations.Querier) error) error {
	tx, err := db.SimulationsDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin simulations tx: %w", err)
	}
	defer tx.Rollback()

	q := simulations.New(tx)
	if err := fn(q); err != nil {
		return err
	}

	return tx.Commit()
}

// WithConfigsTx executes a function within a configs DB transaction using a querier.
func (db *DB) WithConfigsTx(ctx context.Context, fn func(configs.Querier) error) error {
	tx, err := db.ConfigsDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin configs tx: %w", err)
	}
	defer tx.Rollback()

	q := configs.New(tx)
	if err := fn(q); err != nil {
		return err
	}

	return tx.Commit()
}

// WithFinancesTx executes a function within a finances DB transaction using a querier.
func (db *DB) WithFinancesTx(ctx context.Context, fn func(finances.Querier) error) error {
	tx, err := db.FinancesDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin finances tx: %w", err)
	}
	defer tx.Rollback()

	q := finances.New(tx)
	if err := fn(q); err != nil {
		return err
	}

	return tx.Commit()
}

// WithSweepsTx executes a function within a sweeps DB transaction using a querier.
func (db *DB) WithSweepsTx(ctx context.Context, fn func(sweeps.Querier) error) error {
	tx, err := db.SweepsDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin sweeps tx: %w", err)
	}
	defer tx.Rollback()

	q := sweeps.New(tx)
	if err := fn(q); err != nil {
		return err
	}

	return tx.Commit()
}
