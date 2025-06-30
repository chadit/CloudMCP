package linode

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/linode/linodego"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	haStatusStandard         = "Standard"
	haStatusHighAvailability = "High Availability"
	autoscalerStatusDisabled = "Disabled"
)

// handleLKEClustersList lists all LKE clusters.
func (s *Service) handleLKEClustersList(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	clusters, err := account.Client.ListLKEClusters(ctx, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list LKE clusters: %v", err)), nil
	}

	summaries := make([]LKEClusterSummary, 0, len(clusters))
	for _, cluster := range clusters {
		summary := LKEClusterSummary{
			ID:         cluster.ID,
			Label:      cluster.Label,
			Region:     cluster.Region,
			Status:     string(cluster.Status),
			K8sVersion: cluster.K8sVersion,
			Tags:       cluster.Tags,
			Created:    cluster.Created.Format("2006-01-02T15:04:05"),
			Updated:    cluster.Updated.Format("2006-01-02T15:04:05"),
			ControlPlane: LKEControlPlane{
				HighAvailability: cluster.ControlPlane.HighAvailability,
			},
		}
		summaries = append(summaries, summary)
	}

	var stringBuilder strings.Builder

	stringBuilder.WriteString(fmt.Sprintf("Found %d LKE clusters:\n\n", len(summaries)))

	for _, cluster := range summaries {
		haStatus := haStatusStandard
		if cluster.ControlPlane.HighAvailability {
			haStatus = haStatusHighAvailability
		}

		fmt.Fprintf(&stringBuilder, "ID: %d | %s (%s)\n", cluster.ID, cluster.Label, cluster.Region)
		fmt.Fprintf(&stringBuilder, "  Status: %s\n", cluster.Status)
		fmt.Fprintf(&stringBuilder, "  Kubernetes Version: %s\n", cluster.K8sVersion)
		fmt.Fprintf(&stringBuilder, "  Control Plane: %s\n", haStatus)
		if len(cluster.Tags) > 0 {
			fmt.Fprintf(&stringBuilder, "  Tags: %s\n", strings.Join(cluster.Tags, ", "))
		}

		fmt.Fprintf(&stringBuilder, "  Updated: %s\n", cluster.Updated)
		stringBuilder.WriteString("\n")
	}

	return mcp.NewToolResultText(stringBuilder.String()), nil
}

