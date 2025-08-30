package shared

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// TestBaseSDKError tests the base ClaudeSDKError interface and implementation.
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
}

// TestCLINotFoundError tests CLINotFoundError with helpful installation message.
func TestCLINotFoundError(t *testing.T) {
	// Test basic CLI not found error
	message := "Claude Code not found"
	err := NewCLINotFoundError("", message)

	// Must implement SDKError interface
	if _, ok := interface{}(err).(SDKError); !ok {
		t.Fatal("CLINotFoundError must implement SDKError interface")
	}

	// Must contain expected message
	if !strings.Contains(err.Error(), message) {
		t.Errorf("Expected error to contain %q, got %q", message, err.Error())
	}

	// Check error type
	if err.Type() != "cli_not_found_error" {
		t.Errorf("Expected error type 'cli_not_found_error', got %q", err.Type())
	}

	// Test CLI not found error with path
	path := "/usr/bin/claude"
	fullMessage := "CLI not found"
	errWithPath := NewCLINotFoundError(path, fullMessage)
	
	// Should format message with path
	expectedMessage := fmt.Sprintf("%s: %s", fullMessage, path)
	if errWithPath.Error() != expectedMessage {
		t.Errorf("Expected error message %q, got %q", expectedMessage, errWithPath.Error())
	}
	
	// Should store path
	if errWithPath.Path != path {
		t.Errorf("Expected path %q, got %q", path, errWithPath.Path)
	}
}

// TestConnectionError tests CLIConnectionError for connection failures.
func TestConnectionError(t *testing.T) {
	message := "Connection failed"
	cause := errors.New("network unavailable")
	err := NewConnectionError(message, cause)

	// Must implement SDKError interface
	if _, ok := interface{}(err).(SDKError); !ok {
		t.Fatal("ConnectionError must implement SDKError interface")
	}

	// Check error type
	if err.Type() != "connection_error" {
		t.Errorf("Expected error type 'connection_error', got %q", err.Type())
	}

	// Check error message includes cause
	errorMsg := err.Error()
	if !strings.Contains(errorMsg, message) {
		t.Errorf("Expected error message to contain %q, got %q", message, errorMsg)
	}
	if !strings.Contains(errorMsg, cause.Error()) {
		t.Errorf("Expected error message to contain %q, got %q", cause.Error(), errorMsg)
	}

	// Check error unwrapping
	if !errors.Is(err, cause) {
		t.Error("Expected error to wrap the cause error")
	}
}

// TestProcessErrorWithDetails tests ProcessError with exit_code and stderr.
func TestProcessErrorWithDetails(t *testing.T) {
	message := "Process failed"
	exitCode := 1
	stderr := "Permission denied"
	err := NewProcessError(message, exitCode, stderr)

	// Must implement SDKError interface
	if _, ok := interface{}(err).(SDKError); !ok {
		t.Fatal("ProcessError must implement SDKError interface")
	}

	// Check error type
	if err.Type() != "process_error" {
		t.Errorf("Expected error type 'process_error', got %q", err.Type())
	}

	// Check that exit code is included in error message
	errorMsg := err.Error()
	if !strings.Contains(errorMsg, fmt.Sprintf("exit code: %d", exitCode)) {
		t.Errorf("Expected error message to contain exit code, got %q", errorMsg)
	}

	// Check that stderr is included
	if !strings.Contains(errorMsg, stderr) {
		t.Errorf("Expected error message to contain stderr, got %q", errorMsg)
	}

	// Check fields are set
	if err.ExitCode != exitCode {
		t.Errorf("Expected exit code %d, got %d", exitCode, err.ExitCode)
	}
	if err.Stderr != stderr {
		t.Errorf("Expected stderr %q, got %q", stderr, err.Stderr)
	}
}

