package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/severity1/claude-code-sdk-go/internal/shared"
)

// T083: CLI Not Found Error 🔴 RED
func TestCLINotFoundError(t *testing.T) {
	// Test that CLI binary not found returns helpful error with Node.js dependency check

	// Create temporary home directory for testing
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	originalPath := os.Getenv("PATH")

	// Set environment to isolated temp directory to avoid finding real Claude CLI
	if runtime.GOOS == "windows" {
		originalHome = os.Getenv("USERPROFILE")
		os.Setenv("USERPROFILE", tempHome)
	} else {
		os.Setenv("HOME", tempHome)
	}
	os.Setenv("PATH", "/nonexistent/path") // Ensure claude is not found in PATH

	defer func() {
		if runtime.GOOS == "windows" {
			os.Setenv("USERPROFILE", originalHome)
		} else {
			os.Setenv("HOME", originalHome)
		}
		os.Setenv("PATH", originalPath)
	}()

	// When no Claude CLI exists in any location, should get CLINotFoundError
	_, err := FindCLI()

	// Should get an error (either CLI not found or Node.js not found)
	if err == nil {
		t.Error("Expected error when Claude CLI is not found, got nil")
		return
	}

	// Error should be helpful and mention installation instructions
	errMsg := err.Error()
	if !strings.Contains(errMsg, "Claude Code") && !strings.Contains(errMsg, "Node.js") {
		t.Error("Error message should mention Claude Code or Node.js")
	}
	if !strings.Contains(errMsg, "install") {
		t.Error("Error message should include installation guidance")
	}
}

// T084: Build Basic Command 🔴 RED
func TestBuildBasicCommand(t *testing.T) {
	// Build basic CLI command with required flags

	cliPath := "/usr/local/bin/claude"
	options := &shared.Options{}

	// Test one-shot mode (closeStdin = true)
	cmd := BuildCommand(cliPath, options, true)

	// Should include basic required flags
	if cmd[0] != cliPath {
		t.Errorf("Expected CLI path %s, got %s", cliPath, cmd[0])
	}

	// Must include --output-format stream-json --verbose
	if !containsArgs(cmd, "--output-format", "stream-json") {
		t.Error("Command should include --output-format stream-json")
	}
	if !containsArg(cmd, "--verbose") {
		t.Error("Command should include --verbose")
	}

	// One-shot mode should include --print
	if !containsArg(cmd, "--print") {
		t.Error("One-shot mode should include --print flag")
	}

	// Test streaming mode (closeStdin = false)
	streamCmd := BuildCommand(cliPath, options, false)

	// Streaming mode should include --input-format stream-json
	if !containsArgs(streamCmd, "--input-format", "stream-json") {
		t.Error("Streaming mode should include --input-format stream-json")
	}
	if containsArg(streamCmd, "--print") {
		t.Error("Streaming mode should not include --print flag")
	}
}

// T085: CLI Path Accepts Path 🔴 RED
func TestCLIPathAcceptsPath(t *testing.T) {
	// Accept both string and path types for CLI path

	// Test with absolute path
	absPath := "/usr/local/bin/claude"
	cmd := BuildCommand(absPath, &shared.Options{}, true)
	if cmd[0] != absPath {
		t.Errorf("Expected absolute path %s, got %s", absPath, cmd[0])
	}

	// Test with relative path
	relPath := "./claude"
	cmd = BuildCommand(relPath, &shared.Options{}, true)
	if cmd[0] != relPath {
		t.Errorf("Expected relative path %s, got %s", relPath, cmd[0])
	}

	// Test with messy path (should preserve original path as given)
	messyPath := "/usr/local/bin/../bin/./claude"
	cmd = BuildCommand(messyPath, &shared.Options{}, true)
	if cmd[0] != messyPath {
		t.Errorf("Expected messy path %s to be preserved, got %s", messyPath, cmd[0])
	}
}

