package linode

import (
	"context"
	"fmt"
	"strings"

	"github.com/linode/linodego"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	// Maximum description length for display truncation.
	maxDescriptionLength  = 100
	maxDescriptionDisplay = 97 // Length before adding "..."
)

// handleStackScriptsList lists all StackScripts.
func (s *Service) handleStackScriptsList(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	stackscripts, err := account.Client.ListStackscripts(ctx, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list StackScripts: %v", err)), nil
	}

	summaries := make([]StackScriptSummary, 0, len(stackscripts))
	for _, stackScript := range stackscripts {
		summary := StackScriptSummary{
			ID:                stackScript.ID,
			Username:          stackScript.Username,
			Label:             stackScript.Label,
			Description:       stackScript.Description,
			IsPublic:          stackScript.IsPublic,
			Images:            stackScript.Images,
			DeploymentsTotal:  stackScript.DeploymentsTotal,
			DeploymentsActive: stackScript.DeploymentsActive,
			UserGravatarID:    stackScript.UserGravatarID,
			Created:           stackScript.Created.Format("2006-01-02T15:04:05"),
			Updated:           stackScript.Updated.Format("2006-01-02T15:04:05"),
		}
		summaries = append(summaries, summary)
	}

	// Remove unused result variable

	var stringBuilder strings.Builder

	stringBuilder.WriteString(fmt.Sprintf("Found %d StackScripts:\n\n", len(summaries)))

	for _, stackScript := range summaries {
		visibility := visibilityPrivate
		if stackScript.IsPublic {
			visibility = visibilityPublic
		}

		fmt.Fprintf(&stringBuilder, "ID: %d | %s (%s)\n", stackScript.ID, stackScript.Label, visibility)
		fmt.Fprintf(&stringBuilder, "  Author: %s\n", stackScript.Username)

		if stackScript.Description != "" {
			description := stackScript.Description
			if len(description) > maxDescriptionLength {
				description = description[:maxDescriptionDisplay] + "..."
			}

			fmt.Fprintf(&stringBuilder, "  Description: %s\n", description)
		}

		fmt.Fprintf(&stringBuilder, "  Compatible Images: %s\n", strings.Join(stackScript.Images, ", "))
		fmt.Fprintf(&stringBuilder, "  Deployments: %d total, %d active\n", stackScript.DeploymentsTotal, stackScript.DeploymentsActive)
		fmt.Fprintf(&stringBuilder, "  Updated: %s\n", stackScript.Updated)
		stringBuilder.WriteString("\n")
	}

	return mcp.NewToolResultText(stringBuilder.String()), nil
}

