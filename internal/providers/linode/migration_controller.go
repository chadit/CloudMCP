// Package linode provides migration control infrastructure for Batch 1 tool migration.
// This file implements percentage-based traffic routing, feature flags, and rollback
// capabilities for the CloudMCP Factory Pattern Migration.
package linode

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/chadit/CloudMCP/pkg/logger"
)

// Constants for magic numbers.
const (
	PercentageMax          = 100
	PercentageMin          = 0
	DefaultLatencyCapacity = 1000
	MaxLatencyHistory      = 10000
	LatencyTrimSize        = 1000
	Batch1ToolsCount       = 9
	Batch2ToolsCount       = 14
	Batch3ToolsCount       = 12
	Batch4ToolsCount       = 21
	PercentageRandomBound  = 100
)

// Static errors for err113 compliance.
var (
	ErrInvalidPercentageRange = errors.New("percentage must be between 0 and 100")
	ErrToolNotFoundInSettings = errors.New("tool not found in migration settings")
	ErrPercentageExceedsMax   = errors.New("percentage exceeds global maximum")
)

// MigrationController manages migration settings and traffic routing for Batch 1 tools.
type MigrationController struct {
	mu                    sync.RWMutex
	logger                logger.Logger
	toolMigrationSettings map[string]*ToolMigrationSettings
	globalMigrationConfig *GlobalMigrationConfig
	metrics               *MigrationMetrics
}

// ToolMigrationSettings defines migration configuration for a specific tool.
type ToolMigrationSettings struct {
	ToolName            string    `json:"toolName"`
	MigrationEnabled    bool      `json:"migrationEnabled"`
	TrafficPercentage   int       `json:"trafficPercentage"` // 0-100, percentage to route to provider-native
	ForceProviderNative bool      `json:"forceProviderNative"`
	ForceServiceBacked  bool      `json:"forceServiceBacked"`
	LastUpdated         time.Time `json:"lastUpdated"`
	UpdatedBy           string    `json:"updatedBy"`
}

// GlobalMigrationConfig defines global migration settings.
type GlobalMigrationConfig struct {
	MigrationEnabled  bool      `json:"migrationEnabled"`
	DefaultPercentage int       `json:"defaultPercentage"`
	MaxPercentage     int       `json:"maxPercentage"`
	RollbackMode      bool      `json:"rollbackMode"`
	MaintenanceMode   bool      `json:"maintenanceMode"`
	LastUpdated       time.Time `json:"lastUpdated"`
	UpdatedBy         string    `json:"updatedBy"`
}

// MigrationMetrics tracks migration performance and reliability.
type MigrationMetrics struct {
	mu                       sync.RWMutex
	ProviderNativeExecutions map[string]int64   `json:"providerNativeExecutions"`
	ServiceBackedExecutions  map[string]int64   `json:"serviceBackedExecutions"`
	ProviderNativeErrors     map[string]int64   `json:"providerNativeErrors"`
	ServiceBackedErrors      map[string]int64   `json:"serviceBackedErrors"`
	ProviderNativeLatency    map[string][]int64 `json:"providerNativeLatency"`
	ServiceBackedLatency     map[string][]int64 `json:"serviceBackedLatency"`
	LastReset                time.Time          `json:"lastReset"`
}

// NewMigrationController creates a new migration controller for Batch 1 tools.
func NewMigrationController(logger logger.Logger) *MigrationController {
	migrationController := &MigrationController{
		logger:                logger,
		toolMigrationSettings: make(map[string]*ToolMigrationSettings),
		globalMigrationConfig: &GlobalMigrationConfig{
			MigrationEnabled:  true,
			DefaultPercentage: PercentageMin, // Start with 0% migration
			MaxPercentage:     PercentageMax, // Allow up to 100%
			RollbackMode:      false,
			MaintenanceMode:   false,
			LastUpdated:       time.Now(),
			UpdatedBy:         "system",
		},
		metrics: &MigrationMetrics{
			ProviderNativeExecutions: make(map[string]int64),
			ServiceBackedExecutions:  make(map[string]int64),
			ProviderNativeErrors:     make(map[string]int64),
			ServiceBackedErrors:      make(map[string]int64),
			ProviderNativeLatency:    make(map[string][]int64),
			ServiceBackedLatency:     make(map[string][]int64),
			LastReset:                time.Now(),
		},
	}

	// Initialize all batches with default settings
	migrationController.initializeAllBatchTools()

	return migrationController
}

