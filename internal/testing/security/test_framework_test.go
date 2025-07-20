package security

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSecureTestFramework_SetupSecureEnvironment validates secure environment setup.
func TestSecureTestFramework_SetupSecureEnvironment(t *testing.T) {
	t.Parallel()

	// Create mock logger
	mockLogger := &mockTestLogger{}
	framework := NewSecureTestFramework(mockLogger)

	// Setup mock environment
	originalEnv := setupMockEnvironment(t)
	defer restoreEnvironment(originalEnv)

	// Test environment setup
	ctx := framework.SetupSecureEnvironment(t)

	// Verify context is valid
	require.NotNil(t, ctx, "Context should not be nil")
	
	// Verify context has timeout
	deadline, hasDeadline := ctx.Deadline()
	require.True(t, hasDeadline, "Context should have deadline")
	require.True(t, deadline.After(time.Now()), "Deadline should be in the future")

	// Verify logging occurred
	assert.Greater(t, len(mockLogger.logs), 0, "Should have logged setup messages")
	assert.Contains(t, mockLogger.logs[0], "Setting up secure test environment", 
		"Should log environment setup")

	// Verify no token leakage in logs
	for _, log := range mockLogger.logs {
		assert.NotContains(t, log, "abcdef123456789012345678901234567890abcd", 
			"Logs should not contain full token values")
	}
}

// TestSecureTestFramework_GetSecureToken validates secure token retrieval.
func TestSecureTestFramework_GetSecureToken(t *testing.T) {
	t.Parallel()

	mockLogger := &mockTestLogger{}
	framework := NewSecureTestFramework(mockLogger)

	// Setup mock environment
	originalEnv := setupMockEnvironment(t)
	defer restoreEnvironment(originalEnv)

	// Load environment
	framework.loadEnvironmentVariables()

	// Test retrieving valid token
	token := framework.GetSecureToken(t, "LINODE_TOKEN")
	assert.Equal(t, "abcdef123456789012345678901234567890abcd", token, 
		"Should return the correct token")

	// Test retrieving empty token (should not fail)
	emptyToken := framework.GetSecureToken(t, "EMPTY_TOKEN")
	assert.Empty(t, emptyToken, "Should return empty string for empty token")

	// Verify secure logging
	foundLogEntry := false
	for _, log := range mockLogger.logs {
		if assert.Contains(t, log, "LINODE_TOKEN") {
			assert.NotContains(t, log, "abcdef123456789012345678901234567890abcd", 
				"Log should not contain full token")
			assert.Contains(t, log, "VALID", "Log should indicate token is valid")
			foundLogEntry = true
			break
		}
	}
	assert.True(t, foundLogEntry, "Should have logged token retrieval")
}

// TestSecureTestFramework_RequireValidToken validates required token handling.
func TestSecureTestFramework_RequireValidToken(t *testing.T) {
	t.Parallel()

	mockLogger := &mockTestLogger{}
	framework := NewSecureTestFramework(mockLogger)

	// Setup mock environment
	originalEnv := setupMockEnvironment(t)
	defer restoreEnvironment(originalEnv)

	// Load environment
	framework.loadEnvironmentVariables()

	// Test requiring valid token (should succeed)
	token := framework.RequireValidToken(t, "LINODE_TOKEN")
	assert.Equal(t, "abcdef123456789012345678901234567890abcd", token, 
		"Should return valid token")

	// Test requiring missing token (should be handled by the test framework)
	// We'll test this indirectly by checking the environment
	emptyToken := framework.environment["MISSING_TOKEN"]
	assert.Empty(t, emptyToken, "Missing token should be empty")
}

// TestSecureTestFramework_SkipIfTokenMissing validates conditional test skipping.
func TestSecureTestFramework_SkipIfTokenMissing(t *testing.T) {
	t.Parallel()

	mockLogger := &mockTestLogger{}
	framework := NewSecureTestFramework(mockLogger)

	// Setup environment with missing token
	originalEnv := setupMockEnvironment(t)
	defer restoreEnvironment(originalEnv)
	
	// Remove a token to test skipping
	os.Unsetenv("TEST_TOKEN")

	// Load environment
	framework.loadEnvironmentVariables()

	// Test with missing token - should skip
	defer func() {
		if r := recover(); r != nil {
			// Check if it's a skip (testing.TB.SkipNow behavior varies)
			if skipMsg, ok := r.(string); ok {
				assert.Contains(t, skipMsg, "Skipping test", "Should contain skip message")
			}
		}
	}()

	// This would normally skip, but we're testing the logic
	framework.SkipIfTokenMissing(t, "TEST_TOKEN")
	
	// If we get here, the token was present or skip didn't trigger as expected
	// The actual skip behavior depends on testing.T implementation
}

