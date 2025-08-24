package claudecode

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
)

// TestBaseSDKError tests the base ClaudeSDKError interface and implementation.
// Python Reference: test_errors.py::TestErrorTypes::test_base_error
func TestBaseSDKError(t *testing.T) {
	// Test that we can create a base SDK error
	message := "Something went wrong"
	var err error = &BaseError{message: message}

	// Must implement error interface
	if err.Error() != message {
		t.Errorf("Expected error message %q, got %q", message, err.Error())
	}

	// Must implement SDKError interface
	sdkErr, ok := err.(SDKError)
	if !ok {
		t.Fatal("BaseError must implement SDKError interface")
	}

	// Type method should return a type identifier
	errType := sdkErr.Type()
	if errType == "" {
		t.Error("SDKError.Type() should return non-empty string")
	}

	// Should be an instance of error (standard interface) - already validated above
}

// TestCLINotFoundError tests CLINotFoundError with helpful installation message.
// Python Reference: test_errors.py::TestErrorTypes::test_cli_not_found_error
func TestCLINotFoundError(t *testing.T) {
	// Test basic CLI not found error
	message := "Claude Code not found"
	err := NewCLINotFoundError("", message)

	// Must implement SDKError interface
	if _, ok := interface{}(err).(SDKError); !ok {
		t.Fatal("CLINotFoundError must implement SDKError interface")
	}

	// Must contain expected message
	errMsg := err.Error()
	if !strings.Contains(errMsg, "Claude Code not found") {
		t.Errorf("Expected error to contain 'Claude Code not found', got %q", errMsg)
	}

	// Test with path specification
	path := "/path/to/claude"
	errWithPath := NewCLINotFoundError(path, message)
	errMsgWithPath := errWithPath.Error()
	if !strings.Contains(errMsgWithPath, path) {
		t.Errorf("Expected error to contain path %q, got %q", path, errMsgWithPath)
	}

	// Should have correct type
	if errWithPath.Type() != "cli_not_found_error" {
		t.Errorf("Expected type 'cli_not_found_error', got %q", errWithPath.Type())
	}
}

// TestConnectionError tests CLIConnectionError for connection failures.
// Python Reference: test_errors.py::TestErrorTypes::test_connection_error
func TestConnectionError(t *testing.T) {
	// Test basic connection error
	message := "Failed to connect to CLI"
	err := NewConnectionError(message, nil)

	// Must implement SDKError interface
	if _, ok := interface{}(err).(SDKError); !ok {
		t.Fatal("ConnectionError must implement SDKError interface")
	}

	// Must inherit from ClaudeSDKError (compatible)
	if _, ok := interface{}(err).(SDKError); !ok {
		t.Fatal("ConnectionError must inherit from base SDK error")
	}

	// Must contain expected message
	errMsg := err.Error()
	if !strings.Contains(errMsg, "Failed to connect to CLI") {
		t.Errorf("Expected error to contain 'Failed to connect to CLI', got %q", errMsg)
	}

	// Should have correct type
	if err.Type() != "connection_error" {
		t.Errorf("Expected type 'connection_error', got %q", err.Type())
	}
}

// TestProcessErrorWithDetails tests ProcessError with exit code and stderr.
// Python Reference: test_errors.py::TestErrorTypes::test_process_error
func TestProcessErrorWithDetails(t *testing.T) {
	// Test process error with exit code and stderr
	message := "Process failed"
	exitCode := 1
	stderr := "Command not found"
	err := NewProcessError(message, exitCode, stderr)

	// Must implement SDKError interface
	if _, ok := interface{}(err).(SDKError); !ok {
		t.Fatal("ProcessError must implement SDKError interface")
	}

	// Test field access
	if err.ExitCode != exitCode {
		t.Errorf("Expected exit code %d, got %d", exitCode, err.ExitCode)
	}

	if err.Stderr != stderr {
		t.Errorf("Expected stderr %q, got %q", stderr, err.Stderr)
	}

	// Test error message format - should match Python exactly
	errMsg := err.Error()
	if !strings.Contains(errMsg, "Process failed") {
		t.Errorf("Expected error to contain 'Process failed', got %q", errMsg)
	}

	if !strings.Contains(errMsg, "exit code: 1") {
		t.Errorf("Expected error to contain 'exit code: 1', got %q", errMsg)
	}

	if !strings.Contains(errMsg, "Command not found") {
		t.Errorf("Expected error to contain 'Command not found', got %q", errMsg)
	}

	// Should have correct type
	if err.Type() != "process_error" {
		t.Errorf("Expected type 'process_error', got %q", err.Type())
	}
}

