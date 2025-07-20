package metrics_test

import (
	"errors"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/chadit/CloudMCP/pkg/metrics"
)

// Static errors for err113 compliance.
var (
	ErrTest = errors.New("test error")
)

// TestProviderInterface verifies that all backends implement the Provider interface correctly.
func TestProviderInterface(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		factory func() metrics.Provider
	}{
		{
			name:    "NoOp Provider",
			factory: metrics.NewNoOpProvider,
		},
		{
			name: "Prometheus Provider",
			factory: func() metrics.Provider {
				config := &metrics.ProviderConfig{
					Enabled:   true,
					Namespace: "test",
					Backend:   metrics.BackendPrometheus,
					BackendConfig: map[string]any{
						"prometheus": &metrics.PrometheusBackendConfig{
							Registry: prometheus.NewRegistry(),
						},
					},
				}
				provider, _ := metrics.NewProvider(config)

				return provider
			},
		},
		{
			name: "Log Provider",
			factory: func() metrics.Provider {
				config := &metrics.ProviderConfig{
					Enabled:   true,
					Namespace: "test",
					Backend:   metrics.BackendLog,
				}
				provider, _ := metrics.NewProvider(config)

				return provider
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			provider := tc.factory()

			// Test basic functionality.
			testProviderBasicOperations(t, provider)
			testProviderTimers(t, provider)
			testProviderGenericMetrics(t, provider)
		})
	}
}

// testProviderBasicOperations tests the basic provider operations.
func testProviderBasicOperations(_ *testing.T, provider metrics.Provider) {
	// Test tool execution.
	provider.RecordToolExecution("test_tool", "test_account", "success", time.Millisecond*100)

	// Test API request.
	provider.RecordAPIRequest("GET", "/test", "200", time.Millisecond*50)

	// Test cache operations.
	provider.RecordCacheHit("memory", "test_account")
	provider.RecordCacheMiss("memory", "test_account")

	// Test account switching.
	provider.RecordAccountSwitch("account1", "account2", "success")

	// Test connection tracking.
	provider.UpdateActiveConnections("test_account", 5)

	// Test resource tracking.
	provider.UpdateResourceCount("instances", "test_account", 10)

	// Test provider lifecycle.
	provider.RecordProviderLifecycle("test_provider", "init", "success", time.Millisecond*200)
	provider.UpdateProviderHealth("test_provider", true)
}

// testProviderTimers tests the timer functionality.
func testProviderTimers(_ *testing.T, provider metrics.Provider) {
	// Test tool execution timer.
	toolTimer := provider.NewToolExecutionTimer("test_tool", "test_account")

	time.Sleep(time.Millisecond) // Small delay to ensure measurable duration
	toolTimer.Finish("success")

	// Test tool execution timer with tags.
	toolTimerWithTags := provider.NewToolExecutionTimer("test_tool_2", "test_account")

	time.Sleep(time.Millisecond)
	toolTimerWithTags.FinishWithTags("success", map[string]string{"extra": "tag"})

	// Test API request timer.
	apiTimer := provider.NewAPIRequestTimer("POST", "/test")

	time.Sleep(time.Millisecond)
	apiTimer.Finish("201")

	// Test API request timer with tags.
	apiTimerWithTags := provider.NewAPIRequestTimer("PUT", "/test")

	time.Sleep(time.Millisecond)
	apiTimerWithTags.FinishWithTags("200", map[string]string{"region": "us-east"})
}

// testProviderGenericMetrics tests the generic metrics operations.
func testProviderGenericMetrics(_ *testing.T, provider metrics.Provider) {
	tags := map[string]string{
		"environment": "test",
		"component":   "unit_test",
	}

	// Test counter operations.
	provider.IncrementCounter("test_counter", tags)
	provider.IncrementCounterBy("test_counter_by", 5.0, tags)

	// Test gauge operations.
	provider.SetGauge("test_gauge", 42.0, tags)

	// Test histogram operations.
	provider.RecordHistogram("test_histogram", 3.14, tags)

	// Test timing operations.
	provider.RecordTiming("test_timing", time.Millisecond*150, tags)
}

