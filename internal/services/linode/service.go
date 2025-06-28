package linode

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/linode/linodego"
	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"golang.org/x/oauth2"

	"github.com/chadit/CloudMCP/internal/config"
	"github.com/chadit/CloudMCP/pkg/logger"
	"github.com/chadit/CloudMCP/pkg/types"
)

var (
	ErrAccountNoToken       = errors.New("account has no token configured")
	ErrInvalidServerType    = errors.New("invalid server type")
	ErrInvalidParameterType = errors.New("parameter is required and must be a positive number")
)

// Service provides Linode cloud infrastructure management through MCP tools.
// It manages multiple Linode accounts and provides thread-safe access to
// instance, volume, image, and network operations.
type Service struct {
	config         *config.Config
	logger         logger.Logger
	accountManager *AccountManager
}

// AccountManager provides thread-safe management of multiple Linode accounts.
// It maintains a collection of configured accounts and handles account switching
// operations for the service.
type AccountManager struct {
	accounts       map[string]*Account
	currentAccount string
	mu             sync.RWMutex
}

// Account represents a configured Linode account with its associated API client.
// Each account has a name for identification, a human-readable label, and an
// authenticated Linode API client for operations.
type Account struct {
	Name   string
	Label  string
	Client *linodego.Client
}

// New creates a new Linode service instance with the provided configuration and logger.
// It initializes all configured Linode accounts with their API clients and sets up
// the account manager with the default account. Returns an error if account configuration
// is invalid or if account initialization fails.
func New(cfg *config.Config, log logger.Logger) (*Service, error) {
	service := &Service{
		config: cfg,
		logger: log,
		accountManager: &AccountManager{
			accounts:       make(map[string]*Account),
			currentAccount: cfg.DefaultLinodeAccount,
		},
	}

	for name, accCfg := range cfg.LinodeAccounts {
		if accCfg.Token == "" {
			return nil, fmt.Errorf("account %q: %w", name, ErrAccountNoToken)
		}

		tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accCfg.Token})
		oauth2Client := oauth2.NewClient(context.Background(), tokenSource)

		client := linodego.NewClient(oauth2Client)
		client.SetUserAgent("CloudMCP/0.1.0")

		// Set custom API URL if provided (e.g., for development environments)
		if accCfg.APIURL != "" {
			client.SetBaseURL(accCfg.APIURL)
		}

		service.accountManager.accounts[name] = &Account{
			Name:   name,
			Label:  accCfg.Label,
			Client: &client,
		}
	}

	return service, nil
}

// Name returns the service identifier for the Linode cloud service.
func (s *Service) Name() string {
	return "linode"
}

// Initialize verifies the default Linode account connection and prepares the service
// for operation. It validates the default account by fetching the user profile and
// logs the initialization details.
func (s *Service) Initialize(ctx context.Context) error {
	s.logger.Info("Initializing Linode service",
		"accounts", len(s.accountManager.accounts),
		"default", s.config.DefaultLinodeAccount,
	)

	account, err := s.accountManager.GetAccount(s.config.DefaultLinodeAccount)
	if err != nil {
		return err
	}

	profile, err := account.Client.GetProfile(ctx)
	if err != nil {
		return fmt.Errorf("failed to verify default account: %w", err)
	}

	s.logger.Info("Linode service initialized",
		"account", s.config.DefaultLinodeAccount,
		"username", profile.Username,
	)

	return nil
}

// Shutdown performs cleanup operations for the Linode service. Currently this is
// a no-op as the service doesn't maintain persistent connections that need cleanup.
func (s *Service) Shutdown(_ context.Context) error {
	s.logger.Info("Shutting down Linode service")

	return nil
}

