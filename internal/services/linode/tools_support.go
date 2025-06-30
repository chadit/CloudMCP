package linode

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// handleSupportTicketsList lists all support tickets.
func (s *Service) handleSupportTicketsList(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	tickets, err := account.Client.ListTickets(ctx, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list support tickets: %v", err)), nil
	}

	summaries := make([]SupportTicketSummary, 0, len(tickets))
	for _, ticket := range tickets {
		var entity SupportTicketEntity
		if ticket.Entity != nil {
			entity = SupportTicketEntity{
				ID:    ticket.Entity.ID,
				Label: ticket.Entity.Label,
				Type:  ticket.Entity.Type,
				URL:   ticket.Entity.URL,
			}
		}

		summary := SupportTicketSummary{
			ID:          ticket.ID,
			Summary:     ticket.Summary,
			Description: ticket.Description,
			Status:      string(ticket.Status),
			Entity:      entity,
			OpenedBy:    ticket.OpenedBy,
			UpdatedBy:   ticket.UpdatedBy,
			Closeable:   ticket.Closeable,
		}

		if ticket.Opened != nil {
			summary.Opened = ticket.Opened.Format("2006-01-02T15:04:05")
		}
		if ticket.Updated != nil {
			summary.Updated = ticket.Updated.Format("2006-01-02T15:04:05")
		}

		summaries = append(summaries, summary)
	}

	var stringBuilder strings.Builder

	stringBuilder.WriteString(fmt.Sprintf("Found %d support tickets:\n\n", len(summaries)))

	for _, ticket := range summaries {
		fmt.Fprintf(&stringBuilder, "ID: %d | %s\n", ticket.ID, ticket.Summary)
		fmt.Fprintf(&stringBuilder, "  Status: %s", ticket.Status)
		if ticket.Closeable {
			fmt.Fprintf(&stringBuilder, " (Closeable)")
		}
		fmt.Fprintf(&stringBuilder, "\n")

		if ticket.Entity.Type != "" {
			fmt.Fprintf(&stringBuilder, "  Related: %s %s (ID: %d)\n", ticket.Entity.Type, ticket.Entity.Label, ticket.Entity.ID)
		}

		if ticket.OpenedBy != "" {
			fmt.Fprintf(&stringBuilder, "  Opened by: %s", ticket.OpenedBy)

			if ticket.Opened != "" {
				fmt.Fprintf(&stringBuilder, " on %s", ticket.Opened)
			}
			fmt.Fprintf(&stringBuilder, "\n")
		}

		if ticket.UpdatedBy != "" && ticket.Updated != "" {
			fmt.Fprintf(&stringBuilder, "  Last updated by: %s on %s\n", ticket.UpdatedBy, ticket.Updated)
		}

		stringBuilder.WriteString("\n")
	}

	return mcp.NewToolResultText(stringBuilder.String()), nil
}