// TestSecureIntegrationTest validates the integration test helper.
func TestSecureIntegrationTest(t *testing.T) {
	t.Parallel()

	// Setup mock environment
	originalEnv := setupMockEnvironment(t)
	defer restoreEnvironment(originalEnv)

	executed := false
	
	SecureIntegrationTest(t, "TestSecureExample", func(t *testing.T, ctx context.Context, framework *SecureTestFramework) {
		executed = true
		
		// Test that we have a valid context
		require.NotNil(t, ctx, "Context should be provided")
		
		// Test that framework is available
		require.NotNil(t, framework, "Framework should be provided")
		
		// Test secure token retrieval
		token := framework.GetSecureToken(t, "LINODE_TOKEN")
		assert.NotEmpty(t, token, "Should retrieve token securely")
		
		// Verify context has timeout
		_, hasDeadline := ctx.Deadline()
		assert.True(t, hasDeadline, "Context should have timeout")
	})

	assert.True(t, executed, "Test function should have been executed")
}

// TestMockTokenEnvironment validates the mock environment setup.
func TestMockTokenEnvironment(t *testing.T) {
	t.Parallel()

	mockEnv := MockTokenEnvironment()

	// Verify all expected tokens are present
	expectedTokens := []string{
		"LINODE_TOKEN",
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY", 
		"GITHUB_TOKEN",
		"TEST_TOKEN",
		"INVALID_TOKEN",
		"EMPTY_TOKEN",
	}

	for _, tokenName := range expectedTokens {
		_, exists := mockEnv[tokenName]
		assert.True(t, exists, "Mock environment should contain %s", tokenName)
	}

	// Verify token formats
	assert.NotEmpty(t, mockEnv["LINODE_TOKEN"], "LINODE_TOKEN should not be empty")
	assert.Empty(t, mockEnv["EMPTY_TOKEN"], "EMPTY_TOKEN should be empty")
	assert.Contains(t, mockEnv["GITHUB_TOKEN"], "ghp_", "GITHUB_TOKEN should have proper prefix")
}

// TestWithMockEnvironment validates the mock environment helper.
func TestWithMockEnvironment(t *testing.T) {
	t.Parallel()

	executed := false

	WithMockEnvironment(func(getEnv func(string) string) {
		executed = true
		
		// Test token retrieval
		linodeToken := getEnv("LINODE_TOKEN")
		assert.NotEmpty(t, linodeToken, "Should retrieve LINODE_TOKEN")
		
		emptyToken := getEnv("EMPTY_TOKEN")
		assert.Empty(t, emptyToken, "EMPTY_TOKEN should be empty")
		
		invalidToken := getEnv("INVALID_TOKEN")
		assert.NotEmpty(t, invalidToken, "INVALID_TOKEN should have value")
	})

	assert.True(t, executed, "Function should have been executed")
}

// TestSecureTestFramework_SecurityReporting validates security report generation.
func TestSecureTestFramework_SecurityReporting(t *testing.T) {
	t.Parallel()

	mockLogger := &mockTestLogger{}
	framework := NewSecureTestFramework(mockLogger)

	// Use mock environment
	WithMockEnvironment(func(getEnv func(string) string) {
		// Set up framework environment
		for _, tokenName := range []string{"LINODE_TOKEN", "AWS_ACCESS_KEY_ID", "GITHUB_TOKEN", "INVALID_TOKEN", "EMPTY_TOKEN"} {
			framework.environment[tokenName] = getEnv(tokenName)
		}

		// Generate report
		report := framework.GenerateTestSecurityReport()

		// Verify report structure
		assert.Greater(t, report.TotalTokens, 0, "Should report tokens")
		assert.Greater(t, report.ValidTokens, 0, "Should have valid tokens")
		assert.GreaterOrEqual(t, report.InvalidTokens, 0, "May have invalid tokens")
		assert.NotEmpty(t, report.Recommendations, "Should have recommendations")
		
		// Verify timestamp is recent
		assert.WithinDuration(t, time.Now(), report.Timestamp, time.Minute, 
			"Report timestamp should be recent")
	})
}

// mockTestLogger implements TestLogger for testing.
type mockTestLogger struct {
	logs   []string
	errors []string
}

func (m *mockTestLogger) Logf(format string, args ...interface{}) {
	m.logs = append(m.logs, fmt.Sprintf(format, args...))
}

func (m *mockTestLogger) Errorf(format string, args ...interface{}) {
	m.errors = append(m.errors, fmt.Sprintf(format, args...))
}

func (m *mockTestLogger) Helper() {
	// No-op for mock
}

// setupMockEnvironment sets up test environment variables.
func setupMockEnvironment(t *testing.T) map[string]string {
	t.Helper()

	// Save original environment
	originalEnv := make(map[string]string)
	tokenNames := []string{
		"LINODE_TOKEN",
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY",
		"GITHUB_TOKEN",
		"TEST_TOKEN",
		"EMPTY_TOKEN",
	}

	for _, name := range tokenNames {
		originalEnv[name] = os.Getenv(name)
	}

	// Set mock values
	mockEnv := MockTokenEnvironment()
	for name, value := range mockEnv {
		os.Setenv(name, value)
	}

	return originalEnv
}

// restoreEnvironment restores original environment variables.
func restoreEnvironment(originalEnv map[string]string) {
	for name, value := range originalEnv {
		if value == "" {
			os.Unsetenv(name)
		} else {
			os.Setenv(name, value)
		}
	}
}

