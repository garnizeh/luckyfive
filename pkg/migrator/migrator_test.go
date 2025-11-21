package migrator

import (
	"database/sql"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func setupTestMigrationsDir(t *testing.T, migrations map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range migrations {
		path := filepath.Join(dir, name)
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestNew(t *testing.T) {
	db := setupTestDB(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	m := New(db, "/tmp/migrations", logger)
	if m == nil {
		t.Fatal("expected non-nil migrator")
	}
	if m.db != db {
		t.Errorf("expected db to be %v, got %v", db, m.db)
	}
	if m.migrationsPath != "/tmp/migrations" {
		t.Errorf("expected migrationsPath to be %q, got %q", "/tmp/migrations", m.migrationsPath)
	}
	if m.logger == nil {
		t.Error("expected non-nil logger")
	}
}

func TestNew_DefaultLogger(t *testing.T) {
	db := setupTestDB(t)

	m := New(db, "/tmp/migrations", nil)
	if m == nil {
		t.Fatal("expected non-nil migrator")
	}
	if m.logger == nil {
		t.Error("expected non-nil logger")
	}
}

func TestMigrator_ensureSchema(t *testing.T) {
	db := setupTestDB(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	m := New(db, "", logger)

	err := m.ensureSchema()
	if err != nil {
		t.Fatal(err)
	}

	// Verify table was created
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_migrations'").Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("expected 1 table, got %d", count)
	}
}

func TestMigrator_listMigrationFiles(t *testing.T) {
	migrations := map[string]string{
		"001_create_users.sql": "CREATE TABLE users (id INTEGER);",
		"002_add_email.sql":    "ALTER TABLE users ADD COLUMN email TEXT;",
		"003_create_posts.sql": "CREATE TABLE posts (id INTEGER);",
		"not_a_migration.txt":  "some content",
		"invalid.sql":          "some sql",
	}

	dir := setupTestMigrationsDir(t, migrations)
	db := setupTestDB(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	m := NewFromFS(db, os.DirFS(dir), logger)

	filesMap, versions, err := m.listMigrationFiles()
	if err != nil {
		t.Fatal(err)
	}

	expectedVersions := []int{1, 2, 3}
	if len(versions) != len(expectedVersions) {
		t.Errorf("expected %d versions, got %d", len(expectedVersions), len(versions))
	}
	for i, v := range versions {
		if v != expectedVersions[i] {
			t.Errorf("expected version %d at index %d, got %d", expectedVersions[i], i, v)
		}
	}
	if len(filesMap) != 3 {
		t.Errorf("expected 3 files, got %d", len(filesMap))
	}

	for _, ver := range expectedVersions {
		path, exists := filesMap[ver]
		if !exists {
			t.Errorf("version %d should exist", ver)
		}
		if path == "" {
			t.Errorf("expected non-empty path for version %d", ver)
		}
	}
}

func TestMigrator_listMigrationFiles_InvalidDir(t *testing.T) {
	db := setupTestDB(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	m := New(db, "/nonexistent/directory", logger)

	_, _, err := m.listMigrationFiles()
	if err == nil {
		t.Fatal("expected error for invalid directory")
	}
	if !strings.Contains(err.Error(), "read migrations dir") {
		t.Errorf("expected error to contain 'read migrations dir', got %q", err.Error())
	}
}

func TestMigrator_appliedVersions(t *testing.T) {
	db := setupTestDB(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	m := New(db, "", logger)

	// Ensure schema exists
	err := m.ensureSchema()
	if err != nil {
		t.Fatal(err)
	}

	// Insert some test data
	_, err = db.Exec("INSERT INTO schema_migrations (version, applied_at) VALUES (1, '2023-01-01T00:00:00Z'), (3, '2023-01-02T00:00:00Z')")
	if err != nil {
		t.Fatal(err)
	}

	applied, err := m.appliedVersions()
	if err != nil {
		t.Fatal(err)
	}
	if len(applied) != 2 {
		t.Errorf("expected 2 applied versions, got %d", len(applied))
	}

	// Check version 1 exists
	t1, exists := applied[1]
	if !exists {
		t.Error("version 1 should exist")
	}
	if t1.Format("2006-01-02T15:04:05Z") != "2023-01-01T00:00:00Z" {
		t.Errorf("expected time 2023-01-01T00:00:00Z, got %s", t1.Format("2006-01-02T15:04:05Z"))
	}

	// Check version 3 exists
	t3, exists := applied[3]
	if !exists {
		t.Error("version 3 should exist")
	}
	if t3.Format("2006-01-02T15:04:05Z") != "2023-01-02T00:00:00Z" {
		t.Errorf("expected time 2023-01-02T00:00:00Z, got %s", t3.Format("2006-01-02T15:04:05Z"))
	}
}

func TestMigrator_Up(t *testing.T) {
	migrations := map[string]string{
		"001_create_users.sql": `-- Migration: 001_create_users.sql
-- Up migration
CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT);
-- Down migration
-- DROP TABLE users;`,
		"002_add_email.sql": `-- Migration: 002_add_email.sql
-- Up migration
ALTER TABLE users ADD COLUMN email TEXT;
-- Down migration
-- ALTER TABLE users DROP COLUMN email;`,
	}

	dir := setupTestMigrationsDir(t, migrations)
	db := setupTestDB(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	m := NewFromFS(db, os.DirFS(dir), logger)

	// Apply migrations
	err := m.Up()
	if err != nil {
		t.Fatal(err)
	}

	// Verify migrations were applied
	version, err := m.Version()
	if err != nil {
		t.Fatal(err)
	}
	if version != 2 {
		t.Errorf("expected version 2, got %d", version)
	}

	// Verify tables exist
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users'").Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("expected 1 users table, got %d", count)
	}

	// Verify column was added
	err = db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('users') WHERE name='email'").Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("expected 1 email column, got %d", count)
	}
}

func TestMigrator_Up_EmptyMigration(t *testing.T) {
	migrations := map[string]string{
		"001_empty.sql": `-- Migration: 001_empty.sql
-- Up migration
-- This is just a comment
-- Down migration
-- Another comment`,
	}

	dir := setupTestMigrationsDir(t, migrations)
	db := setupTestDB(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	m := NewFromFS(db, os.DirFS(dir), logger)

	// Should not fail with empty migration
	err := m.Up()
	if err != nil {
		t.Fatal(err)
	}

	// Version should still be 1 (migration was recorded)
	version, err := m.Version()
	if err != nil {
		t.Fatal(err)
	}
	if version != 1 {
		t.Errorf("expected version 1, got %d", version)
	}
}

func TestMigrator_Up_InvalidSQL(t *testing.T) {
	migrations := map[string]string{
		"001_invalid.sql": `-- Migration: 001_invalid.sql
-- Up migration
INVALID SQL STATEMENT;
-- Down migration
-- DROP TABLE users;`,
	}

	dir := setupTestMigrationsDir(t, migrations)
	db := setupTestDB(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	m := NewFromFS(db, os.DirFS(dir), logger)

	// Should fail with invalid SQL
	err := m.Up()
	if err == nil {
		t.Fatal("expected error with invalid SQL")
	}
	if !strings.Contains(err.Error(), "execute migration 1") {
		t.Errorf("expected error to contain 'execute migration 1', got %q", err.Error())
	}
}

func TestMigrator_Down(t *testing.T) {
	migrations := map[string]string{
		"001_create_users.sql": `-- Migration: 001_create_users.sql
-- Up migration
CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT);
-- Down migration
DROP TABLE users;`,
	}

	dir := setupTestMigrationsDir(t, migrations)
	db := setupTestDB(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	m := NewFromFS(db, os.DirFS(dir), logger)

	// Apply migration first
	err := m.Up()
	if err != nil {
		t.Fatal(err)
	}

	// Verify table exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users'").Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("expected 1 users table, got %d", count)
	}

	// Rollback
	err = m.Down()
	if err != nil {
		t.Fatal(err)
	}

	// Verify table was dropped
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users'").Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("expected 0 users table, got %d", count)
	}

	// Verify version is 0
	version, err := m.Version()
	if err != nil {
		t.Fatal(err)
	}
	if version != 0 {
		t.Errorf("expected version 0, got %d", version)
	}
}

func TestMigrator_Down_NoMigrations(t *testing.T) {
	dir := setupTestMigrationsDir(t, map[string]string{})
	db := setupTestDB(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	m := NewFromFS(db, os.DirFS(dir), logger)

	err := m.Down()
	if err == nil {
		t.Fatal("expected error when no migrations applied")
	}
	if !strings.Contains(err.Error(), "no migrations have been applied") {
		t.Errorf("expected error to contain 'no migrations have been applied', got %q", err.Error())
	}
}

func TestMigrator_Down_NoDownSection(t *testing.T) {
	migrations := map[string]string{
		"001_create_users.sql": `-- Migration: 001_create_users.sql
-- Up migration
CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT);`,
	}

	dir := setupTestMigrationsDir(t, migrations)
	db := setupTestDB(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	m := NewFromFS(db, os.DirFS(dir), logger)

	// Apply migration first
	err := m.Up()
	if err != nil {
		t.Fatal(err)
	}

	// Try to rollback - should fail because no down section
	err = m.Down()
	if err == nil {
		t.Fatal("expected error when no down section")
	}
	if !strings.Contains(err.Error(), "no down migration found") {
		t.Errorf("expected error to contain 'no down migration found', got %q", err.Error())
	}
}

func TestMigrator_Version(t *testing.T) {
	db := setupTestDB(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	m := New(db, "", logger)

	// Initially 0
	version, err := m.Version()
	if err != nil {
		t.Fatal(err)
	}
	if version != 0 {
		t.Errorf("expected version 0, got %d", version)
	}

	// Ensure schema and add a migration
	err = m.ensureSchema()
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec("INSERT INTO schema_migrations (version, applied_at) VALUES (5, '2023-01-01T00:00:00Z')")
	if err != nil {
		t.Fatal(err)
	}

	version, err = m.Version()
	if err != nil {
		t.Fatal(err)
	}
	if version != 5 {
		t.Errorf("expected version 5, got %d", version)
	}
}

func TestMigrator_ApplyVersion(t *testing.T) {
	migrations := map[string]string{
		"001_create_users.sql": `-- Migration: 001_create_users.sql
-- Up migration
CREATE TABLE users (id INTEGER PRIMARY KEY);
-- Down migration
-- DROP TABLE users;`,
	}

	dir := setupTestMigrationsDir(t, migrations)
	db := setupTestDB(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	m := NewFromFS(db, os.DirFS(dir), logger)

	// Apply specific version
	err := m.ApplyVersion(1)
	if err != nil {
		t.Fatal(err)
	}

	// Verify it was applied
	version, err := m.Version()
	if err != nil {
		t.Fatal(err)
	}
	if version != 1 {
		t.Errorf("expected version 1, got %d", version)
	}

	// Verify table exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users'").Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("expected 1 users table, got %d", count)
	}

	// Try to apply again (should be idempotent)
	err = m.ApplyVersion(1)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMigrator_ApplyVersion_NotFound(t *testing.T) {
	dir := setupTestMigrationsDir(t, map[string]string{})
	db := setupTestDB(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	m := New(db, dir, logger)

	err := m.ApplyVersion(999)
	if err == nil {
		t.Fatal("expected error for non-existent migration")
	}
	if !strings.Contains(err.Error(), "migration file for version 999 not found") {
		t.Errorf("expected error to contain 'migration file for version 999 not found', got %q", err.Error())
	}
}

func TestMigrator_ApplyVersion_InvalidSQL(t *testing.T) {
	migrations := map[string]string{
		"001_invalid.sql": `-- Migration: 001_invalid.sql
-- Up migration
INVALID SQL HERE;
-- Down migration
-- DROP TABLE users;`,
	}

	dir := setupTestMigrationsDir(t, migrations)
	db := setupTestDB(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	m := New(db, dir, logger)

	err := m.ApplyVersion(1)
	if err == nil {
		t.Fatal("expected error with invalid SQL")
	}
	if !strings.Contains(err.Error(), "execute migration 1") {
		t.Errorf("expected error to contain 'execute migration 1', got %q", err.Error())
	}
}

func TestMigrator_Up_Idempotent(t *testing.T) {
	migrations := map[string]string{
		"001_create_users.sql": `-- Migration: 001_create_users.sql
-- Up migration
CREATE TABLE users (id INTEGER PRIMARY KEY);
-- Down migration
-- DROP TABLE users;`,
	}

	dir := setupTestMigrationsDir(t, migrations)
	db := setupTestDB(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	m := New(db, dir, logger)

	// Apply migrations first time
	err := m.Up()
	if err != nil {
		t.Fatal(err)
	}

	version1, err := m.Version()
	if err != nil {
		t.Fatal(err)
	}

	// Apply again - should be idempotent
	err = m.Up()
	if err != nil {
		t.Fatal(err)
	}

	version2, err := m.Version()
	if err != nil {
		t.Fatal(err)
	}
	if version1 != version2 {
		t.Errorf("expected versions to be equal, got %d and %d", version1, version2)
	}
}

func TestMigrator_Up_PartialFailure(t *testing.T) {
	migrations := map[string]string{
		"001_create_users.sql": `-- Migration: 001_create_users.sql
-- Up migration
CREATE TABLE users (id INTEGER PRIMARY KEY);
-- Down migration
-- DROP TABLE users;`,
		"002_invalid.sql": `-- Migration: 002_invalid.sql
-- Up migration
INVALID SQL;
-- Down migration
-- DROP TABLE users;`,
	}

	dir := setupTestMigrationsDir(t, migrations)
	db := setupTestDB(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	m := New(db, dir, logger)

	// Should fail on second migration
	err := m.Up()
	if err == nil {
		t.Fatal("expected error on invalid SQL")
	}
	if !strings.Contains(err.Error(), "execute migration 2") {
		t.Errorf("expected error to contain 'execute migration 2', got %q", err.Error())
	}

	// First migration should still be applied
	version, err := m.Version()
	if err != nil {
		t.Fatal(err)
	}
	if version != 1 {
		t.Errorf("expected version 1, got %d", version)
	}
}
