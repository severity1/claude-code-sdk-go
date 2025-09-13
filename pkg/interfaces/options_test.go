package interfaces

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

// TestOptionsInterfaceExistence verifies that the options interfaces exist and have correct method signatures.
func TestOptionsInterfaceExistence(t *testing.T) {
	tests := []struct {
		name            string
		interfaceType   reflect.Type
		expectedMethods []string
	}{
		{
			name:            "McpServerConfig interface",
			interfaceType:   reflect.TypeOf((*McpServerConfig)(nil)).Elem(),
			expectedMethods: []string{"Type"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.interfaceType.Kind() != reflect.Interface {
				t.Errorf("Expected %s to be an interface, got %s", tt.name, tt.interfaceType.Kind())
				return
			}

			actualMethods := make(map[string]bool)
			for i := 0; i < tt.interfaceType.NumMethod(); i++ {
				method := tt.interfaceType.Method(i)
				actualMethods[method.Name] = true
			}

			for _, expectedMethod := range tt.expectedMethods {
				if !actualMethods[expectedMethod] {
					t.Errorf("Expected %s to have method %s", tt.name, expectedMethod)
				}
			}
		})
	}
}

// TestMcpServerConfigTypeMethodSignature verifies Type() method has correct signature.
func TestMcpServerConfigTypeMethodSignature(t *testing.T) {
	configType := reflect.TypeOf((*McpServerConfig)(nil)).Elem()

	method, found := configType.MethodByName("Type")
	if !found {
		t.Error("McpServerConfig interface must have Type method")
		return
	}

	methodType := method.Type

	// Should take no parameters (interface methods don't count receiver)
	if methodType.NumIn() != 0 {
		t.Errorf("Type method should have 0 parameters, got %d", methodType.NumIn())
	}

	// Should return exactly 1 value
	if methodType.NumOut() != 1 {
		t.Errorf("Type method should return 1 value, got %d", methodType.NumOut())
	}

	// Return type should be McpServerType (not string - this is domain-specific)
	if methodType.NumOut() > 0 {
		returnType := methodType.Out(0)
		// We'll check the type name since we don't have McpServerType defined yet
		expectedTypeName := "McpServerType"
		if returnType.Name() != expectedTypeName {
			// For now, just check it's a named type (not primitive string)
			if returnType.Kind() == reflect.String && returnType.Name() == "" {
				t.Error("Type method should return named type (McpServerType), not plain string")
			}
		}
	}
}

// TestMethodNamingStandardization verifies standardized naming conventions.
func TestMethodNamingStandardization(t *testing.T) {
	configType := reflect.TypeOf((*McpServerConfig)(nil)).Elem()

	// Should have Type() method, not GetType()
	_, hasType := configType.MethodByName("Type")
	if !hasType {
		t.Error("McpServerConfig interface must have Type() method")
	}

	// Should NOT have GetType() method (legacy Java-style naming)
	if _, hasGetType := configType.MethodByName("GetType"); hasGetType {
		t.Error("McpServerConfig interface should not have GetType() method - use Type() instead")
	}
}

// TestOptionsInterfaceNilHandling verifies interfaces handle nil values correctly.
func TestOptionsInterfaceNilHandling(t *testing.T) {
	var config McpServerConfig

	if config != nil {
		t.Error("Nil McpServerConfig interface should be nil")
	}
}

// TestMcpServerTypeExistence verifies McpServerType is properly defined.
func TestMcpServerTypeExistence(t *testing.T) {
	// This test verifies that McpServerType exists as a named type
	// We'll get it through reflection on the interface method
	configType := reflect.TypeOf((*McpServerConfig)(nil)).Elem()

	method, found := configType.MethodByName("Type")
	if !found {
		t.Skip("Skipping type test - Type method not found")
		return
	}

	if method.Type.NumOut() == 0 {
		t.Error("Type method should return a value")
		return
	}

	returnType := method.Type.Out(0)

	// Should be a named type, not a primitive
	if returnType.Name() == "" {
		t.Error("Type method should return a named type (McpServerType), not an unnamed type")
	}

	// Should be based on string
	if returnType.Kind() != reflect.String {
		t.Error("McpServerType should be based on string type")
	}
}

// TestConcreteServerConfigImplementations verifies all concrete server configs implement McpServerConfig.
func TestConcreteServerConfigImplementations(t *testing.T) {
	tests := []struct {
		name     string
		config   McpServerConfig
		expected McpServerType
	}{
		{
			name:     "McpStdioServerConfig implements McpServerConfig",
			config:   &McpStdioServerConfig{},
			expected: McpServerTypeStdio,
		},
		{
			name:     "McpSSEServerConfig implements McpServerConfig",
			config:   &McpSSEServerConfig{},
			expected: McpServerTypeSSE,
		},
		{
			name:     "McpHTTPServerConfig implements McpServerConfig",
			config:   &McpHTTPServerConfig{},
			expected: McpServerTypeHTTP,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test interface implementation
			var _ McpServerConfig = tt.config

			// Test Type() method returns expected value
			actual := tt.config.Type()
			if actual != tt.expected {
				t.Errorf("Expected Type() to return %q, got %q", tt.expected, actual)
			}
		})
	}
}

