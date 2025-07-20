package security

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// SecureTestFramework provides secure token handling capabilities for integration tests.
type SecureTestFramework struct {
	validator   *TokenValidator
	testLogger  TestLogger
	environment map[string]string
}

// TestLogger interface for secure logging in tests.
type TestLogger interface {
	Logf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Helper()
}

// NewSecureTestFramework creates a new secure test framework instance.
func NewSecureTestFramework(logger TestLogger) *SecureTestFramework {
	return &SecureTestFramework{
		validator:   NewTokenValidator(),
		testLogger:  logger,
		environment: make(map[string]string),
	}
}

// SetupSecureEnvironment prepares a secure testing environment with token validation.
func (f *SecureTestFramework) SetupSecureEnvironment(t *testing.T) context.Context {
	t.Helper()
	f.testLogger.Helper()

	f.testLogger.Logf("🔒 Setting up secure test environment...")

	// Load environment variables securely
	f.loadEnvironmentVariables()

	// Validate all tokens before tests
	f.validateAllTokens(t)

	// Create context with timeout for tests
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	
	// Cleanup function to ensure no token leakage
	t.Cleanup(func() {
		f.cleanupSecureEnvironment()
		cancel()
	})

	return ctx
}

// loadEnvironmentVariables securely loads environment variables.
func (f *SecureTestFramework) loadEnvironmentVariables() {
	tokenNames := []string{
		"LINODE_TOKEN",
		"AWS_ACCESS_KEY_ID", 
		"AWS_SECRET_ACCESS_KEY",
		"GITHUB_TOKEN",
		"TEST_TOKEN",
	}

	for _, tokenName := range tokenNames {
		value := os.Getenv(tokenName)
		f.environment[tokenName] = value
	}
}

// validateAllTokens validates all tokens and fails the test if required tokens are invalid.
func (f *SecureTestFramework) validateAllTokens(t *testing.T) {
	t.Helper()

	results := f.validator.ValidateEnvironmentTokens(func(key string) string {
		return f.environment[key]
	})

	validationErrors := 0
	for tokenName, result := range results {
		if !result.Valid && result.Error != nil {
			// Check if this is a required token or has an actual validation error
			if f.environment[tokenName] != "" {
				f.testLogger.Errorf("Token validation failed for %s: %s (redacted: %s)", 
					tokenName, result.Error.Error(), result.Redacted)
				validationErrors++
			} else {
				f.testLogger.Logf("Optional token %s is not set", tokenName)
			}
		} else {
			f.testLogger.Logf("Token %s: %s", tokenName, 
				LogSecureTokenValidation(tokenName, result))
		}
	}

	// Only fail if there are actual validation errors (not missing optional tokens)
	if validationErrors > 0 {
		require.FailNow(t, "Token validation failed", 
			"Found %d token validation errors. Check token configuration.", validationErrors)
	}

	f.testLogger.Logf("✅ All token validations passed")
}

// GetSecureToken retrieves a token with validation and redaction logging.
func (f *SecureTestFramework) GetSecureToken(t *testing.T, tokenName string) string {
	t.Helper()
	f.testLogger.Helper()

	token := f.environment[tokenName]
	
	// Validate token before returning
	result := f.validator.ValidateToken(tokenName, token)
	
	if !result.Valid && result.Error != nil && token != "" {
		// Only fail for invalid non-empty tokens
		require.FailNow(t, "Invalid token", 
			"Token %s validation failed: %s (redacted: %s)", 
			tokenName, result.Error.Error(), result.Redacted)
	}

	f.testLogger.Logf("Retrieved token %s: %s", tokenName, 
		LogSecureTokenValidation(tokenName, result))

	return token
}

// RequireValidToken ensures a token is present and valid, failing the test if not.
func (f *SecureTestFramework) RequireValidToken(t *testing.T, tokenName string) string {
	t.Helper()
	f.testLogger.Helper()

	token := f.GetSecureToken(t, tokenName)
	
	if token == "" {
		require.FailNow(t, "Required token missing", 
			"Token %s is required for this test but not provided", tokenName)
	}

	return token
}

