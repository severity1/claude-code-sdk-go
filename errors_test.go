package claudecode

import (
	"testing"
)

// TestErrorReExports tests that error type re-exports work properly
func TestErrorReExports(t *testing.T) {
	// Test that re-exported constructor functions work
	connErr := NewConnectionError("test connection", nil)
	assertErrorCreated(t, connErr)
	assertErrorType(t, connErr, "connection_error")

	// Test CLI not found error
	cliErr := NewCLINotFoundError("", "CLI not found")
	assertErrorCreated(t, cliErr)
	assertErrorType(t, cliErr, "cli_not_found_error")

	// Test process error
	processErr := NewProcessError("process failed", 1, "error output")
	assertErrorCreated(t, processErr)
	assertErrorType(t, processErr, "process_error")

	// Test JSON decode error
	jsonErr := NewJSONDecodeError("invalid json", 0, nil)
	assertErrorCreated(t, jsonErr)
	assertErrorType(t, jsonErr, "json_decode_error")

	// Test message parse error
	msgErr := NewMessageParseError("parse failed", map[string]any{"type": "unknown"})
	assertErrorCreated(t, msgErr)
	assertErrorType(t, msgErr, "message_parse_error")

	// Test that all re-exported types implement SDKError
	assertSDKErrorInterface(t, connErr)
}

// TestErrorInterfaceCompatibility tests that re-exported errors work with public API
func TestErrorInterfaceCompatibility(t *testing.T) {
	// Create error through public API
	err := NewConnectionError("API test", nil)

	// Should work as SDKError
	assertSDKErrorInterface(t, err)
	assertErrorType(t, err, "connection_error")

	// Should work as regular error
	assertErrorMessage(t, err, false)

	// Test type assertion works
	assertTypeAssertion(t, err)
}

// Helper functions following client_test.go patterns

// assertErrorCreated verifies that error constructor created a non-nil error
func assertErrorCreated(t *testing.T, err SDKError) {
	t.Helper()
	if err == nil {
		t.Fatal("Expected error constructor to create error")
	}
}

// assertErrorType verifies that error has the expected type
func assertErrorType(t *testing.T, err SDKError, expectedType string) {
	t.Helper()
	if err.Type() != expectedType {
		t.Errorf("Expected error type %s, got %s", expectedType, err.Type())
	}
}

// assertErrorMessage verifies error message content
func assertErrorMessage(t *testing.T, err error, shouldBeEmpty bool) {
	t.Helper()
	msg := err.Error()
	if shouldBeEmpty && msg != "" {
		t.Errorf("Expected empty error message, got %s", msg)
	}
	if !shouldBeEmpty && msg == "" {
		t.Error("Expected non-empty error message through standard error interface")
	}
}

// assertSDKErrorInterface verifies that error implements SDKError interface properly
func assertSDKErrorInterface(t *testing.T, err SDKError) {
	t.Helper()
	var sdkErr SDKError = err
	if sdkErr.Error() == "" {
		t.Error("Expected error message from SDKError interface")
	}
}

// assertTypeAssertion verifies that type assertion works correctly for ConnectionError specifically
func assertTypeAssertion(t *testing.T, err SDKError) {
	t.Helper()
	var stdErr error = err
	if connErr, ok := stdErr.(*ConnectionError); !ok {
		t.Error("Expected type assertion to work with ConnectionError type")
	} else if connErr.Type() != "connection_error" {
		t.Errorf("Expected connection_error after type assertion, got %s", connErr.Type())
	}
}
