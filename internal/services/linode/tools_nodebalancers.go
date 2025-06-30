package linode

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/linode/linodego"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/chadit/CloudMCP/pkg/types"
)

const (
	// configurationIndexOffset is the offset for displaying configuration numbers.
	configurationIndexOffset = 1
)

var (
	// ErrFailedToListNodeBalancers is returned when listing NodeBalancers fails.
	ErrFailedToListNodeBalancers = errors.New("failed to list NodeBalancers")
	// ErrFailedToGetNodeBalancer is returned when getting a NodeBalancer fails.
	ErrFailedToGetNodeBalancer = errors.New("failed to get NodeBalancer")
	// ErrFailedToGetNodeBalancerConfigs is returned when getting NodeBalancer configurations fails.
	ErrFailedToGetNodeBalancerConfigs = errors.New("failed to get NodeBalancer configurations")
	// ErrFailedToCreateNodeBalancer is returned when creating a NodeBalancer fails.
	ErrFailedToCreateNodeBalancer = errors.New("failed to create NodeBalancer")
	// ErrFailedToUpdateNodeBalancer is returned when updating a NodeBalancer fails.
	ErrFailedToUpdateNodeBalancer = errors.New("failed to update NodeBalancer")
	// ErrFailedToDeleteNodeBalancer is returned when deleting a NodeBalancer fails.
	ErrFailedToDeleteNodeBalancer = errors.New("failed to delete NodeBalancer")
	// ErrFailedToCreateNodeBalancerConfig is returned when creating a NodeBalancer configuration fails.
	ErrFailedToCreateNodeBalancerConfig = errors.New("failed to create NodeBalancer configuration")
	// ErrFailedToUpdateNodeBalancerConfig is returned when updating a NodeBalancer configuration fails.
	ErrFailedToUpdateNodeBalancerConfig = errors.New("failed to update NodeBalancer configuration")
	// ErrFailedToDeleteNodeBalancerConfig is returned when deleting a NodeBalancer configuration fails.
	ErrFailedToDeleteNodeBalancerConfig = errors.New("failed to delete NodeBalancer configuration")
	// ErrInvalidNodeBalancerParameters is returned when NodeBalancer parameters are invalid.
	ErrInvalidNodeBalancerParameters = errors.New("invalid parameters")
)

// handleNodeBalancersList lists all NodeBalancers.
func (s *Service) handleNodeBalancersList(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	account, accountErr := s.accountManager.GetCurrentAccount()
	if accountErr != nil {
		return nil, fmt.Errorf("failed to get current account: %w", accountErr)
	}

	nodebalancers, nodebalancersErr := account.Client.ListNodeBalancers(ctx, nil)
	if nodebalancersErr != nil {
		return nil, types.NewToolError("linode", "nodebalancers_list", //nolint:wrapcheck // types.NewToolError already wraps the error
			"failed to list NodeBalancers", nodebalancersErr)
	}

	summaries := make([]NodeBalancerSummary, 0, len(nodebalancers))

	for _, nodeBalancer := range nodebalancers {
		summary := NodeBalancerSummary{
			ID:                 nodeBalancer.ID,
			Label:              stringPtrValue(nodeBalancer.Label),
			Region:             nodeBalancer.Region,
			IPv4:               stringPtrValue(nodeBalancer.IPv4),
			IPv6:               nodeBalancer.IPv6,
			ClientConnThrottle: nodeBalancer.ClientConnThrottle,
			Hostname:           stringPtrValue(nodeBalancer.Hostname),
			Tags:               nodeBalancer.Tags,
			Created:            nodeBalancer.Created.Format(timeFormatLayout),
			Updated:            nodeBalancer.Updated.Format(timeFormatLayout),
			Transfer: NodeBalancerTransfer{
				In:    float64PtrValue(nodeBalancer.Transfer.In),
				Out:   float64PtrValue(nodeBalancer.Transfer.Out),
				Total: float64PtrValue(nodeBalancer.Transfer.Total),
			},
		}

		summaries = append(summaries, summary)
	}

	var stringBuilder strings.Builder

	stringBuilder.WriteString(fmt.Sprintf("Found %d NodeBalancers:\n\n", len(summaries)))

	for _, nodeBalancer := range summaries {
		fmt.Fprintf(&stringBuilder, "ID: %d | %s (%s)\n", nodeBalancer.ID, nodeBalancer.Label, nodeBalancer.Region)
		fmt.Fprintf(&stringBuilder, "  IPv4: %s\n", nodeBalancer.IPv4)

		if nodeBalancer.IPv6 != nil {
			fmt.Fprintf(&stringBuilder, "  IPv6: %s\n", *nodeBalancer.IPv6)
		}

		fmt.Fprintf(&stringBuilder, "  Hostname: %s\n", nodeBalancer.Hostname)
		fmt.Fprintf(&stringBuilder, "  Throttle: %d conn/sec\n", nodeBalancer.ClientConnThrottle)
		fmt.Fprintf(&stringBuilder, "  Transfer: %.2f GB in, %.2f GB out\n", nodeBalancer.Transfer.In, nodeBalancer.Transfer.Out)

		if len(nodeBalancer.Tags) > 0 {
			fmt.Fprintf(&stringBuilder, "  Tags: %s\n", strings.Join(nodeBalancer.Tags, ", "))
		}

		stringBuilder.WriteString("\n")
	}

	return mcp.NewToolResultText(stringBuilder.String()), nil
}