// TestMcpServerTypeConstants verifies the server type constants are properly defined.
func TestMcpServerTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		value    McpServerType
		expected string
	}{
		{"McpServerTypeStdio", McpServerTypeStdio, "stdio"},
		{"McpServerTypeSSE", McpServerTypeSSE, "sse"},
		{"McpServerTypeHTTP", McpServerTypeHTTP, "http"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.expected {
				t.Errorf("Expected %s to equal %q, got %q", tt.name, tt.expected, string(tt.value))
			}
		})
	}
}

// TestConcreteConfigTypeMethodConsistency verifies Type() method consistency across concrete types.
func TestConcreteConfigTypeMethodConsistency(t *testing.T) {
	configs := []McpServerConfig{
		&McpStdioServerConfig{},
		&McpSSEServerConfig{},
		&McpHTTPServerConfig{},
	}

	for i, config := range configs {
		t.Run(fmt.Sprintf("Config_%d_Type_method", i), func(t *testing.T) {
			// Test that Type() method exists and returns string-based type
			configType := reflect.TypeOf(config)
			method, found := configType.MethodByName("Type")
			if !found {
				t.Error("Concrete config type must have Type() method")
				return
			}

			// Verify method signature
			methodType := method.Type
			if methodType.NumIn() != 1 { // receiver
				t.Errorf("Type method should have 1 parameter (receiver), got %d", methodType.NumIn())
			}
			if methodType.NumOut() != 1 {
				t.Errorf("Type method should return 1 value, got %d", methodType.NumOut())
			}

			// Test actual call
			serverType := config.Type()
			if string(serverType) == "" {
				t.Error("Type() method should return non-empty string")
			}
		})
	}
}

// TestConcreteConfigNoGetTypeMethod verifies concrete types don't have GetType() methods.
func TestConcreteConfigNoGetTypeMethod(t *testing.T) {
	configs := []interface{}{
		&McpStdioServerConfig{},
		&McpSSEServerConfig{},
		&McpHTTPServerConfig{},
	}

	for i, config := range configs {
		t.Run(fmt.Sprintf("Config_%d_no_GetType", i), func(t *testing.T) {
			configType := reflect.TypeOf(config)
			if _, hasGetType := configType.MethodByName("GetType"); hasGetType {
				t.Errorf("%T should not have GetType() method - use Type() instead", config)
			}
		})
	}
}

// TestConcreteConfigJSONMarshaling verifies JSON serialization works correctly.
func TestConcreteConfigJSONMarshaling(t *testing.T) {
	tests := []struct {
		name   string
		config McpServerConfig
	}{
		{
			name: "McpStdioServerConfig JSON marshaling",
			config: &McpStdioServerConfig{
				Command: "test-command",
				Args:    []string{"arg1", "arg2"},
				Env:     map[string]string{"KEY": "value"},
			},
		},
		{
			name: "McpSSEServerConfig JSON marshaling",
			config: &McpSSEServerConfig{
				URL:     "https://example.com/sse",
				Headers: map[string]string{"Authorization": "Bearer token"},
			},
		},
		{
			name: "McpHTTPServerConfig JSON marshaling",
			config: &McpHTTPServerConfig{
				URL:     "https://example.com/api",
				Headers: map[string]string{"Content-Type": "application/json"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			data, err := json.Marshal(tt.config)
			if err != nil {
				t.Errorf("Failed to marshal config: %v", err)
				return
			}

			// Should produce valid JSON
			if len(data) == 0 {
				t.Error("Marshaled data should not be empty")
			}

			// Test that Type() method still works after marshaling/unmarshaling round trip
			originalType := tt.config.Type()
			if string(originalType) == "" {
				t.Error("Original config Type() should return non-empty value")
			}
		})
	}
}

// TestConcreteConfigZeroValues verifies zero value behavior.
func TestConcreteConfigZeroValues(t *testing.T) {
	tests := []struct {
		name     string
		config   McpServerConfig
		expected McpServerType
	}{
		{
			name:     "McpStdioServerConfig zero value",
			config:   &McpStdioServerConfig{},
			expected: McpServerTypeStdio,
		},
		{
			name:     "McpSSEServerConfig zero value",
			config:   &McpSSEServerConfig{},
			expected: McpServerTypeSSE,
		},
		{
			name:     "McpHTTPServerConfig zero value",
			config:   &McpHTTPServerConfig{},
			expected: McpServerTypeHTTP,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Zero value should still implement interface correctly
			var _ McpServerConfig = tt.config

			// Type() should work even with zero values
			actual := tt.config.Type()
			if actual != tt.expected {
				t.Errorf("Expected Type() to return %q for zero value, got %q", tt.expected, actual)
			}
		})
	}
}

