# CloudMCP

CloudMCP is a Model Context Protocol (MCP) server that enables Large Language Models (LLMs) to manage cloud infrastructure through natural language commands. Currently supporting Linode with multi-account capabilities, designed for seamless expansion to other cloud providers.

## Features

- **Multi-Account Support**: Manage multiple Linode accounts with easy switching
- **Secure Token Management**: Environment-based configuration with token sanitization
- **Comprehensive Linode Coverage**: Compute, networking, storage, and account management
- **Built-in Observability**: Prometheus metrics and structured logging
- **Thread-Safe Operations**: Concurrent request handling with proper synchronization
- **Extensible Architecture**: Plugin-ready design for additional cloud providers

## Prerequisites

- Go 1.24.3 or higher
- Linode API Personal Access Token(s)
- Unix-like environment (Linux, macOS)

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/chadit/CloudMCP.git
cd CloudMCP

# Build the binary
go build -o bin/cloud-mcp cmd/server/main.go

# Or install globally
go install github.com/chadit/CloudMCP/cmd/server@latest
```

## Configuration

CloudMCP uses environment variables for configuration. Create a `.env` file:

```bash
# Core Configuration
CLOUD_MCP_SERVER_NAME="Cloud MCP Server"
LOG_LEVEL=info                    # debug, info, warn, error
ENABLE_METRICS=true
METRICS_PORT=8080

# Linode Accounts
DEFAULT_LINODE_ACCOUNT=primary

# Primary Account
LINODE_ACCOUNTS_PRIMARY_TOKEN=your_production_token_here
LINODE_ACCOUNTS_PRIMARY_LABEL="Production"

# Development Account (optional)
LINODE_ACCOUNTS_DEV_TOKEN=your_dev_token_here
LINODE_ACCOUNTS_DEV_LABEL="Development"

# Additional accounts follow the same pattern:
# LINODE_ACCOUNTS_<NAME>_TOKEN=token
# LINODE_ACCOUNTS_<NAME>_LABEL="Display Name"
```

## Usage

### Running the Server

```bash
# Load environment and run
source .env && ./bin/cloud-mcp

# Or with go run
source .env && go run cmd/server/main.go
```

### Connecting with Claude Desktop

Add to your Claude Desktop configuration (`claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "cloud-mcp": {
      "command": "/path/to/cloud-mcp",
      "env": {
        "CLOUD_MCP_SERVER_NAME": "Cloud MCP",
        "LOG_LEVEL": "info",
        "DEFAULT_LINODE_ACCOUNT": "primary",
        "LINODE_ACCOUNTS_PRIMARY_TOKEN": "your_token_here",
        "LINODE_ACCOUNTS_PRIMARY_LABEL": "Production"
      }
    }
  }
}
```

### Connecting with GitHub Copilot

To use CloudMCP with GitHub Copilot Chat:

1. Install the GitHub Copilot Chat extension in VS Code
2. Configure the MCP server in VS Code settings:

```json
{
  "github.copilot.chat.mcpServers": {
    "cloud-mcp": {
      "command": "/path/to/cloud-mcp",
      "args": [],
      "env": {
        "CLOUD_MCP_SERVER_NAME": "Cloud MCP",
        "LOG_LEVEL": "info",
        "DEFAULT_LINODE_ACCOUNT": "primary",
        "LINODE_ACCOUNTS_PRIMARY_TOKEN": "your_token_here",
        "LINODE_ACCOUNTS_PRIMARY_LABEL": "Production"
      }
    }
  }
}
```

1. Restart VS Code to load the MCP server
2. Use `@mcp` in Copilot Chat to interact with your cloud infrastructure

Example prompts:

- `@mcp list all my Linode instances`
- `@mcp switch to development account`
- `@mcp show current account details`

### Available Tools

**Currently Implemented: 4 tools** | **Planned: 13 tools**

#### Account Management

- ✅ `linode_account_get` - Get current account information
- ✅ `linode_account_switch` - Switch between configured accounts
- ✅ `linode_account_list` - List all configured accounts

#### Compute Operations

- ✅ `linode_instances_list` - List all Linode instances
- ⏳ `linode_instance_get` - Get details of a specific instance *(not implemented yet)*
- ⏳ `linode_instance_create` - Create a new Linode instance *(not implemented yet)*
- ⏳ `linode_instance_delete` - Delete a Linode instance *(not implemented yet)*
- ⏳ `linode_instance_boot` - Boot a Linode instance *(not implemented yet)*
- ⏳ `linode_instance_shutdown` - Shutdown a Linode instance *(not implemented yet)*
- ⏳ `linode_instance_reboot` - Reboot a Linode instance *(not implemented yet)*

#### Networking

- ⏳ `linode_ips_list` - List IP addresses *(not implemented yet)*
- ⏳ `linode_ip_get` - Get IP address details *(not implemented yet)*

#### Storage

- ⏳ `linode_volumes_list` - List block storage volumes *(not implemented yet)*
- ⏳ `linode_volume_get` - Get volume details *(not implemented yet)*
- ⏳ `linode_volume_create` - Create a new volume *(not implemented yet)*
- ⏳ `linode_volume_delete` - Delete a volume *(not implemented yet)*
- ⏳ `linode_volume_attach` - Attach volume to instance *(not implemented yet)*
- ⏳ `linode_volume_detach` - Detach volume from instance *(not implemented yet)*

### Example Commands

Through an LLM interface:

```text
"Show me all my Linode instances"
"Create a new 2GB Linode in Newark with Ubuntu 22.04"
"Switch to the development account"
"List all volumes in the us-east region"
"Shutdown the web-server instance"
```

## Development

### Project Structure

```text
CloudMCP/
├── cmd/
│   └── server/              # Application entry point
├── internal/
│   ├── server/              # MCP server implementation
│   ├── services/
│   │   └── linode/          # Linode service implementation
│   └── config/              # Configuration management
├── pkg/
│   ├── logger/              # Logging abstraction
│   ├── types/               # Shared types
│   └── interfaces/          # Service interfaces
└── docs/                    # Additional documentation
```

### Building and Testing

```bash
# Run tests
go test ./...
go test ./... -race