// handleNodeBalancerGet gets details of a specific NodeBalancer.
func (s *Service) handleNodeBalancerGet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()
	nodebalancerID, parseErr := parseIDFromArguments(arguments, "nodebalancer_id")

	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse nodebalancer ID: %w", parseErr)
	}

	account, accountErr := s.accountManager.GetCurrentAccount()
	if accountErr != nil {
		return nil, fmt.Errorf("failed to get current account: %w", accountErr)
	}

	nodeBalancer, nodeBalancerErr := account.Client.GetNodeBalancer(ctx, nodebalancerID)
	if nodeBalancerErr != nil {
		return nil, types.NewToolError("linode", "nodebalancer_get", //nolint:wrapcheck // types.NewToolError already wraps the error
			"failed to get NodeBalancer", nodeBalancerErr)
	}

	// Get configurations.
	configurations, configurationsErr := account.Client.ListNodeBalancerConfigs(ctx, nodebalancerID, nil)
	if configurationsErr != nil {
		return nil, types.NewToolError("linode", "nodebalancer_configs_list", //nolint:wrapcheck // types.NewToolError already wraps the error
			"failed to get NodeBalancer configurations", configurationsErr)
	}

	configurationDetails := make([]NodeBalancerConfig, 0, len(configurations))

	for _, configuration := range configurations {
		configurationDetails = append(configurationDetails, NodeBalancerConfig{
			ID:             configuration.ID,
			Port:           configuration.Port,
			Protocol:       string(configuration.Protocol),
			Algorithm:      string(configuration.Algorithm),
			Stickiness:     string(configuration.Stickiness),
			Check:          string(configuration.Check),
			CheckInterval:  configuration.CheckInterval,
			CheckTimeout:   configuration.CheckTimeout,
			CheckAttempts:  configuration.CheckAttempts,
			CheckPath:      configuration.CheckPath,
			CheckBody:      configuration.CheckBody,
			CheckPassive:   configuration.CheckPassive,
			ProxyProtocol:  string(configuration.ProxyProtocol),
			CipherSuite:    string(configuration.CipherSuite),
			SSLCommonName:  configuration.SSLCommonName,
			SSLFingerprint: configuration.SSLFingerprint,
			SSLCert:        configuration.SSLCert,
			SSLKey:         configuration.SSLKey,
			NodesStatus: NodeBalancerNodeStatus{
				Up:   configuration.NodesStatus.Up,
				Down: configuration.NodesStatus.Down,
			},
		})
	}

	detail := NodeBalancerDetail{
		ID:                 nodeBalancer.ID,
		Label:              stringPtrValue(nodeBalancer.Label),
		Region:             nodeBalancer.Region,
		IPv4:               stringPtrValue(nodeBalancer.IPv4),
		IPv6:               nodeBalancer.IPv6,
		ClientConnThrottle: nodeBalancer.ClientConnThrottle,
		Hostname:           stringPtrValue(nodeBalancer.Hostname),
		Tags:               nodeBalancer.Tags,
		Created:            nodeBalancer.Created.Format(timeFormatLayout),
		Updated:            nodeBalancer.Updated.Format(timeFormatLayout),
		Transfer: NodeBalancerTransfer{
			In:    float64PtrValue(nodeBalancer.Transfer.In),
			Out:   float64PtrValue(nodeBalancer.Transfer.Out),
			Total: float64PtrValue(nodeBalancer.Transfer.Total),
		},
		Configs: configurationDetails,
	}

	var stringBuilder strings.Builder

	fmt.Fprintf(&stringBuilder, "NodeBalancer Details:\n")
	fmt.Fprintf(&stringBuilder, "ID: %d\n", detail.ID)
	fmt.Fprintf(&stringBuilder, "Label: %s\n", detail.Label)
	fmt.Fprintf(&stringBuilder, "Region: %s\n", detail.Region)
	fmt.Fprintf(&stringBuilder, "IPv4: %s\n", detail.IPv4)

	if detail.IPv6 != nil {
		fmt.Fprintf(&stringBuilder, "IPv6: %s\n", *detail.IPv6)
	}

	fmt.Fprintf(&stringBuilder, "Hostname: %s\n", detail.Hostname)
	fmt.Fprintf(&stringBuilder, "Client Connection Throttle: %d conn/sec\n", detail.ClientConnThrottle)
	fmt.Fprintf(&stringBuilder, "Created: %s\n", detail.Created)
	fmt.Fprintf(&stringBuilder, "Updated: %s\n\n", detail.Updated)

	if len(detail.Tags) > 0 {
		fmt.Fprintf(&stringBuilder, "Tags: %s\n\n", strings.Join(detail.Tags, ", "))
	}

	fmt.Fprintf(&stringBuilder, "Transfer Stats:\n")
	fmt.Fprintf(&stringBuilder, "  In: %.2f GB\n", detail.Transfer.In)
	fmt.Fprintf(&stringBuilder, "  Out: %.2f GB\n", detail.Transfer.Out)
	fmt.Fprintf(&stringBuilder, "  Total: %.2f GB\n\n", detail.Transfer.Total)

	if len(detail.Configs) > 0 {
		fmt.Fprintf(&stringBuilder, "Configurations:\n")

		for configIndex, configuration := range detail.Configs {
			formatNodeBalancerConfiguration(&stringBuilder, configIndex, configuration)
		}
	} else {
		stringBuilder.WriteString("No configurations found.\n")
	}

	return mcp.NewToolResultText(stringBuilder.String()), nil
}

