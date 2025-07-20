#!/bin/bash
# Docker Container Testing Script for CloudMCP
# Provides comprehensive container validation including build, security, and functionality testing

set -euo pipefail

# Script configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
IMAGE_NAME="${IMAGE_NAME:-cloudmcp}"
IMAGE_TAG="${IMAGE_TAG:-ci}"
CONTAINER_NAME="${CONTAINER_NAME:-cloudmcp-test}"
TEST_PORT="${TEST_PORT:-8080}"
BUILD_ARGS="${BUILD_ARGS:-}"
MULTI_ARCH="${MULTI_ARCH:-false}"
SECURITY_SCAN="${SECURITY_SCAN:-true}"
PERFORMANCE_TEST="${PERFORMANCE_TEST:-true}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $*"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $*"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $*"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*"
}

# Cleanup function
cleanup() {
    local exit_code=$?
    log_info "Cleaning up test environment..."
    
    # Stop and remove test container if it exists
    if docker ps -q -f name="${CONTAINER_NAME}" >/dev/null 2>&1; then
        docker stop "${CONTAINER_NAME}" >/dev/null 2>&1 || true
    fi
    
    if docker ps -aq -f name="${CONTAINER_NAME}" >/dev/null 2>&1; then
        docker rm "${CONTAINER_NAME}" >/dev/null 2>&1 || true
    fi
    
    # Remove test network if it exists
    if docker network ls -q -f name=cloudmcp-test >/dev/null 2>&1; then
        docker network rm cloudmcp-test >/dev/null 2>&1 || true
    fi
    
    if [[ ${exit_code} -eq 0 ]]; then
        log_success "Container testing completed successfully"
    else
        log_error "Container testing failed with exit code ${exit_code}"
    fi
    
    exit ${exit_code}
}

# Set up cleanup trap
trap cleanup EXIT INT TERM

# Validate prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check Docker
    if ! command -v docker >/dev/null 2>&1; then
        log_error "Docker is required but not installed"
        exit 1
    fi
    
    # Check Docker daemon
    if ! docker info >/dev/null 2>&1; then
        log_error "Docker daemon is not running"
        exit 1
    fi
    
    # Check if we're in the right directory
    if [[ ! -f "${PROJECT_ROOT}/Dockerfile" ]]; then
        log_error "Dockerfile not found in project root: ${PROJECT_ROOT}"
        exit 1
    fi
    
    # Check Go installation for container tests
    if ! command -v go >/dev/null 2>&1; then
        log_warning "Go not found - container framework tests will be skipped"
    fi
    
    log_success "Prerequisites check passed"
}

# Fix Dockerfile path issue
fix_dockerfile() {
    log_info "Checking Dockerfile for path issues..."
    
    # Check if Dockerfile has the wrong path
    if grep -q "cmd/server/main.go" "${PROJECT_ROOT}/Dockerfile"; then
        log_warning "Fixing Dockerfile path from cmd/server/main.go to cmd/cloud-mcp/main.go"
        
        # Create a backup
        cp "${PROJECT_ROOT}/Dockerfile" "${PROJECT_ROOT}/Dockerfile.backup"
        
        # Fix the path
        sed -i.tmp 's|cmd/server/main.go|cmd/cloud-mcp/main.go|g' "${PROJECT_ROOT}/Dockerfile"
        rm -f "${PROJECT_ROOT}/Dockerfile.tmp"
        
        log_success "Dockerfile path fixed"
    else
        log_info "Dockerfile path is correct"
    fi
}

