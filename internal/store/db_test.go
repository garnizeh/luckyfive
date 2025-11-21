package store

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/garnizeh/luckyfive/internal/store/configs"
	"github.com/garnizeh/luckyfive/internal/store/finances"
	"github.com/garnizeh/luckyfive/internal/store/results"
	"github.com/garnizeh/luckyfive/internal/store/simulations"
)

// TestOpen_Success ensures Open initializes DBs and queriers for in-memory DBs.
func TestOpen_Success(t *testing.T) {
	cfg := Config{
		ResultsPath:     ":memory:",
		SimulationsPath: ":memory:",
		ConfigsPath:     ":memory:",
		FinancesPath:    ":memory:",
	}

	db, err := Open(cfg)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	defer db.Close()

	if db.ResultsDB == nil || db.SimulationsDB == nil || db.ConfigsDB == nil || db.FinancesDB == nil {
		t.Fatalf("expected all DB connections to be non-nil")
	}
}

// TestOpenClose verifies Open on file-backed DBs and that Close releases resources.
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

	// create a table on each DB to exercise the connections
	ctx := context.Background()
	if _, err := db.ResultsDB.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS test_results (id INTEGER PRIMARY KEY)"); err != nil {
		t.Fatalf("create table on results DB failed: %v", err)
	}
	if _, err := db.SimulationsDB.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS test_simulations (id INTEGER PRIMARY KEY)"); err != nil {
		t.Fatalf("create table on simulations DB failed: %v", err)
	}
	if _, err := db.ConfigsDB.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS test_configs (id INTEGER PRIMARY KEY)"); err != nil {
		t.Fatalf("create table on configs DB failed: %v", err)
	}
	if _, err := db.FinancesDB.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS test_finances (id INTEGER PRIMARY KEY)"); err != nil {
		t.Fatalf("create table on finances DB failed: %v", err)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// After Close, Ping should return errors
	if err := db.ResultsDB.Ping(); err == nil {
		t.Fatalf("expected Ping on results DB after Close to return error")
	}
}

// TestWithResultsTx_RollbackOnFnError ensures that when the user function returns an error the transaction is rolled back.
func TestWithResultsTx_RollbackOnFnError(t *testing.T) {
	dbConn, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite: %v", err)
	}
	defer dbConn.Close()

	// minimal draws table for sqlc-generated InsertDraw
	_, err = dbConn.Exec(`CREATE TABLE IF NOT EXISTS draws (
        contest INTEGER PRIMARY KEY,
        draw_date TEXT NOT NULL,
        bola1 INTEGER NOT NULL,
        bola2 INTEGER NOT NULL,
        bola3 INTEGER NOT NULL,
        bola4 INTEGER NOT NULL,
        bola5 INTEGER NOT NULL,
        source TEXT,
        raw_row TEXT
    );`)
	if err != nil {
		t.Fatalf("failed to create draws table: %v", err)
	}

	sdb := &DB{ResultsDB: dbConn, Results: results.New(dbConn)}

	var errTest = errors.New("fn error")
	err = sdb.WithResultsTx(context.Background(), func(q results.Querier) error {
		params := results.InsertDrawParams{
			Contest:  1,
			DrawDate: "2024-01-01",
			Bola1:    1, Bola2: 2, Bola3: 3, Bola4: 4, Bola5: 5,
			Source: sql.NullString{String: "t", Valid: true},
			RawRow: sql.NullString{String: "r", Valid: true},
		}
		if err := q.InsertDraw(context.Background(), params); err != nil {
			return err
		}
		return errTest
	})

	if err == nil {
		t.Fatalf("expected error from WithResultsTx because fn returned error")
	}

	c, err := sdb.Results.CountDraws(context.Background())
	if err != nil {
		t.Fatalf("failed to count draws: %v", err)
	}
	if c != 0 {
		t.Fatalf("expected 0 draws after rollback, got %d", c)
	}
}