// handleNodeBalancerCreate creates a new NodeBalancer.
func (s *Service) handleNodeBalancerCreate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params NodeBalancerCreateParams
	if parseErr := parseArguments(request.Params.Arguments, &params); parseErr != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidNodeBalancerParameters, parseErr)
	}

	account, accountErr := s.accountManager.GetCurrentAccount()
	if accountErr != nil {
		return nil, fmt.Errorf("failed to get current account: %w", accountErr)
	}

	createOptions := linodego.NodeBalancerCreateOptions{
		Label:              &params.Label,
		Region:             params.Region,
		ClientConnThrottle: &params.ClientConnThrottle,
		Tags:               params.Tags,
	}

	nodeBalancer, createErr := account.Client.CreateNodeBalancer(ctx, createOptions)
	if createErr != nil {
		return nil, types.NewToolError("linode", "nodebalancer_create", //nolint:wrapcheck // types.NewToolError already wraps the error
			"failed to create NodeBalancer", createErr)
	}

	return mcp.NewToolResultText(fmt.Sprintf("NodeBalancer created successfully:\nID: %d\nLabel: %s\nRegion: %s\nIPv4: %s\nHostname: %s",
		nodeBalancer.ID, stringPtrValue(nodeBalancer.Label), nodeBalancer.Region, stringPtrValue(nodeBalancer.IPv4), stringPtrValue(nodeBalancer.Hostname))), nil
}

