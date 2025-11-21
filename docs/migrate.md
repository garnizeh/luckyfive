## CLI: migrate

This document describes the `migrate` command-line tool included in this repository. It applies SQL migration files to one or more SQLite databases.

Location
- Built binary: `bin/migrate` (use `make build` to produce binaries in `bin/`)
- Source: `cmd/migrate`

Overview
- Action: `up` — apply pending migrations
- Action: `down` — roll back last applied migration
- Action: `version` — print current applied migration version

Default behavior
- By default the CLI reads migration SQL files from the `migrations/` directory.
- Migrations are files with a numeric prefix, e.g. `001_create_results.sql`, `002_add_table.sql`.
- The default database paths are under `data/db/`:
  - `data/db/results.db`
  - `data/db/simulations.db`
  - `data/db/configs.db`
  - `data/db/finances.db`

Per-database migrations
- If a subdirectory exists under `migrations/` with the same base name as the DB file (for example `migrations/results/`), the CLI will use that directory for that DB instead of the global `migrations/` directory. This lets you keep per-DB migrations when desired.

Command usage

Basic:

```bash
./bin/migrate up -migrations migrations -dbs data/db/results.db
./bin/migrate down -migrations migrations -dbs data/db/results.db
./bin/migrate version -migrations migrations -dbs data/db/results.db
```

Flags
- `-migrations` (default: `migrations`) — path to migrations directory
- `-dbs` (default: `data/db/results.db,data/db/simulations.db,data/db/configs.db,data/db/finances.db`) — comma-separated sqlite DB paths
- `-only` — comma-separated migration versions to apply (e.g. `-only 2,4`). Valid only with the `up` action. Applies only those numeric migrations.
- `-file` — a specific migration file path to apply (only valid with `up`). The CLI derives the numeric version from the file name prefix and applies that one migration.

How migrations are parsed
- The migrator reads each `.sql` file and applies the `up` section. The `up` section is the SQL up to a `-- Down` marker. Any SQL after the `-- Down` marker is considered the rollback section.
- It is common to keep the `Down` section commented (prefixed with `--`) so tools like `sqlc` ignore the destructive statements while still documenting the rollback. The migrator will attempt to strip leading `--` from the down section when executing `down`.

Safety notes
- `down` may execute destructive statements (DROP TABLE, DROP VIEW). Always test rollbacks in a non-production environment and keep backups.
- Consider adding a `--force` flag or confirmation prompt before executing `down` in production environments (not currently implemented).

Examples

- Apply all pending migrations for the default DBs:

```bash
./bin/migrate up
```

- Apply migrations only for a specific version to `results.db`:

```bash
./bin/migrate up -dbs data/db/results.db -only 2
```

- Apply a migration by file name (derives numeric prefix):

```bash
./bin/migrate up -file migrations/002_create_simulations.sql -dbs data/db/simulations.db
```

- Roll back the last applied migration for all default DBs:

```bash
./bin/migrate down
```

Notes about build and CI
- Binaries should be built into `bin/` using `make build` to follow repository policy.
- The Makefile currently suppresses sqlc failures during `make generate`; we recommend making `make generate` fail on sqlc errors in CI so generation issues are visible.

Where to look for more
- Migrations directory: `migrations/`
- Migrator implementation: `internal/store/migrate.go`
- CLI source: `cmd/migrate/main.go`

If you want, I can also:
- Add a `--force` or confirmation behavior for `down` (recommended)
- Add an example CI job that runs migrations against a temporary DB before deploying