// RegisterTools registers all Linode MCP tools with the provided MCP server.
// This includes account management, instance operations, volume management,
// IP address tools, and image operations. Returns an error if tool registration fails.
func (s *Service) RegisterTools(server interface{}) error {
	mcpServer, ok := server.(*mcpserver.MCPServer)
	if !ok {
		return fmt.Errorf("%w", ErrInvalidServerType)
	}

	// Register tools with proper JSON schemas
	// linode_account_get - no parameters
	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_account_get",
		Description: "Get current Linode account information",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
			Required:   []string{},
		},
	}, s.handleAccountGet)

	// linode_account_list - no parameters
	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_account_list",
		Description: "List all configured Linode accounts",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
			Required:   []string{},
		},
	}, s.handleAccountList)

	// linode_account_switch - requires account_name parameter
	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_account_switch",
		Description: "Switch to a different Linode account",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"account_name": map[string]any{
					"type":        "string",
					"description": "Name of the account to switch to",
				},
			},
			Required: []string{"account_name"},
		},
	}, s.handleAccountSwitch)

	// System tools
	mcpServer.AddTool(mcp.Tool{
		Name:        "cloudmcp_version",
		Description: "Get CloudMCP version and build information",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
			Required:   []string{},
		},
	}, s.handleSystemVersion)

	mcpServer.AddTool(mcp.Tool{
		Name:        "cloudmcp_version_json",
		Description: "Get CloudMCP version information in JSON format",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
			Required:   []string{},
		},
	}, s.handleSystemVersionJSON)

	// Configuration management tools
	mcpServer.AddTool(mcp.Tool{
		Name:        "cloudmcp_account_list",
		Description: "List all configured Linode accounts from TOML configuration",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
			Required:   []string{},
		},
	}, s.handleCloudMCPAccountList)

	mcpServer.AddTool(mcp.Tool{
		Name:        "cloudmcp_account_add",
		Description: "Add a new Linode account to TOML configuration",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "Name identifier for the account",
				},
				"token": map[string]any{
					"type":        "string",
					"description": "Linode API token",
				},
				"label": map[string]any{
					"type":        "string",
					"description": "Human-readable label for the account",
				},
				"apiurl": map[string]any{
					"type":        "string",
					"description": "Optional custom API URL (defaults to official Linode API)",
				},
			},
			Required: []string{"name", "token", "label"},
		},
	}, s.handleCloudMCPAccountAdd)

	mcpServer.AddTool(mcp.Tool{
		Name:        "cloudmcp_account_remove",
		Description: "Remove a Linode account from TOML configuration",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "Name of the account to remove",
				},
			},
			Required: []string{"name"},
		},
	}, s.handleCloudMCPAccountRemove)

	mcpServer.AddTool(mcp.Tool{
		Name:        "cloudmcp_account_update",
		Description: "Update an existing Linode account in TOML configuration",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "Name of the account to update",
				},
				"token": map[string]any{
					"type":        "string",
					"description": "New Linode API token (optional)",
				},
				"label": map[string]any{
					"type":        "string",
					"description": "New human-readable label (optional)",
				},
				"apiurl": map[string]any{
					"type":        "string",
					"description": "New custom API URL (optional)",
				},
			},
			Required: []string{"name"},
		},
	}, s.handleCloudMCPAccountUpdate)

	// linode_instances_list - no parameters
	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_instances_list",
		Description: "List all Linode instances",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
			Required:   []string{},
		},
	}, s.handleInstancesList)

	// linode_instance_get - requires instance_id parameter
	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_instance_get",
		Description: "Get details of a specific Linode instance",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"instance_id": map[string]any{
					"type":        "number",
					"description": "ID of the Linode instance",
				},
			},
			Required: []string{"instance_id"},
		},
	}, s.handleInstanceGet)

	// linode_volumes_list - no parameters
	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_volumes_list",
		Description: "List all block storage volumes",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
			Required:   []string{},
		},
	}, s.handleVolumesList)

	// linode_volume_get - requires volume_id parameter
	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_volume_get",
		Description: "Get details of a specific volume",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"volume_id": map[string]any{
					"type":        "number",
					"description": "ID of the volume",
				},
			},
			Required: []string{"volume_id"},
		},
	}, s.handleVolumeGet)

	// linode_ips_list - no parameters
	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_ips_list",
		Description: "List all IP addresses",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
			Required:   []string{},
		},
	}, s.handleIPsList)

	// linode_ip_get - requires address parameter
	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_ip_get",
		Description: "Get details of a specific IP address",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"address": map[string]any{
					"type":        "string",
					"description": "IP address to get details for",
				},
			},
			Required: []string{"address"},
		},
	}, s.handleIPGet)

	// Instance operation tools
	// linode_instance_create
	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_instance_create",
		Description: "Create a new Linode instance",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"region": map[string]any{
					"type":        "string",
					"description": "Region ID where the instance will be created",
				},
				"type": map[string]any{
					"type":        "string",
					"description": "Linode type ID (e.g. g6-nanode-1)",
				},
				"label": map[string]any{
					"type":        "string",
					"description": "Display label for the instance",
				},
				"image": map[string]any{
					"type":        "string",
					"description": "Image ID to deploy (e.g. linode/ubuntu22.04)",
				},
				"root_pass": map[string]any{
					"type":        "string",
					"description": "Root password for the instance",
				},
				"authorized_keys": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "SSH public keys to add to root user",
				},
				"stackscript_id": map[string]any{
					"type":        "number",
					"description": "StackScript ID to run on first boot",
				},
				"backups_enabled": map[string]any{
					"type":        "boolean",
					"description": "Enable automatic backups",
				},
				"private_ip": map[string]any{
					"type":        "boolean",
					"description": "Add a private IP address",
				},
				"tags": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "Tags to apply to the instance",
				},
			},
			Required: []string{"region", "type", "label"},
		},
	}, s.handleInstanceCreate)

	// linode_instance_delete
	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_instance_delete",
		Description: "Delete a Linode instance",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"instance_id": map[string]any{
					"type":        "number",
					"description": "ID of the Linode instance to delete",
				},
			},
			Required: []string{"instance_id"},
		},
	}, s.handleInstanceDelete)

	// linode_instance_boot
	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_instance_boot",
		Description: "Boot a Linode instance",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"instance_id": map[string]any{
					"type":        "number",
					"description": "ID of the Linode instance to boot",
				},
				"config_id": map[string]any{
					"type":        "number",
					"description": "Configuration profile ID to boot",
				},
			},
			Required: []string{"instance_id"},
		},
	}, s.handleInstanceBoot)

	// linode_instance_shutdown
	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_instance_shutdown",
		Description: "Shutdown a Linode instance",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"instance_id": map[string]any{
					"type":        "number",
					"description": "ID of the Linode instance to shutdown",
				},
			},
			Required: []string{"instance_id"},
		},
	}, s.handleInstanceShutdown)

	// linode_instance_reboot
	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_instance_reboot",
		Description: "Reboot a Linode instance",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"instance_id": map[string]any{
					"type":        "number",
					"description": "ID of the Linode instance to reboot",
				},
				"config_id": map[string]any{
					"type":        "number",
					"description": "Configuration profile ID to reboot into",
				},
			},
			Required: []string{"instance_id"},
		},
	}, s.handleInstanceReboot)

	// Volume operation tools
	// linode_volume_create
	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_volume_create",
		Description: "Create a new volume",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"label": map[string]any{
					"type":        "string",
					"description": "Display label for the volume",
				},
				"size": map[string]any{
					"type":        "number",
					"description": "Size of the volume in GB (10-8192)",
				},
				"region": map[string]any{
					"type":        "string",
					"description": "Region ID where the volume will be created",
				},
				"linode_id": map[string]any{
					"type":        "number",
					"description": "ID of Linode to attach the volume to",
				},
				"tags": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "Tags to apply to the volume",
				},
			},
			Required: []string{"label", "size"},
		},
	}, s.handleVolumeCreate)

	// linode_volume_delete
	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_volume_delete",
		Description: "Delete a volume",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"volume_id": map[string]any{
					"type":        "number",
					"description": "ID of the volume to delete",
				},
			},
			Required: []string{"volume_id"},
		},
	}, s.handleVolumeDelete)

	// linode_volume_attach
	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_volume_attach",
		Description: "Attach volume to instance",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"volume_id": map[string]any{
					"type":        "number",
					"description": "ID of the volume to attach",
				},
				"linode_id": map[string]any{
					"type":        "number",
					"description": "ID of the Linode to attach to",
				},
				"persist_across_boots": map[string]any{
					"type":        "boolean",
					"description": "Keep volume attached when Linode reboots",
				},
			},
			Required: []string{"volume_id", "linode_id"},
		},
	}, s.handleVolumeAttach)

	// linode_volume_detach
	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_volume_detach",
		Description: "Detach volume from instance",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"volume_id": map[string]any{
					"type":        "number",
					"description": "ID of the volume to detach",
				},
			},
			Required: []string{"volume_id"},
		},
	}, s.handleVolumeDetach)

	// Image tools
	// linode_images_list
	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_images_list",
		Description: "List all available Linode images",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"is_public": map[string]any{
					"type":        "boolean",
					"description": "Filter to only public (true) or private (false) images",
				},
			},
		},
	}, s.handleImagesList)

	// linode_image_get
	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_image_get",
		Description: "Get details of a specific image",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"image_id": map[string]any{
					"type":        "string",
					"description": "ID of the image (e.g. linode/ubuntu22.04 or private/12345)",
				},
			},
			Required: []string{"image_id"},
		},
	}, s.handleImageGet)

	// linode_image_create
	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_image_create",
		Description: "Create a custom image from a Linode disk",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"disk_id": map[string]any{
					"type":        "number",
					"description": "ID of the Linode disk to create image from",
				},
				"label": map[string]any{
					"type":        "string",
					"description": "Display label for the image",
				},
				"description": map[string]any{
					"type":        "string",
					"description": "Detailed description of the image",
				},
				"cloud_init": map[string]any{
					"type":        "boolean",
					"description": "Whether this image supports cloud-init",
				},
				"tags": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "Tags to apply to the image",
				},
			},
			Required: []string{"disk_id", "label"},
		},
	}, s.handleImageCreate)

	// linode_image_update
	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_image_update",
		Description: "Update an existing custom image",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"image_id": map[string]any{
					"type":        "string",
					"description": "ID of the image to update",
				},
				"label": map[string]any{
					"type":        "string",
					"description": "New display label for the image",
				},
				"description": map[string]any{
					"type":        "string",
					"description": "New description for the image",
				},
				"tags": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "New tags for the image (replaces existing tags)",
				},
			},
			Required: []string{"image_id"},
		},
	}, s.handleImageUpdate)

	// linode_image_delete
	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_image_delete",
		Description: "Delete a custom image",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"image_id": map[string]any{
					"type":        "string",
					"description": "ID of the image to delete",
				},
			},
			Required: []string{"image_id"},
		},
	}, s.handleImageDelete)

	// linode_image_replicate
	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_image_replicate",
		Description: "Replicate a custom image to additional regions",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"image_id": map[string]any{
					"type":        "string",
					"description": "ID of the image to replicate",
				},
				"regions": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "List of region IDs to replicate the image to",
				},
			},
			Required: []string{"image_id", "regions"},
		},
	}, s.handleImageReplicate)

	// linode_image_upload_create
	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_image_upload_create",
		Description: "Create an image upload URL for uploading a custom image file",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"label": map[string]any{
					"type":        "string",
					"description": "Display label for the uploaded image",
				},
				"region": map[string]any{
					"type":        "string",
					"description": "Initial region for the uploaded image",
				},
				"description": map[string]any{
					"type":        "string",
					"description": "Description of the uploaded image",
				},
				"cloud_init": map[string]any{
					"type":        "boolean",
					"description": "Whether this image supports cloud-init",
				},
				"tags": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "Tags to apply to the image",
				},
			},
			Required: []string{"label", "region"},
		},
	}, s.handleImageUploadCreate)

	// Firewall tools
	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_firewalls_list",
		Description: "List all firewalls",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
			Required:   []string{},
		},
	}, s.handleFirewallsList)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_firewall_get",
		Description: "Get details of a specific firewall",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"firewall_id": map[string]any{
					"type":        "number",
					"description": "ID of the firewall",
				},
			},
			Required: []string{"firewall_id"},
		},
	}, s.handleFirewallGet)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_firewall_create",
		Description: "Create a new firewall",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"label": map[string]any{
					"type":        "string",
					"description": "Display label for the firewall",
				},
				"rules": map[string]any{
					"type":        "object",
					"description": "Firewall rules configuration",
				},
				"tags": map[string]any{
					"type":        "array",
					"description": "Tags to apply to the firewall",
				},
			},
			Required: []string{"label"},
		},
	}, s.handleFirewallCreate)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_firewall_update",
		Description: "Update an existing firewall",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"firewall_id": map[string]any{
					"type":        "number",
					"description": "ID of the firewall to update",
				},
				"label": map[string]any{
					"type":        "string",
					"description": "New display label for the firewall",
				},
				"tags": map[string]any{
					"type":        "array",
					"description": "New tags for the firewall",
				},
			},
			Required: []string{"firewall_id"},
		},
	}, s.handleFirewallUpdate)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_firewall_delete",
		Description: "Delete a firewall",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"firewall_id": map[string]any{
					"type":        "number",
					"description": "ID of the firewall to delete",
				},
			},
			Required: []string{"firewall_id"},
		},
	}, s.handleFirewallDelete)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_firewall_rules_update",
		Description: "Update firewall rules",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"firewall_id": map[string]any{
					"type":        "number",
					"description": "ID of the firewall to update rules for",
				},
				"rules": map[string]any{
					"type":        "object",
					"description": "New firewall rules configuration",
				},
			},
			Required: []string{"firewall_id", "rules"},
		},
	}, s.handleFirewallRulesUpdate)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_firewall_device_create",
		Description: "Assign a device to a firewall",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"firewall_id": map[string]any{
					"type":        "number",
					"description": "ID of the firewall",
				},
				"device_id": map[string]any{
					"type":        "number",
					"description": "ID of the device to assign to firewall",
				},
				"device_type": map[string]any{
					"type":        "string",
					"description": "Type of device (linode or nodebalancer)",
				},
			},
			Required: []string{"firewall_id", "device_id", "device_type"},
		},
	}, s.handleFirewallDeviceCreate)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_firewall_device_delete",
		Description: "Remove a device from a firewall",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"firewall_id": map[string]any{
					"type":        "number",
					"description": "ID of the firewall",
				},
				"device_id": map[string]any{
					"type":        "number",
					"description": "ID of the device to remove from firewall",
				},
			},
			Required: []string{"firewall_id", "device_id"},
		},
	}, s.handleFirewallDeviceDelete)

	// NodeBalancer tools
	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_nodebalancers_list",
		Description: "List all NodeBalancers",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
			Required:   []string{},
		},
	}, s.handleNodeBalancersList)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_nodebalancer_get",
		Description: "Get details of a specific NodeBalancer",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"nodebalancer_id": map[string]any{
					"type":        "number",
					"description": "ID of the NodeBalancer",
				},
			},
			Required: []string{"nodebalancer_id"},
		},
	}, s.handleNodeBalancerGet)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_nodebalancer_create",
		Description: "Create a new NodeBalancer",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"label": map[string]any{
					"type":        "string",
					"description": "Display label for the NodeBalancer",
				},
				"region": map[string]any{
					"type":        "string",
					"description": "Region ID where the NodeBalancer will be created",
				},
				"client_conn_throttle": map[string]any{
					"type":        "number",
					"description": "Throttle connections per second (0-20)",
				},
				"tags": map[string]any{
					"type":        "array",
					"description": "Tags to apply to the NodeBalancer",
				},
			},
			Required: []string{"label", "region"},
		},
	}, s.handleNodeBalancerCreate)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_nodebalancer_update",
		Description: "Update an existing NodeBalancer",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"nodebalancer_id": map[string]any{
					"type":        "number",
					"description": "ID of the NodeBalancer to update",
				},
				"label": map[string]any{
					"type":        "string",
					"description": "New display label for the NodeBalancer",
				},
				"client_conn_throttle": map[string]any{
					"type":        "number",
					"description": "New throttle connections per second (0-20)",
				},
				"tags": map[string]any{
					"type":        "array",
					"description": "New tags for the NodeBalancer",
				},
			},
			Required: []string{"nodebalancer_id"},
		},
	}, s.handleNodeBalancerUpdate)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_nodebalancer_delete",
		Description: "Delete a NodeBalancer",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"nodebalancer_id": map[string]any{
					"type":        "number",
					"description": "ID of the NodeBalancer to delete",
				},
			},
			Required: []string{"nodebalancer_id"},
		},
	}, s.handleNodeBalancerDelete)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_nodebalancer_config_create",
		Description: "Create a new NodeBalancer configuration",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"nodebalancer_id": map[string]any{
					"type":        "number",
					"description": "ID of the NodeBalancer",
				},
				"port": map[string]any{
					"type":        "number",
					"description": "Port to configure (1-65534)",
				},
				"protocol": map[string]any{
					"type":        "string",
					"description": "Protocol (http, https, tcp)",
				},
				"algorithm": map[string]any{
					"type":        "string",
					"description": "Balancing algorithm (roundrobin, leastconn, source)",
				},
				"stickiness": map[string]any{
					"type":        "string",
					"description": "Session stickiness (none, table, http_cookie)",
				},
				"check": map[string]any{
					"type":        "string",
					"description": "Health check type (none, connection, http, http_body)",
				},
				"check_interval": map[string]any{
					"type":        "number",
					"description": "Health check interval in seconds",
				},
				"check_timeout": map[string]any{
					"type":        "number",
					"description": "Health check timeout in seconds",
				},
				"check_attempts": map[string]any{
					"type":        "number",
					"description": "Health check attempts before marking down",
				},
				"check_path": map[string]any{
					"type":        "string",
					"description": "HTTP health check path",
				},
				"check_body": map[string]any{
					"type":        "string",
					"description": "Expected response body for http_body check",
				},
				"check_passive": map[string]any{
					"type":        "boolean",
					"description": "Enable passive health checks",
				},
				"proxy_protocol": map[string]any{
					"type":        "string",
					"description": "Proxy protocol version (none, v1, v2)",
				},
				"ssl_cert": map[string]any{
					"type":        "string",
					"description": "SSL certificate for HTTPS",
				},
				"ssl_key": map[string]any{
					"type":        "string",
					"description": "SSL private key for HTTPS",
				},
			},
			Required: []string{"nodebalancer_id", "port", "protocol"},
		},
	}, s.handleNodeBalancerConfigCreate)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_nodebalancer_config_update",
		Description: "Update a NodeBalancer configuration",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"nodebalancer_id": map[string]any{
					"type":        "number",
					"description": "ID of the NodeBalancer",
				},
				"config_id": map[string]any{
					"type":        "number",
					"description": "ID of the configuration to update",
				},
				"port": map[string]any{
					"type":        "number",
					"description": "Port to configure (1-65534)",
				},
				"protocol": map[string]any{
					"type":        "string",
					"description": "Protocol (http, https, tcp)",
				},
				"algorithm": map[string]any{
					"type":        "string",
					"description": "Balancing algorithm (roundrobin, leastconn, source)",
				},
				"stickiness": map[string]any{
					"type":        "string",
					"description": "Session stickiness (none, table, http_cookie)",
				},
				"check": map[string]any{
					"type":        "string",
					"description": "Health check type (none, connection, http, http_body)",
				},
				"check_interval": map[string]any{
					"type":        "number",
					"description": "Health check interval in seconds",
				},
				"check_timeout": map[string]any{
					"type":        "number",
					"description": "Health check timeout in seconds",
				},
				"check_attempts": map[string]any{
					"type":        "number",
					"description": "Health check attempts before marking down",
				},
				"check_path": map[string]any{
					"type":        "string",
					"description": "HTTP health check path",
				},
				"check_body": map[string]any{
					"type":        "string",
					"description": "Expected response body for http_body check",
				},
				"check_passive": map[string]any{
					"type":        "boolean",
					"description": "Enable passive health checks",
				},
				"proxy_protocol": map[string]any{
					"type":        "string",
					"description": "Proxy protocol version (none, v1, v2)",
				},
				"ssl_cert": map[string]any{
					"type":        "string",
					"description": "SSL certificate for HTTPS",
				},
				"ssl_key": map[string]any{
					"type":        "string",
					"description": "SSL private key for HTTPS",
				},
			},
			Required: []string{"nodebalancer_id", "config_id"},
		},
	}, s.handleNodeBalancerConfigUpdate)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_nodebalancer_config_delete",
		Description: "Delete a NodeBalancer configuration",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"nodebalancer_id": map[string]any{
					"type":        "number",
					"description": "ID of the NodeBalancer",
				},
				"config_id": map[string]any{
					"type":        "number",
					"description": "ID of the configuration to delete",
				},
			},
			Required: []string{"nodebalancer_id", "config_id"},
		},
	}, s.handleNodeBalancerConfigDelete)

	// Domain tools
	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_domains_list",
		Description: "List all domains",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
			Required:   []string{},
		},
	}, s.handleDomainsList)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_domain_get",
		Description: "Get details of a specific domain",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"domain_id": map[string]any{
					"type":        "number",
					"description": "ID of the domain",
				},
			},
			Required: []string{"domain_id"},
		},
	}, s.handleDomainGet)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_domain_create",
		Description: "Create a new domain",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"domain": map[string]any{
					"type":        "string",
					"description": "Domain name (e.g. example.com)",
				},
				"type": map[string]any{
					"type":        "string",
					"description": "Domain type (master or slave)",
				},
				"soa_email": map[string]any{
					"type":        "string",
					"description": "SOA email address",
				},
				"description": map[string]any{
					"type":        "string",
					"description": "Description of the domain",
				},
				"retry_sec": map[string]any{
					"type":        "number",
					"description": "Retry interval in seconds",
				},
				"master_ips": map[string]any{
					"type":        "array",
					"description": "Master IPs for slave domains",
				},
				"axfr_ips": map[string]any{
					"type":        "array",
					"description": "IPs allowed to AXFR the entire zone",
				},
				"expire_sec": map[string]any{
					"type":        "number",
					"description": "Expire time in seconds",
				},
				"refresh_sec": map[string]any{
					"type":        "number",
					"description": "Refresh time in seconds",
				},
				"ttl_sec": map[string]any{
					"type":        "number",
					"description": "Default TTL in seconds",
				},
				"tags": map[string]any{
					"type":        "array",
					"description": "Tags to apply to the domain",
				},
				"group": map[string]any{
					"type":        "string",
					"description": "Group for the domain",
				},
			},
			Required: []string{"domain", "type"},
		},
	}, s.handleDomainCreate)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_domain_update",
		Description: "Update an existing domain",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"domain_id": map[string]any{
					"type":        "number",
					"description": "ID of the domain to update",
				},
				"domain": map[string]any{
					"type":        "string",
					"description": "New domain name",
				},
				"type": map[string]any{
					"type":        "string",
					"description": "New domain type (master or slave)",
				},
				"soa_email": map[string]any{
					"type":        "string",
					"description": "New SOA email address",
				},
				"description": map[string]any{
					"type":        "string",
					"description": "New description",
				},
				"retry_sec": map[string]any{
					"type":        "number",
					"description": "New retry interval in seconds",
				},
				"master_ips": map[string]any{
					"type":        "array",
					"description": "New master IPs for slave domains",
				},
				"axfr_ips": map[string]any{
					"type":        "array",
					"description": "New IPs allowed to AXFR",
				},
				"expire_sec": map[string]any{
					"type":        "number",
					"description": "New expire time in seconds",
				},
				"refresh_sec": map[string]any{
					"type":        "number",
					"description": "New refresh time in seconds",
				},
				"ttl_sec": map[string]any{
					"type":        "number",
					"description": "New default TTL in seconds",
				},
				"tags": map[string]any{
					"type":        "array",
					"description": "New tags for the domain",
				},
				"group": map[string]any{
					"type":        "string",
					"description": "New group for the domain",
				},
			},
			Required: []string{"domain_id"},
		},
	}, s.handleDomainUpdate)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_domain_delete",
		Description: "Delete a domain",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"domain_id": map[string]any{
					"type":        "number",
					"description": "ID of the domain to delete",
				},
			},
			Required: []string{"domain_id"},
		},
	}, s.handleDomainDelete)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_domain_records_list",
		Description: "List all records for a domain",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"domain_id": map[string]any{
					"type":        "number",
					"description": "ID of the domain",
				},
			},
			Required: []string{"domain_id"},
		},
	}, s.handleDomainRecordsList)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_domain_record_get",
		Description: "Get details of a specific domain record",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"domain_id": map[string]any{
					"type":        "number",
					"description": "ID of the domain",
				},
				"record_id": map[string]any{
					"type":        "number",
					"description": "ID of the domain record",
				},
			},
			Required: []string{"domain_id", "record_id"},
		},
	}, s.handleDomainRecordGet)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_domain_record_create",
		Description: "Create a new domain record",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"domain_id": map[string]any{
					"type":        "number",
					"description": "ID of the domain",
				},
				"type": map[string]any{
					"type":        "string",
					"description": "Record type (A, AAAA, CNAME, MX, TXT, SRV, PTR, CAA, NS)",
				},
				"name": map[string]any{
					"type":        "string",
					"description": "Record name (subdomain)",
				},
				"target": map[string]any{
					"type":        "string",
					"description": "Record target (IP, hostname, etc.)",
				},
				"priority": map[string]any{
					"type":        "number",
					"description": "Record priority (for MX and SRV records)",
				},
				"weight": map[string]any{
					"type":        "number",
					"description": "Record weight (for SRV records)",
				},
				"port": map[string]any{
					"type":        "number",
					"description": "Record port (for SRV records)",
				},
				"service": map[string]any{
					"type":        "string",
					"description": "Service name (for SRV records)",
				},
				"protocol": map[string]any{
					"type":        "string",
					"description": "Protocol name (for SRV records)",
				},
				"ttl_sec": map[string]any{
					"type":        "number",
					"description": "TTL in seconds",
				},
				"tag": map[string]any{
					"type":        "string",
					"description": "CAA record tag",
				},
			},
			Required: []string{"domain_id", "type", "target"},
		},
	}, s.handleDomainRecordCreate)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_domain_record_update",
		Description: "Update a domain record",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"domain_id": map[string]any{
					"type":        "number",
					"description": "ID of the domain",
				},
				"record_id": map[string]any{
					"type":        "number",
					"description": "ID of the record to update",
				},
				"type": map[string]any{
					"type":        "string",
					"description": "New record type",
				},
				"name": map[string]any{
					"type":        "string",
					"description": "New record name",
				},
				"target": map[string]any{
					"type":        "string",
					"description": "New record target",
				},
				"priority": map[string]any{
					"type":        "number",
					"description": "New record priority",
				},
				"weight": map[string]any{
					"type":        "number",
					"description": "New record weight",
				},
				"port": map[string]any{
					"type":        "number",
					"description": "New record port",
				},
				"service": map[string]any{
					"type":        "string",
					"description": "New service name",
				},
				"protocol": map[string]any{
					"type":        "string",
					"description": "New protocol name",
				},
				"ttl_sec": map[string]any{
					"type":        "number",
					"description": "New TTL in seconds",
				},
				"tag": map[string]any{
					"type":        "string",
					"description": "New CAA record tag",
				},
			},
			Required: []string{"domain_id", "record_id"},
		},
	}, s.handleDomainRecordUpdate)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_domain_record_delete",
		Description: "Delete a domain record",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"domain_id": map[string]any{
					"type":        "number",
					"description": "ID of the domain",
				},
				"record_id": map[string]any{
					"type":        "number",
					"description": "ID of the record to delete",
				},
			},
			Required: []string{"domain_id", "record_id"},
		},
	}, s.handleDomainRecordDelete)

	// StackScript tools
	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_stackscripts_list",
		Description: "List all StackScripts",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
			Required:   []string{},
		},
	}, s.handleStackScriptsList)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_stackscript_get",
		Description: "Get details of a specific StackScript",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"stackscript_id": map[string]any{
					"type":        "number",
					"description": "ID of the StackScript",
				},
			},
			Required: []string{"stackscript_id"},
		},
	}, s.handleStackScriptGet)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_stackscript_create",
		Description: "Create a new StackScript",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"label": map[string]any{
					"type":        "string",
					"description": "Display label for the StackScript",
				},
				"description": map[string]any{
					"type":        "string",
					"description": "Description of the StackScript",
				},
				"images": map[string]any{
					"type":        "array",
					"description": "Compatible image IDs",
				},
				"script": map[string]any{
					"type":        "string",
					"description": "StackScript code",
				},
				"is_public": map[string]any{
					"type":        "boolean",
					"description": "Whether the StackScript should be public",
				},
				"rev_note": map[string]any{
					"type":        "string",
					"description": "Revision note for this version",
				},
			},
			Required: []string{"label", "images", "script"},
		},
	}, s.handleStackScriptCreate)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_stackscript_update",
		Description: "Update an existing StackScript",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"stackscript_id": map[string]any{
					"type":        "number",
					"description": "ID of the StackScript to update",
				},
				"label": map[string]any{
					"type":        "string",
					"description": "New display label",
				},
				"description": map[string]any{
					"type":        "string",
					"description": "New description",
				},
				"images": map[string]any{
					"type":        "array",
					"description": "New compatible image IDs",
				},
				"script": map[string]any{
					"type":        "string",
					"description": "New StackScript code",
				},
				"is_public": map[string]any{
					"type":        "boolean",
					"description": "New public status",
				},
				"rev_note": map[string]any{
					"type":        "string",
					"description": "Revision note for this version",
				},
			},
			Required: []string{"stackscript_id"},
		},
	}, s.handleStackScriptUpdate)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_stackscript_delete",
		Description: "Delete a StackScript",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"stackscript_id": map[string]any{
					"type":        "number",
					"description": "ID of the StackScript to delete",
				},
			},
			Required: []string{"stackscript_id"},
		},
	}, s.handleStackScriptDelete)

	// LKE (Kubernetes) tools
	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_lke_clusters_list",
		Description: "List all LKE clusters",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
			Required:   []string{},
		},
	}, s.handleLKEClustersList)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_lke_cluster_get",
		Description: "Get details of a specific LKE cluster",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"cluster_id": map[string]any{
					"type":        "number",
					"description": "ID of the LKE cluster",
				},
			},
			Required: []string{"cluster_id"},
		},
	}, s.handleLKEClusterGet)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_lke_cluster_create",
		Description: "Create a new LKE cluster",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"label": map[string]any{
					"type":        "string",
					"description": "Display label for the cluster",
				},
				"region": map[string]any{
					"type":        "string",
					"description": "Region ID where the cluster will be created",
				},
				"k8s_version": map[string]any{
					"type":        "string",
					"description": "Kubernetes version (e.g. 1.28)",
				},
				"tags": map[string]any{
					"type":        "array",
					"description": "Tags to apply to the cluster",
				},
				"node_pools": map[string]any{
					"type":        "array",
					"description": "Node pool specifications",
				},
				"control_plane": map[string]any{
					"type":        "object",
					"description": "Control plane configuration",
				},
			},
			Required: []string{"label", "region", "k8s_version", "node_pools"},
		},
	}, s.handleLKEClusterCreate)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_lke_cluster_update",
		Description: "Update an existing LKE cluster",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"cluster_id": map[string]any{
					"type":        "number",
					"description": "ID of the cluster to update",
				},
				"label": map[string]any{
					"type":        "string",
					"description": "New display label",
				},
				"k8s_version": map[string]any{
					"type":        "string",
					"description": "New Kubernetes version",
				},
				"tags": map[string]any{
					"type":        "array",
					"description": "New tags for the cluster",
				},
				"control_plane": map[string]any{
					"type":        "object",
					"description": "New control plane configuration",
				},
			},
			Required: []string{"cluster_id"},
		},
	}, s.handleLKEClusterUpdate)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_lke_cluster_delete",
		Description: "Delete an LKE cluster",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"cluster_id": map[string]any{
					"type":        "number",
					"description": "ID of the cluster to delete",
				},
			},
			Required: []string{"cluster_id"},
		},
	}, s.handleLKEClusterDelete)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_lke_nodepool_create",
		Description: "Create a new node pool in an LKE cluster",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"cluster_id": map[string]any{
					"type":        "number",
					"description": "ID of the cluster",
				},
				"type": map[string]any{
					"type":        "string",
					"description": "Linode type for nodes",
				},
				"count": map[string]any{
					"type":        "number",
					"description": "Number of nodes in pool",
				},
				"disks": map[string]any{
					"type":        "array",
					"description": "Disk configuration for nodes",
				},
				"autoscaler": map[string]any{
					"type":        "object",
					"description": "Autoscaler configuration",
				},
				"tags": map[string]any{
					"type":        "array",
					"description": "Tags for the node pool",
				},
			},
			Required: []string{"cluster_id", "type", "count"},
		},
	}, s.handleLKENodePoolCreate)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_lke_nodepool_update",
		Description: "Update a node pool in an LKE cluster",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"cluster_id": map[string]any{
					"type":        "number",
					"description": "ID of the cluster",
				},
				"pool_id": map[string]any{
					"type":        "number",
					"description": "ID of the node pool to update",
				},
				"count": map[string]any{
					"type":        "number",
					"description": "New number of nodes",
				},
				"autoscaler": map[string]any{
					"type":        "object",
					"description": "New autoscaler configuration",
				},
				"tags": map[string]any{
					"type":        "array",
					"description": "New tags for the node pool",
				},
			},
			Required: []string{"cluster_id", "pool_id"},
		},
	}, s.handleLKENodePoolUpdate)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_lke_nodepool_delete",
		Description: "Delete a node pool from an LKE cluster",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"cluster_id": map[string]any{
					"type":        "number",
					"description": "ID of the cluster",
				},
				"pool_id": map[string]any{
					"type":        "number",
					"description": "ID of the node pool to delete",
				},
			},
			Required: []string{"cluster_id", "pool_id"},
		},
	}, s.handleLKENodePoolDelete)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_lke_kubeconfig",
		Description: "Retrieve the kubeconfig for an LKE cluster",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"cluster_id": map[string]any{
					"type":        "number",
					"description": "ID of the cluster",
				},
			},
			Required: []string{"cluster_id"},
		},
	}, s.handleLKEKubeconfig)

	// Database tools
	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_databases_list",
		Description: "List all databases (both MySQL and PostgreSQL)",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
			Required:   []string{},
		},
	}, s.handleDatabasesList)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_mysql_databases_list",
		Description: "List all MySQL databases",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
			Required:   []string{},
		},
	}, s.handleMySQLDatabasesList)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_postgres_databases_list",
		Description: "List all PostgreSQL databases",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
			Required:   []string{},
		},
	}, s.handlePostgresDatabasesList)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_mysql_database_get",
		Description: "Get details of a specific MySQL database",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"database_id": map[string]any{
					"type":        "number",
					"description": "ID of the MySQL database",
				},
			},
			Required: []string{"database_id"},
		},
	}, s.handleMySQLDatabaseGet)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_postgres_database_get",
		Description: "Get details of a specific PostgreSQL database",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"database_id": map[string]any{
					"type":        "number",
					"description": "ID of the PostgreSQL database",
				},
			},
			Required: []string{"database_id"},
		},
	}, s.handlePostgresDatabaseGet)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_mysql_database_create",
		Description: "Create a new MySQL database",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"label": map[string]any{
					"type":        "string",
					"description": "Display label for the database",
				},
				"region": map[string]any{
					"type":        "string",
					"description": "Region where the database will be created",
				},
				"type": map[string]any{
					"type":        "string",
					"description": "Database type (e.g. g6-nanode-1)",
				},
				"engine": map[string]any{
					"type":        "string",
					"description": "Database engine (e.g. mysql/8.0.30)",
				},
				"cluster_size": map[string]any{
					"type":        "number",
					"description": "Number of nodes in the cluster (1 or 3)",
				},
				"allow_list": map[string]any{
					"type":        "array",
					"description": "List of IP addresses/ranges allowed to access the database",
				},
			},
			Required: []string{"label", "region", "type", "engine"},
		},
	}, s.handleMySQLDatabaseCreate)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_postgres_database_create",
		Description: "Create a new PostgreSQL database",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"label": map[string]any{
					"type":        "string",
					"description": "Display label for the database",
				},
				"region": map[string]any{
					"type":        "string",
					"description": "Region where the database will be created",
				},
				"type": map[string]any{
					"type":        "string",
					"description": "Database type (e.g. g6-nanode-1)",
				},
				"engine": map[string]any{
					"type":        "string",
					"description": "Database engine (e.g. postgresql/14.9)",
				},
				"cluster_size": map[string]any{
					"type":        "number",
					"description": "Number of nodes in the cluster (1 or 3)",
				},
				"allow_list": map[string]any{
					"type":        "array",
					"description": "List of IP addresses/ranges allowed to access the database",
				},
			},
			Required: []string{"label", "region", "type", "engine"},
		},
	}, s.handlePostgresDatabaseCreate)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_mysql_database_update",
		Description: "Update an existing MySQL database",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"database_id": map[string]any{
					"type":        "number",
					"description": "ID of the MySQL database to update",
				},
				"label": map[string]any{
					"type":        "string",
					"description": "New display label for the database",
				},
				"allow_list": map[string]any{
					"type":        "array",
					"description": "Updated list of IP addresses/ranges allowed to access the database",
				},
			},
			Required: []string{"database_id"},
		},
	}, s.handleMySQLDatabaseUpdate)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_postgres_database_update",
		Description: "Update an existing PostgreSQL database",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"database_id": map[string]any{
					"type":        "number",
					"description": "ID of the PostgreSQL database to update",
				},
				"label": map[string]any{
					"type":        "string",
					"description": "New display label for the database",
				},
				"allow_list": map[string]any{
					"type":        "array",
					"description": "Updated list of IP addresses/ranges allowed to access the database",
				},
			},
			Required: []string{"database_id"},
		},
	}, s.handlePostgresDatabaseUpdate)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_mysql_database_delete",
		Description: "Delete a MySQL database",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"database_id": map[string]any{
					"type":        "number",
					"description": "ID of the MySQL database to delete",
				},
			},
			Required: []string{"database_id"},
		},
	}, s.handleMySQLDatabaseDelete)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_postgres_database_delete",
		Description: "Delete a PostgreSQL database",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"database_id": map[string]any{
					"type":        "number",
					"description": "ID of the PostgreSQL database to delete",
				},
			},
			Required: []string{"database_id"},
		},
	}, s.handlePostgresDatabaseDelete)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_mysql_database_credentials",
		Description: "Get root credentials for a MySQL database",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"database_id": map[string]any{
					"type":        "number",
					"description": "ID of the MySQL database",
				},
			},
			Required: []string{"database_id"},
		},
	}, s.handleMySQLDatabaseCredentials)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_postgres_database_credentials",
		Description: "Get root credentials for a PostgreSQL database",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"database_id": map[string]any{
					"type":        "number",
					"description": "ID of the PostgreSQL database",
				},
			},
			Required: []string{"database_id"},
		},
	}, s.handlePostgresDatabaseCredentials)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_mysql_database_credentials_reset",
		Description: "Reset root password for a MySQL database",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"database_id": map[string]any{
					"type":        "number",
					"description": "ID of the MySQL database",
				},
			},
			Required: []string{"database_id"},
		},
	}, s.handleMySQLDatabaseCredentialsReset)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_postgres_database_credentials_reset",
		Description: "Reset root password for a PostgreSQL database",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"database_id": map[string]any{
					"type":        "number",
					"description": "ID of the PostgreSQL database",
				},
			},
			Required: []string{"database_id"},
		},
	}, s.handlePostgresDatabaseCredentialsReset)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_database_engines_list",
		Description: "List all available database engines",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
			Required:   []string{},
		},
	}, s.handleDatabaseEnginesList)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_database_types_list",
		Description: "List all available database types",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
			Required:   []string{},
		},
	}, s.handleDatabaseTypesList)

	// Object Storage tools
	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_objectstorage_buckets_list",
		Description: "List all Object Storage buckets",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
			Required:   []string{},
		},
	}, s.handleObjectStorageBucketsList)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_objectstorage_bucket_get",
		Description: "Get details of a specific Object Storage bucket",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"cluster": map[string]any{
					"type":        "string",
					"description": "Object Storage cluster ID",
				},
				"bucket": map[string]any{
					"type":        "string",
					"description": "Bucket name",
				},
			},
			Required: []string{"cluster", "bucket"},
		},
	}, s.handleObjectStorageBucketGet)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_objectstorage_bucket_create",
		Description: "Create a new Object Storage bucket",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"label": map[string]any{
					"type":        "string",
					"description": "Bucket name/label",
				},
				"cluster": map[string]any{
					"type":        "string",
					"description": "Object Storage cluster ID (deprecated, use region)",
				},
				"region": map[string]any{
					"type":        "string",
					"description": "Object Storage region ID",
				},
				"acl": map[string]any{
					"type":        "string",
					"description": "Access control list (private, public-read, public-read-write, authenticated-read)",
				},
				"cors_enabled": map[string]any{
					"type":        "boolean",
					"description": "Enable CORS",
				},
			},
			Required: []string{"label"},
		},
	}, s.handleObjectStorageBucketCreate)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_objectstorage_bucket_update",
		Description: "Update an existing Object Storage bucket",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"cluster": map[string]any{
					"type":        "string",
					"description": "Object Storage cluster ID",
				},
				"bucket": map[string]any{
					"type":        "string",
					"description": "Bucket name",
				},
				"acl": map[string]any{
					"type":        "string",
					"description": "New access control list",
				},
				"cors_enabled": map[string]any{
					"type":        "boolean",
					"description": "Enable or disable CORS",
				},
			},
			Required: []string{"cluster", "bucket"},
		},
	}, s.handleObjectStorageBucketUpdate)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_objectstorage_bucket_delete",
		Description: "Delete an Object Storage bucket",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"cluster": map[string]any{
					"type":        "string",
					"description": "Object Storage cluster ID",
				},
				"bucket": map[string]any{
					"type":        "string",
					"description": "Bucket name",
				},
			},
			Required: []string{"cluster", "bucket"},
		},
	}, s.handleObjectStorageBucketDelete)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_objectstorage_keys_list",
		Description: "List all Object Storage keys",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
			Required:   []string{},
		},
	}, s.handleObjectStorageKeysList)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_objectstorage_key_get",
		Description: "Get details of a specific Object Storage key",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"key_id": map[string]any{
					"type":        "number",
					"description": "ID of the Object Storage key",
				},
			},
			Required: []string{"key_id"},
		},
	}, s.handleObjectStorageKeyGet)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_objectstorage_key_create",
		Description: "Create a new Object Storage key",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"label": map[string]any{
					"type":        "string",
					"description": "Display label for the key",
				},
				"bucket_access": map[string]any{
					"type":        "array",
					"description": "Bucket access permissions (optional, omit for full access)",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"cluster": map[string]any{
								"type":        "string",
								"description": "Cluster ID",
							},
							"bucket_name": map[string]any{
								"type":        "string",
								"description": "Bucket name",
							},
							"permissions": map[string]any{
								"type":        "string",
								"description": "Permission level (read_only, read_write)",
							},
						},
					},
				},
			},
			Required: []string{"label"},
		},
	}, s.handleObjectStorageKeyCreate)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_objectstorage_key_update",
		Description: "Update an existing Object Storage key",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"key_id": map[string]any{
					"type":        "number",
					"description": "ID of the key to update",
				},
				"label": map[string]any{
					"type":        "string",
					"description": "New display label",
				},
				"bucket_access": map[string]any{
					"type":        "array",
					"description": "New bucket access permissions",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"cluster": map[string]any{
								"type":        "string",
								"description": "Cluster ID",
							},
							"bucket_name": map[string]any{
								"type":        "string",
								"description": "Bucket name",
							},
							"permissions": map[string]any{
								"type":        "string",
								"description": "Permission level (read_only, read_write)",
							},
						},
					},
				},
			},
			Required: []string{"key_id"},
		},
	}, s.handleObjectStorageKeyUpdate)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_objectstorage_key_delete",
		Description: "Delete an Object Storage key",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"key_id": map[string]any{
					"type":        "number",
					"description": "ID of the key to delete",
				},
			},
			Required: []string{"key_id"},
		},
	}, s.handleObjectStorageKeyDelete)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_objectstorage_clusters_list",
		Description: "List all Object Storage clusters",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
			Required:   []string{},
		},
	}, s.handleObjectStorageClustersList)

	// Advanced Networking tools
	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_reserved_ips_list",
		Description: "List all reserved IP addresses",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
			Required:   []string{},
		},
	}, s.handleReservedIPsList)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_reserved_ip_get",
		Description: "Get details of a specific reserved IP address",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"address": map[string]any{
					"type":        "string",
					"description": "IP address to get details for",
				},
			},
			Required: []string{"address"},
		},
	}, s.handleReservedIPGet)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_reserved_ip_allocate",
		Description: "Allocate a new reserved IP address",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"type": map[string]any{
					"type":        "string",
					"description": "Type of IP (ipv4, ipv6)",
				},
				"public": map[string]any{
					"type":        "boolean",
					"description": "Whether the IP should be public",
				},
				"linode_id": map[string]any{
					"type":        "number",
					"description": "Linode to assign the IP to",
				},
				"region": map[string]any{
					"type":        "string",
					"description": "Region for the IP (required if no linode_id)",
				},
			},
			Required: []string{"type"},
		},
	}, s.handleReservedIPAllocate)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_reserved_ip_assign",
		Description: "Assign a reserved IP to a Linode or unassign it",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"address": map[string]any{
					"type":        "string",
					"description": "IP address to assign",
				},
				"linode_id": map[string]any{
					"type":        "number",
					"description": "Linode to assign to (null to unassign)",
				},
			},
			Required: []string{"address"},
		},
	}, s.handleReservedIPAssign)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_reserved_ip_update",
		Description: "Update the reverse DNS for a reserved IP",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"address": map[string]any{
					"type":        "string",
					"description": "IP address to update",
				},
				"rdns": map[string]any{
					"type":        "string",
					"description": "Reverse DNS for the IP",
				},
			},
			Required: []string{"address"},
		},
	}, s.handleReservedIPUpdate)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_vlans_list",
		Description: "List all VLANs",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
			Required:   []string{},
		},
	}, s.handleVLANsList)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_ipv6_pools_list",
		Description: "List all IPv6 pools",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
			Required:   []string{},
		},
	}, s.handleIPv6PoolsList)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_ipv6_ranges_list",
		Description: "List all IPv6 ranges",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
			Required:   []string{},
		},
	}, s.handleIPv6RangesList)

	// Monitoring (Longview) tools
	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_longview_clients_list",
		Description: "List all Longview monitoring clients",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
			Required:   []string{},
		},
	}, s.handleLongviewClientsList)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_longview_client_get",
		Description: "Get details of a specific Longview client",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"client_id": map[string]any{
					"type":        "number",
					"description": "ID of the Longview client",
				},
			},
			Required: []string{"client_id"},
		},
	}, s.handleLongviewClientGet)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_longview_client_create",
		Description: "Create a new Longview monitoring client",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"label": map[string]any{
					"type":        "string",
					"description": "Display label for the Longview client",
				},
			},
			Required: []string{"label"},
		},
	}, s.handleLongviewClientCreate)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_longview_client_update",
		Description: "Update an existing Longview client",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"client_id": map[string]any{
					"type":        "number",
					"description": "ID of the client to update",
				},
				"label": map[string]any{
					"type":        "string",
					"description": "New display label",
				},
			},
			Required: []string{"client_id"},
		},
	}, s.handleLongviewClientUpdate)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_longview_client_delete",
		Description: "Delete a Longview monitoring client",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"client_id": map[string]any{
					"type":        "number",
					"description": "ID of the client to delete",
				},
			},
			Required: []string{"client_id"},
		},
	}, s.handleLongviewClientDelete)

	// Support tools
	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_support_tickets_list",
		Description: "List all support tickets",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
			Required:   []string{},
		},
	}, s.handleSupportTicketsList)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_support_ticket_get",
		Description: "Get details of a specific support ticket",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"ticket_id": map[string]any{
					"type":        "number",
					"description": "ID of the support ticket",
				},
			},
			Required: []string{"ticket_id"},
		},
	}, s.handleSupportTicketGet)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_support_ticket_create",
		Description: "Create a new support ticket",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"summary": map[string]any{
					"type":        "string",
					"description": "Brief summary of the issue",
				},
				"description": map[string]any{
					"type":        "string",
					"description": "Detailed description of the issue",
				},
				"linode_id": map[string]any{
					"type":        "number",
					"description": "ID of related Linode",
				},
				"domain_id": map[string]any{
					"type":        "number",
					"description": "ID of related domain",
				},
				"nodebalancer_id": map[string]any{
					"type":        "number",
					"description": "ID of related NodeBalancer",
				},
				"volume_id": map[string]any{
					"type":        "number",
					"description": "ID of related volume",
				},
			},
			Required: []string{"summary", "description"},
		},
	}, s.handleSupportTicketCreate)

	mcpServer.AddTool(mcp.Tool{
		Name:        "linode_support_ticket_reply",
		Description: "Add a reply to an existing support ticket",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"ticket_id": map[string]any{
					"type":        "number",
					"description": "ID of the ticket to reply to",
				},
				"description": map[string]any{
					"type":        "string",
					"description": "Reply content",
				},
			},
			Required: []string{"ticket_id", "description"},
		},
	}, s.handleSupportTicketReply)

	return nil
}

