package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	_ "modernc.org/sqlite"

	"github.com/garnizeh/luckyfive/pkg/migrator"
)

func main() {
	migrationsDir := flag.String("migrations", "migrations", "path to migrations directory")
	dbsFlag := flag.String("dbs", "data/db/results.db,data/db/simulations.db,data/db/configs.db,data/db/finances.db", "comma-separated list of sqlite db paths")
	onlyFlag := flag.String("only", "", "comma-separated migration versions to apply (e.g. 1,3) — only valid with 'up'")
	fileFlag := flag.String("file", "", "apply a single migration file by path (only valid with 'up')")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "usage: migrate [up|down|version] [flags]\n")
		flag.PrintDefaults()
		os.Exit(2)
	}
	action := strings.ToLower(flag.Arg(0))
	dbs := strings.Split(*dbsFlag, ",")
	logger := slog.Default()

	for _, dbPath := range dbs {
		dbPath = strings.TrimSpace(dbPath)
		if dbPath == "" {
			continue
		}
		// Ensure parent dir exists for file dbs
		if !strings.HasPrefix(dbPath, ":") {
			parent := filepath.Dir(dbPath)
			if parent == "." || parent == "" {
				parent = "."
			}
			// create parent dir recursively (handles data/ and data/db/)
			if err := os.MkdirAll(parent, 0o755); err != nil {
				logger.Error("create parent dir failed", "dir", parent, "err", err)
				os.Exit(1)
			}
		}
		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			logger.Error("open db", "db", dbPath, "err", err)
			os.Exit(1)
		}
		// Choose migrations directory for this DB. If a subdirectory exists under the
		// main migrations directory matching the DB filename (without extension), use it.
		// This allows each logical DB to have its own independent migrations.
		migrationsForDB := *migrationsDir
		if dbPath != ":memory:" {
			base := filepath.Base(dbPath)
			name := strings.TrimSuffix(base, filepath.Ext(base))
			candidate := filepath.Join(*migrationsDir, name)
			if fi, err := os.Stat(candidate); err == nil && fi.IsDir() {
				migrationsForDB = candidate
				logger.Info("using db-specific migrations directory", "db", dbPath, "dir", candidate)
			}
		}
		migrator := migrator.New(db, migrationsForDB, logger)
		switch action {
		case "up":
			logger.Info("migrating up", "db", dbPath)
			// If -only provided, apply only those versions
			if *onlyFlag != "" {
				parts := strings.SplitSeq(*onlyFlag, ",")
				for p := range parts {
					p = strings.TrimSpace(p)
					if p == "" {
						continue
					}
					ver, err := strconv.Atoi(p)
					if err != nil {
						logger.Error("invalid version in -only", "val", p, "err", err)
						db.Close()
						os.Exit(1)
					}
					if err := migrator.ApplyVersion(ver); err != nil {
						logger.Error("apply version failed", "db", dbPath, "ver", ver, "err", err)
						db.Close()
						os.Exit(1)
					}
				}
			} else if *fileFlag != "" {
				// derive version from filename prefix if possible
				base := filepath.Base(*fileFlag)
				// parse leading number
				i := 0
				for i < len(base) && base[i] >= '0' && base[i] <= '9' {
					i++
				}
				if i == 0 {
					logger.Error("cannot determine version from file name", "file", *fileFlag)
					db.Close()
					os.Exit(1)
				}
				ver, err := strconv.Atoi(base[:i])
				if err != nil {
					logger.Error("cannot parse version from file", "file", *fileFlag, "err", err)
					db.Close()
					os.Exit(1)
				}
				if err := migrator.ApplyVersion(ver); err != nil {
					logger.Error("apply file failed", "db", dbPath, "file", *fileFlag, "err", err)
					db.Close()
					os.Exit(1)
				}
			} else {
				if err := migrator.Up(); err != nil {
					logger.Error("migration up failed", "db", dbPath, "err", err)
					db.Close()
					os.Exit(1)
				}
			}
		case "down":
			logger.Info("migrating down", "db", dbPath)
			if err := migrator.Down(); err != nil {
				logger.Error("migration down failed", "db", dbPath, "err", err)
				db.Close()
				os.Exit(1)
			}
		case "version":
			v, err := migrator.Version()
			if err != nil {
				logger.Error("migration version failed", "db", dbPath, "err", err)
				db.Close()
				os.Exit(1)
			}
			fmt.Printf("%s: %d\n", dbPath, v)
		default:
			fmt.Fprintf(os.Stderr, "unknown action: %s\n", action)
			db.Close()
			os.Exit(2)
		}
		db.Close()
	}
}