# Run integration tests (requires LINODE_TEST_TOKEN)
go test ./internal/... -tags=integration

# Format code
gofumpt -w .

# Run linters
golangci-lint run

# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o bin/cloud-mcp-linux cmd/server/main.go
GOOS=darwin GOARCH=arm64 go build -o bin/cloud-mcp-darwin-arm64 cmd/server/main.go
```

### Testing MCP Server Locally

You can test the MCP server without Claude or Copilot using these methods:

#### 1. Direct stdio Testing

Create a test file `test-commands.json`:

```json
{"jsonrpc": "2.0", "method": "initialize", "params": {"protocolVersion": "0.1.0", "capabilities": {}, "clientInfo": {"name": "test-client", "version": "1.0.0"}}, "id": 1}
{"jsonrpc": "2.0", "method": "tools/list", "params": {}, "id": 2}
{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "linode_account_get", "arguments": {}}, "id": 3}
```

Then run:

```bash
cat test-commands.json | ./bin/cloud-mcp
```

#### 2. Using the MCP Inspector

Install and use the MCP Inspector tool:

```bash
# Install MCP Inspector
npm install -g @modelcontextprotocol/inspector

# Run your server with the inspector
mcp-inspector ./bin/cloud-mcp
```

This opens a web interface at `http://localhost:5173` where you can:

- See all available tools
- Execute tools interactively
- View request/response logs
- Test different scenarios

#### 3. Create a Test Client

Use the test client in `test/client/`:

```bash
cd test/client
go run main.go
```

This provides an interactive CLI to test your MCP server.

#### 4. Using curl with SSE Transport

If you modify the server to support HTTP transport:

```bash
# Start server in HTTP mode
./bin/cloud-mcp --transport http --port 8080

# Test with curl
curl -X POST http://localhost:8080/sse \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc": "2.0", "method": "tools/list", "params": {}, "id": 1}'
```

### Adding New Tools

1. Define types in `internal/services/linode/types.go`
2. Implement handler in `internal/services/linode/tools_*.go`
3. Register tool in service initialization
4. Add tests in corresponding `*_test.go` file

## Monitoring

When metrics are enabled, Prometheus metrics are available at `http://localhost:8080/metrics`:

- `cloudmcp_tool_requests_total` - Total tool requests by tool and status
- `cloudmcp_tool_duration_seconds` - Tool execution duration
- `cloudmcp_linode_api_requests_total` - Linode API requests by endpoint
- `cloudmcp_linode_api_duration_seconds` - Linode API call duration

## Security Considerations

- API tokens are never logged or returned in responses
- All sensitive data is sanitized in error messages
- Environment-based configuration keeps secrets out of code
- Rate limiting protects against API abuse
- Request validation prevents malformed inputs

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please ensure:

- All tests pass (`go test ./...`)
- Code is formatted (`gofumpt -w .`)
- Linting passes (`golangci-lint run`)
- Commits follow conventional format

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [MCP SDK for Go](https://github.com/mark3labs/mcp-go) - MCP protocol implementation
- [Linode API](https://www.linode.com/api/v4) - Cloud infrastructure API
- [slog](https://pkg.go.dev/log/slog) - Structured logging

## Support

- Issues: [GitHub Issues](https://github.com/chadit/CloudMCP/issues)
- Discussions: [GitHub Discussions](https://github.com/chadit/CloudMCP/discussions)
- Documentation: See [CLAUDE.md](CLAUDE.md) for development guidance
