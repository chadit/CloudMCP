#!/bin/bash
set -euo pipefail

# CloudMCP Wrapper Script for MCP Clients
# This script loads environment variables and starts the CloudMCP server

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
ENV_FILE="${SCRIPT_DIR}/.env"
ENV_EXAMPLE="${SCRIPT_DIR}/.env.example"
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

# Check if .env file exists, create from example if not
if [[ ! -f "$ENV_FILE" ]]; then
	if [[ -f "$ENV_EXAMPLE" ]]; then
		warning "No .env file found. Creating from .env.example..."
		cp "$ENV_EXAMPLE" "$ENV_FILE"
		info "Please edit $ENV_FILE with your Linode API tokens"
		info "You need to set at least:"
		echo "  - LINODE_ACCOUNTS_PRIMARY_TOKEN=your_token_here"
		echo ""
		info "Opening .env file in default editor..."
		if [[ -n "${EDITOR:-}" ]]; then
			"$EDITOR" "$ENV_FILE"
		elif command -v nano &>/dev/null; then
			nano "$ENV_FILE"
		elif command -v vi &>/dev/null; then
			vi "$ENV_FILE"
		else
			warning "No editor found. Please manually edit $ENV_FILE"
		fi
	else
		error "No .env or .env.example file found"
		error "Please create $ENV_FILE with your configuration"
		exit 1
	fi
fi

# Load environment variables from .env file
if [[ -f "$ENV_FILE" ]]; then
	# Use set -a to export all variables, then source the file
	set -a
	# shellcheck disable=SC1090
	source "$ENV_FILE"
	set +a
else
	error "Configuration file not found: $ENV_FILE"
	exit 1
fi

# Set default values if not already set
export CLOUD_MCP_SERVER_NAME="${CLOUD_MCP_SERVER_NAME:-Cloud MCP Server}"
export LOG_LEVEL="${LOG_LEVEL:-info}"

# Validate that at least one Linode account is configured
if [[ -z "${DEFAULT_LINODE_ACCOUNT:-}" ]]; then
	error "DEFAULT_LINODE_ACCOUNT is not set"
	info "Please configure your environment variables in $ENV_FILE"
	info "Example:"
	echo "  DEFAULT_LINODE_ACCOUNT=primary"
	echo "  LINODE_ACCOUNTS_PRIMARY_TOKEN=your_token_here"
	echo "  LINODE_ACCOUNTS_PRIMARY_LABEL=\"Production\""
	exit 1
fi

# Check if the default account token is set
# Convert account name to uppercase for the token variable
ACCOUNT_UPPER=$(echo "$DEFAULT_LINODE_ACCOUNT" | tr '[:lower:]' '[:upper:]')
TOKEN_VAR="LINODE_ACCOUNTS_${ACCOUNT_UPPER}_TOKEN"
TOKEN_VALUE="${!TOKEN_VAR:-}"

if [[ -z "$TOKEN_VALUE" ]] || [[ "$TOKEN_VALUE" == "your_"*"_token_here" ]] || [[ "$TOKEN_VALUE" == "test_token_"* ]]; then
	error "Token for default account '$DEFAULT_LINODE_ACCOUNT' is not set or is a placeholder"
	info "Please set $TOKEN_VAR in $ENV_FILE with your actual Linode API token"
	info ""
	info "To get a Linode API token:"
	echo "  1. Log in to https://cloud.linode.com"
	echo "  2. Click on your username in the top right"
	echo "  3. Select 'API Tokens'"
	echo "  4. Create a Personal Access Token with the required permissions"
	exit 1
fi

# Check if running in debug mode
if [[ "${DEBUG:-}" == "true" ]] || [[ "${LOG_LEVEL}" == "debug" ]]; then
	info "Debug mode enabled"
	info "Configuration loaded from: $ENV_FILE"
	info "Default account: $DEFAULT_LINODE_ACCOUNT"
	info "Server name: $CLOUD_MCP_SERVER_NAME"
fi

# Start the CloudMCP server
exec "$BINARY" "$@"