// initializeAllBatchTools sets up migration settings for all batch tools.
func (mc *MigrationController) initializeAllBatchTools() {
	mc.initializeBatch1Tools()
	mc.initializeBatch2Tools()
	mc.initializeBatch3Tools()
	mc.initializeBatch4Tools()
}

// initializeBatch1Tools sets up migration settings for all Batch 1 tools.
func (mc *MigrationController) initializeBatch1Tools() {
	batch1Tools := []string{
		"linode_account_info",
		"linode_account_availability",
		"linode_account_invoices_list",
		"linode_account_invoice_get",
		"linode_account_payments_list",
		"linode_account_transfer_get",
		"cloudmcp_account_switch",
		"cloudmcp_version",
		"cloudmcp_version_extended",
	}

	for _, toolName := range batch1Tools {
		mc.toolMigrationSettings[toolName] = &ToolMigrationSettings{
			ToolName:            toolName,
			MigrationEnabled:    true,
			TrafficPercentage:   PercentageMin, // Start with 0% traffic to provider-native
			ForceProviderNative: false,
			ForceServiceBacked:  false,
			LastUpdated:         time.Now(),
			UpdatedBy:           "system",
		}
	}

	mc.logger.Info("Initialized Batch 1 tool migration settings",
		"tools", len(batch1Tools))
}

// initializeBatch2Tools sets up migration settings for all Batch 2 tools.
func (mc *MigrationController) initializeBatch2Tools() {
	batch2Tools := []string{
		"linode_instances_list",
		"linode_instance_get",
		"linode_instance_create",
		"linode_instance_delete",
		"linode_instance_boot",
		"linode_instance_shutdown",
		"linode_instance_reboot",
		"linode_images_list",
		"linode_image_get",
		"linode_image_create",
		"linode_image_update",
		"linode_image_delete",
		"linode_image_replicate",
		"linode_image_upload_create",
	}

	for _, toolName := range batch2Tools {
		// Use conservative percentages for high-risk compute operations
		initialPercentage := PercentageMin
		if toolName == "linode_instance_create" || toolName == "linode_instance_delete" || 
		   toolName == "linode_image_delete" {
			// High-risk operations start with even more conservative settings
			initialPercentage = PercentageMin
		}

		mc.toolMigrationSettings[toolName] = &ToolMigrationSettings{
			ToolName:            toolName,
			MigrationEnabled:    true,
			TrafficPercentage:   initialPercentage,
			ForceProviderNative: false,
			ForceServiceBacked:  false,
			LastUpdated:         time.Now(),
			UpdatedBy:           "system",
		}
	}

	mc.logger.Info("Initialized Batch 2 tool migration settings",
		"tools", len(batch2Tools))
}

// initializeBatch3Tools sets up migration settings for all Batch 3 storage tools.
func (mc *MigrationController) initializeBatch3Tools() {
	batch3Tools := []string{
		"linode_volumes_list",
		"linode_volume_get",
		"linode_volume_create",
		"linode_volume_update",
		"linode_volume_delete",
		"linode_volume_attach",
		"linode_objectstorage_buckets_list",
		"linode_objectstorage_bucket_create",
		"linode_objectstorage_bucket_delete",
		"linode_objectstorage_objects_list",
		"linode_objectstorage_object_create",
		"linode_objectstorage_object_delete",
	}

	for _, toolName := range batch3Tools {
		// Use conservative percentages for high-risk storage operations
		initialPercentage := PercentageMin
		if toolName == "linode_volume_create" || toolName == "linode_volume_delete" || 
		   toolName == "linode_objectstorage_bucket_create" || toolName == "linode_objectstorage_bucket_delete" ||
		   toolName == "linode_objectstorage_object_delete" {
			// High-risk operations start with very conservative settings for data safety
			initialPercentage = PercentageMin
		}

		mc.toolMigrationSettings[toolName] = &ToolMigrationSettings{
			ToolName:            toolName,
			MigrationEnabled:    true,
			TrafficPercentage:   initialPercentage,
			ForceProviderNative: false,
			ForceServiceBacked:  false,
			LastUpdated:         time.Now(),
			UpdatedBy:           "system",
		}
	}

	mc.logger.Info("Initialized Batch 3 storage tool migration settings with enhanced safety controls",
		"tools", len(batch3Tools))
}

