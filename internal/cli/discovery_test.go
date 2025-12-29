package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/severity1/claude-code-sdk-go/internal/shared"
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

// TestCwdNotAddedToCommand tests that WithCwd() doesn't add --cwd flag
func TestCwdNotAddedToCommand(t *testing.T) {
	cwd := "/workspace/test"
	options := &shared.Options{
		Cwd: &cwd,
	}

	cmd := BuildCommand("/usr/local/bin/claude", options, false)

	// Verify --cwd flag is NOT in the command
	assertNotContainsArg(t, cmd, "--cwd")

	// Verify the working directory path is also NOT in the command
	for _, arg := range cmd {
		if arg == cwd {
			t.Errorf("Expected command to not contain working directory path %s as argument, got %v", cwd, cmd)
		}
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

// TestBetasFlagSupport tests SDK beta features CLI flag support
func TestBetasFlagSupport(t *testing.T) {
	tests := []struct {
		name     string
		betas    []shared.SdkBeta
		validate func(*testing.T, []string)
	}{
		{
			name:     "single_beta",
			betas:    []shared.SdkBeta{shared.SdkBetaContext1M},
			validate: validateSingleBetaFlag,
		},
		{
			name:     "multiple_betas",
			betas:    []shared.SdkBeta{shared.SdkBetaContext1M, "other-beta"},
			validate: validateMultipleBetasFlag,
		},
		{
			name:     "empty_betas",
			betas:    []shared.SdkBeta{},
			validate: validateNoBetasFlag,
		},
		{
			name:     "nil_betas",
			betas:    nil,
			validate: validateNoBetasFlag,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			options := &shared.Options{Betas: test.betas}
			cmd := BuildCommand("/usr/local/bin/claude", options, true)
			test.validate(t, cmd)
		})
	}
}

// TestBuildCommandWithPrompt tests CLI command construction with prompt argument
func TestBuildCommandWithPrompt(t *testing.T) {
	tests := []struct {
		name     string
		options  *shared.Options
		prompt   string
		validate func(*testing.T, []string, string)
	}{
		{"basic_prompt", &shared.Options{}, "What is 2+2?", validateBasicPromptCommand},
		{"empty_prompt", nil, "", validateEmptyPromptCommand},
		{"multiline_prompt", &shared.Options{Model: stringPtr("claude-3-sonnet")}, "Line 1\nLine 2", validateBasicPromptCommand},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := BuildCommandWithPrompt("/usr/local/bin/claude", test.options, test.prompt)
			test.validate(t, cmd, test.prompt)
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

func validateSingleBetaFlag(t *testing.T, cmd []string) {
	t.Helper()
	assertContainsArgs(t, cmd, "--betas", "context-1m-2025-08-07")
}

func validateMultipleBetasFlag(t *testing.T, cmd []string) {
	t.Helper()
	assertContainsArgs(t, cmd, "--betas", "context-1m-2025-08-07,other-beta")
}

func validateNoBetasFlag(t *testing.T, cmd []string) {
	t.Helper()
	assertNotContainsArg(t, cmd, "--betas")
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

func validateEmptyPromptCommand(t *testing.T, cmd []string, _ string) {
	t.Helper()
	assertContainsArgs(t, cmd, "--output-format", "stream-json")
	assertContainsArg(t, cmd, "--verbose")
	assertContainsArgs(t, cmd, "--print", "") // Empty prompt should still be there
}

// Helper function for string pointers
// TestFindCLISuccess tests successful CLI discovery paths
func TestFindCLISuccess(t *testing.T) {
	// Test when CLI is found in PATH
	t.Run("cli_found_in_path", func(t *testing.T) {
		// Create a temporary executable file
		tempDir := t.TempDir()
		cliPath := filepath.Join(tempDir, "claude")
		if runtime.GOOS == windowsOS {
			cliPath += ".exe"
		}

		// Create and make executable
		//nolint:gosec // G306: Test file needs execute permission for mock CLI binary
		err := os.WriteFile(cliPath, []byte("#!/bin/bash\necho test"), 0o700)
		if err != nil {
			t.Fatalf("Failed to create test CLI: %v", err)
		}

		// Temporarily modify PATH
		originalPath := os.Getenv("PATH")
		newPath := tempDir + string(os.PathListSeparator) + originalPath
		if err := os.Setenv("PATH", newPath); err != nil {
			t.Fatalf("Failed to set PATH: %v", err)
		}
		defer func() {
			if err := os.Setenv("PATH", originalPath); err != nil {
				t.Logf("Failed to restore PATH: %v", err)
			}
		}()

		found, err := FindCLI()
		if err != nil {
			t.Errorf("Expected CLI to be found, got error: %v", err)
		}
		if !strings.Contains(found, "claude") {
			t.Errorf("Expected found path to contain 'claude', got: %s", found)
		}
	})

	// Test executable validation on Unix
	if runtime.GOOS != windowsOS {
		t.Run("non_executable_file_skipped", func(t *testing.T) {
			// Create a non-executable file in a location that would be found
			tempDir := t.TempDir()
			cliPath := filepath.Join(tempDir, ".npm-global", "bin", "claude")
			if err := os.MkdirAll(filepath.Dir(cliPath), 0o750); err != nil {
				t.Fatalf("Failed to create directory: %v", err)
			}
			if err := os.WriteFile(cliPath, []byte("not executable"), 0o600); err != nil {
				t.Fatalf("Failed to write file: %v", err)
			}

			// Mock home directory
			originalHome := os.Getenv("HOME")
			if err := os.Setenv("HOME", tempDir); err != nil {
				t.Fatalf("Failed to set HOME: %v", err)
			}
			defer func() {
				if err := os.Setenv("HOME", originalHome); err != nil {
					t.Logf("Failed to restore HOME: %v", err)
				}
			}()

			// Isolate PATH to force common location search
			originalPath := os.Getenv("PATH")
			if err := os.Setenv("PATH", "/nonexistent"); err != nil {
				t.Fatalf("Failed to set PATH: %v", err)
			}
			defer func() {
				if err := os.Setenv("PATH", originalPath); err != nil {
					t.Logf("Failed to restore PATH: %v", err)
				}
			}()

			_, err := FindCLI()
			// Should fail because file is not executable
			if err == nil {
				t.Error("Expected error for non-executable file")
			}
		})
	}
}

// TestGetCommonCLILocationsPlatforms tests platform-specific path generation
func TestGetCommonCLILocationsPlatforms(t *testing.T) {
	// Test Windows paths
	if runtime.GOOS == windowsOS {
		t.Run("windows_paths", func(t *testing.T) {
			locations := getCommonCLILocations()

			// Check for Windows-specific patterns
			foundAppData := false
			foundProgramFiles := false

			for _, location := range locations {
				if strings.Contains(location, "AppData") && strings.HasSuffix(location, ".cmd") {
					foundAppData = true
				}
				if strings.Contains(location, "Program Files") && strings.HasSuffix(location, ".cmd") {
					foundProgramFiles = true
				}
			}

			if !foundAppData {
				t.Error("Expected Windows AppData path with .cmd extension")
			}
			if !foundProgramFiles {
				t.Error("Expected Program Files path with .cmd extension")
			}
		})
	}

	// Test home directory fallback
	t.Run("home_directory_fallback", func(t *testing.T) {
		// Temporarily unset home directory env vars
		var originalHome string
		var envVar string

		if runtime.GOOS == windowsOS {
			envVar = "USERPROFILE"
		} else {
			envVar = "HOME"
		}

		originalHome = os.Getenv(envVar)
		if err := os.Unsetenv(envVar); err != nil {
			t.Fatalf("Failed to unset %s: %v", envVar, err)
		}
		defer func() {
			if err := os.Setenv(envVar, originalHome); err != nil {
				t.Logf("Failed to restore %s: %v", envVar, err)
			}
		}()

		locations := getCommonCLILocations()
		// Should still return paths, using current directory as fallback
		if len(locations) == 0 {
			t.Error("Expected fallback paths when home directory unavailable")
		}
	})
}

// TestValidateNodeJSSuccess tests successful Node.js validation
func TestValidateNodeJSSuccess(t *testing.T) {
	// This test assumes Node.js is available in the test environment
	// If Node.js is not available, we'll create a mock
	err := ValidateNodeJS()
	if err != nil {
		// Node.js not found - test the error path
		assertNodeJSValidation(t, err)
	} else {
		// Node.js found - validation should succeed
		t.Log("Node.js validation succeeded")
	}
}

// TestDetectCLIVersionSuccess tests successful version detection
func TestDetectCLIVersionSuccess(t *testing.T) {
	ctx := context.Background()

	// Create a mock CLI that outputs a version
	tempDir := t.TempDir()
	mockCLI := filepath.Join(tempDir, "mock-claude")
	if runtime.GOOS == windowsOS {
		mockCLI += ".bat"
	}

	var script string
	if runtime.GOOS == windowsOS {
		script = "@echo off\necho 1.2.3"
	} else {
		script = "#!/bin/bash\necho '1.2.3'"
	}

	//nolint:gosec // G306: Test file needs execute permission for mock CLI binary
	err := os.WriteFile(mockCLI, []byte(script), 0o700)
	if err != nil {
		t.Fatalf("Failed to create mock CLI: %v", err)
	}

	version, err := DetectCLIVersion(ctx, mockCLI)
	if err != nil {
		t.Errorf("Expected successful version detection, got error: %v", err)
		return
	}

	if version != "1.2.3" {
		t.Errorf("Expected version '1.2.3', got '%s'", version)
	}
}

// TestDetectCLIVersionInvalidFormat tests version format validation
func TestDetectCLIVersionInvalidFormat(t *testing.T) {
	ctx := context.Background()

	// Create a mock CLI that outputs invalid version format
	tempDir := t.TempDir()
	mockCLI := filepath.Join(tempDir, "mock-claude-invalid")
	if runtime.GOOS == windowsOS {
		mockCLI += ".bat"
	}

	var script string
	if runtime.GOOS == windowsOS {
		script = "@echo off\necho invalid-version-format"
	} else {
		script = "#!/bin/bash\necho 'invalid-version-format'"
	}

	//nolint:gosec // G306: Test file needs execute permission for mock CLI binary
	err := os.WriteFile(mockCLI, []byte(script), 0o700)
	if err != nil {
		t.Fatalf("Failed to create mock CLI: %v", err)
	}

	_, err = DetectCLIVersion(ctx, mockCLI)
	if err == nil {
		t.Error("Expected error for invalid version format")
		return
	}

	if !strings.Contains(err.Error(), "invalid version format") {
		t.Errorf("Expected 'invalid version format' error, got: %v", err)
	}
}

// TestAddPermissionFlagsComplete tests all permission flag combinations
func TestAddPermissionFlagsComplete(t *testing.T) {
	tests := []struct {
		name    string
		options *shared.Options
		expect  map[string]string // flag -> value pairs
	}{
		{
			name: "permission_mode_only",
			options: &shared.Options{
				PermissionMode: func() *shared.PermissionMode {
					mode := shared.PermissionModeAcceptEdits
					return &mode
				}(),
			},
			expect: map[string]string{
				"--permission-mode": "acceptEdits",
			},
		},
		{
			name: "permission_prompt_tool_only",
			options: &shared.Options{
				PermissionPromptToolName: stringPtr("custom-tool"),
			},
			expect: map[string]string{
				"--permission-prompt-tool": "custom-tool",
			},
		},
		{
			name: "both_permission_flags",
			options: &shared.Options{
				PermissionMode: func() *shared.PermissionMode {
					mode := shared.PermissionModeBypassPermissions
					return &mode
				}(),
				PermissionPromptToolName: stringPtr("security-tool"),
			},
			expect: map[string]string{
				"--permission-mode":        "bypassPermissions",
				"--permission-prompt-tool": "security-tool",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := BuildCommand("/usr/local/bin/claude", test.options, false)

			for flag, expectedValue := range test.expect {
				assertContainsArgs(t, cmd, flag, expectedValue)
			}
		})
	}
}

// TestWorkingDirectoryValidationStatError tests stat error handling
func TestWorkingDirectoryValidationStatError(t *testing.T) {
	// Test with a path that will cause os.Stat to return a non-IsNotExist error
	// This is platform-dependent and hard to trigger reliably, so we test what we can

	// Test permission denied scenario (where possible)
	if runtime.GOOS != windowsOS {
		t.Run("permission_denied_directory", func(t *testing.T) {
			// Create a directory and remove permissions
			tempDir := t.TempDir()
			restrictedDir := filepath.Join(tempDir, "restricted")
			if err := os.Mkdir(restrictedDir, 0o000); err != nil {
				t.Fatalf("Failed to create restricted directory: %v", err)
			}
			defer func() {
				if err := os.Chmod(restrictedDir, 0o600); err != nil {
					t.Logf("Failed to restore directory permissions: %v", err)
				}
			}()

			// Try to validate a subdirectory of the restricted directory
			testPath := filepath.Join(restrictedDir, "subdir")
			err := ValidateWorkingDirectory(testPath)

			// Should return an error (either not exist or permission denied)
			if err == nil {
				t.Error("Expected error for inaccessible directory")
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}

// TestToolsFlagSupport tests --tools CLI flag generation for both list and preset
func TestToolsFlagSupport(t *testing.T) {
	tests := []struct {
		name     string
		options  *shared.Options
		validate func(*testing.T, []string)
	}{
		{
			name: "tools_list",
			options: &shared.Options{
				Tools: []string{"Read", "Write", "Edit"},
			},
			validate: validateToolsListFlag,
		},
		{
			name: "tools_preset",
			options: &shared.Options{
				Tools: shared.ToolsPreset{Type: "preset", Preset: "claude_code"},
			},
			validate: validateToolsPresetFlag,
		},
		{
			name: "tools_nil",
			options: &shared.Options{
				Tools: nil,
			},
			validate: validateNoToolsFlag,
		},
		{
			name: "tools_empty_list",
			options: &shared.Options{
				Tools: []string{},
			},
			validate: validateNoToolsFlag,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := BuildCommand("/usr/local/bin/claude", test.options, true)
			test.validate(t, cmd)
		})
	}
}

// validateToolsListFlag checks that --tools flag is present with comma-separated list
func validateToolsListFlag(t *testing.T, cmd []string) {
	t.Helper()
	assertContainsArgs(t, cmd, "--tools", "Read,Write,Edit")
}

// validateToolsPresetFlag checks that --tools flag contains JSON preset
func validateToolsPresetFlag(t *testing.T, cmd []string) {
	t.Helper()
	// Find the --tools flag and check its value contains the preset JSON
	for i, arg := range cmd {
		if arg == "--tools" && i+1 < len(cmd) {
			value := cmd[i+1]
			// Should contain type and preset fields
			if !strings.Contains(value, `"type":"preset"`) || !strings.Contains(value, `"preset":"claude_code"`) {
				t.Errorf("Expected --tools value to contain preset JSON, got %q", value)
			}
			return
		}
	}
	t.Error("Expected --tools flag to be present")
}

// validateNoToolsFlag checks that --tools flag is not present
func validateNoToolsFlag(t *testing.T, cmd []string) {
	t.Helper()
	assertNotContainsArg(t, cmd, "--tools")
}

// TestSessionManagementFlagsSupport tests fork_session and setting_sources CLI flags
func TestSessionManagementFlagsSupport(t *testing.T) {
	tests := []struct {
		name     string
		options  *shared.Options
		validate func(*testing.T, []string)
	}{
		{
			name:     "fork_session_enabled",
			options:  &shared.Options{ForkSession: true, SettingSources: []shared.SettingSource{}},
			validate: validateForkSessionEnabled,
		},
		{
			name:     "fork_session_disabled",
			options:  &shared.Options{ForkSession: false, SettingSources: []shared.SettingSource{}},
			validate: validateForkSessionDisabled,
		},
		{
			name:     "setting_sources_single",
			options:  &shared.Options{SettingSources: []shared.SettingSource{shared.SettingSourceUser}},
			validate: validateSettingSourcesSingle,
		},
		{
			name:     "setting_sources_multiple",
			options:  &shared.Options{SettingSources: []shared.SettingSource{shared.SettingSourceUser, shared.SettingSourceProject}},
			validate: validateSettingSourcesMultiple,
		},
		{
			name:     "setting_sources_all",
			options:  &shared.Options{SettingSources: []shared.SettingSource{shared.SettingSourceUser, shared.SettingSourceProject, shared.SettingSourceLocal}},
			validate: validateSettingSourcesAll,
		},
		{
			name:     "setting_sources_empty",
			options:  &shared.Options{SettingSources: []shared.SettingSource{}},
			validate: validateSettingSourcesEmpty,
		},
		{
			name: "fork_session_with_resume",
			options: &shared.Options{
				Resume:         stringPtr("session-123"),
				ForkSession:    true,
				SettingSources: []shared.SettingSource{shared.SettingSourceUser},
			},
			validate: validateForkSessionWithResume,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := BuildCommand("/usr/local/bin/claude", test.options, true)
			test.validate(t, cmd)
		})
	}
}

func validateForkSessionEnabled(t *testing.T, cmd []string) {
	t.Helper()
	assertContainsArg(t, cmd, "--fork-session")
}

func validateForkSessionDisabled(t *testing.T, cmd []string) {
	t.Helper()
	assertNotContainsArg(t, cmd, "--fork-session")
}

func validateSettingSourcesSingle(t *testing.T, cmd []string) {
	t.Helper()
	assertContainsArgs(t, cmd, "--setting-sources", "user")
}

func validateSettingSourcesMultiple(t *testing.T, cmd []string) {
	t.Helper()
	assertContainsArgs(t, cmd, "--setting-sources", "user,project")
}

func validateSettingSourcesAll(t *testing.T, cmd []string) {
	t.Helper()
	assertContainsArgs(t, cmd, "--setting-sources", "user,project,local")
}

func validateSettingSourcesEmpty(t *testing.T, cmd []string) {
	t.Helper()
	assertContainsArgs(t, cmd, "--setting-sources", "")
}

func validateForkSessionWithResume(t *testing.T, cmd []string) {
	t.Helper()
	assertContainsArgs(t, cmd, "--resume", "session-123")
	assertContainsArg(t, cmd, "--fork-session")
	assertContainsArgs(t, cmd, "--setting-sources", "user")
}

// TestPluginsFlagSupport tests --plugin-dir CLI flag generation
func TestPluginsFlagSupport(t *testing.T) {
	tests := []struct {
		name     string
		options  *shared.Options
		validate func(*testing.T, []string)
	}{
		{
			name: "single_local_plugin",
			options: &shared.Options{
				Plugins: []shared.SdkPluginConfig{
					{Type: shared.SdkPluginTypeLocal, Path: "/path/to/plugin"},
				},
			},
			validate: validateSinglePluginFlag,
		},
		{
			name: "multiple_local_plugins",
			options: &shared.Options{
				Plugins: []shared.SdkPluginConfig{
					{Type: shared.SdkPluginTypeLocal, Path: "/plugin1"},
					{Type: shared.SdkPluginTypeLocal, Path: "/plugin2"},
					{Type: shared.SdkPluginTypeLocal, Path: "/plugin3"},
				},
			},
			validate: validateMultiplePluginFlags,
		},
		{
			name: "empty_plugins",
			options: &shared.Options{
				Plugins: []shared.SdkPluginConfig{},
			},
			validate: validateNoPluginFlag,
		},
		{
			name: "nil_plugins",
			options: &shared.Options{
				Plugins: nil,
			},
			validate: validateNoPluginFlag,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := BuildCommand("/usr/local/bin/claude", test.options, true)
			test.validate(t, cmd)
		})
	}
}

// TestSandboxFlagSupport tests --settings flag generation for sandbox configuration
func TestSandboxFlagSupport(t *testing.T) {
	tests := []struct {
		name     string
		options  *shared.Options
		validate func(*testing.T, []string)
	}{
		{
			name: "sandbox_enabled",
			options: &shared.Options{
				Sandbox: &shared.SandboxSettings{
					Enabled:                  true,
					AutoAllowBashIfSandboxed: true,
				},
				SettingSources: []shared.SettingSource{},
			},
			validate: validateSandboxEnabled,
		},
		{
			name: "sandbox_with_excluded_commands",
			options: &shared.Options{
				Sandbox: &shared.SandboxSettings{
					Enabled:          true,
					ExcludedCommands: []string{"docker", "git"},
				},
				SettingSources: []shared.SettingSource{},
			},
			validate: validateSandboxWithExcludedCommands,
		},
		{
			name: "sandbox_with_network_config",
			options: &shared.Options{
				Sandbox: &shared.SandboxSettings{
					Enabled: true,
					Network: &shared.SandboxNetworkConfig{
						AllowUnixSockets:  []string{"/var/run/docker.sock"},
						AllowLocalBinding: true,
					},
				},
				SettingSources: []shared.SettingSource{},
			},
			validate: validateSandboxWithNetwork,
		},
		{
			name: "sandbox_nil",
			options: &shared.Options{
				Sandbox:        nil,
				SettingSources: []shared.SettingSource{},
			},
			validate: validateNoSandboxSettings,
		},
		{
			name: "sandbox_with_ignore_violations",
			options: &shared.Options{
				Sandbox: &shared.SandboxSettings{
					Enabled: true,
					IgnoreViolations: &shared.SandboxIgnoreViolations{
						File:    []string{"/tmp/*"},
						Network: []string{"localhost"},
					},
				},
				SettingSources: []shared.SettingSource{},
			},
			validate: validateSandboxWithIgnoreViolations,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := BuildCommand("/usr/local/bin/claude", test.options, true)
			test.validate(t, cmd)
		})
	}
}

// TestPluginsWithOtherFlags tests plugins work alongside other CLI flags
func TestPluginsWithOtherFlags(t *testing.T) {
	options := &shared.Options{
		Plugins: []shared.SdkPluginConfig{
			{Type: shared.SdkPluginTypeLocal, Path: "/my/plugin"},
		},
		Betas:          []shared.SdkBeta{shared.SdkBetaContext1M},
		AllowedTools:   []string{"Read", "Write"},
		SettingSources: []shared.SettingSource{},
	}

	cmd := BuildCommand("/usr/local/bin/claude", options, true)

	// Verify plugin flag is present
	assertContainsArgs(t, cmd, "--plugin-dir", "/my/plugin")

	// Verify other flags are also present
	assertContainsArgs(t, cmd, "--betas", "context-1m-2025-08-07")
	assertContainsArgs(t, cmd, "--allowed-tools", "Read,Write")
}

// TestPluginsOrderPreserved tests that plugin order is preserved in CLI flags
func TestPluginsOrderPreserved(t *testing.T) {
	options := &shared.Options{
		Plugins: []shared.SdkPluginConfig{
			{Type: shared.SdkPluginTypeLocal, Path: "/first"},
			{Type: shared.SdkPluginTypeLocal, Path: "/second"},
			{Type: shared.SdkPluginTypeLocal, Path: "/third"},
		},
		SettingSources: []shared.SettingSource{},
	}

	cmd := BuildCommand("/usr/local/bin/claude", options, true)

	// Find all --plugin-dir flags and verify order
	var pluginPaths []string
	for i, arg := range cmd {
		if arg == "--plugin-dir" && i+1 < len(cmd) {
			pluginPaths = append(pluginPaths, cmd[i+1])
		}
	}

	expected := []string{"/first", "/second", "/third"}
	if len(pluginPaths) != len(expected) {
		t.Errorf("Expected %d plugin paths, got %d", len(expected), len(pluginPaths))
		return
	}
	for i, exp := range expected {
		if pluginPaths[i] != exp {
			t.Errorf("Expected plugin path[%d] = %q, got %q", i, exp, pluginPaths[i])
		}
	}
}

func validateSinglePluginFlag(t *testing.T, cmd []string) {
	t.Helper()
	assertContainsArgs(t, cmd, "--plugin-dir", "/path/to/plugin")
}

func validateMultiplePluginFlags(t *testing.T, cmd []string) {
	t.Helper()
	// Each plugin should generate a separate --plugin-dir flag
	assertContainsArgs(t, cmd, "--plugin-dir", "/plugin1")
	assertContainsArgs(t, cmd, "--plugin-dir", "/plugin2")
	assertContainsArgs(t, cmd, "--plugin-dir", "/plugin3")
}

func validateNoPluginFlag(t *testing.T, cmd []string) {
	t.Helper()
	assertNotContainsArg(t, cmd, "--plugin-dir")
}

const settingsFlag = "--settings"

// validateSandboxEnabled checks that --settings flag contains sandbox enabled config
func validateSandboxEnabled(t *testing.T, cmd []string) {
	t.Helper()
	// Find the --settings flag and check its value contains sandbox config
	for i, arg := range cmd {
		if arg == settingsFlag && i+1 < len(cmd) {
			value := cmd[i+1]
			// Should contain sandbox enabled fields
			if !strings.Contains(value, `"sandbox"`) {
				t.Errorf("Expected --settings value to contain sandbox config, got %q", value)
			}
			if !strings.Contains(value, `"enabled":true`) {
				t.Errorf("Expected --settings value to contain enabled:true, got %q", value)
			}
			return
		}
	}
	t.Error("Expected --settings flag to be present for sandbox configuration")
}

// validateSandboxWithExcludedCommands checks sandbox with excluded commands
func validateSandboxWithExcludedCommands(t *testing.T, cmd []string) {
	t.Helper()
	for i, arg := range cmd {
		if arg == settingsFlag && i+1 < len(cmd) {
			value := cmd[i+1]
			if !strings.Contains(value, `"excludedCommands"`) {
				t.Errorf("Expected --settings value to contain excludedCommands, got %q", value)
			}
			if !strings.Contains(value, `"docker"`) {
				t.Errorf("Expected --settings value to contain docker command, got %q", value)
			}
			return
		}
	}
	t.Error("Expected --settings flag to be present")
}

// validateSandboxWithNetwork checks sandbox with network configuration
func validateSandboxWithNetwork(t *testing.T, cmd []string) {
	t.Helper()
	for i, arg := range cmd {
		if arg == settingsFlag && i+1 < len(cmd) {
			value := cmd[i+1]
			if !strings.Contains(value, `"network"`) {
				t.Errorf("Expected --settings value to contain network config, got %q", value)
			}
			if !strings.Contains(value, `"allowLocalBinding":true`) {
				t.Errorf("Expected --settings value to contain allowLocalBinding:true, got %q", value)
			}
			return
		}
	}
	t.Error("Expected --settings flag to be present")
}

// validateNoSandboxSettings checks that no sandbox-related --settings is added
func validateNoSandboxSettings(t *testing.T, cmd []string) {
	t.Helper()
	// When sandbox is nil, we should not add a --settings flag for sandbox
	// (unless there's an existing Settings value)
	for i, arg := range cmd {
		if arg == settingsFlag && i+1 < len(cmd) {
			value := cmd[i+1]
			if strings.Contains(value, `"sandbox"`) {
				t.Errorf("Expected no sandbox in --settings when Sandbox is nil, got %q", value)
			}
		}
	}
}

// validateSandboxWithIgnoreViolations checks sandbox with ignore violations
func validateSandboxWithIgnoreViolations(t *testing.T, cmd []string) {
	t.Helper()
	for i, arg := range cmd {
		if arg == settingsFlag && i+1 < len(cmd) {
			value := cmd[i+1]
			if !strings.Contains(value, `"ignoreViolations"`) {
				t.Errorf("Expected --settings value to contain ignoreViolations, got %q", value)
			}
			return
		}
	}
	t.Error("Expected --settings flag to be present")
}

// TestSandboxWithExistingSettings tests sandbox merging with existing settings
func TestSandboxWithExistingSettings(t *testing.T) {
	existingSettings := `{"model":"claude-3-sonnet"}`
	options := &shared.Options{
		Settings: &existingSettings,
		Sandbox: &shared.SandboxSettings{
			Enabled: true,
		},
		SettingSources: []shared.SettingSource{},
	}

	cmd := BuildCommand("/usr/local/bin/claude", options, true)

	// Count --settings flags - must be exactly 1
	settingsCount := 0
	var settingsValue string
	for i, arg := range cmd {
		if arg == settingsFlag && i+1 < len(cmd) {
			settingsCount++
			settingsValue = cmd[i+1]
		}
	}

	if settingsCount != 1 {
		t.Errorf("Expected exactly 1 --settings flag, got %d", settingsCount)
	}

	// MUST contain BOTH sandbox AND model in merged JSON
	if !strings.Contains(settingsValue, `"sandbox"`) {
		t.Errorf("Expected --settings to contain 'sandbox', got %q", settingsValue)
	}
	if !strings.Contains(settingsValue, `"model"`) {
		t.Errorf("Expected --settings to contain 'model', got %q", settingsValue)
	}
	if !strings.Contains(settingsValue, `"enabled":true`) {
		t.Errorf("Expected --settings to contain sandbox enabled:true, got %q", settingsValue)
	}
}

// TestOutputFormatFlagSupport tests --json-schema CLI flag generation for structured output
func TestOutputFormatFlagSupport(t *testing.T) {
	tests := []struct {
		name     string
		options  *shared.Options
		validate func(*testing.T, []string)
	}{
		{
			name: "json_schema_flag_present",
			options: &shared.Options{
				OutputFormat: &shared.OutputFormat{
					Type: "json_schema",
					Schema: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"answer": map[string]any{"type": "string"},
						},
					},
				},
			},
			validate: validateJSONSchemaFlagPresent,
		},
		{
			name: "no_flag_when_nil",
			options: &shared.Options{
				OutputFormat: nil,
			},
			validate: validateNoJSONSchemaFlag,
		},
		{
			name: "no_flag_when_nil_schema",
			options: &shared.Options{
				OutputFormat: &shared.OutputFormat{
					Type:   "json_schema",
					Schema: nil,
				},
			},
			validate: validateNoJSONSchemaFlag,
		},
		{
			name: "complex_nested_schema",
			options: &shared.Options{
				OutputFormat: &shared.OutputFormat{
					Type: "json_schema",
					Schema: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"items": map[string]any{
								"type": "array",
								"items": map[string]any{
									"type": "object",
									"properties": map[string]any{
										"name":  map[string]any{"type": "string"},
										"value": map[string]any{"type": "number"},
									},
								},
							},
						},
						"required": []string{"items"},
					},
				},
			},
			validate: validateJSONSchemaFlagPresent,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := BuildCommand("/usr/local/bin/claude", test.options, true)
			test.validate(t, cmd)
		})
	}
}

