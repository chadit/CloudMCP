#!/bin/bash

# CloudMCP Security Utilities
# Secure tool installation with checksum verification
# Usage: ./scripts/security-utils.sh [command] [options]

set -euo pipefail

# Script configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
CACHE_DIR="${PROJECT_ROOT}/.security-cache"
TOOLS_DIR="${PROJECT_ROOT}/.tools"

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
    echo -e "${RED}[ERROR]${NC} $*" >&2
}

# Verified tool configurations
# Bash 3 compatible approach (no associative arrays)
# Function to get tool configuration
get_tool_config() {
    local tool_key="$1"
    
    case "$tool_key" in
        "golangci-lint-linux-amd64")
            echo "v1.64.8:linux:amd64:https://github.com/golangci/golangci-lint/releases/download/v1.64.8/golangci-lint-1.64.8-linux-amd64.tar.gz:b6270687afb143d019f387c791cd2a6f1cb383be9b3124d241ca11bd3ce2e54e"
            ;;
        "golangci-lint-linux-arm64")
            echo "v1.64.8:linux:arm64:https://github.com/golangci/golangci-lint/releases/download/v1.64.8/golangci-lint-1.64.8-linux-arm64.tar.gz:a6ab58ebcb1c48572622146cdaec2956f56871038a54ed1149f1386e287789a5"
            ;;
        "golangci-lint-darwin-amd64")
            echo "v1.64.2:darwin:amd64:https://github.com/golangci/golangci-lint/releases/download/v1.64.2/golangci-lint-1.64.2-darwin-amd64.tar.gz:ab6a9e08c4f534a9523cb2d25847169cda7857feffb39893f958f29016b21364"
            ;;
        "golangci-lint-darwin-arm64")
            echo "v1.64.2:darwin:arm64:https://github.com/golangci/golangci-lint/releases/download/v1.64.2/golangci-lint-1.64.2-darwin-arm64.tar.gz:8ae47cfa821fa17cb6984d0644e2cba9acb61f37ee524d4186d5f0458918da80"
            ;;
        "hadolint-linux-amd64")
            echo "v2.12.0:linux:amd64:https://github.com/hadolint/hadolint/releases/download/v2.12.0/hadolint-Linux-x86_64:56de6d5e5ec427e17b74fa48d51271c7fc0d61244bf5c90e828aab8362d55010"
            ;;
        "hadolint-linux-arm64")
            echo "v2.12.0:linux:arm64:https://github.com/hadolint/hadolint/releases/download/v2.12.0/hadolint-Linux-arm64:5798551bf19f33951881f15eb238f90aef023f11e7ec7e9f4c37961cb87c5df6"
            ;;
        "hadolint-darwin-amd64")
            echo "v2.12.0:darwin:amd64:https://github.com/hadolint/hadolint/releases/download/v2.12.0/hadolint-Darwin-x86_64:2a5b7afcab91645c39a7cebefcd835b865f7488e69be24567f433dfc3d41cd27"
            ;;
        "hadolint-darwin-arm64")
            echo "v2.12.0:darwin:arm64:https://github.com/hadolint/hadolint/releases/download/v2.12.0/hadolint-Darwin-x86_64:2a5b7afcab91645c39a7cebefcd835b865f7488e69be24567f433dfc3d41cd27"
            ;;
        "trivy-linux-amd64")
            echo "v0.48.3:linux:amd64:https://github.com/aquasecurity/trivy/releases/download/v0.48.3/trivy_0.48.3_Linux-64bit.tar.gz:61f6d5c0fb6ed451c8bab8b13acb5d701f1b532bd6b629f3163f8f57bb10e564"
            ;;
        "trivy-linux-arm64")
            echo "v0.48.3:linux:arm64:https://github.com/aquasecurity/trivy/releases/download/v0.48.3/trivy_0.48.3_Linux-ARM64.tar.gz:01e814fbb0b2aaaa4510b6c29e9a37103fe9818f70be816c3ecbb39e836a61b5"
            ;;
        "trivy-darwin-amd64")
            echo "v0.48.3:darwin:amd64:https://github.com/aquasecurity/trivy/releases/download/v0.48.3/trivy_0.48.3_macOS-64bit.tar.gz:4fc0d1f2ec55869ab4772bd321451023ada4589cc8f9114dae71c7656b2be725"
            ;;
        "trivy-darwin-arm64")
            echo "v0.48.3:darwin:arm64:https://github.com/aquasecurity/trivy/releases/download/v0.48.3/trivy_0.48.3_macOS-ARM64.tar.gz:6553a995a97bd7f57c486b7bd38cc297aeeb1125c2eb647cff0866ad6eeef48d"
            ;;
        "cosign-linux-amd64")
            echo "v2.4.1:linux:amd64:https://github.com/sigstore/cosign/releases/download/v2.4.1/cosign-linux-amd64:8b24b946dd5809c6bd93de08033bcf6bc0ed7d336b7785787c080f574b89249b"
            ;;
        "cosign-linux-arm64")
            echo "v2.4.1:linux:arm64:https://github.com/sigstore/cosign/releases/download/v2.4.1/cosign-linux-arm64:3b2e2e3854d0356c45fe6607047526ccd04742d20bd44afb5be91fa2a6e7cb4a"
            ;;
        "cosign-darwin-amd64")
            echo "v2.4.1:darwin:amd64:https://github.com/sigstore/cosign/releases/download/v2.4.1/cosign-darwin-amd64:666032ca283da92b6f7953965688fd51200fdc891a86c19e05c98b898ea0af4e"
            ;;
        *)
            echo ""
            ;;
    esac
}

