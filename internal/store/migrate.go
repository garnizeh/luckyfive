package store

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"log/slog"
)

// Migrator applies SQL migration files from a directory.
// Migration files must start with a numeric prefix (e.g. 001_create_results.sql).
type Migrator struct {
	db             *sql.DB
	migrationsPath string
	logger         *slog.Logger
}

// NewMigrator creates a new Migrator. migrationsPath is the directory containing .sql files.
func NewMigrator(db *sql.DB, migrationsPath string, logger *slog.Logger) *Migrator {
	if logger == nil {
		logger = slog.Default()
	}
	return &Migrator{db: db, migrationsPath: migrationsPath, logger: logger}
}

// ensureSchema ensures the schema_migrations table exists.
func (m *Migrator) ensureSchema() error {
	const stmt = `CREATE TABLE IF NOT EXISTS schema_migrations (
  version INTEGER PRIMARY KEY,
  applied_at TEXT NOT NULL
);`
	_, err := m.db.Exec(stmt)
	return err
}

// listMigrationFiles returns a map version->path sorted by version asc
func (m *Migrator) listMigrationFiles() (map[int]string, []int, error) {
	files, err := os.ReadDir(m.migrationsPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read migrations dir: %w", err)
	}
	entries := map[int]string{}
	versions := []int{}
	for _, fi := range files {
		if fi.IsDir() {
			continue
		}
		name := fi.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}
		// parse leading number
		i := 0
		for i < len(name) && name[i] >= '0' && name[i] <= '9' {
			i++
		}
		if i == 0 {
			// skip files without numeric prefix
			continue
		}
		verStr := name[:i]
		ver, err := strconv.Atoi(verStr)
		if err != nil {
			continue
		}
		entries[ver] = filepath.Join(m.migrationsPath, name)
		versions = append(versions, ver)
	}
	sort.Ints(versions)
	return entries, versions, nil
}

