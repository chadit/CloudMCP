package linode

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/chadit/CloudMCP/pkg/types"
)

func (s *Service) handleAccountGet(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, err
	}

	profile, err := account.Client.GetProfile(ctx)
	if err != nil {
		return nil, types.NewToolError("linode", "account_get", "failed to get profile", err)
	}

	resultText := fmt.Sprintf("Account: %s (%s)\nUsername: %s\nEmail: %s\nUID: %d\nRestricted: %v",
		account.Name, account.Label, profile.Username, profile.Email, profile.UID, profile.Restricted)

	return mcp.NewToolResultText(resultText), nil
}

func (s *Service) handleAccountList(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	accounts := s.accountManager.ListAccounts()
	currentAccount, _ := s.accountManager.GetCurrentAccount()

	accountList := make([]AccountInfo, 0, len(accounts))
	for name, label := range accounts {
		accountList = append(accountList, AccountInfo{
			Name:      name,
			Label:     label,
			IsCurrent: name == currentAccount.Name,
		})
	}

	var resultText string
	resultText = fmt.Sprintf("Current account: %s\n\nConfigured accounts:\n", currentAccount.Name)

	for _, acc := range accountList {
		if acc.IsCurrent {
			resultText += fmt.Sprintf("* %s: %s (current)\n", acc.Name, acc.Label)
		} else {
			resultText += fmt.Sprintf("  %s: %s\n", acc.Name, acc.Label)
		}
	}

	return mcp.NewToolResultText(resultText), nil
}

func (s *Service) handleAccountSwitch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()
	accountName, ok := arguments["account_name"].(string)

	if !ok || accountName == "" {
		return mcp.NewToolResultError("account_name is required"), nil
	}

	if err := s.accountManager.SwitchAccount(accountName); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to switch account: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, err
	}

	profile, err := account.Client.GetProfile(ctx)
	if err != nil {
		return nil, types.NewToolError("linode", "account_switch", "failed to verify switched account", err)
	}

	s.logger.Info("Switched Linode account",
		"to", accountName,
		"username", profile.Username,
	)

	resultText := fmt.Sprintf("Successfully switched to account: %s (%s)\nUsername: %s",
		accountName, account.Label, profile.Username)

	return mcp.NewToolResultText(resultText), nil
}
