package metrics

import (
	"log"
	"time"
)

// noOpBackend implements Backend with no-operation methods.
type noOpBackend struct{}

// Name implements Backend.
func (b *noOpBackend) Name() string {
	return BackendNoOp
}

// IsEnabled implements Backend.
func (b *noOpBackend) IsEnabled() bool {
	return false
}

// Counter implements Backend.
//
//nolint:ireturn // Counter returns interface to allow multiple metric implementations.
func (b *noOpBackend) Counter(_ string, _ map[string]string) CounterMetric {
	return &noOpCounterMetric{}
}

// Gauge implements Backend.
//
//nolint:ireturn // Gauge returns interface to allow multiple metric implementations.
func (b *noOpBackend) Gauge(_ string, _ map[string]string) GaugeMetric {
	return &noOpGaugeMetric{}
}

// Histogram implements Backend.
//
//nolint:ireturn // Histogram returns interface to allow multiple metric implementations.
func (b *noOpBackend) Histogram(_ string, _ map[string]string) HistogramMetric {
	return &noOpHistogramMetric{}
}

// Timer implements Backend.
//
//nolint:ireturn // Timer returns interface to allow multiple metric implementations.
func (b *noOpBackend) Timer(_ string, _ map[string]string) TimerMetric {
	return &noOpTimerMetric{}
}

// Start implements Backend.
func (b *noOpBackend) Start() error {
	return nil
}

// Stop implements Backend.
func (b *noOpBackend) Stop() error {
	return nil
}

// Health implements Backend.
func (b *noOpBackend) Health() error {
	return nil
}

// noOpCounterMetric implements CounterMetric with no-operation methods.
type noOpCounterMetric struct{}

// Inc implements CounterMetric.
func (m *noOpCounterMetric) Inc() {}

// Add implements CounterMetric.
func (m *noOpCounterMetric) Add(_ float64) {}

// noOpGaugeMetric implements GaugeMetric with no-operation methods.
type noOpGaugeMetric struct{}

// Set implements GaugeMetric.
func (m *noOpGaugeMetric) Set(_ float64) {}

// Inc implements GaugeMetric.
func (m *noOpGaugeMetric) Inc() {}

// Dec implements GaugeMetric.
func (m *noOpGaugeMetric) Dec() {}

// Add implements GaugeMetric.
func (m *noOpGaugeMetric) Add(_ float64) {}

// Sub implements GaugeMetric.
func (m *noOpGaugeMetric) Sub(_ float64) {}

// noOpHistogramMetric implements HistogramMetric with no-operation methods.
type noOpHistogramMetric struct{}

// Observe implements HistogramMetric.
func (m *noOpHistogramMetric) Observe(_ float64) {}

// ObserveDuration implements HistogramMetric.
func (m *noOpHistogramMetric) ObserveDuration(_ time.Duration) {}

// noOpTimerMetric implements TimerMetric with no-operation methods.
type noOpTimerMetric struct{}

// Start implements TimerMetric.
//
//nolint:ireturn // Start returns interface to allow multiple timer implementations.
func (m *noOpTimerMetric) Start() TimerInstance {
	return &noOpTimerInstance{}
}

// Record implements TimerMetric.
func (m *noOpTimerMetric) Record(_ time.Duration) {}

// noOpTimerInstance implements TimerInstance with no-operation methods.
type noOpTimerInstance struct{}

// Stop implements TimerInstance.
func (i *noOpTimerInstance) Stop() {}

// Elapsed implements TimerInstance.
func (i *noOpTimerInstance) Elapsed() time.Duration {
	return 0
}

// logBackend implements Backend using standard Go logging.
type logBackend struct {
	namespace string
	subsystem string
	tags      map[string]string
}

// Name implements Backend.
func (b *logBackend) Name() string {
	return BackendLog
}

// IsEnabled implements Backend.
func (b *logBackend) IsEnabled() bool {
	return true
}

// Counter implements Backend.
//
//nolint:ireturn // Counter returns interface to allow multiple metric implementations.
func (b *logBackend) Counter(name string, tags map[string]string) CounterMetric {
	metricName := buildMetricName(b.namespace, b.subsystem, name)
	mergedTags := mergeTags(b.tags, tags)

	return &logCounterMetric{
		name: metricName,
		tags: mergedTags,
	}
}

// Gauge implements Backend.
//
//nolint:ireturn // Gauge returns interface to allow multiple metric implementations.
func (b *logBackend) Gauge(name string, tags map[string]string) GaugeMetric {
	metricName := buildMetricName(b.namespace, b.subsystem, name)
	mergedTags := mergeTags(b.tags, tags)

	return &logGaugeMetric{
		name: metricName,
		tags: mergedTags,
	}
}