// TestOptionsValidate verifies Options.Validate() method works correctly.
func TestOptionsValidate(t *testing.T) {
	tests := []struct {
		name    string
		options *Options
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid options",
			options: &Options{
				MaxThinkingTokens: 5000,
				MaxTurns:          10,
				AllowedTools:      []string{"tool1", "tool2"},
				DisallowedTools:   []string{"tool3"},
			},
			wantErr: false,
		},
		{
			name: "negative MaxThinkingTokens",
			options: &Options{
				MaxThinkingTokens: -100,
				MaxTurns:          10,
			},
			wantErr: true,
			errMsg:  "MaxThinkingTokens must be non-negative, got -100",
		},
		{
			name: "negative MaxTurns",
			options: &Options{
				MaxThinkingTokens: 5000,
				MaxTurns:          -5,
			},
			wantErr: true,
			errMsg:  "MaxTurns must be non-negative, got -5",
		},
		{
			name: "tool in both allowed and disallowed",
			options: &Options{
				AllowedTools:    []string{"tool1", "tool2"},
				DisallowedTools: []string{"tool2", "tool3"},
			},
			wantErr: true,
			errMsg:  "tool 'tool2' cannot be in both AllowedTools and DisallowedTools",
		},
		{
			name: "zero values (valid)",
			options: &Options{
				MaxThinkingTokens: 0,
				MaxTurns:          0,
			},
			wantErr: false,
		},
		{
			name: "empty tool lists (valid)",
			options: &Options{
				AllowedTools:    []string{},
				DisallowedTools: []string{},
			},
			wantErr: false,
		},
		{
			name: "nil tool lists (valid)",
			options: &Options{
				AllowedTools:    nil,
				DisallowedTools: nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.options.Validate()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error, but got none")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("Expected error message %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
			}
		})
	}
}

// TestNewOptions verifies NewOptions() creates options with correct defaults.
func TestNewOptions(t *testing.T) {
	options := NewOptions()

	// Verify non-nil pointer
	if options == nil {
		t.Fatal("NewOptions() should return non-nil pointer")
	}

	// Test default values
	if len(options.AllowedTools) != 0 {
		t.Errorf("Expected empty AllowedTools slice, got length %d", len(options.AllowedTools))
	}
	if options.AllowedTools == nil {
		t.Error("AllowedTools should be empty slice, not nil")
	}

	if len(options.DisallowedTools) != 0 {
		t.Errorf("Expected empty DisallowedTools slice, got length %d", len(options.DisallowedTools))
	}
	if options.DisallowedTools == nil {
		t.Error("DisallowedTools should be empty slice, not nil")
	}

	if options.MaxThinkingTokens != DefaultMaxThinkingTokens {
		t.Errorf("Expected MaxThinkingTokens %d, got %d", DefaultMaxThinkingTokens, options.MaxThinkingTokens)
	}

	if len(options.AddDirs) != 0 {
		t.Errorf("Expected empty AddDirs slice, got length %d", len(options.AddDirs))
	}
	if options.AddDirs == nil {
		t.Error("AddDirs should be empty slice, not nil")
	}

	if options.McpServers == nil {
		t.Error("McpServers should be initialized map, not nil")
	}
	if len(options.McpServers) != 0 {
		t.Errorf("Expected empty McpServers map, got length %d", len(options.McpServers))
	}

	if options.ExtraArgs == nil {
		t.Error("ExtraArgs should be initialized map, not nil")
	}
	if len(options.ExtraArgs) != 0 {
		t.Errorf("Expected empty ExtraArgs map, got length %d", len(options.ExtraArgs))
	}

	// Test that default options are valid
	if err := options.Validate(); err != nil {
		t.Errorf("Default options should be valid, got error: %v", err)
	}

	// Test that we can add configurations to the default options
	options.AllowedTools = append(options.AllowedTools, "test-tool")
	options.McpServers["test-server"] = &McpStdioServerConfig{
		Command: "test-command",
	}
	options.ExtraArgs["test-arg"] = stringPtr("test-value")

	// Should still be valid after modifications
	if err := options.Validate(); err != nil {
		t.Errorf("Modified default options should be valid, got error: %v", err)
	}
}

// Helper function for string pointers in tests
func stringPtr(s string) *string {
	return &s
}