// TestOutputFormatFlagWithOtherOptions tests --json-schema flag works with other options
func TestOutputFormatFlagWithOtherOptions(t *testing.T) {
	systemPrompt := "You are helpful"
	permissionMode := shared.PermissionModeAcceptEdits

	options := &shared.Options{
		SystemPrompt:   &systemPrompt,
		PermissionMode: &permissionMode,
		OutputFormat: &shared.OutputFormat{
			Type: "json_schema",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"result": map[string]any{"type": "string"},
				},
			},
		},
	}

	cmd := BuildCommand("/usr/local/bin/claude", options, true)

	// Verify all flags are present
	assertContainsArgs(t, cmd, "--system-prompt", "You are helpful")
	assertContainsArgs(t, cmd, "--permission-mode", "acceptEdits")
	validateJSONSchemaFlagPresent(t, cmd)
}

// validateJSONSchemaFlagPresent checks that --json-schema flag is present with valid JSON
func validateJSONSchemaFlagPresent(t *testing.T, cmd []string) {
	t.Helper()
	for i, arg := range cmd {
		if arg == "--json-schema" && i+1 < len(cmd) {
			// Verify it's valid JSON
			var schema map[string]any
			if err := json.Unmarshal([]byte(cmd[i+1]), &schema); err != nil {
				t.Errorf("Expected valid JSON for --json-schema, got error: %v", err)
			}
			return
		}
	}
	t.Error("Expected --json-schema flag to be present")
}

