#!/bin/bash

# CloudMCP Artifact Signing Script
# Signs binaries, checksums, and container images using cosign
# Usage: ./scripts/sign-artifacts.sh [OPTIONS]

set -euo pipefail

# Script configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
BUILD_DIR="${PROJECT_ROOT}/build"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
SIGN_BINARIES=true
SIGN_CHECKSUMS=true
SIGN_CONTAINER=false
CONTAINER_TAG=""
VERIFY_SIGNATURES=false
KEYLESS=true
PRIVATE_KEY=""
VERBOSE=false

# Help function
show_help() {
    cat << EOF
CloudMCP Artifact Signing Script

Usage: $0 [OPTIONS]

Options:
    -h, --help              Show this help message
    -b, --binaries          Sign binary artifacts (default: true)
    -c, --checksums         Sign checksums file (default: true)
    -i, --container TAG     Sign container image with specified tag
    -v, --verify            Verify signatures after signing
    -k, --key FILE          Use private key file instead of keyless signing
    --no-keyless            Disable keyless signing (requires private key)
    --verbose               Enable verbose output
    --build-dir DIR         Specify build directory (default: ./build)

Examples:
    # Sign all binaries and checksums (keyless)
    $0 --binaries --checksums

    # Sign container image
    $0 --container cloudmcp:latest

    # Sign with private key and verify
    $0 --key /path/to/private.key --verify

    # Sign everything and verify
    $0 --binaries --checksums --container cloudmcp:v1.0.0 --verify

Environment Variables:
    COSIGN_YES              Set to true to skip confirmation prompts
    COSIGN_PRIVATE_KEY      Path to private key (alternative to --key)
    COSIGN_PASSWORD         Password for private key

Prerequisites:
    - cosign must be installed and available in PATH
    - For keyless signing: GitHub Actions environment or OIDC provider
    - For private key signing: Valid cosign private key

EOF
}

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
    echo -e "${RED}[ERROR]${NC} $*" >&2
}

log_verbose() {
    if [[ "${VERBOSE}" == "true" ]]; then
        echo -e "${BLUE}[VERBOSE]${NC} $*"
    fi
}

# Check if cosign is installed
check_cosign() {
    if ! command -v cosign &> /dev/null; then
        log_error "cosign is not installed or not in PATH"
        log_info "Install cosign: go install github.com/sigstore/cosign/v2/cmd/cosign@latest"
        exit 1
    fi
    
    local cosign_version
    cosign_version=$(cosign version --json 2>/dev/null | grep -o '"GitVersion":"[^"]*' | cut -d'"' -f4 || echo "unknown")
    log_verbose "Using cosign version: ${cosign_version}"
}

# Validate build directory and artifacts
validate_artifacts() {
    if [[ ! -d "${BUILD_DIR}" ]]; then
        log_error "Build directory not found: ${BUILD_DIR}"
        log_info "Run build process first or specify correct --build-dir"
        exit 1
    fi
    
    if [[ "${SIGN_BINARIES}" == "true" ]]; then
        local binary_count
        binary_count=$(find "${BUILD_DIR}" -name "cloud-mcp-*" -type f | wc -l)
        if [[ "${binary_count}" -eq 0 ]]; then
            log_warning "No binary artifacts found in ${BUILD_DIR}"
            log_info "Available files:"
            ls -la "${BUILD_DIR}/" || true
        else
            log_verbose "Found ${binary_count} binary artifacts"
        fi
    fi
    
    if [[ "${SIGN_CHECKSUMS}" == "true" ]] && [[ ! -f "${BUILD_DIR}/checksums.txt" ]]; then
        log_warning "Checksums file not found: ${BUILD_DIR}/checksums.txt"
        log_info "Generating checksums..."
        generate_checksums
    fi
}

# Generate checksums for binaries
generate_checksums() {
    log_info "Generating SHA256 checksums..."
    
    cd "${BUILD_DIR}"
    
    # Generate checksums for all binaries
    find . -name "cloud-mcp-*" -type f -exec sha256sum {} \; > checksums.txt
    
    if [[ -s checksums.txt ]]; then
        log_success "Generated checksums for $(wc -l < checksums.txt) files"
        log_verbose "Checksums:"
        cat checksums.txt | while read -r line; do
            log_verbose "  ${line}"
        done
    else
        log_error "Failed to generate checksums"
        exit 1
    fi
    
    cd "${PROJECT_ROOT}"
}

# Sign binary artifacts
sign_binaries() {
    log_info "Signing binary artifacts..."
    
    cd "${BUILD_DIR}"
    
    local signed_count=0
    
    for file in cloud-mcp-*; do
        if [[ -f "${file}" ]]; then
            log_info "Signing: ${file}"
            
            if [[ "${KEYLESS}" == "true" ]]; then
                # Keyless signing
                cosign sign-blob \
                    --yes \
                    --bundle "${file}.cosign.bundle" \
                    "${file}"
            else
                # Private key signing
                cosign sign-blob \
                    --key "${PRIVATE_KEY}" \
                    --bundle "${file}.cosign.bundle" \
                    "${file}"
            fi
            
            ((signed_count++))
            log_verbose "Created signature bundle: ${file}.cosign.bundle"
        fi
    done
    
    cd "${PROJECT_ROOT}"
    
    if [[ "${signed_count}" -gt 0 ]]; then
        log_success "Signed ${signed_count} binary artifacts"
    else
        log_warning "No binary artifacts found to sign"
    fi
}