// TestJSONDecodeError tests CLIJSONDecodeError.
// Python Reference: test_errors.py::TestErrorTypes::test_json_decode_error
func TestJSONDecodeError(t *testing.T) {
	// Test JSON decode error with original error
	invalidJSON := "{invalid json}"
	var originalError error

	// Create an actual JSON decode error like Python test
	err := json.Unmarshal([]byte(invalidJSON), &map[string]interface{}{})
	originalError = err

	jsonErr := NewJSONDecodeError(invalidJSON, 0, originalError)

	// Must implement SDKError interface
	if _, ok := interface{}(jsonErr).(SDKError); !ok {
		t.Fatal("JSONDecodeError must implement SDKError interface")
	}

	// Test field access
	if jsonErr.Line != invalidJSON {
		t.Errorf("Expected line %q, got %q", invalidJSON, jsonErr.Line)
	}

	if jsonErr.OriginalError != originalError {
		t.Errorf("Expected original error to be preserved")
	}
	
	// Test that Unwrap() also works
	if jsonErr.Unwrap() != originalError {
		t.Errorf("Expected Unwrap() to return original error")
	}

	// Test error message format - should match Python exactly
	errMsg := jsonErr.Error()
	if !strings.Contains(errMsg, "Failed to decode JSON") {
		t.Errorf("Expected error to contain 'Failed to decode JSON', got %q", errMsg)
	}

	// Should have correct type
	if jsonErr.Type() != "json_decode_error" {
		t.Errorf("Expected type 'json_decode_error', got %q", jsonErr.Type())
	}
}

// TestMessageParseError tests MessageParseError with raw data context preservation.
// Python Reference: test_errors.py (implied from MessageParseError usage)
func TestMessageParseError(t *testing.T) {
	// Test message parse error with raw data
	message := "Unable to parse message"
	data := map[string]interface{}{
		"type":          "unknown",
		"invalid_field": "value",
	}
	err := NewMessageParseError(message, data)

	// Must implement SDKError interface
	if _, ok := interface{}(err).(SDKError); !ok {
		t.Fatal("MessageParseError must implement SDKError interface")
	}

	// Test field access - data should be preserved for debugging
	if err.Data == nil {
		t.Errorf("Expected data to be preserved, got nil")
	}

	// Check if data contains expected fields
	if dataMap, ok := err.Data.(map[string]interface{}); ok {
		if dataMap["type"] != "unknown" {
			t.Errorf("Expected data to contain type 'unknown'")
		}
	} else {
		t.Errorf("Expected data to be a map[string]interface{}")
	}

	// Test error message
	errMsg := err.Error()
	if !strings.Contains(errMsg, message) {
		t.Errorf("Expected error to contain %q, got %q", message, errMsg)
	}

	// Should have correct type
	if err.Type() != "message_parse_error" {
		t.Errorf("Expected type 'message_parse_error', got %q", err.Type())
	}
}

// TestErrorHierarchy verifies all errors implement SDKError interface.
func TestErrorHierarchy(t *testing.T) {
	// Test all error types implement SDKError interface
	testCases := []struct {
		name  string
		error SDKError
	}{
		{"BaseError", &BaseError{message: "test"}},
		{"ConnectionError", NewConnectionError("test", nil)},
		{"CLINotFoundError", NewCLINotFoundError("", "test")},
		{"ProcessError", NewProcessError("test", 1, "stderr")},
		{"JSONDecodeError", NewJSONDecodeError("test", 0, nil)},
		{"MessageParseError", NewMessageParseError("test", nil)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Must implement both error and SDKError interfaces
			if _, ok := tc.error.(error); !ok {
				t.Errorf("%s must implement error interface", tc.name)
			}

			if _, ok := tc.error.(SDKError); !ok {
				t.Errorf("%s must implement SDKError interface", tc.name)
			}

			// Must have non-empty type
			if tc.error.Type() == "" {
				t.Errorf("%s Type() should return non-empty string", tc.name)
			}

			// Error() must return non-empty string
			if tc.error.Error() == "" {
				t.Errorf("%s Error() should return non-empty string", tc.name)
			}
		})
	}
}

