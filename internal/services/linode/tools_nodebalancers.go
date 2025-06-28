package linode

import (
	"context"
	"fmt"
	"strings"

	"github.com/linode/linodego"
	"github.com/mark3labs/mcp-go/mcp"
)

// handleNodeBalancersList lists all NodeBalancers.
func (s *Service) handleNodeBalancersList(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	nodebalancers, err := account.Client.ListNodeBalancers(ctx, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list NodeBalancers: %v", err)), nil
	}

	var summaries []NodeBalancerSummary
	for _, nb := range nodebalancers {
		summary := NodeBalancerSummary{
			ID:                 nb.ID,
			Label:              stringPtrValue(nb.Label),
			Region:             nb.Region,
			IPv4:               stringPtrValue(nb.IPv4),
			IPv6:               nb.IPv6,
			ClientConnThrottle: nb.ClientConnThrottle,
			Hostname:           stringPtrValue(nb.Hostname),
			Tags:               nb.Tags,
			Created:            nb.Created.Format("2006-01-02T15:04:05"),
			Updated:            nb.Updated.Format("2006-01-02T15:04:05"),
			Transfer: NodeBalancerTransfer{
				In:    float64PtrValue(nb.Transfer.In),
				Out:   float64PtrValue(nb.Transfer.Out),
				Total: float64PtrValue(nb.Transfer.Total),
			},
		}
		summaries = append(summaries, summary)
	}

	// Remove unused result variable

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d NodeBalancers:\n\n", len(summaries)))

	for _, nb := range summaries {
		fmt.Fprintf(&sb, "ID: %d | %s (%s)\n", nb.ID, nb.Label, nb.Region)
		fmt.Fprintf(&sb, "  IPv4: %s\n", nb.IPv4)
		if nb.IPv6 != nil {
			fmt.Fprintf(&sb, "  IPv6: %s\n", *nb.IPv6)
		}
		fmt.Fprintf(&sb, "  Hostname: %s\n", nb.Hostname)
		fmt.Fprintf(&sb, "  Throttle: %d conn/sec\n", nb.ClientConnThrottle)
		fmt.Fprintf(&sb, "  Transfer: %.2f GB in, %.2f GB out\n", nb.Transfer.In, nb.Transfer.Out)
		if len(nb.Tags) > 0 {
			fmt.Fprintf(&sb, "  Tags: %s\n", strings.Join(nb.Tags, ", "))
		}
		sb.WriteString("\n")
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// handleNodeBalancerGet gets details of a specific NodeBalancer.
func (s *Service) handleNodeBalancerGet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()
	nodebalancerID, err := parseIDFromArguments(arguments, "nodebalancer_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	nb, err := account.Client.GetNodeBalancer(ctx, nodebalancerID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get NodeBalancer: %v", err)), nil
	}

	// Get configurations
	configs, err := account.Client.ListNodeBalancerConfigs(ctx, nodebalancerID, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get NodeBalancer configurations: %v", err)), nil
	}

	var configDetails []NodeBalancerConfig
	for _, config := range configs {
		configDetails = append(configDetails, NodeBalancerConfig{
			ID:             config.ID,
			Port:           config.Port,
			Protocol:       string(config.Protocol),
			Algorithm:      string(config.Algorithm),
			Stickiness:     string(config.Stickiness),
			Check:          string(config.Check),
			CheckInterval:  config.CheckInterval,
			CheckTimeout:   config.CheckTimeout,
			CheckAttempts:  config.CheckAttempts,
			CheckPath:      config.CheckPath,
			CheckBody:      config.CheckBody,
			CheckPassive:   config.CheckPassive,
			ProxyProtocol:  string(config.ProxyProtocol),
			CipherSuite:    string(config.CipherSuite),
			SSLCommonName:  config.SSLCommonName,
			SSLFingerprint: config.SSLFingerprint,
			SSLCert:        config.SSLCert,
			SSLKey:         config.SSLKey,
			NodesStatus: NodeBalancerNodeStatus{
				Up:   config.NodesStatus.Up,
				Down: config.NodesStatus.Down,
			},
		})
	}

	detail := NodeBalancerDetail{
		ID:                 nb.ID,
		Label:              stringPtrValue(nb.Label),
		Region:             nb.Region,
		IPv4:               stringPtrValue(nb.IPv4),
		IPv6:               nb.IPv6,
		ClientConnThrottle: nb.ClientConnThrottle,
		Hostname:           stringPtrValue(nb.Hostname),
		Tags:               nb.Tags,
		Created:            nb.Created.Format("2006-01-02T15:04:05"),
		Updated:            nb.Updated.Format("2006-01-02T15:04:05"),
		Transfer: NodeBalancerTransfer{
			In:    float64PtrValue(nb.Transfer.In),
			Out:   float64PtrValue(nb.Transfer.Out),
			Total: float64PtrValue(nb.Transfer.Total),
		},
		Configs: configDetails,
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "NodeBalancer Details:\n")
	fmt.Fprintf(&sb, "ID: %d\n", detail.ID)
	fmt.Fprintf(&sb, "Label: %s\n", detail.Label)
	fmt.Fprintf(&sb, "Region: %s\n", detail.Region)
	fmt.Fprintf(&sb, "IPv4: %s\n", detail.IPv4)
	if detail.IPv6 != nil {
		fmt.Fprintf(&sb, "IPv6: %s\n", *detail.IPv6)
	}
	fmt.Fprintf(&sb, "Hostname: %s\n", detail.Hostname)
	fmt.Fprintf(&sb, "Client Connection Throttle: %d conn/sec\n", detail.ClientConnThrottle)
	fmt.Fprintf(&sb, "Created: %s\n", detail.Created)
	fmt.Fprintf(&sb, "Updated: %s\n\n", detail.Updated)

	if len(detail.Tags) > 0 {
		fmt.Fprintf(&sb, "Tags: %s\n\n", strings.Join(detail.Tags, ", "))
	}

	fmt.Fprintf(&sb, "Transfer Stats:\n")
	fmt.Fprintf(&sb, "  In: %.2f GB\n", detail.Transfer.In)
	fmt.Fprintf(&sb, "  Out: %.2f GB\n", detail.Transfer.Out)
	fmt.Fprintf(&sb, "  Total: %.2f GB\n\n", detail.Transfer.Total)

	if len(detail.Configs) > 0 {
		fmt.Fprintf(&sb, "Configurations:\n")
		for i, config := range detail.Configs {
			fmt.Fprintf(&sb, "  %d. Port %d (%s)\n", i+1, config.Port, config.Protocol)
			fmt.Fprintf(&sb, "     Algorithm: %s\n", config.Algorithm)
			fmt.Fprintf(&sb, "     Stickiness: %s\n", config.Stickiness)
			fmt.Fprintf(&sb, "     Health Check: %s\n", config.Check)
			if config.Check != "none" {
				fmt.Fprintf(&sb, "     Check Interval: %ds, Timeout: %ds, Attempts: %d\n",
					config.CheckInterval, config.CheckTimeout, config.CheckAttempts)
				if config.CheckPath != "" {
					fmt.Fprintf(&sb, "     Check Path: %s\n", config.CheckPath)
				}
			}
			fmt.Fprintf(&sb, "     Nodes: %d up, %d down\n", config.NodesStatus.Up, config.NodesStatus.Down)
			fmt.Fprintf(&sb, "     Proxy Protocol: %s\n", config.ProxyProtocol)
			if config.SSLCert != "" {
				fmt.Fprintf(&sb, "     SSL: Enabled (Common Name: %s)\n", config.SSLCommonName)
			}
			sb.WriteString("\n")
		}
	} else {
		sb.WriteString("No configurations found.\n")
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// handleNodeBalancerCreate creates a new NodeBalancer.
func (s *Service) handleNodeBalancerCreate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params NodeBalancerCreateParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	createOpts := linodego.NodeBalancerCreateOptions{
		Label:              &params.Label,
		Region:             params.Region,
		ClientConnThrottle: &params.ClientConnThrottle,
		Tags:               params.Tags,
	}

	nb, err := account.Client.CreateNodeBalancer(ctx, createOpts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create NodeBalancer: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("NodeBalancer created successfully:\nID: %d\nLabel: %s\nRegion: %s\nIPv4: %s\nHostname: %s",
		nb.ID, stringPtrValue(nb.Label), nb.Region, stringPtrValue(nb.IPv4), stringPtrValue(nb.Hostname))), nil
}

