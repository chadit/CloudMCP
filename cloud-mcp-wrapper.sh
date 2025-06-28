#!/bin/bash
set -euo pipefail

# CloudMCP Wrapper Script for MCP Clients
# This script starts the CloudMCP server with TOML configuration

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
error() {
	echo -e "${RED}Error:${NC} $1" >&2
}

warning() {
	echo -e "${YELLOW}Warning:${NC} $1" >&2
}

info() {
	echo -e "${GREEN}Info:${NC} $1" >&2
}

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Default paths
BINARY="${SCRIPT_DIR}/bin/cloud-mcp"

# Check if binary exists
if [[ ! -f "$BINARY" ]]; then
	error "CloudMCP binary not found at $BINARY"
	info "Building the binary for you..."

	# Check if go is installed
	if ! command -v go &>/dev/null; then
		error "Go is not installed. Please install Go 1.24.3 or higher"
		exit 1
	fi

	# Build the binary
	if go build -o "$BINARY" "${SCRIPT_DIR}/cmd/server/main.go"; then
		info "Binary built successfully!"
	else
		error "Failed to build binary"
		exit 1
	fi
fi

# CloudMCP uses TOML configuration automatically
# First run will create default config with placeholder tokens
# Use the MCP account management commands to add your tokens:
#   cloudmcp_account_list
#   cloudmcp_account_add --name primary --token YOUR_TOKEN --label "Primary Account"

# Check if running in debug mode
if [[ "${DEBUG:-}" == "true" ]]; then
	info "Debug mode enabled - using TOML configuration"
fi

# Start the CloudMCP server
exec "$BINARY" "$@"

