package linode

import (
	"context"
	"fmt"
	"strings"

	"github.com/linode/linodego"
	"github.com/mark3labs/mcp-go/mcp"
	"golang.org/x/oauth2"

	"github.com/chadit/CloudMCP/internal/config"
	"github.com/chadit/CloudMCP/pkg/types"
)

// CloudMCPAccountAddParams represents parameters for adding a new account to configuration.
type CloudMCPAccountAddParams struct {
	Name   string `json:"name"`
	Token  string `json:"token"`
	Label  string `json:"label"`
	APIURL string `json:"apiurl,omitempty"`
}

// CloudMCPAccountRemoveParams represents parameters for removing an account from configuration.
type CloudMCPAccountRemoveParams struct {
	Name string `json:"name"`
}

// CloudMCPAccountUpdateParams represents parameters for updating an existing account.
type CloudMCPAccountUpdateParams struct {
	Name   string `json:"name"`
	Token  string `json:"token,omitempty"`
	Label  string `json:"label,omitempty"`
	APIURL string `json:"apiurl,omitempty"`
}

// handleCloudMCPAccountList returns a formatted list of all configured accounts.
func (s *Service) handleCloudMCPAccountList(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Load current TOML configuration
	configPath := config.GetConfigPath()
	tomlConfig, err := config.LoadTOMLConfig(configPath)
	if err != nil {
		// Fall back to current in-memory config
		var result strings.Builder
		result.WriteString(fmt.Sprintf("Current default account: %s\n\n", s.config.DefaultLinodeAccount))
		result.WriteString("Configured accounts (from environment):\n")

		for name, account := range s.config.LinodeAccounts {
			status := ""
			if name == s.config.DefaultLinodeAccount {
				status = " (default)"
			}

			apiURL := account.APIURL
			if apiURL == "" {
				apiURL = "https://api.linode.com/v4 (default)"
			}

			result.WriteString(fmt.Sprintf("• %s: %s%s\n", name, account.Label, status))
			result.WriteString(fmt.Sprintf("  API URL: %s\n", apiURL))
		}

		result.WriteString("\nNote: Configuration is currently loaded from environment variables.\n")
		result.WriteString("Use cloudmcp_account_add to migrate to TOML configuration.\n")

		return mcp.NewToolResultText(result.String()), nil
	}

	// Format TOML configuration
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Current default account: %s\n\n", tomlConfig.System.DefaultAccount))
	result.WriteString("Configured accounts:\n")

	for name, account := range tomlConfig.Accounts {
		status := ""
		if name == tomlConfig.System.DefaultAccount {
			status = " (default)"
		}

		apiURL := account.APIURL
		if apiURL == "" {
			apiURL = "https://api.linode.com/v4 (default)"
		}

		result.WriteString(fmt.Sprintf("• %s: %s%s\n", name, account.Label, status))
		result.WriteString(fmt.Sprintf("  API URL: %s\n", apiURL))
	}

	result.WriteString(fmt.Sprintf("\nConfiguration file: %s\n", configPath))

	return mcp.NewToolResultText(result.String()), nil
}

