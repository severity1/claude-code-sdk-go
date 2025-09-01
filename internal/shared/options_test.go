package shared

import (
	"testing"
)

// TestDefaultOptionsStruct tests the Options struct with default values
func TestDefaultOptionsStruct(t *testing.T) {
	// Test Options with default values to match Python SDK
	options := NewOptions()

	// Verify tool control defaults
	if options.AllowedTools == nil {
		t.Error("Expected AllowedTools to be initialized, got nil")
	}
	if len(options.AllowedTools) != 0 {
		t.Errorf("Expected AllowedTools = [], got %v", options.AllowedTools)
	}

	if options.DisallowedTools == nil {
		t.Error("Expected DisallowedTools to be initialized, got nil")
	}
	if len(options.DisallowedTools) != 0 {
		t.Errorf("Expected DisallowedTools = [], got %v", options.DisallowedTools)
	}

	// Verify system prompts and model defaults (should be nil pointers)
	if options.SystemPrompt != nil {
		t.Errorf("Expected SystemPrompt = nil, got %v", options.SystemPrompt)
	}
	if options.AppendSystemPrompt != nil {
		t.Errorf("Expected AppendSystemPrompt = nil, got %v", options.AppendSystemPrompt)
	}
	if options.Model != nil {
		t.Errorf("Expected Model = nil, got %v", options.Model)
	}

	// Verify max thinking tokens default matches Python SDK
	if options.MaxThinkingTokens != 8000 {
		t.Errorf("Expected MaxThinkingTokens = 8000, got %d", options.MaxThinkingTokens)
	}

	// Verify permission system defaults
	if options.PermissionMode != nil {
		t.Errorf("Expected PermissionMode = nil, got %v", options.PermissionMode)
	}
	if options.PermissionPromptToolName != nil {
		t.Errorf("Expected PermissionPromptToolName = nil, got %v", options.PermissionPromptToolName)
	}

	// Verify session and state management defaults
	if options.ContinueConversation != false {
		t.Errorf("Expected ContinueConversation = false, got %v", options.ContinueConversation)
	}
	if options.Resume != nil {
		t.Errorf("Expected Resume = nil, got %v", options.Resume)
	}
	if options.MaxTurns != 0 {
		t.Errorf("Expected MaxTurns = 0, got %d", options.MaxTurns)
	}
	if options.Settings != nil {
		t.Errorf("Expected Settings = nil, got %v", options.Settings)
	}

	// Verify file system and context defaults
	if options.Cwd != nil {
		t.Errorf("Expected Cwd = nil, got %v", options.Cwd)
	}
	if options.AddDirs == nil {
		t.Error("Expected AddDirs to be initialized, got nil")
	}
	if len(options.AddDirs) != 0 {
		t.Errorf("Expected AddDirs = [], got %v", options.AddDirs)
	}

	// Verify MCP integration defaults
	if options.McpServers == nil {
		t.Error("Expected McpServers to be initialized, got nil")
	}
	if len(options.McpServers) != 0 {
		t.Errorf("Expected McpServers = {}, got %v", options.McpServers)
	}

	// Verify extensibility defaults
	if options.ExtraArgs == nil {
		t.Error("Expected ExtraArgs to be initialized, got nil")
	}
	if len(options.ExtraArgs) != 0 {
		t.Errorf("Expected ExtraArgs = {}, got %v", options.ExtraArgs)
	}
}

