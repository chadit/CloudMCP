package linode

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/chadit/CloudMCP/pkg/types"
)

func (s *Service) handleInstancesList(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, err
	}

	instances, err := account.Client.ListInstances(ctx, nil)
	if err != nil {
		return nil, types.NewToolError("linode", "instances_list",
			"failed to list instances", err)
	}

	summaries := make([]InstanceSummary, 0, len(instances))
	for _, instance := range instances {
		ipv4Addresses := make([]string, 0)
		for _, ip := range instance.IPv4 {
			ipv4Addresses = append(ipv4Addresses, ip.String())
		}

		summary := InstanceSummary{
			ID:      instance.ID,
			Label:   instance.Label,
			Status:  string(instance.Status),
			Region:  instance.Region,
			Type:    instance.Type,
			IPv4:    ipv4Addresses,
			IPv6:    instance.IPv6,
			Created: instance.Created.Format("2006-01-02T15:04:05Z"),
			Updated: instance.Updated.Format("2006-01-02T15:04:05Z"),
		}
		summaries = append(summaries, summary)
	}

	if len(summaries) == 0 {
		return mcp.NewToolResultText("No Linode instances found."), nil
	}

	var resultText string
	resultText = fmt.Sprintf("Found %d Linode instance(s):\n\n", len(summaries))
	for _, inst := range summaries {
		resultText += fmt.Sprintf("ID: %d | %s\n", inst.ID, inst.Label)
		resultText += fmt.Sprintf("  Status: %s | Region: %s | Type: %s\n", inst.Status, inst.Region, inst.Type)
		if len(inst.IPv4) > 0 {
			resultText += fmt.Sprintf("  IPv4: %v\n", inst.IPv4)
		}
		resultText += "\n"
	}

	return mcp.NewToolResultText(resultText), nil
}