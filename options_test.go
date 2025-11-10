package claudecode

import (
	"context"
	"testing"
)

// T015: Default Options Creation - Test functional options integration
func TestDefaultOptions(t *testing.T) {
	// Test that NewOptions() creates proper defaults via shared package
	options := NewOptions()

	// Verify that functional options work with shared types
	assertOptionsMaxThinkingTokens(t, options, 8000)

	// Test that we can apply functional options
	optionsWithPrompt := NewOptions(WithSystemPrompt("test prompt"))
	assertOptionsSystemPrompt(t, optionsWithPrompt, "test prompt")
}

// T016: Options with Tools
func TestOptionsWithTools(t *testing.T) {
	// Test Options with allowed_tools and disallowed_tools to match Python SDK
	options := NewOptions(
		WithAllowedTools("Read", "Write", "Edit"),
		WithDisallowedTools("Bash"),
	)

	// Verify allowed tools
	expectedAllowed := []string{"Read", "Write", "Edit"}
	assertOptionsStringSlice(t, options.AllowedTools, expectedAllowed, "AllowedTools")

	// Verify disallowed tools
	expectedDisallowed := []string{"Bash"}
	assertOptionsStringSlice(t, options.DisallowedTools, expectedDisallowed, "DisallowedTools")

	// Test with empty tools
	emptyOptions := NewOptions(
		WithAllowedTools(),
		WithDisallowedTools(),
	)
	assertOptionsStringSlice(t, emptyOptions.AllowedTools, []string{}, "AllowedTools")
	assertOptionsStringSlice(t, emptyOptions.DisallowedTools, []string{}, "DisallowedTools")
}

// T017: Permission Mode Options
func TestPermissionModeOptions(t *testing.T) {
	// Test all permission modes using table-driven approach
	tests := []struct {
		name string
		mode PermissionMode
	}{
		{"default", PermissionModeDefault},
		{"accept_edits", PermissionModeAcceptEdits},
		{"plan", PermissionModePlan},
		{"bypass_permissions", PermissionModeBypassPermissions},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			options := NewOptions(WithPermissionMode(test.mode))
			assertOptionsPermissionMode(t, options, test.mode)
		})
	}
}

// T018: System Prompt Options
func TestSystemPromptOptions(t *testing.T) {
	// Test system_prompt and append_system_prompt
	systemPrompt := "You are a helpful assistant."
	appendPrompt := "Be concise."

	options := NewOptions(
		WithSystemPrompt(systemPrompt),
		WithAppendSystemPrompt(appendPrompt),
	)

	// Verify system prompt is set
	assertOptionsSystemPrompt(t, options, systemPrompt)
	assertOptionsAppendSystemPrompt(t, options, appendPrompt)

	// Test with only system prompt
	systemOnlyOptions := NewOptions(WithSystemPrompt("Only system prompt"))
	assertOptionsSystemPrompt(t, systemOnlyOptions, "Only system prompt")
	assertOptionsAppendSystemPromptNil(t, systemOnlyOptions)

	// Test with only append prompt
	appendOnlyOptions := NewOptions(WithAppendSystemPrompt("Only append prompt"))
	assertOptionsAppendSystemPrompt(t, appendOnlyOptions, "Only append prompt")
	assertOptionsSystemPromptNil(t, appendOnlyOptions)
}

// T019: Session Continuation Options
func TestSessionContinuationOptions(t *testing.T) {
	// Test continue_conversation and resume options
	sessionID := "session-123"

	options := NewOptions(
		WithContinueConversation(true),
		WithResume(sessionID),
	)

	// Verify continue conversation is set
	assertOptionsContinueConversation(t, options, true)
	assertOptionsResume(t, options, sessionID)

	// Test with continue_conversation false
	falseOptions := NewOptions(WithContinueConversation(false))
	assertOptionsContinueConversation(t, falseOptions, false)
	assertOptionsResumeNil(t, falseOptions)

	// Test with only resume
	resumeOnlyOptions := NewOptions(WithResume("another-session"))
	assertOptionsResume(t, resumeOnlyOptions, "another-session")
	assertOptionsContinueConversation(t, resumeOnlyOptions, false) // default
}

