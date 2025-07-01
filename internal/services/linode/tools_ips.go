package linode

import (
	"context"
	"fmt"
	"net"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/chadit/CloudMCP/pkg/types"
)

const (
	visibilityPublic  = "Public"
	visibilityPrivate = "Private"
)

func (s *Service) handleIPsList(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, err
	}

	// Get all instances to fetch their IPs
	instances, err := account.Client.ListInstances(ctx, nil)
	if err != nil {
		return nil, types.NewToolError("linode", "ips_list",
			"failed to list instances for IPs", err)
	}

	var allIPs []IPInfo

	instanceMap := make(map[int]string) // Map instance ID to label

	// Collect IPs from all instances
	for _, instance := range instances {
		instanceMap[instance.ID] = instance.Label

		// Process IPv4 addresses
		for _, ip := range instance.IPv4 {
			ipInfo := IPInfo{
				Address:  ip.String(),
				Type:     "IPv4",
				Public:   !ip.IsPrivate(),
				LinodeID: instance.ID,
				Region:   instance.Region,
			}
			allIPs = append(allIPs, ipInfo)
		}

		// Process IPv6 address
		if instance.IPv6 != "" {
			ipInfo := IPInfo{
				Address:  instance.IPv6,
				Type:     "IPv6",
				Public:   true, // IPv6 addresses are typically public
				LinodeID: instance.ID,
				Region:   instance.Region,
			}
			allIPs = append(allIPs, ipInfo)
		}
	}

	if len(allIPs) == 0 {
		return mcp.NewToolResultText("No IP addresses found."), nil
	}

	var resultText string
	resultText = fmt.Sprintf("Found %d IP address(es):\n\n", len(allIPs))

	// Group by instance for better readability
	for instanceID, instanceLabel := range instanceMap {
		var instanceIPs []IPInfo

		for _, ip := range allIPs {
			if ip.LinodeID == instanceID {
				instanceIPs = append(instanceIPs, ip)
			}
		}

		if len(instanceIPs) > 0 {
			resultText += fmt.Sprintf("Instance: %s (ID: %d)\n", instanceLabel, instanceID)

			for _, ipInfo := range instanceIPs {
				visibility := visibilityPublic
				if !ipInfo.Public {
					visibility = visibilityPrivate
				}

				resultText += fmt.Sprintf("  - %s (%s, %s) - Region: %s\n",
					ipInfo.Address, ipInfo.Type, visibility, ipInfo.Region)
			}

			resultText += "\n"
		}
	}

	return mcp.NewToolResultText(resultText), nil
}

func (s *Service) handleIPGet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()

	address, ok := arguments["address"].(string)
	if !ok || address == "" {
		return mcp.NewToolResultError("address is required"), nil
	}

	// Validate IP address format
	ip := net.ParseIP(address)
	if ip == nil {
		return mcp.NewToolResultError("invalid IP address format"), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, err
	}

	// Unfortunately, Linode API doesn't have a direct "get IP" endpoint
	// We need to find the IP by searching through instances
	instances, err := account.Client.ListInstances(ctx, nil)
	if err != nil {
		return nil, types.NewToolError("linode", "ip_get",
			"failed to list instances", err)
	}

	// Search for the IP
	for _, instance := range instances {
		// Check IPv4 addresses
		for _, instIP := range instance.IPv4 {
			if instIP.String() == address {
				ipType := "IPv4"

				visibility := visibilityPublic
				if instIP.IsPrivate() {
					visibility = visibilityPrivate
				}

				resultText := fmt.Sprintf(`IP Address Details:
Address: %s
Type: %s
Visibility: %s
Region: %s

Associated Instance:
- ID: %d
- Label: %s
- Status: %s`,
					address,
					ipType,
					visibility,
					instance.Region,
					instance.ID,
					instance.Label,
					instance.Status,
				)

				// Try to get RDNS if it's a public IP
				if visibility == visibilityPublic {
					// Note: Linode API would need specific endpoint for RDNS lookup
					resultText += "\n\nReverse DNS: (Not configured or unavailable)"
				}

				return mcp.NewToolResultText(resultText), nil
			}
		}

		// Check IPv6 address
		if instance.IPv6 == address {
			resultText := fmt.Sprintf(`IP Address Details:
Address: %s
Type: IPv6
Visibility: Public
Region: %s

Associated Instance:
- ID: %d
- Label: %s
- Status: %s

Note: IPv6 addresses are globally routable`,
				address,
				instance.Region,
				instance.ID,
				instance.Label,
				instance.Status,
			)

			return mcp.NewToolResultText(resultText), nil
		}
	}

	return mcp.NewToolResultError(fmt.Sprintf("IP address %s not found in any Linode instance", address)), nil
}
