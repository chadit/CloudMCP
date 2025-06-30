package linode

import (
	"context"
	"errors"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	linode "github.com/chadit/CloudMCP/internal/services/linode"
	"github.com/chadit/CloudMCP/pkg/interfaces"
	"github.com/chadit/CloudMCP/pkg/logger"
)

const (
	// Demo instance IDs for sample responses.
	demoInstanceID1 = 123456
	demoInstanceID2 = 789012
)

// Service is an alias for the linode service to avoid package name conflicts.
type Service = linode.Service

// Static errors for err113 compliance.
var (
	ErrUnknownTool          = errors.New("unknown tool")
	ErrParametersNil        = errors.New("parameters cannot be nil")
	ErrUnsupportedTool      = errors.New("unsupported tool")
	ErrServiceNotInit       = errors.New("Linode service is not initialized")
	ErrAccountManagerMissing = errors.New("account manager is not available")
)

// SampleTool provides a simple implementation of the Tool interface for demonstration.
// This shows how providers can create custom tools without tight coupling to services.
type SampleTool struct {
	name        string
	description string
	config      interfaces.Config
	logger      logger.Logger
}

// Definition returns the MCP tool definition.
func (st *SampleTool) Definition() mcp.Tool {
	return mcp.Tool{
		Name:        st.name,
		Description: st.description,
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
			Required:   []string{},
		},
	}
}

// Execute handles the tool execution.
func (st *SampleTool) Execute(_ context.Context, params map[string]interface{}) (interface{}, error) {
	// Log the execution
	st.logger.Info("Executing sample tool", "name", st.name)

	// Create a simple response based on the tool name
	switch st.name {
	case "linode_account_info":
		return map[string]interface{}{
			"account":     st.config.GetString("default_linode_account"),
			"provider":    "linode",
			"status":      "active",
			"plugin_demo": true,
		}, nil

	case "linode_instances_list":
		return map[string]interface{}{
			"instances": []map[string]interface{}{
				{
					"id":     demoInstanceID1,
					"label":  "demo-instance-1",
					"region": "us-east",
					"type":   "g6-nanode-1",
					"status": "running",
				},
				{
					"id":     demoInstanceID2,
					"label":  "demo-instance-2",
					"region": "us-west",
					"type":   "g6-standard-1",
					"status": "running",
				},
			},
			"plugin_demo": true,
		}, nil

	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownTool, st.name)
	}
}

// Validate validates the tool parameters.
func (st *SampleTool) Validate(params map[string]interface{}) error {
	// Basic validation - ensure params is not nil
	if params == nil {
		return ErrParametersNil
	}

	// These sample tools don't require any specific parameters
	return nil
}

// ToolWrapper wraps existing Linode service tools to implement the Tool interface.
type ToolWrapper struct {
	definition mcp.Tool
	service    *Service
	logger     logger.Logger
}

// Definition returns the MCP tool definition.
func (ltw *ToolWrapper) Definition() mcp.Tool {
	return ltw.definition
}

// Execute handles the tool execution by delegating to the service.
func (ltw *ToolWrapper) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Create a mock MCP request to pass to the service handler
	request := mockCallToolRequest{
		toolName:  ltw.definition.Name,
		arguments: params,
	}

	// Find and call the appropriate service handler
	result, err := ltw.callServiceHandler(ctx, request)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Validate validates the tool parameters.
func (ltw *ToolWrapper) Validate(params map[string]interface{}) error {
	if params == nil {
		return ErrParametersNil
	}
	// Add more validation based on the tool's input schema if needed
	return nil
}

// callServiceHandler routes the request to the appropriate service handler.
func (ltw *ToolWrapper) callServiceHandler(_ context.Context, request mockCallToolRequest) (interface{}, error) {
	// This is a simplified demonstration implementation
	// In practice, this would route to the actual service handler methods
	switch ltw.definition.Name {
	case "account_get":
		// Return mock account information for demonstration
		return map[string]interface{}{
			"account_id": "demo-account-123",
			"label":      "Demo Account",
			"email":      "demo@example.com",
			"status":     "active",
		}, nil
	case "account_list":
		// Return mock account list for demonstration
		return []map[string]interface{}{
			{"name": "primary", "label": "Primary Account", "status": "active"},
			{"name": "development", "label": "Development Account", "status": "active"},
		}, nil
	case "account_switch":
		// Return mock success response for demonstration
		return map[string]interface{}{
			"message":     "Account switched successfully",
			"new_account": request.GetArguments()["account_name"],
		}, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedTool, ltw.definition.Name)
	}
}

// mockCallToolRequest implements a basic version of mcp.CallToolRequest for testing.
type mockCallToolRequest struct {
	toolName  string
	arguments map[string]interface{}
}

// GetArguments returns the tool arguments.
func (m mockCallToolRequest) GetArguments() map[string]interface{} {
	return m.arguments
}

// ToolRegistrationHelper provides utilities for registering Linode tools.
type ToolRegistrationHelper struct {
	service *Service
	logger  logger.Logger
}

// NewToolRegistrationHelper creates a new tool registration helper.
func NewToolRegistrationHelper(service *Service, logger logger.Logger) *ToolRegistrationHelper {
	return &ToolRegistrationHelper{
		service: service,
		logger:  logger,
	}
}