// TestErrorContextPreservation tests error wrapping with fmt.Errorf %w verb.
func TestErrorContextPreservation(t *testing.T) {
	// Test error wrapping and unwrapping with Go error patterns
	originalErr := errors.New("original error")

	// Test ConnectionError wrapping
	connErr := NewConnectionError("connection failed", originalErr)

	// Should be able to unwrap to get original error
	if !errors.Is(connErr, originalErr) {
		t.Errorf("ConnectionError should wrap original error, errors.Is check failed")
	}

	var unwrapped error = errors.Unwrap(connErr)
	if unwrapped != originalErr {
		t.Errorf("Expected unwrapped error to be original, got %v", unwrapped)
	}

	// Test JSONDecodeError wrapping
	jsonErr := NewJSONDecodeError("invalid json", 0, originalErr)
	if !errors.Is(jsonErr, originalErr) {
		t.Errorf("JSONDecodeError should wrap original error, errors.Is check failed")
	}

	// Test fmt.Errorf style wrapping with SDK errors
	wrappedSDK := fmt.Errorf("higher level error: %w", connErr)

	// Should be able to extract ConnectionError with errors.As
	var extractedConn *ConnectionError
	if !errors.As(wrappedSDK, &extractedConn) {
		t.Errorf("Should be able to extract ConnectionError with errors.As")
	}

	// And should still be able to check for original error
	if !errors.Is(wrappedSDK, originalErr) {
		t.Errorf("Should be able to check for original error through chain")
	}
}

// TestErrorMessageFormatting tests helpful error messages with contextual guidance.
func TestErrorMessageFormatting(t *testing.T) {
	// Test CLINotFoundError with helpful installation guidance
	cliErr := NewCLINotFoundError("", `Claude Code not found. Install with:
  npm install -g @anthropic-ai/claude-code

If already installed locally, try:
  export PATH="$HOME/node_modules/.bin:$PATH"

Or specify the path when creating client`)

	msg := cliErr.Error()

	// Should provide actionable guidance
	if !strings.Contains(msg, "npm install") {
		t.Errorf("CLI not found error should suggest npm install")
	}

	if !strings.Contains(msg, "PATH") {
		t.Errorf("CLI not found error should suggest PATH export")
	}

	// Test ProcessError with exit code context
	procErr := NewProcessError("Command execution failed", 127, "command not found")
	procMsg := procErr.Error()

	// Should provide specific exit code and stderr context
	if !strings.Contains(procMsg, "exit code: 127") {
		t.Errorf("Process error should include exit code")
	}

	if !strings.Contains(procMsg, "Error output: command not found") {
		t.Errorf("Process error should include stderr output with prefix")
	}

	// Test JSONDecodeError with truncation
	longJSON := strings.Repeat("x", 120) + "{invalid}"
	jsonErr := NewJSONDecodeError(longJSON, 0, errors.New("syntax error"))
	jsonMsg := jsonErr.Error()

	// Should truncate long lines appropriately
	if !strings.Contains(jsonMsg, "Failed to decode JSON:") {
		t.Errorf("JSON error should start with descriptive prefix")
	}

	if !strings.Contains(jsonMsg, "...") {
		t.Errorf("JSON error should truncate long lines with ...")
	}

	// Should not be longer than reasonable (prefix + 100 chars + ... + some padding)
	if len(jsonMsg) > 150 {
		t.Errorf("JSON error message seems too long: %d chars", len(jsonMsg))
	}
}

// TestPythonSDKExactAlignment tests exact alignment with Python SDK behavior
func TestPythonSDKExactAlignment(t *testing.T) {
	// Test ProcessError format: "message (exit code: X)\nError output: stderr"
	procErr := NewProcessError("Process failed", 1, "Command not found")
	expected := "Process failed (exit code: 1)\nError output: Command not found"
	if procErr.Error() != expected {
		t.Errorf("ProcessError format mismatch:\nExpected: %q\nGot:      %q", expected, procErr.Error())
	}
	
	// Test CLINotFoundError with path: "message: path"
	cliErr := NewCLINotFoundError("/path/to/claude", "Claude Code not found") 
	expectedCli := "Claude Code not found: /path/to/claude"
	if cliErr.Error() != expectedCli {
		t.Errorf("CLINotFoundError format mismatch:\nExpected: %q\nGot:      %q", expectedCli, cliErr.Error())
	}
	
	// Test JSONDecodeError exact Python format: f"Failed to decode JSON: {line[:100]}..."
	longJSON := "this is a very long json line that should be truncated at exactly 100 characters"
	jsonErr := NewJSONDecodeError(longJSON, 0, errors.New("syntax error"))
	expectedJSON := "Failed to decode JSON: " + longJSON + "..."
	if jsonErr.Error() != expectedJSON {
		t.Errorf("JSONDecodeError format mismatch:\nExpected: %q\nGot:      %q", expectedJSON, jsonErr.Error())
	}
	
	// Test truncation with >100 char line
	longJSON2 := strings.Repeat("x", 120)
	jsonErr2 := NewJSONDecodeError(longJSON2, 0, errors.New("syntax error"))
	expectedJSON2 := "Failed to decode JSON: " + longJSON2[:100] + "..."
	if jsonErr2.Error() != expectedJSON2 {
		t.Errorf("JSONDecodeError truncation mismatch:\nExpected: %q\nGot:      %q", expectedJSON2, jsonErr2.Error())
	}
}