// initializeBatch4Tools sets up migration settings for all Batch 4 simple operations tools.
func (mc *MigrationController) initializeBatch4Tools() {
	batch4Tools := []string{
		// DNS Management (10 tools)
		"linode_domains_list",
		"linode_domain_get",
		"linode_domain_create",
		"linode_domain_update",
		"linode_domain_delete",
		"linode_domain_records_list",
		"linode_domain_record_get",
		"linode_domain_record_create",
		"linode_domain_record_update",
		"linode_domain_record_delete",

		// Monitoring (5 tools)
		"linode_longview_clients_list",
		"linode_longview_client_get",
		"linode_longview_client_create",
		"linode_longview_client_update",
		"linode_longview_client_delete",

		// Automation (3 tools)
		"linode_stackscripts_list",
		"linode_stackscript_get",
		"linode_stackscript_create",

		// Support (3 tools)
		"linode_support_tickets_list",
		"linode_support_ticket_get",
		"linode_support_ticket_create",
	}

	for _, toolName := range batch4Tools {
		// Use appropriate percentages based on risk level for simple operations
		initialPercentage := PercentageMin
		if toolName == "linode_domain_delete" {
			// High-risk operations (data loss) start with very conservative settings
			initialPercentage = PercentageMin
		} else if toolName == "linode_domain_create" || toolName == "linode_domain_update" ||
			toolName == "linode_domain_record_create" || toolName == "linode_domain_record_update" ||
			toolName == "linode_domain_record_delete" || toolName == "linode_longview_client_create" ||
			toolName == "linode_longview_client_update" || toolName == "linode_longview_client_delete" ||
			toolName == "linode_stackscript_create" || toolName == "linode_support_ticket_create" {
			// Medium-risk operations (billing/configuration impact) start conservatively
			initialPercentage = PercentageMin
		} else {
			// Low-risk operations (read-only) can start with higher percentage
			initialPercentage = PercentageMin
		}

		mc.toolMigrationSettings[toolName] = &ToolMigrationSettings{
			ToolName:            toolName,
			MigrationEnabled:    true,
			TrafficPercentage:   initialPercentage,
			ForceProviderNative: false,
			ForceServiceBacked:  false,
			LastUpdated:         time.Now(),
			UpdatedBy:           "system",
		}
	}

	mc.logger.Info("Initialized Batch 4 simple operations tool migration settings with DNS-specific safety controls",
		"tools", len(batch4Tools))
}

// ShouldUseProviderNative determines whether to route a request to provider-native implementation.
func (mc *MigrationController) ShouldUseProviderNative(toolName string) bool {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	// Check global migration state
	if !mc.globalMigrationConfig.MigrationEnabled {
		return false
	}

	if mc.globalMigrationConfig.RollbackMode {
		return false
	}

	if mc.globalMigrationConfig.MaintenanceMode {
		return false
	}

	// Get tool-specific settings
	settings, exists := mc.toolMigrationSettings[toolName]
	if !exists {
		return false
	}

	if !settings.MigrationEnabled {
		return false
	}

	// Check force flags
	if settings.ForceServiceBacked {
		return false
	}

	if settings.ForceProviderNative {
		return true
	}

	// Apply percentage-based routing
	return mc.shouldRouteByPercentage(settings.TrafficPercentage)
}

// shouldRouteByPercentage implements percentage-based traffic routing.
func (mc *MigrationController) shouldRouteByPercentage(percentage int) bool {
	if percentage <= PercentageMin {
		return false
	}

	if percentage >= PercentageMax {
		return true
	}

	// Generate cryptographically secure random number for fair distribution
	randomNumber, err := rand.Int(rand.Reader, big.NewInt(PercentageRandomBound))
	if err != nil {
		// Fall back to service-backed on error
		mc.logger.Error("Failed to generate random number for routing", "error", err)

		return false
	}

	return randomNumber.Int64() < int64(percentage)
}

