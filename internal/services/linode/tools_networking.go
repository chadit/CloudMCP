package linode

import (
	"context"
	"fmt"
	"strings"

	"github.com/linode/linodego"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/chadit/CloudMCP/pkg/types"
)

const (
	ipAssignmentUnassigned = "Unassigned"
)

// handleReservedIPsList lists all reserved IP addresses.
func (s *Service) handleReservedIPsList(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, err
	}

	ips, err := account.Client.ListIPAddresses(ctx, nil)
	if err != nil {
		return nil, types.NewToolError("linode", "reserved_ips_list",
			"failed to list IP addresses", err)
	}

	reservedIPs := make([]ReservedIPSummary, 0, len(ips))

	for _, ipAddress := range ips {
		// Only include reserved IPs (not assigned to instances)
		if ipAddress.LinodeID == 0 {
			summary := ReservedIPSummary{
				Address:    ipAddress.Address,
				Gateway:    ipAddress.Gateway,
				SubnetMask: ipAddress.SubnetMask,
				Prefix:     ipAddress.Prefix,
				Type:       string(ipAddress.Type),
				Public:     ipAddress.Public,
				RDNS:       ipAddress.RDNS,
				LinodeID:   intPtr(ipAddress.LinodeID),
				Region:     ipAddress.Region,
			}
			reservedIPs = append(reservedIPs, summary)
		}
	}

	var stringBuilder strings.Builder

	stringBuilder.WriteString(fmt.Sprintf("Found %d reserved IP addresses:\n\n", len(reservedIPs)))

	for _, ipAddress := range reservedIPs {
		assignment := ipAssignmentUnassigned
		if ipAddress.LinodeID != nil && *ipAddress.LinodeID > 0 {
			assignment = fmt.Sprintf("Assigned to Linode %d", *ipAddress.LinodeID)
		}

		visibility := visibilityPrivate
		if ipAddress.Public {
			visibility = visibilityPublic
		}

		fmt.Fprintf(&stringBuilder, "Address: %s (%s %s)\n", ipAddress.Address, ipAddress.Type, visibility)
		fmt.Fprintf(&stringBuilder, "  Gateway: %s | Prefix: %d\n", ipAddress.Gateway, ipAddress.Prefix)
		fmt.Fprintf(&stringBuilder, "  Region: %s | %s\n", ipAddress.Region, assignment)

		if ipAddress.RDNS != "" {
			fmt.Fprintf(&stringBuilder, "  RDNS: %s\n", ipAddress.RDNS)
		}

		stringBuilder.WriteString("\n")
	}

	if len(reservedIPs) == 0 {
		return mcp.NewToolResultText("No reserved IP addresses found."), nil
	}

	return mcp.NewToolResultText(stringBuilder.String()), nil
}

// handleReservedIPGet gets details of a specific reserved IP address.
func (s *Service) handleReservedIPGet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()

	address, addressExists := arguments["address"].(string)
	if !addressExists || address == "" {
		return mcp.NewToolResultError("address parameter is required"), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, err
	}

	ipAddress, err := account.Client.GetIPAddress(ctx, address)
	if err != nil {
		return nil, types.NewToolError("linode", "reserved_ip_get",
			"failed to get IP address "+address, err)
	}

	detail := ReservedIPDetail{
		Address:    ipAddress.Address,
		Gateway:    ipAddress.Gateway,
		SubnetMask: ipAddress.SubnetMask,
		Prefix:     ipAddress.Prefix,
		Type:       string(ipAddress.Type),
		Public:     ipAddress.Public,
		RDNS:       ipAddress.RDNS,
		LinodeID:   intPtr(ipAddress.LinodeID),
		Region:     ipAddress.Region,
	}

	var stringBuilder strings.Builder

	fmt.Fprintf(&stringBuilder, "IP Address Details:\n")
	fmt.Fprintf(&stringBuilder, "Address: %s\n", detail.Address)
	fmt.Fprintf(&stringBuilder, "Type: %s\n", detail.Type)
	fmt.Fprintf(&stringBuilder, "Gateway: %s\n", detail.Gateway)
	fmt.Fprintf(&stringBuilder, "Subnet Mask: %s\n", detail.SubnetMask)
	fmt.Fprintf(&stringBuilder, "Prefix: %d\n", detail.Prefix)
	fmt.Fprintf(&stringBuilder, "Region: %s\n", detail.Region)

	visibility := visibilityPrivate
	if detail.Public {
		visibility = visibilityPublic
	}

	fmt.Fprintf(&stringBuilder, "Visibility: %s\n", visibility)

	if detail.LinodeID != nil && *detail.LinodeID > 0 {
		fmt.Fprintf(&stringBuilder, "Assigned to Linode: %d\n", *detail.LinodeID)
	} else {
		fmt.Fprintf(&stringBuilder, "Assignment: %s\n", ipAssignmentUnassigned)
	}

	if detail.RDNS != "" {
		fmt.Fprintf(&stringBuilder, "Reverse DNS: %s\n", detail.RDNS)
	}

	return mcp.NewToolResultText(stringBuilder.String()), nil
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

	ipAddress, err := account.Client.AllocateReserveIP(ctx, allocateOpts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to allocate reserved IP: %v", err)), nil
	}

	assignment := ipAssignmentUnassigned
	if ipAddress.LinodeID > 0 {
		assignment = fmt.Sprintf("Assigned to Linode %d", ipAddress.LinodeID)
	}

	visibility := visibilityPrivate
	if ipAddress.Public {
		visibility = visibilityPublic
	}

	return mcp.NewToolResultText(fmt.Sprintf("Reserved IP allocated successfully:\nAddress: %s\nType: %s (%s)\nRegion: %s\nAssignment: %s",
		ipAddress.Address, ipAddress.Type, visibility, ipAddress.Region, assignment)), nil
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

	assignment := ipAssignmentUnassigned
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

	if rdns, rdnsExists := arguments["rdns"].(string); rdnsExists {
		updateOpts.RDNS = &rdns
	}

	ipAddress, err := account.Client.UpdateIPAddress(ctx, address, updateOpts)
	if err != nil {
		return nil, types.NewToolError("linode", "reserved_ip_update",
			"failed to update IP address "+address, err)
	}

	return mcp.NewToolResultText(fmt.Sprintf("IP address updated successfully:\nAddress: %s\nReverse DNS: %s", ipAddress.Address, ipAddress.RDNS)), nil
}