// TestProviderFactory tests the provider factory functionality.
func TestProviderFactory(t *testing.T) {
	t.Parallel()

	factory := metrics.DefaultProviderFactory()

	// Test supported backends.
	backends := factory.SupportedBackends()

	if len(backends) == 0 {
		t.Error("Expected at least one supported backend")
	}

	expectedBackends := []string{metrics.BackendPrometheus, metrics.BackendNoOp, metrics.BackendLog}
	for _, expected := range expectedBackends {
		found := false

		for _, backend := range backends {
			if backend == expected {
				found = true

				break
			}
		}

		if !found {
			t.Errorf("Expected backend %s not found in supported backends", expected)
		}
	}

	// Test configuration validation.
	validConfig := metrics.DefaultProviderConfig()

	if err := factory.ValidateConfig(validConfig); err != nil {
		t.Errorf("Expected valid config to pass validation, got error: %v", err)
	}

	// Test invalid configuration.
	invalidConfig := &metrics.ProviderConfig{
		Enabled:   true,
		Namespace: "", // Invalid - empty namespace
		Backend:   metrics.BackendPrometheus,
	}
	if err := factory.ValidateConfig(invalidConfig); err == nil {
		t.Error("Expected invalid config to fail validation")
	}

	// Test provider creation.
	provider, err := factory.CreateProvider(validConfig)
	if err != nil {
		t.Errorf("Expected provider creation to succeed, got error: %v", err)
	}

	if provider == nil {
		t.Error("Expected non-nil provider")
	}
}

// TestPrometheusBackendConfiguration tests Prometheus-specific configuration.
func TestPrometheusBackendConfiguration(t *testing.T) {
	t.Parallel()

	defaultConfig := metrics.DefaultPrometheusBackendConfig()

	if defaultConfig == nil {
		t.Error("Expected non-nil default Prometheus config")

		return
	}

	testConfig := metrics.TestPrometheusBackendConfig()

	if testConfig == nil {
		t.Error("Expected non-nil test Prometheus config")

		return
	}

	// Test that test config has isolated registry.
	if testConfig.Registry == defaultConfig.Registry {
		t.Error("Expected test config to have separate registry from default")
	}
}

// TestConstants verifies that the backend constants are defined correctly.
func TestConstants(t *testing.T) {
	t.Parallel()

	constants := map[string]string{
		"BackendPrometheus": metrics.BackendPrometheus,
		"BackendNoOp":       metrics.BackendNoOp,
		"BackendLog":        metrics.BackendLog,
	}

	for name, value := range constants {
		if value == "" {
			t.Errorf("Expected constant %s to have non-empty value", name)
		}
	}

	// Test that constants are unique.
	values := make(map[string]bool)
	for _, value := range constants {
		if values[value] {
			t.Errorf("Duplicate constant value found: %s", value)
		}

		values[value] = true
	}
}