// SetToolMigrationPercentage sets the migration percentage for a specific tool.
func (mc *MigrationController) SetToolMigrationPercentage(toolName string, percentage int, updatedBy string) error {
	if percentage < PercentageMin || percentage > PercentageMax {
		return fmt.Errorf("%w, got %d", ErrInvalidPercentageRange, percentage)
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	settings, exists := mc.toolMigrationSettings[toolName]
	if !exists {
		return fmt.Errorf("%w: %s", ErrToolNotFoundInSettings, toolName)
	}

	// Check global max percentage limit
	if percentage > mc.globalMigrationConfig.MaxPercentage {
		return fmt.Errorf("%w: %d exceeds %d",
			ErrPercentageExceedsMax, percentage, mc.globalMigrationConfig.MaxPercentage)
	}

	oldPercentage := settings.TrafficPercentage
	settings.TrafficPercentage = percentage
	settings.LastUpdated = time.Now()
	settings.UpdatedBy = updatedBy

	mc.logger.Info("Updated tool migration percentage",
		"tool", toolName,
		"old_percentage", oldPercentage,
		"new_percentage", percentage,
		"updated_by", updatedBy,
	)

	return nil
}

// EnableToolMigration enables migration for a specific tool.
func (mc *MigrationController) EnableToolMigration(toolName string, updatedBy string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	settings, exists := mc.toolMigrationSettings[toolName]
	if !exists {
		return fmt.Errorf("%w: %s", ErrToolNotFoundInSettings, toolName)
	}

	settings.MigrationEnabled = true
	settings.LastUpdated = time.Now()
	settings.UpdatedBy = updatedBy

	mc.logger.Info("Enabled tool migration", "tool", toolName, "updated_by", updatedBy)

	return nil
}

// DisableToolMigration disables migration for a specific tool.
func (mc *MigrationController) DisableToolMigration(toolName string, updatedBy string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	settings, exists := mc.toolMigrationSettings[toolName]
	if !exists {
		return fmt.Errorf("%w: %s", ErrToolNotFoundInSettings, toolName)
	}

	settings.MigrationEnabled = false
	settings.LastUpdated = time.Now()
	settings.UpdatedBy = updatedBy

	mc.logger.Info("Disabled tool migration", "tool", toolName, "updated_by", updatedBy)

	return nil
}

// ForceProviderNative forces a tool to always use provider-native implementation.
func (mc *MigrationController) ForceProviderNative(toolName string, updatedBy string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	settings, exists := mc.toolMigrationSettings[toolName]
	if !exists {
		return fmt.Errorf("%w: %s", ErrToolNotFoundInSettings, toolName)
	}

	settings.ForceProviderNative = true
	settings.ForceServiceBacked = false
	settings.LastUpdated = time.Now()
	settings.UpdatedBy = updatedBy

	mc.logger.Info("Forced tool to provider-native", "tool", toolName, "updated_by", updatedBy)

	return nil
}

// ForceServiceBacked forces a tool to always use service-backed implementation.
func (mc *MigrationController) ForceServiceBacked(toolName string, updatedBy string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	settings, exists := mc.toolMigrationSettings[toolName]
	if !exists {
		return fmt.Errorf("%w: %s", ErrToolNotFoundInSettings, toolName)
	}

	settings.ForceServiceBacked = true
	settings.ForceProviderNative = false
	settings.LastUpdated = time.Now()
	settings.UpdatedBy = updatedBy

	mc.logger.Info("Forced tool to service-backed", "tool", toolName, "updated_by", updatedBy)

	return nil
}

// ClearForceFlags removes force flags for a tool, returning to percentage-based routing.
func (mc *MigrationController) ClearForceFlags(toolName string, updatedBy string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	settings, exists := mc.toolMigrationSettings[toolName]
	if !exists {
		return fmt.Errorf("%w: %s", ErrToolNotFoundInSettings, toolName)
	}

	settings.ForceProviderNative = false
	settings.ForceServiceBacked = false
	settings.LastUpdated = time.Now()
	settings.UpdatedBy = updatedBy

	mc.logger.Info("Cleared force flags for tool", "tool", toolName, "updated_by", updatedBy)

	return nil
}

// EnableGlobalRollback enables global rollback mode (all tools use service-backed).
func (mc *MigrationController) EnableGlobalRollback(updatedBy string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.globalMigrationConfig.RollbackMode = true
	mc.globalMigrationConfig.LastUpdated = time.Now()
	mc.globalMigrationConfig.UpdatedBy = updatedBy

	mc.logger.Warn("Enabled global rollback mode", "updated_by", updatedBy)
}

// DisableGlobalRollback disables global rollback mode.
func (mc *MigrationController) DisableGlobalRollback(updatedBy string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.globalMigrationConfig.RollbackMode = false
	mc.globalMigrationConfig.LastUpdated = time.Now()
	mc.globalMigrationConfig.UpdatedBy = updatedBy

	mc.logger.Info("Disabled global rollback mode", "updated_by", updatedBy)
}

// RecordExecution records metrics for tool execution.
func (mc *MigrationController) RecordExecution(toolName string, isProviderNative bool, success bool, latencyMs int64) {
	mc.metrics.mu.Lock()
	defer mc.metrics.mu.Unlock()

	mc.recordExecutionMetrics(toolName, isProviderNative, success, latencyMs)

	// Limit latency history to prevent memory growth
	if len(mc.metrics.ProviderNativeLatency[toolName]) > MaxLatencyHistory {
		mc.metrics.ProviderNativeLatency[toolName] = mc.metrics.ProviderNativeLatency[toolName][LatencyTrimSize:]
	}

	if len(mc.metrics.ServiceBackedLatency[toolName]) > MaxLatencyHistory {
		mc.metrics.ServiceBackedLatency[toolName] = mc.metrics.ServiceBackedLatency[toolName][LatencyTrimSize:]
	}
}

// GetMigrationStatus returns comprehensive migration status for all tools.
func (mc *MigrationController) GetMigrationStatus() map[string]interface{} {
	mc.mu.RLock()
	globalConfig := *mc.globalMigrationConfig
	toolSettings := make(map[string]*ToolMigrationSettings)

	for k, v := range mc.toolMigrationSettings {
		settingsCopy := *v
		toolSettings[k] = &settingsCopy
	}
	mc.mu.RUnlock()

	mc.metrics.mu.RLock()
	// Create metrics copy without mutex to avoid lock copying
	metricsSnapshot := MigrationMetrics{
		ProviderNativeExecutions: make(map[string]int64),
		ServiceBackedExecutions:  make(map[string]int64),
		ProviderNativeErrors:     make(map[string]int64),
		ServiceBackedErrors:      make(map[string]int64),
		ProviderNativeLatency:    make(map[string][]int64),
		ServiceBackedLatency:     make(map[string][]int64),
		LastReset:                mc.metrics.LastReset,
	}

	// Copy maps to avoid reference sharing
	for k, v := range mc.metrics.ProviderNativeExecutions {
		metricsSnapshot.ProviderNativeExecutions[k] = v
	}

	for k, v := range mc.metrics.ServiceBackedExecutions {
		metricsSnapshot.ServiceBackedExecutions[k] = v
	}

	for k, v := range mc.metrics.ProviderNativeErrors {
		metricsSnapshot.ProviderNativeErrors[k] = v
	}

	for k, v := range mc.metrics.ServiceBackedErrors {
		metricsSnapshot.ServiceBackedErrors[k] = v
	}

	for k, v := range mc.metrics.ProviderNativeLatency {
		metricsSnapshot.ProviderNativeLatency[k] = make([]int64, len(v))
		copy(metricsSnapshot.ProviderNativeLatency[k], v)
	}

	for k, v := range mc.metrics.ServiceBackedLatency {
		metricsSnapshot.ServiceBackedLatency[k] = make([]int64, len(v))
		copy(metricsSnapshot.ServiceBackedLatency[k], v)
	}
	mc.metrics.mu.RUnlock()

	// Calculate total tools
	totalTools := Batch1ToolsCount + Batch2ToolsCount + Batch3ToolsCount + Batch4ToolsCount

	// Create result without embedding mutex
	result := map[string]interface{}{
		"global_config": globalConfig,
		"tool_settings": toolSettings,
		"total_tools":   totalTools, // Top-level total_tools for backward compatibility
		"metrics": map[string]interface{}{
			"providerNativeExecutions": metricsSnapshot.ProviderNativeExecutions,
			"serviceBackedExecutions":  metricsSnapshot.ServiceBackedExecutions,
			"providerNativeErrors":     metricsSnapshot.ProviderNativeErrors,
			"serviceBackedErrors":      metricsSnapshot.ServiceBackedErrors,
			"providerNativeLatency":    metricsSnapshot.ProviderNativeLatency,
			"serviceBackedLatency":     metricsSnapshot.ServiceBackedLatency,
			"lastReset":                metricsSnapshot.LastReset,
		},
		"batch_info": map[string]interface{}{
			"batch_1": map[string]interface{}{
				"name":        "batch-1",
				"description": "Foundation tools: Account Management (7) + System Information (2)",
				"tools_count": Batch1ToolsCount,
				"status":      "active",
			},
			"batch_2": map[string]interface{}{
				"name":        "batch-2", 
				"description": "Compute tools: Instances (7) + Images (7)",
				"tools_count": Batch2ToolsCount,
				"status":      "active",
			},
			"batch_3": map[string]interface{}{
				"name":        "batch-3",
				"description": "Storage tools: Volumes (6) + Object Storage (6)",
				"tools_count": Batch3ToolsCount,
				"status":      "active",
			},
			"batch_4": map[string]interface{}{
				"name":        "batch-4",
				"description": "Simple Operations: DNS (10) + Monitoring (5) + Automation (3) + Support (3)",
				"tools_count": Batch4ToolsCount,
				"status":      "active",
			},
			"total_tools": totalTools,
		},
	}

	return result
}

// GetToolMigrationSettings returns migration settings for a specific tool.
func (mc *MigrationController) GetToolMigrationSettings(toolName string) (*ToolMigrationSettings, bool) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	settings, exists := mc.toolMigrationSettings[toolName]
	if !exists {
		return nil, false
	}

	// Return a copy to prevent external modification
	settingsCopy := *settings

	return &settingsCopy, true
}