// T020: Model Specification Options
func TestModelSpecificationOptions(t *testing.T) {
	// Test model and permission_prompt_tool_name
	model := "claude-3-5-sonnet-20241022"
	toolName := "CustomTool"

	options := NewOptions(
		WithModel(model),
		WithPermissionPromptToolName(toolName),
	)

	// Verify model and tool name are set
	assertOptionsModel(t, options, model)
	assertOptionsPermissionPromptToolName(t, options, toolName)

	// Test with only model
	modelOnlyOptions := NewOptions(WithModel("claude-opus-4"))
	assertOptionsModel(t, modelOnlyOptions, "claude-opus-4")
	assertOptionsPermissionPromptToolNameNil(t, modelOnlyOptions)

	// Test with only permission prompt tool name
	toolOnlyOptions := NewOptions(WithPermissionPromptToolName("OnlyTool"))
	assertOptionsPermissionPromptToolName(t, toolOnlyOptions, "OnlyTool")
	assertOptionsModelNil(t, toolOnlyOptions)
}

// T021: Functional Options Pattern
func TestFunctionalOptionsPattern(t *testing.T) {
	// Test chaining multiple functional options to create a fluent API
	options := NewOptions(
		WithSystemPrompt("You are a helpful assistant"),
		WithAllowedTools("Read", "Write"),
		WithDisallowedTools("Bash"),
		WithPermissionMode(PermissionModeAcceptEdits),
		WithModel("claude-3-5-sonnet-20241022"),
		WithContinueConversation(true),
		WithResume("session-456"),
		WithCwd("/tmp/test"),
		WithAddDirs("/tmp/dir1", "/tmp/dir2"),
		WithMaxThinkingTokens(10000),
		WithPermissionPromptToolName("CustomPermissionTool"),
	)

	// Verify all options are correctly applied
	if options.SystemPrompt == nil || *options.SystemPrompt != "You are a helpful assistant" {
		t.Errorf("Expected SystemPrompt = %q, got %v", "You are a helpful assistant", options.SystemPrompt)
	}

	expectedAllowed := []string{"Read", "Write"}
	if len(options.AllowedTools) != len(expectedAllowed) {
		t.Errorf("Expected AllowedTools length = %d, got %d", len(expectedAllowed), len(options.AllowedTools))
	}

	expectedDisallowed := []string{"Bash"}
	if len(options.DisallowedTools) != len(expectedDisallowed) {
		t.Errorf("Expected DisallowedTools length = %d, got %d", len(expectedDisallowed), len(options.DisallowedTools))
	}

	if options.PermissionMode == nil || *options.PermissionMode != PermissionModeAcceptEdits {
		t.Errorf("Expected PermissionMode = %q, got %v", PermissionModeAcceptEdits, options.PermissionMode)
	}

	if options.Model == nil || *options.Model != "claude-3-5-sonnet-20241022" {
		t.Errorf("Expected Model = %q, got %v", "claude-3-5-sonnet-20241022", options.Model)
	}

	if options.ContinueConversation != true {
		t.Errorf("Expected ContinueConversation = true, got %v", options.ContinueConversation)
	}

	if options.Resume == nil || *options.Resume != "session-456" {
		t.Errorf("Expected Resume = %q, got %v", "session-456", options.Resume)
	}

	if options.Cwd == nil || *options.Cwd != "/tmp/test" {
		t.Errorf("Expected Cwd = %q, got %v", "/tmp/test", options.Cwd)
	}

	expectedAddDirs := []string{"/tmp/dir1", "/tmp/dir2"}
	if len(options.AddDirs) != len(expectedAddDirs) {
		t.Errorf("Expected AddDirs length = %d, got %d", len(expectedAddDirs), len(options.AddDirs))
	}

	if options.MaxThinkingTokens != 10000 {
		t.Errorf("Expected MaxThinkingTokens = 10000, got %d", options.MaxThinkingTokens)
	}

	if options.PermissionPromptToolName == nil || *options.PermissionPromptToolName != "CustomPermissionTool" {
		t.Errorf("Expected PermissionPromptToolName = %q, got %v", "CustomPermissionTool", options.PermissionPromptToolName)
	}
}

