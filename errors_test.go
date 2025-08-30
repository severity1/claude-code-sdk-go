package claudecode

import (
	"testing"
)

// TestErrorReExports tests that error type re-exports work properly
func TestErrorReExports(t *testing.T) {
	// Test that re-exported constructor functions work
	connErr := NewConnectionError("test connection", nil)
	if connErr == nil {
		t.Fatal("Expected NewConnectionError to create error")
	}
	if connErr.Type() != "connection_error" {
		t.Errorf("Expected connection error type, got %s", connErr.Type())
	}

	// Test CLI not found error
	cliErr := NewCLINotFoundError("", "CLI not found")
	if cliErr == nil {
		t.Fatal("Expected NewCLINotFoundError to create error")
	}
	if cliErr.Type() != "cli_not_found_error" {
		t.Errorf("Expected CLI error type, got %s", cliErr.Type())
	}

	// Test process error
	processErr := NewProcessError("process failed", 1, "error output")
	if processErr == nil {
		t.Fatal("Expected NewProcessError to create error")
	}
	if processErr.Type() != "process_error" {
		t.Errorf("Expected process error type, got %s", processErr.Type())
	}

	// Test JSON decode error
	jsonErr := NewJSONDecodeError("invalid json", 0, nil)
	if jsonErr == nil {
		t.Fatal("Expected NewJSONDecodeError to create error")
	}
	if jsonErr.Type() != "json_decode_error" {
		t.Errorf("Expected JSON error type, got %s", jsonErr.Type())
	}

	// Test message parse error
	msgErr := NewMessageParseError("parse failed", map[string]interface{}{"type": "unknown"})
	if msgErr == nil {
		t.Fatal("Expected NewMessageParseError to create error")
	}
	if msgErr.Type() != "message_parse_error" {
		t.Errorf("Expected message error type, got %s", msgErr.Type())
	}

	// Test that all re-exported types implement SDKError
	var sdkErr SDKError = connErr
	if sdkErr.Error() == "" {
		t.Error("Expected error message from re-exported type")
	}
}

// TestErrorInterfaceCompatibility tests that re-exported errors work with public API
func TestErrorInterfaceCompatibility(t *testing.T) {
	// Create error through public API
	err := NewConnectionError("API test", nil)
	
	// Should work as SDKError
	var sdkErr SDKError = err
	if sdkErr.Type() != "connection_error" {
		t.Errorf("Expected connection_error type through interface, got %s", sdkErr.Type())
	}
	
	// Should work as regular error
	var stdErr error = err
	if stdErr.Error() == "" {
		t.Error("Expected non-empty error message through standard error interface")
	}
	
	// Test type assertion works
	if connErr, ok := stdErr.(*ConnectionError); !ok {
		t.Error("Expected type assertion to work with re-exported type")
	} else if connErr.Type() != "connection_error" {
		t.Errorf("Expected connection_error after type assertion, got %s", connErr.Type())
	}
}