// handleNodeBalancerUpdate updates an existing NodeBalancer.
func (s *Service) handleNodeBalancerUpdate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params NodeBalancerUpdateParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	updateOpts := linodego.NodeBalancerUpdateOptions{}
	if params.Label != "" {
		updateOpts.Label = &params.Label
	}
	if params.ClientConnThrottle > 0 {
		updateOpts.ClientConnThrottle = &params.ClientConnThrottle
	}
	if params.Tags != nil {
		updateOpts.Tags = &params.Tags
	}

	nb, err := account.Client.UpdateNodeBalancer(ctx, params.NodeBalancerID, updateOpts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update NodeBalancer: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("NodeBalancer updated successfully:\nID: %d\nLabel: %s\nRegion: %s\nThrottle: %d conn/sec",
		nb.ID, stringPtrValue(nb.Label), nb.Region, nb.ClientConnThrottle)), nil
}

// handleNodeBalancerDelete deletes a NodeBalancer.
func (s *Service) handleNodeBalancerDelete(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()
	nodebalancerID, err := parseIDFromArguments(arguments, "nodebalancer_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	err = account.Client.DeleteNodeBalancer(ctx, nodebalancerID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete NodeBalancer: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("NodeBalancer %d deleted successfully", nodebalancerID)), nil
}

// handleNodeBalancerConfigCreate creates a new NodeBalancer configuration.
func (s *Service) handleNodeBalancerConfigCreate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params NodeBalancerConfigCreateParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	createOpts := linodego.NodeBalancerConfigCreateOptions{
		Port:     params.Port,
		Protocol: linodego.ConfigProtocol(params.Protocol),
	}

	if params.Algorithm != "" {
		createOpts.Algorithm = linodego.ConfigAlgorithm(params.Algorithm)
	}
	if params.Stickiness != "" {
		createOpts.Stickiness = linodego.ConfigStickiness(params.Stickiness)
	}
	if params.Check != "" {
		createOpts.Check = linodego.ConfigCheck(params.Check)
	}
	if params.CheckInterval > 0 {
		createOpts.CheckInterval = params.CheckInterval
	}
	if params.CheckTimeout > 0 {
		createOpts.CheckTimeout = params.CheckTimeout
	}
	if params.CheckAttempts > 0 {
		createOpts.CheckAttempts = params.CheckAttempts
	}
	if params.CheckPath != "" {
		createOpts.CheckPath = params.CheckPath
	}
	if params.CheckBody != "" {
		createOpts.CheckBody = params.CheckBody
	}
	if params.CheckPassive {
		createOpts.CheckPassive = &params.CheckPassive
	}
	if params.ProxyProtocol != "" {
		createOpts.ProxyProtocol = linodego.ConfigProxyProtocol(params.ProxyProtocol)
	}
	if params.SSLCert != "" {
		createOpts.SSLCert = params.SSLCert
	}
	if params.SSLKey != "" {
		createOpts.SSLKey = params.SSLKey
	}

	config, err := account.Client.CreateNodeBalancerConfig(ctx, params.NodeBalancerID, createOpts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create NodeBalancer configuration: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("NodeBalancer configuration created successfully:\nConfig ID: %d\nPort: %d\nProtocol: %s\nAlgorithm: %s",
		config.ID, config.Port, config.Protocol, config.Algorithm)), nil
}

