package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"

	"github.com/garnizeh/luckyfive/internal/store/configs"
	"github.com/garnizeh/luckyfive/internal/store/finances"
	"github.com/garnizeh/luckyfive/internal/store/results"
	"github.com/garnizeh/luckyfive/internal/store/simulations"
)

// Store is a lightweight container that holds the DB connection and generated
// sqlc query objects for the different logical databases/tables.
type Store struct {
	DB          *sql.DB
	Results     *results.Queries
	Simulations *simulations.Queries
	Configs     *configs.Queries
	Finances    *finances.Queries
}

// Open opens a sqlite database at the given path and returns a Store wired with
// the sqlc-generated query objects. For file-backed DBs the path should be a
// filesystem path (e.g. "data/db/results.db"). For an in-memory DB pass
// ":memory:" which will be opened as-is.
func Open(path string) (*Store, error) {
	var dsn string
	// If caller passed a special DSN (starts with file: or :memory:), use it as-is.
	if strings.HasPrefix(path, "file:") || strings.HasPrefix(path, ":") {
		dsn = path
	} else {
		// Use a file: URI so sqlite opens with read-write-create semantics and a shared cache.
		dsn = fmt.Sprintf("file:%s?cache=shared&mode=rwc", path)
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	// Verify connectivity.
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	s := &Store{
		DB:          db,
		Results:     results.New(db),
		Simulations: simulations.New(db),
		Configs:     configs.New(db),
		Finances:    finances.New(db),
	}

	return s, nil
}

// Close closes the underlying DB connection.
func (s *Store) Close() error {
	if s == nil || s.DB == nil {
		return nil
	}
	return s.DB.Close()
}

// BeginTx is a thin helper that starts a transaction with the provided options.
func (s *Store) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	if s == nil || s.DB == nil {
		return nil, fmt.Errorf("store not opened")
	}
	return s.DB.BeginTx(ctx, opts)
}