// T022: MCP Server Configuration
func TestMcpServerConfiguration(t *testing.T) {
	// Test all three MCP server configuration types: stdio, SSE, HTTP

	// Create MCP server configurations
	stdioConfig := &McpStdioServerConfig{
		Type:    McpServerTypeStdio,
		Command: "python",
		Args:    []string{"-m", "my_mcp_server"},
		Env:     map[string]string{"DEBUG": "1"},
	}

	sseConfig := &McpSSEServerConfig{
		Type:    McpServerTypeSSE,
		URL:     "http://localhost:8080/sse",
		Headers: map[string]string{"Authorization": "Bearer token123"},
	}

	httpConfig := &McpHTTPServerConfig{
		Type:    McpServerTypeHTTP,
		URL:     "http://localhost:8080/mcp",
		Headers: map[string]string{"Content-Type": "application/json"},
	}

	servers := map[string]McpServerConfig{
		"stdio_server": stdioConfig,
		"sse_server":   sseConfig,
		"http_server":  httpConfig,
	}

	options := NewOptions(WithMcpServers(servers))

	// Verify MCP servers are set
	if options.McpServers == nil {
		t.Error("Expected McpServers to be set, got nil")
	}

	if len(options.McpServers) != 3 {
		t.Errorf("Expected 3 MCP servers, got %d", len(options.McpServers))
	}

	// Test stdio server configuration
	stdioServer, exists := options.McpServers["stdio_server"]
	if !exists {
		t.Error("Expected stdio_server to exist")
	}
	if stdioServer.GetType() != McpServerTypeStdio {
		t.Errorf("Expected stdio server type = %q, got %q", McpServerTypeStdio, stdioServer.GetType())
	}

	stdioTyped, ok := stdioServer.(*McpStdioServerConfig)
	if !ok {
		t.Errorf("Expected *McpStdioServerConfig, got %T", stdioServer)
	} else {
		if stdioTyped.Command != "python" {
			t.Errorf("Expected Command = %q, got %q", "python", stdioTyped.Command)
		}
		if len(stdioTyped.Args) != 2 || stdioTyped.Args[0] != "-m" {
			t.Errorf("Expected Args = [-m my_mcp_server], got %v", stdioTyped.Args)
		}
		if stdioTyped.Env["DEBUG"] != "1" {
			t.Errorf("Expected Env[DEBUG] = %q, got %q", "1", stdioTyped.Env["DEBUG"])
		}
	}

	// Test SSE server configuration
	sseServer, exists := options.McpServers["sse_server"]
	if !exists {
		t.Error("Expected sse_server to exist")
	}
	if sseServer.GetType() != McpServerTypeSSE {
		t.Errorf("Expected SSE server type = %q, got %q", McpServerTypeSSE, sseServer.GetType())
	}

	sseTyped, ok := sseServer.(*McpSSEServerConfig)
	if !ok {
		t.Errorf("Expected *McpSSEServerConfig, got %T", sseServer)
	} else {
		if sseTyped.URL != "http://localhost:8080/sse" {
			t.Errorf("Expected URL = %q, got %q", "http://localhost:8080/sse", sseTyped.URL)
		}
		if sseTyped.Headers["Authorization"] != "Bearer token123" {
			t.Errorf("Expected Headers[Authorization] = %q, got %q", "Bearer token123", sseTyped.Headers["Authorization"])
		}
	}

	// Test HTTP server configuration
	httpServer, exists := options.McpServers["http_server"]
	if !exists {
		t.Error("Expected http_server to exist")
	}
	if httpServer.GetType() != McpServerTypeHTTP {
		t.Errorf("Expected HTTP server type = %q, got %q", McpServerTypeHTTP, httpServer.GetType())
	}

	httpTyped, ok := httpServer.(*McpHTTPServerConfig)
	if !ok {
		t.Errorf("Expected *McpHTTPServerConfig, got %T", httpServer)
	} else {
		if httpTyped.URL != "http://localhost:8080/mcp" {
			t.Errorf("Expected URL = %q, got %q", "http://localhost:8080/mcp", httpTyped.URL)
		}
		if httpTyped.Headers["Content-Type"] != "application/json" {
			t.Errorf("Expected Headers[Content-Type] = %q, got %q", "application/json", httpTyped.Headers["Content-Type"])
		}
	}
}

