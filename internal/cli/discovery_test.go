package cli

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/severity1/claude-code-sdk-go/internal/shared"
)

const (
	printFlag = "--print"
)

// TestCLIDiscovery tests CLI binary discovery functionality
func TestCLIDiscovery(t *testing.T) {
	tests := []struct {
		name          string
		setupEnv      func(t *testing.T) (cleanup func())
		expectError   bool
		errorContains string
	}{
		{
			name:          "cli_not_found_error",
			setupEnv:      setupIsolatedEnvironment,
			expectError:   true,
			errorContains: "install",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cleanup := test.setupEnv(t)
			defer cleanup()

			_, err := FindCLI()
			assertCLIDiscoveryError(t, err, test.expectError, test.errorContains)
		})
	}
}

// TestCommandBuilding tests CLI command construction with various options
func TestCommandBuilding(t *testing.T) {
	tests := []struct {
		name       string
		cliPath    string
		options    *shared.Options
		closeStdin bool
		validate   func(*testing.T, []string)
	}{
		{
			name:       "basic_oneshot_command",
			cliPath:    "/usr/local/bin/claude",
			options:    &shared.Options{},
			closeStdin: true,
			validate:   validateOneshotCommand,
		},
		{
			name:       "basic_streaming_command",
			cliPath:    "/usr/local/bin/claude",
			options:    &shared.Options{},
			closeStdin: false,
			validate:   validateStreamingCommand,
		},
		{
			name:       "all_options_command",
			cliPath:    "/usr/local/bin/claude",
			options:    createFullOptionsSet(),
			closeStdin: false,
			validate:   validateFullOptionsCommand,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := BuildCommand(test.cliPath, test.options, test.closeStdin)
			test.validate(t, cmd)
		})
	}
}

// TestCLIPathHandling tests various CLI path formats
func TestCLIPathHandling(t *testing.T) {
	tests := []struct {
		name     string
		cliPath  string
		expected string
	}{
		{"absolute_path", "/usr/local/bin/claude", "/usr/local/bin/claude"},
		{"relative_path", "./claude", "./claude"},
		{"complex_path", "/usr/local/bin/../bin/./claude", "/usr/local/bin/../bin/./claude"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := BuildCommand(test.cliPath, &shared.Options{}, true)
			assertCLIPath(t, cmd, test.expected)
		})
	}
}

// TestCLIDiscoveryLocations tests CLI discovery path generation
func TestCLIDiscoveryLocations(t *testing.T) {
	locations := getCommonCLILocations()

	assertDiscoveryLocations(t, locations)
	assertPlatformSpecificPaths(t, locations)
}

// TestNodeJSDependencyValidation tests Node.js validation
func TestNodeJSDependencyValidation(t *testing.T) {
	err := ValidateNodeJS()
	assertNodeJSValidation(t, err)
}

// TestExtraArgsSupport tests arbitrary CLI flag support
func TestExtraArgsSupport(t *testing.T) {
	tests := []struct {
		name      string
		extraArgs map[string]*string
		validate  func(*testing.T, []string)
	}{
		{
			name:      "boolean_flags",
			extraArgs: map[string]*string{"debug": nil, "trace": nil},
			validate:  validateBooleanExtraArgs,
		},
		{
			name:      "value_flags",
			extraArgs: map[string]*string{"log-level": &[]string{"info"}[0]},
			validate:  validateValueExtraArgs,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			options := &shared.Options{ExtraArgs: test.extraArgs}
			cmd := BuildCommand("/usr/local/bin/claude", options, true)
			test.validate(t, cmd)
		})
	}
}