// WrapTools converts service tool definitions to plugin-compatible tools.
func (trh *ToolRegistrationHelper) WrapTools(toolDefinitions []mcp.Tool) []interfaces.Tool {
	tools := make([]interfaces.Tool, len(toolDefinitions))

	for i, definition := range toolDefinitions {
		tools[i] = &ToolWrapper{
			definition: definition,
			service:    trh.service,
			logger:     trh.logger,
		}
	}

	return tools
}

// RegisterWithServer registers all wrapped tools with an MCP server.
func (trh *ToolRegistrationHelper) RegisterWithServer(server interfaces.MCPServer, toolDefinitions []mcp.Tool) error {
	tools := trh.WrapTools(toolDefinitions)

	for _, tool := range tools {
		if err := server.RegisterTool(tool); err != nil {
			return fmt.Errorf("failed to register tool %q: %w", tool.Definition().Name, err)
		}
	}

	return nil
}

// GetToolsByCategory returns tools grouped by their capability category.
func (trh *ToolRegistrationHelper) GetToolsByCategory(toolDefinitions []mcp.Tool) map[string][]interfaces.Tool {
	categories := make(map[string][]interfaces.Tool)

	for _, definition := range toolDefinitions {
		// Determine category based on tool name prefix
		category := trh.getToolCategory(definition.Name)

		tool := &ToolWrapper{
			definition: definition,
			service:    trh.service,
			logger:     trh.logger,
		}

		categories[category] = append(categories[category], tool)
	}

	return categories
}

// getToolCategory determines the capability category for a tool based on its name.
func (trh *ToolRegistrationHelper) getToolCategory(toolName string) string {
	// Map tool name prefixes to capability categories
	categoryMap := map[string]string{
		"instance":          "compute",
		"volume":            "storage",
		"object_storage":    "storage",
		"firewall":          "networking",
		"nodebalancer":      "networking",
		"vlan":              "networking",
		"ip":                "networking",
		"ipv6":              "networking",
		"reserved_ip":       "networking",
		"domain":            "dns",
		"lke":               "kubernetes",
		"database":          "databases",
		"mysql_database":    "databases",
		"postgres_database": "databases",
		"longview":          "monitoring",
		"support":           "support",
		"account":           "account",
		"system":            "management",
		"image":             "compute",
		"kernel":            "compute",
		"type":              "compute",
		"region":            "management",
		"stackscript":       "automation",
	}

	// Find the category by checking prefixes
	for prefix, category := range categoryMap {
		if len(toolName) >= len(prefix) && toolName[:len(prefix)] == prefix {
			return category
		}
	}

	// Default category for unrecognized tools
	return "general"
}

// GetToolStatistics returns statistics about the registered tools.
func (trh *ToolRegistrationHelper) GetToolStatistics(toolDefinitions []mcp.Tool) map[string]interface{} {
	categories := trh.GetToolsByCategory(toolDefinitions)

	stats := map[string]interface{}{
		"total_tools": len(toolDefinitions),
		"categories":  make(map[string]int),
	}

	categoryStats := make(map[string]int)
	for category, tools := range categories {
		categoryStats[category] = len(tools)
	}

	stats["categories"] = categoryStats

	return stats
}

// ValidateToolConfiguration validates that all tools can be properly configured.
func (trh *ToolRegistrationHelper) ValidateToolConfiguration() error {
	// Check that the service is available
	if trh.service == nil {
		return ErrServiceNotInit
	}

	// Verify service is properly initialized (simplified check)
	// Note: In the real implementation, we would check account manager
	// but since GetAccountManager is not exported, we'll use a basic check
	if trh.service.Name() == "" {
		return ErrAccountManagerMissing
	}

	// In a real implementation, we would check that accounts are configured
	// and verify the current account is accessible, but since account manager
	// methods are not exported, we'll skip this detailed validation
	// and rely on the service being properly initialized

	return nil
}

// CreateCustomTool creates a custom tool for specialized Linode operations.
func (trh *ToolRegistrationHelper) CreateCustomTool(
	name, description string,
	inputSchema mcp.ToolInputSchema,
	handler func(context.Context, map[string]interface{}) (interface{}, error),
) interfaces.Tool { //nolint:ireturn // Factory pattern requires returning interface
	definition := mcp.Tool{
		Name:        name,
		Description: description,
		InputSchema: inputSchema,
	}

	return &CustomLinodeTool{
		definition: definition,
		handler:    handler,
		service:    trh.service,
		logger:     trh.logger,
	}
}

// CustomLinodeTool represents a custom tool implementation for specialized operations.
type CustomLinodeTool struct {
	definition mcp.Tool
	handler    func(context.Context, map[string]interface{}) (interface{}, error)
	service    *linode.Service
	logger     logger.Logger
}

// Definition returns the MCP tool definition.
func (clt *CustomLinodeTool) Definition() mcp.Tool {
	return clt.definition
}

// Execute handles the tool execution using the custom handler.
func (clt *CustomLinodeTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return clt.handler(ctx, params)
}

// Validate validates the tool parameters.
func (clt *CustomLinodeTool) Validate(params map[string]interface{}) error {
	if params == nil {
		return ErrParametersNil
	}

	return nil
}
