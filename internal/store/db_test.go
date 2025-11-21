package store

import (
	"context"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

// TestOpenClose_FileDB verifies that Open creates a file-backed sqlite DB and
// that basic DB operations (Exec, BeginTx, Commit) work as expected. It also
// checks that resources are released on Close.
func TestOpenClose_FileDB(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.db")

	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open(file) failed: %v", err)
	}
	if s == nil {
		t.Fatalf("Open returned nil store")
	}
	if s.DB == nil {
		t.Fatalf("store.DB is nil")
	}
	if s.Results == nil || s.Simulations == nil || s.Configs == nil || s.Finances == nil {
		t.Fatalf("one or more query objects are nil")
	}

	ctx := context.Background()
	if _, err := s.DB.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS foo (id INTEGER PRIMARY KEY)"); err != nil {
		t.Fatalf("create table failed: %v", err)
	}

	tx, err := s.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx failed: %v", err)
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO foo(id) VALUES (1)"); err != nil {
		tx.Rollback()
		t.Fatalf("tx insert failed: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("tx commit failed: %v", err)
	}

	var cnt int
	if err := s.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM foo").Scan(&cnt); err != nil {
		t.Fatalf("select count failed: %v", err)
	}
	if cnt != 1 {
		t.Fatalf("expected 1 row, got %d", cnt)
	}

	if err := s.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
	// After Close, Ping should return an error.
	if err := s.DB.Ping(); err == nil {
		t.Fatalf("expected Ping after Close to return error")
	}
}

// TestOpen_MemoryDB verifies in-memory DB opening and transaction behaviour.
func TestOpen_MemoryDB(t *testing.T) {
	s, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open(:memory:) failed: %v", err)
	}
	defer s.Close()

	ctx := context.Background()
	if _, err := s.DB.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS items (id INTEGER PRIMARY KEY, name TEXT)"); err != nil {
		t.Fatalf("create table failed: %v", err)
	}

	tx, err := s.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx failed: %v", err)
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO items(name) VALUES ('a')"); err != nil {
		tx.Rollback()
		t.Fatalf("tx insert failed: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("tx commit failed: %v", err)
	}

	var n int
	if err := s.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM items").Scan(&n); err != nil {
		t.Fatalf("select count failed: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 row, got %d", n)
	}
}