func (am *AccountManager) GetCurrentAccount() (*Account, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	account, exists := am.accounts[am.currentAccount]
	if !exists {
		return nil, types.NewServiceError("linode",
			fmt.Sprintf("current account %q not found", am.currentAccount), nil)
	}

	return account, nil
}

func (am *AccountManager) GetAccount(name string) (*Account, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	account, exists := am.accounts[name]
	if !exists {
		return nil, types.NewServiceError("linode",
			fmt.Sprintf("account %q not found", name), nil)
	}

	return account, nil
}

func (am *AccountManager) SwitchAccount(name string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if _, exists := am.accounts[name]; !exists {
		return types.NewServiceError("linode",
			fmt.Sprintf("account %q not found", name), nil)
	}

	am.currentAccount = name

	return nil
}

func (am *AccountManager) ListAccounts() map[string]string {
	am.mu.RLock()
	defer am.mu.RUnlock()

	accounts := make(map[string]string)
	for name, account := range am.accounts {
		accounts[name] = account.Label
	}

	return accounts
}

// NewAccountManagerForTesting creates an empty AccountManager instance for testing.
// This allows external test packages to create isolated AccountManager instances.
func NewAccountManagerForTesting() *AccountManager {
	return &AccountManager{
		accounts:       make(map[string]*Account),
		currentAccount: "",
		mu:             sync.RWMutex{},
	}
}

