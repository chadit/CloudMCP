package linode

import (
	"context"
	"fmt"
	"strings"

	"github.com/linode/linodego"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/chadit/CloudMCP/pkg/types"
)

// handleListImages lists all available images.
func (s *Service) handleListImages(ctx context.Context, params ImagesListParams) (*ImagesListResult, error) {
	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, types.NewToolError("linode", "list_images", "failed to get current account", err) //nolint:wrapcheck // types.NewToolError already wraps the error
	}

	images, err := account.Client.ListImages(ctx, nil)
	if err != nil {
		return nil, types.NewToolError("linode", "list_images", "failed to list images", err) //nolint:wrapcheck // types.NewToolError already wraps the error
	}

	result := &ImagesListResult{
		Images: make([]ImageSummary, 0, len(images)),
		Count:  len(images),
	}

	for _, img := range images {
		// Apply filter if specified
		if params.IsPublic != nil && img.IsPublic != *params.IsPublic {
			continue
		}

		summary := ImageSummary{
			ID:           img.ID,
			Label:        img.Label,
			Description:  img.Description,
			Created:      img.Created.Format("2006-01-02T15:04:05"),
			CreatedBy:    img.CreatedBy,
			Deprecated:   img.Deprecated,
			IsPublic:     img.IsPublic,
			Size:         img.Size,
			Type:         img.Type,
			Vendor:       img.Vendor,
			Status:       string(img.Status),
			TotalSize:    img.TotalSize,
			Capabilities: img.Capabilities,
		}

		// Convert regions
		if img.Regions != nil {
			summary.Regions = make([]ImageRegion, len(img.Regions))
			for i, r := range img.Regions {
				summary.Regions[i] = ImageRegion{
					Region: r.Region,
					Status: string(r.Status),
				}
			}
		}

		// Convert tags
		if len(img.Tags) > 0 {
			summary.Tags = img.Tags
		}

		result.Images = append(result.Images, summary)
	}

	// Update count after filtering
	result.Count = len(result.Images)

	return result, nil
}

// handleGetImage retrieves details of a specific image.
func (s *Service) handleGetImage(ctx context.Context, params ImageGetParams) (*ImageDetail, error) {
	// Validate input
	if params.ImageID == "" {
		return nil, types.NewToolError("linode", "get_image", "image ID cannot be empty", nil) //nolint:wrapcheck // types.NewToolError already wraps the error
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, types.NewToolError("linode", "get_image", "failed to get current account", err) //nolint:wrapcheck // types.NewToolError already wraps the error
	}

	image, err := account.Client.GetImage(ctx, params.ImageID)
	if err != nil {
		return nil, types.NewToolError("linode", "get_image", //nolint:wrapcheck // types.NewToolError already wraps the error
			"failed to get image "+params.ImageID, err)
	}

	if image == nil {
		return nil, types.NewToolError("linode", "get_image", //nolint:wrapcheck // types.NewToolError already wraps the error
			"image "+params.ImageID+" not found", nil)
	}

	detail := &ImageDetail{
		ID:           image.ID,
		Label:        image.Label,
		Description:  image.Description,
		CreatedBy:    image.CreatedBy,
		Deprecated:   image.Deprecated,
		IsPublic:     image.IsPublic,
		Size:         image.Size,
		Type:         image.Type,
		Vendor:       image.Vendor,
		Status:       string(image.Status),
		TotalSize:    image.TotalSize,
		Capabilities: image.Capabilities,
	}

	// Handle Created field
	if image.Created != nil {
		detail.Created = image.Created.Format("2006-01-02T15:04:05")
	}

	// Add optional fields
	if image.Updated != nil {
		detail.Updated = image.Updated.Format("2006-01-02T15:04:05")
	}

	if image.Expiry != nil {
		expiry := image.Expiry.Format("2006-01-02T15:04:05")
		detail.Expiry = &expiry
	}

	// Convert regions
	if image.Regions != nil {
		detail.Regions = make([]ImageRegion, len(image.Regions))
		for i, r := range image.Regions {
			detail.Regions[i] = ImageRegion{
				Region: r.Region,
				Status: string(r.Status),
			}
		}
	}

	// Convert tags
	if len(image.Tags) > 0 {
		detail.Tags = image.Tags
	}

	return detail, nil
}