// handleLKEClusterGet gets details of a specific LKE cluster.
func (s *Service) handleLKEClusterGet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params LKEClusterGetParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	cluster, err := account.Client.GetLKECluster(ctx, params.ClusterID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get LKE cluster: %v", err)), nil
	}

	// Get node pools
	pools, err := account.Client.ListLKENodePools(ctx, params.ClusterID, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get cluster node pools: %v", err)), nil
	}

	nodePools := make([]LKENodePool, 0, len(pools))
	for _, pool := range pools {
		disks := make([]LKENodePoolDisk, 0, len(pool.Disks))
		for _, disk := range pool.Disks {
			disks = append(disks, LKENodePoolDisk{
				Size: disk.Size,
				Type: disk.Type,
			})
		}

		nodes := make([]LKENode, 0, len(pool.Linodes))
		for _, node := range pool.Linodes {
			nodes = append(nodes, LKENode{
				ID:         node.ID,
				InstanceID: node.InstanceID,
				Status:     string(node.Status),
			})
		}

		nodePool := LKENodePool{
			ID:    pool.ID,
			Count: pool.Count,
			Type:  pool.Type,
			Disks: disks,
			Nodes: nodes,
			Autoscaler: LKENodePoolAutoscaler{
				Enabled: pool.Autoscaler.Enabled,
				Min:     pool.Autoscaler.Min,
				Max:     pool.Autoscaler.Max,
			},
			Tags: pool.Tags,
		}
		nodePools = append(nodePools, nodePool)
	}

	detail := LKEClusterDetail{
		ID:         cluster.ID,
		Label:      cluster.Label,
		Region:     cluster.Region,
		Status:     string(cluster.Status),
		K8sVersion: cluster.K8sVersion,
		Tags:       cluster.Tags,
		Created:    cluster.Created.Format("2006-01-02T15:04:05"),
		Updated:    cluster.Updated.Format("2006-01-02T15:04:05"),
		ControlPlane: LKEControlPlane{
			HighAvailability: cluster.ControlPlane.HighAvailability,
		},
		NodePools: nodePools,
	}

	var stringBuilder strings.Builder
	fmt.Fprintf(&stringBuilder, "LKE Cluster Details:\n")
	fmt.Fprintf(&stringBuilder, "ID: %d\n", detail.ID)
	fmt.Fprintf(&stringBuilder, "Label: %s\n", detail.Label)
	fmt.Fprintf(&stringBuilder, "Region: %s\n", detail.Region)
	fmt.Fprintf(&stringBuilder, "Status: %s\n", detail.Status)
	fmt.Fprintf(&stringBuilder, "Kubernetes Version: %s\n", detail.K8sVersion)

	haStatus := haStatusStandard
	if detail.ControlPlane.HighAvailability {
		haStatus = haStatusHighAvailability
	}

	fmt.Fprintf(&stringBuilder, "Control Plane: %s\n", haStatus)
	fmt.Fprintf(&stringBuilder, "Created: %s\n", detail.Created)
	fmt.Fprintf(&stringBuilder, "Updated: %s\n\n", detail.Updated)

	if len(detail.Tags) > 0 {
		fmt.Fprintf(&stringBuilder, "Tags: %s\n\n", strings.Join(detail.Tags, ", "))
	}

	if len(detail.NodePools) > 0 {
		fmt.Fprintf(&stringBuilder, "Node Pools:\n")
		for i, pool := range detail.NodePools {
			fmt.Fprintf(&stringBuilder, "  %d. Pool ID: %d\n", i+1, pool.ID)
			fmt.Fprintf(&stringBuilder, "     Type: %s\n", pool.Type)
			fmt.Fprintf(&stringBuilder, "     Count: %d nodes\n", pool.Count)

			if pool.Autoscaler.Enabled {
				fmt.Fprintf(&stringBuilder, "     Autoscaler: Enabled (Min: %d, Max: %d)\n", pool.Autoscaler.Min, pool.Autoscaler.Max)
			} else {
				fmt.Fprintf(&stringBuilder, "     Autoscaler: %s\n", autoscalerStatusDisabled)
			}

			if len(pool.Disks) > 0 {
				fmt.Fprintf(&stringBuilder, "     Disks:\n")
				for _, disk := range pool.Disks {
					fmt.Fprintf(&stringBuilder, "       - %s: %d GB\n", disk.Type, disk.Size)
				}
			}

			if len(pool.Nodes) > 0 {
				fmt.Fprintf(&stringBuilder, "     Nodes:\n")
				for _, node := range pool.Nodes {
					fmt.Fprintf(&stringBuilder, "       - %s (Instance: %d, Status: %s)\n", node.ID, node.InstanceID, node.Status)
				}
			}

			if len(pool.Tags) > 0 {
				fmt.Fprintf(&stringBuilder, "     Tags: %s\n", strings.Join(pool.Tags, ", "))
			}

			stringBuilder.WriteString("\n")
		}
	} else {
		stringBuilder.WriteString("No node pools found.\n")
	}

	return mcp.NewToolResultText(stringBuilder.String()), nil
}