// handleNodeBalancerConfigUpdate updates a NodeBalancer configuration.
func (s *Service) handleNodeBalancerConfigUpdate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params NodeBalancerConfigUpdateParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	updateOpts := linodego.NodeBalancerConfigUpdateOptions{}

	if params.Port > 0 {
		updateOpts.Port = params.Port
	}
	if params.Protocol != "" {
		updateOpts.Protocol = linodego.ConfigProtocol(params.Protocol)
	}
	if params.Algorithm != "" {
		updateOpts.Algorithm = linodego.ConfigAlgorithm(params.Algorithm)
	}
	if params.Stickiness != "" {
		updateOpts.Stickiness = linodego.ConfigStickiness(params.Stickiness)
	}
	if params.Check != "" {
		updateOpts.Check = linodego.ConfigCheck(params.Check)
	}
	if params.CheckInterval > 0 {
		updateOpts.CheckInterval = params.CheckInterval
	}
	if params.CheckTimeout > 0 {
		updateOpts.CheckTimeout = params.CheckTimeout
	}
	if params.CheckAttempts > 0 {
		updateOpts.CheckAttempts = params.CheckAttempts
	}
	if params.CheckPath != "" {
		updateOpts.CheckPath = params.CheckPath
	}
	if params.CheckBody != "" {
		updateOpts.CheckBody = params.CheckBody
	}
	if params.CheckPassive {
		updateOpts.CheckPassive = &params.CheckPassive
	}
	if params.ProxyProtocol != "" {
		updateOpts.ProxyProtocol = linodego.ConfigProxyProtocol(params.ProxyProtocol)
	}
	if params.SSLCert != "" {
		updateOpts.SSLCert = params.SSLCert
	}
	if params.SSLKey != "" {
		updateOpts.SSLKey = params.SSLKey
	}

	config, err := account.Client.UpdateNodeBalancerConfig(ctx, params.NodeBalancerID, params.ConfigID, updateOpts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update NodeBalancer configuration: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("NodeBalancer configuration updated successfully:\nConfig ID: %d\nPort: %d\nProtocol: %s\nAlgorithm: %s",
		config.ID, config.Port, config.Protocol, config.Algorithm)), nil
}

// handleNodeBalancerConfigDelete deletes a NodeBalancer configuration.
func (s *Service) handleNodeBalancerConfigDelete(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()
	nodebalancerID, err := parseIDFromArguments(arguments, "nodebalancer_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	configID, err := parseIDFromArguments(arguments, "config_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	err = account.Client.DeleteNodeBalancerConfig(ctx, nodebalancerID, configID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete NodeBalancer configuration: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("NodeBalancer configuration %d deleted successfully from NodeBalancer %d",
		configID, nodebalancerID)), nil
}

// Helper functions for pointer handling
func float64PtrValue(ptr *float64) float64 {
	if ptr == nil {
		return 0
	}
	return *ptr
}

// stringPtrValue is already defined in tools_domains.go
