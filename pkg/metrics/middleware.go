package metrics

import "context"

// Middleware wraps tool execution with metrics collection.
type Middleware struct {
	collector *Collector
	next      func(ctx context.Context, tool string, account string) error
}

// NewMiddleware creates a middleware that collects metrics for tool execution.
func NewMiddleware(collector *Collector, next func(ctx context.Context, tool string, account string) error) *Middleware {
	return &Middleware{
		collector: collector,
		next:      next,
	}
}

// Execute wraps the execution with metrics collection.
func (m *Middleware) Execute(ctx context.Context, tool string, account string) error {
	timer := m.collector.NewToolExecutionTimer(tool, account)

	err := m.next(ctx, tool, account)

	status := "success"
	if err != nil {
		status = "error"
	}

	timer.Finish(status)

	return err
}