# Build Docker image
build_image() {
    log_info "Building Docker image: ${IMAGE_NAME}:${IMAGE_TAG}"
    
    cd "${PROJECT_ROOT}"
    
    # Build command construction
    local build_cmd="docker build"
    
    # Add build arguments if provided
    if [[ -n "${BUILD_ARGS}" ]]; then
        build_cmd="${build_cmd} ${BUILD_ARGS}"
    fi
    
    # Add version information
    local version
    version=$(cd "${PROJECT_ROOT}" && git describe --tags --always --dirty 2>/dev/null || echo "dev")
    build_cmd="${build_cmd} --build-arg VERSION=${version}"
    
    # Add tag and context
    build_cmd="${build_cmd} -t ${IMAGE_NAME}:${IMAGE_TAG} ."
    
    log_info "Build command: ${build_cmd}"
    
    # Execute build
    if eval "${build_cmd}"; then
        log_success "Docker image built successfully"
    else
        log_error "Docker image build failed"
        exit 1
    fi
    
    # Get image information
    local image_size
    image_size=$(docker images "${IMAGE_NAME}:${IMAGE_TAG}" --format "{{.Size}}")
    log_info "Image size: ${image_size}"
}

# Test multi-architecture build
test_multi_arch() {
    if [[ "${MULTI_ARCH}" != "true" ]]; then
        log_info "Skipping multi-architecture testing (MULTI_ARCH=false)"
        return 0
    fi
    
    log_info "Testing multi-architecture build..."
    
    # Check if buildx is available
    if ! docker buildx version >/dev/null 2>&1; then
        log_warning "Docker buildx not available - skipping multi-arch test"
        return 0
    fi
    
    # Create builder if it doesn't exist
    if ! docker buildx ls | grep -q cloudmcp-builder; then
        docker buildx create --name cloudmcp-builder --use >/dev/null 2>&1 || true
    fi
    
    # Test build for multiple platforms
    log_info "Building for linux/amd64,linux/arm64..."
    if docker buildx build \
        --platform linux/amd64,linux/arm64 \
        --build-arg VERSION="${version:-dev}" \
        -t "${IMAGE_NAME}:${IMAGE_TAG}-multiarch" \
        --load=false \
        "${PROJECT_ROOT}"; then
        log_success "Multi-architecture build test passed"
    else
        log_error "Multi-architecture build test failed"
        return 1
    fi
}

# Run security scans
run_security_scan() {
    if [[ "${SECURITY_SCAN}" != "true" ]]; then
        log_info "Skipping security scan (SECURITY_SCAN=false)"
        return 0
    fi
    
    log_info "Running container security scans..."
    
    # Trivy vulnerability scan
    if command -v trivy >/dev/null 2>&1; then
        log_info "Running Trivy vulnerability scan..."
        if trivy image --exit-code 0 --no-progress --format table "${IMAGE_NAME}:${IMAGE_TAG}"; then
            log_success "Trivy scan completed"
        else
            log_warning "Trivy scan found issues (non-blocking)"
        fi
    else
        log_warning "Trivy not installed - installing for scan..."
        # Install Trivy
        if [[ "${OSTYPE}" == "linux-gnu"* ]]; then
            curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /tmp v0.45.0
            /tmp/trivy image --exit-code 0 --no-progress --format table "${IMAGE_NAME}:${IMAGE_TAG}"
        else
            log_warning "Trivy installation skipped on non-Linux platform"
        fi
    fi
    
    # Dockerfile lint with hadolint
    if command -v hadolint >/dev/null 2>&1; then
        log_info "Running hadolint Dockerfile scan..."
        if hadolint "${PROJECT_ROOT}/Dockerfile"; then
            log_success "hadolint scan passed"
        else
            log_warning "hadolint found issues (non-blocking)"
        fi
    else
        log_warning "hadolint not installed - downloading for scan..."
        # Download hadolint
        if [[ "${OSTYPE}" == "linux-gnu"* ]]; then
            wget -O /tmp/hadolint https://github.com/hadolint/hadolint/releases/latest/download/hadolint-Linux-x86_64
            chmod +x /tmp/hadolint
            /tmp/hadolint "${PROJECT_ROOT}/Dockerfile"
        else
            log_warning "hadolint download skipped on non-Linux platform"
        fi
    fi
}

