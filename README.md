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

Setup guide
----------

Prerequisites
- Go 1.21+ (this project was developed and tested with Go 1.25.x)
- Git
- Optional: sqlite3 for manual inspection of DB files

Environment
- Copy `configs/dev.env` (or `configs/.env.example`) to `.env` or provide an `--env-file` when running the server. The service reads configuration such as database paths and server host/port from env vars via `internal/config`.

Common commands
- Generate code (sqlc + mocks) and OpenAPI docs:

```bash
make generate
```

- Build binaries:

```bash
make build
```

- Run API (development):

```bash
make run-api
# or run directly with 'go run'
go run ./cmd/api/main.go --env-file=configs/dev.env
```

Tests
- Run the full test suite:

```bash
go test ./... -v
```

- Run package-specific tests, e.g. handlers only:

```bash
go test ./internal/handlers -v
```

Swagger / OpenAPI
-----------------
This project uses `swag` (swaggo) to generate OpenAPI artifacts from source annotations. The generated files live under `./api` (or `./docs` depending on your generation target) and the server serves the JSON at `/swagger/doc.json` and a UI at `/swagger/`.

Generate swagger docs (example):

```bash
# install swag CLI if you don't have it
go install github.com/swaggo/swag/cmd/swag@latest

# from project root
swag init -g cmd/api/main.go -o api
```

Then start the server and open:

	http://localhost:8080/swagger/

Notes & Troubleshooting
- If `swag init` emits a warning like "no Go files in /path" it usually means the CLI tried to inspect a directory without Go files (for example the repo root). This is harmless when you point `-g` to `cmd/api/main.go` — the generator will still scan subpackages and produce operations. See `Makefile` target `swagger-generate` which wraps a recommended invocation.
- If code generation (`make generate`) fails, ensure `sqlc`, `mockgen` and `swag` are installed and on your PATH.

Contributing
------------
Follow project conventions:

- Use Go idioms and keep functions small and testable
- Run linters and tests locally before opening PRs:

```bash
golangci-lint run
go test ./... -v
```

If you modify SQL files, run `make generate` to regenerate sqlc code and mocks.

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

Swagger / OpenAPI (swag)

This project can generate Swagger/OpenAPI documentation using `swag` (https://github.com/swaggo/swag) and serves a Swagger UI at `/swagger/` when the generated JSON is available under `./docs/swagger.json`.

To generate docs locally:

```bash
# install the swag CLI (if you haven't already)
go install github.com/swaggo/swag/cmd/swag@latest

# from project root, generate docs (scans annotations in cmd/api/main.go and handlers)
swag init -g cmd/api/main.go -o docs
```

Start the server and open:

	http://localhost:8080/swagger/index.html

The server exposes the generated JSON at `/swagger/doc.json` and the UI at `/swagger/`.
