package linode

import (
	"context"
	"fmt"
	"strings"

	"github.com/linode/linodego"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/chadit/CloudMCP/pkg/types"
)

// handleLongviewClientsList lists all Longview clients.
func (s *Service) handleLongviewClientsList(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, err
	}

	clients, err := account.Client.ListLongviewClients(ctx, nil)
	if err != nil {
		return nil, types.NewToolError("linode", "longview_clients_list",
			"failed to list Longview clients", err)
	}

	summaries := make([]LongviewClientSummary, 0, len(clients))

	for _, client := range clients {
		summary := LongviewClientSummary{
			ID:      client.ID,
			Label:   client.Label,
			APIKey:  "***REDACTED***", // Don't expose API keys in listings
			Created: client.Created.Format("2006-01-02T15:04:05"),
			Updated: client.Updated.Format("2006-01-02T15:04:05"),
		}
		summaries = append(summaries, summary)
	}

	var stringBuilder strings.Builder

	stringBuilder.WriteString(fmt.Sprintf("Found %d Longview clients:\n\n", len(summaries)))

	for _, client := range summaries {
		fmt.Fprintf(&stringBuilder, "ID: %d | %s\n", client.ID, client.Label)
		fmt.Fprintf(&stringBuilder, "  Created: %s\n", client.Created)
		fmt.Fprintf(&stringBuilder, "  Updated: %s\n", client.Updated)
		stringBuilder.WriteString("\n")
	}

	if len(summaries) == 0 {
		return mcp.NewToolResultText("No Longview clients found."), nil
	}

	return mcp.NewToolResultText(stringBuilder.String()), nil
}

// handleLongviewClientGet gets details of a specific Longview client.
func (s *Service) handleLongviewClientGet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()

	clientID, err := parseIDFromArguments(arguments, "client_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, err
	}

	client, err := account.Client.GetLongviewClient(ctx, clientID)
	if err != nil {
		return nil, types.NewToolError("linode", "longview_client_get",
			fmt.Sprintf("failed to get Longview client %d", clientID), err)
	}

	detail := LongviewClientDetail{
		ID:      client.ID,
		Label:   client.Label,
		APIKey:  client.APIKey, // Show API key in detail view
		Created: client.Created.Format("2006-01-02T15:04:05"),
		Updated: client.Updated.Format("2006-01-02T15:04:05"),
		Apps:    map[string]interface{}{}, // Apps field needs conversion
	}

	var stringBuilder strings.Builder

	fmt.Fprintf(&stringBuilder, "Longview Client Details:\n")
	fmt.Fprintf(&stringBuilder, "ID: %d\n", detail.ID)
	fmt.Fprintf(&stringBuilder, "Label: %s\n", detail.Label)
	fmt.Fprintf(&stringBuilder, "API Key: %s\n", detail.APIKey)
	fmt.Fprintf(&stringBuilder, "Created: %s\n", detail.Created)
	fmt.Fprintf(&stringBuilder, "Updated: %s\n", detail.Updated)

	if len(detail.Apps) > 0 {
		fmt.Fprintf(&stringBuilder, "\nMonitored Applications:\n")

		for app := range detail.Apps {
			fmt.Fprintf(&stringBuilder, "  - %s\n", app)
		}
	}

	fmt.Fprintf(&stringBuilder, "\nInstallation Instructions:\n")
	fmt.Fprintf(&stringBuilder, "1. Install the Longview client on your server\n")
	fmt.Fprintf(&stringBuilder, "2. Configure the API key: %s\n", detail.APIKey)
	fmt.Fprintf(&stringBuilder, "3. Monitor your system metrics in the Linode Cloud Manager\n")

	return mcp.NewToolResultText(stringBuilder.String()), nil
}

// handleLongviewClientCreate creates a new Longview client.
func (s *Service) handleLongviewClientCreate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()

	label, labelPresent := arguments["label"].(string)
	if !labelPresent || label == "" {
		return mcp.NewToolResultError("label parameter is required"), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, err
	}

	createOpts := linodego.LongviewClientCreateOptions{
		Label: label,
	}

	client, err := account.Client.CreateLongviewClient(ctx, createOpts)
	if err != nil {
		return nil, types.NewToolError("linode", "longview_client_create",
			"failed to create Longview client", err)
	}

	return mcp.NewToolResultText(fmt.Sprintf("Longview client created successfully:\nID: %d\nLabel: %s\nAPI Key: %s\n\nUse this API key to configure monitoring on your server.",
		client.ID, client.Label, client.APIKey)), nil
}

// handleLongviewClientUpdate updates an existing Longview client.
func (s *Service) handleLongviewClientUpdate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()

	clientID, err := parseIDFromArguments(arguments, "client_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, err
	}

	updateOpts := linodego.LongviewClientUpdateOptions{}

	if label, labelExists := arguments["label"].(string); labelExists && label != "" {
		updateOpts.Label = label
	}

	client, err := account.Client.UpdateLongviewClient(ctx, clientID, updateOpts)
	if err != nil {
		return nil, types.NewToolError("linode", "longview_client_update",
			fmt.Sprintf("failed to update Longview client %d", clientID), err)
	}

	return mcp.NewToolResultText(fmt.Sprintf("Longview client updated successfully:\nID: %d\nLabel: %s", client.ID, client.Label)), nil
}

// handleLongviewClientDelete deletes a Longview client.
func (s *Service) handleLongviewClientDelete(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()

	clientID, err := parseIDFromArguments(arguments, "client_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, err
	}

	err = account.Client.DeleteLongviewClient(ctx, clientID)
	if err != nil {
		return nil, types.NewToolError("linode", "longview_client_delete",
			fmt.Sprintf("failed to delete Longview client %d", clientID), err)
	}

	return mcp.NewToolResultText(fmt.Sprintf("Longview client %d deleted successfully", clientID)), nil
}