// handleCloudMCPAccountAdd adds a new account to the TOML configuration.
func (s *Service) handleCloudMCPAccountAdd(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params CloudMCPAccountAddParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return nil, types.NewToolError("cloudmcp", "account_add", "invalid parameters", err) //nolint:wrapcheck // types.NewToolError already wraps the error
	}

	// Validate required parameters
	if params.Name == "" {
		return nil, types.NewToolError("cloudmcp", "account_add", "account name is required", nil) //nolint:wrapcheck // types.NewToolError already wraps the error
	}
	if params.Token == "" {
		return nil, types.NewToolError("cloudmcp", "account_add", "account token is required", nil) //nolint:wrapcheck // types.NewToolError already wraps the error
	}
	if params.Label == "" {
		return nil, types.NewToolError("cloudmcp", "account_add", "account label is required", nil) //nolint:wrapcheck // types.NewToolError already wraps the error
	}

	// Validate token by testing API connection
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: params.Token})
	oauth2Client := oauth2.NewClient(ctx, tokenSource)
	client := linodego.NewClient(oauth2Client)

	if params.APIURL != "" {
		client.SetBaseURL(params.APIURL)
	}

	// Test the connection
	if _, err := client.GetProfile(ctx); err != nil {
		return nil, types.NewToolError("cloudmcp", "account_add", "invalid token or API URL", err) //nolint:wrapcheck // types.NewToolError already wraps the error
	}

	// Load or create TOML configuration
	configPath := config.GetConfigPath()
	configManager := config.NewTOMLConfigManager(configPath)

	if err := configManager.LoadOrCreate(); err != nil {
		return nil, types.NewToolError("cloudmcp", "account_add", "failed to load configuration", err) //nolint:wrapcheck // types.NewToolError already wraps the error
	}

	// Add the account
	accountConfig := config.AccountConfig{
		Token:  params.Token,
		Label:  params.Label,
		APIURL: params.APIURL,
	}

	if err := configManager.AddAccount(params.Name, accountConfig); err != nil {
		return nil, types.NewToolError("cloudmcp", "account_add", "failed to add account", err) //nolint:wrapcheck // types.NewToolError already wraps the error
	}

	// Update service configuration in memory for immediate use
	s.config.LinodeAccounts[params.Name] = config.LinodeAccount{
		Token:  params.Token,
		Label:  params.Label,
		APIURL: params.APIURL,
	}

	// Add to account manager
	client.SetUserAgent("CloudMCP/0.1.0")
	s.accountManager.mu.Lock()
	s.accountManager.accounts[params.Name] = &Account{
		Name:   params.Name,
		Label:  params.Label,
		Client: &client,
	}
	s.accountManager.mu.Unlock()

	result := fmt.Sprintf("Account '%s' (%s) added successfully.\n", params.Name, params.Label)
	result += fmt.Sprintf("Configuration saved to: %s\n", configPath)
	result += "The account is now available for use."

	return mcp.NewToolResultText(result), nil
}

// handleCloudMCPAccountRemove removes an account from the TOML configuration.
func (s *Service) handleCloudMCPAccountRemove(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params CloudMCPAccountRemoveParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return nil, types.NewToolError("cloudmcp", "account_remove", "invalid parameters", err) //nolint:wrapcheck // types.NewToolError already wraps the error
	}

	if params.Name == "" {
		return nil, types.NewToolError("cloudmcp", "account_remove", "account name is required", nil) //nolint:wrapcheck // types.NewToolError already wraps the error
	}

	// Load TOML configuration
	configPath := config.GetConfigPath()
	configManager := config.NewTOMLConfigManager(configPath)

	if err := configManager.LoadOrCreate(); err != nil {
		return nil, types.NewToolError("cloudmcp", "account_remove", "failed to load configuration", err) //nolint:wrapcheck // types.NewToolError already wraps the error
	}

	// Prevent removing the default account
	currentConfig := configManager.GetConfig()
	if params.Name == currentConfig.System.DefaultAccount {
		return nil, types.NewToolError("cloudmcp", "account_remove", fmt.Sprintf("cannot remove the default account '%s'. Change the default account first", params.Name), nil) //nolint:wrapcheck // types.NewToolError already wraps the error
	}

	// Remove the account
	if err := configManager.RemoveAccount(params.Name); err != nil {
		return nil, types.NewToolError("cloudmcp", "account_remove", "failed to remove account", err) //nolint:wrapcheck // types.NewToolError already wraps the error
	}

	// Remove from service configuration
	delete(s.config.LinodeAccounts, params.Name)

	// Remove from account manager
	s.accountManager.mu.Lock()
	delete(s.accountManager.accounts, params.Name)
	s.accountManager.mu.Unlock()

	result := fmt.Sprintf("Account '%s' removed successfully.\n", params.Name)
	result += fmt.Sprintf("Configuration saved to: %s", configPath)

	return mcp.NewToolResultText(result), nil
}

