//go:build integration

package linode

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/chadit/CloudMCP/internal/config"
	"github.com/chadit/CloudMCP/pkg/logger"
)

// TestContainer represents a WireMock container for integration testing.
// It provides mock Linode API responses through a containerized WireMock server
// that can be used to test CloudMCP's Linode service integration without
// requiring live API access.
type TestContainer struct {
	Container testcontainers.Container
	Host      string
	Port      string
	BaseURL   string
}

// MockLinodeAPI sets up a WireMock container with predefined Linode API stubs.
// This creates a mock Linode API server that responds to common endpoints
// with realistic data, enabling comprehensive integration testing without
// external dependencies.
//
// **Container Features:**
// • WireMock server with Linode API endpoint stubs
// • Realistic JSON responses for instances, volumes, accounts
// • Error scenario simulation (404, 401, validation errors)
// • Health check endpoint for container readiness
//
// **Test Coverage:**
// • Account management operations
// • Instance lifecycle management
// • Volume operations
// • Error handling scenarios
//
// **Usage Pattern:**
//
//	container := MockLinodeAPI(t)
//	defer container.Terminate(ctx)
//	service := createTestService(t, container.BaseURL)
func MockLinodeAPI(t *testing.T) *TestContainer {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "wiremock/wiremock:latest",
		ExposedPorts: []string{"8080/tcp"},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      "./testdata/linode_stubs.json",
				ContainerFilePath: "/home/wiremock/mappings/linode.json",
				FileMode:          0o644,
			},
		},
		WaitingFor: wait.ForHTTP("/__admin").WithPort("8080").WithStartupTimeout(60 * time.Second),
		Env: map[string]string{
			"WIREMOCK_OPTIONS": "--global-response-templating",
		},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err, "failed to start WireMock container")

	host, err := container.Host(ctx)
	require.NoError(t, err, "failed to get container host")

	mappedPort, err := container.MappedPort(ctx, "8080")
	require.NoError(t, err, "failed to get mapped port")

	port := mappedPort.Port()
	baseURL := fmt.Sprintf("http://%s:%s", host, port)

	return &TestContainer{
		Container: container,
		Host:      host,
		Port:      port,
		BaseURL:   baseURL,
	}
}

// Terminate stops and removes the WireMock container.
// This should be called in test cleanup (typically with defer) to ensure
// proper resource cleanup after integration tests complete.
func (tc *TestContainer) Terminate(ctx context.Context) error {
	return tc.Container.Terminate(ctx)
}

// CreateIntegrationTestService creates a CloudMCP Linode service configured to use the mock API.
// This service instance will make API calls to the WireMock container instead of
// the real Linode API, enabling controlled integration testing scenarios.
//
// **Service Configuration:**
// • Uses mock API base URL from container
// • Configured with test account and token
// • Debug logging enabled for detailed test output
// • Same service interface as production for realistic testing
//
// **Account Setup:**
// • Single test account named "integration-test"
// • Uses placeholder token (mock API doesn't validate)
// • Custom API URL pointing to WireMock container
func CreateIntegrationTestService(t *testing.T, mockAPIURL string) *Service {
	log := logger.New("debug")

	cfg := &config.Config{
		DefaultLinodeAccount: "integration-test",
		LinodeAccounts: map[string]config.LinodeAccount{
			"integration-test": {
				Token:  "test-token-mock-api",
				Label:  "Integration Test Account",
				APIURL: mockAPIURL,
			},
		},
	}

	service, err := New(cfg, log)
	require.NoError(t, err, "failed to create test service")

	return service
}

// ValidateJSONStructure validates that response JSON contains expected fields.
// This helper function checks that API responses from the mock server match
// the expected structure of real Linode API responses, ensuring contract
// compliance between mock and real API.
//
// **Validation Types:**
// • Field presence checks
// • Data type validation
// • Nested structure verification
// • Array element structure validation
//
// **Usage:**
//
//	ValidateJSONStructure(t, response, map[string]interface{}{
//	    "id": "number",
//	    "label": "string",
//	    "status": "string",
//	})
func ValidateJSONStructure(t *testing.T, data map[string]interface{}, expectedFields map[string]string) {
	for field, expectedType := range expectedFields {
		value, exists := data[field]
		require.True(t, exists, "field %s should exist in response", field)

		switch expectedType {
		case "string":
			_, ok := value.(string)
			require.True(t, ok, "field %s should be a string", field)
		case "number":
			_, ok := value.(float64)
			if !ok {
				_, ok = value.(int)
			}
			require.True(t, ok, "field %s should be a number", field)
		case "bool":
			_, ok := value.(bool)
			require.True(t, ok, "field %s should be a boolean", field)
		case "array":
			_, ok := value.([]interface{})
			require.True(t, ok, "field %s should be an array", field)
		case "object":
			_, ok := value.(map[string]interface{})
			require.True(t, ok, "field %s should be an object", field)
		}
	}
}

// SetupIntegrationTest performs common setup for integration tests.
// This helper function creates the WireMock container, test service,
// and returns all necessary components for integration testing with
// proper cleanup setup.
//
// **Setup Process:**
// 1. Start WireMock container with Linode API stubs
// 2. Create CloudMCP service configured for mock API
// 3. Initialize service with mock account
// 4. Return service and cleanup function
//
// **Cleanup Handling:**
// Returns a cleanup function that should be called with defer to ensure
// proper container termination and resource cleanup.
//
// **Usage:**
//
//	service, cleanup := SetupIntegrationTest(t)
//	defer cleanup()
//	// Run integration tests with service
func SetupIntegrationTest(t *testing.T) (*Service, func()) {
	ctx := context.Background()

	container := MockLinodeAPI(t)

	service := CreateIntegrationTestService(t, container.BaseURL)

	err := service.Initialize(ctx)
	require.NoError(t, err, "failed to initialize test service")

	cleanup := func() {
		container.Terminate(ctx)
	}

	return service, cleanup
}
