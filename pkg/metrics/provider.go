package metrics

import (
	"time"
)

// unifiedProvider implements Provider by wrapping a Backend.
// and providing the unified metrics interface that supports both specific.
// and generic metrics operations.
type unifiedProvider struct {
	config  *ProviderConfig
	backend Backend
	enabled bool
}

// IsEnabled implements Provider.
func (p *unifiedProvider) IsEnabled() bool {
	return p.enabled && p.backend.IsEnabled()
}

// RecordToolExecution implements Provider.
func (p *unifiedProvider) RecordToolExecution(tool, account, status string, duration time.Duration) {
	if !p.IsEnabled() {
		return
	}

	tags := mergeTags(p.config.Tags, map[string]string{
		"tool":    tool,
		"account": account,
		"status":  status,
	})

	// Record duration as histogram.
	durationMetric := p.backend.Histogram("tool_execution_duration_seconds", tags)
	durationMetric.ObserveDuration(duration)

	// Record count as counter.
	countMetric := p.backend.Counter("tool_execution_total", tags)
	countMetric.Inc()
}

// NewToolExecutionTimer implements Provider.
//
//nolint:ireturn // NewToolExecutionTimer returns Timer interface to provide timing abstraction for tool execution metrics.
func (p *unifiedProvider) NewToolExecutionTimer(tool, account string) ToolTimer {
	return &toolTimer{
		provider: p,
		tool:     tool,
		account:  account,
		start:    time.Now(),
	}
}

// RecordAPIRequest implements Provider.
func (p *unifiedProvider) RecordAPIRequest(method, endpoint, status string, duration time.Duration) {
	if !p.IsEnabled() {
		return
	}

	tags := mergeTags(p.config.Tags, map[string]string{
		"method":   method,
		"endpoint": endpoint,
		"status":   status,
	})

	// Record duration as histogram.
	durationMetric := p.backend.Histogram("api_duration_seconds", tags)
	durationMetric.ObserveDuration(duration)

	// Record count as counter.
	countMetric := p.backend.Counter("api_requests_total", tags)
	countMetric.Inc()
}

// NewAPIRequestTimer implements Provider.
//
//nolint:ireturn // NewAPIRequestTimer returns Timer interface to provide timing abstraction for API request metrics.
func (p *unifiedProvider) NewAPIRequestTimer(method, endpoint string) APITimer {
	return &apiTimer{
		provider: p,
		method:   method,
		endpoint: endpoint,
		start:    time.Now(),
	}
}

// RecordCacheHit implements Provider.
func (p *unifiedProvider) RecordCacheHit(cacheType, account string) {
	if !p.IsEnabled() {
		return
	}

	tags := mergeTags(p.config.Tags, map[string]string{
		"cache_type": cacheType,
		"account":    account,
	})

	countMetric := p.backend.Counter("cache_hits_total", tags)
	countMetric.Inc()
}

// RecordCacheMiss implements Provider.
func (p *unifiedProvider) RecordCacheMiss(cacheType, account string) {
	if !p.IsEnabled() {
		return
	}

	tags := mergeTags(p.config.Tags, map[string]string{
		"cache_type": cacheType,
		"account":    account,
	})

	countMetric := p.backend.Counter("cache_misses_total", tags)
	countMetric.Inc()
}

// RecordAccountSwitch implements Provider.
func (p *unifiedProvider) RecordAccountSwitch(fromAccount, toAccount, status string) {
	if !p.IsEnabled() {
		return
	}

	tags := mergeTags(p.config.Tags, map[string]string{
		"from_account": fromAccount,
		"to_account":   toAccount,
		"status":       status,
	})

	countMetric := p.backend.Counter("account_switches_total", tags)
	countMetric.Inc()
}

// UpdateActiveConnections implements Provider.
func (p *unifiedProvider) UpdateActiveConnections(account string, count int) {
	if !p.IsEnabled() {
		return
	}

	tags := mergeTags(p.config.Tags, map[string]string{
		"account": account,
	})

	gaugeMetric := p.backend.Gauge("active_connections", tags)
	gaugeMetric.Set(float64(count))
}