// handleCreateImage creates a new image from a Linode disk.
func (s *Service) handleCreateImage(ctx context.Context, params ImageCreateParams) (*ImageDetail, error) {
	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, types.NewToolError("linode", "create_image", "failed to get current account", err) //nolint:wrapcheck // types.NewToolError already wraps the error
	}

	createOpts := linodego.ImageCreateOptions{
		DiskID:      params.DiskID,
		Label:       params.Label,
		Description: params.Description,
	}

	createOpts.CloudInit = params.CloudInit

	image, err := account.Client.CreateImage(ctx, createOpts)
	if err != nil {
		return nil, types.NewToolError("linode", "create_image", //nolint:wrapcheck // types.NewToolError already wraps the error
			fmt.Sprintf("failed to create image from disk %d", params.DiskID), err)
	}

	// Wait for initial image creation to complete
	s.logger.Info("image creation initiated", "image_id", image.ID, "label", image.Label)

	// Convert to our detail type
	detail := &ImageDetail{
		ID:           image.ID,
		Label:        image.Label,
		Description:  image.Description,
		Created:      image.Created.Format("2006-01-02T15:04:05"),
		CreatedBy:    image.CreatedBy,
		Deprecated:   image.Deprecated,
		IsPublic:     image.IsPublic,
		Size:         image.Size,
		Type:         image.Type,
		Vendor:       image.Vendor,
		Status:       string(image.Status),
		TotalSize:    image.TotalSize,
		Capabilities: image.Capabilities,
	}

	// Add optional fields
	if image.Updated != nil {
		detail.Updated = image.Updated.Format("2006-01-02T15:04:05")
	}

	if image.Expiry != nil {
		expiry := image.Expiry.Format("2006-01-02T15:04:05")
		detail.Expiry = &expiry
	}

	// Convert regions
	if image.Regions != nil {
		detail.Regions = make([]ImageRegion, len(image.Regions))
		for i, r := range image.Regions {
			detail.Regions[i] = ImageRegion{
				Region: r.Region,
				Status: string(r.Status),
			}
		}
	}

	// Apply tags after creation if provided
	if len(params.Tags) > 0 {
		// Linodego doesn't support tags on creation, need to update after
		_, err = account.Client.UpdateImage(ctx, image.ID, linodego.ImageUpdateOptions{
			Tags: &params.Tags,
		})
		if err != nil {
			// Log but don't fail the operation
			s.logger.Warn("failed to apply tags to new image", "image_id", image.ID, "error", err)
		} else {
			detail.Tags = params.Tags
		}
	}

	return detail, nil
}

// handleUpdateImage updates an existing image.
func (s *Service) handleUpdateImage(ctx context.Context, params ImageUpdateParams) (*ImageDetail, error) {
	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, types.NewToolError("linode", "update_image", "failed to get current account", err) //nolint:wrapcheck // types.NewToolError already wraps the error
	}

	updateOpts := linodego.ImageUpdateOptions{}

	if params.Label != "" {
		updateOpts.Label = params.Label
	}

	if params.Description != "" {
		updateOpts.Description = &params.Description
	}

	if len(params.Tags) > 0 {
		updateOpts.Tags = &params.Tags
	}

	image, err := account.Client.UpdateImage(ctx, params.ImageID, updateOpts)
	if err != nil {
		return nil, types.NewToolError("linode", "update_image", //nolint:wrapcheck // types.NewToolError already wraps the error
			"failed to update image "+params.ImageID, err)
	}

	// Convert to our detail type
	detail := &ImageDetail{
		ID:           image.ID,
		Label:        image.Label,
		Description:  image.Description,
		Created:      image.Created.Format("2006-01-02T15:04:05"),
		CreatedBy:    image.CreatedBy,
		Deprecated:   image.Deprecated,
		IsPublic:     image.IsPublic,
		Size:         image.Size,
		Type:         image.Type,
		Vendor:       image.Vendor,
		Status:       string(image.Status),
		TotalSize:    image.TotalSize,
		Capabilities: image.Capabilities,
	}

	// Add optional fields
	if image.Updated != nil {
		detail.Updated = image.Updated.Format("2006-01-02T15:04:05")
	}

	if image.Expiry != nil {
		expiry := image.Expiry.Format("2006-01-02T15:04:05")
		detail.Expiry = &expiry
	}

	// Convert regions
	if image.Regions != nil {
		detail.Regions = make([]ImageRegion, len(image.Regions))
		for i, r := range image.Regions {
			detail.Regions[i] = ImageRegion{
				Region: r.Region,
				Status: string(r.Status),
			}
		}
	}

	// Convert tags
	if len(image.Tags) > 0 {
		detail.Tags = image.Tags
	}

	return detail, nil
}

