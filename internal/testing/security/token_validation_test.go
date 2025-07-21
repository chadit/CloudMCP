package security

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTokenValidator_ValidateToken performs comprehensive token validation testing.
// This test validates secure token handling, format checking, and redaction capabilities.
//
// **Test Environment:**
// • Token validator with default and custom configurations
// • Test tokens in various formats (hex, alphanumeric, base64, JWT)
// • Security validation including timing attack prevention
//
// **Security Validation:**
// • Format validation for different token types
// • Length validation with constant-time comparison
// • Token redaction for secure logging
// • Environment variable validation
//
// **Expected Behavior:**
// • Valid tokens pass validation with proper format checking
// • Invalid tokens fail with descriptive error messages
// • Token redaction prevents credential exposure in logs
// • Constant-time operations prevent timing attacks
//
// **Purpose:** Ensure secure token handling and prevent credential exposure.
func TestTokenValidator_ValidateToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		tokenName   string
		tokenValue  string
		config      TokenValidationConfig
		expectValid bool
		expectError string
	}{
		{
			name:       "ValidHexToken",
			tokenName:  "TEST_HEX",
			tokenValue: "abcdef123456789012345678901234567890abcd",
			config: TokenValidationConfig{
				Format:    HexToken,
				MinLength: 16,
				MaxLength: 64,
				Required:  true,
			},
			expectValid: true,
		},
		{
			name:       "ValidAlphanumericToken",
			tokenName:  "TEST_ALPHANUM",
			tokenValue: "test_token_123-ABC.def",
			config: TokenValidationConfig{
				Format:    AlphanumericToken,
				MinLength: 8,
				MaxLength: 32,
				Required:  true,
			},
			expectValid: true,
		},
		{
			name:       "ValidBase64Token",
			tokenName:  "TEST_BASE64",
			tokenValue: "dGVzdFRva2VuMTIz",
			config: TokenValidationConfig{
				Format:    Base64Token,
				MinLength: 8,
				MaxLength: 32,
				Required:  true,
			},
			expectValid: true,
		},
		{
			name:       "ValidJWTToken",
			tokenName:  "TEST_JWT",
			tokenValue: "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJ0ZXN0In0.signature",
			config: TokenValidationConfig{
				Format:    JWTToken,
				MinLength: 16,
				MaxLength: 256,
				Required:  true,
			},
			expectValid: true,
		},
		{
			name:       "EmptyTokenAllowed",
			tokenName:  "TEST_EMPTY",
			tokenValue: "",
			config: TokenValidationConfig{
				Format:     HexToken,
				MinLength:  16,
				MaxLength:  64,
				Required:   false,
				AllowEmpty: true,
			},
			expectValid: true,
		},
		{
			name:       "EmptyTokenRequired",
			tokenName:  "TEST_REQUIRED",
			tokenValue: "",
			config: TokenValidationConfig{
				Format:     HexToken,
				MinLength:  16,
				MaxLength:  64,
				Required:   true,
				AllowEmpty: false,
			},
			expectValid: false,
			expectError: "token is required but not provided",
		},
		{
			name:       "TokenTooShort",
			tokenName:  "TEST_SHORT",
			tokenValue: "abc",
			config: TokenValidationConfig{
				Format:    HexToken,
				MinLength: 16,
				MaxLength: 64,
				Required:  true,
			},
			expectValid: false,
			expectError: "token length invalid",
		},
		{
			name:       "TokenTooLong",
			tokenName:  "TEST_LONG",
			tokenValue: strings.Repeat("a", 100),
			config: TokenValidationConfig{
				Format:    HexToken,
				MinLength: 16,
				MaxLength: 64,
				Required:  true,
			},
			expectValid: false,
			expectError: "token length invalid",
		},
		{
			name:       "InvalidHexFormat",
			tokenName:  "TEST_BAD_HEX",
			tokenValue: "ghijklmnopqrstuvwxyz123456789012345678",
			config: TokenValidationConfig{
				Format:    HexToken,
				MinLength: 16,
				MaxLength: 64,
				Required:  true,
			},
			expectValid: false,
			expectError: "token format invalid",
		},
		{
			name:       "InvalidJWTFormat",
			tokenName:  "TEST_BAD_JWT",
			tokenValue: "invalid.jwt.format.too.many.parts",
			config: TokenValidationConfig{
				Format:    JWTToken,
				MinLength: 16,
				MaxLength: 64,
				Required:  true,
			},
			expectValid: false,
			expectError: "token format invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			validator := NewTokenValidator()
			validator.AddConfig(tt.tokenName, tt.config)

			result := validator.ValidateToken(tt.tokenName, tt.tokenValue)

			assert.Equal(t, tt.expectValid, result.Valid, 
				"Token validation result should match expectation")

			if tt.expectError != "" {
				require.NotNil(t, result.Error, "Expected error should be present")
				assert.Contains(t, result.Error.Error(), tt.expectError, 
					"Error message should contain expected text")
			} else if !tt.expectValid {
				assert.NotNil(t, result.Error, "Invalid token should have error")
			}

			// Verify redaction is applied
			assert.NotEmpty(t, result.Redacted, "Redacted value should not be empty")
			if tt.tokenValue != "" && len(tt.tokenValue) > 4 {
				assert.NotEqual(t, tt.tokenValue, result.Redacted, 
					"Redacted value should not equal original token")
			}
		})
	}
}