// UpdateResourceCount implements Provider.
func (p *unifiedProvider) UpdateResourceCount(resourceType, account string, count int) {
	if !p.IsEnabled() {
		return
	}

	tags := mergeTags(p.config.Tags, map[string]string{
		"resource_type": resourceType,
		"account":       account,
	})

	gaugeMetric := p.backend.Gauge("resources", tags)
	gaugeMetric.Set(float64(count))
}

// IncrementCounter implements Provider.
func (p *unifiedProvider) IncrementCounter(name string, tags map[string]string) {
	p.IncrementCounterBy(name, 1.0, tags)
}

// IncrementCounterBy implements Provider.
func (p *unifiedProvider) IncrementCounterBy(name string, value float64, tags map[string]string) {
	if !p.IsEnabled() {
		return
	}

	mergedTags := mergeTags(p.config.Tags, tags)
	countMetric := p.backend.Counter(name, mergedTags)
	countMetric.Add(value)
}

// SetGauge implements Provider.
func (p *unifiedProvider) SetGauge(name string, value float64, tags map[string]string) {
	if !p.IsEnabled() {
		return
	}

	mergedTags := mergeTags(p.config.Tags, tags)
	gaugeMetric := p.backend.Gauge(name, mergedTags)
	gaugeMetric.Set(value)
}

// RecordHistogram implements Provider.
func (p *unifiedProvider) RecordHistogram(name string, value float64, tags map[string]string) {
	if !p.IsEnabled() {
		return
	}

	mergedTags := mergeTags(p.config.Tags, tags)
	histogramMetric := p.backend.Histogram(name, mergedTags)
	histogramMetric.Observe(value)
}

// RecordTiming implements Provider.
func (p *unifiedProvider) RecordTiming(name string, duration time.Duration, tags map[string]string) {
	if !p.IsEnabled() {
		return
	}

	mergedTags := mergeTags(p.config.Tags, tags)
	histogramMetric := p.backend.Histogram(name, mergedTags)
	histogramMetric.ObserveDuration(duration)
}

// RecordProviderLifecycle implements Provider.
func (p *unifiedProvider) RecordProviderLifecycle(provider, event, status string, duration time.Duration) {
	if !p.IsEnabled() {
		return
	}

	tags := mergeTags(p.config.Tags, map[string]string{
		"provider": provider,
		"event":    event,
		"status":   status,
	})

	// Record duration if meaningful.
	if duration > 0 {
		durationMetric := p.backend.Histogram("provider_lifecycle_duration_seconds", tags)
		durationMetric.ObserveDuration(duration)
	}

	// Record count.
	countMetric := p.backend.Counter("provider_lifecycle_events_total", tags)
	countMetric.Inc()
}

// UpdateProviderHealth implements Provider.
func (p *unifiedProvider) UpdateProviderHealth(provider string, healthy bool) {
	if !p.IsEnabled() {
		return
	}

	tags := mergeTags(p.config.Tags, map[string]string{
		"provider": provider,
	})

	gaugeMetric := p.backend.Gauge("provider_health", tags)

	if healthy {
		gaugeMetric.Set(1.0)
	} else {
		gaugeMetric.Set(0.0)
	}
}

// toolTimer implements ToolTimer interface.
type toolTimer struct {
	provider *unifiedProvider
	tool     string
	account  string
	start    time.Time
}

// Finish implements ToolTimer.
func (t *toolTimer) Finish(status string) {
	t.FinishWithTags(status, nil)
}

// FinishWithTags implements ToolTimer.
func (t *toolTimer) FinishWithTags(status string, tags map[string]string) {
	duration := time.Since(t.start)

	// Record using the provider's RecordToolExecution method.
	t.provider.RecordToolExecution(t.tool, t.account, status, duration)

	// Record as additional generic timing metric with different name to avoid label conflicts.
	if len(tags) > 0 {
		finalTags := mergeTags(tags, map[string]string{
			"tool":    t.tool,
			"account": t.account,
			"status":  status,
		})
		t.provider.RecordTiming("tool_execution_extra_duration_seconds", duration, finalTags)
	}
}

