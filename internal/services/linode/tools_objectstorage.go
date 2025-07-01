package linode

import (
	"context"
	"fmt"
	"strings"

	"github.com/linode/linodego"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	bucketAccessFull    = "Full Access"
	bucketAccessLimited = "Limited Access"

	// Byte conversion constants.
	bytesPerKB = 1024
	bytesPerMB = 1024 * 1024
	bytesPerGB = 1024 * 1024 * 1024
)

// handleObjectStorageBucketsList lists all Object Storage buckets.
func (s *Service) handleObjectStorageBucketsList(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	buckets, err := account.Client.ListObjectStorageBuckets(ctx, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list Object Storage buckets: %v", err)), nil
	}

	summaries := make([]ObjectStorageBucketSummary, 0, len(buckets))

	for _, bucket := range buckets {
		summary := ObjectStorageBucketSummary{
			Label:    bucket.Label,
			Region:   bucket.Region,
			Hostname: bucket.Hostname,
			Created:  bucket.Created.Format("2006-01-02T15:04:05"),
			Size:     int64(bucket.Size),
			Objects:  bucket.Objects,
		}
		summaries = append(summaries, summary)
	}

	// Remove unused result variable

	var stringBuilder strings.Builder

	stringBuilder.WriteString(fmt.Sprintf("Found %d Object Storage buckets:\n\n", len(summaries)))

	for _, bucket := range summaries {
		sizeDisplay := "0 bytes"

		if bucket.Size > 0 {
			switch {
			case bucket.Size >= bytesPerGB:
				sizeDisplay = fmt.Sprintf("%.2f GB", float64(bucket.Size)/bytesPerGB)
			case bucket.Size >= bytesPerMB:
				sizeDisplay = fmt.Sprintf("%.2f MB", float64(bucket.Size)/bytesPerMB)
			case bucket.Size >= bytesPerKB:
				sizeDisplay = fmt.Sprintf("%.2f KB", float64(bucket.Size)/bytesPerKB)
			default:
				sizeDisplay = fmt.Sprintf("%d bytes", bucket.Size)
			}
		}

		fmt.Fprintf(&stringBuilder, "Name: %s (%s)\n", bucket.Label, bucket.Region)
		fmt.Fprintf(&stringBuilder, "  Hostname: %s\n", bucket.Hostname)
		fmt.Fprintf(&stringBuilder, "  Size: %s | Objects: %d\n", sizeDisplay, bucket.Objects)
		fmt.Fprintf(&stringBuilder, "  Created: %s\n", bucket.Created)
		stringBuilder.WriteString("\n")
	}

	return mcp.NewToolResultText(stringBuilder.String()), nil
}

// handleObjectStorageBucketGet gets details of a specific Object Storage bucket.
func (s *Service) handleObjectStorageBucketGet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params ObjectStorageBucketGetParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	bucket, err := account.Client.GetObjectStorageBucket(ctx, params.Region, params.Bucket)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Object Storage bucket: %v", err)), nil
	}

	detail := ObjectStorageBucketDetail{
		Label:    bucket.Label,
		Region:   bucket.Region,
		Hostname: bucket.Hostname,
		Created:  bucket.Created.Format("2006-01-02T15:04:05"),
		Size:     int64(bucket.Size),
		Objects:  bucket.Objects,
	}

	var stringBuilder strings.Builder

	fmt.Fprintf(&stringBuilder, "Object Storage Bucket Details:\n")
	fmt.Fprintf(&stringBuilder, "Name: %s\n", detail.Label)
	fmt.Fprintf(&stringBuilder, "Region: %s\n", detail.Region)
	fmt.Fprintf(&stringBuilder, "Hostname: %s\n", detail.Hostname)
	fmt.Fprintf(&stringBuilder, "Created: %s\n", detail.Created)

	sizeDisplay := "0 bytes"

	if detail.Size > 0 {
		switch {
		case detail.Size >= bytesPerGB:
			sizeDisplay = fmt.Sprintf("%.2f GB", float64(detail.Size)/bytesPerGB)
		case detail.Size >= bytesPerMB:
			sizeDisplay = fmt.Sprintf("%.2f MB", float64(detail.Size)/bytesPerMB)
		case detail.Size >= bytesPerKB:
			sizeDisplay = fmt.Sprintf("%.2f KB", float64(detail.Size)/bytesPerKB)
		default:
			sizeDisplay = fmt.Sprintf("%d bytes", detail.Size)
		}
	}

	fmt.Fprintf(&stringBuilder, "Size: %s\n", sizeDisplay)
	fmt.Fprintf(&stringBuilder, "Objects: %d\n", detail.Objects)

	return mcp.NewToolResultText(stringBuilder.String()), nil
}