// handleNodeBalancerUpdate updates an existing NodeBalancer.
func (s *Service) handleNodeBalancerUpdate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params NodeBalancerUpdateParams
	if parseErr := parseArguments(request.Params.Arguments, &params); parseErr != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidNodeBalancerParameters, parseErr)
	}

	account, accountErr := s.accountManager.GetCurrentAccount()
	if accountErr != nil {
		return nil, fmt.Errorf("failed to get current account: %w", accountErr)
	}

	updateOptions := linodego.NodeBalancerUpdateOptions{}
	if params.Label != "" {
		updateOptions.Label = &params.Label
	}

	if params.ClientConnThrottle > 0 {
		updateOptions.ClientConnThrottle = &params.ClientConnThrottle
	}

	if params.Tags != nil {
		updateOptions.Tags = &params.Tags
	}

	nodeBalancer, updateErr := account.Client.UpdateNodeBalancer(ctx, params.NodeBalancerID, updateOptions)
	if updateErr != nil {
		return nil, types.NewToolError("linode", "nodebalancer_update", //nolint:wrapcheck // types.NewToolError already wraps the error
			"failed to update NodeBalancer", updateErr)
	}

	return mcp.NewToolResultText(fmt.Sprintf("NodeBalancer updated successfully:\nID: %d\nLabel: %s\nRegion: %s\nThrottle: %d conn/sec",
		nodeBalancer.ID, stringPtrValue(nodeBalancer.Label), nodeBalancer.Region, nodeBalancer.ClientConnThrottle)), nil
}

// handleNodeBalancerDelete deletes a NodeBalancer.
func (s *Service) handleNodeBalancerDelete(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()
	nodebalancerID, parseErr := parseIDFromArguments(arguments, "nodebalancer_id")

	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse nodebalancer ID: %w", parseErr)
	}

	account, accountErr := s.accountManager.GetCurrentAccount()
	if accountErr != nil {
		return nil, fmt.Errorf("failed to get current account: %w", accountErr)
	}

	deleteErr := account.Client.DeleteNodeBalancer(ctx, nodebalancerID)
	if deleteErr != nil {
		return nil, types.NewToolError("linode", "nodebalancer_delete", //nolint:wrapcheck // types.NewToolError already wraps the error
			"failed to delete NodeBalancer", deleteErr)
	}

	return mcp.NewToolResultText(fmt.Sprintf("NodeBalancer %d deleted successfully", nodebalancerID)), nil
}

