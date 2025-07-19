package metrics

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

// Static errors for err113 compliance.
var (
	ErrUnsupportedBackend      = errors.New("unsupported metrics backend")
	ErrConfigurationNil        = errors.New("configuration cannot be nil")
	ErrNamespaceEmpty          = errors.New("namespace cannot be empty when metrics are enabled")
	ErrBackendNotSpecified     = errors.New("backend must be specified when metrics are enabled")
	ErrInvalidPrometheusConfig = errors.New("invalid prometheus configuration type")
	ErrPrometheusRegistryNil   = errors.New("prometheus registry is nil")
)

const (
	// MetricNameComponents is the maximum number of components in a metric name.
	MetricNameComponents = 3 // namespace, subsystem, name
)

// defaultProviderFactory implements ProviderFactory and provides the standard.
// metrics provider creation functionality.
type defaultProviderFactory struct {
	mutex            sync.RWMutex
	backendFactories map[string]BackendFactory
}

// DefaultProviderFactory creates a new provider factory with default backend support.
//
//nolint:ireturn // DefaultProviderFactory returns interface to provide factory abstraction.
func DefaultProviderFactory() ProviderFactory {
	factory := &defaultProviderFactory{
		backendFactories: make(map[string]BackendFactory),
	}

	// Register default backend factories.
	factory.RegisterBackendFactory(&prometheusBackendFactory{})
	factory.RegisterBackendFactory(&noOpBackendFactory{})
	factory.RegisterBackendFactory(&logBackendFactory{})

	return factory
}

// RegisterBackendFactory registers a backend factory for a specific backend type.
func (f *defaultProviderFactory) RegisterBackendFactory(factory BackendFactory) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.backendFactories[factory.BackendName()] = factory
}

// CreateProvider implements ProviderFactory.
//
//nolint:ireturn // CreateProvider returns interface to provide metrics provider abstraction.
func (f *defaultProviderFactory) CreateProvider(config *ProviderConfig) (Provider, error) {
	if config == nil {
		config = DefaultProviderConfig()
	}

	// Validate configuration.
	if err := f.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid metrics provider configuration: %w", err)
	}

	// Handle disabled case.
	if !config.Enabled {
		return &noOpProvider{}, nil
	}

	// Get backend factory.
	f.mutex.RLock()
	backendFactory, exists := f.backendFactories[config.Backend]
	f.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedBackend, config.Backend)
	}

	// Create backend.
	backend, err := backendFactory.CreateBackend(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics backend %s: %w", config.Backend, err)
	}

	// Create provider wrapping the backend.
	provider := &unifiedProvider{
		config:  config,
		backend: backend,
		enabled: config.Enabled,
	}

	// Initialize the backend.
	if err := backend.Start(); err != nil {
		return nil, fmt.Errorf("failed to start metrics backend %s: %w", config.Backend, err)
	}

	return provider, nil
}

// SupportedBackends implements ProviderFactory.
func (f *defaultProviderFactory) SupportedBackends() []string {
	backends := make([]string, 0, len(f.backendFactories))
	for name := range f.backendFactories {
		backends = append(backends, name)
	}

	return backends
}

// ValidateConfig implements ProviderFactory.
func (f *defaultProviderFactory) ValidateConfig(config *ProviderConfig) error {
	if config == nil {
		return ErrConfigurationNil
	}

	// Validate namespace.
	if config.Enabled && strings.TrimSpace(config.Namespace) == "" {
		return ErrNamespaceEmpty
	}

	// Validate backend.
	if config.Enabled {
		if config.Backend == "" {
			return ErrBackendNotSpecified
		}

		f.mutex.RLock()
		backendFactory, exists := f.backendFactories[config.Backend]
		f.mutex.RUnlock()

		if !exists {
			return fmt.Errorf("%w: %s", ErrUnsupportedBackend, config.Backend)
		}

		// Validate backend-specific configuration.
		if err := backendFactory.ValidateConfig(config); err != nil {
			return fmt.Errorf("invalid configuration for backend %s: %w", config.Backend, err)
		}
	}

	return nil
}

// noOpBackendFactory creates no-operation backends.
type noOpBackendFactory struct{}

// CreateBackend implements BackendFactory.
//
//nolint:ireturn // CreateBackend returns interface to allow multiple backend implementations.
func (f *noOpBackendFactory) CreateBackend(_ *ProviderConfig) (Backend, error) {
	return &noOpBackend{}, nil
}

// BackendName implements BackendFactory.
func (f *noOpBackendFactory) BackendName() string {
	return BackendNoOp
}

// ValidateConfig implements BackendFactory.
func (f *noOpBackendFactory) ValidateConfig(_ *ProviderConfig) error {
	// No-op backend accepts any configuration.
	return nil
}

// logBackendFactory creates log-based backends.
type logBackendFactory struct{}

// CreateBackend implements BackendFactory.
//
//nolint:ireturn // CreateBackend returns interface to allow multiple backend implementations.
func (f *logBackendFactory) CreateBackend(config *ProviderConfig) (Backend, error) {
	return &logBackend{
		namespace: config.Namespace,
		subsystem: config.Subsystem,
		tags:      copyTags(config.Tags),
	}, nil
}

// BackendName implements BackendFactory.
func (f *logBackendFactory) BackendName() string {
	return BackendLog
}

// ValidateConfig implements BackendFactory.
func (f *logBackendFactory) ValidateConfig(_ *ProviderConfig) error {
	// Log backend has minimal validation requirements.
	return nil
}

// copyTags creates a copy of the tags map to avoid shared references.
func copyTags(tags map[string]string) map[string]string {
	if tags == nil {
		return make(map[string]string)
	}

	copied := make(map[string]string, len(tags))
	for k, v := range tags {
		copied[k] = v
	}

	return copied
}

// mergeTags combines two tag maps, with the second map taking precedence for duplicate keys.
func mergeTags(base, additional map[string]string) map[string]string {
	if base == nil && additional == nil {
		return make(map[string]string)
	}

	if base == nil {
		return copyTags(additional)
	}

	if additional == nil {
		return copyTags(base)
	}

	merged := make(map[string]string, len(base)+len(additional))

	for k, v := range base {
		merged[k] = v
	}

	for k, v := range additional {
		merged[k] = v
	}

	return merged
}

// buildMetricName constructs a metric name from namespace, subsystem, and name components.
func buildMetricName(namespace, subsystem, name string) string {
	parts := make([]string, 0, MetricNameComponents)

	if namespace != "" {
		parts = append(parts, namespace)
	}

	if subsystem != "" {
		parts = append(parts, subsystem)
	}

	if name != "" {
		parts = append(parts, name)
	}

	return strings.Join(parts, "_")
}
