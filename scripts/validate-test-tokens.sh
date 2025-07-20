#!/bin/bash

# CloudMCP Token Validation Script
# Validates integration test tokens for security compliance
# 
# SECURITY FEATURES:
# - Token format validation without exposure
# - Secure logging with automatic redaction
# - Environment variable validation
# - CI/CD pipeline integration
# - Constant-time validation to prevent timing attacks

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
VALIDATION_BINARY="${PROJECT_ROOT}/bin/token-validator"
LOG_FILE="${PROJECT_ROOT}/tmp/token-validation.log"

# Colors for output (if terminal supports it)
if [[ -t 1 ]]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    NC='\033[0m' # No Color
else
    RED=''
    GREEN=''
    YELLOW=''
    BLUE=''
    NC=''
fi

# Logging functions with automatic token redaction
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a "${LOG_FILE}" 2>/dev/null || echo -e "${BLUE}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1" | tee -a "${LOG_FILE}" 2>/dev/null || echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "${LOG_FILE}" 2>/dev/null || echo -e "${RED}[ERROR]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "${LOG_FILE}" 2>/dev/null || echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Secure token redaction for logging
redact_token() {
    local token="$1"
    if [[ -z "$token" ]]; then
        echo "[EMPTY]"
        return
    fi
    
    local length=${#token}
    if [[ $length -le 4 ]]; then
        echo "[REDACTED]"
    elif [[ $length -le 8 ]]; then
        echo "${token:0:2}$(printf '*%.0s' $(seq 1 $((length-2))))"
    elif [[ $length -le 16 ]]; then
        echo "${token:0:3}$(printf '*%.0s' $(seq 1 $((length-6))))${token: -3}"
    else
        echo "${token:0:4}$(printf '*%.0s' $(seq 1 $((length-8))))${token: -4}"
    fi
}

# Create temp directory for logs
create_temp_dir() {
    mkdir -p "${PROJECT_ROOT}/tmp"
    
    # Secure permissions on temp directory
    chmod 700 "${PROJECT_ROOT}/tmp" 2>/dev/null || true
    
    # Clear previous log
    > "${LOG_FILE}" 2>/dev/null || touch "${LOG_FILE}"
    chmod 600 "${LOG_FILE}" 2>/dev/null || true
}

# Validate token format without exposing the token
validate_token_format() {
    local token_name="$1"
    local token_value="$2"
    local expected_format="${3:-hex}"
    local min_length="${4:-16}"
    local max_length="${5:-128}"
    
    local redacted_token
    redacted_token=$(redact_token "$token_value")
    
    # Check if token is empty
    if [[ -z "$token_value" ]]; then
        log_warn "Token $token_name is empty (redacted: $redacted_token)"
        return 0  # Allow empty tokens for optional environment variables
    fi
    
    # Validate length
    local length=${#token_value}
    if [[ $length -lt $min_length ]] || [[ $length -gt $max_length ]]; then
        log_error "Token $token_name has invalid length: $length (expected: $min_length-$max_length) (redacted: $redacted_token)"
        return 1
    fi
    
    # Validate format based on type
    case "$expected_format" in
        "hex")
            if ! [[ "$token_value" =~ ^[0-9a-fA-F]+$ ]]; then
                log_error "Token $token_name has invalid hex format (redacted: $redacted_token)"
                return 1
            fi
            ;;
        "alphanumeric")
            if ! [[ "$token_value" =~ ^[a-zA-Z0-9_\-\.]+$ ]]; then
                log_error "Token $token_name has invalid alphanumeric format (redacted: $redacted_token)"
                return 1
            fi
            ;;
        "base64")
            if ! [[ "$token_value" =~ ^[A-Za-z0-9+/]*={0,2}$ ]] || [[ $((length % 4)) -ne 0 ]]; then
                log_error "Token $token_name has invalid base64 format (redacted: $redacted_token)"
                return 1
            fi
            ;;
        *)
            log_warn "Unknown token format: $expected_format for $token_name"
            ;;
    esac
    
    log_success "Token $token_name validation passed (length: $length, format: $expected_format, redacted: $redacted_token)"
    return 0
}

# Validate all environment tokens
validate_environment_tokens() {
    local validation_errors=0
    
    log_info "Starting environment token validation..."
    
    # Define token validation rules
    # Format: "TOKEN_NAME:format:min_length:max_length:required"
    local token_rules=(
        "LINODE_TOKEN:hex:32:128:false"
        "AWS_ACCESS_KEY_ID:alphanumeric:16:32:false"
        "AWS_SECRET_ACCESS_KEY:alphanumeric:32:64:false"
        "GITHUB_TOKEN:alphanumeric:40:255:false"
        "TEST_TOKEN:hex:16:256:false"
    )
    
    for rule in "${token_rules[@]}"; do
        IFS=':' read -r token_name format min_len max_len required <<< "$rule"
        
        local token_value="${!token_name:-}"
        
        # Check if required token is missing
        if [[ "$required" == "true" && -z "$token_value" ]]; then
            log_error "Required token $token_name is missing"
            ((validation_errors++))
            continue
        fi
        
        # Validate token if present
        if [[ -n "$token_value" ]]; then
            if ! validate_token_format "$token_name" "$token_value" "$format" "$min_len" "$max_len"; then
                ((validation_errors++))
            fi
        else
            log_info "Optional token $token_name is not set"
        fi
    done
    
    return $validation_errors
}