// handleNodeBalancerConfigCreate creates a new NodeBalancer configuration.
func (s *Service) handleNodeBalancerConfigCreate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params NodeBalancerConfigCreateParams
	if parseErr := parseArguments(request.Params.Arguments, &params); parseErr != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidNodeBalancerParameters, parseErr)
	}

	account, accountErr := s.accountManager.GetCurrentAccount()
	if accountErr != nil {
		return nil, fmt.Errorf("failed to get current account: %w", accountErr)
	}

	createOptions := linodego.NodeBalancerConfigCreateOptions{
		Port:     params.Port,
		Protocol: linodego.ConfigProtocol(params.Protocol),
	}

	if params.Algorithm != "" {
		createOptions.Algorithm = linodego.ConfigAlgorithm(params.Algorithm)
	}

	if params.Stickiness != "" {
		createOptions.Stickiness = linodego.ConfigStickiness(params.Stickiness)
	}

	if params.Check != "" {
		createOptions.Check = linodego.ConfigCheck(params.Check)
	}

	if params.CheckInterval > 0 {
		createOptions.CheckInterval = params.CheckInterval
	}

	if params.CheckTimeout > 0 {
		createOptions.CheckTimeout = params.CheckTimeout
	}

	if params.CheckAttempts > 0 {
		createOptions.CheckAttempts = params.CheckAttempts
	}

	if params.CheckPath != "" {
		createOptions.CheckPath = params.CheckPath
	}

	if params.CheckBody != "" {
		createOptions.CheckBody = params.CheckBody
	}

	if params.CheckPassive {
		createOptions.CheckPassive = &params.CheckPassive
	}

	if params.ProxyProtocol != "" {
		createOptions.ProxyProtocol = linodego.ConfigProxyProtocol(params.ProxyProtocol)
	}

	if params.SSLCert != "" {
		createOptions.SSLCert = params.SSLCert
	}

	if params.SSLKey != "" {
		createOptions.SSLKey = params.SSLKey
	}

	configuration, createErr := account.Client.CreateNodeBalancerConfig(ctx, params.NodeBalancerID, createOptions)
	if createErr != nil {
		return nil, types.NewToolError("linode", "nodebalancer_config_create", //nolint:wrapcheck // types.NewToolError already wraps the error
			"failed to create NodeBalancer configuration", createErr)
	}

	return mcp.NewToolResultText(fmt.Sprintf("NodeBalancer configuration created successfully:\nConfig ID: %d\nPort: %d\nProtocol: %s\nAlgorithm: %s",
		configuration.ID, configuration.Port, configuration.Protocol, configuration.Algorithm)), nil
}

// handleNodeBalancerConfigUpdate updates a NodeBalancer configuration.
func (s *Service) handleNodeBalancerConfigUpdate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params NodeBalancerConfigUpdateParams
	if parseErr := parseArguments(request.Params.Arguments, &params); parseErr != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidNodeBalancerParameters, parseErr)
	}

	account, accountErr := s.accountManager.GetCurrentAccount()
	if accountErr != nil {
		return nil, fmt.Errorf("failed to get current account: %w", accountErr)
	}

	updateOptions := linodego.NodeBalancerConfigUpdateOptions{}

	if params.Port > 0 {
		updateOptions.Port = params.Port
	}

	if params.Protocol != "" {
		updateOptions.Protocol = linodego.ConfigProtocol(params.Protocol)
	}

	if params.Algorithm != "" {
		updateOptions.Algorithm = linodego.ConfigAlgorithm(params.Algorithm)
	}

	if params.Stickiness != "" {
		updateOptions.Stickiness = linodego.ConfigStickiness(params.Stickiness)
	}

	if params.Check != "" {
		updateOptions.Check = linodego.ConfigCheck(params.Check)
	}

	if params.CheckInterval > 0 {
		updateOptions.CheckInterval = params.CheckInterval
	}

	if params.CheckTimeout > 0 {
		updateOptions.CheckTimeout = params.CheckTimeout
	}

	if params.CheckAttempts > 0 {
		updateOptions.CheckAttempts = params.CheckAttempts
	}

	if params.CheckPath != "" {
		updateOptions.CheckPath = params.CheckPath
	}

	if params.CheckBody != "" {
		updateOptions.CheckBody = params.CheckBody
	}

	if params.CheckPassive {
		updateOptions.CheckPassive = &params.CheckPassive
	}

	if params.ProxyProtocol != "" {
		updateOptions.ProxyProtocol = linodego.ConfigProxyProtocol(params.ProxyProtocol)
	}

	if params.SSLCert != "" {
		updateOptions.SSLCert = params.SSLCert
	}

	if params.SSLKey != "" {
		updateOptions.SSLKey = params.SSLKey
	}

	configuration, updateErr := account.Client.UpdateNodeBalancerConfig(ctx, params.NodeBalancerID, params.ConfigID, updateOptions)
	if updateErr != nil {
		return nil, types.NewToolError("linode", "nodebalancer_config_update", //nolint:wrapcheck // types.NewToolError already wraps the error
			"failed to update NodeBalancer configuration", updateErr)
	}

	return mcp.NewToolResultText(fmt.Sprintf("NodeBalancer configuration updated successfully:\nConfig ID: %d\nPort: %d\nProtocol: %s\nAlgorithm: %s",
		configuration.ID, configuration.Port, configuration.Protocol, configuration.Algorithm)), nil
}