// handleObjectStorageBucketCreate creates a new Object Storage bucket.
func (s *Service) handleObjectStorageBucketCreate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params ObjectStorageBucketCreateParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	createOpts := linodego.ObjectStorageBucketCreateOptions{
		Label: params.Label,
	}

	createOpts.Region = params.Region

	if params.ACL != "" {
		createOpts.ACL = linodego.ObjectStorageACL(params.ACL)
	}

	if params.CORS {
		createOpts.CorsEnabled = &params.CORS
	}

	bucket, err := account.Client.CreateObjectStorageBucket(ctx, createOpts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create Object Storage bucket: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Object Storage bucket created successfully:\nName: %s\nRegion: %s\nHostname: %s",
		bucket.Label, bucket.Region, bucket.Hostname)), nil
}

// handleObjectStorageBucketUpdate updates an existing Object Storage bucket's access settings.
func (s *Service) handleObjectStorageBucketUpdate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params ObjectStorageBucketUpdateParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	updateOpts := linodego.ObjectStorageBucketUpdateAccessOptions{}

	if params.ACL != "" {
		updateOpts.ACL = linodego.ObjectStorageACL(params.ACL)
	}

	if params.CORS != nil {
		updateOpts.CorsEnabled = params.CORS
	}

	err = account.Client.UpdateObjectStorageBucketAccess(ctx, params.Region, params.Bucket, updateOpts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update Object Storage bucket access: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Object Storage bucket '%s' access updated successfully in region '%s'",
		params.Bucket, params.Region)), nil
}

// handleObjectStorageBucketDelete deletes an Object Storage bucket.
func (s *Service) handleObjectStorageBucketDelete(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params ObjectStorageBucketDeleteParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	err = account.Client.DeleteObjectStorageBucket(ctx, params.Region, params.Bucket)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete Object Storage bucket: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Object Storage bucket '%s' deleted successfully from region '%s'",
		params.Bucket, params.Region)), nil
}

// handleObjectStorageKeysList lists all Object Storage keys.
func (s *Service) handleObjectStorageKeysList(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	keys, err := account.Client.ListObjectStorageKeys(ctx, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list Object Storage keys: %v", err)), nil
	}

	summaries := make([]ObjectStorageKeySummary, 0, len(keys))

	for _, key := range keys {
		var bucketAccess []ObjectStorageBucketAccess

		if key.BucketAccess != nil {
			bucketAccess = make([]ObjectStorageBucketAccess, 0, len(*key.BucketAccess))
			for _, access := range *key.BucketAccess {
				bucketAccess = append(bucketAccess, ObjectStorageBucketAccess{
					Region:      access.Region,
					BucketName:  access.BucketName,
					Permissions: access.Permissions,
				})
			}
		}

		summary := ObjectStorageKeySummary{
			ID:           key.ID,
			Label:        key.Label,
			AccessKey:    key.AccessKey,
			SecretKey:    "***REDACTED***",
			Limited:      key.Limited,
			BucketAccess: bucketAccess,
		}
		summaries = append(summaries, summary)
	}

	// Remove unused result variable

	var stringBuilder strings.Builder

	stringBuilder.WriteString(fmt.Sprintf("Found %d Object Storage keys:\n\n", len(summaries)))

	for _, key := range summaries {
		accessType := bucketAccessFull
		if key.Limited {
			accessType = bucketAccessLimited
		}

		fmt.Fprintf(&stringBuilder, "ID: %d | %s (%s)\n", key.ID, key.Label, accessType)
		fmt.Fprintf(&stringBuilder, "  Access Key: %s\n", key.AccessKey)

		if key.Limited && len(key.BucketAccess) > 0 {
			fmt.Fprintf(&stringBuilder, "  Bucket Access:\n")

			for _, access := range key.BucketAccess {
				fmt.Fprintf(&stringBuilder, "    - %s/%s: %s\n", access.Region, access.BucketName, access.Permissions)
			}
		}

		stringBuilder.WriteString("\n")
	}

	return mcp.NewToolResultText(stringBuilder.String()), nil
}

