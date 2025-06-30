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
	maxDescriptionLength = 100
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

	var summaries []StackScriptSummary
	for _, ss := range stackscripts {
		summary := StackScriptSummary{
			ID:                ss.ID,
			Username:          ss.Username,
			Label:             ss.Label,
			Description:       ss.Description,
			IsPublic:          ss.IsPublic,
			Images:            ss.Images,
			DeploymentsTotal:  ss.DeploymentsTotal,
			DeploymentsActive: ss.DeploymentsActive,
			UserGravatarID:    ss.UserGravatarID,
			Created:           ss.Created.Format("2006-01-02T15:04:05"),
			Updated:           ss.Updated.Format("2006-01-02T15:04:05"),
		}
		summaries = append(summaries, summary)
	}

	// Remove unused result variable

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Found %d StackScripts:\n\n", len(summaries)))

	for _, ss := range summaries {
		visibility := "Private"
		if ss.IsPublic {
			visibility = "Public"
		}

		fmt.Fprintf(&sb, "ID: %d | %s (%s)\n", ss.ID, ss.Label, visibility)
		fmt.Fprintf(&sb, "  Author: %s\n", ss.Username)

		if ss.Description != "" {
			description := ss.Description
			if len(description) > maxDescriptionLength {
				description = description[:maxDescriptionDisplay] + "..."
			}

			fmt.Fprintf(&sb, "  Description: %s\n", description)
		}

		fmt.Fprintf(&sb, "  Compatible Images: %s\n", strings.Join(ss.Images, ", "))
		fmt.Fprintf(&sb, "  Deployments: %d total, %d active\n", ss.DeploymentsTotal, ss.DeploymentsActive)
		fmt.Fprintf(&sb, "  Updated: %s\n", ss.Updated)
		sb.WriteString("\n")
	}

	return mcp.NewToolResultText(sb.String()), nil
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

	ss, err := account.Client.GetStackscript(ctx, params.StackScriptID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get StackScript: %v", err)), nil
	}

	var udfs []StackScriptUserDefinedField

	if ss.UserDefinedFields != nil {
		for _, udf := range *ss.UserDefinedFields {
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
		ID:                ss.ID,
		Username:          ss.Username,
		Label:             ss.Label,
		Description:       ss.Description,
		Ordinal:           ss.Ordinal,
		LogoURL:           ss.LogoURL,
		Images:            ss.Images,
		DeploymentsTotal:  ss.DeploymentsTotal,
		DeploymentsActive: ss.DeploymentsActive,
		IsPublic:          ss.IsPublic,
		Mine:              ss.Mine,
		Created:           ss.Created.Format("2006-01-02T15:04:05"),
		Updated:           ss.Updated.Format("2006-01-02T15:04:05"),
		RevNote:           ss.RevNote,
		Script:            ss.Script,
		UserDefinedFields: udfs,
		UserGravatarID:    ss.UserGravatarID,
	}

	var sb strings.Builder

	fmt.Fprintf(&sb, "StackScript Details:\n")
	fmt.Fprintf(&sb, "ID: %d\n", detail.ID)
	fmt.Fprintf(&sb, "Label: %s\n", detail.Label)
	fmt.Fprintf(&sb, "Author: %s\n", detail.Username)

	visibility := "Private"
	if detail.IsPublic {
		visibility = "Public"
	}

	ownership := ""
	if detail.Mine {
		ownership = " (Mine)"
	}

	fmt.Fprintf(&sb, "Visibility: %s%s\n", visibility, ownership)

	if detail.Description != "" {
		fmt.Fprintf(&sb, "Description: %s\n", detail.Description)
	}

	if detail.RevNote != "" {
		fmt.Fprintf(&sb, "Revision Note: %s\n", detail.RevNote)
	}

	fmt.Fprintf(&sb, "Compatible Images: %s\n", strings.Join(detail.Images, ", "))
	fmt.Fprintf(&sb, "Deployments: %d total, %d active\n", detail.DeploymentsTotal, detail.DeploymentsActive)
	fmt.Fprintf(&sb, "Created: %s\n", detail.Created)
	fmt.Fprintf(&sb, "Updated: %s\n\n", detail.Updated)

	if len(detail.UserDefinedFields) > 0 {
		fmt.Fprintf(&sb, "User-Defined Fields:\n")

		for _, udf := range detail.UserDefinedFields {
			fmt.Fprintf(&sb, "  - %s (%s)\n", udf.Name, udf.Label)
			if udf.Default != "" {
				fmt.Fprintf(&sb, "    Default: %s\n", udf.Default)
			}

			if udf.Example != "" {
				fmt.Fprintf(&sb, "    Example: %s\n", udf.Example)
			}

			if udf.OneOf != "" {
				fmt.Fprintf(&sb, "    Options: %s\n", udf.OneOf)
			}

			if udf.ManyOf != "" {
				fmt.Fprintf(&sb, "    Multiple Options: %s\n", udf.ManyOf)
			}
		}

		sb.WriteString("\n")
	}

	fmt.Fprintf(&sb, "Script Content:\n")
	fmt.Fprintf(&sb, "```bash\n%s\n```\n", detail.Script)

	return mcp.NewToolResultText(sb.String()), nil
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

	ss, err := account.Client.CreateStackscript(ctx, createOpts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create StackScript: %v", err)), nil
	}

	visibility := "Private"
	if ss.IsPublic {
		visibility = "Public"
	}

	return mcp.NewToolResultText(fmt.Sprintf("StackScript created successfully:\nID: %d\nLabel: %s\nVisibility: %s\nCompatible Images: %s",
		ss.ID, ss.Label, visibility, strings.Join(ss.Images, ", "))), nil
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

	ss, err := account.Client.UpdateStackscript(ctx, params.StackScriptID, updateOpts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update StackScript: %v", err)), nil
	}

	visibility := "Private"
	if ss.IsPublic {
		visibility = "Public"
	}

	return mcp.NewToolResultText(fmt.Sprintf("StackScript updated successfully:\nID: %d\nLabel: %s\nVisibility: %s\nCompatible Images: %s",
		ss.ID, ss.Label, visibility, strings.Join(ss.Images, ", "))), nil
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