// handleDeleteImage deletes a custom image.
func (s *Service) handleDeleteImage(ctx context.Context, params ImageDeleteParams) (string, error) {
	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return "", types.NewToolError("linode", "delete_image", "failed to get current account", err) //nolint:wrapcheck // types.NewToolError already wraps the error
	}

	// Check if image exists and is private
	image, err := account.Client.GetImage(ctx, params.ImageID)
	if err != nil {
		return "", types.NewToolError("linode", "delete_image", //nolint:wrapcheck // types.NewToolError already wraps the error
			"failed to get image "+params.ImageID, err)
	}

	if image.IsPublic {
		return "", types.NewToolError("linode", "delete_image", //nolint:wrapcheck // types.NewToolError already wraps the error
			"cannot delete public images", nil)
	}

	err = account.Client.DeleteImage(ctx, params.ImageID)
	if err != nil {
		return "", types.NewToolError("linode", "delete_image", //nolint:wrapcheck // types.NewToolError already wraps the error
			"failed to delete image "+params.ImageID, err)
	}

	return fmt.Sprintf("Successfully deleted image %s (%s)", params.ImageID, image.Label), nil
}

// handleReplicateImage replicates an image to additional regions.
func (s *Service) handleReplicateImage(ctx context.Context, params ImageReplicateParams) (*ImageDetail, error) {
	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, types.NewToolError("linode", "replicate_image", "failed to get current account", err) //nolint:wrapcheck // types.NewToolError already wraps the error
	}

	// Validate the image exists and is private
	existingImage, err := account.Client.GetImage(ctx, params.ImageID)
	if err != nil {
		return nil, types.NewToolError("linode", "replicate_image", //nolint:wrapcheck // types.NewToolError already wraps the error
			"failed to get image "+params.ImageID, err)
	}

	if existingImage.IsPublic {
		return nil, types.NewToolError("linode", "replicate_image", //nolint:wrapcheck // types.NewToolError already wraps the error
			"cannot replicate public images", nil)
	}

	// Check which regions already have the image
	existingRegions := make(map[string]bool)
	for _, r := range existingImage.Regions {
		existingRegions[r.Region] = true
	}

	// Filter out regions that already have the image
	newRegions := make([]string, 0)

	for _, region := range params.Regions {
		if !existingRegions[region] {
			newRegions = append(newRegions, region)
		}
	}

	if len(newRegions) == 0 {
		return nil, types.NewToolError("linode", "replicate_image", //nolint:wrapcheck // types.NewToolError already wraps the error
			"image already exists in all specified regions", nil)
	}

	// Replicate to new regions
	image, err := account.Client.ReplicateImage(ctx, params.ImageID, linodego.ImageReplicateOptions{
		Regions: newRegions,
	})
	if err != nil {
		return nil, types.NewToolError("linode", "replicate_image", //nolint:wrapcheck // types.NewToolError already wraps the error
			fmt.Sprintf("failed to replicate image %s to regions %s",
				params.ImageID, strings.Join(newRegions, ", ")), err)
	}

	s.logger.Info("image replication initiated",
		"image_id", image.ID,
		"new_regions", strings.Join(newRegions, ", "))

	// Convert to our detail type
	detail := &ImageDetail{
		ID:           image.ID,
		Label:        image.Label,
		Description:  image.Description,
		Created:      image.Created.Format("2006-01-02T15:04:05"),
		CreatedBy:    image.CreatedBy,
		Deprecated:   image.Deprecated,
		IsPublic:     image.IsPublic,
		Size:         image.Size,
		Type:         image.Type,
		Vendor:       image.Vendor,
		Status:       string(image.Status),
		TotalSize:    image.TotalSize,
		Capabilities: image.Capabilities,
	}

	// Add optional fields
	if image.Updated != nil {
		detail.Updated = image.Updated.Format("2006-01-02T15:04:05")
	}

	if image.Expiry != nil {
		expiry := image.Expiry.Format("2006-01-02T15:04:05")
		detail.Expiry = &expiry
	}

	// Convert regions
	if image.Regions != nil {
		detail.Regions = make([]ImageRegion, len(image.Regions))
		for i, r := range image.Regions {
			detail.Regions[i] = ImageRegion{
				Region: r.Region,
				Status: string(r.Status),
			}
		}
	}

	// Convert tags
	if len(image.Tags) > 0 {
		detail.Tags = image.Tags
	}

	return detail, nil
}