// appliedVersions returns a map of applied versions
func (m *Migrator) appliedVersions() (map[int]time.Time, error) {
	rows, err := m.db.Query(`SELECT version, applied_at FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("query schema_migrations: %w", err)
	}
	defer rows.Close()
	res := map[int]time.Time{}
	for rows.Next() {
		var v int
		var at string
		if err := rows.Scan(&v, &at); err != nil {
			return nil, err
		}
		t, _ := time.Parse(time.RFC3339, at)
		res[v] = t
	}
	return res, nil
}

// Up applies all pending migrations.
func (m *Migrator) Up() error {
	if err := m.ensureSchema(); err != nil {
		return err
	}
	filesMap, versions, err := m.listMigrationFiles()
	if err != nil {
		return err
	}
	applied, err := m.appliedVersions()
	if err != nil {
		return err
	}
	for _, ver := range versions {
		if _, ok := applied[ver]; ok {
			continue // already applied
		}
		path := filesMap[ver]
		m.logger.Info("applying migration", "version", ver, "file", path)
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		// Only run the up section (before a "-- Down" marker)
		sqlText := string(content)
		lines := strings.Split(sqlText, "\n")
		upLines := []string{}
		for _, L := range lines {
			trimmed := strings.TrimSpace(L)
			if strings.HasPrefix(strings.ToLower(trimmed), "-- down") {
				break
			}
			upLines = append(upLines, L)
		}
		upSQL := strings.TrimSpace(strings.Join(upLines, "\n"))
		if upSQL == "" {
			m.logger.Info("no up SQL found for migration; skipping execution", "version", ver)
		} else {
			tx, err := m.db.Begin()
			if err != nil {
				return err
			}
			if _, err := tx.Exec(upSQL); err != nil {
				tx.Rollback()
				return fmt.Errorf("execute migration %d: %w", ver, err)
			}
			if _, err := tx.Exec(`INSERT INTO schema_migrations(version, applied_at) VALUES(?, ?)`, ver, time.Now().UTC().Format(time.RFC3339)); err != nil {
				tx.Rollback()
				return err
			}
			if err := tx.Commit(); err != nil {
				return err
			}
		}
	}
	return nil
}

// Down rolls back the last applied migration. It looks for a "-- Down" marker in the migration file
// and will attempt to uncomment lines that start with '--' before executing.
func (m *Migrator) Down() error {
	if err := m.ensureSchema(); err != nil {
		return err
	}
	// find last applied version
	var ver sql.NullInt64
	row := m.db.QueryRow(`SELECT MAX(version) FROM schema_migrations`)
	if err := row.Scan(&ver); err != nil {
		return fmt.Errorf("query max version: %w", err)
	}
	if !ver.Valid || ver.Int64 == 0 {
		return errors.New("no migrations have been applied")
	}
	last := int(ver.Int64)
	filesMap, _, err := m.listMigrationFiles()
	if err != nil {
		return err
	}
	path, ok := filesMap[last]
	if !ok {
		return fmt.Errorf("migration file for version %d not found", last)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	// extract down section
	sqlText := string(content)
	lines := strings.Split(sqlText, "\n")
	downLines := []string{}
	inDown := false
	for _, L := range lines {
		trimmed := strings.TrimSpace(L)
		if strings.HasPrefix(strings.ToLower(trimmed), "-- down") {
			inDown = true
			continue
		}
		if inDown {
			// if line starts with --, remove leading --
			if strings.HasPrefix(strings.TrimSpace(L), "--") {
				// remove first occurrence of --
				idx := strings.Index(L, "--")
				if idx >= 0 {
					L = L[idx+2:]
				}
			}
			downLines = append(downLines, L)
		}
	}
	downSQL := strings.TrimSpace(strings.Join(downLines, "\n"))
	if downSQL == "" {
		return fmt.Errorf("no down migration found for version %d", last)
	}
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	if _, err := tx.Exec(downSQL); err != nil {
		tx.Rollback()
		return fmt.Errorf("execute down migration %d: %w", last, err)
	}
	if _, err := tx.Exec(`DELETE FROM schema_migrations WHERE version = ?`, last); err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	m.logger.Info("rolled back migration", "version", last)
	return nil
}

// Version returns the current applied migration version (max) or 0 if none.
func (m *Migrator) Version() (int, error) {
	if err := m.ensureSchema(); err != nil {
		return 0, err
	}
	var ver sql.NullInt64
	row := m.db.QueryRow(`SELECT MAX(version) FROM schema_migrations`)
	if err := row.Scan(&ver); err != nil {
		return 0, err
	}
	if !ver.Valid {
		return 0, nil
	}
	return int(ver.Int64), nil
}

// ApplyVersion applies a single migration version if it exists and is not applied yet.
func (m *Migrator) ApplyVersion(version int) error {
	if err := m.ensureSchema(); err != nil {
		return err
	}
	filesMap, _, err := m.listMigrationFiles()
	if err != nil {
		return err
	}
	path, ok := filesMap[version]
	if !ok {
		return fmt.Errorf("migration file for version %d not found", version)
	}
	applied, err := m.appliedVersions()
	if err != nil {
		return err
	}
	if _, ok := applied[version]; ok {
		m.logger.Info("migration already applied; skipping", "version", version, "file", path)
		return nil
	}
	m.logger.Info("applying migration", "version", version, "file", path)
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	// extract up section (before -- Down)
	sqlText := string(content)
	lines := strings.Split(sqlText, "\n")
	upLines := []string{}
	for _, L := range lines {
		trimmed := strings.TrimSpace(L)
		if strings.HasPrefix(strings.ToLower(trimmed), "-- down") {
			break
		}
		upLines = append(upLines, L)
	}
	upSQL := strings.TrimSpace(strings.Join(upLines, "\n"))
	if upSQL == "" {
		m.logger.Info("no up SQL found for migration; skipping execution", "version", version)
		return nil
	}
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	if _, err := tx.Exec(upSQL); err != nil {
		tx.Rollback()
		return fmt.Errorf("execute migration %d: %w", version, err)
	}
	if _, err := tx.Exec(`INSERT INTO schema_migrations(version, applied_at) VALUES(?, ?)`, version, time.Now().UTC().Format(time.RFC3339)); err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
