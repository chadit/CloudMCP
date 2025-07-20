// Package types provides common types and error handling utilities for the CloudMCP system.
// It defines structured error types that provide context about which service and tool.
// encountered an error, enabling better error reporting and debugging across the MCP server.
package types

import "fmt"

type CloudMCPServerError struct {
	Service string
	Tool    string
	Message string
	Err     error
}

func (e *CloudMCPServerError) Error() string {
	if e.Tool != "" {
		return fmt.Sprintf("[%s/%s] %s", e.Service, e.Tool, e.Message)
	}

	return fmt.Sprintf("[%s] %s", e.Service, e.Message)
}

func (e *CloudMCPServerError) Unwrap() error {
	return e.Err
}

func NewServiceError(service, message string, err error) error {
	return &CloudMCPServerError{
		Service: service,
		Message: message,
		Err:     err,
	}
}

func NewToolError(service, tool, message string, err error) error {
	return &CloudMCPServerError{
		Service: service,
		Tool:    tool,
		Message: message,
		Err:     err,
	}
}
