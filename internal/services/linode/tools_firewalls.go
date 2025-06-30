package linode

import (
	"context"
	"fmt"
	"strings"

	"github.com/linode/linodego"
	"github.com/mark3labs/mcp-go/mcp"
)

// handleFirewallsList lists all firewalls.
func (s *Service) handleFirewallsList(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	firewalls, err := account.Client.ListFirewalls(ctx, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list firewalls: %v", err)), nil
	}

	summaries := make([]FirewallSummary, 0, len(firewalls))

	for _, firewall := range firewalls {
		// Convert rules
		inboundRules := make([]FirewallRule, 0, len(firewall.Rules.Inbound))

		outboundRules := make([]FirewallRule, 0, len(firewall.Rules.Outbound))

		for _, rule := range firewall.Rules.Inbound {
			inboundRules = append(inboundRules, FirewallRule{
				Ports:       rule.Ports,
				Protocol:    string(rule.Protocol),
				Action:      rule.Action,
				Label:       rule.Label,
				Description: rule.Description,
				Addresses: FirewallAddress{
					IPv4: stringSlicePtrValue(rule.Addresses.IPv4),
					IPv6: stringSlicePtrValue(rule.Addresses.IPv6),
				},
			})
		}

		for _, rule := range firewall.Rules.Outbound {
			outboundRules = append(outboundRules, FirewallRule{
				Ports:       rule.Ports,
				Protocol:    string(rule.Protocol),
				Action:      rule.Action,
				Label:       rule.Label,
				Description: rule.Description,
				Addresses: FirewallAddress{
					IPv4: stringSlicePtrValue(rule.Addresses.IPv4),
					IPv6: stringSlicePtrValue(rule.Addresses.IPv6),
				},
			})
		}

		summary := FirewallSummary{
			ID:     firewall.ID,
			Label:  firewall.Label,
			Status: string(firewall.Status),
			Tags:   firewall.Tags,
			Rules: FirewallRuleSet{
				Inbound:        inboundRules,
				InboundPolicy:  string(firewall.Rules.InboundPolicy),
				Outbound:       outboundRules,
				OutboundPolicy: firewall.Rules.OutboundPolicy,
			},
			Devices: []FirewallDevice{}, // Empty slice - devices need to be fetched separately
			Created: firewall.Created.Format("2006-01-02T15:04:05"),
			Updated: firewall.Updated.Format("2006-01-02T15:04:05"),
		}
		summaries = append(summaries, summary)
	}

	var stringBuilder strings.Builder

	stringBuilder.WriteString(fmt.Sprintf("Found %d firewalls:\n\n", len(summaries)))

	for _, firewall := range summaries {
		var devicesList strings.Builder

		for deviceIndex, device := range firewall.Devices {
			if deviceIndex > 0 {
				devicesList.WriteString(", ")
			}

			fmt.Fprintf(&devicesList, "%s:%d", device.Type, device.ID)
		}

		fmt.Fprintf(&stringBuilder, "ID: %d | %s (%s)\n", firewall.ID, firewall.Label, firewall.Status)
		fmt.Fprintf(&stringBuilder, "  Rules: %d inbound, %d outbound\n", len(firewall.Rules.Inbound), len(firewall.Rules.Outbound))
		fmt.Fprintf(&stringBuilder, "  Devices: %s\n", devicesList.String())
		if len(firewall.Tags) > 0 {
			fmt.Fprintf(&stringBuilder, "  Tags: %s\n", strings.Join(firewall.Tags, ", "))
		}

		stringBuilder.WriteString("\n")
	}

	return mcp.NewToolResultText(stringBuilder.String()), nil
}