// handleObjectStorageKeyGet gets details of a specific Object Storage key.
func (s *Service) handleObjectStorageKeyGet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params ObjectStorageKeyGetParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	key, err := account.Client.GetObjectStorageKey(ctx, params.KeyID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Object Storage key: %v", err)), nil
	}

	var bucketAccess []ObjectStorageBucketAccess

	if key.BucketAccess != nil {
		bucketAccess = make([]ObjectStorageBucketAccess, 0, len(*key.BucketAccess))
		for _, access := range *key.BucketAccess {
			bucketAccess = append(bucketAccess, ObjectStorageBucketAccess{
				Region:      access.Region,
				BucketName:  access.BucketName,
				Permissions: access.Permissions,
			})
		}
	}

	detail := ObjectStorageKeyDetail{
		ID:           key.ID,
		Label:        key.Label,
		AccessKey:    key.AccessKey,
		SecretKey:    "***REDACTED***",
		Limited:      key.Limited,
		BucketAccess: bucketAccess,
	}

	var stringBuilder strings.Builder

	fmt.Fprintf(&stringBuilder, "Object Storage Key Details:\n")
	fmt.Fprintf(&stringBuilder, "ID: %d\n", detail.ID)
	fmt.Fprintf(&stringBuilder, "Label: %s\n", detail.Label)
	fmt.Fprintf(&stringBuilder, "Access Key: %s\n", detail.AccessKey)
	fmt.Fprintf(&stringBuilder, "Secret Key: %s\n", detail.SecretKey)

	accessType := bucketAccessFull
	if detail.Limited {
		accessType = bucketAccessLimited
	}

	fmt.Fprintf(&stringBuilder, "Access Type: %s\n", accessType)

	if detail.Limited && len(detail.BucketAccess) > 0 {
		fmt.Fprintf(&stringBuilder, "\nBucket Access Permissions:\n")

		for _, access := range detail.BucketAccess {
			fmt.Fprintf(&stringBuilder, "  - Region: %s\n", access.Region)
			fmt.Fprintf(&stringBuilder, "    Bucket: %s\n", access.BucketName)
			fmt.Fprintf(&stringBuilder, "    Permissions: %s\n", access.Permissions)
			stringBuilder.WriteString("\n")
		}
	}

	return mcp.NewToolResultText(stringBuilder.String()), nil
}

// handleObjectStorageKeyCreate creates a new Object Storage key.
func (s *Service) handleObjectStorageKeyCreate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params ObjectStorageKeyCreateParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	createOpts := linodego.ObjectStorageKeyCreateOptions{
		Label: params.Label,
	}

	if len(params.BucketAccess) > 0 {
		bucketAccess := make([]linodego.ObjectStorageKeyBucketAccess, 0, len(params.BucketAccess))
		for _, access := range params.BucketAccess {
			bucketAccess = append(bucketAccess, linodego.ObjectStorageKeyBucketAccess{
				Region:      access.Region,
				BucketName:  access.BucketName,
				Permissions: access.Permissions,
			})
		}

		createOpts.BucketAccess = &bucketAccess
	}

	key, err := account.Client.CreateObjectStorageKey(ctx, createOpts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create Object Storage key: %v", err)), nil
	}

	accessType := bucketAccessFull
	if key.Limited {
		accessType = bucketAccessLimited
	}

	return mcp.NewToolResultText(fmt.Sprintf("Object Storage key created successfully:\nID: %d\nLabel: %s\nAccess Key: %s\nSecret Key: %s\nAccess Type: %s",
		key.ID, key.Label, key.AccessKey, key.SecretKey, accessType)), nil
}

