# CloudMCP

CloudMCP is a minimal Model Context Protocol (MCP) server shell designed as a
foundation for cloud infrastructure management through natural language commands.

**Current Status**: 🎉 **Minimal Shell Architecture** - Clean Foundation
Ready for Extension

## 🚀 Quick Start

### Prerequisites

You need one of these MCP-compatible AI tools:

- [Claude Desktop](https://claude.ai/desktop) (recommended)
- [Claude Code](https://claude.ai/code)
- [VS Code with MCP Extension](https://marketplace.visualstudio.com/items?itemName=mark3labs.mcp)

### Installation Methods

#### Option 1: Go Install (Recommended)

```bash
# Install CloudMCP server
go install github.com/chadit/CloudMCP/cmd/cloud-mcp@latest

# Add to Claude Code
claude mcp add cloudmcp -- cloud-mcp

# Test it works
# Ask your AI assistant: "Check CloudMCP health status"
```

#### Option 2: Claude Desktop Integration

1. **Install the server:**

   ```bash
   go install github.com/chadit/CloudMCP/cmd/cloud-mcp@latest
   ```

2. **Configure Claude Desktop** - Edit `claude_desktop_config.json`:

   ```json
   {
     "mcpServers": {
       "cloudmcp": {
         "command": "cloud-mcp",
         "args": []
       }
     }
   }
   ```

3. **Restart Claude Desktop**

#### Option 3: VSCode MCP Support

1. **Install the server:**

   ```bash
   go install github.com/chadit/CloudMCP/cmd/cloud-mcp@latest
   ```

2. **Install MCP Extension** from VS Code marketplace

3. **The project already includes `.vscode/settings.json`** for automatic
   configuration

### Test It Works

Ask your AI assistant:
*"Check CloudMCP health status"*

**That's it!** You now have a minimal MCP server foundation ready for
extension.

## 💡 Current Capabilities

CloudMCP is currently a **minimal shell** providing foundation functionality:

### Health and System Tools

- *"Check CloudMCP health status"* - Server health and service discovery
- *"Show server information"* - Get server details and uptime
- *"What tools are available?"* - List available MCP tools

### System Features

- Health status monitoring with sub-microsecond response times
- Service discovery and capability listing
- Metrics collection via Prometheus endpoints
- Secure configuration with optional TLS and authentication

### Framework Ready

The framework is ready for:

- **Provider Registry**: Ready for cloud provider implementations
- **Tool Interface**: Clean tool implementation pattern
- **Metrics System**: Unified metrics with Prometheus backend
- **Security**: Authentication, rate limiting, and TLS support

**Architecture designed for extensibility** - add new providers by
implementing the `CloudProvider` interface.

## 📦 Detailed Installation

### From Source (Developers)

For development and customization:

```bash
git clone https://github.com/chadit/CloudMCP.git
cd CloudMCP
go build -o bin/cloud-mcp cmd/cloud-mcp/main.go

# Run locally for development
go run cmd/cloud-mcp/main.go

# Run tests
go test ./...
```

## ⚙️ Configuration

CloudMCP automatically creates TOML configuration files on first run:

**Config locations:**

- **Linux**: `~/.config/cloudmcp/config.toml`
- **macOS**: `~/Library/Application Support/CloudMCP/config.toml`
- **Windows**: `%APPDATA%\CloudMCP\config.toml`

**Basic configuration:**

```toml
[system]
server_name = "CloudMCP Shell"
log_level = "info"
enable_metrics = true
metrics_port = 8080

# Logging configuration
log_max_size = 10      # MB
log_max_backups = 5    # Number of files to keep
log_max_age = 30       # Days to retain logs
```

### Environment Variables

You can override configuration with environment variables:

```bash
export CLOUD_MCP_SERVER_NAME="My CloudMCP"
export LOG_LEVEL="debug"
export ENABLE_METRICS="true"
export METRICS_PORT="8080"

# Optional metrics security
export METRICS_AUTH_USERNAME="admin"
export METRICS_AUTH_PASSWORD="secure_password"
export METRICS_TLS_ENABLED="false"
```

## 🛠️ Troubleshooting

### CloudMCP not appearing in your AI tool?

1. Verify installation: `which cloud-mcp` (after go install)
2. Check MCP configuration files (`.mcp.json` or Claude Desktop config)
3. Restart your AI tool completely

### Health check not working?

1. Test the server directly:
   `echo '{"jsonrpc": "2.0", "method": "tools/list", "params": {}, "id": 1}' | cloud-mcp`
2. Check logs in config directory
3. Verify environment variables if using custom configuration

### Metrics server issues?

1. Check if port 8080 is available: `lsof -i :8080`
2. Test metrics endpoint: `curl http://localhost:8080/health`
3. Check authentication if configured

## 📚 Available Tools

CloudMCP currently provides **minimal shell** functionality:

### Health Monitoring

- `health_check` - Server health status and service discovery
  - Reports server status and uptime
  - Lists available tools and capabilities  
  - Provides service discovery information
  - Sub-microsecond response times

### Future Extensions

The framework is ready for:

- **Cloud Provider Tools** - AWS, Azure, GCP, Linode integrations
- **Infrastructure Tools** - Terraform, Kubernetes, Docker management
- **Monitoring Tools** - Observability and alerting capabilities
- **Security Tools** - Compliance and vulnerability scanning

## 🔧 Development

### Project Structure

```text
CloudMCP/
├── cmd/
│   └── cloud-mcp/           # Main MCP server entry point
├── internal/
│   ├── server/              # MCP server implementation with metrics
│   ├── tools/               # Health check tool implementation
│   ├── config/              # Configuration management
│   ├── registry/            # Provider registry framework
│   └── providers/           # Provider implementations (ready for future)
├── pkg/
│   ├── logger/              # Flexible logging package
│   ├── metrics/             # Unified metrics with Prometheus backend
│   ├── types/               # Shared types
│   └── interfaces/          # Core interfaces for extensibility
```

### Building and Testing

```bash
# Run tests
go test ./...

# Format code
gofumpt -w .

# Run linters
golangci-lint run

# Build
go build -o bin/cloud-mcp cmd/server/main.go
```

### Testing Without AI Tools

**Using MCP Inspector:**

```bash
npm install -g @modelcontextprotocol/inspector
mcp-inspector ./bin/cloud-mcp
```

**Direct stdio testing:**

```bash
echo '{"jsonrpc": "2.0", "method": "tools/list", "params": {}, "id": 1}' | ./bin/cloud-mcp
```

## 📋 Advanced Configuration

<!-- markdownlint-disable MD033 -->
<details>
<summary>Click to expand detailed configuration options</summary>
<!-- markdownlint-enable MD033 -->

### Full TOML Configuration

```toml
[system]
server_name = "Cloud MCP Server"
log_level = "info"                # debug, info, warn, error
enable_metrics = true
metrics_port = 8080
default_account = "primary"

# Logging configuration
log_max_size = 10      # MB
log_max_backups = 5    # Number of files to keep
log_max_age = 30       # Days to retain logs

[account.primary]
token = "your_production_token_here"
label = "Production"

[account.development]
token = "your_dev_token_here"
label = "Development"
```

### Log File Locations

- **Linux**: `~/.local/share/CloudMCP/cloudmcp.log`
- **macOS**: `~/Library/Application Support/CloudMCP/logs/cloudmcp.log`
- **Windows**: `%APPDATA%\CloudMCP\logs\cloudmcp.log`

### Monitoring

When metrics are enabled, Prometheus metrics are available at `http://localhost:8080/metrics`:

- `cloudmcp_tool_requests_total` - Tool request metrics
- `cloudmcp_linode_api_requests_total` - API request metrics

</details>

## 🔒 Security

- API tokens are never logged or returned in responses
- All sensitive data is sanitized in error messages
- TOML-based configuration keeps secrets secure
- Rate limiting protects against API abuse
- Request validation prevents malformed inputs

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes with tests
4. Ensure code is formatted (`gofumpt -w .`)
5. Run linters (`golangci-lint run`)
6. Submit a pull request

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [MCP SDK for Go](https://github.com/mark3labs/mcp-go) - MCP protocol implementation
- [Linode API](https://www.linode.com/api/v4) - Cloud infrastructure API

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/chadit/CloudMCP/issues)
- **Discussions**: [GitHub Discussions](https://github.com/chadit/CloudMCP/discussions)
- **API Reference**: [LINODE_API_COVERAGE.md](LINODE_API_COVERAGE.md)
