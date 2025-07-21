#!/bin/bash

# CloudMCP Simplified SBOM Generation Script
# Lightweight SBOM generation without over-engineering
# Usage: ./scripts/generate-sbom-simple.sh

set -euo pipefail

# Script configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
SBOM_DIR="${PROJECT_ROOT}/build/sbom"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $*"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $*"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*" >&2
}

# Help function
show_help() {
    cat << EOF
CloudMCP Simplified SBOM Generation

USAGE:
    $0 [OPTIONS]

OPTIONS:
    -h, --help              Show this help message
    -v, --version VERSION   Component version (auto-detected if not specified)
    -o, --output DIR        Output directory [default: ./build/sbom]

DESCRIPTION:
    Generates a lightweight Software Bill of Materials (SBOM) in SPDX and CycloneDX formats.
    This is a simplified version that focuses on essential functionality without complexity.

EXAMPLES:
    # Generate SBOM with auto-detected version
    $0

    # Generate with specific version
    $0 --version v1.2.3

    # Generate to custom directory
    $0 --output /tmp/sbom

EOF
}

# Detect component version
detect_version() {
    local version=""
    
    # Try to get version from version.go
    if [[ -f "${PROJECT_ROOT}/internal/version/version.go" ]]; then
        version=$(grep -o 'Version = "[^"]*"' "${PROJECT_ROOT}/internal/version/version.go" | cut -d'"' -f2 2>/dev/null || echo "")
    fi
    
    # Fallback to git tag
    if [[ -z "$version" ]]; then
        version=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
    fi
    
    # Fallback to git commit
    if [[ -z "$version" ]]; then
        local git_commit
        git_commit=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
        version="dev-${git_commit}"
    fi
    
    echo "$version"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    if ! command -v syft &> /dev/null; then
        log_error "syft is not installed"
        log_info "Install syft: curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh"
        exit 1
    fi
    
    local syft_version
    syft_version=$(syft version --output json 2>/dev/null | jq -r '.version' 2>/dev/null || syft version 2>/dev/null | head -1 || echo "unknown")
    log_info "Using syft version: ${syft_version}"
}

# Generate SBOMs
generate_sboms() {
    local version="$1"
    
    log_info "Generating SBOMs for CloudMCP ${version}..."
    
    # Create output directory
    mkdir -p "${SBOM_DIR}"
    
    cd "${PROJECT_ROOT}"
    
    # Generate SPDX format
    log_info "Generating SPDX SBOM..."
    syft . -o spdx-json="${SBOM_DIR}/sbom.spdx.json" --quiet
    
    # Generate CycloneDX format  
    log_info "Generating CycloneDX SBOM..."
    syft . -o cyclonedx-json="${SBOM_DIR}/sbom.cyclonedx.json" --quiet
    
    # Create simple text summary
    log_info "Creating summary report..."
    cat > "${SBOM_DIR}/README.md" << EOF
# CloudMCP Software Bill of Materials (SBOM)

**Component:** CloudMCP  
**Version:** ${version}  
**Generated:** $(date -u +%Y-%m-%dT%H:%M:%SZ)  
**Tool:** Syft $(syft version --output json 2>/dev/null | jq -r '.version' 2>/dev/null || echo "unknown")

## Files

- **sbom.spdx.json**: SPDX format SBOM (industry standard)
- **sbom.cyclonedx.json**: CycloneDX format SBOM (security-focused)

## Usage

### View Dependencies
\`\`\`bash
# View all packages from SPDX
jq '.packages[] | select(.name != "CloudMCP") | {name, versionInfo}' sbom.spdx.json

# View components from CycloneDX  
jq '.components[] | {name, version, type}' sbom.cyclonedx.json
\`\`\`

### Verification
These SBOMs provide a complete inventory of all software components, libraries, 
and dependencies used in CloudMCP. They can be used for:

- Supply chain security analysis
- License compliance verification  
- Vulnerability management
- Dependency tracking

## Statistics

EOF

    # Add basic statistics if jq is available
    if command -v jq >/dev/null 2>&1; then
        local spdx_packages cyclonedx_components
        spdx_packages=$(jq '[.packages[] | select(.name != "CloudMCP")] | length' "${SBOM_DIR}/sbom.spdx.json" 2>/dev/null || echo "unknown")
        cyclonedx_components=$(jq '[.components[]?] | length' "${SBOM_DIR}/sbom.cyclonedx.json" 2>/dev/null || echo "unknown")
        
        cat >> "${SBOM_DIR}/README.md" << EOF
- **Total packages (SPDX):** ${spdx_packages}
- **Total components (CycloneDX):** ${cyclonedx_components}
- **Go modules analyzed:** Yes
- **Container image scanned:** No (source-only SBOM)

EOF
    else
        echo "- Install \`jq\` for detailed statistics" >> "${SBOM_DIR}/README.md"
        echo "" >> "${SBOM_DIR}/README.md"
    fi
    
    # Create checksums
    log_info "Generating checksums..."
    cd "${SBOM_DIR}"
    sha256sum *.json > checksums.txt
    
    log_success "SBOM generation completed!"
    log_info "Output directory: ${SBOM_DIR}"
    
    # Show file listing
    echo ""
    echo "Generated files:"
    ls -la "${SBOM_DIR}/"
}

# Main function
main() {
    local version=""
    local output_dir=""
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -v|--version)
                version="$2"
                shift 2
                ;;
            -o|--output)
                output_dir="$2"
                shift 2
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # Set output directory if specified
    if [[ -n "$output_dir" ]]; then
        SBOM_DIR="$output_dir"
    fi
    
    # Detect version if not provided
    if [[ -z "$version" ]]; then
        version=$(detect_version)
    fi
    
    log_info "CloudMCP Simplified SBOM Generator"
    log_info "================================="
    log_info "Version: ${version}"
    log_info "Output: ${SBOM_DIR}"
    
    # Check prerequisites and generate
    check_prerequisites
    generate_sboms "$version"
    
    echo ""
    log_success "SBOM generation completed successfully!"
    
    # Usage examples
    echo ""
    echo "Usage examples:"
    echo "  # View dependencies"
    echo "  jq '.packages[] | select(.name != \"CloudMCP\") | {name, versionInfo}' ${SBOM_DIR}/sbom.spdx.json"
    echo ""
    echo "  # Verify checksums"  
    echo "  cd ${SBOM_DIR} && sha256sum -c checksums.txt"
}

# Run main function with all arguments
main "$@"