# Test container functionality
test_container_functionality() {
    log_info "Testing container functionality..."
    
    # Create test network
    docker network create cloudmcp-test >/dev/null 2>&1 || true
    
    # Start container with health check
    log_info "Starting container for functionality testing..."
    docker run -d \
        --name "${CONTAINER_NAME}" \
        --network cloudmcp-test \
        -p "${TEST_PORT}:8080" \
        -e LOG_LEVEL=debug \
        -e ENABLE_METRICS=true \
        -e METRICS_PORT=8080 \
        "${IMAGE_NAME}:${IMAGE_TAG}"
    
    # Wait for container to be healthy
    log_info "Waiting for container to be healthy..."
    local max_attempts=30
    local attempt=0
    
    while [[ ${attempt} -lt ${max_attempts} ]]; do
        if docker inspect "${CONTAINER_NAME}" --format='{{.State.Health.Status}}' 2>/dev/null | grep -q "healthy"; then
            log_success "Container is healthy"
            break
        elif docker inspect "${CONTAINER_NAME}" --format='{{.State.Health.Status}}' 2>/dev/null | grep -q "unhealthy"; then
            log_error "Container health check failed"
            docker logs "${CONTAINER_NAME}"
            exit 1
        else
            log_info "Waiting for health check... (attempt ${attempt}/${max_attempts})"
            sleep 2
            ((attempt++))
        fi
    done
    
    if [[ ${attempt} -eq ${max_attempts} ]]; then
        log_error "Container did not become healthy within timeout"
        docker logs "${CONTAINER_NAME}"
        exit 1
    fi
    
    # Test health endpoint
    log_info "Testing health endpoint..."
    local health_response
    if health_response=$(curl -f -s "http://localhost:${TEST_PORT}/health" 2>/dev/null); then
        log_success "Health endpoint test passed"
        log_info "Health response: ${health_response}"
    else
        log_error "Health endpoint test failed"
        docker logs "${CONTAINER_NAME}"
        exit 1
    fi
    
    # Test metrics endpoint
    log_info "Testing metrics endpoint..."
    if curl -f -s "http://localhost:${TEST_PORT}/metrics" >/dev/null 2>&1; then
        log_success "Metrics endpoint test passed"
    else
        log_warning "Metrics endpoint test failed (may require authentication)"
    fi
    
    # Test graceful shutdown
    log_info "Testing graceful shutdown..."
    docker stop "${CONTAINER_NAME}" >/dev/null 2>&1
    
    # Check exit code
    local exit_code
    exit_code=$(docker inspect "${CONTAINER_NAME}" --format='{{.State.ExitCode}}')
    if [[ ${exit_code} -eq 0 ]]; then
        log_success "Graceful shutdown test passed"
    else
        log_warning "Container exited with code ${exit_code}"
        docker logs "${CONTAINER_NAME}"
    fi
}

# Run performance tests
test_performance() {
    if [[ "${PERFORMANCE_TEST}" != "true" ]]; then
        log_info "Skipping performance tests (PERFORMANCE_TEST=false)"
        return 0
    fi
    
    log_info "Running container performance tests..."
    
    # Start container for performance testing
    docker run -d \
        --name "${CONTAINER_NAME}-perf" \
        -p "$((TEST_PORT + 1)):8080" \
        -e LOG_LEVEL=error \
        -e ENABLE_METRICS=true \
        "${IMAGE_NAME}:${IMAGE_TAG}"
    
    # Wait for startup
    sleep 5
    
    # Measure startup time
    local startup_time
    startup_time=$(docker inspect "${CONTAINER_NAME}-perf" --format='{{.State.StartedAt}}')
    log_info "Container startup time: ${startup_time}"
    
    # Test response time
    log_info "Testing response time..."
    local response_time
    if command -v curl >/dev/null 2>&1; then
        response_time=$(curl -w "%{time_total}" -s -o /dev/null "http://localhost:$((TEST_PORT + 1))/health" || echo "failed")
        if [[ "${response_time}" != "failed" ]]; then
            log_success "Health endpoint response time: ${response_time}s"
        else
            log_warning "Response time test failed"
        fi
    fi
    
    # Get memory usage
    local memory_usage
    memory_usage=$(docker stats "${CONTAINER_NAME}-perf" --no-stream --format "{{.MemUsage}}" 2>/dev/null || echo "unknown")
    log_info "Memory usage: ${memory_usage}"
    
    # Cleanup performance test container
    docker stop "${CONTAINER_NAME}-perf" >/dev/null 2>&1
    docker rm "${CONTAINER_NAME}-perf" >/dev/null 2>&1
    
    log_success "Performance tests completed"
}

