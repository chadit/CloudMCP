package linode

import (
	"context"
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

type Service struct {
	config         *config.Config
	logger         logger.Logger
	accountManager *AccountManager
}

type AccountManager struct {
	accounts       map[string]*Account
	currentAccount string
	mu             sync.RWMutex
}

type Account struct {
	Name   string
	Label  string
	Client *linodego.Client
}

func New(cfg *config.Config, log logger.Logger) (*Service, error) {
	s := &Service{
		config: cfg,
		logger: log,
		accountManager: &AccountManager{
			accounts:       make(map[string]*Account),
			currentAccount: cfg.DefaultLinodeAccount,
		},
	}

	for name, accCfg := range cfg.LinodeAccounts {
		if accCfg.Token == "" {
			return nil, fmt.Errorf("account %q has no token configured", name)
		}

		tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accCfg.Token})
		oauth2Client := oauth2.NewClient(context.Background(), tokenSource)

		client := linodego.NewClient(oauth2Client)
		client.SetUserAgent("CloudMCP/0.1.0")

		// Set custom API URL if provided (e.g., for development environments)
		if accCfg.APIURL != "" {
			client.SetBaseURL(accCfg.APIURL)
		}

		s.accountManager.accounts[name] = &Account{
			Name:   name,
			Label:  accCfg.Label,
			Client: &client,
		}
	}

	return s, nil
}

func (s *Service) Name() string {
	return "linode"
}

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

func (s *Service) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down Linode service")
	return nil
}

func (s *Service) RegisterTools(server interface{}) error {
	mcpServer, ok := server.(*mcpserver.MCPServer)
	if !ok {
		return fmt.Errorf("invalid server type")
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
