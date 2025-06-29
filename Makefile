.PHONY: help build run test test-race test-integration clean lint fmt install-tools setup-mcp

# Default target
help:
	@echo "CloudMCP Development Commands:"
	@echo ""
	@echo "Build Commands:"
	@echo "  make build           - Build the cloud-mcp binary"
	@echo "  make build-all       - Build for multiple platforms"
	@echo "  make run             - Run the server (requires .env)"
	@echo ""
	@echo "Testing Commands:"
	@echo "  make test            - Run unit tests only (no integration)"
	@echo "  make test-unit       - Run unit tests only (alias for test)"
	@echo "  make test-integration - Run integration tests only (mock-based)"
	@echo "  make test-all        - Run all tests (unit + integration)"
	@echo "  make test-race       - Run tests with race detector"
	@echo "  make coverage        - Generate test coverage report"
	@echo ""
	@echo "Development Tools:"
	@echo "  make test-client     - Run interactive test client"
	@echo "  make test-stdio      - Test with stdio commands"
	@echo "  make lint            - Run golangci-lint"
	@echo "  make fmt             - Format code with gofumpt"
	@echo "  make tidy            - Tidy and verify dependencies"
	@echo ""
	@echo "Setup Commands:"
	@echo "  make clean           - Clean build artifacts"
	@echo "  make install-tools   - Install development tools"
	@echo "  make setup-mcp       - Setup CloudMCP for Claude Desktop and Claude Code"

# Build binary
build:
	@echo "Building cloud-mcp..."
	@go build -o bin/cloud-mcp cmd/server/main.go
	@echo "Building cloud-mcp-setup..."
	@go build -o bin/cloud-mcp-setup cmd/cloud-mcp-setup/main.go

# Run server
run: build
	@echo "Running cloud-mcp..."
	@./bin/cloud-mcp

# Run unit tests only (excludes integration tests)
test:
	@echo "Running unit tests only..."
	@go test -v ./... -short

# Alias for unit tests
test-unit:
	@echo "Running unit tests only..."
	@go test -v ./... -short

# Run integration tests only (mock-based, no live APIs)
test-integration:
	@echo "Running integration tests (mock-based)..."
	@go test -v ./... -tags=integration -run="Integration"

# Run all tests (unit + integration)
test-all:
	@echo "Running all tests (unit + integration)..."
	@go test -race -v ./...

# Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	@go test -race ./...

# Run unit tests with race detector only
test-race-unit:
	@echo "Running unit tests with race detector..."
	@go test -race -v ./... -short

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/ dist/ coverage.txt coverage.html *.coverprofile

# Run linters (using reliable approach)
lint:
	@echo "Running linters..."
	@echo "• Running go vet..."
	@go vet ./...
	@echo "• Running go fmt check..."
	@test -z "$$(gofmt -l .)" || (echo "Code needs formatting. Run 'make fmt'" && exit 1)
	@echo "• Running staticcheck for deprecation warnings..."
	@golangci-lint run --disable-all --enable=staticcheck,gosimple,unused,ineffassign,misspell,unconvert --exclude-use-default=false
	@echo "✅ All linting checks passed!"

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
	@go install github.com/testcontainers/testcontainers-go@latest
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

# Setup CloudMCP for Claude Desktop and Claude Code
.PHONY: setup-mcp
setup-mcp: build
	@echo "Setting up CloudMCP for Claude..."
	@./bin/cloud-mcp-setup -local