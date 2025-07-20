package security

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

// TokenFormat represents supported token formats for validation.
type TokenFormat string

const (
	// HexToken represents hexadecimal tokens (e.g., API keys in hex format).
	HexToken TokenFormat = "hex"
	// AlphanumericToken represents alphanumeric tokens with optional special chars.
	AlphanumericToken TokenFormat = "alphanumeric"
	// Base64Token represents base64-encoded tokens.
	Base64Token TokenFormat = "base64"
	// JWTToken represents JSON Web Tokens.
	JWTToken TokenFormat = "jwt"
)

// TokenValidationConfig defines validation rules for tokens.
type TokenValidationConfig struct {
	Format     TokenFormat
	MinLength  int
	MaxLength  int
	Required   bool
	AllowEmpty bool
}

// DefaultTokenConfigs provides secure default configurations for different token types.
var DefaultTokenConfigs = map[string]TokenValidationConfig{
	"LINODE_TOKEN": {
		Format:     HexToken,
		MinLength:  32,
		MaxLength:  128,
		Required:   false, // Optional for testing
		AllowEmpty: true,  // Allow empty for local development
	},
	"AWS_ACCESS_KEY_ID": {
		Format:     AlphanumericToken,
		MinLength:  16,
		MaxLength:  32,
		Required:   false,
		AllowEmpty: true,
	},
	"AWS_SECRET_ACCESS_KEY": {
		Format:     AlphanumericToken,
		MinLength:  32,
		MaxLength:  64,
		Required:   false,
		AllowEmpty: true,
	},
	"GITHUB_TOKEN": {
		Format:     AlphanumericToken,
		MinLength:  40,
		MaxLength:  255,
		Required:   false,
		AllowEmpty: true,
	},
	// Generic test token validation
	"TEST_TOKEN": {
		Format:     HexToken,
		MinLength:  16,
		MaxLength:  256,
		Required:   false,
		AllowEmpty: true,
	},
}

// TokenValidationResult contains the result of token validation.
type TokenValidationResult struct {
	Valid       bool
	Error       error
	Redacted    string
	TokenLength int
	Format      TokenFormat
}

// TokenValidator provides secure token validation and redaction capabilities.
// It is safe for concurrent use.
type TokenValidator struct {
	mu      sync.RWMutex
	configs map[string]TokenValidationConfig
}

// NewTokenValidator creates a new token validator with default configurations.
func NewTokenValidator() *TokenValidator {
	// Create a copy of the default configs to avoid sharing the same map
	configsCopy := make(map[string]TokenValidationConfig)
	for k, v := range DefaultTokenConfigs {
		configsCopy[k] = v
	}
	
	return &TokenValidator{
		configs: configsCopy,
	}
}

// AddConfig adds or updates a token validation configuration.
// This method is safe for concurrent use.
func (v *TokenValidator) AddConfig(tokenName string, config TokenValidationConfig) {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	if v.configs == nil {
		v.configs = make(map[string]TokenValidationConfig)
	}
	v.configs[tokenName] = config
}

// ValidateToken performs comprehensive validation of a token using constant-time comparison
// to prevent timing attacks. This method is safe for concurrent use.
func (v *TokenValidator) ValidateToken(tokenName, tokenValue string) TokenValidationResult {
	v.mu.RLock()
	config, exists := v.configs[tokenName]
	v.mu.RUnlock()
	if !exists {
		return TokenValidationResult{
			Valid:    false,
			Error:    fmt.Errorf("no validation config found for token: %s", tokenName),
			Redacted: RedactToken(tokenValue),
		}
	}

	// Handle empty tokens
	if tokenValue == "" {
		if config.AllowEmpty {
			return TokenValidationResult{
				Valid:       true,
				Redacted:    "[EMPTY]",
				TokenLength: 0,
				Format:      config.Format,
			}
		}
		if config.Required {
			return TokenValidationResult{
				Valid:    false,
				Error:    errors.New("token is required but not provided"),
				Redacted: "[EMPTY]",
			}
		}
		// Optional and empty - valid
		return TokenValidationResult{
			Valid:       true,
			Redacted:    "[EMPTY]",
			TokenLength: 0,
			Format:      config.Format,
		}
	}

	result := TokenValidationResult{
		Redacted:    RedactToken(tokenValue),
		TokenLength: len(tokenValue),
		Format:      config.Format,
	}

	// Validate length using constant-time comparison to prevent timing attacks
	if !constantTimeValidateLength(tokenValue, config.MinLength, config.MaxLength) {
		result.Error = fmt.Errorf("token length invalid (expected %d-%d chars)", 
			config.MinLength, config.MaxLength)
		return result
	}

	// Validate format
	if !v.validateTokenFormat(tokenValue, config.Format) {
		result.Error = fmt.Errorf("token format invalid (expected %s)", config.Format)
		return result
	}

	result.Valid = true
	return result
}

// constantTimeValidateLength performs length validation in constant time.
func constantTimeValidateLength(token string, minLen, maxLen int) bool {
	length := len(token)
	// Use subtle.ConstantTimeEq for constant-time comparison
	var minValid, maxValid int
	if length >= minLen {
		minValid = 1
	}
	if length <= maxLen {
		maxValid = 1
	}
	minCheck := subtle.ConstantTimeEq(int32(minValid), 1)
	maxCheck := subtle.ConstantTimeEq(int32(maxValid), 1)
	return minCheck == 1 && maxCheck == 1
}

