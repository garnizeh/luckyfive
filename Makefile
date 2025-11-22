PEMPTY :=
.PHONY: build test test-coverage run-api run-worker migrate clean lint generate sqlc-generate mock-generate swagger-generate reset-db
 
.PHONY: import-quina

build:
	@echo "Building binaries..."
	@mkdir -p bin
	@echo "Building all cmd/* packages"
	@go build -o bin/api ./cmd/api
	@go build -o bin/import ./cmd/import
	@go build -o bin/migrate ./cmd/migrate
	@go build -o bin/worker ./cmd/worker

test:
	@echo "Running tests..."
	@go test -v -cover ./...

test-coverage:
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

# Run tests with coverage across all packages (includes coverage for dependent packages)
test-coverage-all:
	@echo "Running tests with coverage across all packages..."
	@go test ./... -covermode=atomic -coverpkg=./... -coverprofile=coverage.out
	@go tool cover -func=coverage.out | tee coverage.txt
	@go tool cover -html=coverage.out -o coverage.html

run-api:
	@go run ./cmd/api/main.go --env-file=configs/dev.env

run-worker:
	@go run ./cmd/worker/main.go

migrate:
	@go run ./cmd/migrate/main.go --env-file=configs/dev.env up

clean:
	@rm -rf bin/
	@rm -f coverage.out coverage.html coverage.txt

lint:
	@golangci-lint run

sqlc-generate:
	@echo "Generating sqlc code..."
	@sqlc generate

mock-generate: sqlc-generate
	@echo "Generating mocks..."
	@mockgen -source=internal/store/results/querier.go -destination=internal/store/results/mock/querier.go -package=mock
	@mockgen -source=internal/store/simulations/querier.go -destination=internal/store/simulations/mock/querier.go -package=mock
	@mockgen -source=internal/store/configs/querier.go -destination=internal/store/configs/mock/querier.go -package=mock
	@mockgen -source=internal/store/finances/querier.go -destination=internal/store/finances/mock/querier.go -package=mock
	@mockgen -source=internal/services/simulation.go -destination=internal/services/mock/simulation_service.go -package=mock

swagger-generate:
	@echo "Generating OpenAPI/Swagger docs (if swag CLI is available)";
	@if command -v swag >/dev/null 2>&1; then swag init -g cmd/api/main.go --parseInternal -o api || echo "swag init failed"; else echo "swag CLI not found, skipping swagger generation (install with 'go install github.com/swaggo/swag/cmd/swag@latest')"; fi

generate: sqlc-generate mock-generate swagger-generate
	@echo "Code generation complete"

# Reset databases: remove DB files, regenerate code and run migrations
reset-db:
	@echo "Resetting databases..."
	@mkdir -p data/db
	@rm -f data/db/*.db || true
	@$(MAKE) generate
	@$(MAKE) migrate

# Run import CLI using the canonical Quina XLSX and dev env
import-quina:
	@echo "Importing data/results/Quina.xlsx using configs/dev.env"
	@./bin/import --env-file=configs/dev.env --xlsx=data/results/Quina.xlsx