// TestBuildCommandWithPrompt tests CLI command construction with prompt argument
func TestBuildCommandWithPrompt(t *testing.T) {
	tests := []struct {
		name     string
		cliPath  string
		options  *shared.Options
		prompt   string
		validate func(*testing.T, []string, string)
	}{
		{
			name:     "basic_prompt_command",
			cliPath:  "/usr/local/bin/claude",
			options:  &shared.Options{},
			prompt:   "What is 2+2?",
			validate: validateBasicPromptCommand,
		},
		{
			name:    "prompt_with_system_prompt",
			cliPath: "/usr/local/bin/claude",
			options: &shared.Options{
				SystemPrompt: stringPtr("You are a helpful assistant"),
			},
			prompt:   "Hello there",
			validate: validatePromptWithSystemPrompt,
		},
		{
			name:     "prompt_with_full_options",
			cliPath:  "/usr/local/bin/claude",
			options:  createFullOptionsSet(),
			prompt:   "Complex query with all options",
			validate: validatePromptWithFullOptions,
		},
		{
			name:     "empty_prompt",
			cliPath:  "/usr/local/bin/claude",
			options:  &shared.Options{},
			prompt:   "",
			validate: validateEmptyPromptCommand,
		},
		{
			name:     "multiline_prompt",
			cliPath:  "/usr/local/bin/claude",
			options:  &shared.Options{},
			prompt:   "Line 1\nLine 2\nLine 3",
			validate: validateMultilinePromptCommand,
		},
		{
			name:     "special_characters_prompt",
			cliPath:  "/usr/local/bin/claude",
			options:  &shared.Options{},
			prompt:   "Test with \"quotes\" and 'apostrophes' and $variables",
			validate: validateSpecialCharactersPromptCommand,
		},
		{
			name:     "nil_options",
			cliPath:  "/usr/local/bin/claude",
			options:  nil,
			prompt:   "Test with nil options",
			validate: validateNilOptionsPromptCommand,
		},
		{
			name:    "prompt_with_tools_and_model",
			cliPath: "/usr/local/bin/claude",
			options: &shared.Options{
				AllowedTools:    []string{"Read", "Write"},
				DisallowedTools: []string{"Bash"},
				Model:           stringPtr("claude-sonnet-3-5-20241022"),
			},
			prompt:   "Use tools to help me",
			validate: validatePromptWithToolsAndModel,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := BuildCommandWithPrompt(test.cliPath, test.options, test.prompt)
			test.validate(t, cmd, test.prompt)
		})
	}
}

// TestBuildCommandWithPromptVsBuildCommand tests differences between prompt and regular command building
func TestBuildCommandWithPromptVsBuildCommand(t *testing.T) {
	cliPath := "/usr/local/bin/claude"
	options := &shared.Options{
		SystemPrompt: stringPtr("Test prompt"),
		Model:        stringPtr("claude-sonnet-3-5-20241022"),
	}
	prompt := "Test query"

	// Build both types of commands
	regularCommand := BuildCommand(cliPath, options, true) // closeStdin=true for one-shot
	promptCommand := BuildCommandWithPrompt(cliPath, options, prompt)

	tests := []struct {
		name  string
		check func(t *testing.T)
	}{
		{
			name: "both_have_print_flag",
			check: func(t *testing.T) {
				assertContainsArg(t, regularCommand, "--print")
				assertContainsArg(t, promptCommand, "--print")
			},
		},
		{
			name: "prompt_command_includes_prompt_as_argument",
			check: func(t *testing.T) {
				// promptCommand should have the prompt as an argument after --print
				found := false
				for i, arg := range promptCommand {
					if arg == printFlag && i+1 < len(promptCommand) && promptCommand[i+1] == prompt {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected prompt command to have prompt %q as argument after --print, got %v", prompt, promptCommand)
				}
			},
		},
		{
			name: "regular_command_no_prompt_argument",
			check: func(t *testing.T) {
				// Regular command should not have prompt as argument
				for i, arg := range regularCommand {
					if arg == printFlag && i+1 < len(regularCommand) && regularCommand[i+1] == prompt {
						t.Errorf("Expected regular command to not have prompt as argument, got %v", regularCommand)
						break
					}
				}
			},
		},
		{
			name: "both_contain_same_options",
			check: func(t *testing.T) {
				// Both should contain the same options flags
				assertContainsArgs(t, regularCommand, "--system-prompt", "Test prompt")
				assertContainsArgs(t, promptCommand, "--system-prompt", "Test prompt")
				assertContainsArgs(t, regularCommand, "--model", "claude-sonnet-3-5-20241022")
				assertContainsArgs(t, promptCommand, "--model", "claude-sonnet-3-5-20241022")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.check(t)
		})
	}
}

// TestWorkingDirectoryValidation tests working directory validation
func TestWorkingDirectoryValidation(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(t *testing.T) string
		expectError   bool
		errorContains string
	}{
		{
			name:        "existing_directory",
			setup:       func(t *testing.T) string { return t.TempDir() },
			expectError: false,
		},
		{
			name:        "empty_path",
			setup:       func(_ *testing.T) string { return "" },
			expectError: false,
		},
		{
			name: "nonexistent_directory",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "does-not-exist")
			},
			expectError: true,
		},
		{
			name: "file_not_directory",
			setup: func(t *testing.T) string {
				tempFile := filepath.Join(t.TempDir(), "testfile")
				if err := os.WriteFile(tempFile, []byte("test"), 0o600); err != nil {
					t.Fatalf("Failed to write test file: %v", err)
				}
				return tempFile
			},
			expectError:   true,
			errorContains: "not a directory",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := test.setup(t)
			err := ValidateWorkingDirectory(path)
			assertValidationError(t, err, test.expectError, test.errorContains)
		})
	}
}