// TestWithResultsTx_BeginTxFailure verifies BeginTx error is returned when DB is closed.
func TestWithResultsTx_BeginTxFailure(t *testing.T) {
	dbConn, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite: %v", err)
	}

	if err := dbConn.Close(); err != nil {
		t.Fatalf("failed to close DB: %v", err)
	}

	sdb := &DB{ResultsDB: dbConn, Results: results.New(dbConn)}

	err = sdb.WithResultsTx(context.Background(), func(q results.Querier) error { return nil })
	if err == nil {
		t.Fatalf("expected BeginTx to fail on closed DB")
	}
}

// TestCloseErrors ensures Close is idempotent and handles nil DB.
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

	if err := db.Close(); err != nil {
		t.Fatalf("first Close failed: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("second Close failed: %v", err)
	}

	var nilDB *DB
	if err := nilDB.Close(); err != nil {
		t.Fatalf("Close on nil DB failed: %v", err)
	}
}

// TestWithResultsTx_Success ensures that a successful function commits the transaction.
func TestWithResultsTx_Success(t *testing.T) {
	dbConn, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite: %v", err)
	}
	defer dbConn.Close()

	// minimal draws table
	if _, err := dbConn.Exec(`CREATE TABLE IF NOT EXISTS draws (
        contest INTEGER PRIMARY KEY,
        draw_date TEXT NOT NULL,
        bola1 INTEGER NOT NULL,
        bola2 INTEGER NOT NULL,
        bola3 INTEGER NOT NULL,
        bola4 INTEGER NOT NULL,
        bola5 INTEGER NOT NULL,
        source TEXT,
        imported_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
        raw_row TEXT
    );`); err != nil {
		t.Fatalf("failed to create draws table: %v", err)
	}

	sdb := &DB{ResultsDB: dbConn, Results: results.New(dbConn)}

	if err := sdb.WithResultsTx(context.Background(), func(q results.Querier) error {
		p := results.InsertDrawParams{
			Contest:  10,
			DrawDate: "2025-01-02",
			Bola1:    1, Bola2: 2, Bola3: 3, Bola4: 4, Bola5: 5,
			Source: sql.NullString{String: "s", Valid: true},
			RawRow: sql.NullString{String: "r", Valid: true},
		}
		return q.InsertDraw(context.Background(), p)
	}); err != nil {
		t.Fatalf("WithResultsTx commit failed: %v", err)
	}

	// verify inserted
	cnt, err := sdb.Results.CountDraws(context.Background())
	if err != nil {
		t.Fatalf("CountDraws failed: %v", err)
	}
	if cnt != 1 {
		t.Fatalf("expected 1 draw after commit, got %d", cnt)
	}
}

// TestWithOtherTx_BasicPaths checks commit and rollback behavior for other DBs.
func TestWithOtherTx_BasicPaths(t *testing.T) {
	// Simulations
	simDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open simulations db: %v", err)
	}
	defer simDB.Close()
	sdb := &DB{SimulationsDB: simDB, Simulations: simulations.New(simDB)}

	// success (commit)
	if err := sdb.WithSimulationsTx(context.Background(), func(q simulations.Querier) error {
		// no-op
		return nil
	}); err != nil {
		t.Fatalf("WithSimulationsTx commit failed: %v", err)
	}

	// rollback when fn returns error
	var sentinel = errors.New("fail")
	if err := sdb.WithSimulationsTx(context.Background(), func(q simulations.Querier) error {
		return sentinel
	}); err == nil {
		t.Fatalf("expected error from WithSimulationsTx when fn returns error")
	}

	// Configs
	cfgDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open configs db: %v", err)
	}
	defer cfgDB.Close()
	cdb := &DB{ConfigsDB: cfgDB, Configs: configs.New(cfgDB)}

	if err := cdb.WithConfigsTx(context.Background(), func(q configs.Querier) error { return nil }); err != nil {
		t.Fatalf("WithConfigsTx commit failed: %v", err)
	}
	if err := cdb.WithConfigsTx(context.Background(), func(q configs.Querier) error { return sentinel }); err == nil {
		t.Fatalf("expected error from WithConfigsTx when fn returns error")
	}

	// Finances
	fDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open finances db: %v", err)
	}
	defer fDB.Close()
	fdb := &DB{FinancesDB: fDB, Finances: finances.New(fDB)}

	if err := fdb.WithFinancesTx(context.Background(), func(q finances.Querier) error { return nil }); err != nil {
		t.Fatalf("WithFinancesTx commit failed: %v", err)
	}
	if err := fdb.WithFinancesTx(context.Background(), func(q finances.Querier) error { return sentinel }); err == nil {
		t.Fatalf("expected error from WithFinancesTx when fn returns error")
	}
}

