package linode

import (
	"context"
	"fmt"
	"strings"

	"github.com/linode/linodego"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/chadit/CloudMCP/pkg/types"
)

func (s *Service) handleVolumesList(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, err
	}

	volumes, err := account.Client.ListVolumes(ctx, nil)
	if err != nil {
		return nil, types.NewToolError("linode", "volumes_list",
			"failed to list volumes", err)
	}

	if len(volumes) == 0 {
		return mcp.NewToolResultText("No volumes found."), nil
	}

	var resultText string
	resultText = fmt.Sprintf("Found %d volume(s):\n\n", len(volumes))
	for _, vol := range volumes {
		resultText += fmt.Sprintf("ID: %d | %s\n", vol.ID, vol.Label)
		resultText += fmt.Sprintf("  Status: %s | Size: %d GB | Region: %s\n", vol.Status, vol.Size, vol.Region)
		if vol.LinodeID != nil && *vol.LinodeID > 0 {
			resultText += fmt.Sprintf("  Attached to: Linode %d", *vol.LinodeID)
			if vol.LinodeLabel != "" {
				resultText += fmt.Sprintf(" (%s)", vol.LinodeLabel)
			}
			resultText += "\n"
			if vol.FilesystemPath != "" {
				resultText += fmt.Sprintf("  Mount Path: %s\n", vol.FilesystemPath)
			}
		} else {
			resultText += "  Status: Unattached\n"
		}
		if len(vol.Tags) > 0 {
			resultText += fmt.Sprintf("  Tags: %s\n", strings.Join(vol.Tags, ", "))
		}
		resultText += "\n"
	}

	return mcp.NewToolResultText(resultText), nil
}

func (s *Service) handleVolumeGet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()
	volumeID, ok := arguments["volume_id"].(float64)
	if !ok || volumeID <= 0 {
		return mcp.NewToolResultError("volume_id is required and must be a positive number"), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, err
	}

	volume, err := account.Client.GetVolume(ctx, int(volumeID))
	if err != nil {
		return nil, types.NewToolError("linode", "volume_get",
			fmt.Sprintf("failed to get volume %d", int(volumeID)), err)
	}

	// Format the detailed volume information
	resultText := fmt.Sprintf(`Volume Details:
ID: %d
Label: %s
Status: %s
Size: %d GB
Region: %s

Created: %s
Updated: %s`,
		volume.ID,
		volume.Label,
		volume.Status,
		volume.Size,
		volume.Region,
		volume.Created.Format("2006-01-02 15:04:05"),
		volume.Updated.Format("2006-01-02 15:04:05"),
	)

	if volume.LinodeID != nil && *volume.LinodeID > 0 {
		resultText += fmt.Sprintf("\n\nAttached to Linode: %d", *volume.LinodeID)
		if volume.LinodeLabel != "" {
			resultText += fmt.Sprintf(" (%s)", volume.LinodeLabel)
		}
		if volume.FilesystemPath != "" {
			resultText += fmt.Sprintf("\nMount Path: %s", volume.FilesystemPath)
		}
	} else {
		resultText += "\n\nStatus: Unattached"
	}

	if len(volume.Tags) > 0 {
		resultText += fmt.Sprintf("\n\nTags: %s", strings.Join(volume.Tags, ", "))
	}

	return mcp.NewToolResultText(resultText), nil
}

func (s *Service) handleVolumeCreate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()

	// Required parameters
	label, ok := arguments["label"].(string)
	if !ok || label == "" {
		return mcp.NewToolResultError("label is required"), nil
	}

	size, ok := arguments["size"].(float64)
	if !ok || size < 10 || size > 8192 {
		return mcp.NewToolResultError("size is required and must be between 10 and 8192 GB"), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, err
	}

	// Build volume create options
	createOpts := linodego.VolumeCreateOptions{
		Label: label,
		Size:  int(size),
	}

	// Optional parameters
	if region, ok := arguments["region"].(string); ok && region != "" {
		createOpts.Region = region
	}

	if linodeID, ok := arguments["linode_id"].(float64); ok && linodeID > 0 {
		id := int(linodeID)
		createOpts.LinodeID = id
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

	// Create the volume
	volume, err := account.Client.CreateVolume(ctx, createOpts)
	if err != nil {
		return nil, types.NewToolError("linode", "volume_create",
			"failed to create volume", err)
	}

	resultText := fmt.Sprintf(`Volume created successfully!

ID: %d
Label: %s
Size: %d GB
Region: %s
Status: %s`,
		volume.ID,
		volume.Label,
		volume.Size,
		volume.Region,
		volume.Status,
	)

	if volume.LinodeID != nil && *volume.LinodeID > 0 {
		resultText += fmt.Sprintf("\nAttached to Linode: %d", *volume.LinodeID)
		if volume.FilesystemPath != "" {
			resultText += fmt.Sprintf("\nMount Path: %s", volume.FilesystemPath)
		}
	}

	return mcp.NewToolResultText(resultText), nil
}