// T023: Extra Args Support
func TestExtraArgsSupport(t *testing.T) {
	// Test arbitrary CLI flag support via ExtraArgs map[string]*string

	// Create extra args - nil values represent boolean flags, non-nil represent flags with values
	debugFlag := "verbose"
	extraArgs := map[string]*string{
		"--debug":   &debugFlag,        // Flag with value: --debug=verbose
		"--verbose": nil,               // Boolean flag: --verbose
		"--output":  stringPtr("json"), // Flag with value: --output=json
		"--quiet":   nil,               // Boolean flag: --quiet
	}

	options := NewOptions(WithExtraArgs(extraArgs))

	// Verify extra args are set
	if options.ExtraArgs == nil {
		t.Error("Expected ExtraArgs to be set, got nil")
	}

	if len(options.ExtraArgs) != 4 {
		t.Errorf("Expected 4 extra args, got %d", len(options.ExtraArgs))
	}

	// Test flag with value
	debugValue, exists := options.ExtraArgs["--debug"]
	if !exists {
		t.Error("Expected --debug flag to exist")
	}
	if debugValue == nil {
		t.Error("Expected --debug to have a value, got nil")
		return
	}
	if *debugValue != "verbose" {
		t.Errorf("Expected --debug = %q, got %q", "verbose", *debugValue)
	}

	// Test boolean flag
	verboseValue, exists := options.ExtraArgs["--verbose"]
	if !exists {
		t.Error("Expected --verbose flag to exist")
	}
	if verboseValue != nil {
		t.Errorf("Expected --verbose to be boolean flag (nil), got %v", verboseValue)
	}

	// Test another flag with value
	outputValue, exists := options.ExtraArgs["--output"]
	if !exists {
		t.Error("Expected --output flag to exist")
	}
	if outputValue == nil {
		t.Error("Expected --output to have a value, got nil")
		return
	}
	if *outputValue != "json" {
		t.Errorf("Expected --output = %q, got %q", "json", *outputValue)
	}

	// Test another boolean flag
	quietValue, exists := options.ExtraArgs["--quiet"]
	if !exists {
		t.Error("Expected --quiet flag to exist")
	}
	if quietValue != nil {
		t.Errorf("Expected --quiet to be boolean flag (nil), got %v", quietValue)
	}

	// Test empty extra args
	emptyOptions := NewOptions(WithExtraArgs(map[string]*string{}))
	if emptyOptions.ExtraArgs == nil {
		t.Error("Expected ExtraArgs to be initialized, got nil")
	}
	if len(emptyOptions.ExtraArgs) != 0 {
		t.Errorf("Expected empty ExtraArgs, got %v", emptyOptions.ExtraArgs)
	}
}

// T024: Options Validation
func TestOptionsValidationIntegration(t *testing.T) {
	// Test that validation works through functional options API (detailed tests in internal/shared)
	validOptions := NewOptions(
		WithAllowedTools("Read", "Write"),
		WithMaxThinkingTokens(8000),
		WithSystemPrompt("Valid prompt"),
	)
	assertOptionsValidationError(t, validOptions, false, "valid options should pass validation")

	// Test that functional options can create invalid options that validation catches
	invalidOptions := NewOptions(WithMaxThinkingTokens(-100))
	assertOptionsValidationError(t, invalidOptions, true, "negative max thinking tokens should fail validation")
}

// T025: NewOptions Constructor
func TestNewOptionsConstructor(t *testing.T) {
	// Test Options creation with functional options applied correctly with defaults

	// Test NewOptions with no arguments should return defaults
	defaultOptions := NewOptions()
	assertOptionsMaxThinkingTokens(t, defaultOptions, 8000)
	assertOptionsStringSlice(t, defaultOptions.AllowedTools, []string{}, "AllowedTools")

	// Test NewOptions with single functional option
	singleOptionOptions := NewOptions(WithSystemPrompt("Single option test"))
	assertOptionsSystemPrompt(t, singleOptionOptions, "Single option test")
	// Should still have defaults for other fields
	assertOptionsMaxThinkingTokens(t, singleOptionOptions, 8000)

	// Test NewOptions with multiple functional options applied in order
	multipleOptions := NewOptions(
		WithMaxThinkingTokens(5000),               // Override default
		WithAllowedTools("Read"),                  // Add tools
		WithSystemPrompt("First prompt"),          // Set system prompt
		WithMaxThinkingTokens(12000),              // Override again (should win)
		WithAllowedTools("Read", "Write", "Edit"), // Override tools (should win)
		WithSystemPrompt("Second prompt"),         // Override again (should win)
		WithDisallowedTools("Bash"),
		WithPermissionMode(PermissionModeAcceptEdits),
		WithContinueConversation(true),
		WithMaxTurns(5),                        // Test WithMaxTurns
		WithSettings("/path/to/settings.json"), // Test WithSettings
	)

	// Verify options are applied in order (later options override earlier ones)
	assertOptionsMaxThinkingTokens(t, multipleOptions, 12000) // final override
	assertOptionsStringSlice(t, multipleOptions.AllowedTools, []string{"Read", "Write", "Edit"}, "AllowedTools")
	assertOptionsSystemPrompt(t, multipleOptions, "Second prompt") // final override
	assertOptionsStringSlice(t, multipleOptions.DisallowedTools, []string{"Bash"}, "DisallowedTools")
	assertOptionsPermissionMode(t, multipleOptions, PermissionModeAcceptEdits)
	assertOptionsContinueConversation(t, multipleOptions, true)
	assertOptionsMaxTurns(t, multipleOptions, 5)
	assertOptionsSettings(t, multipleOptions, "/path/to/settings.json")

	// Test that unmodified fields retain defaults
	assertOptionsResumeNil(t, multipleOptions)
	assertOptionsCwdNil(t, multipleOptions)

	// Test that maps are properly initialized even with options
	if multipleOptions.McpServers == nil {
		t.Error("Expected McpServers to be initialized, got nil")
	} else {
		assertOptionsMapInitialized(t, len(multipleOptions.McpServers), "McpServers")
	}

	if multipleOptions.ExtraArgs == nil {
		t.Error("Expected ExtraArgs to be initialized, got nil")
	} else {
		assertOptionsMapInitialized(t, len(multipleOptions.ExtraArgs), "ExtraArgs")
	}
}