# Sign checksums file
sign_checksums() {
    log_info "Signing checksums file..."
    
    cd "${BUILD_DIR}"
    
    if [[ -f "checksums.txt" ]]; then
        log_info "Signing: checksums.txt"
        
        if [[ "${KEYLESS}" == "true" ]]; then
            # Keyless signing
            cosign sign-blob \
                --yes \
                --bundle "checksums.txt.cosign.bundle" \
                "checksums.txt"
        else
            # Private key signing
            cosign sign-blob \
                --key "${PRIVATE_KEY}" \
                --bundle "checksums.txt.cosign.bundle" \
                "checksums.txt"
        fi
        
        log_success "Signed checksums file"
        log_verbose "Created signature bundle: checksums.txt.cosign.bundle"
    else
        log_error "Checksums file not found: checksums.txt"
        exit 1
    fi
    
    cd "${PROJECT_ROOT}"
}

# Sign container image
sign_container() {
    log_info "Signing container image: ${CONTAINER_TAG}"
    
    # Check if image exists locally
    if ! docker image inspect "${CONTAINER_TAG}" &> /dev/null; then
        log_error "Container image not found locally: ${CONTAINER_TAG}"
        log_info "Build the image first or pull from registry"
        exit 1
    fi
    
    if [[ "${KEYLESS}" == "true" ]]; then
        # Keyless signing
        cosign sign \
            --yes \
            "${CONTAINER_TAG}"
    else
        # Private key signing
        cosign sign \
            --key "${PRIVATE_KEY}" \
            "${CONTAINER_TAG}"
    fi
    
    log_success "Signed container image: ${CONTAINER_TAG}"
}

# Verify signatures
verify_signatures() {
    log_info "Verifying signatures..."
    
    local verification_failed=false
    
    # Verify binary signatures
    if [[ "${SIGN_BINARIES}" == "true" ]]; then
        cd "${BUILD_DIR}"
        
        for file in cloud-mcp-*; do
            if [[ -f "${file}" ]] && [[ -f "${file}.cosign.bundle" ]]; then
                log_info "Verifying: ${file}"
                
                if [[ "${KEYLESS}" == "true" ]]; then
                    # Keyless verification
                    if cosign verify-blob \
                        --bundle "${file}.cosign.bundle" \
                        --certificate-identity-regexp="https://github.com/chadit/CloudMCP" \
                        --certificate-oidc-issuer="https://token.actions.githubusercontent.com" \
                        "${file}" &> /dev/null; then
                        log_success "✓ ${file} signature verified"
                    else
                        log_error "✗ ${file} signature verification failed"
                        verification_failed=true
                    fi
                else
                    # Public key verification (requires public key)
                    log_warning "Public key verification not implemented in this script"
                fi
            fi
        done
        
        cd "${PROJECT_ROOT}"
    fi
    
    # Verify checksums signature
    if [[ "${SIGN_CHECKSUMS}" == "true" ]] && [[ -f "${BUILD_DIR}/checksums.txt.cosign.bundle" ]]; then
        cd "${BUILD_DIR}"
        
        log_info "Verifying: checksums.txt"
        
        if [[ "${KEYLESS}" == "true" ]]; then
            if cosign verify-blob \
                --bundle "checksums.txt.cosign.bundle" \
                --certificate-identity-regexp="https://github.com/chadit/CloudMCP" \
                --certificate-oidc-issuer="https://token.actions.githubusercontent.com" \
                "checksums.txt" &> /dev/null; then
                log_success "✓ checksums.txt signature verified"
            else
                log_error "✗ checksums.txt signature verification failed"
                verification_failed=true
            fi
        fi
        
        cd "${PROJECT_ROOT}"
    fi
    
    # Verify container signature
    if [[ "${SIGN_CONTAINER}" == "true" ]] && [[ -n "${CONTAINER_TAG}" ]]; then
        log_info "Verifying container: ${CONTAINER_TAG}"
        
        if [[ "${KEYLESS}" == "true" ]]; then
            if cosign verify \
                --certificate-identity-regexp="https://github.com/chadit/CloudMCP" \
                --certificate-oidc-issuer="https://token.actions.githubusercontent.com" \
                "${CONTAINER_TAG}" &> /dev/null; then
                log_success "✓ ${CONTAINER_TAG} signature verified"
            else
                log_error "✗ ${CONTAINER_TAG} signature verification failed"
                verification_failed=true
            fi
        fi
    fi
    
    if [[ "${verification_failed}" == "true" ]]; then
        log_error "Some signature verifications failed"
        exit 1
    else
        log_success "All signatures verified successfully"
    fi
}

