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
	mu             sync.RWMutex
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

