# CloudMCP Container Testing Guide

## Overview

This document provides comprehensive guidance for CloudMCP container build validation, testing, and deployment readiness validation. The container testing infrastructure ensures CloudMCP deployments are secure, performant, and compliant with MCP protocol standards.

## 🏗️ Container Testing Architecture

### Core Components

1. **Container Testing Framework** (`internal/testing/container/`)
   - Comprehensive container validation
   - MCP protocol compliance testing
   - Performance metrics collection
   - Security scanning integration

2. **Docker Test Script** (`scripts/docker-test.sh`)
   - Automated container build and validation
   - Multi-architecture testing support
   - Security scanning with Trivy and hadolint
   - Performance benchmarking

3. **CI/CD Integration** (`.github/workflows/ci.yml`)
   - Multi-platform container builds (amd64, arm64)
   - Comprehensive security scanning
   - Health check validation
   - Performance metrics collection

4. **Docker Compose Testing** (`docker-compose.test.yml`)
   - Complete testing environment
   - Security scanner service
   - Performance testing with k6
   - Test result collection

## 🚀 Quick Start

### Running Container Tests Locally

```bash
# Run comprehensive container validation
./scripts/docker-test.sh

# Run with custom configuration
./scripts/docker-test.sh \
  --name cloudmcp \
  --tag test \
  --port 8080 \
  --multi-arch \
  --skip-security

# Run specific test scenarios
./scripts/docker-test.sh --skip-performance  # Skip performance tests
./scripts/docker-test.sh --skip-security     # Skip security scans
```

### Using Docker Compose for Testing

```bash
# Run complete test suite
docker-compose -f docker-compose.test.yml up --build

# Run specific services
docker-compose -f docker-compose.test.yml up cloudmcp security-scanner

# View test results
docker-compose -f docker-compose.test.yml logs test-runner
```

### Manual Container Testing

```bash
# Build container
docker build -t cloudmcp:test .

# Run container
docker run -d -p 8080:8080 \
  -e LOG_LEVEL=debug \
  -e ENABLE_METRICS=true \
  --name cloudmcp-test \
  cloudmcp:test

# Test health endpoint
curl http://localhost:8080/health

# Check container health
docker inspect cloudmcp-test --format='{{.State.Health.Status}}'
```

## 🔧 Testing Framework Usage

### Go Container Tests

```go
package main

import (
    "context"
    "time"
    
    "github.com/chadit/CloudMCP/internal/testing/container"
    "github.com/chadit/CloudMCP/pkg/logger"
)

func main() {
    log := logger.New(logger.LogConfig{Level: "debug"})
    tf := container.NewTestFramework(log, "cloudmcp:test", "http://localhost:8080")
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()
    
    suite, err := tf.RunTestSuite(ctx, "cloudmcp:test")
    if err != nil {
        log.Error("Test suite failed", "error", err)
        return
    }
    
    log.Info("Test suite completed",
        "total", suite.TotalTests,
        "passed", suite.PassedTests,
        "failed", suite.FailedTests,
    )
}
```

### Environment Variables

```bash
# Container test configuration
export CONTAINER_TEST_MODE=true
export CONTAINER_IMAGE=cloudmcp:test
export CONTAINER_PORT=8080

# Docker test script configuration
export IMAGE_NAME=cloudmcp
export IMAGE_TAG=test
export MULTI_ARCH=true
export SECURITY_SCAN=true
export PERFORMANCE_TEST=true
```

## 🧪 Test Categories

### 1. Container Build Validation

**Purpose**: Ensure containers build successfully across platforms

**Tests**:
- Multi-architecture builds (amd64, arm64)
- Dockerfile optimization validation
- Build argument handling
- Layer caching efficiency

**Example**:
```bash
# Test multi-arch build
docker buildx build --platform linux/amd64,linux/arm64 -t cloudmcp:multiarch .
```

### 2. Security Scanning

**Purpose**: Validate container security and compliance

**Tools**:
- **Trivy**: Vulnerability scanning
- **hadolint**: Dockerfile linting
- **Custom scans**: Configuration validation

**Tests**:
- Base image vulnerabilities
- Dependency security issues
- Dockerfile best practices
- Configuration security

**Example**:
```bash
# Run Trivy scan
trivy image --severity HIGH,CRITICAL cloudmcp:test

# Run hadolint
hadolint Dockerfile
```

### 3. Functionality Testing

**Purpose**: Verify container behaves correctly in runtime environment

**Tests**:
- Container startup validation
- Health check functionality
- Signal handling (SIGTERM, SIGINT)
- Resource limit compliance
- Port accessibility

**Example**:
```bash
# Test health check
docker run -d --name test-container cloudmcp:test
sleep 10
curl -f http://localhost:8080/health || exit 1
```

### 4. MCP Protocol Validation

**Purpose**: Ensure MCP protocol compliance in containerized environment

**Tests**:
- Protocol version compliance
- Tool availability and functionality
- Request/response validation
- Error handling
- Performance within MCP constraints

**Example**:
```bash
# Run MCP compliance tests
go test -v ./internal/testing/mcp/... \
  -args -test.server-url=http://localhost:8080
```

### 5. Performance Testing

