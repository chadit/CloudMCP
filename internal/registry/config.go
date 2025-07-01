package registry

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/chadit/CloudMCP/pkg/interfaces"
)

// Package-level errors.
var (
	ErrConfigurationDataNil = errors.New("configuration data is nil")
	ErrMissingRequiredKeys  = errors.New("missing required configuration keys")
)

// ConfigAdapter adapts environment-based configuration to the interfaces.Config interface.
// This allows providers to access configuration through a standardized interface
// regardless of the underlying configuration source.
type ConfigAdapter struct {
	data map[string]interface{}
}

// NewConfigAdapter creates a new configuration adapter from a map of values.
// The values can be strings, booleans, integers, or maps as supported by the Config interface.
func NewConfigAdapter(data map[string]interface{}) *ConfigAdapter {
	return &ConfigAdapter{
		data: data,
	}
}

// NewConfigFromEnvironment creates a configuration adapter from environment variables.
// It takes a prefix and scans for environment variables that start with that prefix.
func NewConfigFromEnvironment(envVars map[string]string, prefix string) *ConfigAdapter {
	data := make(map[string]interface{})

	for key, value := range envVars {
		if strings.HasPrefix(key, prefix) {
			// Remove prefix and convert to lowercase
			configKey := strings.ToLower(strings.TrimPrefix(key, prefix))
			configKey = strings.TrimPrefix(configKey, "_")

			// Try to parse as different types
			if parsedBool, err := strconv.ParseBool(value); err == nil {
				data[configKey] = parsedBool
			} else if parsedInt, err := strconv.Atoi(value); err == nil {
				data[configKey] = parsedInt
			} else {
				data[configKey] = value
			}
		}
	}

	return &ConfigAdapter{data: data}
}

// GetString retrieves a string configuration value by key.
// Returns empty string if the key doesn't exist or the value is not a string.
func (c *ConfigAdapter) GetString(key string) string {
	if value, exists := c.data[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
		// Try to convert other types to string
		return fmt.Sprintf("%v", value)
	}

	return ""
}

// GetBool retrieves a boolean configuration value by key.
// Returns false if the key doesn't exist or the value is not a boolean.
func (c *ConfigAdapter) GetBool(key string) bool {
	if value, exists := c.data[key]; exists {
		if boolean, ok := value.(bool); ok {
			return boolean
		}
		// Try to parse string values as boolean
		if str, ok := value.(string); ok {
			if parsed, err := strconv.ParseBool(str); err == nil {
				return parsed
			}
		}
	}

	return false
}

// GetInt retrieves an integer configuration value by key.
// Returns 0 if the key doesn't exist or the value is not an integer.
func (c *ConfigAdapter) GetInt(key string) int {
	if value, exists := c.data[key]; exists {
		if integer, ok := value.(int); ok {
			return integer
		}
		// Try to parse string values as integer
		if str, ok := value.(string); ok {
			if parsed, err := strconv.Atoi(str); err == nil {
				return parsed
			}
		}
	}

	return 0
}

// GetStringMap retrieves a map of string values by key.
// Returns empty map if the key doesn't exist or the value is not a map.
func (c *ConfigAdapter) GetStringMap(key string) map[string]string {
	if value, exists := c.data[key]; exists {
		if stringMap, ok := value.(map[string]string); ok {
			return stringMap
		}
		// Try to convert map[string]interface{} to map[string]string
		if interfaceMap, ok := value.(map[string]interface{}); ok {
			result := make(map[string]string)
			for k, v := range interfaceMap {
				result[k] = fmt.Sprintf("%v", v)
			}

			return result
		}
	}

	return make(map[string]string)
}

// IsSet checks if a configuration key has been set.
func (c *ConfigAdapter) IsSet(key string) bool {
	_, exists := c.data[key]

	return exists
}

// Validate validates the entire configuration.
// This default implementation always returns nil, but can be overridden
// for specific validation logic.
func (c *ConfigAdapter) Validate() error {
	// Basic validation - ensure data is not nil
	if c.data == nil {
		return ErrConfigurationDataNil
	}

	return nil
}

// SetValue sets a configuration value.
// This method allows dynamic configuration updates.
func (c *ConfigAdapter) SetValue(key string, value interface{}) {
	if c.data == nil {
		c.data = make(map[string]interface{})
	}

	c.data[key] = value
}

// GetValue retrieves a raw configuration value by key.
// Returns nil if the key doesn't exist.
func (c *ConfigAdapter) GetValue(key string) interface{} {
	return c.data[key]
}

// GetAllKeys returns all configuration keys.
func (c *ConfigAdapter) GetAllKeys() []string {
	keys := make([]string, 0, len(c.data))
	for key := range c.data {
		keys = append(keys, key)
	}

	return keys
}

// Clone creates a deep copy of the configuration.
func (c *ConfigAdapter) Clone() interfaces.Config { //nolint:ireturn // Clone method should return interface
	clone := make(map[string]interface{})
	for key, value := range c.data {
		clone[key] = value
	}

	return &ConfigAdapter{data: clone}
}

// Merge merges another configuration into this one.
// Values from the other configuration will override existing values.
func (c *ConfigAdapter) Merge(other interfaces.Config) {
	if adapter, ok := other.(*ConfigAdapter); ok {
		for key, value := range adapter.data {
			c.data[key] = value
		}
	}
}

// ValidateRequiredKeys validates that all required keys are present.
func (c *ConfigAdapter) ValidateRequiredKeys(requiredKeys []string) error {
	var missing []string

	for _, key := range requiredKeys {
		if !c.IsSet(key) {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("%w: %s", ErrMissingRequiredKeys, strings.Join(missing, ", "))
	}

	return nil
}

// ToMap returns the configuration as a map.
// This is useful for debugging and logging.
func (c *ConfigAdapter) ToMap() map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range c.data {
		result[key] = value
	}

	return result
}