// handleLKEClusterCreate creates a new LKE cluster.
func (s *Service) handleLKEClusterCreate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params LKEClusterCreateParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	nodePools := make([]linodego.LKENodePoolCreateOptions, 0, len(params.NodePools))
	for _, pool := range params.NodePools {
		disks := make([]linodego.LKENodePoolDisk, 0, len(pool.Disks))
		for _, disk := range pool.Disks {
			disks = append(disks, linodego.LKENodePoolDisk{
				Size: disk.Size,
			})
		}

		nodePool := linodego.LKENodePoolCreateOptions{
			Type:  pool.Type,
			Count: pool.Count,
			Disks: disks,
			Tags:  pool.Tags,
		}

		if pool.Autoscaler != nil {
			nodePool.Autoscaler = &linodego.LKENodePoolAutoscaler{
				Enabled: pool.Autoscaler.Enabled,
				Min:     pool.Autoscaler.Min,
				Max:     pool.Autoscaler.Max,
			}
		}

		nodePools = append(nodePools, nodePool)
	}

	createOpts := linodego.LKEClusterCreateOptions{
		Label:      params.Label,
		Region:     params.Region,
		K8sVersion: params.K8sVersion,
		Tags:       params.Tags,
		NodePools:  nodePools,
		ControlPlane: &linodego.LKEClusterControlPlaneOptions{
			HighAvailability: &params.ControlPlane.HighAvailability,
		},
	}

	cluster, err := account.Client.CreateLKECluster(ctx, createOpts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create LKE cluster: %v", err)), nil
	}

	haStatus := haStatusStandard
	if cluster.ControlPlane.HighAvailability {
		haStatus = haStatusHighAvailability
	}

	return mcp.NewToolResultText(fmt.Sprintf("LKE cluster created successfully:\nID: %d\nLabel: %s\nRegion: %s\nKubernetes Version: %s\nControl Plane: %s\nStatus: %s",
		cluster.ID, cluster.Label, cluster.Region, cluster.K8sVersion, haStatus, cluster.Status)), nil
}

// handleLKEClusterUpdate updates an existing LKE cluster.
func (s *Service) handleLKEClusterUpdate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params LKEClusterUpdateParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	updateOpts := linodego.LKEClusterUpdateOptions{}

	if params.Label != "" {
		updateOpts.Label = params.Label
	}
	if params.K8sVersion != "" {
		updateOpts.K8sVersion = params.K8sVersion
	}
	if params.Tags != nil {
		updateOpts.Tags = &params.Tags
	}
	if params.ControlPlane.HighAvailability {
		updateOpts.ControlPlane = &linodego.LKEClusterControlPlaneOptions{
			HighAvailability: &params.ControlPlane.HighAvailability,
		}
	}

	cluster, err := account.Client.UpdateLKECluster(ctx, params.ClusterID, updateOpts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update LKE cluster: %v", err)), nil
	}

	haStatus := haStatusStandard
	if cluster.ControlPlane.HighAvailability {
		haStatus = haStatusHighAvailability
	}

	return mcp.NewToolResultText(fmt.Sprintf("LKE cluster updated successfully:\nID: %d\nLabel: %s\nKubernetes Version: %s\nControl Plane: %s\nStatus: %s",
		cluster.ID, cluster.Label, cluster.K8sVersion, haStatus, cluster.Status)), nil
}

// handleLKEClusterDelete deletes an LKE cluster.
func (s *Service) handleLKEClusterDelete(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params LKEClusterDeleteParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	err = account.Client.DeleteLKECluster(ctx, params.ClusterID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete LKE cluster: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("LKE cluster %d deleted successfully", params.ClusterID)), nil
}