// handleFirewallGet gets details of a specific firewall.
func (s *Service) handleFirewallGet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments, argumentsValid := request.Params.Arguments.(map[string]interface{})
	if !argumentsValid {
		return mcp.NewToolResultError("Invalid arguments format"), nil
	}

	firewallID, err := parseIDFromArguments(arguments, "firewall_id")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid firewall_id parameter: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	firewall, err := account.Client.GetFirewall(ctx, firewallID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get firewall: %v", err)), nil
	}

	// Note: Device information needs to be fetched separately via firewall devices API
	devices := []FirewallDevice{}

	// Convert rules
	inboundRules := make([]FirewallRule, 0, len(firewall.Rules.Inbound))

	outboundRules := make([]FirewallRule, 0, len(firewall.Rules.Outbound))

	for _, rule := range firewall.Rules.Inbound {
		inboundRules = append(inboundRules, FirewallRule{
			Ports:       rule.Ports,
			Protocol:    string(rule.Protocol),
			Action:      rule.Action,
			Label:       rule.Label,
			Description: rule.Description,
			Addresses: FirewallAddress{
				IPv4: stringSlicePtrValue(rule.Addresses.IPv4),
				IPv6: stringSlicePtrValue(rule.Addresses.IPv6),
			},
		})
	}

	for _, rule := range firewall.Rules.Outbound {
		outboundRules = append(outboundRules, FirewallRule{
			Ports:       rule.Ports,
			Protocol:    string(rule.Protocol),
			Action:      rule.Action,
			Label:       rule.Label,
			Description: rule.Description,
			Addresses: FirewallAddress{
				IPv4: stringSlicePtrValue(rule.Addresses.IPv4),
				IPv6: stringSlicePtrValue(rule.Addresses.IPv6),
			},
		})
	}

	detail := FirewallDetail{
		ID:     firewall.ID,
		Label:  firewall.Label,
		Status: string(firewall.Status),
		Tags:   firewall.Tags,
		Rules: FirewallRuleSet{
			Inbound:        inboundRules,
			InboundPolicy:  string(firewall.Rules.InboundPolicy),
			Outbound:       outboundRules,
			OutboundPolicy: string(firewall.Rules.OutboundPolicy),
		},
		Devices: devices,
		Created: firewall.Created.Format("2006-01-02T15:04:05"),
		Updated: firewall.Updated.Format("2006-01-02T15:04:05"),
	}

	var stringBuilder strings.Builder

	fmt.Fprintf(&stringBuilder, "Firewall Details:\n")
	fmt.Fprintf(&stringBuilder, "ID: %d\n", detail.ID)
	fmt.Fprintf(&stringBuilder, "Label: %s\n", detail.Label)
	fmt.Fprintf(&stringBuilder, "Status: %s\n", detail.Status)
	fmt.Fprintf(&stringBuilder, "Created: %s\n", detail.Created)
	fmt.Fprintf(&stringBuilder, "Updated: %s\n\n", detail.Updated)

	if len(detail.Tags) > 0 {
		fmt.Fprintf(&stringBuilder, "Tags: %s\n\n", strings.Join(detail.Tags, ", "))
	}

	fmt.Fprintf(&stringBuilder, "Inbound Rules (Policy: %s):\n", detail.Rules.InboundPolicy)
	for ruleIndex, rule := range detail.Rules.Inbound {
		fmt.Fprintf(&stringBuilder, "  %d. %s %s:%s -> %s\n", ruleIndex+1, rule.Action, rule.Protocol, rule.Ports, strings.Join(rule.Addresses.IPv4, ", "))
		if rule.Label != "" {
			fmt.Fprintf(&stringBuilder, "     Label: %s\n", rule.Label)
		}
		if rule.Description != "" {
			fmt.Fprintf(&stringBuilder, "     Description: %s\n", rule.Description)
		}
	}

	fmt.Fprintf(&stringBuilder, "\nOutbound Rules (Policy: %s):\n", detail.Rules.OutboundPolicy)
	for ruleIndex, rule := range detail.Rules.Outbound {
		fmt.Fprintf(&stringBuilder, "  %d. %s %s:%s -> %s\n", ruleIndex+1, rule.Action, rule.Protocol, rule.Ports, strings.Join(rule.Addresses.IPv4, ", "))
		if rule.Label != "" {
			fmt.Fprintf(&stringBuilder, "     Label: %s\n", rule.Label)
		}
		if rule.Description != "" {
			fmt.Fprintf(&stringBuilder, "     Description: %s\n", rule.Description)
		}
	}

	if len(detail.Devices) > 0 {
		fmt.Fprintf(&stringBuilder, "\nAssigned Devices:\n")

		for _, device := range detail.Devices {
			fmt.Fprintf(&stringBuilder, "  - %s: %s (ID: %d)\n", device.Type, device.Label, device.ID)
		}
	}

	return mcp.NewToolResultText(stringBuilder.String()), nil
}