// handleNodeBalancerConfigDelete deletes a NodeBalancer configuration.
func (s *Service) handleNodeBalancerConfigDelete(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()
	nodebalancerID, nodebalancerParseErr := parseIDFromArguments(arguments, "nodebalancer_id")

	if nodebalancerParseErr != nil {
		return nil, fmt.Errorf("failed to parse nodebalancer ID: %w", nodebalancerParseErr)
	}

	configurationID, configurationParseErr := parseIDFromArguments(arguments, "config_id")

	if configurationParseErr != nil {
		return nil, fmt.Errorf("failed to parse configuration ID: %w", configurationParseErr)
	}

	account, accountErr := s.accountManager.GetCurrentAccount()
	if accountErr != nil {
		return nil, fmt.Errorf("failed to get current account: %w", accountErr)
	}

	deleteErr := account.Client.DeleteNodeBalancerConfig(ctx, nodebalancerID, configurationID)
	if deleteErr != nil {
		return nil, types.NewToolError("linode", "nodebalancer_config_delete", //nolint:wrapcheck // types.NewToolError already wraps the error
			"failed to delete NodeBalancer configuration", deleteErr)
	}

	return mcp.NewToolResultText(fmt.Sprintf("NodeBalancer configuration %d deleted successfully from NodeBalancer %d",
		configurationID, nodebalancerID)), nil
}

// formatNodeBalancerConfiguration formats a single NodeBalancer configuration for display.
func formatNodeBalancerConfiguration(stringBuilder *strings.Builder, configIndex int, configuration NodeBalancerConfig) {
	fmt.Fprintf(stringBuilder, "  %d. Port %d (%s)\n", configIndex+configurationIndexOffset, configuration.Port, configuration.Protocol)
	fmt.Fprintf(stringBuilder, "     Algorithm: %s\n", configuration.Algorithm)
	fmt.Fprintf(stringBuilder, "     Stickiness: %s\n", configuration.Stickiness)
	fmt.Fprintf(stringBuilder, "     Health Check: %s\n", configuration.Check)

	if configuration.Check != "none" {
		fmt.Fprintf(stringBuilder, "     Check Interval: %ds, Timeout: %ds, Attempts: %d\n",
			configuration.CheckInterval, configuration.CheckTimeout, configuration.CheckAttempts)

		if configuration.CheckPath != "" {
			fmt.Fprintf(stringBuilder, "     Check Path: %s\n", configuration.CheckPath)
		}
	}

	fmt.Fprintf(stringBuilder, "     Nodes: %d up, %d down\n", configuration.NodesStatus.Up, configuration.NodesStatus.Down)
	fmt.Fprintf(stringBuilder, "     Proxy Protocol: %s\n", configuration.ProxyProtocol)

	if configuration.SSLCert != "" {
		fmt.Fprintf(stringBuilder, "     SSL: Enabled (Common Name: %s)\n", configuration.SSLCommonName)
	}

	stringBuilder.WriteString("\n")
}

// Helper functions for pointer handling.
func float64PtrValue(ptr *float64) float64 {
	if ptr == nil {
		return 0
	}

	return *ptr
}

// stringPtrValue is already defined in tools_domains.go.
