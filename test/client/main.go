package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func main() {
	// Check for environment variables
	if os.Getenv("DEFAULT_LINODE_ACCOUNT") == "" {
		fmt.Println("Please set environment variables first:")
		fmt.Println("export DEFAULT_LINODE_ACCOUNT=primary")
		fmt.Println("export LINODE_ACCOUNTS_PRIMARY_TOKEN=your_token")
		fmt.Println("export LINODE_ACCOUNTS_PRIMARY_LABEL=\"Production\"")
		os.Exit(1)
	}

	fmt.Println("Starting CloudMCP Test Client...")
	fmt.Println("This will launch the cloud-mcp server and interact with it")
	fmt.Println()

	// Create MCP client
	// Get the binary path relative to where we're running from
	binaryPath := "./bin/cloud-mcp"
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		// Try from test/client directory
		binaryPath = "../../bin/cloud-mcp"
	}
	
	mcpClient, err := client.NewStdioMCPClient(
		binaryPath,
		[]string{},
	)
	if err != nil {
		log.Fatalf("Failed to create MCP client: %v", err)
	}

	ctx := context.Background()

	// Start the client
	fmt.Println("Starting CloudMCP server...")
	err = mcpClient.Start(ctx)
	if err != nil {
		log.Fatalf("Failed to start client: %v", err)
	}
	defer mcpClient.Close()

	// Initialize connection
	fmt.Println("Initializing connection...")
	initResult, err := mcpClient.Initialize(ctx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: "0.1.0",
			ClientInfo: mcp.Implementation{
				Name:    "test-client",
				Version: "1.0.0",
			},
			Capabilities: mcp.ClientCapabilities{},
		},
	})
	if err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}

	fmt.Println("Connected successfully!")
	fmt.Printf("Server: %s (version %s)\n", initResult.ServerInfo.Name, initResult.ServerInfo.Version)
	fmt.Println()

	// List available tools
	toolsResult, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		log.Fatalf("Failed to list tools: %v", err)
	}

	fmt.Println("Available tools:")
	for _, tool := range toolsResult.Tools {
		fmt.Printf("  - %s: %s\n", tool.Name, tool.Description)
	}
	fmt.Println()

	// Interactive loop
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("\nAvailable commands:")
		fmt.Println("  1. Get current account")
		fmt.Println("  2. List all accounts")
		fmt.Println("  3. Switch account")
		fmt.Println("  4. List instances")
		fmt.Println("  5. Raw tool call (advanced)")
		fmt.Println("  q. Quit")
		fmt.Print("\nEnter command: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "1":
			callTool(ctx, mcpClient, "linode_account_get", nil)
		case "2":
			callTool(ctx, mcpClient, "linode_account_list", nil)
		case "3":
			fmt.Print("Enter account name to switch to: ")
			accountName, _ := reader.ReadString('\n')
			accountName = strings.TrimSpace(accountName)
			callTool(ctx, mcpClient, "linode_account_switch", map[string]interface{}{
				"account_name": accountName,
			})
		case "4":
			callTool(ctx, mcpClient, "linode_instances_list", nil)
		case "5":
			fmt.Print("Enter tool name: ")
			toolName, _ := reader.ReadString('\n')
			toolName = strings.TrimSpace(toolName)
			
			fmt.Print("Enter arguments as JSON (or press enter for none): ")
			argsStr, _ := reader.ReadString('\n')
			argsStr = strings.TrimSpace(argsStr)
			
			var args map[string]interface{}
			if argsStr != "" {
				if err := json.Unmarshal([]byte(argsStr), &args); err != nil {
					fmt.Printf("Invalid JSON: %v\n", err)
					continue
				}
			}
			
			callTool(ctx, mcpClient, toolName, args)
		case "q", "quit", "exit":
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Println("Invalid command")
		}
	}
}

func callTool(ctx context.Context, mcpClient *client.Client, toolName string, args map[string]interface{}) {
	fmt.Printf("\nCalling tool: %s\n", toolName)
	
	result, err := mcpClient.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      toolName,
			Arguments: args,
		},
	})
	if err != nil {
		fmt.Printf("Error calling tool: %v\n", err)
		return
	}

	fmt.Println("Result:")
	// Since Content is an interface, we need to handle it as raw JSON
	if len(result.Content) > 0 {
		// Try to get the text representation
		for i, content := range result.Content {
			// Marshal to JSON to see the content
			data, _ := json.Marshal(content)
			var contentMap map[string]interface{}
			if err := json.Unmarshal(data, &contentMap); err == nil {
				if text, ok := contentMap["text"].(string); ok {
					fmt.Println(text)
				} else {
					fmt.Printf("Content %d: %s\n", i, string(data))
				}
			}
		}
	}
	
	if result.IsError {
		fmt.Println("(Tool returned an error)")
	}
}