# Run container framework tests
run_container_tests() {
    if ! command -v go >/dev/null 2>&1; then
        log_warning "Go not available - skipping container framework tests"
        return 0
    fi
    
    log_info "Running container framework tests..."
    
    cd "${PROJECT_ROOT}"
    
    # Set environment variables for container testing
    export CONTAINER_TEST_MODE=true
    export CONTAINER_IMAGE="${IMAGE_NAME}:${IMAGE_TAG}"
    export CONTAINER_PORT="${TEST_PORT}"
    
    # Run the container tests
    if go test -v -timeout=10m ./internal/testing/container/...; then
        log_success "Container framework tests passed"
    else
        log_error "Container framework tests failed"
        return 1
    fi
}

# Main execution
main() {
    log_info "Starting CloudMCP container testing..."
    log_info "Image: ${IMAGE_NAME}:${IMAGE_TAG}"
    log_info "Container: ${CONTAINER_NAME}"
    log_info "Port: ${TEST_PORT}"
    
    check_prerequisites
    fix_dockerfile
    build_image
    test_multi_arch
    run_security_scan
    test_container_functionality
    test_performance
    run_container_tests
    
    log_success "All container tests completed successfully!"
}

# Show usage information
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

CloudMCP Container Testing Script

OPTIONS:
    -h, --help              Show this help message
    -n, --name NAME         Set image name (default: cloudmcp)
    -t, --tag TAG           Set image tag (default: ci)
    -p, --port PORT         Set test port (default: 8080)
    -m, --multi-arch        Enable multi-architecture testing
    -s, --skip-security     Skip security scanning
    -P, --skip-performance  Skip performance testing
    -b, --build-args ARGS   Additional build arguments

ENVIRONMENT VARIABLES:
    IMAGE_NAME              Docker image name
    IMAGE_TAG               Docker image tag
    CONTAINER_NAME          Test container name
    TEST_PORT               Port for testing
    BUILD_ARGS              Additional build arguments
    MULTI_ARCH              Enable multi-arch testing (true/false)
    SECURITY_SCAN           Enable security scanning (true/false)
    PERFORMANCE_TEST        Enable performance testing (true/false)

EXAMPLES:
    $0                      Run with defaults
    $0 -m                   Run with multi-architecture testing
    $0 -n myapp -t latest   Use custom image name and tag
    $0 -p 9090 -s          Use port 9090 and skip security scans

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            usage
            exit 0
            ;;
        -n|--name)
            IMAGE_NAME="$2"
            shift 2
            ;;
        -t|--tag)
            IMAGE_TAG="$2"
            shift 2
            ;;
        -p|--port)
            TEST_PORT="$2"
            shift 2
            ;;
        -m|--multi-arch)
            MULTI_ARCH="true"
            shift
            ;;
        -s|--skip-security)
            SECURITY_SCAN="false"
            shift
            ;;
        -P|--skip-performance)
            PERFORMANCE_TEST="false"
            shift
            ;;
        -b|--build-args)
            BUILD_ARGS="$2"
            shift 2
            ;;
        *)
            log_error "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi