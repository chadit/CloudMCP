#!/bin/bash

# CloudMCP SBOM Generation Script
# Generates Software Bill of Materials in multiple formats
# Usage: ./scripts/generate-sbom.sh [options]

set -euo pipefail

# Script configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
BUILD_DIR="${PROJECT_ROOT}/build"
SBOM_DIR="${PROJECT_ROOT}/build/sbom"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
OUTPUT_FORMATS=("spdx" "cyclonedx")
SIGN_SBOM=false
VERBOSE=false
SCAN_VULNERABILITIES=false
COMPONENT_NAME="CloudMCP"
COMPONENT_VERSION=""
SUPPLIER="chadit"
NAMESPACE="https://github.com/chadit/CloudMCP"

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

# Help function
show_help() {
    cat << EOF
CloudMCP SBOM Generation Script

USAGE:
    $0 [OPTIONS]

OPTIONS:
    -h, --help              Show this help message
    -f, --format FORMAT     SBOM format (spdx, cyclonedx, or both) [default: both]
    -v, --version VERSION   Component version (auto-detected if not specified)
    -s, --sign              Sign SBOM files with cosign
    --scan                  Include vulnerability scanning results
    --verbose               Enable verbose output
    --build-dir DIR         Specify build directory [default: ./build]
    --output-dir DIR        Specify SBOM output directory [default: ./build/sbom]

EXAMPLES:
    # Generate SBOM in all formats
    $0

    # Generate SPDX format only
    $0 --format spdx

    # Generate and sign SBOM
    $0 --sign

    # Generate with vulnerability scanning
    $0 --scan --sign

SUPPORTED FORMATS:
    - SPDX: Software Package Data Exchange (JSON and tag-value)
    - CycloneDX: Industry standard for security-focused SBOM

PREREQUISITES:
    - syft: Install via 'curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh'
    - cosign: For signing (optional, installed via security-utils.sh)
    - grype: For vulnerability scanning (optional)

EOF
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check for syft
    if ! command -v syft &> /dev/null; then
        log_error "syft is not installed"
        log_info "Install syft: curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh"
        exit 1
    fi
    
    local syft_version
    syft_version=$(syft version --output json 2>/dev/null | jq -r '.version' || echo "unknown")
    log_verbose "Using syft version: ${syft_version}"
    
    # Check for cosign if signing is enabled
    if [[ "${SIGN_SBOM}" == "true" ]]; then
        if ! command -v cosign &> /dev/null; then
            log_warning "cosign not found in PATH, checking security-utils..."
            if [[ -x "${PROJECT_ROOT}/scripts/security-utils.sh" ]]; then
                local cosign_path
                if cosign_path=$("${PROJECT_ROOT}/scripts/security-utils.sh" get-path cosign 2>/dev/null); then
                    log_info "Using cosign from security-utils: $cosign_path"
                    ln -sf "$cosign_path" /tmp/cosign
                    export PATH="/tmp:$PATH"
                else
                    log_error "cosign required for signing but not available"
                    exit 1
                fi
            else
                log_error "cosign required for signing but not available"
                exit 1
            fi
        fi
    fi
    
    # Check for grype if vulnerability scanning is enabled
    if [[ "${SCAN_VULNERABILITIES}" == "true" ]]; then
        if ! command -v grype &> /dev/null; then
            log_warning "grype not found for vulnerability scanning"
            log_info "Install grype: curl -sSfL https://raw.githubusercontent.com/anchore/grype/main/install.sh | sh"
            SCAN_VULNERABILITIES=false
        fi
    fi
    
    log_success "Prerequisites check completed"
}

# Detect component version
detect_version() {
    if [[ -z "${COMPONENT_VERSION}" ]]; then
        # Try to get version from version.go
        if [[ -f "${PROJECT_ROOT}/internal/version/version.go" ]]; then
            COMPONENT_VERSION=$(grep -o 'Version = "[^"]*"' "${PROJECT_ROOT}/internal/version/version.go" | cut -d'"' -f2 || echo "")
        fi
        
        # Fallback to git tag
        if [[ -z "${COMPONENT_VERSION}" ]]; then
            COMPONENT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "dev")
        fi
        
        # Fallback to git commit
        if [[ "${COMPONENT_VERSION}" == "dev" ]] || [[ -z "${COMPONENT_VERSION}" ]]; then
            local git_commit
            git_commit=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
            COMPONENT_VERSION="dev-${git_commit}"
        fi
    fi
    
    log_info "Component version: ${COMPONENT_VERSION}"
}

# Generate Go module information
generate_go_mod_info() {
    log_info "Analyzing Go modules..."
    
    cd "${PROJECT_ROOT}"
    
    # Create temporary file for module analysis
    local mod_info_file="${SBOM_DIR}/go-modules.json"
    
    # Get direct dependencies
    go list -m -json all > "${mod_info_file}"
    
    log_verbose "Go module information saved to: ${mod_info_file}"
    
    # Count dependencies
    local dep_count
    dep_count=$(go list -m all | grep -v "^$(go list -m)" | wc -l)
    log_info "Found ${dep_count} Go module dependencies"
}

# Generate SPDX SBOM
generate_spdx() {
    log_info "Generating SPDX SBOM..."
    
    local spdx_file="${SBOM_DIR}/sbom.spdx.json"
    local spdx_tv_file="${SBOM_DIR}/sbom.spdx"
    
    # Generate SPDX JSON format
    syft "${PROJECT_ROOT}" \
        --output spdx-json="${spdx_file}" \
        --name "${COMPONENT_NAME}" \
        --version "${COMPONENT_VERSION}" \
        --source-name "${COMPONENT_NAME}" \
        --source-version "${COMPONENT_VERSION}" \
        --select-catalogers "+go-module-binary,+go-module"
    
    # Generate SPDX tag-value format
    syft "${PROJECT_ROOT}" \
        --output spdx-tag-value="${spdx_tv_file}" \
        --name "${COMPONENT_NAME}" \
        --version "${COMPONENT_VERSION}" \
        --source-name "${COMPONENT_NAME}" \
        --source-version "${COMPONENT_VERSION}" \
        --select-catalogers "+go-module-binary,+go-module"
    
    # Enhance SPDX with additional metadata
    if command -v jq &> /dev/null; then
        local enhanced_spdx="${SBOM_DIR}/sbom.spdx.enhanced.json"
        jq --arg supplier "$SUPPLIER" \
           --arg namespace "$NAMESPACE" \
           --arg timestamp "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
           '. + {
               "creationInfo": (.creationInfo + {
                   "created": $timestamp,
                   "creators": (.creators + ["Tool: CloudMCP-SBOM-Generator"]),
                   "licenseListVersion": "3.21"
               }),
               "documentDescribes": [.SPDXID],
               "documentNamespace": ($namespace + "/sbom/" + $timestamp)
           }' "${spdx_file}" > "${enhanced_spdx}"
        
        mv "${enhanced_spdx}" "${spdx_file}"
        log_verbose "Enhanced SPDX metadata"
    fi
    
    log_success "SPDX SBOM generated: ${spdx_file}, ${spdx_tv_file}"
}

# Generate CycloneDX SBOM
generate_cyclonedx() {
    log_info "Generating CycloneDX SBOM..."
    
    local cyclonedx_file="${SBOM_DIR}/sbom.cdx.json"
    local cyclonedx_xml_file="${SBOM_DIR}/sbom.cdx.xml"
    
    # Generate CycloneDX JSON format
    syft "${PROJECT_ROOT}" \
        --output cyclonedx-json="${cyclonedx_file}" \
        --name "${COMPONENT_NAME}" \
        --version "${COMPONENT_VERSION}" \
        --source-name "${COMPONENT_NAME}" \
        --source-version "${COMPONENT_VERSION}" \
        --select-catalogers "+go-module-binary,+go-module"
    
    # Generate CycloneDX XML format
    syft "${PROJECT_ROOT}" \
        --output cyclonedx-xml="${cyclonedx_xml_file}" \
        --name "${COMPONENT_NAME}" \
        --version "${COMPONENT_VERSION}" \
        --source-name "${COMPONENT_NAME}" \
        --source-version "${COMPONENT_VERSION}" \
        --select-catalogers "+go-module-binary,+go-module"
    
    # Enhance CycloneDX with additional metadata
    if command -v jq &> /dev/null; then
        local enhanced_cdx="${SBOM_DIR}/sbom.cdx.enhanced.json"
        jq --arg supplier "$SUPPLIER" \
           --arg namespace "$NAMESPACE" \
           --arg timestamp "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
           '. + {
               "metadata": (.metadata + {
                   "timestamp": $timestamp,
                   "supplier": {
                       "name": $supplier,
                       "url": [$namespace]
                   },
                   "licenses": [
                       {
                           "license": {
                               "name": "MIT"
                           }
                       }
                   ]
               })
           }' "${cyclonedx_file}" > "${enhanced_cdx}"
        
        mv "${enhanced_cdx}" "${cyclonedx_file}"
        log_verbose "Enhanced CycloneDX metadata"
    fi
    
    log_success "CycloneDX SBOM generated: ${cyclonedx_file}, ${cyclonedx_xml_file}"
}

# Generate vulnerability report
generate_vulnerability_report() {
    if [[ "${SCAN_VULNERABILITIES}" == "false" ]]; then
        return 0
    fi
    
    log_info "Generating vulnerability report..."
    
    local vuln_report="${SBOM_DIR}/vulnerabilities.json"
    local vuln_report_txt="${SBOM_DIR}/vulnerabilities.txt"
    
    # Scan for vulnerabilities using grype
    if command -v grype &> /dev/null; then
        grype "${PROJECT_ROOT}" \
            --output json \
            --file "${vuln_report}" \
            --fail-on=""
        
        grype "${PROJECT_ROOT}" \
            --output table \
            --file "${vuln_report_txt}" \
            --fail-on=""
        
        # Generate summary
        if command -v jq &> /dev/null && [[ -f "${vuln_report}" ]]; then
            local vuln_summary="${SBOM_DIR}/vulnerability-summary.json"
            jq '{
                "scan_timestamp": .descriptor.timestamp,
                "total_vulnerabilities": (.matches | length),
                "by_severity": (
                    .matches | 
                    group_by(.vulnerability.severity) | 
                    map({
                        "severity": .[0].vulnerability.severity,
                        "count": length
                    })
                ),
                "by_package": (
                    .matches | 
                    group_by(.artifact.name) | 
                    map({
                        "package": .[0].artifact.name,
                        "version": .[0].artifact.version,
                        "vulnerability_count": length,
                        "highest_severity": (map(.vulnerability.severity) | max)
                    }) |
                    sort_by(.vulnerability_count) |
                    reverse
                )
            }' "${vuln_report}" > "${vuln_summary}"
            
            log_info "Vulnerability summary:"
            jq -r '.by_severity[] | "  \(.severity): \(.count)"' "${vuln_summary}"
        fi
        
        log_success "Vulnerability report generated: ${vuln_report}"
    else
        log_warning "grype not available, skipping vulnerability scanning"
    fi
}

# Sign SBOM files
sign_sbom_files() {
    if [[ "${SIGN_SBOM}" == "false" ]]; then
        return 0
    fi
    
    log_info "Signing SBOM files..."
    
    local signed_count=0
    
    # Sign all SBOM files
    for file in "${SBOM_DIR}"/sbom.*; do
        if [[ -f "$file" ]] && [[ ! "$file" == *.sig ]]; then
            log_verbose "Signing: $(basename "$file")"
            
            # Sign with cosign
            cosign sign-blob \
                --yes \
                --bundle "${file}.cosign.bundle" \
                "$file"
            
            ((signed_count++))
        fi
    done
    
    if [[ "${signed_count}" -gt 0 ]]; then
        log_success "Signed ${signed_count} SBOM files"
    else
        log_warning "No SBOM files found to sign"
    fi
}