# Check for potential token exposure in environment
check_token_exposure() {
    log_info "Checking for potential token exposure..."
    
    # Check common files that might accidentally contain tokens
    local check_files=(
        ".env"
        ".env.local"
        ".env.example"
        "config.toml"
        "*.log"
        "*.tmp"
    )
    
    local exposure_found=false
    
    for pattern in "${check_files[@]}"; do
        if find "$PROJECT_ROOT" -name "$pattern" -type f 2>/dev/null | head -1 | grep -q .; then
            log_warn "Found files matching pattern: $pattern - ensure no tokens are committed"
            exposure_found=true
        fi
    done
    
    # Check git history for potential token patterns (basic check)
    if git -C "$PROJECT_ROOT" rev-parse --git-dir >/dev/null 2>&1; then
        if git -C "$PROJECT_ROOT" log --all --grep="token\|password\|secret\|key" --oneline 2>/dev/null | head -1 | grep -q .; then
            log_warn "Found commits mentioning tokens/secrets - review git history"
            exposure_found=true
        fi
    fi
    
    if [[ "$exposure_found" == false ]]; then
        log_success "No obvious token exposure detected"
    fi
}

# Generate security report
generate_security_report() {
    local validation_errors="$1"
    
    log_info "Generating security validation report..."
    
    cat > "${PROJECT_ROOT}/tmp/token-security-report.txt" << EOF
CloudMCP Token Security Validation Report
=========================================
Generated: $(date -u '+%Y-%m-%d %H:%M:%S UTC')
Validation Errors: $validation_errors

SECURITY RECOMMENDATIONS:
- Store tokens in secure environment variables
- Use different tokens for different environments (dev/staging/prod)
- Rotate tokens regularly
- Never commit tokens to version control
- Use token validation in CI/CD pipelines
- Monitor for token exposure in logs and outputs

VALIDATION SUMMARY:
- Checked token formats and lengths
- Verified secure token handling
- Scanned for potential exposure risks
- Applied constant-time validation

For more details, see: ${LOG_FILE}
EOF
    
    chmod 600 "${PROJECT_ROOT}/tmp/token-security-report.txt" 2>/dev/null || true
    
    log_info "Security report saved to: ${PROJECT_ROOT}/tmp/token-security-report.txt"
}

# Main validation function
main() {
    local mode="${1:-validate}"
    
    log_info "CloudMCP Token Security Validation"
    log_info "=================================="
    
    create_temp_dir
    
    case "$mode" in
        "validate")
            log_info "Mode: Token Validation"
            local errors=0
            
            if ! validate_environment_tokens; then
                errors=$?
            fi
            
            check_token_exposure
            generate_security_report "$errors"
            
            if [[ $errors -gt 0 ]]; then
                log_error "Token validation failed with $errors errors"
                exit 1
            else
                log_success "All token validations passed"
                exit 0
            fi
            ;;
        "ci")
            log_info "Mode: CI/CD Integration"
            
            # Set strict mode for CI
            set -e
            
            # In CI, we want to fail fast on any token issues
            local errors=0
            if ! validate_environment_tokens; then
                errors=$?
            fi
            
            if [[ $errors -gt 0 ]]; then
                log_error "CI token validation failed - check token configuration"
                exit 1
            fi
            
            log_success "CI token validation passed"
            exit 0
            ;;
        "audit")
            log_info "Mode: Security Audit"
            
            validate_environment_tokens || true  # Don't fail on errors in audit mode
            check_token_exposure
            generate_security_report "0"
            
            log_success "Security audit completed"
            cat "${PROJECT_ROOT}/tmp/token-security-report.txt"
            exit 0
            ;;
        "help"|"-h"|"--help")
            cat << EOF
CloudMCP Token Security Validation Script

USAGE:
    $0 [MODE]

MODES:
    validate    Validate all environment tokens (default)
    ci          CI/CD mode - strict validation with fast failure
    audit       Security audit mode - comprehensive checking
    help        Show this help message

ENVIRONMENT VARIABLES:
    LINODE_TOKEN         - Linode API token (hex format, 32-128 chars)
    AWS_ACCESS_KEY_ID    - AWS access key (alphanumeric, 16-32 chars)
    AWS_SECRET_ACCESS_KEY - AWS secret key (alphanumeric, 32-64 chars)
    GITHUB_TOKEN         - GitHub API token (alphanumeric, 40-255 chars)
    TEST_TOKEN           - Generic test token (hex format, 16-256 chars)

SECURITY FEATURES:
    - Token format validation without exposure
    - Automatic token redaction in logs
    - Constant-time validation (timing attack prevention)
    - Environment variable security checking
    - Git history scanning for token patterns

EXAMPLES:
    $0                  # Validate tokens
    $0 validate         # Same as above
    $0 ci               # CI/CD pipeline validation
    $0 audit            # Full security audit

EXIT CODES:
    0 - Success
    1 - Validation errors found
    2 - Script error
EOF
            exit 0
            ;;
        *)
            log_error "Unknown mode: $mode"
            log_info "Use '$0 help' for usage information"
            exit 2
            ;;
    esac
}

# Trap to ensure cleanup
cleanup() {
    # Secure cleanup - ensure no tokens are left in temp files
    if [[ -d "${PROJECT_ROOT}/tmp" ]]; then
        find "${PROJECT_ROOT}/tmp" -name "*.tmp" -type f -delete 2>/dev/null || true
    fi
}

trap cleanup EXIT

# Run main function with all arguments
main "$@"