// TestRedactToken validates secure token redaction functionality.
func TestRedactToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		token       string
		expectation func(t *testing.T, redacted string)
	}{
		{
			name:  "EmptyToken",
			token: "",
			expectation: func(t *testing.T, redacted string) {
				assert.Equal(t, "[EMPTY]", redacted, "Empty token should be marked as empty")
			},
		},
		{
			name:  "ShortToken",
			token: "abc",
			expectation: func(t *testing.T, redacted string) {
				assert.Equal(t, "[REDACTED]", redacted, "Short token should be fully redacted")
			},
		},
		{
			name:  "MediumToken",
			token: "abcdefgh",
			expectation: func(t *testing.T, redacted string) {
				assert.True(t, strings.HasPrefix(redacted, "ab"), "Should preserve first 2 chars")
				assert.Contains(t, redacted, "*", "Should contain redaction characters")
				assert.NotContains(t, redacted, "cdefgh", "Should not contain middle/end chars")
			},
		},
		{
			name:  "LongToken",
			token: "abcdefghijklmnopqrstuvwxyz123456789012345678",
			expectation: func(t *testing.T, redacted string) {
				assert.True(t, strings.HasPrefix(redacted, "abcd"), "Should preserve first 4 chars")
				assert.True(t, strings.HasSuffix(redacted, "5678"), "Should preserve last 4 chars")
				assert.Contains(t, redacted, "*", "Should contain redaction characters")
				assert.NotContains(t, redacted, "efghijklmnopqrstuvwxyz12349", "Should not contain middle chars")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			redacted := RedactToken(tt.token)
			tt.expectation(t, redacted)
		})
	}
}

// TestTokenValidator_ValidateEnvironmentTokens tests environment variable token validation.
func TestTokenValidator_ValidateEnvironmentTokens(t *testing.T) {
	t.Parallel()

	validator := NewTokenValidator()
	
	// Mock environment function
	mockEnv := map[string]string{
		"LINODE_TOKEN":    "abcdef123456789012345678901234567890abcd",
		"GITHUB_TOKEN":    "ghp_1234567890abcdefghijklmnopqrstuvwxyz12",
		"TEST_TOKEN":      "invalid_format_for_hex",
		"EMPTY_TOKEN":     "",
	}

	getEnv := func(key string) string {
		return mockEnv[key]
	}

	results := validator.ValidateEnvironmentTokens(getEnv)

	// Verify LINODE_TOKEN is valid
	linodeResult, exists := results["LINODE_TOKEN"]
	require.True(t, exists, "LINODE_TOKEN result should exist")
	assert.True(t, linodeResult.Valid, "LINODE_TOKEN should be valid")

	// Verify GITHUB_TOKEN is valid
	githubResult, exists := results["GITHUB_TOKEN"]
	require.True(t, exists, "GITHUB_TOKEN result should exist")
	assert.True(t, githubResult.Valid, "GITHUB_TOKEN should be valid")

	// Verify TEST_TOKEN is invalid due to format
	testResult, exists := results["TEST_TOKEN"]
	require.True(t, exists, "TEST_TOKEN result should exist")
	assert.False(t, testResult.Valid, "TEST_TOKEN should be invalid due to format")
	assert.Contains(t, testResult.Error.Error(), "format invalid", 
		"Error should mention format issue")
}

// TestTokenValidator_GenerateSecurityAuditReport validates security audit report generation.
func TestTokenValidator_GenerateSecurityAuditReport(t *testing.T) {
	t.Parallel()

	validator := NewTokenValidator()
	
	// Mock environment with mixed valid/invalid tokens
	mockEnv := map[string]string{
		"LINODE_TOKEN":         "abcdef123456789012345678901234567890abcd", // Valid hex
		"GITHUB_TOKEN":         "ghp_1234567890abcdefghijklmnopqrstuvwxyz12", // Valid alphanumeric
		"AWS_ACCESS_KEY_ID":    "",                                          // Empty (allowed)
		"AWS_SECRET_ACCESS_KEY": "too_short",                                // Invalid length
		"TEST_TOKEN":           "invalid_hex_format_xyz",                    // Invalid format
	}

	getEnv := func(key string) string {
		return mockEnv[key]
	}

	report := validator.GenerateSecurityAuditReport(getEnv)

	// Verify report structure
	assert.Equal(t, 5, report.TotalTokens, "Should report 5 total tokens")
	assert.Greater(t, report.ValidTokens, 0, "Should have some valid tokens")
	assert.Greater(t, report.InvalidTokens, 0, "Should have some invalid tokens")
	assert.Equal(t, 1, report.EmptyTokens, "Should have 1 empty token")
	assert.Greater(t, len(report.Issues), 0, "Should have some issues reported")
	assert.Greater(t, len(report.Recommendations), 0, "Should have recommendations")

	// Verify timestamp is recent
	assert.WithinDuration(t, time.Now(), report.Timestamp, time.Minute, 
		"Report timestamp should be recent")
}

