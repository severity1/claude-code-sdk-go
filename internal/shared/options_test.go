package shared

import (
	"testing"
)

// TestOptionsDefaults tests Options struct default values using table-driven approach
func TestOptionsDefaults(t *testing.T) {
	options := NewOptions()

	tests := []struct {
		name     string
		field    interface{}
		expected interface{}
	}{
		{"MaxThinkingTokens", options.MaxThinkingTokens, 8000},
		{"ContinueConversation", options.ContinueConversation, false},
		{"MaxTurns", options.MaxTurns, 0},
		{"AllowedTools_initialized", options.AllowedTools == nil, false},
		{"AllowedTools_empty", len(options.AllowedTools), 0},
		{"DisallowedTools_initialized", options.DisallowedTools == nil, false},
		{"DisallowedTools_empty", len(options.DisallowedTools), 0},
		{"Betas_initialized", options.Betas == nil, false},
		{"Betas_empty", len(options.Betas), 0},
		{"AddDirs_initialized", options.AddDirs == nil, false},
		{"AddDirs_empty", len(options.AddDirs), 0},
		{"McpServers_initialized", options.McpServers == nil, false},
		{"McpServers_empty", len(options.McpServers), 0},
		{"ExtraArgs_initialized", options.ExtraArgs == nil, false},
		{"ExtraArgs_empty", len(options.ExtraArgs), 0},
		{"ExtraEnv_initialized", options.ExtraEnv == nil, false},
		{"ExtraEnv_empty", len(options.ExtraEnv), 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertOptionsField(t, test.field, test.expected, test.name)
		})
	}

	// Test nil pointer fields
	nilTests := []struct {
		name  string
		check func() bool
	}{
		{"SystemPrompt", func() bool { return options.SystemPrompt == nil }},
		{"AppendSystemPrompt", func() bool { return options.AppendSystemPrompt == nil }},
		{"Model", func() bool { return options.Model == nil }},
		{"PermissionMode", func() bool { return options.PermissionMode == nil }},
		{"PermissionPromptToolName", func() bool { return options.PermissionPromptToolName == nil }},
		{"Resume", func() bool { return options.Resume == nil }},
		{"Settings", func() bool { return options.Settings == nil }},
		{"Cwd", func() bool { return options.Cwd == nil }},
	}

	for _, test := range nilTests {
		t.Run(test.name+"_nil", func(t *testing.T) {
			if !test.check() {
				t.Errorf("Expected %s to be nil", test.name)
			}
		})
	}
}

// TestOptionsValidation tests critical validation edge cases
func TestOptionsValidation(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Options
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid_options",
			setup: func() *Options {
				opts := NewOptions()
				opts.AllowedTools = []string{"Read", "Write"}
				return opts
			},
			wantErr: false,
		},
		{
			name: "negative_thinking_tokens",
			setup: func() *Options {
				opts := NewOptions()
				opts.MaxThinkingTokens = -100
				return opts
			},
			wantErr: true,
			errMsg:  "MaxThinkingTokens must be non-negative, got -100",
		},
		{
			name: "conflicting_tools",
			setup: func() *Options {
				opts := NewOptions()
				opts.AllowedTools = []string{"Read", "Write"}
				opts.DisallowedTools = []string{"Write", "Bash"}
				return opts
			},
			wantErr: true,
			errMsg:  "tool 'Write' cannot be in both AllowedTools and DisallowedTools",
		},
		{
			name: "negative_max_turns",
			setup: func() *Options {
				opts := NewOptions()
				opts.MaxTurns = -5
				return opts
			},
			wantErr: true,
			errMsg:  "MaxTurns must be non-negative, got -5",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			options := test.setup()
			err := options.Validate()
			assertValidationError(t, err, test.wantErr, test.errMsg)
		})
	}
}

// TestMcpServerTypes tests MCP server configuration interface compliance
func TestMcpServerTypes(t *testing.T) {
	tests := []struct {
		name         string
		config       McpServerConfig
		expectedType McpServerType
	}{
		{
			name: "stdio_server",
			config: &McpStdioServerConfig{
				Type:    McpServerTypeStdio,
				Command: "node",
				Args:    []string{"server.js"},
			},
			expectedType: McpServerTypeStdio,
		},
		{
			name: "sse_server",
			config: &McpSSEServerConfig{
				Type: McpServerTypeSSE,
				URL:  "https://example.com/sse",
			},
			expectedType: McpServerTypeSSE,
		},
		{
			name: "http_server",
			config: &McpHTTPServerConfig{
				Type: McpServerTypeHTTP,
				URL:  "https://example.com/http",
			},
			expectedType: McpServerTypeHTTP,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertMcpServerType(t, test.config, test.expectedType)
		})
	}
}

// TestPermissionModeConstants tests permission mode constant values
func TestPermissionModeConstants(t *testing.T) {
	tests := []struct {
		mode     PermissionMode
		expected string
	}{
		{PermissionModeDefault, "default"},
		{PermissionModeAcceptEdits, "acceptEdits"},
		{PermissionModePlan, "plan"},
		{PermissionModeBypassPermissions, "bypassPermissions"},
	}

	for _, test := range tests {
		t.Run(string(test.mode), func(t *testing.T) {
			assertOptionsField(t, string(test.mode), test.expected, "PermissionMode")
		})
	}
}

// Helper functions

// assertOptionsField verifies field values with proper error reporting
func assertOptionsField(t *testing.T, actual, expected interface{}, fieldName string) {
	t.Helper()
	// Handle nil pointer comparisons properly
	if expected == nil {
		if actual != nil {
			t.Errorf("Expected %s = nil, got %v", fieldName, actual)
		}
		return
	}
	if actual != expected {
		t.Errorf("Expected %s = %v, got %v", fieldName, expected, actual)
	}
}

// assertValidationError verifies validation error behavior
func assertValidationError(t *testing.T, err error, wantErr bool, expectedMsg string) {
	t.Helper()
	if (err != nil) != wantErr {
		t.Errorf("error = %v, wantErr %v", err, wantErr)
		return
	}
	if wantErr && expectedMsg != "" && err.Error() != expectedMsg {
		t.Errorf("error = %v, expected message %q", err, expectedMsg)
	}
}

// assertMcpServerType verifies MCP server configuration types
func assertMcpServerType(t *testing.T, config McpServerConfig, expectedType McpServerType) {
	t.Helper()
	if config.GetType() != expectedType {
		t.Errorf("Expected server type %s, got %s", expectedType, config.GetType())
	}
}
