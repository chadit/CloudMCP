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
	mbToGB = 1024
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

func (s *Service) handleInstanceGet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()
	instanceID, ok := arguments["instance_id"].(float64)
	if !ok || instanceID <= 0 {
		return mcp.NewToolResultError("instance_id is required and must be a positive number"), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, err
	}

	instance, err := account.Client.GetInstance(ctx, int(instanceID))
	if err != nil {
		return nil, types.NewToolError("linode", "instance_get",
			fmt.Sprintf("failed to get instance %d", int(instanceID)), err)
	}

	// Format IPv4 addresses
	ipv4Addresses := make([]string, 0)
	for _, ip := range instance.IPv4 {
		ipv4Addresses = append(ipv4Addresses, ip.String())
	}

	// Format the detailed instance information
	resultText := fmt.Sprintf(`Instance Details:
ID: %d
Label: %s
Status: %s
Region: %s
Type: %s
Image: %s

Specifications:
- CPUs: %d
- Memory: %d MB
- Disk: %d GB
- Transfer: %d GB

Network:
- IPv4: %s
- IPv6: %s

Created: %s
Updated: %s

Backups: %s
Watchdog: %s`,
		instance.ID,
		instance.Label,
		instance.Status,
		instance.Region,
		instance.Type,
		instance.Image,
		instance.Specs.VCPUs,
		instance.Specs.Memory,
		instance.Specs.Disk/mbToGB,     // Convert MB to GB
		instance.Specs.Transfer/mbToGB, // Convert MB to GB
		strings.Join(ipv4Addresses, ", "),
		instance.IPv6,
		instance.Created.Format("2006-01-02 15:04:05"),
		instance.Updated.Format("2006-01-02 15:04:05"),
		formatBool(instance.Backups.Enabled),
		formatBool(instance.WatchdogEnabled),
	)

	if len(instance.Tags) > 0 {
		resultText += fmt.Sprintf("\nTags: %s", strings.Join(instance.Tags, ", "))
	}

	return mcp.NewToolResultText(resultText), nil
}

func formatBool(b bool) string {
	if b {
		return "Enabled"
	}
	return "Disabled"
}

func (s *Service) handleInstanceCreate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()

	// Required parameters
	region, ok := arguments["region"].(string)
	if !ok || region == "" {
		return mcp.NewToolResultError("region is required"), nil
	}

	instanceType, ok := arguments["type"].(string)
	if !ok || instanceType == "" {
		return mcp.NewToolResultError("type is required"), nil
	}

	label, ok := arguments["label"].(string)
	if !ok || label == "" {
		return mcp.NewToolResultError("label is required"), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, err
	}

	// Build instance create options
	createOpts := linodego.InstanceCreateOptions{
		Region: region,
		Type:   instanceType,
		Label:  label,
	}

	// Optional parameters
	if image, ok := arguments["image"].(string); ok && image != "" {
		createOpts.Image = image
	}

	if rootPass, ok := arguments["root_pass"].(string); ok && rootPass != "" {
		createOpts.RootPass = rootPass
	}

	if authorizedKeys, ok := arguments["authorized_keys"].([]interface{}); ok {
		keys := make([]string, 0, len(authorizedKeys))
		for _, key := range authorizedKeys {
			if k, ok := key.(string); ok {
				keys = append(keys, k)
			}
		}
		createOpts.AuthorizedKeys = keys
	}

	if stackscriptID, ok := arguments["stackscript_id"].(float64); ok && stackscriptID > 0 {
		createOpts.StackScriptID = int(stackscriptID)
	}

	if backupsEnabled, ok := arguments["backups_enabled"].(bool); ok {
		createOpts.BackupsEnabled = backupsEnabled
	}

	if privateIP, ok := arguments["private_ip"].(bool); ok {
		createOpts.PrivateIP = privateIP
	}

	if tags, ok := arguments["tags"].([]interface{}); ok {
		tagList := make([]string, 0, len(tags))
		for _, tag := range tags {
			if t, ok := tag.(string); ok {
				tagList = append(tagList, t)
			}
		}
		createOpts.Tags = tagList
	}

	// Create the instance
	instance, err := account.Client.CreateInstance(ctx, createOpts)
	if err != nil {
		return nil, types.NewToolError("linode", "instance_create",
			"failed to create instance", err)
	}

	// Format IPv4 addresses
	ipv4Addresses := make([]string, 0)
	for _, ip := range instance.IPv4 {
		ipv4Addresses = append(ipv4Addresses, ip.String())
	}

	resultText := fmt.Sprintf(`Instance created successfully!

ID: %d
Label: %s
Status: %s
Region: %s
Type: %s
IPv4: %s
IPv6: %s

The instance is now being provisioned. Use linode_instance_get with ID %d to check its status.`,
		instance.ID,
		instance.Label,
		instance.Status,
		instance.Region,
		instance.Type,
		strings.Join(ipv4Addresses, ", "),
		instance.IPv6,
		instance.ID,
	)

	return mcp.NewToolResultText(resultText), nil
}