// TestBeginTxFailures checks that BeginTx returns errors when DB is closed.
func TestBeginTxFailures(t *testing.T) {
	// Results
	rdb, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open results db: %v", err)
	}
	if err := rdb.Close(); err != nil {
		t.Fatalf("close results db: %v", err)
	}
	rd := &DB{ResultsDB: rdb, Results: results.New(rdb)}
	if err := rd.WithResultsTx(context.Background(), func(q results.Querier) error { return nil }); err == nil {
		t.Fatalf("expected BeginTx to fail on closed ResultsDB")
	}

	// Simulations
	sdbConn, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open simulations db: %v", err)
	}
	if err := sdbConn.Close(); err != nil {
		t.Fatalf("close simulations db: %v", err)
	}
	sd := &DB{SimulationsDB: sdbConn, Simulations: simulations.New(sdbConn)}
	if err := sd.WithSimulationsTx(context.Background(), func(q simulations.Querier) error { return nil }); err == nil {
		t.Fatalf("expected BeginTx to fail on closed SimulationsDB")
	}

	// Configs
	cdbConn, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open configs db: %v", err)
	}
	if err := cdbConn.Close(); err != nil {
		t.Fatalf("close configs db: %v", err)
	}
	cd := &DB{ConfigsDB: cdbConn, Configs: configs.New(cdbConn)}
	if err := cd.WithConfigsTx(context.Background(), func(q configs.Querier) error { return nil }); err == nil {
		t.Fatalf("expected BeginTx to fail on closed ConfigsDB")
	}

	// Finances
	fdbConn, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open finances db: %v", err)
	}
	if err := fdbConn.Close(); err != nil {
		t.Fatalf("close finances db: %v", err)
	}
	fd := &DB{FinancesDB: fdbConn, Finances: finances.New(fdbConn)}
	if err := fd.WithFinancesTx(context.Background(), func(q finances.Querier) error { return nil }); err == nil {
		t.Fatalf("expected BeginTx to fail on closed FinancesDB")
	}
}

// TestOpen_PingFailures exercises Open branches where Ping fails for each DB in turn.
func TestOpen_PingFailures(t *testing.T) {
	// Results ping fails when path points to non-existent directory in read-only mode
	cfg := Config{ResultsPath: "file:/this_dir_should_not_exist_12345/results.db?mode=ro", SimulationsPath: ":memory:", ConfigsPath: ":memory:", FinancesPath: ":memory:"}
	if _, err := Open(cfg); err == nil {
		t.Fatalf("expected Open to fail when Results ping fails")
	}

	// Simulations ping fails
	cfg = Config{ResultsPath: ":memory:", SimulationsPath: "file:/this_dir_should_not_exist_12345/sim.db?mode=ro", ConfigsPath: ":memory:", FinancesPath: ":memory:"}
	if _, err := Open(cfg); err == nil {
		t.Fatalf("expected Open to fail when Simulations ping fails")
	}

	// Configs ping fails
	cfg = Config{ResultsPath: ":memory:", SimulationsPath: ":memory:", ConfigsPath: "file:/this_dir_should_not_exist_12345/configs.db?mode=ro", FinancesPath: ":memory:"}
	if _, err := Open(cfg); err == nil {
		t.Fatalf("expected Open to fail when Configs ping fails")
	}

	// Finances ping fails
	cfg = Config{ResultsPath: ":memory:", SimulationsPath: ":memory:", ConfigsPath: ":memory:", FinancesPath: "file:/this_dir_should_not_exist_12345/finances.db?mode=ro"}
	if _, err := Open(cfg); err == nil {
		t.Fatalf("expected Open to fail when Finances ping fails")
	}
}
