PEMPTY :=
.PHONY: build test test-coverage run-api run-worker migrate clean lint generate sqlc-generate mock-generate

build:
	@echo "Building binaries..."
	@mkdir -p bin
	@echo "Building all cmd/* packages"
	@go build -o bin/admin ./cmd/admin
	@go build -o bin/api ./cmd/api
	@go build -o bin/worker ./cmd/worker
	@go build -o bin/migrate ./cmd/migrate

test:
	@echo "Running tests..."
	@go test -v -cover ./...

test-coverage:
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

run-api:
	@go run ./cmd/api/main.go

run-worker:
	@go run ./cmd/worker/main.go

migrate:
	@go run ./cmd/migrate/main.go up

clean:
	@rm -rf bin/
	@rm -f coverage.out coverage.html

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

generate: sqlc-generate mock-generate
	@echo "Code generation complete"