// apiTimer implements APITimer interface.
type apiTimer struct {
	provider *unifiedProvider
	method   string
	endpoint string
	start    time.Time
}

// Finish implements APITimer.
func (t *apiTimer) Finish(status string) {
	t.FinishWithTags(status, nil)
}

// FinishWithTags implements APITimer.
func (t *apiTimer) FinishWithTags(status string, tags map[string]string) {
	duration := time.Since(t.start)

	// Record using the provider's RecordAPIRequest method.
	t.provider.RecordAPIRequest(t.method, t.endpoint, status, duration)

	// Record as additional generic timing metric with different name to avoid label conflicts.
	if len(tags) > 0 {
		finalTags := mergeTags(tags, map[string]string{
			"method":   t.method,
			"endpoint": t.endpoint,
			"status":   status,
		})
		t.provider.RecordTiming("api_request_extra_duration_seconds", duration, finalTags)
	}
}

// noOpProvider implements Provider with no-operation methods.
type noOpProvider struct{}

// IsEnabled implements Provider.
func (n *noOpProvider) IsEnabled() bool { return false }

// RecordToolExecution implements Provider.
func (n *noOpProvider) RecordToolExecution(_, _, _ string, _ time.Duration) {}

// NewToolExecutionTimer implements Provider.
//
//nolint:ireturn // NewToolExecutionTimer implements Null Object pattern - returns Timer interface for consistent API.
func (n *noOpProvider) NewToolExecutionTimer(_, _ string) ToolTimer {
	return &noOpToolTimer{}
}

// RecordAPIRequest implements Provider.
func (n *noOpProvider) RecordAPIRequest(_, _, _ string, _ time.Duration) {}

// NewAPIRequestTimer implements Provider.
//
//nolint:ireturn // NewAPIRequestTimer implements Null Object pattern - returns Timer interface for consistent API.
func (n *noOpProvider) NewAPIRequestTimer(_, _ string) APITimer {
	return &noOpAPITimer{}
}

// RecordCacheHit implements Provider.
func (n *noOpProvider) RecordCacheHit(_, _ string) {}

// RecordCacheMiss implements Provider.
func (n *noOpProvider) RecordCacheMiss(_, _ string) {}

// RecordAccountSwitch implements Provider.
func (n *noOpProvider) RecordAccountSwitch(_, _, _ string) {}

// UpdateActiveConnections implements Provider.
func (n *noOpProvider) UpdateActiveConnections(_ string, _ int) {}

// UpdateResourceCount implements Provider.
func (n *noOpProvider) UpdateResourceCount(_, _ string, _ int) {}

// IncrementCounter implements Provider.
func (n *noOpProvider) IncrementCounter(_ string, _ map[string]string) {}

// IncrementCounterBy implements Provider.
func (n *noOpProvider) IncrementCounterBy(_ string, _ float64, _ map[string]string) {
}

// SetGauge implements Provider.
func (n *noOpProvider) SetGauge(_ string, _ float64, _ map[string]string) {}

// RecordHistogram implements Provider.
func (n *noOpProvider) RecordHistogram(_ string, _ float64, _ map[string]string) {}

// RecordTiming implements Provider.
func (n *noOpProvider) RecordTiming(_ string, _ time.Duration, _ map[string]string) {}

// RecordProviderLifecycle implements Provider.
func (n *noOpProvider) RecordProviderLifecycle(_, _, _ string, _ time.Duration) {
}

// UpdateProviderHealth implements Provider.
func (n *noOpProvider) UpdateProviderHealth(_ string, _ bool) {}

// noOpToolTimer implements ToolTimer with no-operation methods.
type noOpToolTimer struct{}

// Finish implements ToolTimer.
func (n *noOpToolTimer) Finish(_ string) {}

// FinishWithTags implements ToolTimer.
func (n *noOpToolTimer) FinishWithTags(_ string, _ map[string]string) {}

// noOpAPITimer implements APITimer with no-operation methods.
type noOpAPITimer struct{}

// Finish implements APITimer.
func (n *noOpAPITimer) Finish(_ string) {}

// FinishWithTags implements APITimer.
func (n *noOpAPITimer) FinishWithTags(_ string, _ map[string]string) {}