# Function to list available tools
list_available_tools() {
    echo "golangci-lint-linux-amd64"
    echo "golangci-lint-linux-arm64"
    echo "golangci-lint-darwin-amd64"
    echo "golangci-lint-darwin-arm64"
    echo "hadolint-linux-amd64"
    echo "hadolint-linux-arm64"
    echo "hadolint-darwin-amd64"
    echo "hadolint-darwin-arm64"
    echo "trivy-linux-amd64"
    echo "trivy-linux-arm64"
    echo "trivy-darwin-amd64"
    echo "trivy-darwin-arm64"
    echo "cosign-linux-amd64"
    echo "cosign-linux-arm64"
    echo "cosign-darwin-amd64"
}

# Function to detect platform
detect_platform() {
    local os arch
    
    case "$(uname -s)" in
        Linux*)  os="linux" ;;
        Darwin*) os="darwin" ;;
        CYGWIN*|MINGW*) os="windows" ;;
        *) 
            log_error "Unsupported operating system: $(uname -s)"
            exit 1
            ;;
    esac
    
    case "$(uname -m)" in
        x86_64|amd64) arch="amd64" ;;
        arm64|aarch64) arch="arm64" ;;
        armv7*) arch="arm" ;;
        *) 
            log_error "Unsupported architecture: $(uname -m)"
            exit 1
            ;;
    esac
    
    echo "${os}-${arch}"
}

# Function to verify SHA256 checksum
verify_checksum() {
    local file="$1"
    local expected_checksum="$2"
    
    if [[ ! -f "$file" ]]; then
        log_error "File not found for checksum verification: $file"
        return 1
    fi
    
    local actual_checksum
    if command -v sha256sum >/dev/null 2>&1; then
        actual_checksum=$(sha256sum "$file" | cut -d' ' -f1)
    elif command -v shasum >/dev/null 2>&1; then
        actual_checksum=$(shasum -a 256 "$file" | cut -d' ' -f1)
    else
        log_error "No SHA256 tool found (sha256sum or shasum required)"
        return 1
    fi
    
    if [[ "$actual_checksum" == "$expected_checksum" ]]; then
        log_success "Checksum verification passed: $file"
        return 0
    else
        log_error "Checksum verification failed for: $file"
        log_error "Expected: $expected_checksum"
        log_error "Actual:   $actual_checksum"
        return 1
    fi
}

# Function to download file with retries and integrity checks
secure_download() {
    local url="$1"
    local output_file="$2"
    local expected_checksum="$3"
    local max_retries=3
    local retry_delay=5
    
    log_info "Downloading securely: $(basename "$output_file")"
    
    # Create output directory if needed
    mkdir -p "$(dirname "$output_file")"
    
    for ((i=1; i<=max_retries; i++)); do
        log_info "Download attempt $i/$max_retries..."
        
        # Download with curl (with security options)
        if curl -fsSL \
            --connect-timeout 30 \
            --max-time 300 \
            --retry 2 \
            --retry-delay 3 \
            --user-agent "CloudMCP-Security-Utils/1.0" \
            -o "$output_file" \
            "$url"; then
            
            # Verify checksum
            if verify_checksum "$output_file" "$expected_checksum"; then
                log_success "Secure download completed: $(basename "$output_file")"
                return 0
            else
                log_warning "Checksum verification failed, retrying..."
                rm -f "$output_file"
            fi
        else
            log_warning "Download failed, retrying in ${retry_delay}s..."
        fi
        
        if [[ $i -lt $max_retries ]]; then
            sleep $retry_delay
            retry_delay=$((retry_delay * 2))  # Exponential backoff
        fi
    done
    
    log_error "Failed to download after $max_retries attempts: $url"
    return 1
}