// handleCreateImageUpload initiates an image upload.
func (s *Service) handleCreateImageUpload(ctx context.Context, params ImageUploadParams) (*ImageUploadResult, error) {
	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, types.NewToolError("linode", "create_image_upload", "failed to get current account", err) //nolint:wrapcheck // types.NewToolError already wraps the error
	}

	uploadOpts := linodego.ImageCreateUploadOptions{
		Label:       params.Label,
		Region:      params.Region,
		Description: params.Description,
	}

	uploadOpts.CloudInit = params.CloudInit

	if len(params.Tags) > 0 {
		uploadOpts.Tags = &params.Tags
	}

	image, uploadURL, err := account.Client.CreateImageUpload(ctx, uploadOpts)
	if err != nil {
		return nil, types.NewToolError("linode", "create_image_upload", //nolint:wrapcheck // types.NewToolError already wraps the error
			"failed to create image upload", err)
	}

	return &ImageUploadResult{
		ImageID:  image.ID,
		UploadTo: uploadURL,
	}, nil
}

// MCP Handler wrappers.
func (s *Service) handleImagesList(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()

	// Parse optional is_public filter
	var isPublic *bool
	if val, ok := arguments["is_public"].(bool); ok {
		isPublic = &val
	}

	params := ImagesListParams{
		IsPublic: isPublic,
	}

	result, err := s.handleListImages(ctx, params)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(formatImagesListResult(result)), nil
}

func (s *Service) handleImageGet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()

	imageID, imageIDPresent := arguments["image_id"].(string)
	if !imageIDPresent || imageID == "" {
		return mcp.NewToolResultError("image_id is required"), nil
	}

	params := ImageGetParams{
		ImageID: imageID,
	}

	result, err := s.handleGetImage(ctx, params)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(formatImageDetail(result)), nil
}

func (s *Service) handleImageCreate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()

	diskIDFloat, diskIDPresent := arguments["disk_id"].(float64)
	if !diskIDPresent {
		return mcp.NewToolResultError("disk_id is required"), nil
	}

	diskID := int(diskIDFloat)

	label, labelPresent := arguments["label"].(string)
	if !labelPresent || label == "" {
		return mcp.NewToolResultError("label is required"), nil
	}

	params := ImageCreateParams{
		DiskID: diskID,
		Label:  label,
	}

	if desc, ok := arguments["description"].(string); ok {
		params.Description = desc
	}

	if ci, ok := arguments["cloud_init"].(bool); ok {
		params.CloudInit = ci
	}

	if tagsInterface, ok := arguments["tags"].([]interface{}); ok {
		tags := make([]string, len(tagsInterface))

		for i, tag := range tagsInterface {
			if tagStr, ok := tag.(string); ok {
				tags[i] = tagStr
			}
		}

		params.Tags = tags
	}

	result, err := s.handleCreateImage(ctx, params)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(formatImageDetail(result)), nil
}

func (s *Service) handleImageUpdate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()

	imageID, imageIDPresent := arguments["image_id"].(string)
	if !imageIDPresent || imageID == "" {
		return mcp.NewToolResultError("image_id is required"), nil
	}

	params := ImageUpdateParams{
		ImageID: imageID,
	}

	if label, ok := arguments["label"].(string); ok {
		params.Label = label
	}

	if desc, ok := arguments["description"].(string); ok {
		params.Description = desc
	}

	if tagsInterface, ok := arguments["tags"].([]interface{}); ok {
		tags := make([]string, len(tagsInterface))

		for i, tag := range tagsInterface {
			if tagStr, ok := tag.(string); ok {
				tags[i] = tagStr
			}
		}

		params.Tags = tags
	}

	result, err := s.handleUpdateImage(ctx, params)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(formatImageDetail(result)), nil
}

func (s *Service) handleImageDelete(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()

	imageID, imageIDPresent := arguments["image_id"].(string)
	if !imageIDPresent || imageID == "" {
		return mcp.NewToolResultError("image_id is required"), nil
	}

	params := ImageDeleteParams{
		ImageID: imageID,
	}

	result, err := s.handleDeleteImage(ctx, params)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(result), nil
}

