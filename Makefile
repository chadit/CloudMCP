.PHONY: help build run test test-race test-integration clean lint fmt install-tools

# Default target
help:
	@echo "CloudMCP Development Commands:"
	@echo "  make build           - Build the cloud-mcp binary"
	@echo "  make run             - Run the server (requires .env)"
	@echo "  make test            - Run unit tests"
	@echo "  make test-race       - Run tests with race detector"
	@echo "  make test-integration - Run integration tests (requires LINODE_TEST_TOKEN)"
	@echo "  make test-client     - Run interactive test client"
	@echo "  make test-stdio      - Test with stdio commands"
	@echo "  make lint            - Run golangci-lint"
	@echo "  make fmt             - Format code with gofumpt"
	@echo "  make clean           - Clean build artifacts"
	@echo "  make install-tools   - Install development tools"

# Build binary
build:
	@echo "Building cloud-mcp..."
	@go build -o bin/cloud-mcp cmd/server/main.go

# Run server
run: build
	@echo "Running cloud-mcp..."
	@./bin/cloud-mcp

# Run tests
test:
	@echo "Running tests..."
	@go test ./...

# Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	@go test -race ./...

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	@go test ./internal/... -tags=integration

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/ dist/ coverage.txt coverage.html *.coverprofile

# Run linters
lint:
	@echo "Running linters..."
	@golangci-lint run

# Format code
fmt:
	@echo "Formatting code..."
	@gofumpt -w .

# Install development tools
install-tools:
	@echo "Installing development tools..."
	@go install mvdan.cc/gofumpt@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/vektra/mockery/v2@latest
	@echo "Tools installed successfully"

# Build for multiple platforms
.PHONY: build-all
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p dist
	@GOOS=linux GOARCH=amd64 go build -o dist/cloud-mcp-linux-amd64 cmd/server/main.go
	@GOOS=darwin GOARCH=amd64 go build -o dist/cloud-mcp-darwin-amd64 cmd/server/main.go
	@GOOS=darwin GOARCH=arm64 go build -o dist/cloud-mcp-darwin-arm64 cmd/server/main.go
	@GOOS=windows GOARCH=amd64 go build -o dist/cloud-mcp-windows-amd64.exe cmd/server/main.go
	@echo "Build complete. Binaries in dist/"

# Run with environment file
.PHONY: run-env
run-env: build
	@if [ -f .env ]; then \
		echo "Loading .env and running..."; \
		set -a; . ./.env; set +a; ./bin/cloud-mcp; \
	else \
		echo "No .env file found. Create one from .env.example"; \
		exit 1; \
	fi

# Generate test coverage
.PHONY: coverage
coverage:
	@echo "Generating test coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Check for security vulnerabilities
.PHONY: security
security:
	@echo "Checking for vulnerabilities..."
	@go list -m all | nancy sleuth || true
	@govulncheck ./... || true

# Tidy dependencies
.PHONY: tidy
tidy:
	@echo "Tidying dependencies..."
	@go mod tidy
	@go mod verify

# Run interactive test client
.PHONY: test-client
test-client: build
	@if [ -f .env ]; then \
		echo "Loading .env and running test client..."; \
		set -a; . ./.env; set +a; go run test/client/main.go; \
	else \
		echo "No .env file found. Create one from .env.example"; \
		exit 1; \
	fi

# Test with stdio commands
.PHONY: test-stdio
test-stdio: build
	@if [ -f .env ]; then \
		echo "Testing with stdio commands..."; \
		set -a; . ./.env; set +a; cat test-commands.json | ./bin/cloud-mcp | jq .; \
	else \
		echo "No .env file found. Create one from .env.example"; \
		exit 1; \
	fi