# Function to install a verified tool
install_tool() {
    local tool_name="$1"
    local platform="${2:-$(detect_platform)}"
    local force_install="${3:-false}"
    
    local tool_key="${tool_name}-${platform}"
    
    local config
    config=$(get_tool_config "$tool_key")
    
    if [[ -z "$config" ]]; then
        log_error "Tool configuration not found: $tool_key"
        log_info "Available tools:"
        list_available_tools | grep "^$tool_name-" || echo "  No tools found for $tool_name"
        return 1
    fi
    
    # Parse configuration (handle URLs with colons correctly)
    # Format: version:os:arch:url:checksum
    # Extract each field carefully to handle URLs with colons
    version=$(echo "$config" | cut -d':' -f1)
    os=$(echo "$config" | cut -d':' -f2)
    arch=$(echo "$config" | cut -d':' -f3)
    # Extract URL by removing version:os:arch: prefix and :checksum suffix
    temp_config="${config#*:*:*:}"  # Remove first 3 fields
    checksum="${temp_config##*:}"   # Get last field (checksum)
    url="${temp_config%:*}"         # Remove checksum, leaving URL
    
    local tool_dir="${TOOLS_DIR}/${tool_name}"
    local tool_binary="${tool_dir}/${tool_name}"
    local cache_file="${CACHE_DIR}/${tool_name}-${version}-${platform}"
    
    # Check if already installed and up to date
    if [[ -f "$tool_binary" ]] && [[ "$force_install" != "true" ]]; then
        if [[ -f "${tool_dir}/.version" ]] && [[ "$(cat "${tool_dir}/.version")" == "$version" ]]; then
            log_info "Tool already installed: $tool_name $version"
            return 0
        fi
    fi
    
    log_info "Installing verified tool: $tool_name $version for $platform"
    
    # Create directories
    mkdir -p "$tool_dir" "$CACHE_DIR"
    
    # Download to cache if not present
    if [[ ! -f "$cache_file" ]]; then
        secure_download "$url" "$cache_file" "$checksum"
    else
        log_info "Using cached download"
        verify_checksum "$cache_file" "$checksum"
    fi
    
    # Extract and install
    if [[ "$url" == *.tar.gz ]]; then
        # Extract tar.gz
        log_info "Extracting archive..."
        
        # Special handling for different tool archive structures
        case "$tool_name" in
            "trivy")
                # Trivy archives have the binary directly in the root
                tar -xzf "$cache_file" -C "$tool_dir"
                ;;
            *)
                # Standard extraction with strip-components for other tools
                tar -xzf "$cache_file" -C "$tool_dir" --strip-components=1
                ;;
        esac
        
        # Find the actual binary (may be in subdirectories or different names)
        local binary_path
        binary_path=$(find "$tool_dir" -name "$tool_name" -type f | head -1)
        
        # If binary not found by exact name, try common variations
        if [[ -z "$binary_path" ]]; then
            binary_path=$(find "$tool_dir" -type f -executable | grep -E "${tool_name}(\\.exe)?$" | head -1)
        fi
        
        # If found and different from expected location, move it
        if [[ -n "$binary_path" ]] && [[ "$binary_path" != "$tool_binary" ]]; then
            log_info "Moving binary from $binary_path to $tool_binary"
            mv "$binary_path" "$tool_binary"
        elif [[ -z "$binary_path" ]]; then
            log_error "Binary not found after extraction for tool: $tool_name"
            log_info "Contents of $tool_dir:"
            ls -la "$tool_dir"
            return 1
        fi
    else
        # Single binary file
        cp "$cache_file" "$tool_binary"
    fi
    
    # Make executable
    chmod +x "$tool_binary"
    
    # Store version
    echo "$version" > "${tool_dir}/.version"
    
    # Verify installation
    if [[ -x "$tool_binary" ]]; then
        log_success "Successfully installed: $tool_name $version"
        
        # Test the binary
        if "$tool_binary" --version >/dev/null 2>&1 || "$tool_binary" version >/dev/null 2>&1; then
            log_success "Tool verification passed: $tool_name"
        else
            log_warning "Tool may not be working correctly: $tool_name"
        fi
    else
        log_error "Installation failed: $tool_binary not executable"
        return 1
    fi
}