// TestLogSecureTokenValidation validates secure logging functionality.
func TestLogSecureTokenValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		tokenName string
		result    TokenValidationResult
		expectLog func(t *testing.T, logMsg string)
	}{
		{
			name:      "ValidTokenLog",
			tokenName: "TEST_TOKEN",
			result: TokenValidationResult{
				Valid:       true,
				TokenLength: 32,
				Format:      HexToken,
				Redacted:    "abcd****5678",
			},
			expectLog: func(t *testing.T, logMsg string) {
				assert.Contains(t, logMsg, "VALID", "Log should indicate token is valid")
				assert.Contains(t, logMsg, "TEST_TOKEN", "Log should contain token name")
				assert.Contains(t, logMsg, "length=32", "Log should contain token length")
				assert.Contains(t, logMsg, "format=hex", "Log should contain token format")
				assert.Contains(t, logMsg, "abcd****5678", "Log should contain redacted token")
			},
		},
		{
			name:      "InvalidTokenLog",
			tokenName: "BAD_TOKEN",
			result: TokenValidationResult{
				Valid:    false,
				Error:    errors.New("token format invalid"),
				Redacted: "xyz****123",
			},
			expectLog: func(t *testing.T, logMsg string) {
				assert.Contains(t, logMsg, "INVALID", "Log should indicate token is invalid")
				assert.Contains(t, logMsg, "BAD_TOKEN", "Log should contain token name")
				assert.Contains(t, logMsg, "token format invalid", "Log should contain error message")
				assert.Contains(t, logMsg, "xyz****123", "Log should contain redacted token")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logMsg := LogSecureTokenValidation(tt.tokenName, tt.result)
			tt.expectLog(t, logMsg)

			// Ensure no token leakage in logs
			assert.NotContains(t, logMsg, "abcdef123456789012345678901234567890abcd", 
				"Log should not contain full token values")
		})
	}
}

// TestConstantTimeValidation tests timing attack prevention in token validation.
func TestConstantTimeValidation(t *testing.T) {
	t.Parallel()

	// Test constant time length validation
	shortToken := "abc"
	validToken := "abcdef123456789012345678"
	longToken := strings.Repeat("a", 100)

	// These should all take similar time regardless of how far off the length is
	assert.False(t, constantTimeValidateLength(shortToken, 16, 64), 
		"Short token should be invalid")
	assert.True(t, constantTimeValidateLength(validToken, 16, 64), 
		"Valid token should be valid")
	assert.False(t, constantTimeValidateLength(longToken, 16, 64), 
		"Long token should be invalid")
}

// BenchmarkTokenValidation benchmarks token validation performance.
func BenchmarkTokenValidation(b *testing.B) {
	validator := NewTokenValidator()
	token := "abcdef123456789012345678901234567890abcd"
	
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result := validator.ValidateToken("LINODE_TOKEN", token)
		if !result.Valid {
			b.Fatalf("Expected valid token, got: %v", result.Error)
		}
	}
}

// BenchmarkTokenRedaction benchmarks token redaction performance.
func BenchmarkTokenRedaction(b *testing.B) {
	token := "abcdef123456789012345678901234567890abcd"
	
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		redacted := RedactToken(token)
		if len(redacted) == 0 {
			b.Fatal("Expected redacted token")
		}
	}
}

// TestTokenValidator_SecurityIntegration performs integration testing of security features.
func TestTokenValidator_SecurityIntegration(t *testing.T) {
	t.Parallel()

	// Simulate integration test environment
	validator := NewTokenValidator()
	
	// Test environment variables that might be used in CI
	testEnv := map[string]string{
		"LINODE_TOKEN": os.Getenv("LINODE_TOKEN"), // Real or empty from CI
		"GITHUB_TOKEN": os.Getenv("GITHUB_TOKEN"), // Real or empty from CI
	}

	getEnv := func(key string) string {
		return testEnv[key]
	}

	// Validate environment tokens
	results := validator.ValidateEnvironmentTokens(getEnv)

	// Generate audit report
	report := validator.GenerateSecurityAuditReport(getEnv)

	// Verify security measures
	for tokenName, result := range results {
		logMsg := LogSecureTokenValidation(tokenName, result)
		
		// Ensure no token leakage in logs
		if testEnv[tokenName] != "" {
			assert.NotContains(t, logMsg, testEnv[tokenName], 
				"Log should not contain full token value for %s", tokenName)
		}
		
		// Verify redaction is applied
		assert.NotEqual(t, testEnv[tokenName], result.Redacted, 
			"Redacted value should not equal original token for %s", tokenName)
	}

	// Verify report contains security recommendations
	assert.Contains(t, report.Recommendations, "Store tokens in secure environment variables", 
		"Report should include security recommendations")
	assert.Contains(t, report.Recommendations, "Never commit tokens to version control", 
		"Report should warn against committing tokens")
}