// TestWithCLIPath tests the WithCLIPath option function
func TestWithCLIPath(t *testing.T) {
	tests := []struct {
		name     string
		cliPath  string
		expected *string
	}{
		{
			name:     "valid_cli_path",
			cliPath:  "/usr/local/bin/claude",
			expected: stringPtr("/usr/local/bin/claude"),
		},
		{
			name:     "relative_cli_path",
			cliPath:  "./claude",
			expected: stringPtr("./claude"),
		},
		{
			name:     "empty_cli_path",
			cliPath:  "",
			expected: stringPtr(""),
		},
		{
			name:     "windows_cli_path",
			cliPath:  "C:\\Program Files\\Claude\\claude.exe",
			expected: stringPtr("C:\\Program Files\\Claude\\claude.exe"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			options := NewOptions(WithCLIPath(test.cliPath))

			if options.CLIPath == nil && test.expected != nil {
				t.Errorf("Expected CLIPath to be set to %q, got nil", *test.expected)
			}

			if options.CLIPath != nil && test.expected == nil {
				t.Errorf("Expected CLIPath to be nil, got %q", *options.CLIPath)
			}

			if options.CLIPath != nil && test.expected != nil && *options.CLIPath != *test.expected {
				t.Errorf("Expected CLIPath %q, got %q", *test.expected, *options.CLIPath)
			}
		})
	}

	// Test integration with other options
	t.Run("cli_path_with_other_options", func(t *testing.T) {
		options := NewOptions(
			WithCLIPath("/custom/claude"),
			WithSystemPrompt("Test system prompt"),
			WithModel("claude-sonnet-3-5-20241022"),
		)

		if options.CLIPath == nil || *options.CLIPath != "/custom/claude" {
			t.Errorf("Expected CLIPath to be preserved with other options")
		}

		assertOptionsSystemPrompt(t, options, "Test system prompt")
		assertOptionsModel(t, options, "claude-sonnet-3-5-20241022")
	})
}

// TestWithTransport tests the WithTransport option function
func TestWithTransport(t *testing.T) {
	// Create a mock transport for testing
	mockTransport := &mockTransportForOptions{}

	t.Run("transport_marker_in_extra_args", func(t *testing.T) {
		options := NewOptions(WithTransport(mockTransport))

		if options.ExtraArgs == nil {
			t.Fatal("Expected ExtraArgs to be initialized")
		}

		marker, exists := options.ExtraArgs["__transport_marker__"]
		if !exists {
			t.Error("Expected transport marker to be set in ExtraArgs")
		}

		if marker == nil || *marker != customTransportMarker {
			t.Errorf("Expected transport marker value 'custom_transport', got %v", marker)
		}
	})

	t.Run("transport_with_existing_extra_args", func(t *testing.T) {
		options := NewOptions(
			WithExtraArgs(map[string]*string{"existing": stringPtr("value")}),
			WithTransport(mockTransport),
		)

		if options.ExtraArgs == nil {
			t.Fatal("Expected ExtraArgs to be preserved")
		}

		// Check existing arg is preserved
		existing, exists := options.ExtraArgs["existing"]
		if !exists || existing == nil || *existing != "value" {
			t.Error("Expected existing ExtraArgs to be preserved")
		}

		// Check transport marker is added
		marker, exists := options.ExtraArgs["__transport_marker__"]
		if !exists || marker == nil || *marker != customTransportMarker {
			t.Error("Expected transport marker to be added to existing ExtraArgs")
		}
	})

	t.Run("transport_with_nil_extra_args", func(t *testing.T) {
		// Create options with nil ExtraArgs
		options := &Options{}

		// Apply WithTransport option
		WithTransport(mockTransport)(options)

		if options.ExtraArgs == nil {
			t.Error("Expected ExtraArgs to be initialized")
		}

		marker, exists := options.ExtraArgs["__transport_marker__"]
		if !exists || marker == nil || *marker != customTransportMarker {
			t.Error("Expected transport marker to be set when ExtraArgs was nil")
		}
	})

	t.Run("multiple_transport_calls", func(t *testing.T) {
		anotherMockTransport := &mockTransportForOptions{}

		options := NewOptions(
			WithTransport(mockTransport),
			WithTransport(anotherMockTransport), // Should overwrite
		)

		// Should only have one transport marker (last one wins)
		marker, exists := options.ExtraArgs["__transport_marker__"]
		if !exists || marker == nil || *marker != customTransportMarker {
			t.Error("Expected last transport to set the marker")
		}
	})
}