# Generate SBOM summary
generate_summary() {
    log_info "Generating SBOM summary..."
    
    local summary_file="${SBOM_DIR}/sbom-summary.json"
    
    # Create summary with metadata
    cat > "${summary_file}" << EOF
{
    "component": {
        "name": "${COMPONENT_NAME}",
        "version": "${COMPONENT_VERSION}",
        "supplier": "${SUPPLIER}",
        "namespace": "${NAMESPACE}"
    },
    "generation": {
        "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
        "tool": "CloudMCP-SBOM-Generator",
        "formats": $(printf '%s\n' "${OUTPUT_FORMATS[@]}" | jq -R . | jq -s .),
        "signed": ${SIGN_SBOM},
        "vulnerability_scan": ${SCAN_VULNERABILITIES}
    },
    "files": {
EOF

    # Add file information
    local first=true
    for file in "${SBOM_DIR}"/sbom.*; do
        if [[ -f "$file" ]] && [[ ! "$file" == *.bundle ]]; then
            if [[ "$first" == "true" ]]; then
                first=false
            else
                echo "," >> "${summary_file}"
            fi
            
            local filename
            filename=$(basename "$file")
            local filesize
            filesize=$(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null)
            
            printf '        "%s": {"size": %d, "type": "%s"}' \
                "$filename" \
                "$filesize" \
                "${filename##*.}" >> "${summary_file}"
        fi
    done
    
    cat >> "${summary_file}" << EOF

    }
}
EOF

    log_success "SBOM summary generated: ${summary_file}"
}

# Show SBOM information
show_sbom_info() {
    log_info "SBOM Generation Complete"
    log_info "========================"
    
    echo "Component: ${COMPONENT_NAME} v${COMPONENT_VERSION}"
    echo "Output directory: ${SBOM_DIR}"
    echo "Formats generated: $(printf '%s ' "${OUTPUT_FORMATS[@]}")"
    echo "Signed: ${SIGN_SBOM}"
    echo "Vulnerability scan: ${SCAN_VULNERABILITIES}"
    echo ""
    
    echo "Generated files:"
    if [[ -d "${SBOM_DIR}" ]]; then
        ls -la "${SBOM_DIR}/"
    fi
    
    echo ""
    echo "Usage examples:"
    echo "  # Verify SBOM signature (if signed)"
    echo "  cosign verify-blob --bundle sbom.spdx.json.cosign.bundle sbom.spdx.json"
    echo ""
    echo "  # View component summary"
    echo "  jq '.component' ${SBOM_DIR}/sbom-summary.json"
    echo ""
    echo "  # List dependencies from SPDX"
    echo "  jq '.packages[] | select(.name != \"${COMPONENT_NAME}\") | {name, versionInfo}' ${SBOM_DIR}/sbom.spdx.json"
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -f|--format)
                case "$2" in
                    spdx)
                        OUTPUT_FORMATS=("spdx")
                        ;;
                    cyclonedx)
                        OUTPUT_FORMATS=("cyclonedx")
                        ;;
                    both)
                        OUTPUT_FORMATS=("spdx" "cyclonedx")
                        ;;
                    *)
                        log_error "Invalid format: $2. Use spdx, cyclonedx, or both"
                        exit 1
                        ;;
                esac
                shift 2
                ;;
            -v|--version)
                COMPONENT_VERSION="$2"
                shift 2
                ;;
            -s|--sign)
                SIGN_SBOM=true
                shift
                ;;
            --scan)
                SCAN_VULNERABILITIES=true
                shift
                ;;
            --verbose)
                VERBOSE=true
                shift
                ;;
            --build-dir)
                BUILD_DIR="$2"
                SBOM_DIR="${BUILD_DIR}/sbom"
                shift 2
                ;;
            --output-dir)
                SBOM_DIR="$2"
                shift 2
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# Main function
main() {
    log_info "CloudMCP SBOM Generator"
    log_info "======================"
    
    parse_args "$@"
    check_prerequisites
    detect_version
    
    # Create output directory
    mkdir -p "${SBOM_DIR}"
    
    # Generate Go module information
    generate_go_mod_info
    
    # Generate SBOM in requested formats
    for format in "${OUTPUT_FORMATS[@]}"; do
        case "$format" in
            spdx)
                generate_spdx
                ;;
            cyclonedx)
                generate_cyclonedx
                ;;
        esac
    done
    
    # Generate vulnerability report if requested
    generate_vulnerability_report
    
    # Sign SBOM files if requested
    sign_sbom_files
    
    # Generate summary
    generate_summary
    
    # Show information
    show_sbom_info
    
    log_success "SBOM generation completed successfully!"
}

# Run main function with all arguments
main "$@"