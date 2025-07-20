.PHONY: help build run test test-race test-integration clean lint fmt install-tools

# Default target
help:
	@echo "CloudMCP Development Commands:"
	@echo ""
	@echo "Build Commands:"
	@echo "  make build           - Build the cloud-mcp binary (development)"
	@echo "  make build-prod      - Build optimized binary (production)"
	@echo "  make build-all       - Build for multiple platforms"
	@echo "  make docker-build    - Build Docker image"
	@echo "  make docker-run      - Run Docker container"
	@echo "  make run             - Run the server (requires .env)"
	@echo ""
	@echo "Testing Commands:"
	@echo "  make test            - Run unit tests only (no integration)"
	@echo "  make test-quick      - Run quick tests (fast packages only)"
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
	@echo "  make analyze         - Analyze dependencies and binary size"
	@echo ""
	@echo "Setup Commands:"
	@echo "  make clean           - Clean build artifacts"
	@echo "  make install-tools   - Install development tools"

# Build binary (development - fast build)
build:
	@echo "Building cloud-mcp (development)..."
	@go build -o bin/cloud-mcp cmd/cloud-mcp/main.go

# Build optimized binary (production - smaller, faster)
build-prod:
	@echo "Building optimized cloud-mcp (production)..."
	@go build -ldflags="-s -w" -trimpath -o bin/cloud-mcp cmd/cloud-mcp/main.go
	@echo "Optimized build complete!"
	@ls -lah bin/cloud-mcp

# Run server
run: build
	@echo "Running cloud-mcp..."
	@./bin/cloud-mcp

# Run unit tests only (excludes integration tests)
test:
	@echo "Running unit tests only..."
	@go test -v ./... -short

# Quick test (fast packages only)
test-quick:
	@echo "Running quick tests (fast packages only)..."
	@go test -short -parallel 8 ./pkg/... ./internal/config/... ./internal/version/...

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

# Run linters (following Go rules)
lint:
	@echo "Running linters..."
	@echo "• Running golangci-lint with full ruleset..."
	@golangci-lint run
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

# Build for multiple platforms (optimized)
.PHONY: build-all
build-all:
	@echo "Building for multiple platforms (optimized)..."
	@mkdir -p dist
	@GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o dist/cloud-mcp-linux-amd64 cmd/cloud-mcp/main.go
	@GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o dist/cloud-mcp-darwin-amd64 cmd/cloud-mcp/main.go
	@GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -trimpath -o dist/cloud-mcp-darwin-arm64 cmd/cloud-mcp/main.go
	@GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o dist/cloud-mcp-windows-amd64.exe cmd/cloud-mcp/main.go
	@echo "Build complete. Optimized binaries in dist/"
	@ls -lah dist/

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


# Build Docker image
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	@docker build -t cloudmcp:latest .
	@echo "Docker image built successfully!"
	@docker images cloudmcp:latest

# Run Docker container
.PHONY: docker-run
docker-run:
	@echo "Running CloudMCP in Docker..."
	@docker run --rm -it \
		-p 8080:8080 \
		-e LOG_LEVEL=debug \
		-e ENABLE_METRICS=true \
		--name cloudmcp-container \
		cloudmcp:latest

# Analyze dependencies and binary size
.PHONY: analyze
analyze: build-prod
	@echo "=== Dependency Analysis ==="
	@echo "Total dependencies for main binary:"
	@go list -deps ./cmd/cloud-mcp | wc -l
	@echo "Total modules in dependency graph:"
	@go mod graph | wc -l
	@echo ""
	@echo "=== Binary Size Analysis ==="
	@echo "Development build:"
	@ls -lah bin/cloud-mcp 2>/dev/null || echo "  (Run 'make build' first)"
	@echo "Production build:"
	@ls -lah bin/cloud-mcp
	@echo ""
	@echo "Binary details:"
	@file bin/cloud-mcp
	@echo ""
	@echo "=== Large Dependencies ==="
	@echo "Top 10 largest modules by disk usage:"
	@go mod download -json | jq -r '.Path + " " + .Dir' | head -10 2>/dev/null || echo "  (Install jq for detailed analysis)"