// validateNoJSONSchemaFlag checks that --json-schema flag is not present
func validateNoJSONSchemaFlag(t *testing.T, cmd []string) {
	t.Helper()
	assertNotContainsArg(t, cmd, "--json-schema")
}

const agentsFlag = "--agents"

// TestAgentsFlagSupport tests --agents CLI flag generation
func TestAgentsFlagSupport(t *testing.T) {
	tests := []struct {
		name     string
		options  *shared.Options
		validate func(*testing.T, []string)
	}{
		{
			name: "single_agent",
			options: &shared.Options{
				Agents: map[string]shared.AgentDefinition{
					"code-reviewer": {
						Description: "Reviews code",
						Prompt:      "You are a reviewer...",
						Tools:       []string{"Read", "Grep"},
						Model:       shared.AgentModelSonnet,
					},
				},
			},
			validate: validateSingleAgentFlag,
		},
		{
			name: "multiple_agents",
			options: &shared.Options{
				Agents: map[string]shared.AgentDefinition{
					"reviewer": {
						Description: "Reviews",
						Prompt:      "Reviewer prompt",
					},
					"tester": {
						Description: "Tests",
						Prompt:      "Tester prompt",
					},
				},
			},
			validate: validateMultipleAgentsFlag,
		},
		{
			name: "omit_nil_fields",
			options: &shared.Options{
				Agents: map[string]shared.AgentDefinition{
					"minimal": {
						Description: "Minimal agent",
						Prompt:      "Minimal prompt",
						// Tools and Model are empty/nil
					},
				},
			},
			validate: validateMinimalAgentFlag,
		},
		{
			name: "empty_agents",
			options: &shared.Options{
				Agents: map[string]shared.AgentDefinition{},
			},
			validate: validateNoAgentsFlag,
		},
		{
			name: "nil_agents",
			options: &shared.Options{
				Agents: nil,
			},
			validate: validateNoAgentsFlag,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := BuildCommand("/usr/local/bin/claude", test.options, true)
			test.validate(t, cmd)
		})
	}
}