func (s *Service) handleImageReplicate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()

	imageID, imageIDPresent := arguments["image_id"].(string)
	if !imageIDPresent || imageID == "" {
		return mcp.NewToolResultError("image_id is required"), nil
	}

	regionsInterface, ok := arguments["regions"].([]interface{})
	if !ok || len(regionsInterface) == 0 {
		return mcp.NewToolResultError("regions is required"), nil
	}

	regions := make([]string, len(regionsInterface))

	for i, region := range regionsInterface {
		if regionStr, ok := region.(string); ok {
			regions[i] = regionStr
		}
	}

	params := ImageReplicateParams{
		ImageID: imageID,
		Regions: regions,
	}

	result, err := s.handleReplicateImage(ctx, params)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(formatImageDetail(result)), nil
}

func (s *Service) handleImageUploadCreate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()

	label, labelPresent := arguments["label"].(string)
	if !labelPresent || label == "" {
		return mcp.NewToolResultError("label is required"), nil
	}

	region, ok := arguments["region"].(string)
	if !ok || region == "" {
		return mcp.NewToolResultError("region is required"), nil
	}

	params := ImageUploadParams{
		Label:  label,
		Region: region,
	}

	if desc, ok := arguments["description"].(string); ok {
		params.Description = desc
	}

	if ci, ok := arguments["cloud_init"].(bool); ok {
		params.CloudInit = ci
	}

	if tagsInterface, ok := arguments["tags"].([]interface{}); ok {
		tags := make([]string, len(tagsInterface))

		for i, tag := range tagsInterface {
			if tagStr, ok := tag.(string); ok {
				tags[i] = tagStr
			}
		}

		params.Tags = tags
	}

	result, err := s.handleCreateImageUpload(ctx, params)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(fmt.Sprintf("Image upload created: %s\nUpload URL: %s", result.ImageID, result.UploadTo)), nil
}

// Formatting functions.
func formatImagesListResult(result *ImagesListResult) string {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("Found %d images:\n\n", result.Count))

	for _, img := range result.Images {
		output.WriteString(fmt.Sprintf("ID: %s\n", img.ID))
		output.WriteString(fmt.Sprintf("Label: %s\n", img.Label))
		output.WriteString(fmt.Sprintf("Description: %s\n", img.Description))
		output.WriteString(fmt.Sprintf("Type: %s\n", img.Type))
		output.WriteString(fmt.Sprintf("Status: %s\n", img.Status))
		output.WriteString(fmt.Sprintf("Size: %d MB\n", img.Size))
		output.WriteString(fmt.Sprintf("Public: %t\n", img.IsPublic))
		output.WriteString(fmt.Sprintf("Created: %s\n", img.Created))

		if len(img.Regions) > 0 {
			output.WriteString("Regions:\n")

			for _, region := range img.Regions {
				output.WriteString(fmt.Sprintf("  %s: %s\n", region.Region, region.Status))
			}
		}

		if len(img.Tags) > 0 {
			output.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(img.Tags, ", ")))
		}

		output.WriteString("\n")
	}

	return output.String()
}

func formatImageDetail(detail *ImageDetail) string {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("Image: %s (%s)\n", detail.ID, detail.Label))
	output.WriteString(fmt.Sprintf("Description: %s\n", detail.Description))
	output.WriteString(fmt.Sprintf("Type: %s\n", detail.Type))
	output.WriteString(fmt.Sprintf("Status: %s\n", detail.Status))
	output.WriteString(fmt.Sprintf("Size: %d MB\n", detail.Size))
	output.WriteString(fmt.Sprintf("Total Size: %d MB\n", detail.TotalSize))
	output.WriteString(fmt.Sprintf("Public: %t\n", detail.IsPublic))
	output.WriteString(fmt.Sprintf("Deprecated: %t\n", detail.Deprecated))
	output.WriteString(fmt.Sprintf("Created: %s\n", detail.Created))
	output.WriteString(fmt.Sprintf("Created By: %s\n", detail.CreatedBy))

	if detail.Updated != "" {
		output.WriteString(fmt.Sprintf("Updated: %s\n", detail.Updated))
	}

	if detail.Expiry != nil {
		output.WriteString(fmt.Sprintf("Expires: %s\n", *detail.Expiry))
	}

	if len(detail.Regions) > 0 {
		output.WriteString("Regions:\n")

		for _, region := range detail.Regions {
			output.WriteString(fmt.Sprintf("  %s: %s\n", region.Region, region.Status))
		}
	}

	if len(detail.Capabilities) > 0 {
		output.WriteString(fmt.Sprintf("Capabilities: %s\n", strings.Join(detail.Capabilities, ", ")))
	}

	if len(detail.Tags) > 0 {
		output.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(detail.Tags, ", ")))
	}

	return output.String()
}