// SkipIfTokenMissing skips a test if a required token is not available.
func (f *SecureTestFramework) SkipIfTokenMissing(t *testing.T, tokenName string) {
	t.Helper()
	f.testLogger.Helper()

	token := f.environment[tokenName]
	if token == "" {
		t.Skipf("Skipping test: %s token not provided", tokenName)
	}

	// Validate token if present
	result := f.validator.ValidateToken(tokenName, token)
	if !result.Valid && result.Error != nil {
		t.Skipf("Skipping test: %s token validation failed: %s", 
			tokenName, result.Error.Error())
	}
}

// cleanupSecureEnvironment performs secure cleanup after tests.
func (f *SecureTestFramework) cleanupSecureEnvironment() {
	// Clear environment map to prevent token leakage
	for key := range f.environment {
		f.environment[key] = ""
	}
	clear(f.environment)
}

// GenerateTestSecurityReport generates a security report for test execution.
func (f *SecureTestFramework) GenerateTestSecurityReport() SecurityAuditReport {
	return f.validator.GenerateSecurityAuditReport(func(key string) string {
		return f.environment[key]
	})
}

// SecureIntegrationTest provides a helper for integration tests with secure token handling.
func SecureIntegrationTest(t *testing.T, testName string, testFunc func(t *testing.T, ctx context.Context, framework *SecureTestFramework)) {
	t.Helper()

	// Create test logger adapter
	testLogger := &testLoggerAdapter{t: t}
	
	// Setup secure framework
	framework := NewSecureTestFramework(testLogger)
	
	t.Logf("🔒 Starting secure integration test: %s", testName)
	
	// Setup secure environment
	ctx := framework.SetupSecureEnvironment(t)
	
	// Run the test function with secure context
	testFunc(t, ctx, framework)
	
	// Generate security report
	report := framework.GenerateTestSecurityReport()
	t.Logf("🔒 Test security report: %d total tokens, %d valid, %d invalid", 
		report.TotalTokens, report.ValidTokens, report.InvalidTokens)
}

// testLoggerAdapter adapts *testing.T to TestLogger interface.
type testLoggerAdapter struct {
	t *testing.T
}

func (a *testLoggerAdapter) Logf(format string, args ...interface{}) {
	a.t.Helper()
	a.t.Logf(format, args...)
}

func (a *testLoggerAdapter) Errorf(format string, args ...interface{}) {
	a.t.Helper()
	a.t.Errorf(format, args...)
}

func (a *testLoggerAdapter) Helper() {
	a.t.Helper()
}

// MockTokenEnvironment creates a mock environment for testing token validation.
func MockTokenEnvironment() map[string]string {
	return map[string]string{
		"LINODE_TOKEN":           "abcdef123456789012345678901234567890abcd", // Valid hex
		"AWS_ACCESS_KEY_ID":      "AKIATEST12345678ABCD",                      // Valid alphanumeric
		"AWS_SECRET_ACCESS_KEY":  "test_secret_key_1234567890abcdefghijklmnop", // Valid alphanumeric
		"GITHUB_TOKEN":           "ghp_1234567890abcdefghijklmnopqrstuvwxyz12", // Valid GitHub token format
		"TEST_TOKEN":             "deadbeef12345678",                          // Valid hex test token
		"INVALID_TOKEN":          "invalid_format_xyz",                        // Invalid for hex format
		"EMPTY_TOKEN":            "",                                          // Empty token
	}
}

// WithMockEnvironment sets up tests with a mock token environment.
func WithMockEnvironment(testFunc func(getEnv func(string) string)) {
	mockEnv := MockTokenEnvironment()
	getEnv := func(key string) string {
		return mockEnv[key]
	}
	testFunc(getEnv)
}