func validateSingleAgentFlag(t *testing.T, cmd []string) {
	t.Helper()
	// Find the --agents flag and verify JSON content
	for i, arg := range cmd {
		if arg == agentsFlag && i+1 < len(cmd) {
			value := cmd[i+1]
			// Should contain the agent definition with all fields
			if !strings.Contains(value, `"code-reviewer"`) {
				t.Errorf("Expected --agents value to contain code-reviewer, got %q", value)
			}
			if !strings.Contains(value, `"description":"Reviews code"`) {
				t.Errorf("Expected --agents value to contain description, got %q", value)
			}
			if !strings.Contains(value, `"prompt":"You are a reviewer..."`) {
				t.Errorf("Expected --agents value to contain prompt, got %q", value)
			}
			if !strings.Contains(value, `"tools"`) {
				t.Errorf("Expected --agents value to contain tools, got %q", value)
			}
			if !strings.Contains(value, `"model":"sonnet"`) {
				t.Errorf("Expected --agents value to contain model, got %q", value)
			}
			return
		}
	}
	t.Error("Expected --agents flag to be present")
}

func validateMultipleAgentsFlag(t *testing.T, cmd []string) {
	t.Helper()
	for i, arg := range cmd {
		if arg == agentsFlag && i+1 < len(cmd) {
			value := cmd[i+1]
			if !strings.Contains(value, `"reviewer"`) {
				t.Errorf("Expected --agents value to contain reviewer, got %q", value)
			}
			if !strings.Contains(value, `"tester"`) {
				t.Errorf("Expected --agents value to contain tester, got %q", value)
			}
			return
		}
	}
	t.Error("Expected --agents flag to be present")
}