// handleObjectStorageKeyUpdate updates an existing Object Storage key.
func (s *Service) handleObjectStorageKeyUpdate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params ObjectStorageKeyUpdateParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	updateOpts := linodego.ObjectStorageKeyUpdateOptions{}

	if params.Label != "" {
		updateOpts.Label = params.Label
	}

	if len(params.BucketAccess) > 0 {
		// BucketAccess field not available in ObjectStorageKeyUpdateOptions
		// Once the Linode API supports updating bucket access, implement this:
		// var bucketAccess []linodego.ObjectStorageKeyBucketAccess
		// for _, access := range params.BucketAccess {
		//     bucketAccess = append(bucketAccess, linodego.ObjectStorageKeyBucketAccess{
		//         Region:      access.Region,
		//         BucketName:  access.BucketName,
		//         Permissions: string(access.Permissions),
		//     })
		// }
		// updateOpts.BucketAccess = &bucketAccess
		_ = params.BucketAccess // Acknowledge parameter until API supports it
	}

	key, err := account.Client.UpdateObjectStorageKey(ctx, params.KeyID, updateOpts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update Object Storage key: %v", err)), nil
	}

	accessType := bucketAccessFull
	if key.Limited {
		accessType = bucketAccessLimited
	}

	return mcp.NewToolResultText(fmt.Sprintf("Object Storage key updated successfully:\nID: %d\nLabel: %s\nAccess Type: %s",
		key.ID, key.Label, accessType)), nil
}

// handleObjectStorageKeyDelete deletes an Object Storage key.
func (s *Service) handleObjectStorageKeyDelete(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params ObjectStorageKeyDeleteParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	err = account.Client.DeleteObjectStorageKey(ctx, params.KeyID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete Object Storage key: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Object Storage key %d deleted successfully", params.KeyID)), nil
}

// handleObjectStorageClustersList lists all Object Storage clusters.
func (s *Service) handleObjectStorageClustersList(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	clusters, err := account.Client.ListObjectStorageClusters(ctx, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list Object Storage clusters: %v", err)), nil
	}

	summaries := make([]ObjectStorageClusterSummary, 0, len(clusters))

	for _, cluster := range clusters {
		summary := ObjectStorageClusterSummary{
			ID:               cluster.ID,
			Domain:           cluster.Domain,
			Status:           cluster.Status,
			Region:           cluster.Region,
			StaticSiteDomain: cluster.StaticSiteDomain,
		}
		summaries = append(summaries, summary)
	}

	// Remove unused result variable

	var stringBuilder strings.Builder

	stringBuilder.WriteString(fmt.Sprintf("Found %d Object Storage clusters:\n\n", len(summaries)))

	for _, cluster := range summaries {
		fmt.Fprintf(&stringBuilder, "ID: %s | Region: %s\n", cluster.ID, cluster.Region)
		fmt.Fprintf(&stringBuilder, "  Domain: %s\n", cluster.Domain)
		fmt.Fprintf(&stringBuilder, "  Status: %s\n", cluster.Status)

		if cluster.StaticSiteDomain != "" {
			fmt.Fprintf(&stringBuilder, "  Static Site Domain: %s\n", cluster.StaticSiteDomain)
		}

		stringBuilder.WriteString("\n")
	}

	return mcp.NewToolResultText(stringBuilder.String()), nil
}