**Purpose**: Validate container performance characteristics

**Metrics**:
- Startup time
- Memory usage
- CPU utilization
- Response time
- Throughput

**Tests**:
- Load testing with k6
- Resource consumption monitoring
- Stress testing
- Sustained load validation

**Example**:
```bash
# Run k6 performance tests
k6 run test/performance/load-test.js
```

## 📊 Test Results and Reporting

### Test Result Structure

```json
{
  "name": "CloudMCP Container Validation",
  "container_tag": "cloudmcp:test",
  "total_tests": 8,
  "passed_tests": 7,
  "failed_tests": 1,
  "duration": "2m30s",
  "timestamp": "2024-01-15T10:30:00Z",
  "results": [
    {
      "test_name": "container_startup",
      "passed": true,
      "duration": "5s",
      "details": {
        "startup_time": "3s",
        "status": "healthy"
      }
    }
  ]
}
```

### Viewing Test Results

```bash
# View CI test results
gh run view --log  # GitHub CLI

# Local test results
cat test-results/container-validation.json | jq '.'

# Performance results
cat performance-results/k6-results.json | jq '.metrics'
```

## 🔒 Security Considerations

### Container Security Features

1. **Non-root execution**: Container runs as non-privileged user
2. **Read-only filesystem**: Where applicable
3. **Resource limits**: Memory and CPU constraints
4. **Security headers**: Proper HTTP security headers
5. **Minimal attack surface**: Alpine base image with minimal packages

### Security Testing

```bash
# Comprehensive security scan
./scripts/docker-test.sh --security-scan

# Manual security validation
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
  aquasec/trivy image cloudmcp:test

# Check for hardcoded secrets
grep -r "password\|token\|key" . --exclude-dir=.git
```

## ⚡ Performance Optimization

### Container Size Optimization

- Multi-stage builds
- Minimal base images (Alpine)
- Layer caching strategies
- .dockerignore optimization

### Runtime Performance

- Health check tuning
- Resource limit optimization
- Startup time minimization
- Memory usage optimization

### Performance Targets

| Metric | Target | Validation |
|--------|--------|------------|
| Startup Time | < 30s | Automated test |
| Memory Usage | < 100MB | Container stats |
| Response Time | < 1s | Load testing |
| Image Size | < 50MB | Build validation |
| Health Check | < 500ms | CI validation |

## 🔧 Troubleshooting

### Common Issues

1. **Build Failures**
   ```bash
   # Check Dockerfile syntax
   hadolint Dockerfile
   
   # Verify build context
   docker build --no-cache -t debug .
   ```

2. **Health Check Failures**
   ```bash
   # Check container logs
   docker logs container-name
   
   # Test health endpoint manually
   curl -v http://localhost:8080/health
   ```

3. **Performance Issues**
   ```bash
   # Monitor container resources
   docker stats container-name
   
   # Profile application
   go tool pprof http://localhost:8080/debug/pprof/profile
   ```

### Debug Mode

```bash
# Run tests with debug output
CONTAINER_TEST_DEBUG=true ./scripts/docker-test.sh

# Run container with debug logging
docker run -e LOG_LEVEL=debug cloudmcp:test
```

## 📋 CI/CD Integration

### GitHub Actions

The CI pipeline automatically runs container validation:

1. **Multi-architecture builds**: Linux amd64 and arm64
2. **Security scanning**: Trivy and hadolint
3. **Functionality testing**: Health checks and endpoints
4. **Performance validation**: Load testing and metrics
5. **MCP compliance**: Protocol validation

### Workflow Triggers

- Push to main/develop branches
- Pull requests
- Manual workflow dispatch
- Release tags

### Artifact Collection

- Security scan reports
- Performance test results
- Container test logs
- Build artifacts

## 🎯 Best Practices

### Container Development

1. **Use multi-stage builds** for size optimization
2. **Implement proper health checks** for orchestration
3. **Handle signals gracefully** for clean shutdown
4. **Set resource limits** for predictable behavior
5. **Use non-root users** for security

### Testing

1. **Test locally** before CI/CD
2. **Validate on target platforms** (amd64, arm64)
3. **Include performance tests** in validation
4. **Monitor resource usage** during tests
5. **Document test scenarios** and expected outcomes

### Security

1. **Scan regularly** for vulnerabilities
2. **Update base images** frequently
3. **Minimize attack surface** with minimal images
4. **Validate configurations** for security compliance
5. **Monitor runtime behavior** for anomalies

## 📖 References

- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [Container Security Guide](https://cloud.google.com/architecture/best-practices-for-operating-containers)
- [Kubernetes Security](https://kubernetes.io/docs/concepts/security/)
- [MCP Protocol Specification](https://modelcontextprotocol.io/docs)
- [Performance Testing with k6](https://k6.io/docs/)

## 🤝 Contributing

When adding new container tests:

1. **Follow the test framework patterns** in `internal/testing/container/`
2. **Add comprehensive test coverage** for new functionality
3. **Update CI workflows** if new validation is needed
4. **Document test scenarios** and expected behaviors
5. **Validate locally** before submitting PRs

For questions or issues with container testing, please open an issue or contact the CloudMCP team.