// TestCLIVersionDetection tests CLI version detection
func TestCLIVersionDetection(t *testing.T) {
	nonExistentPath := "/this/path/does/not/exist/claude"
	ctx := context.Background()
	_, err := DetectCLIVersion(ctx, nonExistentPath)
	assertVersionDetectionError(t, err)
}

// Helper Functions

func setupIsolatedEnvironment(t *testing.T) func() {
	t.Helper()
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	originalPath := os.Getenv("PATH")

	if runtime.GOOS == windowsOS {
		originalHome = os.Getenv("USERPROFILE")
		_ = os.Setenv("USERPROFILE", tempHome)
	} else {
		_ = os.Setenv("HOME", tempHome)
	}
	_ = os.Setenv("PATH", "/nonexistent/path")

	return func() {
		if runtime.GOOS == windowsOS {
			_ = os.Setenv("USERPROFILE", originalHome)
		} else {
			_ = os.Setenv("HOME", originalHome)
		}
		_ = os.Setenv("PATH", originalPath)
	}
}

func createFullOptionsSet() *shared.Options {
	systemPrompt := "You are a helpful assistant"
	appendPrompt := "Additional context"
	model := "claude-3-sonnet"
	permissionMode := shared.PermissionModeAcceptEdits
	resume := "session123"
	settings := "/path/to/settings.json"
	cwd := "/workspace"
	testValue := "test"

	return &shared.Options{
		AllowedTools:         []string{"Read", "Write"},
		DisallowedTools:      []string{"Bash", "Delete"},
		SystemPrompt:         &systemPrompt,
		AppendSystemPrompt:   &appendPrompt,
		Model:                &model,
		MaxThinkingTokens:    10000,
		PermissionMode:       &permissionMode,
		ContinueConversation: true,
		Resume:               &resume,
		MaxTurns:             25,
		Settings:             &settings,
		Cwd:                  &cwd,
		AddDirs:              []string{"/extra/dir1", "/extra/dir2"},
		McpServers:           make(map[string]shared.McpServerConfig),
		ExtraArgs:            map[string]*string{"custom-flag": nil, "with-value": &testValue},
	}
}

// Assertion helpers

func assertCLIDiscoveryError(t *testing.T, err error, expectError bool, errorContains string) {
	t.Helper()
	if (err != nil) != expectError {
		t.Errorf("error = %v, expectError %v", err, expectError)
		return
	}
	if expectError && errorContains != "" && !strings.Contains(err.Error(), errorContains) {
		t.Errorf("error = %v, expected message to contain %q", err, errorContains)
	}
}