// TestOptionsValidation tests options field validation and constraints
func TestOptionsValidation(t *testing.T) {
	// Test valid options should pass validation
	validOptions := NewOptions()
	validOptions.AllowedTools = []string{"Read", "Write"}
	validOptions.MaxThinkingTokens = 8000
	systemPrompt := "Valid prompt"
	validOptions.SystemPrompt = &systemPrompt

	if err := validOptions.Validate(); err != nil {
		t.Errorf("Expected valid options to pass validation, got error: %v", err)
	}

	// Test negative max thinking tokens should fail
	invalidTokensOptions := NewOptions()
	invalidTokensOptions.MaxThinkingTokens = -100
	if err := invalidTokensOptions.Validate(); err == nil {
		t.Error("Expected negative max thinking tokens to fail validation")
	} else if err.Error() != "MaxThinkingTokens must be non-negative, got -100" {
		t.Errorf("Expected specific error for negative tokens, got: %v", err)
	}

	// Test zero max thinking tokens should be valid
	zeroTokensOptions := NewOptions()
	zeroTokensOptions.MaxThinkingTokens = 0
	if err := zeroTokensOptions.Validate(); err != nil {
		t.Errorf("Expected zero max thinking tokens to be valid, got error: %v", err)
	}

	// Test empty system prompt should be valid (nil is also valid)
	emptyPromptOptions := NewOptions()
	emptyPrompt := ""
	emptyPromptOptions.SystemPrompt = &emptyPrompt
	if err := emptyPromptOptions.Validate(); err != nil {
		t.Errorf("Expected empty system prompt to be valid, got error: %v", err)
	}

	// Test conflicting tool lists (same tool in both allowed and disallowed)
	conflictingOptions := NewOptions()
	conflictingOptions.AllowedTools = []string{"Read", "Write", "Edit"}
	conflictingOptions.DisallowedTools = []string{"Write", "Bash"}
	if err := conflictingOptions.Validate(); err == nil {
		t.Error("Expected conflicting tool lists to fail validation")
	} else if err.Error() != "tool 'Write' cannot be in both AllowedTools and DisallowedTools" {
		t.Errorf("Expected specific error for conflicting tools, got: %v", err)
	}

	// Test negative MaxTurns should fail
	invalidTurnsOptions := NewOptions()
	invalidTurnsOptions.MaxTurns = -5
	if err := invalidTurnsOptions.Validate(); err == nil {
		t.Error("Expected negative MaxTurns to fail validation")
	} else if err.Error() != "MaxTurns must be non-negative, got -5" {
		t.Errorf("Expected specific error for negative turns, got: %v", err)
	}

	// Test zero MaxTurns should be valid
	zeroTurnsOptions := NewOptions()
	zeroTurnsOptions.MaxTurns = 0
	if err := zeroTurnsOptions.Validate(); err != nil {
		t.Errorf("Expected zero MaxTurns to be valid, got error: %v", err)
	}
}

// TestMcpServerConfiguration tests MCP server configuration types
func TestMcpServerConfiguration(t *testing.T) {
	// Test McpStdioServerConfig
	stdioConfig := &McpStdioServerConfig{
		Type:    McpServerTypeStdio,
		Command: "node",
		Args:    []string{"server.js"},
		Env:     map[string]string{"NODE_ENV": "production"},
	}

	if stdioConfig.GetType() != McpServerTypeStdio {
		t.Errorf("Expected stdio config type to be %s, got %s", McpServerTypeStdio, stdioConfig.GetType())
	}

	// Test McpSSEServerConfig
	sseConfig := &McpSSEServerConfig{
		Type: McpServerTypeSSE,
		URL:  "https://example.com/sse",
		Headers: map[string]string{
			"Authorization": "Bearer token",
		},
	}

	if sseConfig.GetType() != McpServerTypeSSE {
		t.Errorf("Expected SSE config type to be %s, got %s", McpServerTypeSSE, sseConfig.GetType())
	}

	// Test McpHTTPServerConfig
	httpConfig := &McpHTTPServerConfig{
		Type: McpServerTypeHTTP,
		URL:  "https://example.com/http",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	if httpConfig.GetType() != McpServerTypeHTTP {
		t.Errorf("Expected HTTP config type to be %s, got %s", McpServerTypeHTTP, httpConfig.GetType())
	}

	// Test that all configs implement McpServerConfig interface
	var configs []McpServerConfig = []McpServerConfig{
		stdioConfig,
		sseConfig,
		httpConfig,
	}

	expectedTypes := []McpServerType{
		McpServerTypeStdio,
		McpServerTypeSSE,
		McpServerTypeHTTP,
	}

	for i, config := range configs {
		if config.GetType() != expectedTypes[i] {
			t.Errorf("Config %d: expected type %s, got %s", i, expectedTypes[i], config.GetType())
		}
	}
}

// TestPermissionModeConstants tests permission mode constants
func TestPermissionModeConstants(t *testing.T) {
	// Test that all permission modes have expected values
	expectedModes := map[PermissionMode]string{
		PermissionModeDefault:           "default",
		PermissionModeAcceptEdits:       "acceptEdits",
		PermissionModePlan:              "plan",
		PermissionModeBypassPermissions: "bypassPermissions",
	}

	for mode, expectedValue := range expectedModes {
		if string(mode) != expectedValue {
			t.Errorf("Expected permission mode %v to have value %s, got %s", mode, expectedValue, string(mode))
		}
	}
}