// Helper Functions - following client_test.go patterns

// assertOptionsMaxThinkingTokens verifies MaxThinkingTokens value
func assertOptionsMaxThinkingTokens(t *testing.T, options *Options, expected int) {
	t.Helper()
	if options.MaxThinkingTokens != expected {
		t.Errorf("Expected MaxThinkingTokens = %d, got %d", expected, options.MaxThinkingTokens)
	}
}

// assertOptionsSystemPrompt verifies SystemPrompt value
func assertOptionsSystemPrompt(t *testing.T, options *Options, expected string) {
	t.Helper()
	if options.SystemPrompt == nil {
		t.Error("Expected SystemPrompt to be set, got nil")
		return
	}
	actual := *options.SystemPrompt
	if actual != expected {
		t.Errorf("Expected SystemPrompt = %q, got %q", expected, actual)
	}
}

// assertOptionsSystemPromptNil verifies SystemPrompt is nil
func assertOptionsSystemPromptNil(t *testing.T, options *Options) {
	t.Helper()
	if options.SystemPrompt != nil {
		t.Errorf("Expected SystemPrompt = nil, got %v", *options.SystemPrompt)
	}
}

// assertOptionsAppendSystemPrompt verifies AppendSystemPrompt value
func assertOptionsAppendSystemPrompt(t *testing.T, options *Options, expected string) {
	t.Helper()
	if options.AppendSystemPrompt == nil {
		t.Error("Expected AppendSystemPrompt to be set, got nil")
		return
	}
	if *options.AppendSystemPrompt != expected {
		t.Errorf("Expected AppendSystemPrompt = %q, got %q", expected, *options.AppendSystemPrompt)
	}
}

// assertOptionsAppendSystemPromptNil verifies AppendSystemPrompt is nil
func assertOptionsAppendSystemPromptNil(t *testing.T, options *Options) {
	t.Helper()
	if options.AppendSystemPrompt != nil {
		t.Errorf("Expected AppendSystemPrompt = nil, got %v", *options.AppendSystemPrompt)
	}
}

// assertOptionsStringSlice verifies string slice values
func assertOptionsStringSlice(t *testing.T, actual, expected []string, fieldName string) {
	t.Helper()
	if len(actual) != len(expected) {
		t.Errorf("Expected %s length = %d, got %d", fieldName, len(expected), len(actual))
		return
	}
	for i, expectedVal := range expected {
		if i >= len(actual) || actual[i] != expectedVal {
			t.Errorf("Expected %s[%d] = %q, got %q", fieldName, i, expectedVal, actual[i])
		}
	}
}

// assertOptionsPermissionMode verifies PermissionMode value
func assertOptionsPermissionMode(t *testing.T, options *Options, expected PermissionMode) {
	t.Helper()
	if options.PermissionMode == nil {
		t.Error("Expected PermissionMode to be set, got nil")
		return
	}
	if *options.PermissionMode != expected {
		t.Errorf("Expected PermissionMode = %q, got %q", expected, *options.PermissionMode)
	}
}

// assertOptionsContinueConversation verifies ContinueConversation value
func assertOptionsContinueConversation(t *testing.T, options *Options, expected bool) {
	t.Helper()
	if options.ContinueConversation != expected {
		t.Errorf("Expected ContinueConversation = %v, got %v", expected, options.ContinueConversation)
	}
}

// assertOptionsResume verifies Resume value
func assertOptionsResume(t *testing.T, options *Options, expected string) {
	t.Helper()
	if options.Resume == nil {
		t.Error("Expected Resume to be set, got nil")
		return
	}
	if *options.Resume != expected {
		t.Errorf("Expected Resume = %q, got %q", expected, *options.Resume)
	}
}

