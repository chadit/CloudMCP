# CloudMCP

Full disclosure: I am an employee of Linode/Akamai, however this project is being developed independently in my own free time and is not associated in any way with Linode, LLC or Akamai Technologies, Inc. Only publicly-available documentation and information is being used in the development of this project.

CloudMCP enables you to manage your Linode cloud infrastructure using **natural language** through AI assistants like Claude and GitHub Copilot.

**Instead of complex CLI commands, just say:**

- *"Create a 2GB Linode in Newark with Ubuntu 22.04"*
- *"List all my instances and their status"*
- *"Switch to my development account"*

## 🚀 Quick Start

### Prerequisites

You need one of these AI tools installed:

- [Claude Desktop](https://claude.ai/desktop) (recommended)
- [Claude Code](https://claude.ai/code)
- [VS Code with GitHub Copilot Chat](https://marketplace.visualstudio.com/items?itemName=GitHub.copilot-chat)

### Installation (2 minutes)

```bash
# Install CloudMCP with automated setup
go install github.com/chadit/CloudMCP/cmd/server@latest
go install github.com/chadit/CloudMCP/cmd/cloud-mcp-setup@latest

# Auto-configure your AI tools
cloud-mcp-setup
```

### Add Your Linode Token

Through your AI assistant, say:
*"Add a Linode account named 'primary' with token YOUR_LINODE_TOKEN and label 'Production'"*

### Test It Works

Ask your AI assistant:
*"List my Linode instances"*

**That's it!** You're now managing Linode through natural language.

## 💡 What You Can Do

CloudMCP provides complete Linode API coverage through natural language:

### Compute Management

- *"Create a 4GB Linode in Dallas with Ubuntu 22.04"*
- *"Shutdown my web-server instance"*
- *"Boot all instances in the us-east region"*

### Storage Operations

- *"List all my block storage volumes"*
- *"Create a 50GB volume and attach it to my database server"*
- *"Create an image backup from my production server"*

### Account Operations

- *"Switch to development account"*
- *"Show current account usage and billing"*
- *"List all configured accounts"*

### Multi-Account Support

- Manage multiple Linode accounts seamlessly
- Switch between accounts with simple commands
- Secure token management with validation

## 📦 Installation Options

### Option 1: Automated Setup (Recommended)

For most users who want everything configured automatically:

```bash
go install github.com/chadit/CloudMCP/cmd/server@latest
go install github.com/chadit/CloudMCP/cmd/cloud-mcp-setup@latest
cloud-mcp-setup
```

This will:

- Install CloudMCP server
- Automatically register with Claude Desktop, Claude Code, and VS Code
- Create configuration templates
- Provide setup guidance

### Option 2: Manual Setup (Advanced)

If you prefer manual control or automated setup doesn't work:

1. **Install server only:**

   ```bash
   go install github.com/chadit/CloudMCP/cmd/server@latest
   ```

2. **Configure your AI tool manually:**

   **Claude Desktop** - Edit `~/Library/Application Support/Claude/claude_desktop_config.json`:

   ```json
   {
     "mcpServers": {
       "cloud-mcp": {
         "command": "cloud-mcp",
         "args": []
       }
     }
   }
   ```

   **Claude Code:**

   ```bash
   claude mcp add -s user cloud-mcp cloud-mcp
   ```

   **VS Code** - Add to settings.json:

   ```json
   {
     "github.copilot.chat.mcpServers": {
       "cloud-mcp": {
         "command": "cloud-mcp",
         "args": []
       }
     }
   }
   ```

3. **Restart your AI tool**

### Option 3: From Source (Developers)

```bash
git clone https://github.com/chadit/CloudMCP.git
cd CloudMCP
go build -o bin/cloud-mcp cmd/server/main.go
go build -o bin/cloud-mcp-setup cmd/cloud-mcp-setup/main.go
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
server_name = "Cloud MCP Server"
log_level = "info"
default_account = "primary"

[account.primary]
token = "your_linode_token_here"
label = "Production"
```

### Managing Accounts

**Through your AI assistant:**

- *"Add account named 'staging' with token TOKEN and label 'Staging Environment'"*
- *"List all my configured accounts"*
- *"Switch to development account"*
- *"Remove the old-staging account"*

## 🛠️ Troubleshooting

### CloudMCP not appearing in your AI tool?

1. Verify installation: `which cloud-mcp`
2. Check that config files were updated by cloud-mcp-setup
3. Restart your AI tool completely

### "No accounts configured" error?

1. Add your token through your AI assistant: *"Add Linode account with token YOUR_TOKEN"*
2. Verify: *"List my Linode accounts"*

### Commands not working?

1. Test connection: *"Get current account information"*
2. Check your Linode token has proper permissions
3. Verify token in Linode Cloud Manager

## 📚 Available Commands

CloudMCP provides **100% coverage** of production-ready Linode API services:

### System Information

- `cloudmcp_version` - Get version and build information
- `cloudmcp_version_json` - Version info in JSON format

### Account Management

- `linode_account_get` - Current account information
- `linode_account_switch` - Switch between accounts
- `linode_account_list` - List configured accounts

### Compute Operations

- `linode_instances_list` - List all instances
- `linode_instance_create` - Create new instances
- `linode_instance_delete` - Delete instances
- `linode_instance_boot/shutdown/reboot` - Control instance power

### Storage Management

- `linode_volumes_*` - Manage block storage volumes
- `linode_images_*` - Manage custom images

### Networking

- `linode_ips_*` - Manage IP addresses

**[→ See complete API coverage](LINODE_API_COVERAGE.md)**

## 🔧 Development

### Project Structure

```text
CloudMCP/
├── cmd/
│   ├── server/          # Main MCP server
│   └── cloud-mcp-setup/ # Optional setup tool
├── internal/
│   ├── server/          # MCP server implementation
│   ├── services/linode/ # Linode API integration
│   └── config/          # Configuration management
└── pkg/                 # Shared packages
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