func validateMinimalAgentFlag(t *testing.T, cmd []string) {
	t.Helper()
	for i, arg := range cmd {
		if arg == agentsFlag && i+1 < len(cmd) {
			value := cmd[i+1]
			// Should contain description and prompt
			if !strings.Contains(value, `"description":"Minimal agent"`) {
				t.Errorf("Expected --agents value to contain description, got %q", value)
			}
			if !strings.Contains(value, `"prompt":"Minimal prompt"`) {
				t.Errorf("Expected --agents value to contain prompt, got %q", value)
			}
			// Should NOT contain tools or model (they're empty)
			if strings.Contains(value, `"tools"`) {
				t.Errorf("Expected --agents value to NOT contain empty tools, got %q", value)
			}
			if strings.Contains(value, `"model"`) {
				t.Errorf("Expected --agents value to NOT contain empty model, got %q", value)
			}
			return
		}
	}
	t.Error("Expected --agents flag to be present")
}

func validateNoAgentsFlag(t *testing.T, cmd []string) {
	t.Helper()
	assertNotContainsArg(t, cmd, agentsFlag)
}

// TestIncludePartialMessagesFlagSupport tests CLI flag for partial message streaming
func TestIncludePartialMessagesFlagSupport(t *testing.T) {
	tests := []struct {
		name     string
		options  *shared.Options
		validate func(*testing.T, []string)
	}{
		{
			name: "flag_added_when_true",
			options: &shared.Options{
				IncludePartialMessages: true,
			},
			validate: func(t *testing.T, cmd []string) {
				t.Helper()
				assertContainsArg(t, cmd, "--include-partial-messages")
			},
		},
		{
			name: "flag_not_added_when_false",
			options: &shared.Options{
				IncludePartialMessages: false,
			},
			validate: func(t *testing.T, cmd []string) {
				t.Helper()
				assertNotContainsArg(t, cmd, "--include-partial-messages")
			},
		},
		{
			name:    "flag_not_added_by_default",
			options: &shared.Options{},
			validate: func(t *testing.T, cmd []string) {
				t.Helper()
				assertNotContainsArg(t, cmd, "--include-partial-messages")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := BuildCommand("/usr/local/bin/claude", test.options, false)
			test.validate(t, cmd)
		})
	}
}

// TestIncludePartialMessagesWithOtherOptions tests flag interaction with other options
func TestIncludePartialMessagesWithOtherOptions(t *testing.T) {
	options := &shared.Options{
		IncludePartialMessages: true,
		MaxTurns:               5,
		ContinueConversation:   true,
	}

	cmd := BuildCommand("/usr/local/bin/claude", options, false)

	// Verify all flags are present
	assertContainsArg(t, cmd, "--include-partial-messages")
	assertContainsArgs(t, cmd, "--max-turns", "5")
	assertContainsArg(t, cmd, "--continue")
}