// handleFirewallCreate creates a new firewall.
func (s *Service) handleFirewallCreate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params FirewallCreateParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Convert rules to linodego format
	inboundRules := make([]linodego.FirewallRule, 0, len(params.Rules.Inbound))
	outboundRules := make([]linodego.FirewallRule, 0, len(params.Rules.Outbound))

	for _, rule := range params.Rules.Inbound {
		inboundRules = append(inboundRules, linodego.FirewallRule{
			Ports:       rule.Ports,
			Protocol:    linodego.NetworkProtocol(rule.Protocol),
			Action:      rule.Action,
			Label:       rule.Label,
			Description: rule.Description,
			Addresses: linodego.NetworkAddresses{
				IPv4: &rule.Addresses.IPv4,
				IPv6: &rule.Addresses.IPv6,
			},
		})
	}

	for _, rule := range params.Rules.Outbound {
		outboundRules = append(outboundRules, linodego.FirewallRule{
			Ports:       rule.Ports,
			Protocol:    linodego.NetworkProtocol(rule.Protocol),
			Action:      rule.Action,
			Label:       rule.Label,
			Description: rule.Description,
			Addresses: linodego.NetworkAddresses{
				IPv4: &rule.Addresses.IPv4,
				IPv6: &rule.Addresses.IPv6,
			},
		})
	}

	createOpts := linodego.FirewallCreateOptions{
		Label: params.Label,
		Rules: linodego.FirewallRuleSet{
			Inbound:        inboundRules,
			InboundPolicy:  params.Rules.InboundPolicy,
			Outbound:       outboundRules,
			OutboundPolicy: params.Rules.OutboundPolicy,
		},
		Tags: params.Tags,
	}

	firewall, err := account.Client.CreateFirewall(ctx, createOpts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create firewall: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Firewall created successfully:\nID: %d\nLabel: %s\nStatus: %s", firewall.ID, firewall.Label, firewall.Status)), nil
}

// handleFirewallUpdate updates an existing firewall.
func (s *Service) handleFirewallUpdate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments, argumentsValid := request.Params.Arguments.(map[string]interface{})
	if !argumentsValid {
		return mcp.NewToolResultError("Invalid arguments format"), nil
	}

	firewallID, err := parseIDFromArguments(arguments, "firewall_id")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid firewall_id parameter: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Parse additional update parameters
	updateOpts := linodego.FirewallUpdateOptions{}
	if label, labelExists := arguments["label"].(string); labelExists && label != "" {
		updateOpts.Label = label
	}
	if tagsRaw, tagsExists := arguments["tags"]; tagsExists {
		if tagsSlice, tagsSliceValid := tagsRaw.([]interface{}); tagsSliceValid {
			tags := make([]string, len(tagsSlice))

			for tagIndex, tag := range tagsSlice {
				if tagStr, tagValid := tag.(string); tagValid {
					tags[tagIndex] = tagStr
				}
			}
			updateOpts.Tags = &tags
		}
	}

	firewall, err := account.Client.UpdateFirewall(ctx, firewallID, updateOpts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update firewall: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Firewall updated successfully:\nID: %d\nLabel: %s\nStatus: %s", firewall.ID, firewall.Label, firewall.Status)), nil
}