func assertCLIPath(t *testing.T, cmd []string, expected string) {
	t.Helper()
	if len(cmd) == 0 || cmd[0] != expected {
		t.Errorf("Expected CLI path %s, got %v", expected, cmd)
	}
}

func assertDiscoveryLocations(t *testing.T, locations []string) {
	t.Helper()
	if len(locations) == 0 {
		t.Fatal("Expected at least one CLI location, got none")
	}
}

func assertPlatformSpecificPaths(t *testing.T, locations []string) {
	t.Helper()
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	expectedNpmGlobal := filepath.Join(homeDir, ".npm-global", "bin", "claude")
	if runtime.GOOS == windowsOS {
		expectedNpmGlobal = filepath.Join(homeDir, ".npm-global", "claude.cmd")
	}

	found := false
	for _, location := range locations {
		if location == expectedNpmGlobal {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected npm-global location %s in discovery paths", expectedNpmGlobal)
	}
}

func assertNodeJSValidation(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		errMsg := err.Error()
		if !strings.Contains(errMsg, "Node.js") {
			t.Error("Error message should mention Node.js")
		}
		if !strings.Contains(errMsg, "https://nodejs.org") {
			t.Error("Error message should include Node.js download URL")
		}
	}
}

func assertValidationError(t *testing.T, err error, expectError bool, errorContains string) {
	t.Helper()
	if (err != nil) != expectError {
		t.Errorf("error = %v, expectError %v", err, expectError)
		return
	}
	if expectError && errorContains != "" && !strings.Contains(err.Error(), errorContains) {
		t.Errorf("error = %v, expected message to contain %q", err, errorContains)
	}
}

func assertVersionDetectionError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Error("Expected error when CLI path does not exist")
		return
	}
	if !strings.Contains(err.Error(), "version") {
		t.Error("Error message should mention version detection failure")
	}
}

// Command validation helpers

func validateOneshotCommand(t *testing.T, cmd []string) {
	t.Helper()
	assertContainsArgs(t, cmd, "--output-format", "stream-json")
	assertContainsArg(t, cmd, "--verbose")
	assertContainsArg(t, cmd, "--print")
	assertNotContainsArgs(t, cmd, "--input-format", "stream-json")
}

func validateStreamingCommand(t *testing.T, cmd []string) {
	t.Helper()
	assertContainsArgs(t, cmd, "--output-format", "stream-json")
	assertContainsArg(t, cmd, "--verbose")
	assertContainsArgs(t, cmd, "--input-format", "stream-json")
	assertNotContainsArg(t, cmd, "--print")
}

func validateFullOptionsCommand(t *testing.T, cmd []string) {
	t.Helper()
	assertContainsArgs(t, cmd, "--allowed-tools", "Read,Write")
	assertContainsArgs(t, cmd, "--disallowed-tools", "Bash,Delete")
	assertContainsArgs(t, cmd, "--system-prompt", "You are a helpful assistant")
	assertContainsArgs(t, cmd, "--model", "claude-3-sonnet")
	assertContainsArg(t, cmd, "--continue")
	assertContainsArgs(t, cmd, "--resume", "session123")
	assertContainsArg(t, cmd, "--custom-flag")
	assertContainsArgs(t, cmd, "--with-value", "test")
}

func validateBooleanExtraArgs(t *testing.T, cmd []string) {
	t.Helper()
	assertContainsArg(t, cmd, "--debug")
	assertContainsArg(t, cmd, "--trace")
}

func validateValueExtraArgs(t *testing.T, cmd []string) {
	t.Helper()
	assertContainsArgs(t, cmd, "--log-level", "info")
}

// Low-level assertion helpers

func assertContainsArg(t *testing.T, args []string, target string) {
	t.Helper()
	for _, arg := range args {
		if arg == target {
			return
		}
	}
	t.Errorf("Expected command to contain %s, got %v", target, args)
}

func assertNotContainsArg(t *testing.T, args []string, target string) {
	t.Helper()
	for _, arg := range args {
		if arg == target {
			t.Errorf("Expected command to not contain %s, got %v", target, args)
			return
		}
	}
}

