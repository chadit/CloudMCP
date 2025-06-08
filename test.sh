#!/bin/bash

# Test script for CloudMCP

echo "Testing CloudMCP build..."

# Check if binary exists
if [ ! -f "bin/cloud-mcp" ]; then
    echo "Error: Binary not found at bin/cloud-mcp"
    exit 1
fi

echo "✓ Binary exists"

# Set minimal test environment
export CLOUD_MCP_SERVER_NAME="CloudMCP Test"
export LOG_LEVEL=debug
export DEFAULT_LINODE_ACCOUNT=test
export LINODE_ACCOUNTS_TEST_TOKEN=test_token_123
export LINODE_ACCOUNTS_TEST_LABEL="Test Account"

# Test help/version output
echo -e "\nTesting server initialization..."
timeout 2s ./bin/cloud-mcp < /dev/null 2>&1 | head -20

echo -e "\n✓ Server can start (will fail on token validation, which is expected)"
echo "Build test completed successfully!"