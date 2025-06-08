#!/bin/bash

# Local MCP Server Testing Script

set -e

echo "CloudMCP Local Testing"
echo "====================="
echo

# Check if binary exists
if [ ! -f "../bin/cloud-mcp" ]; then
    echo "Error: Binary not found. Please run 'make build' first."
    exit 1
fi

# Check for .env file
if [ ! -f "../.env" ]; then
    echo "Error: .env file not found. Copy .env.example to .env and configure it."
    exit 1
fi

# Load environment variables
set -a
source ../.env
set +a

echo "1. Testing basic stdio communication..."
echo "--------------------------------------"
echo '{"jsonrpc": "2.0", "method": "initialize", "params": {"protocolVersion": "0.1.0", "capabilities": {}, "clientInfo": {"name": "test-script", "version": "1.0.0"}}, "id": 1}' | ../bin/cloud-mcp | jq .

echo -e "\n2. Listing available tools..."
echo "-----------------------------"
echo '{"jsonrpc": "2.0", "method": "tools/list", "params": {}, "id": 2}' | ../bin/cloud-mcp | jq '.result.tools[] | {name: .name, description: .description}'

echo -e "\n3. Testing tool execution..."
echo "---------------------------"
echo "Getting current account:"
cat <<EOF | ../bin/cloud-mcp | jq '.result'
{"jsonrpc": "2.0", "method": "initialize", "params": {"protocolVersion": "0.1.0", "capabilities": {}, "clientInfo": {"name": "test-script", "version": "1.0.0"}}, "id": 1}
{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "linode_account_get", "arguments": {}}, "id": 2}
EOF

echo -e "\nTesting complete!"
echo "================="
echo
echo "To run more tests:"
echo "  - Interactive client: make test-client"
echo "  - MCP Inspector: npm install -g @modelcontextprotocol/inspector && mcp-inspector ./bin/cloud-mcp"
echo "  - Custom commands: Edit test-commands.json and run: make test-stdio"