// TestNoOpProviderComprehensive tests all NoOp provider functionality.
func TestNoOpProviderComprehensive(t *testing.T) {
	t.Parallel()

	provider := metrics.NewNoOpProvider()

	// Test basic methods.
	if provider.IsEnabled() {
		t.Error("NoOp provider should not be enabled")
	}

	// Test all recording methods (should not panic).
	provider.RecordToolExecution("test", "test", "success", time.Second)
	provider.RecordAPIRequest("GET", "/test", "200", time.Second)
	provider.RecordCacheHit("memory", "test")
	provider.RecordCacheMiss("memory", "test")
	provider.RecordAccountSwitch("from", "to", "success")
	provider.UpdateActiveConnections("test", 5)
	provider.UpdateResourceCount("instances", "test", 10)

	// Test counter operations.
	provider.IncrementCounter("test_counter", map[string]string{"test": "value"})
	provider.IncrementCounterBy("test_counter", 5, map[string]string{"test": "value"})

	// Test gauge operations.
	provider.SetGauge("test_gauge", 42, map[string]string{"test": "value"})

	// Test histogram operations.
	provider.RecordHistogram("test_histogram", 3.14, map[string]string{"test": "value"})
	provider.RecordTiming("test_timing", time.Millisecond*150, map[string]string{"test": "value"})

	// Test provider lifecycle.
	provider.RecordProviderLifecycle("test_provider", "init", "success", time.Millisecond*200)
	provider.UpdateProviderHealth("test_provider", true)

	// Test timers.
	toolTimer := provider.NewToolExecutionTimer("test_tool", "test_account")
	toolTimer.Finish("success")
	toolTimer.FinishWithTags("success", map[string]string{"extra": "tag"})

	apiTimer := provider.NewAPIRequestTimer("GET", "/test")
	apiTimer.Finish("success")
	apiTimer.FinishWithTags("success", map[string]string{"extra": "tag"})
}

// TestConfigurationFunctions tests all configuration-related functions.
func TestConfigurationFunctions(t *testing.T) {
	t.Parallel()

	// Test DefaultConfig.
	defaultConfig := metrics.DefaultConfig()

	if defaultConfig.Enabled != true {
		t.Error("Default config should be enabled")
	}

	if defaultConfig.Namespace != "cloudmcp" {
		t.Error("Default namespace should be 'cloudmcp'")
	}

	// Test TestConfig.
	testConfig := metrics.TestConfig()

	if testConfig.Enabled != true {
		t.Error("Test config should be enabled")
	}

	if testConfig.Namespace != "cloudmcp_test" {
		t.Error("Test namespace should be 'cloudmcp_test'")
	}

	// Test TestConfigWithNamespace.
	customNamespace := "custom_test"
	customConfig := metrics.TestConfigWithNamespace(customNamespace)

	if customConfig.Enabled != true {
		t.Error("Custom test config should be enabled")
	}

	if customConfig.Namespace != customNamespace {
		t.Errorf("Custom namespace should be '%s', got '%s'", customNamespace, customConfig.Namespace)
	}
}

// TestCollectorFunctions tests collector-related functions.
func TestCollectorFunctions(t *testing.T) {
	t.Parallel()

	// Test NewCollector with nil config (should use default).
	collector := metrics.NewCollector(nil)

	if collector == nil {
		t.Error("NewCollector should not return nil")
	}

	// Test disabled collector.
	disabledConfig := &metrics.Config{
		Enabled:   false,
		Namespace: "test",
		Subsystem: "",
		Registry:  nil,
	}
	disabledCollector := metrics.NewCollector(disabledConfig)

	if disabledCollector.IsEnabled() {
		t.Error("Disabled collector should not be enabled")
	}

	// Test all recording methods on disabled collector (should not panic).
	disabledCollector.RecordToolExecution("test", "test", "success", time.Second)
	disabledCollector.RecordAPIRequest("GET", "/test", "200", time.Second)
	disabledCollector.RecordCacheHit("memory", "test")
	disabledCollector.RecordCacheMiss("memory", "test")
	disabledCollector.RecordAccountSwitch("from", "to", "success")
	disabledCollector.UpdateActiveConnections("test", 5)
	disabledCollector.UpdateResourceCount("instances", "test", 10)

	// Test GetMetricsRegistry.
	registry := metrics.GetMetricsRegistry()

	if registry == nil {
		t.Error("GetMetricsRegistry should not return nil")
	}
}

// TestTimerFunctionality tests timer creation and usage.
func TestTimerFunctionality(t *testing.T) {
	t.Parallel()

	// Test timer creation through provider interface.
	provider := metrics.NewNoOpProvider()

	// Test tool execution timer.
	toolTimer := provider.NewToolExecutionTimer("test_tool", "test_account")

	if toolTimer == nil {
		t.Error("NewToolExecutionTimer should not return nil")
	}

	toolTimer.Finish("success")

	// Test API request timer.
	apiTimer := provider.NewAPIRequestTimer("GET", "/test")

	if apiTimer == nil {
		t.Error("NewAPIRequestTimer should not return nil")
	}

	apiTimer.Finish("200")
}