# Show signing summary
show_summary() {
    log_info "Signing Summary:"
    
    if [[ "${SIGN_BINARIES}" == "true" ]]; then
        local bundle_count
        bundle_count=$(find "${BUILD_DIR}" -name "*.cosign.bundle" -type f | grep -c "cloud-mcp-" || echo "0")
        echo "  Binary signatures: ${bundle_count}"
    fi
    
    if [[ "${SIGN_CHECKSUMS}" == "true" ]] && [[ -f "${BUILD_DIR}/checksums.txt.cosign.bundle" ]]; then
        echo "  Checksums signature: ✓"
    fi
    
    if [[ "${SIGN_CONTAINER}" == "true" ]] && [[ -n "${CONTAINER_TAG}" ]]; then
        echo "  Container signature: ✓ ${CONTAINER_TAG}"
    fi
    
    echo "  Signing method: $([ "${KEYLESS}" == "true" ] && echo "Keyless (OIDC)" || echo "Private key")"
    
    # Show verification instructions
    cat << EOF

Verification Instructions:
========================

# Install cosign
go install github.com/sigstore/cosign/v2/cmd/cosign@latest

# Verify binary (example)
cosign verify-blob \\
  --bundle cloud-mcp-linux-amd64.cosign.bundle \\
  --certificate-identity-regexp="https://github.com/chadit/CloudMCP" \\
  --certificate-oidc-issuer="https://token.actions.githubusercontent.com" \\
  cloud-mcp-linux-amd64

# Verify checksums
cosign verify-blob \\
  --bundle checksums.txt.cosign.bundle \\
  --certificate-identity-regexp="https://github.com/chadit/CloudMCP" \\
  --certificate-oidc-issuer="https://token.actions.githubusercontent.com" \\
  checksums.txt

EOF

    if [[ "${SIGN_CONTAINER}" == "true" ]] && [[ -n "${CONTAINER_TAG}" ]]; then
        cat << EOF
# Verify container image
cosign verify \\
  --certificate-identity-regexp="https://github.com/chadit/CloudMCP" \\
  --certificate-oidc-issuer="https://token.actions.githubusercontent.com" \\
  ${CONTAINER_TAG}

EOF
    fi
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -b|--binaries)
                SIGN_BINARIES=true
                shift
                ;;
            -c|--checksums)
                SIGN_CHECKSUMS=true
                shift
                ;;
            -i|--container)
                SIGN_CONTAINER=true
                CONTAINER_TAG="$2"
                shift 2
                ;;
            -v|--verify)
                VERIFY_SIGNATURES=true
                shift
                ;;
            -k|--key)
                PRIVATE_KEY="$2"
                KEYLESS=false
                shift 2
                ;;
            --no-keyless)
                KEYLESS=false
                shift
                ;;
            --verbose)
                VERBOSE=true
                shift
                ;;
            --build-dir)
                BUILD_DIR="$2"
                shift 2
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # Validate arguments
    if [[ "${KEYLESS}" == "false" ]] && [[ -z "${PRIVATE_KEY}" ]] && [[ -z "${COSIGN_PRIVATE_KEY}" ]]; then
        log_error "Private key required when --no-keyless is specified"
        log_info "Use --key /path/to/key or set COSIGN_PRIVATE_KEY environment variable"
        exit 1
    fi
    
    if [[ -n "${PRIVATE_KEY}" ]] && [[ ! -f "${PRIVATE_KEY}" ]]; then
        log_error "Private key file not found: ${PRIVATE_KEY}"
        exit 1
    fi
    
    if [[ "${SIGN_CONTAINER}" == "true" ]] && [[ -z "${CONTAINER_TAG}" ]]; then
        log_error "Container tag required when --container is specified"
        exit 1
    fi
    
    # Use environment variable if private key not specified via argument
    if [[ "${KEYLESS}" == "false" ]] && [[ -z "${PRIVATE_KEY}" ]]; then
        PRIVATE_KEY="${COSIGN_PRIVATE_KEY}"
    fi
}

# Main function
main() {
    log_info "CloudMCP Artifact Signing Script"
    log_info "================================"
    
    parse_args "$@"
    check_cosign
    validate_artifacts
    
    # Perform signing operations
    if [[ "${SIGN_BINARIES}" == "true" ]]; then
        sign_binaries
    fi
    
    if [[ "${SIGN_CHECKSUMS}" == "true" ]]; then
        sign_checksums
    fi
    
    if [[ "${SIGN_CONTAINER}" == "true" ]]; then
        sign_container
    fi
    
    # Verify signatures if requested
    if [[ "${VERIFY_SIGNATURES}" == "true" ]]; then
        verify_signatures
    fi
    
    # Show summary
    show_summary
    
    log_success "Artifact signing completed successfully!"
}

# Run main function with all arguments
main "$@"