// NewAccountForTesting creates an Account instance for testing.
// This allows external test packages to create isolated Account instances.
func NewAccountForTesting(name, label string) *Account {
	return &Account{
		Name:   name,
		Label:  label,
		Client: nil, // Use nil for unit tests
	}
}

// AddAccountForTesting adds an account to the AccountManager for testing.
// This allows external test packages to set up test accounts.
func (am *AccountManager) AddAccountForTesting(account *Account) {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.accounts[account.Name] = account
	if am.currentAccount == "" {
		am.currentAccount = account.Name
	}
}

// SetCurrentAccountForTesting sets the current account for testing.
// This allows external test packages to control which account is current.
func (am *AccountManager) SetCurrentAccountForTesting(name string) {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.currentAccount = name
}

// NewForTesting creates a service instance for testing with the provided configuration and dependencies.
// This function allows external test packages to create service instances with custom dependencies.
func NewForTesting(cfg *config.Config, log logger.Logger, accountManager *AccountManager) *Service {
	return &Service{
		config:         cfg,
		logger:         log,
		accountManager: accountManager,
	}
}

// CallToolForTesting provides a public interface for testing tool handlers.
// This allows external test packages to call tool handlers through the public API.
func (s *Service) CallToolForTesting(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	switch request.Params.Name {
	case "linode_account_get":
		return s.handleAccountGet(ctx, request)
	case "linode_account_list":
		return s.handleAccountList(ctx, request)
	case "linode_account_switch":
		return s.handleAccountSwitch(ctx, request)
	case "linode_instances_list":
		return s.handleInstancesList(ctx, request)
	case "linode_instance_get":
		return s.handleInstanceGet(ctx, request)
	case "linode_instance_create":
		return s.handleInstanceCreate(ctx, request)
	case "linode_instance_delete":
		return s.handleInstanceDelete(ctx, request)
	case "linode_instance_boot":
		return s.handleInstanceBoot(ctx, request)
	case "linode_instance_shutdown":
		return s.handleInstanceShutdown(ctx, request)
	case "linode_instance_reboot":
		return s.handleInstanceReboot(ctx, request)
	case "linode_volumes_list":
		return s.handleVolumesList(ctx, request)
	case "linode_volume_get":
		return s.handleVolumeGet(ctx, request)
	case "linode_volume_create":
		return s.handleVolumeCreate(ctx, request)
	case "linode_volume_delete":
		return s.handleVolumeDelete(ctx, request)
	case "linode_volume_attach":
		return s.handleVolumeAttach(ctx, request)
	case "linode_volume_detach":
		return s.handleVolumeDetach(ctx, request)
	case "linode_images_list":
		return s.handleImagesList(ctx, request)
	case "linode_image_get":
		return s.handleImageGet(ctx, request)
	case "linode_image_create":
		return s.handleImageCreate(ctx, request)
	case "linode_image_update":
		return s.handleImageUpdate(ctx, request)
	case "linode_image_delete":
		return s.handleImageDelete(ctx, request)
	case "linode_image_replicate":
		return s.handleImageReplicate(ctx, request)
	case "linode_image_upload_create":
		return s.handleImageUploadCreate(ctx, request)
	case "linode_ips_list":
		return s.handleIPsList(ctx, request)
	case "linode_ip_get":
		return s.handleIPGet(ctx, request)
	case "linode_firewalls_list":
		return s.handleFirewallsList(ctx, request)
	case "linode_firewall_get":
		return s.handleFirewallGet(ctx, request)
	case "linode_firewall_create":
		return s.handleFirewallCreate(ctx, request)
	case "linode_firewall_update":
		return s.handleFirewallUpdate(ctx, request)
	case "linode_firewall_delete":
		return s.handleFirewallDelete(ctx, request)
	case "linode_firewall_rules_update":
		return s.handleFirewallRulesUpdate(ctx, request)
	case "linode_firewall_device_create":
		return s.handleFirewallDeviceCreate(ctx, request)
	case "linode_firewall_device_delete":
		return s.handleFirewallDeviceDelete(ctx, request)
	case "linode_nodebalancers_list":
		return s.handleNodeBalancersList(ctx, request)
	case "linode_nodebalancer_get":
		return s.handleNodeBalancerGet(ctx, request)
	case "linode_nodebalancer_create":
		return s.handleNodeBalancerCreate(ctx, request)
	case "linode_nodebalancer_update":
		return s.handleNodeBalancerUpdate(ctx, request)
	case "linode_nodebalancer_delete":
		return s.handleNodeBalancerDelete(ctx, request)
	case "linode_nodebalancer_config_create":
		return s.handleNodeBalancerConfigCreate(ctx, request)
	case "linode_nodebalancer_config_update":
		return s.handleNodeBalancerConfigUpdate(ctx, request)
	case "linode_nodebalancer_config_delete":
		return s.handleNodeBalancerConfigDelete(ctx, request)
	case "linode_domains_list":
		return s.handleDomainsList(ctx, request)
	case "linode_domain_get":
		return s.handleDomainGet(ctx, request)
	case "linode_domain_create":
		return s.handleDomainCreate(ctx, request)
	case "linode_domain_update":
		return s.handleDomainUpdate(ctx, request)
	case "linode_domain_delete":
		return s.handleDomainDelete(ctx, request)
	case "linode_domain_records_list":
		return s.handleDomainRecordsList(ctx, request)
	case "linode_domain_record_get":
		return s.handleDomainRecordGet(ctx, request)
	case "linode_domain_record_create":
		return s.handleDomainRecordCreate(ctx, request)
	case "linode_domain_record_update":
		return s.handleDomainRecordUpdate(ctx, request)
	case "linode_domain_record_delete":
		return s.handleDomainRecordDelete(ctx, request)
	case "linode_stackscripts_list":
		return s.handleStackScriptsList(ctx, request)
	case "linode_stackscript_get":
		return s.handleStackScriptGet(ctx, request)
	case "linode_stackscript_create":
		return s.handleStackScriptCreate(ctx, request)
	case "linode_stackscript_update":
		return s.handleStackScriptUpdate(ctx, request)
	case "linode_stackscript_delete":
		return s.handleStackScriptDelete(ctx, request)
	case "linode_lke_clusters_list":
		return s.handleLKEClustersList(ctx, request)
	case "linode_lke_cluster_get":
		return s.handleLKEClusterGet(ctx, request)
	case "linode_lke_cluster_create":
		return s.handleLKEClusterCreate(ctx, request)
	case "linode_lke_cluster_update":
		return s.handleLKEClusterUpdate(ctx, request)
	case "linode_lke_cluster_delete":
		return s.handleLKEClusterDelete(ctx, request)
	case "linode_lke_node_pool_create":
		return s.handleLKENodePoolCreate(ctx, request)
	case "linode_lke_node_pool_update":
		return s.handleLKENodePoolUpdate(ctx, request)
	case "linode_lke_node_pool_delete":
		return s.handleLKENodePoolDelete(ctx, request)
	case "linode_lke_kubeconfig":
		return s.handleLKEKubeconfig(ctx, request)
	case "linode_databases_list":
		return s.handleDatabasesList(ctx, request)
	case "linode_mysql_databases_list":
		return s.handleMySQLDatabasesList(ctx, request)
	case "linode_postgres_databases_list":
		return s.handlePostgresDatabasesList(ctx, request)
	case "linode_mysql_database_get":
		return s.handleMySQLDatabaseGet(ctx, request)
	case "linode_postgres_database_get":
		return s.handlePostgresDatabaseGet(ctx, request)
	case "linode_mysql_database_create":
		return s.handleMySQLDatabaseCreate(ctx, request)
	case "linode_mysql_database_delete":
		return s.handleMySQLDatabaseDelete(ctx, request)
	case "linode_mysql_database_credentials_reset":
		return s.handleMySQLDatabaseCredentialsReset(ctx, request)
	case "linode_database_engines_list":
		return s.handleDatabaseEnginesList(ctx, request)
	case "linode_object_storage_buckets_list":
		return s.handleObjectStorageBucketsList(ctx, request)
	case "linode_object_storage_bucket_get":
		return s.handleObjectStorageBucketGet(ctx, request)
	case "linode_object_storage_bucket_create":
		return s.handleObjectStorageBucketCreate(ctx, request)
	case "linode_object_storage_bucket_update":
		return s.handleObjectStorageBucketUpdate(ctx, request)
	case "linode_object_storage_bucket_delete":
		return s.handleObjectStorageBucketDelete(ctx, request)
	case "linode_object_storage_keys_list":
		return s.handleObjectStorageKeysList(ctx, request)
	case "linode_object_storage_key_get":
		return s.handleObjectStorageKeyGet(ctx, request)
	case "linode_object_storage_key_create":
		return s.handleObjectStorageKeyCreate(ctx, request)
	case "linode_object_storage_key_update":
		return s.handleObjectStorageKeyUpdate(ctx, request)
	case "linode_object_storage_key_delete":
		return s.handleObjectStorageKeyDelete(ctx, request)
	case "linode_object_storage_clusters_list":
		return s.handleObjectStorageClustersList(ctx, request)
	case "linode_reserved_ips_list":
		return s.handleReservedIPsList(ctx, request)
	case "linode_reserved_ip_get":
		return s.handleReservedIPGet(ctx, request)
	case "linode_reserved_ip_allocate":
		return s.handleReservedIPAllocate(ctx, request)
	case "linode_reserved_ip_assign":
		return s.handleReservedIPAssign(ctx, request)
	case "linode_reserved_ip_update":
		return s.handleReservedIPUpdate(ctx, request)
	case "linode_vlans_list":
		return s.handleVLANsList(ctx, request)
	case "linode_ipv6_pools_list":
		return s.handleIPv6PoolsList(ctx, request)
	case "linode_ipv6_ranges_list":
		return s.handleIPv6RangesList(ctx, request)
	case "linode_support_tickets_list":
		return s.handleSupportTicketsList(ctx, request)
	case "linode_support_ticket_get":
		return s.handleSupportTicketGet(ctx, request)
	case "linode_support_ticket_create":
		return s.handleSupportTicketCreate(ctx, request)
	case "linode_support_ticket_reply":
		return s.handleSupportTicketReply(ctx, request)
	case "linode_longview_clients_list":
		return s.handleLongviewClientsList(ctx, request)
	case "linode_longview_client_get":
		return s.handleLongviewClientGet(ctx, request)
	case "linode_longview_client_create":
		return s.handleLongviewClientCreate(ctx, request)
	case "linode_longview_client_update":
		return s.handleLongviewClientUpdate(ctx, request)
	case "linode_longview_client_delete":
		return s.handleLongviewClientDelete(ctx, request)
	case "cloudmcp_version":
		return s.handleSystemVersion(ctx, request)
	case "cloudmcp_version_json":
		return s.handleSystemVersionJSON(ctx, request)
	default:
		return nil, fmt.Errorf("unsupported tool for testing: %s", request.Params.Name)
	}
}

// GetTextContentForTesting extracts text content from CallToolResult for testing.
// This allows external test packages to access the text content helper function.
func GetTextContentForTesting(t interface {
	Helper()
	Errorf(format string, args ...interface{})
	FailNow()
}, result *mcp.CallToolResult,
) string {
	t.Helper()

	if result == nil {
		t.Errorf("result should not be nil")
		t.FailNow()
	}

	if len(result.Content) == 0 {
		t.Errorf("result should have content")
		t.FailNow()
	}

	if len(result.Content) != 1 {
		t.Errorf("result should have exactly one content item, got %d", len(result.Content))
		t.FailNow()
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Errorf("result content should be text content")
		t.FailNow()
	}

	if textContent.Text == "" {
		t.Errorf("result text should not be empty")
		t.FailNow()
	}

	return textContent.Text
}