// handleCloudMCPAccountUpdate updates an existing account in the TOML configuration.
func (s *Service) handleCloudMCPAccountUpdate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params CloudMCPAccountUpdateParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return nil, types.NewToolError("cloudmcp", "account_update", "invalid parameters", err) //nolint:wrapcheck // types.NewToolError already wraps the error
	}

	if params.Name == "" {
		return nil, types.NewToolError("cloudmcp", "account_update", "account name is required", nil) //nolint:wrapcheck // types.NewToolError already wraps the error
	}

	// Load TOML configuration
	configPath := config.GetConfigPath()
	configManager := config.NewTOMLConfigManager(configPath)

	if err := configManager.LoadOrCreate(); err != nil {
		return nil, types.NewToolError("cloudmcp", "account_update", "failed to load configuration", err) //nolint:wrapcheck // types.NewToolError already wraps the error
	}

	// Get current account configuration
	currentConfig := configManager.GetConfig()
	existingAccount, exists := currentConfig.Accounts[params.Name]
	if !exists {
		return nil, types.NewToolError("cloudmcp", "account_update", fmt.Sprintf("account '%s' does not exist", params.Name), nil) //nolint:wrapcheck // types.NewToolError already wraps the error
	}

	// Build updated account config (preserve existing values if not provided)
	updatedAccount := config.AccountConfig{
		Token:  existingAccount.Token,
		Label:  existingAccount.Label,
		APIURL: existingAccount.APIURL,
	}

	if params.Token != "" {
		updatedAccount.Token = params.Token
	}
	if params.Label != "" {
		updatedAccount.Label = params.Label
	}
	if params.APIURL != "" {
		updatedAccount.APIURL = params.APIURL
	}

	// Validate token if provided
	if params.Token != "" {
		tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: updatedAccount.Token})
		oauth2Client := oauth2.NewClient(ctx, tokenSource)
		client := linodego.NewClient(oauth2Client)

		if updatedAccount.APIURL != "" {
			client.SetBaseURL(updatedAccount.APIURL)
		}

		if _, err := client.GetProfile(ctx); err != nil {
			return nil, types.NewToolError("cloudmcp", "account_update", "invalid token or API URL", err) //nolint:wrapcheck // types.NewToolError already wraps the error
		}
	}

	// Update the account
	if err := configManager.UpdateAccount(params.Name, updatedAccount); err != nil {
		return nil, types.NewToolError("cloudmcp", "account_update", "failed to update account", err) //nolint:wrapcheck // types.NewToolError already wraps the error
	}

	// Update service configuration
	s.config.LinodeAccounts[params.Name] = config.LinodeAccount{
		Token:  updatedAccount.Token,
		Label:  updatedAccount.Label,
		APIURL: updatedAccount.APIURL,
	}

	// Update account manager if token changed
	if params.Token != "" {
		tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: updatedAccount.Token})
		oauth2Client := oauth2.NewClient(ctx, tokenSource)
		client := linodego.NewClient(oauth2Client)
		client.SetUserAgent("CloudMCP/0.1.0")

		if updatedAccount.APIURL != "" {
			client.SetBaseURL(updatedAccount.APIURL)
		}

		s.accountManager.mu.Lock()
		s.accountManager.accounts[params.Name] = &Account{
			Name:   params.Name,
			Label:  updatedAccount.Label,
			Client: &client,
		}
		s.accountManager.mu.Unlock()
	}

	result := fmt.Sprintf("Account '%s' updated successfully.\n", params.Name)
	result += fmt.Sprintf("Configuration saved to: %s", configPath)

	return mcp.NewToolResultText(result), nil
}