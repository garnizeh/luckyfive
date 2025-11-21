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

Dependencies & developer tools

This project depends on several Go libraries and development tools. Install the Go libraries via `go get`/`go mod tidy` (already handled by `make generate`) and install the developer tools for code generation and linting as shown below.

Go libraries (managed in `go.mod`):
- modernc.org/sqlite (SQLite driver)
- github.com/go-chi/chi/v5 (HTTP router)
- github.com/xuri/excelize/v2 (XLSX parsing)
- github.com/stretchr/testify (testing helpers)
- github.com/go-playground/validator/v10 (input validation)
- github.com/golang/mock (mocking tools)

Developer tools (install in your shell; WSL example):

```bash
# install sqlc (code generation)
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# install mockgen (mock generator)
go install github.com/golang/mock/mockgen@latest

# install golangci-lint (linters)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

After installing the tools you can generate code and mocks:

```bash
make generate   # runs sqlc generate and mock generation (if configured)
```

Contributing

Please follow Go idioms and run `golangci-lint` before opening PRs.
