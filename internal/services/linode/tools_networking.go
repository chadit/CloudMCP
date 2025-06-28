package linode

import (
	"context"
	"fmt"
	"strings"

	"github.com/linode/linodego"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/chadit/CloudMCP/pkg/types"
)

// handleReservedIPsList lists all reserved IP addresses.
func (s *Service) handleReservedIPsList(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, err
	}

	ips, err := account.Client.ListIPAddresses(ctx, nil)
	if err != nil {
		return nil, types.NewToolError("linode", "reserved_ips_list", //nolint:wrapcheck // types.NewToolError already wraps the error
			"failed to list IP addresses", err)
	}

	var reservedIPs []ReservedIPSummary
	for _, ip := range ips {
		// Only include reserved IPs (not assigned to instances)
		if ip.LinodeID == 0 {
			summary := ReservedIPSummary{
				Address:    ip.Address,
				Gateway:    ip.Gateway,
				SubnetMask: ip.SubnetMask,
				Prefix:     ip.Prefix,
				Type:       string(ip.Type),
				Public:     ip.Public,
				RDNS:       ip.RDNS,
				LinodeID:   intPtr(ip.LinodeID),
				Region:     ip.Region,
			}
			reservedIPs = append(reservedIPs, summary)
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d reserved IP addresses:\n\n", len(reservedIPs)))

	for _, ip := range reservedIPs {
		assignment := "Unassigned"
		if ip.LinodeID != nil && *ip.LinodeID > 0 {
			assignment = fmt.Sprintf("Assigned to Linode %d", *ip.LinodeID)
		}

		visibility := "Private"
		if ip.Public {
			visibility = "Public"
		}

		fmt.Fprintf(&sb, "Address: %s (%s %s)\n", ip.Address, ip.Type, visibility)
		fmt.Fprintf(&sb, "  Gateway: %s | Prefix: %d\n", ip.Gateway, ip.Prefix)
		fmt.Fprintf(&sb, "  Region: %s | %s\n", ip.Region, assignment)
		if ip.RDNS != "" {
			fmt.Fprintf(&sb, "  RDNS: %s\n", ip.RDNS)
		}
		sb.WriteString("\n")
	}

	if len(reservedIPs) == 0 {
		return mcp.NewToolResultText("No reserved IP addresses found."), nil
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// handleReservedIPGet gets details of a specific reserved IP address.
func (s *Service) handleReservedIPGet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()
	address, ok := arguments["address"].(string)
	if !ok || address == "" {
		return mcp.NewToolResultError("address parameter is required"), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, err
	}

	ip, err := account.Client.GetIPAddress(ctx, address)
	if err != nil {
		return nil, types.NewToolError("linode", "reserved_ip_get", //nolint:wrapcheck // types.NewToolError already wraps the error
			fmt.Sprintf("failed to get IP address %s", address), err)
	}

	detail := ReservedIPDetail{
		Address:    ip.Address,
		Gateway:    ip.Gateway,
		SubnetMask: ip.SubnetMask,
		Prefix:     ip.Prefix,
		Type:       string(ip.Type),
		Public:     ip.Public,
		RDNS:       ip.RDNS,
		LinodeID:   intPtr(ip.LinodeID),
		Region:     ip.Region,
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "IP Address Details:\n")
	fmt.Fprintf(&sb, "Address: %s\n", detail.Address)
	fmt.Fprintf(&sb, "Type: %s\n", detail.Type)
	fmt.Fprintf(&sb, "Gateway: %s\n", detail.Gateway)
	fmt.Fprintf(&sb, "Subnet Mask: %s\n", detail.SubnetMask)
	fmt.Fprintf(&sb, "Prefix: %d\n", detail.Prefix)
	fmt.Fprintf(&sb, "Region: %s\n", detail.Region)

	visibility := "Private"
	if detail.Public {
		visibility = "Public"
	}
	fmt.Fprintf(&sb, "Visibility: %s\n", visibility)

	if detail.LinodeID != nil && *detail.LinodeID > 0 {
		fmt.Fprintf(&sb, "Assigned to Linode: %d\n", *detail.LinodeID)
	} else {
		fmt.Fprintf(&sb, "Assignment: Unassigned\n")
	}

	if detail.RDNS != "" {
		fmt.Fprintf(&sb, "Reverse DNS: %s\n", detail.RDNS)
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// handleReservedIPAllocate allocates a new reserved IP address.
func (s *Service) handleReservedIPAllocate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params ReservedIPAllocateParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	allocateOpts := linodego.AllocateReserveIPOptions{
		Type:     params.Type,
		Public:   params.Public,
		Reserved: true, // Always reserved for this handler
	}

	if params.Region != "" {
		allocateOpts.Region = params.Region
	}
	if params.LinodeID != 0 {
		allocateOpts.LinodeID = params.LinodeID
	}

	ip, err := account.Client.AllocateReserveIP(ctx, allocateOpts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to allocate reserved IP: %v", err)), nil
	}

	assignment := "Unassigned"
	if ip.LinodeID > 0 {
		assignment = fmt.Sprintf("Assigned to Linode %d", ip.LinodeID)
	}

	visibility := "Private"
	if ip.Public {
		visibility = "Public"
	}

	return mcp.NewToolResultText(fmt.Sprintf("Reserved IP allocated successfully:\nAddress: %s\nType: %s (%s)\nRegion: %s\nAssignment: %s",
		ip.Address, ip.Type, visibility, ip.Region, assignment)), nil
}

// handleReservedIPAssign assigns a reserved IP to a Linode or unassigns it.
func (s *Service) handleReservedIPAssign(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params ReservedIPAssignParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Use the IP assignment API to assign/unassign the IP
	assignOpts := linodego.LinodesAssignIPsOptions{
		Region: params.Region,
		Assignments: []linodego.LinodeIPAssignment{
			{
				Address:  params.Address,
				LinodeID: params.LinodeID,
			},
		},
	}

	err = account.Client.InstancesAssignIPs(ctx, assignOpts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to assign IP address: %v", err)), nil
	}

	assignment := "Unassigned"
	if params.LinodeID > 0 {
		assignment = fmt.Sprintf("Assigned to Linode %d", params.LinodeID)
	}

	return mcp.NewToolResultText(fmt.Sprintf("IP address assignment updated:\nAddress: %s\nAssignment: %s", params.Address, assignment)), nil
}

// handleReservedIPUpdate updates the reverse DNS for a reserved IP.
func (s *Service) handleReservedIPUpdate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()

	address, addressPresent := arguments["address"].(string)
	if !addressPresent || address == "" {
		return mcp.NewToolResultError("address parameter is required"), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, err
	}

	updateOpts := linodego.IPAddressUpdateOptions{}

	if rdns, ok := arguments["rdns"].(string); ok {
		updateOpts.RDNS = &rdns
	}

	ip, err := account.Client.UpdateIPAddress(ctx, address, updateOpts)
	if err != nil {
		return nil, types.NewToolError("linode", "reserved_ip_update", //nolint:wrapcheck // types.NewToolError already wraps the error
			fmt.Sprintf("failed to update IP address %s", address), err)
	}

	return mcp.NewToolResultText(fmt.Sprintf("IP address updated successfully:\nAddress: %s\nReverse DNS: %s", ip.Address, ip.RDNS)), nil
}

// handleVLANsList lists all VLANs.
func (s *Service) handleVLANsList(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, err
	}

	vlans, err := account.Client.ListVLANs(ctx, nil)
	if err != nil {
		return nil, types.NewToolError("linode", "vlans_list", //nolint:wrapcheck // types.NewToolError already wraps the error
			"failed to list VLANs", err)
	}

	var summaries []VLANSummary
	for _, vlan := range vlans {
		summary := VLANSummary{
			Label:   vlan.Label,
			Linodes: vlan.Linodes,
			Region:  vlan.Region,
			Created: vlan.Created.Format("2006-01-02T15:04:05"),
		}
		summaries = append(summaries, summary)
	}

	// Remove unused result variable

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d VLANs:\n\n", len(summaries)))

	for _, vlan := range summaries {
		fmt.Fprintf(&sb, "Label: %s (%s)\n", vlan.Label, vlan.Region)
		if len(vlan.Linodes) > 0 {
			fmt.Fprintf(&sb, "  Attached Linodes: %v\n", vlan.Linodes)
		} else {
			fmt.Fprintf(&sb, "  Attached Linodes: None\n")
		}
		fmt.Fprintf(&sb, "  Created: %s\n", vlan.Created)
		sb.WriteString("\n")
	}

	if len(summaries) == 0 {
		return mcp.NewToolResultText("No VLANs found."), nil
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// handleIPv6PoolsList lists all IPv6 pools.
func (s *Service) handleIPv6PoolsList(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, err
	}

	pools, err := account.Client.ListIPv6Pools(ctx, nil)
	if err != nil {
		return nil, types.NewToolError("linode", "ipv6_pools_list", //nolint:wrapcheck // types.NewToolError already wraps the error
			"failed to list IPv6 pools", err)
	}

	var summaries []IPv6PoolSummary
	for _, pool := range pools {
		summary := IPv6PoolSummary{
			Range:  pool.Range,
			Region: pool.Region,
		}
		summaries = append(summaries, summary)
	}

	// Remove unused result variable

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d IPv6 pools:\n\n", len(summaries)))

	for _, pool := range summaries {
		fmt.Fprintf(&sb, "Range: %s\n", pool.Range)
		fmt.Fprintf(&sb, "  Region: %s\n", pool.Region)
		sb.WriteString("\n")
	}

	if len(summaries) == 0 {
		return mcp.NewToolResultText("No IPv6 pools found."), nil
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// handleIPv6RangesList lists all IPv6 ranges.
func (s *Service) handleIPv6RangesList(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, err
	}

	ranges, err := account.Client.ListIPv6Ranges(ctx, nil)
	if err != nil {
		return nil, types.NewToolError("linode", "ipv6_ranges_list", //nolint:wrapcheck // types.NewToolError already wraps the error
			"failed to list IPv6 ranges", err)
	}

	var summaries []IPv6RangeSummary
	for _, ipRange := range ranges {
		summary := IPv6RangeSummary{
			Range:       ipRange.Range,
			Region:      ipRange.Region,
			Prefix:      ipRange.Prefix,
			RouteTarget: ipRange.RouteTarget,
		}
		summaries = append(summaries, summary)
	}

	// Remove unused result variable

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d IPv6 ranges:\n\n", len(summaries)))

	for _, ipRange := range summaries {
		fmt.Fprintf(&sb, "Range: %s/%d\n", ipRange.Range, ipRange.Prefix)
		fmt.Fprintf(&sb, "  Region: %s\n", ipRange.Region)
		fmt.Fprintf(&sb, "  Route Target: %s\n", ipRange.RouteTarget)
		sb.WriteString("\n")
	}

	if len(summaries) == 0 {
		return mcp.NewToolResultText("No IPv6 ranges found."), nil
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// intPtr function already defined in tools_domains.go