// assertOptionsResumeNil verifies Resume is nil
func assertOptionsResumeNil(t *testing.T, options *Options) {
	t.Helper()
	if options.Resume != nil {
		t.Errorf("Expected Resume = nil, got %v", *options.Resume)
	}
}

// assertOptionsModel verifies Model value
func assertOptionsModel(t *testing.T, options *Options, expected string) {
	t.Helper()
	if options.Model == nil {
		t.Error("Expected Model to be set, got nil")
		return
	}
	if *options.Model != expected {
		t.Errorf("Expected Model = %q, got %q", expected, *options.Model)
	}
}

// assertOptionsModelNil verifies Model is nil
func assertOptionsModelNil(t *testing.T, options *Options) {
	t.Helper()
	if options.Model != nil {
		t.Errorf("Expected Model = nil, got %v", *options.Model)
	}
}

// assertOptionsPermissionPromptToolName verifies PermissionPromptToolName value
func assertOptionsPermissionPromptToolName(t *testing.T, options *Options, expected string) {
	t.Helper()
	if options.PermissionPromptToolName == nil {
		t.Error("Expected PermissionPromptToolName to be set, got nil")
		return
	}
	if *options.PermissionPromptToolName != expected {
		t.Errorf("Expected PermissionPromptToolName = %q, got %q", expected, *options.PermissionPromptToolName)
	}
}

// assertOptionsPermissionPromptToolNameNil verifies PermissionPromptToolName is nil
func assertOptionsPermissionPromptToolNameNil(t *testing.T, options *Options) {
	t.Helper()
	if options.PermissionPromptToolName != nil {
		t.Errorf("Expected PermissionPromptToolName = nil, got %v", *options.PermissionPromptToolName)
	}
}

// assertOptionsCwdNil verifies Cwd is nil
func assertOptionsCwdNil(t *testing.T, options *Options) {
	t.Helper()
	if options.Cwd != nil {
		t.Errorf("Expected Cwd = nil, got %v", *options.Cwd)
	}
}

// assertOptionsMaxTurns verifies MaxTurns value
func assertOptionsMaxTurns(t *testing.T, options *Options, expected int) {
	t.Helper()
	if options.MaxTurns != expected {
		t.Errorf("Expected MaxTurns = %d, got %d", expected, options.MaxTurns)
	}
}

// assertOptionsSettings verifies Settings value
func assertOptionsSettings(t *testing.T, options *Options, expected string) {
	t.Helper()
	if options.Settings == nil {
		t.Error("Expected Settings to be set, got nil")
		return
	}
	if *options.Settings != expected {
		t.Errorf("Expected Settings = %q, got %q", expected, *options.Settings)
	}
}

// assertOptionsMapInitialized verifies a map field is initialized but empty
func assertOptionsMapInitialized(t *testing.T, actualLen int, fieldName string) {
	t.Helper()
	if actualLen != 0 {
		t.Errorf("Expected %s = {} (empty but initialized), got length %d", fieldName, actualLen)
	}
}

// assertOptionsValidationError verifies validation returns error
func assertOptionsValidationError(t *testing.T, options *Options, shouldError bool, description string) {
	t.Helper()
	err := options.Validate()
	if shouldError && err == nil {
		t.Errorf("%s: expected validation error, got nil", description)
	}
	if !shouldError && err != nil {
		t.Errorf("%s: expected no validation error, got: %v", description, err)
	}
}

// Helper function for creating string pointers
func stringPtr(s string) *string {
	return &s
}

// mockTransportForOptions is a minimal mock transport for testing options
type mockTransportForOptions struct{}

func (m *mockTransportForOptions) Connect(_ context.Context) error { return nil }
func (m *mockTransportForOptions) SendMessage(_ context.Context, _ StreamMessage) error {
	return nil
}

func (m *mockTransportForOptions) ReceiveMessages(_ context.Context) (<-chan Message, <-chan error) {
	return nil, nil
}
func (m *mockTransportForOptions) Interrupt(_ context.Context) error { return nil }
func (m *mockTransportForOptions) Close() error                      { return nil }
func (m *mockTransportForOptions) GetValidator() *StreamValidator    { return &StreamValidator{} }