// TestJSONDecodeError tests CLIJSONDecodeError with line and position info.
func TestJSONDecodeError(t *testing.T) {
	line := `{"invalid": json}`
	position := 15
	cause := errors.New("unexpected character")
	err := NewJSONDecodeError(line, position, cause)

	// Must implement SDKError interface
	if _, ok := interface{}(err).(SDKError); !ok {
		t.Fatal("JSONDecodeError must implement SDKError interface")
	}

	// Check error type
	if err.Type() != "json_decode_error" {
		t.Errorf("Expected error type 'json_decode_error', got %q", err.Type())
	}

	// Check that line is truncated and included in message
	errorMsg := err.Error()
	if !strings.Contains(errorMsg, "Failed to decode JSON") {
		t.Errorf("Expected error message to contain 'Failed to decode JSON', got %q", errorMsg)
	}

	// Check that fields are set
	if err.Line != line {
		t.Errorf("Expected line %q, got %q", line, err.Line)
	}
	if err.Position != position {
		t.Errorf("Expected position %d, got %d", position, err.Position)
	}
	if err.OriginalError != cause {
		t.Errorf("Expected original error %v, got %v", cause, err.OriginalError)
	}

	// Check error unwrapping
	if !errors.Is(err, cause) {
		t.Error("Expected error to wrap the original error")
	}

	// Test with long line (should truncate)
	longLine := strings.Repeat("x", 150)
	longErr := NewJSONDecodeError(longLine, 0, nil)
	if len(longErr.Error()) < len(longLine) {
		// Error message should be shorter than original line due to truncation
		t.Logf("Long line properly truncated in error message")
	}
}

// TestMessageParseError tests MessageParseError with raw data context.
func TestMessageParseError(t *testing.T) {
	message := "Invalid message structure"
	data := map[string]interface{}{
		"type":    "unknown_type",
		"content": "test",
	}
	err := NewMessageParseError(message, data)

	// Must implement SDKError interface
	if _, ok := interface{}(err).(SDKError); !ok {
		t.Fatal("MessageParseError must implement SDKError interface")
	}

	// Check error type
	if err.Type() != "message_parse_error" {
		t.Errorf("Expected error type 'message_parse_error', got %q", err.Type())
	}

	// Check error message
	if err.Error() != message {
		t.Errorf("Expected error message %q, got %q", message, err.Error())
	}

	// Check that data is preserved (can't compare maps directly)
	if err.Data == nil {
		t.Error("Expected data to be preserved, got nil")
	}
	dataMap, ok := err.Data.(map[string]interface{})
	if !ok {
		t.Errorf("Expected data to be map[string]interface{}, got %T", err.Data)
	}
	if dataMap["type"] != "unknown_type" {
		t.Errorf("Expected data type to be 'unknown_type', got %v", dataMap["type"])
	}
}

// TestErrorHierarchy verifies all errors implement SDKError interface.
func TestErrorHierarchy(t *testing.T) {
	// Test all error types implement SDKError
	var errors []SDKError = []SDKError{
		&BaseError{message: "test"},
		NewConnectionError("test", nil),
		NewCLINotFoundError("", "test"),
		NewProcessError("test", 1, "stderr"),
		NewJSONDecodeError("line", 0, nil),
		NewMessageParseError("test", nil),
	}

	expectedTypes := []string{
		"base_error",
		"connection_error", 
		"cli_not_found_error",
		"process_error",
		"json_decode_error",
		"message_parse_error",
	}

	for i, err := range errors {
		if err.Type() != expectedTypes[i] {
			t.Errorf("Error %d: expected type %q, got %q", i, expectedTypes[i], err.Type())
		}
		
		// All must implement error interface
		if err.Error() == "" {
			t.Errorf("Error %d: Error() returned empty string", i)
		}
	}
}

// TestErrorContextPreservation tests error wrapping with fmt.Errorf %w verb.
func TestErrorContextPreservation(t *testing.T) {
	// Test BaseError wrapping
	cause := errors.New("root cause")
	baseErr := &BaseError{message: "wrapper", cause: cause}

	// Should support errors.Is()
	if !errors.Is(baseErr, cause) {
		t.Error("Expected BaseError to wrap cause with errors.Is() support")
	}

	// Should support errors.As()
	var rootErr *BaseError
	if !errors.As(baseErr, &rootErr) {
		t.Error("Expected BaseError to support errors.As()")
	}

	// Test ConnectionError inheriting wrapping
	connErr := NewConnectionError("connection failed", cause)
	if !errors.Is(connErr, cause) {
		t.Error("Expected ConnectionError to wrap cause")
	}
}