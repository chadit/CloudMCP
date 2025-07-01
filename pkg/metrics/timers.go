package metrics

import "time"

// ToolExecutionTimer provides a convenient way to time tool executions.
type ToolExecutionTimer struct {
	collector *Collector
	tool      string
	account   string
	startTime time.Time
}

// NewToolExecutionTimer creates a new timer for tool execution.
func (c *Collector) NewToolExecutionTimer(tool, account string) *ToolExecutionTimer {
	return &ToolExecutionTimer{
		collector: c,
		tool:      tool,
		account:   account,
		startTime: time.Now(),
	}
}

// Finish records the tool execution metrics with the specified status.
func (t *ToolExecutionTimer) Finish(status string) {
	duration := time.Since(t.startTime)
	t.collector.RecordToolExecution(t.tool, t.account, status, duration)
}

// APIRequestTimer provides a convenient way to time API requests.
type APIRequestTimer struct {
	collector *Collector
	method    string
	endpoint  string
	startTime time.Time
}

// NewAPIRequestTimer creates a new timer for API requests.
func (c *Collector) NewAPIRequestTimer(method, endpoint string) *APIRequestTimer {
	return &APIRequestTimer{
		collector: c,
		method:    method,
		endpoint:  endpoint,
		startTime: time.Now(),
	}
}

// Finish records the API request metrics with the specified status.
func (t *APIRequestTimer) Finish(status string) {
	duration := time.Since(t.startTime)
	t.collector.RecordAPIRequest(t.method, t.endpoint, status, duration)
}