// TestWithEnvOptions tests environment variable functional options following table-driven pattern
func TestWithEnvOptions(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *Options
		expected  map[string]string
		wantPanic bool
	}{
		{
			name: "single_env_var",
			setup: func() *Options {
				return NewOptions(WithEnvVar("DEBUG", "1"))
			},
			expected: map[string]string{"DEBUG": "1"},
		},
		{
			name: "multiple_env_vars",
			setup: func() *Options {
				return NewOptions(WithEnv(map[string]string{
					"HTTP_PROXY": "http://proxy:8080",
					"CUSTOM_VAR": "value",
				}))
			},
			expected: map[string]string{
				"HTTP_PROXY": "http://proxy:8080",
				"CUSTOM_VAR": "value",
			},
		},
		{
			name: "merge_with_env_and_envvar",
			setup: func() *Options {
				return NewOptions(
					WithEnv(map[string]string{"VAR1": "val1"}),
					WithEnvVar("VAR2", "val2"),
				)
			},
			expected: map[string]string{
				"VAR1": "val1",
				"VAR2": "val2",
			},
		},
		{
			name: "override_existing",
			setup: func() *Options {
				return NewOptions(
					WithEnvVar("KEY", "original"),
					WithEnvVar("KEY", "updated"),
				)
			},
			expected: map[string]string{"KEY": "updated"},
		},
		{
			name: "empty_env_map",
			setup: func() *Options {
				return NewOptions(WithEnv(map[string]string{}))
			},
			expected: map[string]string{},
		},
		{
			name: "nil_env_map_initializes",
			setup: func() *Options {
				opts := &Options{} // ExtraEnv is nil
				WithEnvVar("TEST", "value")(opts)
				return opts
			},
			expected: map[string]string{"TEST": "value"},
		},
		{
			name: "proxy_configuration_example",
			setup: func() *Options {
				return NewOptions(
					WithEnv(map[string]string{
						"HTTP_PROXY":  "http://proxy.example.com:8080",
						"HTTPS_PROXY": "http://proxy.example.com:8080",
						"NO_PROXY":    "localhost,127.0.0.1",
					}),
				)
			},
			expected: map[string]string{
				"HTTP_PROXY":  "http://proxy.example.com:8080",
				"HTTPS_PROXY": "http://proxy.example.com:8080",
				"NO_PROXY":    "localhost,127.0.0.1",
			},
		},
		{
			name: "path_override_example",
			setup: func() *Options {
				return NewOptions(
					WithEnvVar("PATH", "/custom/bin:/usr/bin"),
				)
			},
			expected: map[string]string{
				"PATH": "/custom/bin:/usr/bin",
			},
		},
		{
			name: "nil_env_map_to_WithEnv",
			setup: func() *Options {
				opts := &Options{} // ExtraEnv is nil
				WithEnv(map[string]string{"TEST": "value"})(opts)
				return opts
			},
			expected: map[string]string{"TEST": "value"},
		},
		{
			name: "nil_map_passed_to_WithEnv",
			setup: func() *Options {
				return NewOptions(WithEnv(nil))
			},
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := tt.setup()
			assertEnvVars(t, options.ExtraEnv, tt.expected)
		})
	}
}

// TestWithEnvIntegration tests environment variable options integration with other options
func TestWithEnvIntegration(t *testing.T) {
	options := NewOptions(
		WithSystemPrompt("You are a helpful assistant"),
		WithEnvVar("DEBUG", "1"),
		WithModel("claude-3-5-sonnet-20241022"),
		WithEnv(map[string]string{
			"HTTP_PROXY": "http://proxy:8080",
			"CUSTOM":     "value",
		}),
		WithEnvVar("OVERRIDE", "final"),
	)

	// Test that env vars are correctly set
	expected := map[string]string{
		"DEBUG":      "1",
		"HTTP_PROXY": "http://proxy:8080",
		"CUSTOM":     "value",
		"OVERRIDE":   "final",
	}
	assertEnvVars(t, options.ExtraEnv, expected)

	// Test that other options are preserved
	assertOptionsSystemPrompt(t, options, "You are a helpful assistant")
	assertOptionsModel(t, options, "claude-3-5-sonnet-20241022")
}

// Helper function following client_test.go patterns
func assertEnvVars(t *testing.T, actual, expected map[string]string) {
	t.Helper()
	if len(actual) != len(expected) {
		t.Errorf("Expected %d env vars, got %d. Expected: %v, Actual: %v",
			len(expected), len(actual), expected, actual)
		return
	}
	for k, v := range expected {
		if actual[k] != v {
			t.Errorf("Expected %s=%s, got %s=%s", k, v, k, actual[k])
		}
	}
}