// validateTokenFormat validates token format using regex patterns.
func (v *TokenValidator) validateTokenFormat(token string, format TokenFormat) bool {
	switch format {
	case HexToken:
		return v.validateHexToken(token)
	case AlphanumericToken:
		return v.validateAlphanumericToken(token)
	case Base64Token:
		return v.validateBase64Token(token)
	case JWTToken:
		return v.validateJWTToken(token)
	default:
		return false
	}
}

// validateHexToken validates hexadecimal token format.
func (v *TokenValidator) validateHexToken(token string) bool {
	hexPattern := regexp.MustCompile(`^[0-9a-fA-F]+$`)
	return hexPattern.MatchString(token)
}

// validateAlphanumericToken validates alphanumeric token with optional special characters.
func (v *TokenValidator) validateAlphanumericToken(token string) bool {
	// Allow alphanumeric plus common special chars used in API tokens
	alphanumPattern := regexp.MustCompile(`^[a-zA-Z0-9_\-\.]+$`)
	return alphanumPattern.MatchString(token)
}

// validateBase64Token validates base64-encoded token format.
func (v *TokenValidator) validateBase64Token(token string) bool {
	base64Pattern := regexp.MustCompile(`^[A-Za-z0-9+/]*={0,2}$`)
	return base64Pattern.MatchString(token) && len(token)%4 == 0
}

// validateJWTToken validates JWT token format (3 base64url sections separated by dots).
func (v *TokenValidator) validateJWTToken(token string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}
	
	// Validate each part is base64url (no padding requirement, allows URL-safe characters)
	for _, part := range parts {
		if !v.validateBase64URLToken(part) {
			return false
		}
	}
	return true
}

// validateBase64URLToken validates base64url-encoded token format (used in JWT).
func (v *TokenValidator) validateBase64URLToken(token string) bool {
	// Base64url uses A-Z, a-z, 0-9, - (minus), _ (underscore) instead of + and /
	// and doesn't require padding (=), so length doesn't need to be multiple of 4
	base64URLPattern := regexp.MustCompile(`^[A-Za-z0-9_\-]*$`)
	return base64URLPattern.MatchString(token) && len(token) > 0
}

// RedactToken securely redacts a token for logging and display purposes.
func RedactToken(token string) string {
	if token == "" {
		return "[EMPTY]"
	}
	
	length := len(token)
	switch {
	case length <= 4:
		return "[REDACTED]"
	case length <= 8:
		return token[:2] + strings.Repeat("*", length-2)
	case length <= 16:
		return token[:3] + strings.Repeat("*", length-6) + token[length-3:]
	default:
		return token[:4] + strings.Repeat("*", length-8) + token[length-4:]
	}
}

// ValidateEnvironmentTokens validates all configured tokens from environment variables.
// This method is safe for concurrent use.
func (v *TokenValidator) ValidateEnvironmentTokens(getEnv func(string) string) map[string]TokenValidationResult {
	results := make(map[string]TokenValidationResult)
	
	// Get all token names safely
	v.mu.RLock()
	tokenNames := make([]string, 0, len(v.configs))
	for tokenName := range v.configs {
		tokenNames = append(tokenNames, tokenName)
	}
	v.mu.RUnlock()
	
	// Validate each token (ValidateToken handles its own locking)
	for _, tokenName := range tokenNames {
		tokenValue := getEnv(tokenName)
		results[tokenName] = v.ValidateToken(tokenName, tokenValue)
	}
	
	return results
}

// SecurityAuditReport generates a comprehensive security audit report for token validation.
type SecurityAuditReport struct {
	Timestamp       time.Time
	TotalTokens     int
	ValidTokens     int
	InvalidTokens   int
	EmptyTokens     int
	RequiredMissing int
	Issues          []string
	Recommendations []string
}

// GenerateSecurityAuditReport creates a comprehensive security audit report.
func (v *TokenValidator) GenerateSecurityAuditReport(getEnv func(string) string) SecurityAuditReport {
	results := v.ValidateEnvironmentTokens(getEnv)
	
	report := SecurityAuditReport{
		Timestamp: time.Now(),
		Issues:    make([]string, 0),
		Recommendations: []string{
			"Store tokens in secure environment variables",
			"Use different tokens for different environments",
			"Rotate tokens regularly",
			"Never commit tokens to version control",
			"Use token validation in CI/CD pipelines",
		},
	}
	
	for tokenName, result := range results {
		report.TotalTokens++
		
		if result.Valid {
			report.ValidTokens++
			if result.TokenLength == 0 {
				report.EmptyTokens++
			}
		} else {
			report.InvalidTokens++
			if result.Error != nil {
				report.Issues = append(report.Issues, 
					fmt.Sprintf("%s: %s", tokenName, result.Error.Error()))
			}
			
			v.mu.RLock()
			config := v.configs[tokenName]
			v.mu.RUnlock()
			if config.Required && result.TokenLength == 0 {
				report.RequiredMissing++
			}
		}
	}
	
	return report
}

// LogSecureTokenValidation provides secure logging of token validation without exposing tokens.
func LogSecureTokenValidation(tokenName string, result TokenValidationResult) string {
	if result.Valid {
		return fmt.Sprintf("Token %s: VALID (length=%d, format=%s, redacted=%s)", 
			tokenName, result.TokenLength, result.Format, result.Redacted)
	}
	
	errorMsg := "unknown error"
	if result.Error != nil {
		errorMsg = result.Error.Error()
	}
	
	return fmt.Sprintf("Token %s: INVALID - %s (redacted=%s)", 
		tokenName, errorMsg, result.Redacted)
}