// handleSupportTicketGet gets details of a specific support ticket.
func (s *Service) handleSupportTicketGet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params SupportTicketGetParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	ticket, err := account.Client.GetTicket(ctx, params.TicketID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get support ticket: %v", err)), nil
	}

	var entity SupportTicketEntity
	if ticket.Entity != nil {
		entity = SupportTicketEntity{
			ID:    ticket.Entity.ID,
			Label: ticket.Entity.Label,
			Type:  ticket.Entity.Type,
			URL:   ticket.Entity.URL,
		}
	}

	detail := SupportTicketDetail{
		ID:          ticket.ID,
		Summary:     ticket.Summary,
		Description: ticket.Description,
		Status:      string(ticket.Status),
		Entity:      entity,
		OpenedBy:    ticket.OpenedBy,
		UpdatedBy:   ticket.UpdatedBy,
		Closeable:   ticket.Closeable,
		GravatarID:  ticket.GravatarID,
		Attachments: ticket.Attachments,
	}

	if ticket.Opened != nil {
		detail.Opened = ticket.Opened.Format("2006-01-02T15:04:05")
	}

	if ticket.Updated != nil {
		detail.Updated = ticket.Updated.Format("2006-01-02T15:04:05")
	}
	if ticket.Closed != nil {
		detail.ClosedBy = ticket.Closed.Format("2006-01-02T15:04:05")
	}

	var stringBuilder strings.Builder
	fmt.Fprintf(&stringBuilder, "Support Ticket Details:\n")
	fmt.Fprintf(&stringBuilder, "ID: %d\n", detail.ID)
	fmt.Fprintf(&stringBuilder, "Summary: %s\n", detail.Summary)
	fmt.Fprintf(&stringBuilder, "Status: %s", detail.Status)
	if detail.Closeable {
		fmt.Fprintf(&stringBuilder, " (Closeable)")
	}
	fmt.Fprintf(&stringBuilder, "\n\n")

	fmt.Fprintf(&stringBuilder, "Description:\n%s\n\n", detail.Description)

	if detail.Entity.Type != "" {
		fmt.Fprintf(&stringBuilder, "Related Entity:\n")
		fmt.Fprintf(&stringBuilder, "  Type: %s\n", detail.Entity.Type)
		fmt.Fprintf(&stringBuilder, "  Label: %s\n", detail.Entity.Label)
		fmt.Fprintf(&stringBuilder, "  ID: %d\n", detail.Entity.ID)
		if detail.Entity.URL != "" {
			fmt.Fprintf(&stringBuilder, "  URL: %s\n", detail.Entity.URL)
		}
		stringBuilder.WriteString("\n")
	}

	if detail.OpenedBy != "" {
		fmt.Fprintf(&stringBuilder, "Opened by: %s", detail.OpenedBy)

		if detail.Opened != "" {
			fmt.Fprintf(&stringBuilder, " on %s", detail.Opened)
		}

		fmt.Fprintf(&stringBuilder, "\n")
	}

	if detail.UpdatedBy != "" && detail.Updated != "" {
		fmt.Fprintf(&stringBuilder, "Last updated by: %s on %s\n", detail.UpdatedBy, detail.Updated)
	}

	if detail.ClosedBy != "" {
		fmt.Fprintf(&stringBuilder, "Closed on: %s\n", detail.ClosedBy)
	}

	if len(detail.Attachments) > 0 {
		fmt.Fprintf(&stringBuilder, "\nAttachments:\n")

		for _, attachment := range detail.Attachments {
			fmt.Fprintf(&stringBuilder, "  - %s\n", attachment)
		}
	}

	if detail.GravatarID != "" {
		fmt.Fprintf(&stringBuilder, "\nContact Gravatar ID: %s\n", detail.GravatarID)
	}

	// Note about functionality limitations
	fmt.Fprintf(&stringBuilder, "\nNote: This is a read-only view. Ticket creation and replies are not yet supported by the current linodego library version.")

	return mcp.NewToolResultText(stringBuilder.String()), nil
}

// handleSupportTicketCreate creates a new support ticket.
func (s *Service) handleSupportTicketCreate(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params SupportTicketCreateParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	_, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Note: Ticket creation requires manual API implementation since linodego doesn't support it
	// For now, provide information about what would be created
	return mcp.NewToolResultText(fmt.Sprintf("Support ticket would be created with:\nSummary: %s\nDescription: %s\n\nNote: This feature requires custom HTTP implementation that is not yet available in the current linodego library version. The API endpoint exists but needs manual implementation.",
		params.Summary, params.Description)), nil
}

// handleSupportTicketReply creates a reply to a support ticket.
func (s *Service) handleSupportTicketReply(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params SupportTicketReplyParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	_, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Note: Ticket replies require manual API implementation since linodego doesn't support it
	// For now, provide information about what would be created
	return mcp.NewToolResultText(fmt.Sprintf("Support ticket reply would be created for:\nTicket ID: %d\nReply content: %s\n\nNote: This feature requires custom HTTP implementation that is not yet available in the current linodego library version. The API endpoint exists but needs manual implementation.",
		params.TicketID, params.Description)), nil
}
