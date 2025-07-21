#!/bin/bash

# CloudMCP Simple SBOM Generator
# Generates a Software Bill of Materials for the CloudMCP project
# Usage: ./scripts/generate-sbom-simple.sh [--version VERSION] [--output DIR]

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
DEFAULT_OUTPUT_DIR="${PROJECT_ROOT}/build/sbom"

# Variables
VERSION=""
OUTPUT_DIR=""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging function
log_info() {
    echo -e "${BLUE}[INFO]${NC} $*"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $*"
}

# Usage function
show_usage() {
    cat << EOF
CloudMCP Simple SBOM Generator

USAGE:
    $0 [--version VERSION] [--output DIR]

OPTIONS:
    --version VERSION    Set the version for the SBOM (default: auto-detect from git)
    --output DIR         Output directory for SBOM files (default: build/sbom)
    --help               Show this help message

EXAMPLES:
    # Generate SBOM with auto-detected version
    $0

    # Generate SBOM with specific version and output directory
    $0 --version "1.2.3" --output "dist/sbom"

GENERATED FILES:
    - sbom.spdx.json     SPDX format SBOM
    - sbom.cyclonedx.json CycloneDX format SBOM
    - README.md          SBOM documentation
    - dependencies.txt   Plain text dependency list
EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --version)
            VERSION="$2"
            shift 2
            ;;
        --output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        --help|-h)
            show_usage
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Set defaults
if [[ -z "$OUTPUT_DIR" ]]; then
    OUTPUT_DIR="$DEFAULT_OUTPUT_DIR"
fi

if [[ -z "$VERSION" ]]; then
    # Auto-detect version from git
    VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "dev")
    VERSION=${VERSION#v} # Remove 'v' prefix if present
fi

log_info "Generating SBOM for CloudMCP version: $VERSION"
log_info "Output directory: $OUTPUT_DIR"

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Change to project root
cd "$PROJECT_ROOT"

# Check if syft is available, if not provide fallback
if command -v syft >/dev/null 2>&1; then
    log_info "Using syft for SBOM generation"
    
    # Generate SPDX format SBOM
    syft . -o spdx-json="$OUTPUT_DIR/sbom.spdx.json"
    
    # Generate CycloneDX format SBOM
    syft . -o cyclonedx-json="$OUTPUT_DIR/sbom.cyclonedx.json"
    
else
    log_info "syft not available, generating simple SBOM"
    
    # Generate simple SPDX-like JSON
    cat > "$OUTPUT_DIR/sbom.spdx.json" << EOF
{
  "SPDXID": "SPDXRef-DOCUMENT",
  "spdxVersion": "SPDX-2.3",
  "creationInfo": {
    "created": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "creators": ["Tool: CloudMCP-SBOM-Generator"],
    "licenseListVersion": "3.19"
  },
  "name": "CloudMCP",
  "documentNamespace": "https://github.com/chadit/CloudMCP/sbom-$VERSION",
  "packages": [
    {
      "SPDXID": "SPDXRef-Package-CloudMCP",
      "name": "CloudMCP",
      "downloadLocation": "https://github.com/chadit/CloudMCP",
      "filesAnalyzed": false,
      "packageVersion": "$VERSION",
      "supplier": "Person: chadit",
      "licenseConcluded": "NOASSERTION",
      "licenseDeclared": "NOASSERTION",
      "copyrightText": "NOASSERTION"
    }
  ]
}
EOF

    # Generate simple CycloneDX JSON
    cat > "$OUTPUT_DIR/sbom.cyclonedx.json" << EOF
{
  "bomFormat": "CycloneDX",
  "specVersion": "1.4",
  "serialNumber": "urn:uuid:$(uuidgen 2>/dev/null || echo "generated-uuid")",
  "version": 1,
  "metadata": {
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "tools": [
      {
        "vendor": "CloudMCP",
        "name": "simple-sbom-generator",
        "version": "1.0.0"
      }
    ],
    "component": {
      "type": "application",
      "name": "CloudMCP",
      "version": "$VERSION",
      "purl": "pkg:golang/github.com/chadit/CloudMCP@$VERSION"
    }
  },
  "components": [
    {
      "type": "application",
      "name": "CloudMCP",
      "version": "$VERSION",
      "purl": "pkg:golang/github.com/chadit/CloudMCP@$VERSION",
      "scope": "required"
    }
  ]
}
EOF
fi

# Generate Go dependencies list
log_info "Generating dependency list"
if [[ -f "go.mod" ]]; then
    go list -m all > "$OUTPUT_DIR/dependencies.txt"
else
    echo "No go.mod found" > "$OUTPUT_DIR/dependencies.txt"
fi

# Generate README
log_info "Generating SBOM documentation"
cat > "$OUTPUT_DIR/README.md" << EOF
# CloudMCP Software Bill of Materials (SBOM)

Generated: $(date -u +%Y-%m-%dT%H:%M:%SZ)
Version: $VERSION

## Files

- \`sbom.spdx.json\`: SPDX format Software Bill of Materials
- \`sbom.cyclonedx.json\`: CycloneDX format Software Bill of Materials
- \`dependencies.txt\`: Go module dependencies list
- \`README.md\`: This documentation file

## Description

This SBOM catalogs all dependencies and components used in CloudMCP version $VERSION.

## Formats

### SPDX 2.3
The Software Package Data Exchange (SPDX) format is an industry standard for communicating software bill of materials information.

### CycloneDX 1.4
CycloneDX is a lightweight SBOM standard designed for use in application security contexts and supply chain component analysis.

## Verification

To verify the authenticity of these SBOM files, check the digital signatures provided with the release artifacts.

## Statistics
EOF

# Add statistics if possible
if [[ -f "$OUTPUT_DIR/dependencies.txt" ]]; then
    dep_count=$(grep -c '^github.com\|^golang.org\|^gopkg.in' "$OUTPUT_DIR/dependencies.txt" 2>/dev/null || echo "0")
    echo "- Dependencies: $dep_count Go modules" >> "$OUTPUT_DIR/README.md"
fi

total_files=$(find "$OUTPUT_DIR" -type f | wc -l)
echo "- Generated files: $total_files" >> "$OUTPUT_DIR/README.md"

log_success "SBOM generation completed!"
log_info "Generated files:"
find "$OUTPUT_DIR" -type f -exec basename {} \; | sort | sed 's/^/  - /'

log_info "SBOM files are available in: $OUTPUT_DIR"