// handleStackScriptGet gets details of a specific StackScript.
func (s *Service) handleStackScriptGet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params StackScriptGetParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	stackScript, err := account.Client.GetStackscript(ctx, params.StackScriptID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get StackScript: %v", err)), nil
	}

	var udfs []StackScriptUserDefinedField

	if stackScript.UserDefinedFields != nil {
		udfs = make([]StackScriptUserDefinedField, 0, len(*stackScript.UserDefinedFields))
		for _, udf := range *stackScript.UserDefinedFields {
			udfs = append(udfs, StackScriptUserDefinedField{
				Name:    udf.Name,
				Label:   udf.Label,
				Example: udf.Example,
				OneOf:   udf.OneOf,
				ManyOf:  udf.ManyOf,
				Default: udf.Default,
			})
		}
	}

	detail := StackScriptDetail{
		ID:                stackScript.ID,
		Username:          stackScript.Username,
		Label:             stackScript.Label,
		Description:       stackScript.Description,
		Ordinal:           stackScript.Ordinal,
		LogoURL:           stackScript.LogoURL,
		Images:            stackScript.Images,
		DeploymentsTotal:  stackScript.DeploymentsTotal,
		DeploymentsActive: stackScript.DeploymentsActive,
		IsPublic:          stackScript.IsPublic,
		Mine:              stackScript.Mine,
		Created:           stackScript.Created.Format("2006-01-02T15:04:05"),
		Updated:           stackScript.Updated.Format("2006-01-02T15:04:05"),
		RevNote:           stackScript.RevNote,
		Script:            stackScript.Script,
		UserDefinedFields: udfs,
		UserGravatarID:    stackScript.UserGravatarID,
	}

	var stringBuilder strings.Builder

	fmt.Fprintf(&stringBuilder, "StackScript Details:\n")
	fmt.Fprintf(&stringBuilder, "ID: %d\n", detail.ID)
	fmt.Fprintf(&stringBuilder, "Label: %s\n", detail.Label)
	fmt.Fprintf(&stringBuilder, "Author: %s\n", detail.Username)

	visibility := visibilityPrivate
	if detail.IsPublic {
		visibility = visibilityPublic
	}

	ownership := ""
	if detail.Mine {
		ownership = " (Mine)"
	}

	fmt.Fprintf(&stringBuilder, "Visibility: %s%s\n", visibility, ownership)

	if detail.Description != "" {
		fmt.Fprintf(&stringBuilder, "Description: %s\n", detail.Description)
	}

	if detail.RevNote != "" {
		fmt.Fprintf(&stringBuilder, "Revision Note: %s\n", detail.RevNote)
	}

	fmt.Fprintf(&stringBuilder, "Compatible Images: %s\n", strings.Join(detail.Images, ", "))
	fmt.Fprintf(&stringBuilder, "Deployments: %d total, %d active\n", detail.DeploymentsTotal, detail.DeploymentsActive)
	fmt.Fprintf(&stringBuilder, "Created: %s\n", detail.Created)
	fmt.Fprintf(&stringBuilder, "Updated: %s\n\n", detail.Updated)

	if len(detail.UserDefinedFields) > 0 {
		fmt.Fprintf(&stringBuilder, "User-Defined Fields:\n")

		for _, udf := range detail.UserDefinedFields {
			fmt.Fprintf(&stringBuilder, "  - %s (%s)\n", udf.Name, udf.Label)
			if udf.Default != "" {
				fmt.Fprintf(&stringBuilder, "    Default: %s\n", udf.Default)
			}

			if udf.Example != "" {
				fmt.Fprintf(&stringBuilder, "    Example: %s\n", udf.Example)
			}

			if udf.OneOf != "" {
				fmt.Fprintf(&stringBuilder, "    Options: %s\n", udf.OneOf)
			}

			if udf.ManyOf != "" {
				fmt.Fprintf(&stringBuilder, "    Multiple Options: %s\n", udf.ManyOf)
			}
		}

		stringBuilder.WriteString("\n")
	}

	fmt.Fprintf(&stringBuilder, "Script Content:\n")
	fmt.Fprintf(&stringBuilder, "```bash\n%s\n```\n", detail.Script)

	return mcp.NewToolResultText(stringBuilder.String()), nil
}

// handleStackScriptCreate creates a new StackScript.
func (s *Service) handleStackScriptCreate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params StackScriptCreateParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	createOpts := linodego.StackscriptCreateOptions{
		Label:       params.Label,
		Description: params.Description,
		Images:      params.Images,
		Script:      params.Script,
		IsPublic:    params.IsPublic,
		RevNote:     params.RevNote,
	}

	stackScript, err := account.Client.CreateStackscript(ctx, createOpts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create StackScript: %v", err)), nil
	}

	visibility := visibilityPrivate
	if stackScript.IsPublic {
		visibility = visibilityPublic
	}

	return mcp.NewToolResultText(fmt.Sprintf("StackScript created successfully:\nID: %d\nLabel: %s\nVisibility: %s\nCompatible Images: %s",
		stackScript.ID, stackScript.Label, visibility, strings.Join(stackScript.Images, ", "))), nil
}

// handleStackScriptUpdate updates an existing StackScript.
func (s *Service) handleStackScriptUpdate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params StackScriptUpdateParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	updateOpts := linodego.StackscriptUpdateOptions{}

	if params.Label != "" {
		updateOpts.Label = params.Label
	}

	if params.Description != "" {
		updateOpts.Description = params.Description
	}

	if len(params.Images) > 0 {
		updateOpts.Images = params.Images
	}

	if params.Script != "" {
		updateOpts.Script = params.Script
	}

	if params.IsPublic {
		updateOpts.IsPublic = params.IsPublic
	}

	if params.RevNote != "" {
		updateOpts.RevNote = params.RevNote
	}

	stackScript, err := account.Client.UpdateStackscript(ctx, params.StackScriptID, updateOpts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update StackScript: %v", err)), nil
	}

	visibility := visibilityPrivate
	if stackScript.IsPublic {
		visibility = visibilityPublic
	}

	return mcp.NewToolResultText(fmt.Sprintf("StackScript updated successfully:\nID: %d\nLabel: %s\nVisibility: %s\nCompatible Images: %s",
		stackScript.ID, stackScript.Label, visibility, strings.Join(stackScript.Images, ", "))), nil
}

// handleStackScriptDelete deletes a StackScript.
func (s *Service) handleStackScriptDelete(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params StackScriptDeleteParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	err = account.Client.DeleteStackscript(ctx, params.StackScriptID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete StackScript: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("StackScript %d deleted successfully", params.StackScriptID)), nil
}
