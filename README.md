# CloudMCP

CloudMCP is a minimal Model Context Protocol (MCP) server designed as a
lightweight foundation for cloud infrastructure management through natural language commands.

**Current Status**: 🎉 **Minimal MCP Server** - Simple and Ready

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
# Ask your AI assistant: "Hello CloudMCP" or "What version are you?"
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
- *"Hello CloudMCP"*
- *"What version are you?"*

**That's it!** You now have a minimal MCP server foundation ready for
extension.

## 💡 Current Capabilities

CloudMCP is currently a **minimal MCP server** with two simple tools:

### Available Tools

#### `hello` - Friendly Greeting
Say hello to CloudMCP and get a friendly response.

**Examples:**
- *"Hello CloudMCP"* → "Hello, World! CloudMCP server is running and ready to help."
- *"Say hello to Alice"* → "Hello, Alice! CloudMCP server is running and ready to help."

#### `version` - Version Information  
Get detailed version and build information about the CloudMCP server.

**Example:**
*"What version are you?"* returns:
```json
{
  "version": "0.1.0",
  "api_version": "0.1.0", 
  "build_date": "unknown",
  "git_commit": "dev",
  "git_branch": "main",
  "go_version": "go1.24.2",
  "platform": "darwin/arm64",
  "features": {
    "tools": "hello,version",
    "logging": "basic", 
    "protocol": "mcp",
    "mode": "minimal"
  }
}
```

### Framework Ready

This minimal foundation is ready for extension with:

- **Cloud Provider Tools** - AWS, Linode, GCP, Azure integrations
- **Infrastructure Tools** - Terraform, Kubernetes, Docker management
- **Custom Tools** - Any functionality you need via MCP protocol

## 📦 Installation from Source

### From Source (Developers)

For development and customization:

```bash
git clone https://github.com/chadit/CloudMCP.git
cd CloudMCP
go build -o bin/cloud-mcp cmd/cloud-mcp/main.go

# Run locally for development
go run cmd/cloud-mcp/main.go

# Run tests (when available)
go test ./...
```

## ⚙️ Configuration

CloudMCP uses simple environment variable configuration:

```bash
# Optional customization
export CLOUD_MCP_SERVER_NAME="My CloudMCP Server"
export LOG_LEVEL="info"  # debug, info, warn, error
```

**Default values:**
- Server Name: "CloudMCP Minimal"
- Log Level: "info"

## 🛠️ Troubleshooting

### CloudMCP not appearing in your AI tool?

1. Verify installation: `which cloud-mcp` (after go install)
2. Check MCP configuration files (`.mcp.json` or Claude Desktop config)
3. Restart your AI tool completely

### Tools not working?

1. Test the server directly:
   ```bash
   echo '{"jsonrpc": "2.0", "method": "tools/list", "params": {}, "id": 1}' | cloud-mcp
   ```
2. Check server logs for error messages
3. Verify CloudMCP is properly configured in your MCP client

## 📚 Available Tools Reference

### hello Tool

**Purpose**: Responds with a friendly greeting message

**Parameters**: 
- `name` (optional string): Name to include in greeting

**Examples**:
```
User: "Hello CloudMCP"
Response: "Hello, World! CloudMCP server is running and ready to help."

User: "Say hello to John"  
Response: "Hello, John! CloudMCP server is running and ready to help."
```

### version Tool

**Purpose**: Returns detailed version and build information

**Parameters**: None

**Response**: JSON object containing:
- `version`: Semantic version
- `api_version`: MCP API version
- `build_date`: When the binary was built
- `git_commit`: Git commit hash
- `git_branch`: Git branch
- `go_version`: Go compiler version
- `platform`: Operating system and architecture
- `features`: Available feature set

## 🔧 Development

### Project Structure

```text
CloudMCP/
├── cmd/
│   └── cloud-mcp/           # Main MCP server entry point
├── internal/
│   ├── server/              # Minimal MCP server implementation
│   ├── tools/               # Hello and version tools
│   ├── config/              # Environment-based configuration
│   └── version/             # Version information
└── pkg/
    └── interfaces/          # Tool interface definitions
```

### Building and Testing

```bash
# Build
go build -o bin/cloud-mcp cmd/cloud-mcp/main.go

# Format code
gofumpt -w .

# Run linters
golangci-lint run
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

## 🏗️ Architecture

CloudMCP follows a **minimal MCP server** architecture:

- **Pure MCP Protocol**: Communicates via stdin/stdout only
- **No HTTP Server**: No web endpoints, metrics, or complex infrastructure  
- **Simple Tools**: Clean tool interface for easy extension
- **Environment Config**: Simple configuration via environment variables
- **Standard Logging**: Uses Go's standard log package

This design makes CloudMCP:
- **Lightweight**: Minimal resource usage
- **Simple**: Easy to understand and extend
- **Focused**: Pure MCP protocol implementation
- **Extensible**: Ready for cloud provider tools

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

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/chadit/CloudMCP/issues)
- **Discussions**: [GitHub Discussions](https://github.com/chadit/CloudMCP/discussions)