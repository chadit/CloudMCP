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

	var summaries []ObjectStorageBucketSummary
	for _, bucket := range buckets {
		summary := ObjectStorageBucketSummary{
			Label:    bucket.Label,
			Cluster:  bucket.Cluster,
			Hostname: bucket.Hostname,
			Created:  bucket.Created.Format("2006-01-02T15:04:05"),
			Size:     int64(bucket.Size),
			Objects:  bucket.Objects,
		}
		summaries = append(summaries, summary)
	}

	// Remove unused result variable

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d Object Storage buckets:\n\n", len(summaries)))

	for _, bucket := range summaries {
		sizeDisplay := "0 bytes"

		if bucket.Size > 0 {
			if bucket.Size >= 1024*1024*1024 {
				sizeDisplay = fmt.Sprintf("%.2f GB", float64(bucket.Size)/(1024*1024*1024))
			} else if bucket.Size >= 1024*1024 {
				sizeDisplay = fmt.Sprintf("%.2f MB", float64(bucket.Size)/(1024*1024))
			} else if bucket.Size >= 1024 {
				sizeDisplay = fmt.Sprintf("%.2f KB", float64(bucket.Size)/1024)
			} else {
				sizeDisplay = fmt.Sprintf("%d bytes", bucket.Size)
			}
		}

		fmt.Fprintf(&sb, "Name: %s (%s)\n", bucket.Label, bucket.Cluster)
		fmt.Fprintf(&sb, "  Hostname: %s\n", bucket.Hostname)
		fmt.Fprintf(&sb, "  Size: %s | Objects: %d\n", sizeDisplay, bucket.Objects)
		fmt.Fprintf(&sb, "  Created: %s\n", bucket.Created)
		sb.WriteString("\n")
	}

	return mcp.NewToolResultText(sb.String()), nil
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

	bucket, err := account.Client.GetObjectStorageBucket(ctx, params.Cluster, params.Bucket)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Object Storage bucket: %v", err)), nil
	}

	detail := ObjectStorageBucketDetail{
		Label:    bucket.Label,
		Cluster:  bucket.Cluster,
		Hostname: bucket.Hostname,
		Created:  bucket.Created.Format("2006-01-02T15:04:05"),
		Size:     int64(bucket.Size),
		Objects:  bucket.Objects,
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Object Storage Bucket Details:\n")
	fmt.Fprintf(&sb, "Name: %s\n", detail.Label)
	fmt.Fprintf(&sb, "Cluster: %s\n", detail.Cluster)
	fmt.Fprintf(&sb, "Hostname: %s\n", detail.Hostname)
	fmt.Fprintf(&sb, "Created: %s\n", detail.Created)

	sizeDisplay := "0 bytes"

	if detail.Size > 0 {
		if detail.Size >= 1024*1024*1024 {
			sizeDisplay = fmt.Sprintf("%.2f GB", float64(detail.Size)/(1024*1024*1024))
		} else if detail.Size >= 1024*1024 {
			sizeDisplay = fmt.Sprintf("%.2f MB", float64(detail.Size)/(1024*1024))
		} else if detail.Size >= 1024 {
			sizeDisplay = fmt.Sprintf("%.2f KB", float64(detail.Size)/1024)
		} else {
			sizeDisplay = fmt.Sprintf("%d bytes", detail.Size)
		}
	}

	fmt.Fprintf(&sb, "Size: %s\n", sizeDisplay)
	fmt.Fprintf(&sb, "Objects: %d\n", detail.Objects)

	return mcp.NewToolResultText(sb.String()), nil
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

	// Support both deprecated Cluster and new Region parameters
	if params.Region != "" {
		createOpts.Region = params.Region
	} else if params.Cluster != "" {
		createOpts.Cluster = params.Cluster
	}

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

	return mcp.NewToolResultText(fmt.Sprintf("Object Storage bucket created successfully:\nName: %s\nCluster: %s\nHostname: %s",
		bucket.Label, bucket.Cluster, bucket.Hostname)), nil
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

	err = account.Client.UpdateObjectStorageBucketAccess(ctx, params.Cluster, params.Bucket, updateOpts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update Object Storage bucket access: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Object Storage bucket '%s' access updated successfully in cluster '%s'",
		params.Bucket, params.Cluster)), nil
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

	err = account.Client.DeleteObjectStorageBucket(ctx, params.Cluster, params.Bucket)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete Object Storage bucket: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Object Storage bucket '%s' deleted successfully from cluster '%s'",
		params.Bucket, params.Cluster)), nil
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

	var summaries []ObjectStorageKeySummary
	for _, key := range keys {
		var bucketAccess []ObjectStorageBucketAccess
		if key.BucketAccess != nil {
			for _, access := range *key.BucketAccess {
				bucketAccess = append(bucketAccess, ObjectStorageBucketAccess{
					Cluster:     access.Cluster,
					BucketName:  access.BucketName,
					Permissions: string(access.Permissions),
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

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d Object Storage keys:\n\n", len(summaries)))

	for _, key := range summaries {
		accessType := bucketAccessFull
		if key.Limited {
			accessType = bucketAccessLimited
		}

		fmt.Fprintf(&sb, "ID: %d | %s (%s)\n", key.ID, key.Label, accessType)
		fmt.Fprintf(&sb, "  Access Key: %s\n", key.AccessKey)

		if key.Limited && len(key.BucketAccess) > 0 {
			fmt.Fprintf(&sb, "  Bucket Access:\n")
			for _, access := range key.BucketAccess {
				fmt.Fprintf(&sb, "    - %s/%s: %s\n", access.Cluster, access.BucketName, access.Permissions)
			}
		}
		sb.WriteString("\n")
	}

	return mcp.NewToolResultText(sb.String()), nil
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
		for _, access := range *key.BucketAccess {
			bucketAccess = append(bucketAccess, ObjectStorageBucketAccess{
				Cluster:     access.Cluster,
				BucketName:  access.BucketName,
				Permissions: string(access.Permissions),
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

	var sb strings.Builder
	fmt.Fprintf(&sb, "Object Storage Key Details:\n")
	fmt.Fprintf(&sb, "ID: %d\n", detail.ID)
	fmt.Fprintf(&sb, "Label: %s\n", detail.Label)
	fmt.Fprintf(&sb, "Access Key: %s\n", detail.AccessKey)
	fmt.Fprintf(&sb, "Secret Key: %s\n", detail.SecretKey)

	accessType := bucketAccessFull
	if detail.Limited {
		accessType = "Limited Access"
	}
	fmt.Fprintf(&sb, "Access Type: %s\n", accessType)

	if detail.Limited && len(detail.BucketAccess) > 0 {
		fmt.Fprintf(&sb, "\nBucket Access Permissions:\n")
		for _, access := range detail.BucketAccess {
			fmt.Fprintf(&sb, "  - Cluster: %s\n", access.Cluster)
			fmt.Fprintf(&sb, "    Bucket: %s\n", access.BucketName)
			fmt.Fprintf(&sb, "    Permissions: %s\n", access.Permissions)
			sb.WriteString("\n")
		}
	}

	return mcp.NewToolResultText(sb.String()), nil
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
		var bucketAccess []linodego.ObjectStorageKeyBucketAccess
		for _, access := range params.BucketAccess {
			bucketAccess = append(bucketAccess, linodego.ObjectStorageKeyBucketAccess{
				Cluster:     access.Cluster,
				BucketName:  access.BucketName,
				Permissions: string(access.Permissions),
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
		accessType = "Limited Access"
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
		// TODO: BucketAccess field not available in ObjectStorageKeyUpdateOptions
		// Once the Linode API supports updating bucket access, implement this:
		// var bucketAccess []linodego.ObjectStorageKeyBucketAccess
		// for _, access := range params.BucketAccess {
		//     bucketAccess = append(bucketAccess, linodego.ObjectStorageKeyBucketAccess{
		//         Cluster:     access.Cluster,
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
		accessType = "Limited Access"
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

	var summaries []ObjectStorageClusterSummary
	for _, cluster := range clusters {
		summary := ObjectStorageClusterSummary{
			ID:               cluster.ID,
			Domain:           cluster.Domain,
			Status:           string(cluster.Status),
			Region:           cluster.Region,
			StaticSiteDomain: cluster.StaticSiteDomain,
		}
		summaries = append(summaries, summary)
	}

	// Remove unused result variable

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d Object Storage clusters:\n\n", len(summaries)))

	for _, cluster := range summaries {
		fmt.Fprintf(&sb, "ID: %s | Region: %s\n", cluster.ID, cluster.Region)
		fmt.Fprintf(&sb, "  Domain: %s\n", cluster.Domain)
		fmt.Fprintf(&sb, "  Status: %s\n", cluster.Status)
		if cluster.StaticSiteDomain != "" {
			fmt.Fprintf(&sb, "  Static Site Domain: %s\n", cluster.StaticSiteDomain)
		}
		sb.WriteString("\n")
	}

	return mcp.NewToolResultText(sb.String()), nil
}