func (s *Service) handleVolumeDelete(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()
	volumeID, ok := arguments["volume_id"].(float64)
	if !ok || volumeID <= 0 {
		return mcp.NewToolResultError("volume_id is required and must be a positive number"), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, err
	}

	// Get volume details first to show what we're deleting
	volume, err := account.Client.GetVolume(ctx, int(volumeID))
	if err != nil {
		return nil, types.NewToolError("linode", "volume_delete",
			fmt.Sprintf("failed to get volume %d", int(volumeID)), err)
	}

	// Delete the volume
	err = account.Client.DeleteVolume(ctx, int(volumeID))
	if err != nil {
		return nil, types.NewToolError("linode", "volume_delete",
			fmt.Sprintf("failed to delete volume %d", int(volumeID)), err)
	}

	resultText := fmt.Sprintf(`Volume deleted successfully!

Deleted Volume:
- ID: %d
- Label: %s
- Size: %d GB
- Region: %s

The volume has been permanently deleted.`,
		volume.ID,
		volume.Label,
		volume.Size,
		volume.Region,
	)

	s.logger.Info("Deleted volume",
		"volume_id", volume.ID,
		"label", volume.Label,
	)

	return mcp.NewToolResultText(resultText), nil
}

func (s *Service) handleVolumeAttach(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()

	volumeID, ok := arguments["volume_id"].(float64)
	if !ok || volumeID <= 0 {
		return mcp.NewToolResultError("volume_id is required and must be a positive number"), nil
	}

	linodeID, ok := arguments["linode_id"].(float64)
	if !ok || linodeID <= 0 {
		return mcp.NewToolResultError("linode_id is required and must be a positive number"), nil
	}

	persistAcrossBoots := true // Default to true
	if persist, ok := arguments["persist_across_boots"].(bool); ok {
		persistAcrossBoots = persist
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, err
	}

	// Attach the volume
	attachOpts := linodego.VolumeAttachOptions{
		LinodeID:           int(linodeID),
		PersistAcrossBoots: &persistAcrossBoots,
	}

	volume, err := account.Client.AttachVolume(ctx, int(volumeID), &attachOpts)
	if err != nil {
		return nil, types.NewToolError("linode", "volume_attach",
			fmt.Sprintf("failed to attach volume %d to instance %d", int(volumeID), int(linodeID)), err)
	}

	resultText := fmt.Sprintf(`Volume attached successfully!

Volume: %s (ID: %d)
Attached to Linode: %d
Mount Path: %s
Persist Across Boots: %v

To mount the volume, SSH into your Linode and run:
mkdir -p /mnt/%s
mount %s /mnt/%s`,
		volume.Label,
		volume.ID,
		int(linodeID),
		volume.FilesystemPath,
		persistAcrossBoots,
		volume.Label,
		volume.FilesystemPath,
		volume.Label,
	)

	return mcp.NewToolResultText(resultText), nil
}

func (s *Service) handleVolumeDetach(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()
	volumeID, ok := arguments["volume_id"].(float64)
	if !ok || volumeID <= 0 {
		return mcp.NewToolResultError("volume_id is required and must be a positive number"), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, err
	}

	// Get volume details first
	volume, err := account.Client.GetVolume(ctx, int(volumeID))
	if err != nil {
		return nil, types.NewToolError("linode", "volume_detach",
			fmt.Sprintf("failed to get volume %d", int(volumeID)), err)
	}

	var previousLinodeID int
	if volume.LinodeID != nil {
		previousLinodeID = *volume.LinodeID
	}

	// Detach the volume
	err = account.Client.DetachVolume(ctx, int(volumeID))
	if err != nil {
		return nil, types.NewToolError("linode", "volume_detach",
			fmt.Sprintf("failed to detach volume %d", int(volumeID)), err)
	}

	resultText := fmt.Sprintf(`Volume detached successfully!

Volume: %s (ID: %d)
Size: %d GB
Region: %s`,
		volume.Label,
		volume.ID,
		volume.Size,
		volume.Region,
	)

	if previousLinodeID > 0 {
		resultText += fmt.Sprintf("\nDetached from Linode: %d", previousLinodeID)
	}

	resultText += "\n\nThe volume is now available to attach to another Linode instance."

	return mcp.NewToolResultText(resultText), nil
}