func (s *Service) handleInstanceDelete(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()
	instanceID, ok := arguments["instance_id"].(float64)
	if !ok || instanceID <= 0 {
		return mcp.NewToolResultError("instance_id is required and must be a positive number"), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, err
	}

	// Get instance details first to show what we're deleting
	instance, err := account.Client.GetInstance(ctx, int(instanceID))
	if err != nil {
		return nil, types.NewToolError("linode", "instance_delete",
			fmt.Sprintf("failed to get instance %d", int(instanceID)), err)
	}

	// Delete the instance
	err = account.Client.DeleteInstance(ctx, int(instanceID))
	if err != nil {
		return nil, types.NewToolError("linode", "instance_delete",
			fmt.Sprintf("failed to delete instance %d", int(instanceID)), err)
	}

	resultText := fmt.Sprintf(`Instance deleted successfully!

Deleted Instance:
- ID: %d
- Label: %s
- Region: %s
- Type: %s

The instance and all its disks have been permanently deleted.`,
		instance.ID,
		instance.Label,
		instance.Region,
		instance.Type,
	)

	s.logger.Info("Deleted Linode instance",
		"instance_id", instance.ID,
		"label", instance.Label,
	)

	return mcp.NewToolResultText(resultText), nil
}

func (s *Service) handleInstanceBoot(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()
	instanceID, ok := arguments["instance_id"].(float64)
	if !ok || instanceID <= 0 {
		return mcp.NewToolResultError("instance_id is required and must be a positive number"), nil
	}

	var configID int
	if cID, ok := arguments["config_id"].(float64); ok && cID > 0 {
		configID = int(cID)
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, err
	}

	// Boot the instance
	err = account.Client.BootInstance(ctx, int(instanceID), configID)
	if err != nil {
		return nil, types.NewToolError("linode", "instance_boot",
			fmt.Sprintf("failed to boot instance %d", int(instanceID)), err)
	}

	// Get updated instance status
	instance, err := account.Client.GetInstance(ctx, int(instanceID))
	if err != nil {
		return nil, types.NewToolError("linode", "instance_boot",
			"failed to get updated instance status", err)
	}

	resultText := fmt.Sprintf(`Instance boot initiated successfully!

Instance: %s (ID: %d)
Status: %s
Region: %s

The instance is now booting up.`,
		instance.Label,
		instance.ID,
		instance.Status,
		instance.Region,
	)

	return mcp.NewToolResultText(resultText), nil
}

func (s *Service) handleInstanceShutdown(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()
	instanceID, ok := arguments["instance_id"].(float64)
	if !ok || instanceID <= 0 {
		return mcp.NewToolResultError("instance_id is required and must be a positive number"), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, err
	}

	// Shutdown the instance
	err = account.Client.ShutdownInstance(ctx, int(instanceID))
	if err != nil {
		return nil, types.NewToolError("linode", "instance_shutdown",
			fmt.Sprintf("failed to shutdown instance %d", int(instanceID)), err)
	}

	// Get updated instance status
	instance, err := account.Client.GetInstance(ctx, int(instanceID))
	if err != nil {
		return nil, types.NewToolError("linode", "instance_shutdown",
			"failed to get updated instance status", err)
	}

	resultText := fmt.Sprintf(`Instance shutdown initiated successfully!

Instance: %s (ID: %d)
Status: %s
Region: %s

The instance is now shutting down.`,
		instance.Label,
		instance.ID,
		instance.Status,
		instance.Region,
	)

	return mcp.NewToolResultText(resultText), nil
}

func (s *Service) handleInstanceReboot(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()
	instanceID, ok := arguments["instance_id"].(float64)
	if !ok || instanceID <= 0 {
		return mcp.NewToolResultError("instance_id is required and must be a positive number"), nil
	}

	var configID int
	if cID, ok := arguments["config_id"].(float64); ok && cID > 0 {
		configID = int(cID)
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, err
	}

	// Reboot the instance
	err = account.Client.RebootInstance(ctx, int(instanceID), configID)
	if err != nil {
		return nil, types.NewToolError("linode", "instance_reboot",
			fmt.Sprintf("failed to reboot instance %d", int(instanceID)), err)
	}

	// Get updated instance status
	instance, err := account.Client.GetInstance(ctx, int(instanceID))
	if err != nil {
		return nil, types.NewToolError("linode", "instance_reboot",
			"failed to get updated instance status", err)
	}

	resultText := fmt.Sprintf(`Instance reboot initiated successfully!

Instance: %s (ID: %d)
Status: %s
Region: %s

The instance is now rebooting.`,
		instance.Label,
		instance.ID,
		instance.Status,
		instance.Region,
	)

	return mcp.NewToolResultText(resultText), nil
}