# Function to update tool checksums (for maintenance)
update_checksums() {
    log_info "Updating tool checksums (maintenance function)"
    log_warning "This function downloads tools to verify checksums"
    log_warning "Only run this when updating tool versions"
    
    local platform
    platform=$(detect_platform)
    
    list_available_tools | while read -r tool_key; do
        if [[ "$tool_key" == *"-${platform}" ]]; then
            local tool_name="${tool_key%-*}"
            local config
            config=$(get_tool_config "$tool_key")
            # Parse configuration (handle URLs with colons correctly)
            version=$(echo "$config" | cut -d':' -f1)
            os=$(echo "$config" | cut -d':' -f2)
            arch=$(echo "$config" | cut -d':' -f3)
            # Extract URL by removing version:os:arch: prefix and :checksum suffix
            temp_config="${config#*:*:*:}"  # Remove first 3 fields
            checksum="${temp_config##*:}"   # Get last field (checksum)
            url="${temp_config%:*}"         # Remove checksum, leaving URL
            
            log_info "Verifying checksum for: $tool_name"
            
            local temp_file="/tmp/${tool_name}-checksum-verify"
            if curl -fsSL -o "$temp_file" "$url"; then
                local actual_checksum
                actual_checksum=$(sha256sum "$temp_file" | cut -d' ' -f1)
                
                if [[ "$actual_checksum" == "$checksum" ]]; then
                    log_success "Checksum verified: $tool_name"
                else
                    log_error "Checksum mismatch: $tool_name"
                    log_error "Expected: $checksum"
                    log_error "Actual:   $actual_checksum"
                fi
                
                rm -f "$temp_file"
            else
                log_error "Failed to download for verification: $tool_name"
            fi
        fi
    done
}

# Function to clean cache
clean_cache() {
    log_info "Cleaning security cache..."
    
    if [[ -d "$CACHE_DIR" ]]; then
        rm -rf "$CACHE_DIR"
        log_success "Cache cleaned"
    else
        log_info "Cache already clean"
    fi
}

# Function to list installed tools
list_tools() {
    log_info "Installed security tools:"
    
    if [[ ! -d "$TOOLS_DIR" ]]; then
        log_info "No tools installed"
        return 0
    fi
    
    for tool_dir in "$TOOLS_DIR"/*; do
        if [[ -d "$tool_dir" ]]; then
            local tool_name
            tool_name=$(basename "$tool_dir")
            local version="unknown"
            
            if [[ -f "${tool_dir}/.version" ]]; then
                version=$(cat "${tool_dir}/.version")
            fi
            
            local status="❌"
            if [[ -x "${tool_dir}/${tool_name}" ]]; then
                status="✅"
            fi
            
            echo "  $status $tool_name ($version)"
        fi
    done
}

# Function to get tool path
get_tool_path() {
    local tool_name="$1"
    local tool_binary="${TOOLS_DIR}/${tool_name}/${tool_name}"
    
    if [[ -x "$tool_binary" ]]; then
        echo "$tool_binary"
    else
        log_error "Tool not installed: $tool_name"
        return 1
    fi
}

# Function to install all common tools
install_all() {
    local platform
    platform=$(detect_platform)
    
    log_info "Installing all security tools for platform: $platform"
    
    local tools=("golangci-lint" "hadolint" "trivy" "cosign")
    
    for tool in "${tools[@]}"; do
        if install_tool "$tool" "$platform"; then
            log_success "✅ $tool"
        else
            log_error "❌ $tool"
        fi
    done
    
    log_info "Installation complete. Run 'list' to see installed tools."
}

# Help function
show_help() {
    cat << EOF
CloudMCP Security Utilities

USAGE:
    $0 <command> [options]

COMMANDS:
    install <tool> [platform]   Install a verified tool
    install-all                 Install all common security tools
    list                        List installed tools
    get-path <tool>             Get path to installed tool
    clean                       Clean download cache
    update-checksums            Update and verify tool checksums (maintenance)
    help                        Show this help

EXAMPLES:
    # Install golangci-lint for current platform
    $0 install golangci-lint

    # Install hadolint for specific platform  
    $0 install hadolint linux-amd64

    # Install all common tools
    $0 install-all

    # Get path to installed tool
    PATH_TO_TOOL=\$($0 get-path golangci-lint)

    # List all installed tools
    $0 list

SUPPORTED TOOLS:
    - golangci-lint: Go linting tool
    - hadolint: Dockerfile linting tool  
    - trivy: Container security scanning
    - cosign: Container signing and verification

PLATFORMS:
    - linux-amd64, linux-arm64
    - darwin-amd64, darwin-arm64

SECURITY FEATURES:
    ✅ SHA256 checksum verification
    ✅ Secure download with retries
    ✅ Platform detection
    ✅ Tool verification after installation
    ✅ Cached downloads for efficiency

EOF
}

# Main function
main() {
    case "${1:-help}" in
        install)
            if [[ -z "${2:-}" ]]; then
                log_error "Tool name required"
                show_help
                exit 1
            fi
            install_tool "$2" "${3:-}" "${4:-false}"
            ;;
        install-all)
            install_all
            ;;
        list)
            list_tools
            ;;
        get-path)
            if [[ -z "${2:-}" ]]; then
                log_error "Tool name required"
                exit 1
            fi
            get_tool_path "$2"
            ;;
        clean)
            clean_cache
            ;;
        update-checksums)
            update_checksums
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "Unknown command: ${1:-}"
            show_help
            exit 1
            ;;
    esac
}

# Run main function
main "$@"