// T086-T091: CLI Discovery Location Tests 🔴 RED
// These tests verify that the discovery functions check the correct locations
func TestCLIDiscoveryLocations(t *testing.T) {
	// Test that getCommonCLILocations returns expected paths
	locations := getCommonCLILocations()

	if len(locations) == 0 {
		t.Fatal("Expected at least one CLI location, got none")
	}

	// Should include npm-global location
	homeDir, _ := os.UserHomeDir()
	expectedNpmGlobal := filepath.Join(homeDir, ".npm-global", "bin", "claude")
	if runtime.GOOS == "windows" {
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

	// Should include system-wide location (except on Windows)
	if runtime.GOOS != "windows" {
		systemWide := "/usr/local/bin/claude"
		found = false
		for _, location := range locations {
			if location == systemWide {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected system-wide location %s in discovery paths", systemWide)
		}
	}

	// Should include user local location
	expectedUserLocal := filepath.Join(homeDir, ".local", "bin", "claude")
	if runtime.GOOS == "windows" {
		// Windows doesn't typically use ~/.local/bin
		return
	}

	found = false
	for _, location := range locations {
		if location == expectedUserLocal {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected user local location %s in discovery paths", expectedUserLocal)
	}
}

// T092: Node.js Dependency Validation 🔴 RED
func TestNodeJSDependencyValidation(t *testing.T) {
	// Test Node.js validation (this will depend on the actual system)
	err := ValidateNodeJS()

	// If error occurs, it should be helpful
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

// T093: Command Building All Options 🔴 RED
func TestCommandBuildingAllOptions(t *testing.T) {
	// Build command with all configuration options

	cliPath := "/usr/local/bin/claude"
	systemPrompt := "You are a helpful assistant"
	appendPrompt := "Additional context"
	model := "claude-3-sonnet"
	permissionMode := shared.PermissionModeAcceptEdits
	resume := "session123"
	settings := "/path/to/settings.json"
	cwd := "/workspace"

	options := &shared.Options{
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
		ExtraArgs:            map[string]*string{"custom-flag": nil, "with-value": &[]string{"test"}[0]},
	}

	cmd := BuildCommand(cliPath, options, false)

	// Verify all options are included as CLI flags
	if !containsArgs(cmd, "--allowed-tools", "Read,Write") {
		t.Error("Should include --allowed-tools Read,Write")
	}
	if !containsArgs(cmd, "--disallowed-tools", "Bash,Delete") {
		t.Error("Should include --disallowed-tools Bash,Delete")
	}
	if !containsArgs(cmd, "--system-prompt", systemPrompt) {
		t.Error("Should include --system-prompt")
	}
	if !containsArgs(cmd, "--append-system-prompt", appendPrompt) {
		t.Error("Should include --append-system-prompt")
	}
	if !containsArgs(cmd, "--model", model) {
		t.Error("Should include --model")
	}
	if !containsArgs(cmd, "--max-thinking-tokens", "10000") {
		t.Error("Should include --max-thinking-tokens 10000")
	}
	if !containsArgs(cmd, "--permission-mode", "acceptEdits") {
		t.Error("Should include --permission-mode acceptEdits")
	}
	if !containsArg(cmd, "--continue") {
		t.Error("Should include --continue flag")
	}
	if !containsArgs(cmd, "--resume", resume) {
		t.Error("Should include --resume")
	}
	if !containsArgs(cmd, "--max-turns", "25") {
		t.Error("Should include --max-turns 25")
	}
	if !containsArgs(cmd, "--settings", settings) {
		t.Error("Should include --settings")
	}
	if !containsArgs(cmd, "--cwd", cwd) {
		t.Error("Should include --cwd")
	}
	if !containsArgs(cmd, "--add-dir", "/extra/dir1") {
		t.Error("Should include --add-dir for first directory")
	}
	if !containsArgs(cmd, "--add-dir", "/extra/dir2") {
		t.Error("Should include --add-dir for second directory")
	}

	// Custom args
	if !containsArg(cmd, "--custom-flag") {
		t.Error("Should include --custom-flag (boolean)")
	}
	if !containsArgs(cmd, "--with-value", "test") {
		t.Error("Should include --with-value test")
	}
}

// T094: ExtraArgs Support 🔴 RED
func TestExtraArgsSupport(t *testing.T) {
	// Support arbitrary CLI flags via ExtraArgs

	cliPath := "/usr/local/bin/claude"

	// Test boolean flags (nil value)
	options := &shared.Options{
		ExtraArgs: map[string]*string{
			"debug":    nil,
			"trace":    nil,
			"no-cache": nil,
		},
	}

	cmd := BuildCommand(cliPath, options, true)

	if !containsArg(cmd, "--debug") {
		t.Error("Should include boolean flag --debug")
	}
	if !containsArg(cmd, "--trace") {
		t.Error("Should include boolean flag --trace")
	}
	if !containsArg(cmd, "--no-cache") {
		t.Error("Should include boolean flag --no-cache")
	}

	// Test flags with values
	value1 := "info"
	value2 := "/tmp/custom"
	options = &shared.Options{
		ExtraArgs: map[string]*string{
			"log-level":  &value1,
			"custom-dir": &value2,
		},
	}

	cmd = BuildCommand(cliPath, options, true)

	if !containsArgs(cmd, "--log-level", "info") {
		t.Error("Should include --log-level info")
	}
	if !containsArgs(cmd, "--custom-dir", "/tmp/custom") {
		t.Error("Should include --custom-dir /tmp/custom")
	}
}

// T095: Close Stdin Flag Handling 🔴 RED
func TestCloseStdinFlagHandling(t *testing.T) {
	// Handle --print vs --input-format based on closeStdin

	cliPath := "/usr/local/bin/claude"
	options := &shared.Options{}

	// Test one-shot mode (closeStdin = true) should use --print
	oneShot := BuildCommand(cliPath, options, true)
	if !containsArg(oneShot, "--print") {
		t.Error("One-shot mode should include --print flag")
	}
	if containsArgs(oneShot, "--input-format", "stream-json") {
		t.Error("One-shot mode should not include --input-format stream-json")
	}

	// Test streaming mode (closeStdin = false) should use --input-format
	streaming := BuildCommand(cliPath, options, false)
	if containsArg(streaming, "--print") {
		t.Error("Streaming mode should not include --print flag")
	}
	if !containsArgs(streaming, "--input-format", "stream-json") {
		t.Error("Streaming mode should include --input-format stream-json")
	}

	// Both modes should include --output-format stream-json
	if !containsArgs(oneShot, "--output-format", "stream-json") {
		t.Error("One-shot mode should include --output-format stream-json")
	}
	if !containsArgs(streaming, "--output-format", "stream-json") {
		t.Error("Streaming mode should include --output-format stream-json")
	}
}

// T096: Working Directory Validation 🔴 RED
func TestWorkingDirectoryValidation(t *testing.T) {
	// Validate working directory exists

	// Test with existing directory
	tempDir := t.TempDir()
	err := ValidateWorkingDirectory(tempDir)
	if err != nil {
		t.Errorf("Expected no error for existing directory, got: %v", err)
	}

	// Test with empty string (should be valid - uses current directory)
	err = ValidateWorkingDirectory("")
	if err != nil {
		t.Errorf("Expected no error for empty directory, got: %v", err)
	}

	// Test with non-existent directory
	nonExistent := filepath.Join(tempDir, "does-not-exist")
	err = ValidateWorkingDirectory(nonExistent)
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}

	// Test with file instead of directory
	tempFile := filepath.Join(tempDir, "testfile")
	os.WriteFile(tempFile, []byte("test"), 0644)
	err = ValidateWorkingDirectory(tempFile)
	if err == nil {
		t.Error("Expected error when path is a file, not directory")
	}

	// Error messages should be helpful
	if err != nil && !strings.Contains(err.Error(), "not a directory") {
		t.Error("Error message should indicate path is not a directory")
	}
}

// T097: CLI Version Detection 🔴 RED
func TestCLIVersionDetection(t *testing.T) {
	// Detect Claude CLI version for compatibility

	// This test will only pass if Claude CLI is actually installed
	// For now, we'll test the function exists and handles errors appropriately

	nonExistentPath := "/this/path/does/not/exist/claude"
	_, err := DetectCLIVersion(nonExistentPath)
	if err == nil {
		t.Error("Expected error when CLI path does not exist")
	}

	// Error should mention version detection failure
	if err != nil && !strings.Contains(err.Error(), "version") {
		t.Error("Error message should mention version detection failure")
	}
}

// Helper functions for testing
func containsArg(args []string, target string) bool {
	for _, arg := range args {
		if arg == target {
			return true
		}
	}
	return false
}

func containsArgs(args []string, flag, value string) bool {
	for i, arg := range args {
		if arg == flag && i+1 < len(args) && args[i+1] == value {
			return true
		}
	}
	return false
}
