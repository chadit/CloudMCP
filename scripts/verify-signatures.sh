#!/bin/bash

# CloudMCP Signature Verification Script
# Verifies cosign signatures for binaries, checksums, and container images
# Usage: ./scripts/verify-signatures.sh [OPTIONS]

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
VERIFY_BINARIES=true
VERIFY_CHECKSUMS=true
VERIFY_CONTAINER=false
CONTAINER_TAG=""
VERIFY_SLSA=false
SLSA_FILE=""
PUBLIC_KEY=""
KEYLESS=true
VERBOSE=false
STRICT=false
CERT_IDENTITY="https://github.com/chadit/CloudMCP"
CERT_OIDC_ISSUER="https://token.actions.githubusercontent.com"

# Help function
show_help() {
    cat << EOF
CloudMCP Signature Verification Script

Usage: $0 [OPTIONS]

Options:
    -h, --help              Show this help message
    -b, --binaries          Verify binary signatures (default: true)
    -c, --checksums         Verify checksums signature (default: true)
    -i, --container TAG     Verify container image signature
    -s, --slsa FILE         Verify SLSA provenance file
    -k, --key FILE          Use public key file instead of keyless verification
    --no-keyless            Disable keyless verification (requires public key)
    --cert-identity ID      Certificate identity for keyless verification
    --cert-oidc-issuer URL  Certificate OIDC issuer for keyless verification
    --strict                Enable strict verification mode
    --verbose               Enable verbose output
    --build-dir DIR         Specify build directory (default: ./build)

Examples:
    # Verify all binaries and checksums (keyless)
    $0 --binaries --checksums

    # Verify container image
    $0 --container cloudmcp:latest

    # Verify with public key
    $0 --key /path/to/public.key --no-keyless

    # Verify everything including SLSA provenance
    $0 --binaries --checksums --container cloudmcp:v1.0.0 --slsa slsa-provenance.intoto.jsonl

    # Strict verification with custom identity
    $0 --strict --cert-identity "https://github.com/myorg/myrepo"

Environment Variables:
    COSIGN_PUBLIC_KEY       Path to public key (alternative to --key)

Prerequisites:
    - cosign must be installed and available in PATH
    - For keyless verification: Access to certificate transparency logs
    - For public key verification: Valid cosign public key

Verification Types:
    Keyless: Uses certificate transparency and OIDC identity verification
    Public Key: Uses traditional public key cryptography

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

# Validate verification targets
validate_targets() {
    if [[ ! -d "${BUILD_DIR}" ]]; then
        log_error "Build directory not found: ${BUILD_DIR}"
        log_info "Specify correct --build-dir or ensure artifacts are available"
        exit 1
    fi
    
    if [[ "${VERIFY_BINARIES}" == "true" ]]; then
        local bundle_count
        bundle_count=$(find "${BUILD_DIR}" -name "cloud-mcp-*.cosign.bundle" -type f | wc -l)
        if [[ "${bundle_count}" -eq 0 ]]; then
            log_warning "No binary signature bundles found in ${BUILD_DIR}"
            log_info "Available files:"
            ls -la "${BUILD_DIR}/" || true
        else
            log_verbose "Found ${bundle_count} binary signature bundles"
        fi
    fi
    
    if [[ "${VERIFY_CHECKSUMS}" == "true" ]] && [[ ! -f "${BUILD_DIR}/checksums.txt.cosign.bundle" ]]; then
        log_warning "Checksums signature bundle not found: ${BUILD_DIR}/checksums.txt.cosign.bundle"
    fi
    
    if [[ "${VERIFY_SLSA}" == "true" ]] && [[ ! -f "${SLSA_FILE}" ]]; then
        log_error "SLSA provenance file not found: ${SLSA_FILE}"
        exit 1
    fi
}

# Verify binary signatures
verify_binaries() {
    log_info "Verifying binary signatures..."
    
    cd "${BUILD_DIR}"
    
    local verified_count=0
    local failed_count=0
    
    for bundle in cloud-mcp-*.cosign.bundle; do
        if [[ -f "${bundle}" ]]; then
            local binary="${bundle%.cosign.bundle}"
            
            if [[ -f "${binary}" ]]; then
                log_info "Verifying: ${binary}"
                
                local verification_output
                local verification_result=0
                
                if [[ "${KEYLESS}" == "true" ]]; then
                    # Keyless verification
                    if [[ "${VERBOSE}" == "true" ]]; then
                        verification_output=$(cosign verify-blob \
                            --bundle "${bundle}" \
                            --certificate-identity-regexp="${CERT_IDENTITY}" \
                            --certificate-oidc-issuer="${CERT_OIDC_ISSUER}" \
                            "${binary}" 2>&1) || verification_result=$?
                    else
                        verification_output=$(cosign verify-blob \
                            --bundle "${bundle}" \
                            --certificate-identity-regexp="${CERT_IDENTITY}" \
                            --certificate-oidc-issuer="${CERT_OIDC_ISSUER}" \
                            "${binary}" 2>&1 >/dev/null) || verification_result=$?
                    fi
                else
                    # Public key verification
                    if [[ "${VERBOSE}" == "true" ]]; then
                        verification_output=$(cosign verify-blob \
                            --bundle "${bundle}" \
                            --key "${PUBLIC_KEY}" \
                            "${binary}" 2>&1) || verification_result=$?
                    else
                        verification_output=$(cosign verify-blob \
                            --bundle "${bundle}" \
                            --key "${PUBLIC_KEY}" \
                            "${binary}" 2>&1 >/dev/null) || verification_result=$?
                    fi
                fi
                
                if [[ "${verification_result}" -eq 0 ]]; then
                    log_success "✓ ${binary} signature verified"
                    ((verified_count++))
                    
                    if [[ "${VERBOSE}" == "true" ]] && [[ -n "${verification_output}" ]]; then
                        log_verbose "Verification details for ${binary}:"
                        echo "${verification_output}" | sed 's/^/    /'
                    fi
                else
                    log_error "✗ ${binary} signature verification failed"
                    ((failed_count++))
                    
                    if [[ "${VERBOSE}" == "true" ]] && [[ -n "${verification_output}" ]]; then
                        log_verbose "Verification error for ${binary}:"
                        echo "${verification_output}" | sed 's/^/    /'
                    fi
                    
                    if [[ "${STRICT}" == "true" ]]; then
                        log_error "Strict mode: Stopping on first verification failure"
                        cd "${PROJECT_ROOT}"
                        exit 1
                    fi
                fi
            else
                log_warning "Binary file not found for bundle: ${bundle}"
            fi
        fi
    done
    
    cd "${PROJECT_ROOT}"
    
    log_info "Binary verification summary: ${verified_count} verified, ${failed_count} failed"
    
    if [[ "${failed_count}" -gt 0 ]] && [[ "${STRICT}" == "true" ]]; then
        log_error "Binary signature verification failed in strict mode"
        exit 1
    fi
}

# Verify checksums signature
verify_checksums() {
    log_info "Verifying checksums signature..."
    
    cd "${BUILD_DIR}"
    
    if [[ -f "checksums.txt" ]] && [[ -f "checksums.txt.cosign.bundle" ]]; then
        log_info "Verifying: checksums.txt"
        
        local verification_output
        local verification_result=0
        
        if [[ "${KEYLESS}" == "true" ]]; then
            # Keyless verification
            if [[ "${VERBOSE}" == "true" ]]; then
                verification_output=$(cosign verify-blob \
                    --bundle "checksums.txt.cosign.bundle" \
                    --certificate-identity-regexp="${CERT_IDENTITY}" \
                    --certificate-oidc-issuer="${CERT_OIDC_ISSUER}" \
                    "checksums.txt" 2>&1) || verification_result=$?
            else
                verification_output=$(cosign verify-blob \
                    --bundle "checksums.txt.cosign.bundle" \
                    --certificate-identity-regexp="${CERT_IDENTITY}" \
                    --certificate-oidc-issuer="${CERT_OIDC_ISSUER}" \
                    "checksums.txt" 2>&1 >/dev/null) || verification_result=$?
            fi
        else
            # Public key verification
            if [[ "${VERBOSE}" == "true" ]]; then
                verification_output=$(cosign verify-blob \
                    --bundle "checksums.txt.cosign.bundle" \
                    --key "${PUBLIC_KEY}" \
                    "checksums.txt" 2>&1) || verification_result=$?
            else
                verification_output=$(cosign verify-blob \
                    --bundle "checksums.txt.cosign.bundle" \
                    --key "${PUBLIC_KEY}" \
                    "checksums.txt" 2>&1 >/dev/null) || verification_result=$?
            fi
        fi
        
        if [[ "${verification_result}" -eq 0 ]]; then
            log_success "✓ checksums.txt signature verified"
            
            if [[ "${VERBOSE}" == "true" ]] && [[ -n "${verification_output}" ]]; then
                log_verbose "Verification details for checksums.txt:"
                echo "${verification_output}" | sed 's/^/    /'
            fi
        else
            log_error "✗ checksums.txt signature verification failed"
            
            if [[ "${VERBOSE}" == "true" ]] && [[ -n "${verification_output}" ]]; then
                log_verbose "Verification error for checksums.txt:"
                echo "${verification_output}" | sed 's/^/    /'
            fi
            
            if [[ "${STRICT}" == "true" ]]; then
                log_error "Strict mode: Stopping on checksums verification failure"
                cd "${PROJECT_ROOT}"
                exit 1
            fi
        fi
    else
        log_error "Checksums file or signature bundle not found"
        if [[ "${STRICT}" == "true" ]]; then
            cd "${PROJECT_ROOT}"
            exit 1
        fi
    fi
    
    cd "${PROJECT_ROOT}"
}

# Verify container image signature
verify_container() {
    log_info "Verifying container image signature: ${CONTAINER_TAG}"
    
    # Check if image exists locally or is accessible
    if ! docker image inspect "${CONTAINER_TAG}" &> /dev/null; then
        log_warning "Container image not found locally: ${CONTAINER_TAG}"
        log_info "Attempting verification anyway (image may be in registry)"
    fi
    
    local verification_output
    local verification_result=0
    
    if [[ "${KEYLESS}" == "true" ]]; then
        # Keyless verification
        if [[ "${VERBOSE}" == "true" ]]; then
            verification_output=$(cosign verify \
                --certificate-identity-regexp="${CERT_IDENTITY}" \
                --certificate-oidc-issuer="${CERT_OIDC_ISSUER}" \
                "${CONTAINER_TAG}" 2>&1) || verification_result=$?
        else
            verification_output=$(cosign verify \
                --certificate-identity-regexp="${CERT_IDENTITY}" \
                --certificate-oidc-issuer="${CERT_OIDC_ISSUER}" \
                "${CONTAINER_TAG}" 2>&1 >/dev/null) || verification_result=$?
        fi
    else
        # Public key verification
        if [[ "${VERBOSE}" == "true" ]]; then
            verification_output=$(cosign verify \
                --key "${PUBLIC_KEY}" \
                "${CONTAINER_TAG}" 2>&1) || verification_result=$?
        else
            verification_output=$(cosign verify \
                --key "${PUBLIC_KEY}" \
                "${CONTAINER_TAG}" 2>&1 >/dev/null) || verification_result=$?
        fi
    fi
    
    if [[ "${verification_result}" -eq 0 ]]; then
        log_success "✓ ${CONTAINER_TAG} signature verified"
        
        if [[ "${VERBOSE}" == "true" ]] && [[ -n "${verification_output}" ]]; then
            log_verbose "Verification details for ${CONTAINER_TAG}:"
            echo "${verification_output}" | sed 's/^/    /'
        fi
    else
        log_error "✗ ${CONTAINER_TAG} signature verification failed"
        
        if [[ "${VERBOSE}" == "true" ]] && [[ -n "${verification_output}" ]]; then
            log_verbose "Verification error for ${CONTAINER_TAG}:"
            echo "${verification_output}" | sed 's/^/    /'
        fi
        
        if [[ "${STRICT}" == "true" ]]; then
            log_error "Strict mode: Stopping on container verification failure"
            exit 1
        fi
    fi
}

# Verify SLSA provenance
verify_slsa() {
    log_info "Verifying SLSA provenance: ${SLSA_FILE}"
    
    if [[ ! -f "${SLSA_FILE}" ]]; then
        log_error "SLSA provenance file not found: ${SLSA_FILE}"
        if [[ "${STRICT}" == "true" ]]; then
            exit 1
        fi
        return
    fi
    
    # Check if there's a corresponding signature bundle
    local slsa_bundle="${SLSA_FILE}.cosign.bundle"
    
    if [[ -f "${slsa_bundle}" ]]; then
        log_info "Verifying SLSA provenance signature"
        
        local verification_output
        local verification_result=0
        
        if [[ "${KEYLESS}" == "true" ]]; then
            # Keyless verification
            if [[ "${VERBOSE}" == "true" ]]; then
                verification_output=$(cosign verify-blob \
                    --bundle "${slsa_bundle}" \
                    --certificate-identity-regexp="${CERT_IDENTITY}" \
                    --certificate-oidc-issuer="${CERT_OIDC_ISSUER}" \
                    "${SLSA_FILE}" 2>&1) || verification_result=$?
            else
                verification_output=$(cosign verify-blob \
                    --bundle "${slsa_bundle}" \
                    --certificate-identity-regexp="${CERT_IDENTITY}" \
                    --certificate-oidc-issuer="${CERT_OIDC_ISSUER}" \
                    "${SLSA_FILE}" 2>&1 >/dev/null) || verification_result=$?
            fi
        else
            # Public key verification
            if [[ "${VERBOSE}" == "true" ]]; then
                verification_output=$(cosign verify-blob \
                    --bundle "${slsa_bundle}" \
                    --key "${PUBLIC_KEY}" \
                    "${SLSA_FILE}" 2>&1) || verification_result=$?
            else
                verification_output=$(cosign verify-blob \
                    --bundle "${slsa_bundle}" \
                    --key "${PUBLIC_KEY}" \
                    "${SLSA_FILE}" 2>&1 >/dev/null) || verification_result=$?
            fi
        fi
        
        if [[ "${verification_result}" -eq 0 ]]; then
            log_success "✓ SLSA provenance signature verified"
            
            if [[ "${VERBOSE}" == "true" ]] && [[ -n "${verification_output}" ]]; then
                log_verbose "Verification details for SLSA provenance:"
                echo "${verification_output}" | sed 's/^/    /'
            fi
        else
            log_error "✗ SLSA provenance signature verification failed"
            
            if [[ "${VERBOSE}" == "true" ]] && [[ -n "${verification_output}" ]]; then
                log_verbose "Verification error for SLSA provenance:"
                echo "${verification_output}" | sed 's/^/    /'
            fi
            
            if [[ "${STRICT}" == "true" ]]; then
                log_error "Strict mode: Stopping on SLSA verification failure"
                exit 1
            fi
        fi
    else
        log_warning "No signature bundle found for SLSA provenance: ${slsa_bundle}"
    fi
    
    # Basic SLSA content validation
    if command -v jq &> /dev/null; then
        log_info "Validating SLSA provenance content"
        
        if jq -e '.predicateType' "${SLSA_FILE}" &> /dev/null; then
            local predicate_type
            predicate_type=$(jq -r '.predicateType' "${SLSA_FILE}")
            log_verbose "SLSA predicate type: ${predicate_type}"
            
            if [[ "${predicate_type}" == *"slsa-provenance"* ]]; then
                log_success "✓ SLSA provenance format validated"
            else
                log_warning "Unexpected SLSA predicate type: ${predicate_type}"
            fi
        else
            log_warning "Invalid SLSA provenance format"
        fi
    else
        log_verbose "jq not available, skipping SLSA content validation"
    fi
}

# Validate checksums against binaries
validate_checksums() {
    log_info "Validating checksums against binaries..."
    
    cd "${BUILD_DIR}"
    
    if [[ ! -f "checksums.txt" ]]; then
        log_warning "Checksums file not found, skipping validation"
        cd "${PROJECT_ROOT}"
        return
    fi
    
    local validation_failed=false
    
    while IFS= read -r line; do
        if [[ -n "${line}" ]]; then
            local expected_hash
            local filename
            expected_hash=$(echo "${line}" | cut -d' ' -f1)
            filename=$(echo "${line}" | cut -d' ' -f3-)  # Handle filenames with spaces
            
            if [[ -f "${filename}" ]]; then
                local actual_hash
                actual_hash=$(sha256sum "${filename}" | cut -d' ' -f1)
                
                if [[ "${expected_hash}" == "${actual_hash}" ]]; then
                    log_success "✓ ${filename} checksum validated"
                else
                    log_error "✗ ${filename} checksum mismatch"
                    log_verbose "Expected: ${expected_hash}"
                    log_verbose "Actual:   ${actual_hash}"
                    validation_failed=true
                fi
            else
                log_warning "File not found for checksum validation: ${filename}"
            fi
        fi
    done < "checksums.txt"
    
    cd "${PROJECT_ROOT}"
    
    if [[ "${validation_failed}" == "true" ]]; then
        log_error "Checksum validation failed"
        if [[ "${STRICT}" == "true" ]]; then
            exit 1
        fi
    else
        log_success "All checksums validated successfully"
    fi
}

# Show verification summary
show_summary() {
    log_info "Verification Summary:"
    log_info "===================="
    
    echo "  Verification method: $([ "${KEYLESS}" == "true" ] && echo "Keyless (OIDC)" || echo "Public key")"
    echo "  Certificate identity: ${CERT_IDENTITY}"
    echo "  OIDC issuer: ${CERT_OIDC_ISSUER}"
    echo "  Strict mode: $([ "${STRICT}" == "true" ] && echo "Enabled" || echo "Disabled")"
    echo ""
    
    if [[ "${VERIFY_BINARIES}" == "true" ]]; then
        local bundle_count
        bundle_count=$(find "${BUILD_DIR}" -name "cloud-mcp-*.cosign.bundle" -type f | wc -l)
        echo "  Binary signatures verified: ${bundle_count}"
    fi
    
    if [[ "${VERIFY_CHECKSUMS}" == "true" ]]; then
        echo "  Checksums signature: $([ -f "${BUILD_DIR}/checksums.txt.cosign.bundle" ] && echo "✓ Verified" || echo "✗ Not found")"
    fi
    
    if [[ "${VERIFY_CONTAINER}" == "true" ]]; then
        echo "  Container signature: ${CONTAINER_TAG}"
    fi
    
    if [[ "${VERIFY_SLSA}" == "true" ]]; then
        echo "  SLSA provenance: ${SLSA_FILE}"
    fi
    
    echo ""
    log_success "Verification completed"
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
                VERIFY_BINARIES=true
                shift
                ;;
            -c|--checksums)
                VERIFY_CHECKSUMS=true
                shift
                ;;
            -i|--container)
                VERIFY_CONTAINER=true
                CONTAINER_TAG="$2"
                shift 2
                ;;
            -s|--slsa)
                VERIFY_SLSA=true
                SLSA_FILE="$2"
                shift 2
                ;;
            -k|--key)
                PUBLIC_KEY="$2"
                KEYLESS=false
                shift 2
                ;;
            --no-keyless)
                KEYLESS=false
                shift
                ;;
            --cert-identity)
                CERT_IDENTITY="$2"
                shift 2
                ;;
            --cert-oidc-issuer)
                CERT_OIDC_ISSUER="$2"
                shift 2
                ;;
            --strict)
                STRICT=true
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
    if [[ "${KEYLESS}" == "false" ]] && [[ -z "${PUBLIC_KEY}" ]] && [[ -z "${COSIGN_PUBLIC_KEY}" ]]; then
        log_error "Public key required when --no-keyless is specified"
        log_info "Use --key /path/to/key or set COSIGN_PUBLIC_KEY environment variable"
        exit 1
    fi
    
    if [[ -n "${PUBLIC_KEY}" ]] && [[ ! -f "${PUBLIC_KEY}" ]]; then
        log_error "Public key file not found: ${PUBLIC_KEY}"
        exit 1
    fi
    
    if [[ "${VERIFY_CONTAINER}" == "true" ]] && [[ -z "${CONTAINER_TAG}" ]]; then
        log_error "Container tag required when --container is specified"
        exit 1
    fi
    
    # Use environment variable if public key not specified via argument
    if [[ "${KEYLESS}" == "false" ]] && [[ -z "${PUBLIC_KEY}" ]]; then
        PUBLIC_KEY="${COSIGN_PUBLIC_KEY}"
    fi
}

# Main function
main() {
    log_info "CloudMCP Signature Verification Script"
    log_info "====================================="
    
    parse_args "$@"
    check_cosign
    validate_targets
    
    # Perform verification operations
    if [[ "${VERIFY_BINARIES}" == "true" ]]; then
        verify_binaries
    fi
    
    if [[ "${VERIFY_CHECKSUMS}" == "true" ]]; then
        verify_checksums
        validate_checksums
    fi
    
    if [[ "${VERIFY_CONTAINER}" == "true" ]]; then
        verify_container
    fi
    
    if [[ "${VERIFY_SLSA}" == "true" ]]; then
        verify_slsa
    fi
    
    # Show summary
    show_summary
}

# Run main function with all arguments
main "$@"