// handleVLANsList lists all VLANs.
func (s *Service) handleVLANsList(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, err
	}

	vlans, err := account.Client.ListVLANs(ctx, nil)
	if err != nil {
		return nil, types.NewToolError("linode", "vlans_list",
			"failed to list VLANs", err)
	}

	summaries := make([]VLANSummary, 0, len(vlans))

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

	var stringBuilder strings.Builder

	stringBuilder.WriteString(fmt.Sprintf("Found %d VLANs:\n\n", len(summaries)))

	for _, vlan := range summaries {
		fmt.Fprintf(&stringBuilder, "Label: %s (%s)\n", vlan.Label, vlan.Region)

		if len(vlan.Linodes) > 0 {
			fmt.Fprintf(&stringBuilder, "  Attached Linodes: %v\n", vlan.Linodes)
		} else {
			fmt.Fprintf(&stringBuilder, "  Attached Linodes: None\n")
		}

		fmt.Fprintf(&stringBuilder, "  Created: %s\n", vlan.Created)
		stringBuilder.WriteString("\n")
	}

	if len(summaries) == 0 {
		return mcp.NewToolResultText("No VLANs found."), nil
	}

	return mcp.NewToolResultText(stringBuilder.String()), nil
}

// handleIPv6PoolsList lists all IPv6 pools.
func (s *Service) handleIPv6PoolsList(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, err
	}

	pools, err := account.Client.ListIPv6Pools(ctx, nil)
	if err != nil {
		return nil, types.NewToolError("linode", "ipv6_pools_list",
			"failed to list IPv6 pools", err)
	}

	summaries := make([]IPv6PoolSummary, 0, len(pools))

	for _, pool := range pools {
		summary := IPv6PoolSummary{
			Range:  pool.Range,
			Region: pool.Region,
		}
		summaries = append(summaries, summary)
	}

	// Remove unused result variable

	var stringBuilder strings.Builder

	stringBuilder.WriteString(fmt.Sprintf("Found %d IPv6 pools:\n\n", len(summaries)))

	for _, pool := range summaries {
		fmt.Fprintf(&stringBuilder, "Range: %s\n", pool.Range)
		fmt.Fprintf(&stringBuilder, "  Region: %s\n", pool.Region)
		stringBuilder.WriteString("\n")
	}

	if len(summaries) == 0 {
		return mcp.NewToolResultText("No IPv6 pools found."), nil
	}

	return mcp.NewToolResultText(stringBuilder.String()), nil
}

// handleIPv6RangesList lists all IPv6 ranges.
func (s *Service) handleIPv6RangesList(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, err
	}

	ranges, err := account.Client.ListIPv6Ranges(ctx, nil)
	if err != nil {
		return nil, types.NewToolError("linode", "ipv6_ranges_list",
			"failed to list IPv6 ranges", err)
	}

	summaries := make([]IPv6RangeSummary, 0, len(ranges))

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

	var stringBuilder strings.Builder

	stringBuilder.WriteString(fmt.Sprintf("Found %d IPv6 ranges:\n\n", len(summaries)))

	for _, ipRange := range summaries {
		fmt.Fprintf(&stringBuilder, "Range: %s/%d\n", ipRange.Range, ipRange.Prefix)
		fmt.Fprintf(&stringBuilder, "  Region: %s\n", ipRange.Region)
		fmt.Fprintf(&stringBuilder, "  Route Target: %s\n", ipRange.RouteTarget)
		stringBuilder.WriteString("\n")
	}

	if len(summaries) == 0 {
		return mcp.NewToolResultText("No IPv6 ranges found."), nil
	}

	return mcp.NewToolResultText(stringBuilder.String()), nil
}

// intPtr function already defined in tools_domains.go
