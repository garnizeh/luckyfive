# luckyfive

Quina Lottery Simulation Platform

Quickstart

Prerequisites:
- Go 1.25+
- sqlite3 (optional for manual DB inspection)

Install dependencies and generate code:

```bash
# From project root
make generate

# Build binaries
make build
```

Run API (dev):

```bash
make run-api
```

Project layout

- `cmd/` - entry points (api, worker, migrate)
- `internal/` - application code (services, handlers, store)
- `pkg/` - reusable packages (predictor, utils)
- `data/` - SQLite database files
- `migrations/` - SQL migration files
- `docs/` - design and user docs

Contributing

Please follow Go idioms and run `golangci-lint` before opening PRs.