// ResetMetrics resets all migration metrics.
func (mc *MigrationController) ResetMetrics(updatedBy string) {
	mc.metrics.mu.Lock()
	defer mc.metrics.mu.Unlock()

	mc.metrics.ProviderNativeExecutions = make(map[string]int64)
	mc.metrics.ServiceBackedExecutions = make(map[string]int64)
	mc.metrics.ProviderNativeErrors = make(map[string]int64)
	mc.metrics.ServiceBackedErrors = make(map[string]int64)
	mc.metrics.ProviderNativeLatency = make(map[string][]int64)
	mc.metrics.ServiceBackedLatency = make(map[string][]int64)
	mc.metrics.LastReset = time.Now()

	mc.logger.Info("Reset migration metrics", "updated_by", updatedBy)
}

// recordExecutionMetrics handles the actual metrics recording logic.
func (mc *MigrationController) recordExecutionMetrics(toolName string, isProviderNative bool, success bool, latencyMs int64) {
	if isProviderNative {
		mc.recordProviderNativeMetrics(toolName, success, latencyMs)
	} else {
		mc.recordServiceBackedMetrics(toolName, success, latencyMs)
	}
}

// recordProviderNativeMetrics records metrics for provider-native executions.
func (mc *MigrationController) recordProviderNativeMetrics(toolName string, success bool, latencyMs int64) {
	mc.metrics.ProviderNativeExecutions[toolName]++
	if !success {
		mc.metrics.ProviderNativeErrors[toolName]++
	}

	if mc.metrics.ProviderNativeLatency[toolName] == nil {
		mc.metrics.ProviderNativeLatency[toolName] = make([]int64, 0, DefaultLatencyCapacity)
	}

	mc.metrics.ProviderNativeLatency[toolName] = append(
		mc.metrics.ProviderNativeLatency[toolName], latencyMs)
}

// recordServiceBackedMetrics records metrics for service-backed executions.
func (mc *MigrationController) recordServiceBackedMetrics(toolName string, success bool, latencyMs int64) {
	mc.metrics.ServiceBackedExecutions[toolName]++
	if !success {
		mc.metrics.ServiceBackedErrors[toolName]++
	}

	if mc.metrics.ServiceBackedLatency[toolName] == nil {
		mc.metrics.ServiceBackedLatency[toolName] = make([]int64, 0, DefaultLatencyCapacity)
	}

	mc.metrics.ServiceBackedLatency[toolName] = append(
		mc.metrics.ServiceBackedLatency[toolName], latencyMs)
}
