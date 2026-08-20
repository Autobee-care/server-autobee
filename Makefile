.PHONY: dev build test test-coverage lint format swagger docker-up docker-down seed clean

BINARY_NAME=autobee-server
BINARY_PATH=bin/$(BINARY_NAME)
MAIN_PATH=./cmd/api
COVERAGE_OUT=coverage.out
COVERAGE_HTML=coverage.html
GO=go
GOPATH_BIN=$(shell $(GO) env GOPATH)/bin

## ─── Development ─────────────────────────────────────────────────────────────

# Run the server locally (requires .env file)
dev:
	$(GO) run $(MAIN_PATH)

# Build the binary
build:
	mkdir -p bin
	$(GO) build -ldflags="-w -s" -o $(BINARY_PATH) $(MAIN_PATH)

## ─── Testing ──────────────────────────────────────────────────────────────────

# Run all unit tests (excludes integration tests — no Docker needed)
test:
	$(GO) test ./... -short -count=1 -timeout=60s

# Run integration tests only (requires Docker for Testcontainers)
test-integration:
	$(GO) test -tags=integration ./tests/integration/... -v -count=1 -timeout=180s

# Run all tests with coverage (unit only, no integration)
test-coverage:
	$(GO) test ./... -coverprofile=$(COVERAGE_OUT) -covermode=atomic -count=1 -timeout=120s
	$(GO) tool cover -html=$(COVERAGE_OUT) -o $(COVERAGE_HTML)
	$(GO) tool cover -func=$(COVERAGE_OUT) | tail -n 1

## ─── Code Quality ─────────────────────────────────────────────────────────────

# Run golangci-lint
lint:
	golangci-lint run ./...

# Format all Go files with gofmt and goimports
format:
	gofmt -w -s .
	goimports -w .

# Check formatting (used in CI)
format-check:
	@output=$$(gofmt -l .); \
	if [ -n "$$output" ]; then \
		echo "The following files are not formatted:"; \
		echo "$$output"; \
		exit 1; \
	fi

## ─── Swagger ─────────────────────────────────────────────────────────────────

# Generate Swagger docs
swagger:
	@which swag > /dev/null 2>&1 && swag init -g cmd/api/main.go --output docs || $(GOPATH_BIN)/swag init -g cmd/api/main.go --output docs

## ─── Docker ──────────────────────────────────────────────────────────────────

# Start all Docker services
docker-up:
	docker-compose up -d

# Stop all Docker services
docker-down:
	docker-compose down

# Stop and remove volumes
docker-clean:
	docker-compose down -v

## ─── Database ────────────────────────────────────────────────────────────────

# Run development seed script
seed:
	$(GO) run scripts/seed.go

## ─── Utilities ───────────────────────────────────────────────────────────────

# Remove build artifacts
clean:
	rm -rf bin/ $(COVERAGE_OUT) $(COVERAGE_HTML)

# Install local dev tools
tools:
	$(GO) install github.com/swaggo/swag/cmd/swag@latest
	$(GO) install golang.org/x/tools/cmd/goimports@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH_BIN) latest

# Print help
help:
	@echo "Available targets:"
	@echo "  dev              Run server locally"
	@echo "  build            Build binary to bin/"
	@echo "  test             Run unit tests"
	@echo "  test-integration Run integration tests (needs Docker)"
	@echo "  test-coverage    Run all tests with coverage report"
	@echo "  lint             Run golangci-lint"
	@echo "  format           Format code (gofmt + goimports)"
	@echo "  format-check     Check formatting (CI)"
	@echo "  swagger          Generate Swagger docs"
	@echo "  docker-up        Start Docker services"
	@echo "  docker-down      Stop Docker services"
	@echo "  seed             Run dev seed script"
	@echo "  clean            Remove build artifacts"
	@echo "  tools            Install dev tools"