// TestProviderConfigMethods tests provider configuration methods.
func TestProviderConfigMethods(t *testing.T) {
	t.Parallel()

	// Test DefaultProviderConfig.
	defaultConfig := metrics.DefaultProviderConfig()

	if !defaultConfig.Enabled {
		t.Error("Default provider config should be enabled")
	}

	if defaultConfig.Backend != "prometheus" {
		t.Error("Default backend should be prometheus")
	}

	// Test DefaultPrometheusBackendConfig.
	promConfig := metrics.DefaultPrometheusBackendConfig()

	if promConfig == nil {
		t.Error("Default prometheus config should not be nil")
	}

	// Test TestPrometheusBackendConfig.
	testPromConfig := metrics.TestPrometheusBackendConfig()

	if testPromConfig == nil {
		t.Error("Test prometheus config should not be nil")
	}
}

// TestProviderFactoryMethods tests the provider factory methods.
func TestProviderFactoryMethods(t *testing.T) {
	t.Parallel()

	// Test NewProvider.
	config := metrics.DefaultProviderConfig()

	provider, err := metrics.NewProvider(config)
	if err != nil {
		t.Errorf("NewProvider should not return error: %v", err)
	}

	if provider == nil {
		t.Error("NewProvider should not return nil")
	}

	// Test SetProviderFactory (should not panic).
	// Save original factory to restore after test.
	originalFactory := metrics.GetProviderFactory()
	metrics.SetProviderFactory(metrics.DefaultProviderFactory())

	// Restore the original factory to avoid affecting other tests.
	defer metrics.SetProviderFactory(originalFactory)

	// Test NewPrometheusProvider.
	promConfig := metrics.DefaultProviderConfig()

	promProvider, err := metrics.NewPrometheusProvider(promConfig)
	if err != nil {
		t.Errorf("NewPrometheusProvider should not return error: %v", err)
	}

	if promProvider == nil {
		t.Error("NewPrometheusProvider should not return nil")
	}
}

// TestErrorConditions tests error conditions and edge cases for better coverage.
func TestErrorConditions(t *testing.T) {
	t.Parallel()

	// Test NewPrometheusProvider with nil config (should use default).
	provider, err := metrics.NewPrometheusProvider(nil)
	if err != nil {
		t.Errorf("NewPrometheusProvider with nil config should not error: %v", err)
	}

	if provider == nil {
		t.Error("NewPrometheusProvider should not return nil")
	}

	// Test provider factory validation with invalid config.
	factory := metrics.DefaultProviderFactory()

	// Test with nil config.
	err = factory.ValidateConfig(nil)

	if err == nil {
		t.Error("ValidateConfig should return error for nil config")
	}

	// Test with enabled but empty namespace.
	invalidConfig := &metrics.ProviderConfig{
		Enabled:   true,
		Namespace: "",
		Backend:   metrics.BackendPrometheus,
	}
	err = factory.ValidateConfig(invalidConfig)

	if err == nil {
		t.Error("ValidateConfig should return error for empty namespace when enabled")
	}

	// Test with enabled but empty backend.
	invalidConfig2 := &metrics.ProviderConfig{
		Enabled:   true,
		Namespace: "test",
		Backend:   "",
	}
	err = factory.ValidateConfig(invalidConfig2)

	if err == nil {
		t.Error("ValidateConfig should return error for empty backend when enabled")
	}

	// Test with unsupported backend.
	invalidConfig3 := &metrics.ProviderConfig{
		Enabled:   true,
		Namespace: "test",
		Backend:   "unsupported_backend",
	}
	err = factory.ValidateConfig(invalidConfig3)

	if err == nil {
		t.Error("ValidateConfig should return error for unsupported backend")
	}
}
