package claudecode

import (
	"testing"
)

// T015: Default Options Creation
func TestDefaultOptions(t *testing.T) {
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

// T016: Options with Tools
func TestOptionsWithTools(t *testing.T) {
	// Test Options with allowed_tools and disallowed_tools to match Python SDK
	options := NewOptions(
		WithAllowedTools("Read", "Write", "Edit"),
		WithDisallowedTools("Bash"),
	)

	// Verify allowed tools
	expectedAllowed := []string{"Read", "Write", "Edit"}
	if len(options.AllowedTools) != len(expectedAllowed) {
		t.Errorf("Expected AllowedTools length = %d, got %d", len(expectedAllowed), len(options.AllowedTools))
	}
	for i, tool := range expectedAllowed {
		if i >= len(options.AllowedTools) || options.AllowedTools[i] != tool {
			t.Errorf("Expected AllowedTools[%d] = %q, got %q", i, tool, options.AllowedTools[i])
		}
	}

	// Verify disallowed tools
	expectedDisallowed := []string{"Bash"}
	if len(options.DisallowedTools) != len(expectedDisallowed) {
		t.Errorf("Expected DisallowedTools length = %d, got %d", len(expectedDisallowed), len(options.DisallowedTools))
	}
	for i, tool := range expectedDisallowed {
		if i >= len(options.DisallowedTools) || options.DisallowedTools[i] != tool {
			t.Errorf("Expected DisallowedTools[%d] = %q, got %q", i, tool, options.DisallowedTools[i])
		}
	}

	// Test with empty tools
	emptyOptions := NewOptions(
		WithAllowedTools(),
		WithDisallowedTools(),
	)
	if len(emptyOptions.AllowedTools) != 0 {
		t.Errorf("Expected empty AllowedTools, got %v", emptyOptions.AllowedTools)
	}
	if len(emptyOptions.DisallowedTools) != 0 {
		t.Errorf("Expected empty DisallowedTools, got %v", emptyOptions.DisallowedTools)
	}
}

// T017: Permission Mode Options
func TestPermissionModeOptions(t *testing.T) {
	// Test all permission modes: default, acceptEdits, plan, bypassPermissions

	// Test default permission mode
	defaultOptions := NewOptions(WithPermissionMode(PermissionModeDefault))
	if defaultOptions.PermissionMode == nil {
		t.Error("Expected PermissionMode to be set, got nil")
	}
	if *defaultOptions.PermissionMode != PermissionModeDefault {
		t.Errorf("Expected PermissionMode = %q, got %q", PermissionModeDefault, *defaultOptions.PermissionMode)
	}

	// Test acceptEdits permission mode
	acceptEditsOptions := NewOptions(WithPermissionMode(PermissionModeAcceptEdits))
	if acceptEditsOptions.PermissionMode == nil {
		t.Error("Expected PermissionMode to be set, got nil")
	}
	if *acceptEditsOptions.PermissionMode != PermissionModeAcceptEdits {
		t.Errorf("Expected PermissionMode = %q, got %q", PermissionModeAcceptEdits, *acceptEditsOptions.PermissionMode)
	}

	// Test plan permission mode
	planOptions := NewOptions(WithPermissionMode(PermissionModePlan))
	if planOptions.PermissionMode == nil {
		t.Error("Expected PermissionMode to be set, got nil")
	}
	if *planOptions.PermissionMode != PermissionModePlan {
		t.Errorf("Expected PermissionMode = %q, got %q", PermissionModePlan, *planOptions.PermissionMode)
	}

	// Test bypassPermissions permission mode
	bypassOptions := NewOptions(WithPermissionMode(PermissionModeBypassPermissions))
	if bypassOptions.PermissionMode == nil {
		t.Error("Expected PermissionMode to be set, got nil")
	}
	if *bypassOptions.PermissionMode != PermissionModeBypassPermissions {
		t.Errorf("Expected PermissionMode = %q, got %q", PermissionModeBypassPermissions, *bypassOptions.PermissionMode)
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
	if options.SystemPrompt == nil {
		t.Error("Expected SystemPrompt to be set, got nil")
	}
	if *options.SystemPrompt != systemPrompt {
		t.Errorf("Expected SystemPrompt = %q, got %q", systemPrompt, *options.SystemPrompt)
	}

	// Verify append system prompt is set
	if options.AppendSystemPrompt == nil {
		t.Error("Expected AppendSystemPrompt to be set, got nil")
	}
	if *options.AppendSystemPrompt != appendPrompt {
		t.Errorf("Expected AppendSystemPrompt = %q, got %q", appendPrompt, *options.AppendSystemPrompt)
	}

	// Test with only system prompt
	systemOnlyOptions := NewOptions(WithSystemPrompt("Only system prompt"))
	if systemOnlyOptions.SystemPrompt == nil {
		t.Error("Expected SystemPrompt to be set, got nil")
	}
	if *systemOnlyOptions.SystemPrompt != "Only system prompt" {
		t.Errorf("Expected SystemPrompt = %q, got %q", "Only system prompt", *systemOnlyOptions.SystemPrompt)
	}
	if systemOnlyOptions.AppendSystemPrompt != nil {
		t.Errorf("Expected AppendSystemPrompt = nil, got %v", systemOnlyOptions.AppendSystemPrompt)
	}

	// Test with only append prompt
	appendOnlyOptions := NewOptions(WithAppendSystemPrompt("Only append prompt"))
	if appendOnlyOptions.AppendSystemPrompt == nil {
		t.Error("Expected AppendSystemPrompt to be set, got nil")
	}
	if *appendOnlyOptions.AppendSystemPrompt != "Only append prompt" {
		t.Errorf("Expected AppendSystemPrompt = %q, got %q", "Only append prompt", *appendOnlyOptions.AppendSystemPrompt)
	}
	if appendOnlyOptions.SystemPrompt != nil {
		t.Errorf("Expected SystemPrompt = nil, got %v", appendOnlyOptions.SystemPrompt)
	}
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
	if options.ContinueConversation != true {
		t.Errorf("Expected ContinueConversation = true, got %v", options.ContinueConversation)
	}

	// Verify resume session ID is set
	if options.Resume == nil {
		t.Error("Expected Resume to be set, got nil")
	}
	if *options.Resume != sessionID {
		t.Errorf("Expected Resume = %q, got %q", sessionID, *options.Resume)
	}

	// Test with continue_conversation false
	falseOptions := NewOptions(WithContinueConversation(false))
	if falseOptions.ContinueConversation != false {
		t.Errorf("Expected ContinueConversation = false, got %v", falseOptions.ContinueConversation)
	}
	if falseOptions.Resume != nil {
		t.Errorf("Expected Resume = nil, got %v", falseOptions.Resume)
	}

	// Test with only resume
	resumeOnlyOptions := NewOptions(WithResume("another-session"))
	if resumeOnlyOptions.Resume == nil {
		t.Error("Expected Resume to be set, got nil")
	}
	if *resumeOnlyOptions.Resume != "another-session" {
		t.Errorf("Expected Resume = %q, got %q", "another-session", *resumeOnlyOptions.Resume)
	}
	if resumeOnlyOptions.ContinueConversation != false {
		t.Errorf("Expected ContinueConversation = false (default), got %v", resumeOnlyOptions.ContinueConversation)
	}
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

	// Verify model is set
	if options.Model == nil {
		t.Error("Expected Model to be set, got nil")
	}
	if *options.Model != model {
		t.Errorf("Expected Model = %q, got %q", model, *options.Model)
	}

	// Verify permission prompt tool name is set
	if options.PermissionPromptToolName == nil {
		t.Error("Expected PermissionPromptToolName to be set, got nil")
	}
	if *options.PermissionPromptToolName != toolName {
		t.Errorf("Expected PermissionPromptToolName = %q, got %q", toolName, *options.PermissionPromptToolName)
	}

	// Test with only model
	modelOnlyOptions := NewOptions(WithModel("claude-opus-4"))
	if modelOnlyOptions.Model == nil {
		t.Error("Expected Model to be set, got nil")
	}
	if *modelOnlyOptions.Model != "claude-opus-4" {
		t.Errorf("Expected Model = %q, got %q", "claude-opus-4", *modelOnlyOptions.Model)
	}
	if modelOnlyOptions.PermissionPromptToolName != nil {
		t.Errorf("Expected PermissionPromptToolName = nil, got %v", modelOnlyOptions.PermissionPromptToolName)
	}

	// Test with only permission prompt tool name
	toolOnlyOptions := NewOptions(WithPermissionPromptToolName("OnlyTool"))
	if toolOnlyOptions.PermissionPromptToolName == nil {
		t.Error("Expected PermissionPromptToolName to be set, got nil")
	}
	if *toolOnlyOptions.PermissionPromptToolName != "OnlyTool" {
		t.Errorf("Expected PermissionPromptToolName = %q, got %q", "OnlyTool", *toolOnlyOptions.PermissionPromptToolName)
	}
	if toolOnlyOptions.Model != nil {
		t.Errorf("Expected Model = nil, got %v", toolOnlyOptions.Model)
	}
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
func TestOptionsValidation(t *testing.T) {
	// Test options field validation and constraints

	// Test valid options should pass validation
	validOptions := NewOptions(
		WithAllowedTools("Read", "Write"),
		WithMaxThinkingTokens(8000),
		WithSystemPrompt("Valid prompt"),
	)
	if err := validOptions.Validate(); err != nil {
		t.Errorf("Expected valid options to pass validation, got error: %v", err)
	}

	// Test negative max thinking tokens should fail
	invalidTokensOptions := NewOptions(WithMaxThinkingTokens(-100))
	if err := invalidTokensOptions.Validate(); err == nil {
		t.Error("Expected negative max thinking tokens to fail validation")
	} else if err.Error() != "MaxThinkingTokens must be non-negative, got -100" {
		t.Errorf("Expected specific error for negative tokens, got: %v", err)
	}

	// Test zero max thinking tokens should be valid
	zeroTokensOptions := NewOptions(WithMaxThinkingTokens(0))
	if err := zeroTokensOptions.Validate(); err != nil {
		t.Errorf("Expected zero max thinking tokens to be valid, got error: %v", err)
	}

	// Test empty system prompt should be valid (nil is also valid)
	emptyPromptOptions := NewOptions(WithSystemPrompt(""))
	if err := emptyPromptOptions.Validate(); err != nil {
		t.Errorf("Expected empty system prompt to be valid, got error: %v", err)
	}

	// Test conflicting tool lists (same tool in both allowed and disallowed)
	conflictingOptions := NewOptions(
		WithAllowedTools("Read", "Write", "Edit"),
		WithDisallowedTools("Write", "Bash"),
	)
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
		t.Errorf("Expected specific error for negative MaxTurns, got: %v", err)
	}

	// Test zero MaxTurns should be valid (means no limit)
	zeroTurnsOptions := NewOptions()
	zeroTurnsOptions.MaxTurns = 0
	if err := zeroTurnsOptions.Validate(); err != nil {
		t.Errorf("Expected zero MaxTurns to be valid, got error: %v", err)
	}
}

// T025: NewOptions Constructor
func TestNewOptionsConstructor(t *testing.T) {
	// Test Options creation with functional options applied correctly with defaults

	// Test NewOptions with no arguments should return defaults
	defaultOptions := NewOptions()
	if defaultOptions.MaxThinkingTokens != 8000 {
		t.Errorf("Expected default MaxThinkingTokens = 8000, got %d", defaultOptions.MaxThinkingTokens)
	}
	if len(defaultOptions.AllowedTools) != 0 {
		t.Errorf("Expected default AllowedTools = [], got %v", defaultOptions.AllowedTools)
	}

	// Test NewOptions with single functional option
	singleOptionOptions := NewOptions(WithSystemPrompt("Single option test"))
	if singleOptionOptions.SystemPrompt == nil || *singleOptionOptions.SystemPrompt != "Single option test" {
		t.Errorf("Expected SystemPrompt = %q, got %v", "Single option test", singleOptionOptions.SystemPrompt)
	}
	// Should still have defaults for other fields
	if singleOptionOptions.MaxThinkingTokens != 8000 {
		t.Errorf("Expected default MaxThinkingTokens = 8000, got %d", singleOptionOptions.MaxThinkingTokens)
	}

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
	if multipleOptions.MaxThinkingTokens != 12000 {
		t.Errorf("Expected MaxThinkingTokens = 12000 (final override), got %d", multipleOptions.MaxThinkingTokens)
	}

	expectedTools := []string{"Read", "Write", "Edit"}
	if len(multipleOptions.AllowedTools) != len(expectedTools) {
		t.Errorf("Expected AllowedTools length = %d, got %d", len(expectedTools), len(multipleOptions.AllowedTools))
	}
	for i, tool := range expectedTools {
		if i >= len(multipleOptions.AllowedTools) || multipleOptions.AllowedTools[i] != tool {
			t.Errorf("Expected AllowedTools[%d] = %q, got %q", i, tool, multipleOptions.AllowedTools[i])
		}
	}

	if multipleOptions.SystemPrompt == nil || *multipleOptions.SystemPrompt != "Second prompt" {
		t.Errorf("Expected SystemPrompt = %q (final override), got %v", "Second prompt", multipleOptions.SystemPrompt)
	}

	if len(multipleOptions.DisallowedTools) != 1 || multipleOptions.DisallowedTools[0] != "Bash" {
		t.Errorf("Expected DisallowedTools = [Bash], got %v", multipleOptions.DisallowedTools)
	}

	if multipleOptions.PermissionMode == nil || *multipleOptions.PermissionMode != PermissionModeAcceptEdits {
		t.Errorf("Expected PermissionMode = %q, got %v", PermissionModeAcceptEdits, multipleOptions.PermissionMode)
	}

	if multipleOptions.ContinueConversation != true {
		t.Errorf("Expected ContinueConversation = true, got %v", multipleOptions.ContinueConversation)
	}

	if multipleOptions.MaxTurns != 5 {
		t.Errorf("Expected MaxTurns = 5, got %d", multipleOptions.MaxTurns)
	}

	if multipleOptions.Settings == nil || *multipleOptions.Settings != "/path/to/settings.json" {
		t.Errorf("Expected Settings = %q, got %v", "/path/to/settings.json", multipleOptions.Settings)
	}

	// Test that unmodified fields retain defaults
	if multipleOptions.Resume != nil {
		t.Errorf("Expected Resume = nil (default), got %v", multipleOptions.Resume)
	}

	if multipleOptions.Cwd != nil {
		t.Errorf("Expected Cwd = nil (default), got %v", multipleOptions.Cwd)
	}

	// Test that maps are properly initialized even with options
	if multipleOptions.McpServers == nil {
		t.Error("Expected McpServers to be initialized, got nil")
	}
	if len(multipleOptions.McpServers) != 0 {
		t.Errorf("Expected McpServers = {} (default), got %v", multipleOptions.McpServers)
	}

	if multipleOptions.ExtraArgs == nil {
		t.Error("Expected ExtraArgs to be initialized, got nil")
	}
	if len(multipleOptions.ExtraArgs) != 0 {
		t.Errorf("Expected ExtraArgs = {} (default), got %v", multipleOptions.ExtraArgs)
	}
}

// Helper function for creating string pointers
func stringPtr(s string) *string {
	return &s
}
