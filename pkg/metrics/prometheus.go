package metrics

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// prometheusBackendFactory creates Prometheus-based metrics backends.
type prometheusBackendFactory struct{}

// CreateBackend implements BackendFactory.
//
//nolint:ireturn // CreateBackend returns interface to provide backend abstraction.
func (f *prometheusBackendFactory) CreateBackend(config *ProviderConfig) (Backend, error) {
	promConfig, err := extractPrometheusConfig(config)
	if err != nil {
		return nil, err
	}

	backend := &prometheusBackend{
		namespace: config.Namespace,
		subsystem: config.Subsystem,
		registry:  promConfig.Registry,
		buckets:   promConfig.DefaultBuckets,
		enabled:   config.Enabled,
		metrics:   make(map[string]any),
	}

	return backend, nil
}

// BackendName implements BackendFactory.
func (f *prometheusBackendFactory) BackendName() string {
	return BackendPrometheus
}

// ValidateConfig implements BackendFactory.
func (f *prometheusBackendFactory) ValidateConfig(config *ProviderConfig) error {
	_, err := extractPrometheusConfig(config)

	return err
}

// extractPrometheusConfig extracts Prometheus-specific configuration from the generic config.
func extractPrometheusConfig(config *ProviderConfig) (*PrometheusBackendConfig, error) {
	if config.BackendConfig == nil {
		return DefaultPrometheusBackendConfig(), nil
	}

	// Try to extract from BackendConfig.
	if promConfig, ok := config.BackendConfig["prometheus"]; ok {
		if promConfigTyped, ok := promConfig.(*PrometheusBackendConfig); ok {
			return promConfigTyped, nil
		}

		return nil, ErrInvalidPrometheusConfig
	}

	// Fall back to defaults.

	return DefaultPrometheusBackendConfig(), nil
}

// prometheusBackend implements Backend using Prometheus metrics.
type prometheusBackend struct {
	namespace string
	subsystem string
	registry  prometheus.Registerer
	buckets   []float64
	enabled   bool

	mu      sync.RWMutex
	metrics map[string]any // Cache for created metrics
}

// Name implements Backend.
func (b *prometheusBackend) Name() string {
	return BackendPrometheus
}

// IsEnabled implements Backend.
func (b *prometheusBackend) IsEnabled() bool {
	return b.enabled
}

// getOrCreateMetric is a helper method that implements the double-checked locking pattern.
// for creating Prometheus metrics with consistent behavior across all metric types.
func (b *prometheusBackend) getOrCreateMetric(name string, tags map[string]string, createFunc func(promauto.Factory, string, []string) any) any {
	metricName := buildMetricName(b.namespace, b.subsystem, name)
	labelNames := b.extractLabelNames(tags)

	b.mu.RLock()
	if metric, exists := b.metrics[metricName]; exists {
		b.mu.RUnlock()

		return metric
	}
	b.mu.RUnlock()

	// Create new metric.
	b.mu.Lock()
	defer b.mu.Unlock()

	// Double-check pattern.
	if metric, exists := b.metrics[metricName]; exists {
		return metric
	}

	// Create new metric using the provided factory function.
	factory := promauto.With(b.registry)
	metric := createFunc(factory, metricName, labelNames)
	b.metrics[metricName] = metric

	return metric
}

// Counter implements Backend.
//
//nolint:ireturn // Counter returns interface to allow multiple metric backend implementations.
func (b *prometheusBackend) Counter(name string, tags map[string]string) CounterMetric {
	metric := b.getOrCreateMetric(name, tags, func(factory promauto.Factory, metricName string, labelNames []string) any {
		return factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: b.namespace,
				Subsystem: b.subsystem,
				Name:      name,
				Help:      "Counter metric: " + metricName,
			},
			labelNames,
		)
	})

	counterVec, ok := metric.(*prometheus.CounterVec)

	if !ok {
		// This should never happen with our current implementation.
		panic("internal error: metric type mismatch for counter")
	}

	labelNames := b.extractLabelNames(tags)

	return &prometheusCounter{
		counter: counterVec.WithLabelValues(b.extractLabelValues(tags, labelNames)...),
	}
}

// Gauge implements Backend.
//
//nolint:ireturn // Gauge returns interface to allow multiple metric backend implementations.
func (b *prometheusBackend) Gauge(name string, tags map[string]string) GaugeMetric {
	metric := b.getOrCreateMetric(name, tags, func(factory promauto.Factory, metricName string, labelNames []string) any {
		return factory.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: b.namespace,
				Subsystem: b.subsystem,
				Name:      name,
				Help:      "Gauge metric: " + metricName,
			},
			labelNames,
		)
	})

	gaugeVec, ok := metric.(*prometheus.GaugeVec)

	if !ok {
		// This should never happen with our current implementation.
		panic("internal error: metric type mismatch for gauge")
	}

	labelNames := b.extractLabelNames(tags)

	return &prometheusGauge{
		gauge: gaugeVec.WithLabelValues(b.extractLabelValues(tags, labelNames)...),
	}
}