// Histogram implements Backend.
//
//nolint:ireturn // Histogram returns interface to allow multiple metric implementations.
func (b *logBackend) Histogram(name string, tags map[string]string) HistogramMetric {
	metricName := buildMetricName(b.namespace, b.subsystem, name)
	mergedTags := mergeTags(b.tags, tags)

	return &logHistogramMetric{
		name: metricName,
		tags: mergedTags,
	}
}

// Timer implements Backend.
//
//nolint:ireturn // Timer returns interface to allow multiple metric implementations.
func (b *logBackend) Timer(name string, tags map[string]string) TimerMetric {
	metricName := buildMetricName(b.namespace, b.subsystem, name)
	mergedTags := mergeTags(b.tags, tags)

	return &logTimerMetric{
		name: metricName,
		tags: mergedTags,
	}
}

// Start implements Backend.
func (b *logBackend) Start() error {
	log.Printf("Starting log metrics backend: namespace=%s, subsystem=%s", b.namespace, b.subsystem)

	return nil
}

// Stop implements Backend.
func (b *logBackend) Stop() error {
	log.Printf("Stopping log metrics backend: namespace=%s, subsystem=%s", b.namespace, b.subsystem)

	return nil
}

// Health implements Backend.
func (b *logBackend) Health() error {
	return nil
}

// logCounterMetric implements CounterMetric using logging.
type logCounterMetric struct {
	name string
	tags map[string]string
}

// Inc implements CounterMetric.
func (m *logCounterMetric) Inc() {
	m.Add(1.0)
}

// Add implements CounterMetric.
func (m *logCounterMetric) Add(value float64) {
	log.Printf("METRIC counter %s: +%f %v", m.name, value, m.tags)
}

// logGaugeMetric implements GaugeMetric using logging.
type logGaugeMetric struct {
	name string
	tags map[string]string
}

// Set implements GaugeMetric.
func (m *logGaugeMetric) Set(value float64) {
	log.Printf("METRIC gauge %s: =%f %v", m.name, value, m.tags)
}

// Inc implements GaugeMetric.
func (m *logGaugeMetric) Inc() {
	m.Add(1.0)
}

// Dec implements GaugeMetric.
func (m *logGaugeMetric) Dec() {
	m.Sub(1.0)
}

// Add implements GaugeMetric.
func (m *logGaugeMetric) Add(value float64) {
	log.Printf("METRIC gauge %s: +%f %v", m.name, value, m.tags)
}

// Sub implements GaugeMetric.
func (m *logGaugeMetric) Sub(value float64) {
	log.Printf("METRIC gauge %s: -%f %v", m.name, value, m.tags)
}

// logHistogramMetric implements HistogramMetric using logging.
type logHistogramMetric struct {
	name string
	tags map[string]string
}

// Observe implements HistogramMetric.
func (m *logHistogramMetric) Observe(value float64) {
	log.Printf("METRIC histogram %s: %f %v", m.name, value, m.tags)
}

// ObserveDuration implements HistogramMetric.
func (m *logHistogramMetric) ObserveDuration(duration time.Duration) {
	log.Printf("METRIC histogram %s: %v (%fs) %v", m.name, duration, duration.Seconds(), m.tags)
}

// logTimerMetric implements TimerMetric using logging.
type logTimerMetric struct {
	name string
	tags map[string]string
}

// Start implements TimerMetric.
//
//nolint:ireturn // Start returns interface to allow multiple timer implementations.
func (m *logTimerMetric) Start() TimerInstance {
	return &logTimerInstance{
		name:  m.name,
		tags:  m.tags,
		start: time.Now(),
	}
}

// Record implements TimerMetric.
func (m *logTimerMetric) Record(duration time.Duration) {
	log.Printf("METRIC timer %s: %v (%fs) %v", m.name, duration, duration.Seconds(), m.tags)
}

// logTimerInstance implements TimerInstance using logging.
type logTimerInstance struct {
	name  string
	tags  map[string]string
	start time.Time
}

// Stop implements TimerInstance.
func (i *logTimerInstance) Stop() {
	duration := time.Since(i.start)
	log.Printf("METRIC timer %s: %v (%fs) %v", i.name, duration, duration.Seconds(), i.tags)
}

// Elapsed implements TimerInstance.
func (i *logTimerInstance) Elapsed() time.Duration {
	return time.Since(i.start)
}