// handleFirewallDelete deletes a firewall.
func (s *Service) handleFirewallDelete(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments, argumentsValid := request.Params.Arguments.(map[string]interface{})
	if !argumentsValid {
		return mcp.NewToolResultError("Invalid arguments format"), nil
	}

	firewallID, err := parseIDFromArguments(arguments, "firewall_id")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid firewall_id parameter: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	err = account.Client.DeleteFirewall(ctx, firewallID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete firewall: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Firewall %d deleted successfully", firewallID)), nil
}

// handleFirewallRulesUpdate updates firewall rules.
func (s *Service) handleFirewallRulesUpdate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments, argumentsValid := request.Params.Arguments.(map[string]interface{})
	if !argumentsValid {
		return mcp.NewToolResultError("Invalid arguments format"), nil
	}

	firewallID, err := parseIDFromArguments(arguments, "firewall_id")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid firewall_id parameter: %v", err)), nil
	}

	// Parse rules from arguments - for now, use simplified parsing for the integration test
	var params FirewallRulesUpdateParams
	if err := parseArguments(arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Convert rules to linodego format
	inboundRules := make([]linodego.FirewallRule, 0, len(params.Rules.Inbound))

	outboundRules := make([]linodego.FirewallRule, 0, len(params.Rules.Outbound))

	for _, rule := range params.Rules.Inbound {
		inboundRules = append(inboundRules, linodego.FirewallRule{
			Ports:       rule.Ports,
			Protocol:    linodego.NetworkProtocol(rule.Protocol),
			Action:      rule.Action,
			Label:       rule.Label,
			Description: rule.Description,
			Addresses: linodego.NetworkAddresses{
				IPv4: &rule.Addresses.IPv4,
				IPv6: &rule.Addresses.IPv6,
			},
		})
	}

	for _, rule := range params.Rules.Outbound {
		outboundRules = append(outboundRules, linodego.FirewallRule{
			Ports:       rule.Ports,
			Protocol:    linodego.NetworkProtocol(rule.Protocol),
			Action:      rule.Action,
			Label:       rule.Label,
			Description: rule.Description,
			Addresses: linodego.NetworkAddresses{
				IPv4: &rule.Addresses.IPv4,
				IPv6: &rule.Addresses.IPv6,
			},
		})
	}

	ruleSet := linodego.FirewallRuleSet{
		Inbound:        inboundRules,
		InboundPolicy:  params.Rules.InboundPolicy,
		Outbound:       outboundRules,
		OutboundPolicy: params.Rules.OutboundPolicy,
	}

	_, err = account.Client.UpdateFirewallRules(ctx, firewallID, ruleSet)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update firewall rules: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Firewall rules updated successfully for firewall %d", firewallID)), nil
}

// handleFirewallDeviceCreate assigns a device to a firewall.
func (s *Service) handleFirewallDeviceCreate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments format"), nil
	}

	firewallID, err := parseIDFromArguments(arguments, "firewall_id")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid firewall_id parameter: %v", err)), nil
	}

	deviceID, err := parseIDFromArguments(arguments, "device_id")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid device_id parameter: %v", err)), nil
	}

	// Parse device type
	deviceType := "linode" // default
	if typeRaw, ok := arguments["device_type"].(string); ok {
		deviceType = typeRaw
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	createOpts := linodego.FirewallDeviceCreateOptions{
		ID:   deviceID,
		Type: linodego.FirewallDeviceType(deviceType),
	}

	device, err := account.Client.CreateFirewallDevice(ctx, firewallID, createOpts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to assign device to firewall: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Device assigned successfully:\nDevice ID: %d (%s)\nFirewall ID: %d", device.ID, deviceType, firewallID)), nil
}

// handleFirewallDeviceDelete removes a device from a firewall.
func (s *Service) handleFirewallDeviceDelete(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments format"), nil
	}

	firewallID, err := parseIDFromArguments(arguments, "firewall_id")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid firewall_id parameter: %v", err)), nil
	}

	deviceID, err := parseIDFromArguments(arguments, "device_id")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid device_id parameter: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	err = account.Client.DeleteFirewallDevice(ctx, firewallID, deviceID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to remove device from firewall: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Device %d removed from firewall %d successfully", deviceID, firewallID)), nil
}

// Helper functions for pointer handling.
func stringSlicePtrValue(ptr *[]string) []string {
	if ptr == nil {
		return []string{}
	}

	return *ptr
}