// Histogram implements Backend.
//
//nolint:ireturn // Histogram returns interface to allow multiple metric backend implementations.
func (b *prometheusBackend) Histogram(name string, tags map[string]string) HistogramMetric {
	metric := b.getOrCreateMetric(name, tags, func(factory promauto.Factory, metricName string, labelNames []string) any {
		return factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: b.namespace,
				Subsystem: b.subsystem,
				Name:      name,
				Help:      "Histogram metric: " + metricName,
				Buckets:   b.buckets,
			},
			labelNames,
		)
	})

	histogramVec, ok := metric.(*prometheus.HistogramVec)

	if !ok {
		// This should never happen with our current implementation.
		panic("internal error: metric type mismatch for histogram")
	}

	labelNames := b.extractLabelNames(tags)

	return &prometheusHistogram{
		observer: histogramVec.WithLabelValues(b.extractLabelValues(tags, labelNames)...),
	}
}

// Timer implements Backend.
//
//nolint:ireturn // Timer returns interface to allow multiple metric backend implementations.
func (b *prometheusBackend) Timer(name string, tags map[string]string) TimerMetric {
	// Use histogram for timing metrics.
	histogram := b.Histogram(name, tags)

	return &prometheusTimer{
		histogram: histogram,
	}
}

// Start implements Backend.
func (b *prometheusBackend) Start() error {
	// Prometheus doesn't require explicit start.
	return nil
}

// Stop implements Backend.
func (b *prometheusBackend) Stop() error {
	// Prometheus doesn't require explicit stop.
	return nil
}

// Health implements Backend.
func (b *prometheusBackend) Health() error {
	// Simple health check - verify registry is accessible.
	if b.registry == nil {
		return ErrPrometheusRegistryNil
	}

	return nil
}

// extractLabelNames extracts sorted label names from tags for consistent metric registration.
func (b *prometheusBackend) extractLabelNames(tags map[string]string) []string {
	if len(tags) == 0 {
		return nil
	}

	labelNames := make([]string, 0, len(tags))
	for name := range tags {
		labelNames = append(labelNames, name)
	}

	// Sort to ensure consistent ordering.
	for i := range len(labelNames) - 1 {
		for j := range len(labelNames) - i - 1 {
			if labelNames[i] > labelNames[i+j+1] {
				labelNames[i], labelNames[i+j+1] = labelNames[i+j+1], labelNames[i]
			}
		}
	}

	return labelNames
}

// extractLabelValues extracts label values in the same order as labelNames.
func (b *prometheusBackend) extractLabelValues(tags map[string]string, labelNames []string) []string {
	if len(labelNames) == 0 {
		return nil
	}

	labelValues := make([]string, len(labelNames))
	for i, name := range labelNames {
		labelValues[i] = tags[name]
	}

	return labelValues
}

// prometheusCounter implements CounterMetric for Prometheus.
type prometheusCounter struct {
	counter prometheus.Counter
}

// Inc implements CounterMetric.
func (c *prometheusCounter) Inc() {
	c.counter.Inc()
}

// Add implements CounterMetric.
func (c *prometheusCounter) Add(value float64) {
	c.counter.Add(value)
}

// prometheusGauge implements GaugeMetric for Prometheus.
type prometheusGauge struct {
	gauge prometheus.Gauge
}

// Set implements GaugeMetric.
func (g *prometheusGauge) Set(value float64) {
	g.gauge.Set(value)
}

// Inc implements GaugeMetric.
func (g *prometheusGauge) Inc() {
	g.gauge.Inc()
}

// Dec implements GaugeMetric.
func (g *prometheusGauge) Dec() {
	g.gauge.Dec()
}

// Add implements GaugeMetric.
func (g *prometheusGauge) Add(value float64) {
	g.gauge.Add(value)
}

// Sub implements GaugeMetric.
func (g *prometheusGauge) Sub(value float64) {
	g.gauge.Sub(value)
}

// prometheusHistogram implements HistogramMetric for Prometheus.
type prometheusHistogram struct {
	observer prometheus.Observer
}

// Observe implements HistogramMetric.
func (h *prometheusHistogram) Observe(value float64) {
	h.observer.Observe(value)
}

// ObserveDuration implements HistogramMetric.
func (h *prometheusHistogram) ObserveDuration(duration time.Duration) {
	h.observer.Observe(duration.Seconds())
}

// prometheusTimer implements TimerMetric for Prometheus.
type prometheusTimer struct {
	histogram HistogramMetric
}

// Start implements TimerMetric.
//
//nolint:ireturn // Start returns interface to allow multiple timer instance implementations.
func (t *prometheusTimer) Start() TimerInstance {
	return &prometheusTimerInstance{
		start:     time.Now(),
		histogram: t.histogram,
	}
}

// Record implements TimerMetric.
func (t *prometheusTimer) Record(duration time.Duration) {
	t.histogram.ObserveDuration(duration)
}

// prometheusTimerInstance implements TimerInstance for Prometheus.
type prometheusTimerInstance struct {
	start     time.Time
	histogram HistogramMetric
}

// Stop implements TimerInstance.
func (i *prometheusTimerInstance) Stop() {
	duration := time.Since(i.start)
	i.histogram.ObserveDuration(duration)
}

// Elapsed implements TimerInstance.
func (i *prometheusTimerInstance) Elapsed() time.Duration {
	return time.Since(i.start)
}
