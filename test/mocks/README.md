# CloudMCP Test Mocks

This directory contains mock implementations for testing CloudMCP Linode service functionality.

## Overview

The mock interfaces provide testable implementations of Linode API operations without requiring actual API calls. These mocks are designed to support comprehensive unit testing of all CloudMCP tools.

## Available Mocks

### Core Services
- **`firewall_mock.go`** - Firewall management operations
- **`database_mock.go`** - MySQL and PostgreSQL database operations  
- **`nodebalancer_mock.go`** - Load balancer operations
- **`lke_mock.go`** - Kubernetes Engine (LKE) operations
- **`domain_mock.go`** - DNS domain and record management

### Usage Patterns

Each mock follows a consistent pattern:

1. **Mock Interface** - Implements relevant linodego.Client methods
2. **Test Helpers** - Convenience methods for common test scenarios
3. **Error Simulation** - Support for testing error conditions

## Example Usage

```go
package linode_test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/require"
    "github.com/chadit/CloudMCP/test/mocks"
    "github.com/linode/linodego"
)

func TestFirewallsList(t *testing.T) {
    // Setup mock
    mockFirewall := &mocks.MockFirewallService{}
    
    // Configure expected behavior
    expectedFirewalls := []linodego.Firewall{
        {
            ID:     1,
            Label:  "test-firewall",
            Status: "enabled",
        },
    }
    mockFirewall.SetupListFirewallsSuccess(expectedFirewalls)
    
    // Test your service logic here
    // service := NewService(mockFirewall)
    // result, err := service.HandleFirewallsList(context.Background(), request)
    
    // Verify expectations
    mockFirewall.AssertExpectations(t)
}

func TestFirewallsListError(t *testing.T) {
    // Setup mock for error scenario
    mockFirewall := &mocks.MockFirewallService{}
    
    expectedError := errors.New("API error")
    mockFirewall.SetupListFirewallsError(expectedError)
    
    // Test error handling logic here
    // Verify appropriate error responses
    
    mockFirewall.AssertExpectations(t)
}
```

## Mock Helper Methods

Each mock provides helper methods to simplify test setup:

### Success Scenarios
- `SetupListXXXSuccess(items)` - Mock successful list operations
- `SetupGetXXXSuccess(id, item)` - Mock successful get operations  
- `SetupCreateXXXSuccess(opts, item)` - Mock successful create operations

### Error Scenarios  
- `SetupListXXXError(err)` - Mock list operation failures
- `SetupGetXXXError(id, err)` - Mock get operation failures
- `SetupCreateXXXError(opts, err)` - Mock create operation failures

## Testing Best Practices

### 1. Test Both Success and Error Paths
```go
func TestDatabaseCreate_Success(t *testing.T) {
    // Test successful database creation
}

func TestDatabaseCreate_APIError(t *testing.T) {
    // Test API error handling
}

func TestDatabaseCreate_ValidationError(t *testing.T) {
    // Test input validation errors
}
```

### 2. Use Descriptive Test Names
- Follow pattern: `TestMethodName_Scenario`
- Examples: `TestFirewallGet_NotFound`, `TestLKEClusterCreate_InvalidRegion`

### 3. Verify Mock Expectations
Always call `mockService.AssertExpectations(t)` to ensure all expected calls were made.

### 4. Test Boundary Conditions
- Empty lists
- Invalid IDs
- Missing required fields
- Edge case values

## Test Data Helpers

Consider creating test data helpers for consistent test fixtures:

```go
// test/fixtures/firewall.go
func NewTestFirewall() *linodego.Firewall {
    return &linodego.Firewall{
        ID:     1,
        Label:  "test-firewall",
        Status: "enabled",
        Rules: linodego.FirewallRuleSet{
            InboundPolicy:  "ACCEPT",
            OutboundPolicy: "ACCEPT",
            Inbound:        []linodego.FirewallRule{},
            Outbound:       []linodego.FirewallRule{},
        },
        Tags: []string{"test"},
    }
}
```

## Integration with Service Tests

The mocks are designed to integrate with CloudMCP service testing:

```go
// Replace real linodego.Client with mock in tests
type TestService struct {
    *Service
    MockClient interface{
        *mocks.MockFirewallService
        *mocks.MockDatabaseService
        // ... other mocks
    }
}
```

## Mock Completeness

Current mock coverage includes:

### ✅ Implemented
- Firewall operations (create, read, update, delete, rules, devices)
- Database operations (MySQL & PostgreSQL CRUD, credentials)
- NodeBalancer operations (CRUD, configurations)
- LKE operations (clusters, node pools, kubeconfig)
- Domain operations (domains, DNS records)

### 🔄 Future Additions
- Object Storage operations
- Advanced Networking operations  
- StackScript operations
- Monitoring (Longview) operations
- Support ticket operations

## Contributing

When adding new mocks:

1. Follow the established naming conventions
2. Include both success and error helper methods
3. Add comprehensive documentation
4. Provide usage examples
5. Ensure thread-safety for concurrent tests

## Dependencies

- `github.com/stretchr/testify/mock` - Mock framework
- `github.com/linode/linodego` - Linode API types

## Notes

- Mocks are stateless and should be recreated for each test
- Use `mock.Anything` for context parameters unless specific context testing is needed
- Helper methods use sensible defaults for test data
- All mocks are designed to be thread-safe for parallel test execution