// handleLKENodePoolCreate creates a new node pool in an LKE cluster.
func (s *Service) handleLKENodePoolCreate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params LKENodePoolCreateParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	disks := make([]linodego.LKENodePoolDisk, 0, len(params.Disks))
	for _, disk := range params.Disks {
		disks = append(disks, linodego.LKENodePoolDisk{
			Size: disk.Size,
		})
	}

	createOpts := linodego.LKENodePoolCreateOptions{
		Type:  params.Type,
		Count: params.Count,
		Disks: disks,
		Tags:  params.Tags,
	}

	if params.Autoscaler != nil {
		createOpts.Autoscaler = &linodego.LKENodePoolAutoscaler{
			Enabled: params.Autoscaler.Enabled,
			Min:     params.Autoscaler.Min,
			Max:     params.Autoscaler.Max,
		}
	}

	pool, err := account.Client.CreateLKENodePool(ctx, params.ClusterID, createOpts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create node pool: %v", err)), nil
	}

	autoscalerStatus := autoscalerStatusDisabled
	if pool.Autoscaler.Enabled {
		autoscalerStatus = fmt.Sprintf("Enabled (Min: %d, Max: %d)", pool.Autoscaler.Min, pool.Autoscaler.Max)
	}

	return mcp.NewToolResultText(fmt.Sprintf("Node pool created successfully:\nPool ID: %d\nType: %s\nCount: %d\nAutoscaler: %s",
		pool.ID, pool.Type, pool.Count, autoscalerStatus)), nil
}

// handleLKENodePoolUpdate updates a node pool in an LKE cluster.
func (s *Service) handleLKENodePoolUpdate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params LKENodePoolUpdateParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	updateOpts := linodego.LKENodePoolUpdateOptions{}

	if params.Count > 0 {
		updateOpts.Count = params.Count
	}
	if params.Tags != nil {
		updateOpts.Tags = &params.Tags
	}
	if params.Autoscaler != nil {
		updateOpts.Autoscaler = &linodego.LKENodePoolAutoscaler{
			Enabled: params.Autoscaler.Enabled,
			Min:     params.Autoscaler.Min,
			Max:     params.Autoscaler.Max,
		}
	}

	pool, err := account.Client.UpdateLKENodePool(ctx, params.ClusterID, params.PoolID, updateOpts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update node pool: %v", err)), nil
	}

	autoscalerStatus := autoscalerStatusDisabled
	if pool.Autoscaler.Enabled {
		autoscalerStatus = fmt.Sprintf("Enabled (Min: %d, Max: %d)", pool.Autoscaler.Min, pool.Autoscaler.Max)
	}

	return mcp.NewToolResultText(fmt.Sprintf("Node pool updated successfully:\nPool ID: %d\nType: %s\nCount: %d\nAutoscaler: %s",
		pool.ID, pool.Type, pool.Count, autoscalerStatus)), nil
}

// handleLKENodePoolDelete deletes a node pool from an LKE cluster.
func (s *Service) handleLKENodePoolDelete(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params LKENodePoolDeleteParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	err = account.Client.DeleteLKENodePool(ctx, params.ClusterID, params.PoolID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete node pool: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Node pool %d deleted successfully from cluster %d",
		params.PoolID, params.ClusterID)), nil
}

// handleLKEKubeconfig retrieves the kubeconfig for an LKE cluster.
func (s *Service) handleLKEKubeconfig(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params LKEKubeconfigParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	kubeconfig, err := account.Client.GetLKEClusterKubeconfig(ctx, params.ClusterID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get kubeconfig: %v", err)), nil
	}

	// Decode the base64 kubeconfig
	kubeconfigBytes, err := base64.StdEncoding.DecodeString(kubeconfig.KubeConfig)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to decode kubeconfig: %v", err)), nil
	}

	var stringBuilder strings.Builder
	fmt.Fprintf(&stringBuilder, "Kubeconfig for LKE cluster %d:\n\n", params.ClusterID)
	fmt.Fprintf(&stringBuilder, "```yaml\n%s\n```\n\n", string(kubeconfigBytes))
	fmt.Fprintf(&stringBuilder, "To use this kubeconfig:\n")
	fmt.Fprintf(&stringBuilder, "1. Save the content to a file (e.g., ~/.kube/config)\n")
	fmt.Fprintf(&stringBuilder, "2. Set KUBECONFIG environment variable: export KUBECONFIG=~/.kube/config\n")
	fmt.Fprintf(&stringBuilder, "3. Test connection: kubectl get nodes\n")

	return mcp.NewToolResultText(stringBuilder.String()), nil
}