func assertContainsArgs(t *testing.T, args []string, flag, value string) {
	t.Helper()
	for i, arg := range args {
		if arg == flag && i+1 < len(args) && args[i+1] == value {
			return
		}
	}
	t.Errorf("Expected command to contain %s %s, got %v", flag, value, args)
}

func assertNotContainsArgs(t *testing.T, args []string, flag, value string) {
	t.Helper()
	for i, arg := range args {
		if arg == flag && i+1 < len(args) && args[i+1] == value {
			t.Errorf("Expected command to not contain %s %s, got %v", flag, value, args)
			return
		}
	}
}

// Validation functions for BuildCommandWithPrompt tests

func validateBasicPromptCommand(t *testing.T, cmd []string, prompt string) {
	t.Helper()
	assertContainsArgs(t, cmd, "--output-format", "stream-json")
	assertContainsArg(t, cmd, "--verbose")
	assertContainsArgs(t, cmd, "--print", prompt)
}

func validatePromptWithSystemPrompt(t *testing.T, cmd []string, prompt string) {
	t.Helper()
	validateBasicPromptCommand(t, cmd, prompt)
	assertContainsArgs(t, cmd, "--system-prompt", "You are a helpful assistant")
}

func validatePromptWithFullOptions(t *testing.T, cmd []string, prompt string) {
	t.Helper()
	validateBasicPromptCommand(t, cmd, prompt)
	assertContainsArgs(t, cmd, "--allowed-tools", "Read,Write")
	assertContainsArgs(t, cmd, "--disallowed-tools", "Bash,Delete")
	assertContainsArgs(t, cmd, "--system-prompt", "You are a helpful assistant")
	assertContainsArgs(t, cmd, "--model", "claude-3-sonnet")
	assertContainsArg(t, cmd, "--continue")
	assertContainsArgs(t, cmd, "--resume", "session123")
}

func validateEmptyPromptCommand(t *testing.T, cmd []string, _ string) {
	t.Helper()
	assertContainsArgs(t, cmd, "--output-format", "stream-json")
	assertContainsArg(t, cmd, "--verbose")
	assertContainsArgs(t, cmd, "--print", "") // Empty prompt should still be there
}

func validateMultilinePromptCommand(t *testing.T, cmd []string, prompt string) {
	t.Helper()
	validateBasicPromptCommand(t, cmd, prompt)
	// Verify multiline prompt is preserved as single argument
	found := false
	for i, arg := range cmd {
		if arg == "--print" && i+1 < len(cmd) && cmd[i+1] == prompt {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected multiline prompt %q to be preserved as single argument", prompt)
	}
}

func validateSpecialCharactersPromptCommand(t *testing.T, cmd []string, prompt string) {
	t.Helper()
	validateBasicPromptCommand(t, cmd, prompt)
	// Verify special characters are preserved
	found := false
	for i, arg := range cmd {
		if arg == "--print" && i+1 < len(cmd) && cmd[i+1] == prompt {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected special characters prompt %q to be preserved", prompt)
	}
}

func validateNilOptionsPromptCommand(t *testing.T, cmd []string, prompt string) {
	t.Helper()
	// With nil options, should still have basic prompt command structure
	assertContainsArgs(t, cmd, "--output-format", "stream-json")
	assertContainsArg(t, cmd, "--verbose")
	assertContainsArgs(t, cmd, "--print", prompt)
	// Should not contain any option-specific flags
	assertNotContainsArg(t, cmd, "--system-prompt")
	assertNotContainsArg(t, cmd, "--model")
}

func validatePromptWithToolsAndModel(t *testing.T, cmd []string, prompt string) {
	t.Helper()
	validateBasicPromptCommand(t, cmd, prompt)
	assertContainsArgs(t, cmd, "--allowed-tools", "Read,Write")
	assertContainsArgs(t, cmd, "--disallowed-tools", "Bash")
	assertContainsArgs(t, cmd, "--model", "claude-sonnet-3-5-20241022")
}

// Helper function for string pointers
func stringPtr(s string) *string {
	return &s
}
