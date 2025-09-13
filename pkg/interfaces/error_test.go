package interfaces

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// TestErrorInterfaceExistence verifies that the error interfaces exist and have correct method signatures.
func TestErrorInterfaceExistence(t *testing.T) {
	tests := []struct {
		name            string
		interfaceType   reflect.Type
		expectedMethods []string
	}{
		{
			name:            "SDKError interface",
			interfaceType:   reflect.TypeOf((*SDKError)(nil)).Elem(),
			expectedMethods: []string{"Error", "Type"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.interfaceType.Kind() != reflect.Interface {
				t.Errorf("Expected %s to be an interface, got %s", tt.name, tt.interfaceType.Kind())
				return
			}

			actualMethods := make(map[string]bool)
			for i := 0; i < tt.interfaceType.NumMethod(); i++ {
				method := tt.interfaceType.Method(i)
				actualMethods[method.Name] = true
			}

			for _, expectedMethod := range tt.expectedMethods {
				if !actualMethods[expectedMethod] {
					t.Errorf("Expected %s to have method %s", tt.name, expectedMethod)
				}
			}
		})
	}
}

// TestSDKErrorInterfaceEmbedding verifies SDKError embeds the standard error interface.
func TestSDKErrorInterfaceEmbedding(t *testing.T) {
	sdkErrorType := reflect.TypeOf((*SDKError)(nil)).Elem()
	stdErrorType := reflect.TypeOf((*error)(nil)).Elem()

	// SDKError should embed the standard error interface
	var hasErrorMethod bool
	for i := 0; i < sdkErrorType.NumMethod(); i++ {
		method := sdkErrorType.Method(i)
		if method.Name == "Error" {
			hasErrorMethod = true

			// Check signature matches standard error interface
			stdErrorMethod, _ := stdErrorType.MethodByName("Error")
			if !methodSignaturesEqual(method.Type, stdErrorMethod.Type) {
				t.Error("SDKError.Error() method should have same signature as standard error interface")
			}
			break
		}
	}

	if !hasErrorMethod {
		t.Error("SDKError interface should embed standard error interface (have Error() method)")
	}
}

// TestSDKErrorTypeMethodSignature verifies Type() method has correct signature.
func TestSDKErrorTypeMethodSignature(t *testing.T) {
	errorType := reflect.TypeOf((*SDKError)(nil)).Elem()

	method, found := errorType.MethodByName("Type")
	if !found {
		t.Error("SDKError interface must have Type method")
		return
	}

	methodType := method.Type

	// Should take no parameters (interface methods don't count receiver)
	if methodType.NumIn() != 0 {
		t.Errorf("Type method should have 0 parameters, got %d", methodType.NumIn())
	}

	// Should return exactly 1 value (string)
	if methodType.NumOut() != 1 {
		t.Errorf("Type method should return 1 value, got %d", methodType.NumOut())
	}

	// Return type should be string
	if methodType.NumOut() > 0 {
		returnType := methodType.Out(0)
		if returnType.Kind() != reflect.String {
			t.Errorf("Type method should return string, got %s", returnType.Kind())
		}
	}
}

// TestErrorInterfaceCompatibility verifies SDKError is compatible with standard error interface.
func TestErrorInterfaceCompatibility(t *testing.T) {
	// This test ensures SDKError can be used wherever error is expected
	sdkErrorType := reflect.TypeOf((*SDKError)(nil)).Elem()
	stdErrorType := reflect.TypeOf((*error)(nil)).Elem()

	// SDKError should be assignable to error
	if !sdkErrorType.Implements(stdErrorType) {
		t.Error("SDKError should implement standard error interface")
	}

	// Test with errors package functions
	// These will be tested with actual implementations later, but interface should support it
}

// TestErrorInterfaceNilHandling verifies interfaces handle nil values correctly.
func TestErrorInterfaceNilHandling(t *testing.T) {
	var sdkError SDKError

	if sdkError != nil {
		t.Error("Nil SDKError interface should be nil")
	}

	// Test error interface compatibility with nil
	var err error = sdkError
	if err != nil {
		t.Error("Nil SDKError assigned to error interface should be nil")
	}
}

// TestErrorInterfaceWithErrorsPackage verifies compatibility with errors package patterns.
func TestErrorInterfaceWithErrorsPackage(t *testing.T) {
	// This test verifies that SDKError interface supports Go 1.13+ error patterns
	// We can't test the actual implementation here, but we can verify interface compatibility

	sdkErrorType := reflect.TypeOf((*SDKError)(nil)).Elem()

	// Should implement standard error interface for errors.Is()
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	if !sdkErrorType.Implements(errorType) {
		t.Error("SDKError must implement error interface for errors.Is/As compatibility")
	}

	// Test basic error creation pattern would work
	var sdkErr SDKError
	var stdErr error = sdkErr

	// Should be able to use with errors.Is (when implementations exist)
	if stdErr != nil {
		errors.Is(stdErr, stdErr) // This should compile
	}
}

// TestErrorTypesConsistency verifies error types follow consistent patterns.
func TestErrorTypesConsistency(t *testing.T) {
	// All error types should have Type() method that returns string
	// This enforces consistency with Message and ContentBlock interfaces
	sdkErrorType := reflect.TypeOf((*SDKError)(nil)).Elem()

	typeMethod, found := sdkErrorType.MethodByName("Type")
	if !found {
		t.Error("SDKError must have Type() method for consistency with other interfaces")
		return
	}

	// Should match signature pattern used by Message and ContentBlock
	methodType := typeMethod.Type
	if methodType.NumIn() != 0 || methodType.NumOut() != 1 {
		t.Error("SDKError.Type() should have same signature pattern as Message.Type() and ContentBlock.Type()")
	}

	if methodType.NumOut() > 0 && methodType.Out(0).Kind() != reflect.String {
		t.Error("SDKError.Type() should return string for consistency with other Type() methods")
	}
}

// TestConcreteErrorImplementations verifies all concrete error types implement SDKError.
func TestConcreteErrorImplementations(t *testing.T) {
	tests := []struct {
		name         string
		error        SDKError
		expectedType string
	}{
		{
			name:         "ConnectionError implements SDKError",
			error:        NewConnectionError("test connection error", nil),
			expectedType: "connection_error",
		},
		{
			name:         "DiscoveryError implements SDKError",
			error:        NewDiscoveryError("test-service", "test discovery error", nil),
			expectedType: "discovery_error",
		},
		{
			name:         "ValidationError implements SDKError",
			error:        NewValidationError("test-field", "test validation error", "invalid-value"),
			expectedType: "validation_error",
		},
		{
			name:         "CLINotFoundError implements SDKError",
			error:        NewCLINotFoundError("/usr/bin/claude", "CLI not found"),
			expectedType: "cli_not_found_error",
		},
		{
			name:         "ProcessError implements SDKError",
			error:        NewProcessError("process failed", 1, "stderr output"),
			expectedType: "process_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test interface implementation
			var _ SDKError = tt.error
			var _ error = tt.error

			// Test Type() method returns expected value
			actual := tt.error.Type()
			if actual != tt.expectedType {
				t.Errorf("Expected Type() to return %q, got %q", tt.expectedType, actual)
			}

			// Test Error() method works
			errorMsg := tt.error.Error()
			if errorMsg == "" {
				t.Error("Error() method should return non-empty string")
			}
		})
	}
}

// TestConcreteErrorTypeMethodConsistency verifies Type() method consistency across concrete error types.
func TestConcreteErrorTypeMethodConsistency(t *testing.T) {
	errors := []SDKError{
		NewConnectionError("test", nil),
		NewDiscoveryError("service", "test", nil),
		NewValidationError("field", "test", nil),
		NewCLINotFoundError("path", "test"),
		NewProcessError("test", 1, "stderr"),
	}

	for i, err := range errors {
		t.Run(fmt.Sprintf("Error_%d_Type_method", i), func(t *testing.T) {
			// Test that Type() method exists and returns string
			errorType := reflect.TypeOf(err)
			method, found := errorType.MethodByName("Type")
			if !found {
				t.Error("Concrete error type must have Type() method")
				return
			}

			// Verify method signature
			methodType := method.Type
			if methodType.NumIn() != 1 { // receiver
				t.Errorf("Type method should have 1 parameter (receiver), got %d", methodType.NumIn())
			}
			if methodType.NumOut() != 1 {
				t.Errorf("Type method should return 1 value, got %d", methodType.NumOut())
			}
			if methodType.NumOut() > 0 && methodType.Out(0).Kind() != reflect.String {
				t.Error("Type method should return string")
			}

			// Test actual call
			errorTypeStr := err.Type()
			if errorTypeStr == "" {
				t.Error("Type() method should return non-empty string")
			}
		})
	}
}

// TestErrorWrappingBehavior verifies error wrapping follows Go 1.13+ patterns.
func TestErrorWrappingBehavior(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := NewConnectionError("connection failed", originalErr)

	// Test Unwrap behavior - ConnectionError embeds BaseError which has Unwrap()
	unwrapped := wrappedErr.BaseError.Unwrap()
	if !errors.Is(unwrapped, originalErr) {
		t.Error("Unwrap should return the original error")
	}

	// Test errors.Is compatibility
	if !errors.Is(wrappedErr, originalErr) {
		t.Error("errors.Is should find the original error in the wrapped error")
	}

	// Test error message includes both messages
	errorMsg := wrappedErr.Error()
	if !strings.Contains(errorMsg, "connection failed") {
		t.Error("Error message should contain the wrapper message")
	}
	if !strings.Contains(errorMsg, "original error") {
		t.Error("Error message should contain the original error message")
	}
}

// TestErrorsWithoutWrapping verifies errors work without wrapped causes.
func TestErrorsWithoutWrapping(t *testing.T) {
	tests := []struct {
		name string
		err  SDKError
	}{
		{"ConnectionError without cause", NewConnectionError("standalone error", nil)},
		{"ValidationError without cause", NewValidationError("field", "validation failed", "value")},
		{"CLINotFoundError without cause", NewCLINotFoundError("/path", "not found")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Error message should work without wrapped error
			errorMsg := tt.err.Error()
			if errorMsg == "" {
				t.Error("Error() should return non-empty message even without wrapped error")
			}

			// Unwrap should return nil for errors without causes
			if unwrapper, ok := tt.err.(interface{ Unwrap() error }); ok {
				if unwrapper.Unwrap() != nil {
					t.Error("Unwrap() should return nil for errors without wrapped causes")
				}
			}

			// Type() should still work
			errorType := tt.err.Type()
			if errorType == "" {
				t.Error("Type() should return non-empty string even without wrapped error")
			}
		})
	}
}

// TestErrorFieldAccess verifies structured error fields are accessible.
func TestErrorFieldAccess(t *testing.T) {
	// Test DiscoveryError field access
	discoveryErr := NewDiscoveryError("test-service", "discovery failed", nil)
	if discoveryErr.Service != "test-service" {
		t.Errorf("Expected DiscoveryError.Service to be %q, got %q", "test-service", discoveryErr.Service)
	}

	// Test ValidationError field access
	validationErr := NewValidationError("test-field", "validation failed", "test-value")
	if validationErr.Field != "test-field" {
		t.Errorf("Expected ValidationError.Field to be %q, got %q", "test-field", validationErr.Field)
	}
	if validationErr.Value != "test-value" {
		t.Errorf("Expected ValidationError.Value to be %q, got %q", "test-value", validationErr.Value)
	}

	// Test CLINotFoundError field access
	cliErr := NewCLINotFoundError("/usr/bin/claude", "CLI not found")
	if cliErr.Path != "/usr/bin/claude" {
		t.Errorf("Expected CLINotFoundError.Path to be %q, got %q", "/usr/bin/claude", cliErr.Path)
	}

	// Test ProcessError field access
	processErr := NewProcessError("process failed", 42, "stderr output")
	if processErr.ExitCode != 42 {
		t.Errorf("Expected ProcessError.ExitCode to be %d, got %d", 42, processErr.ExitCode)
	}
	if processErr.Stderr != "stderr output" {
		t.Errorf("Expected ProcessError.Stderr to be %q, got %q", "stderr output", processErr.Stderr)
	}
}

// TestErrorZeroValues verifies zero value behavior for error types.
func TestErrorZeroValues(t *testing.T) {
	tests := []struct {
		name string
		err  SDKError
	}{
		{"ConnectionError zero value", &ConnectionError{}},
		{"DiscoveryError zero value", &DiscoveryError{}},
		{"ValidationError zero value", &ValidationError{}},
		{"CLINotFoundError zero value", &CLINotFoundError{}},
		{"ProcessError zero value", &ProcessError{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Zero value should still implement interface correctly
			var _ SDKError = tt.err

			// Type() should work even with zero values
			errorType := tt.err.Type()
			if errorType == "" {
				t.Error("Type() should return non-empty string even for zero values")
			}

			// Error() should not panic with zero values
			errorMsg := tt.err.Error()
			// Error message might be empty for zero values, but should not panic
			_ = errorMsg
		})
	}
}

// TestBaseErrorType verifies BaseError.Type() method works correctly.
func TestBaseErrorType(t *testing.T) {
	baseErr := &BaseError{message: "test error", cause: nil}

	expectedType := "base_error"
	actualType := baseErr.Type()

	if actualType != expectedType {
		t.Errorf("Expected BaseError.Type() to return %q, got %q", expectedType, actualType)
	}
}

// TestJSONDecodeError verifies JSONDecodeError functionality.
func TestJSONDecodeError(t *testing.T) {
	t.Run("NewJSONDecodeError creation", func(t *testing.T) {
		originalErr := errors.New("json parse error")
		line := "this is a very long line that should be truncated when it exceeds the maximum display length of 100 characters"
		position := 50

		jsonErr := NewJSONDecodeError(line, position, originalErr)

		// Test Type() method
		if jsonErr.Type() != "json_decode_error" {
			t.Errorf("Expected Type() to return 'json_decode_error', got %q", jsonErr.Type())
		}

		// Test Unwrap() method
		unwrapped := jsonErr.Unwrap()
		if !errors.Is(unwrapped, originalErr) {
			t.Error("Unwrap() should return the original error")
		}

		// Test Error() message contains truncated line
		errorMsg := jsonErr.Error()
		if !strings.Contains(errorMsg, "Failed to decode JSON") {
			t.Errorf("Error message should contain 'Failed to decode JSON', got: %s", errorMsg)
		}

		// Test field access
		if jsonErr.Line != line {
			t.Errorf("Expected Line field to be %q, got %q", line, jsonErr.Line)
		}
		if jsonErr.Position != position {
			t.Errorf("Expected Position field to be %d, got %d", position, jsonErr.Position)
		}
		if jsonErr.OriginalError != originalErr {
			t.Errorf("Expected OriginalError field to be %v, got %v", originalErr, jsonErr.OriginalError)
		}
	})

	t.Run("Line truncation for long lines", func(t *testing.T) {
		// Create a line longer than 100 characters
		longLine := strings.Repeat("x", 150)
		jsonErr := NewJSONDecodeError(longLine, 10, errors.New("test"))

		// The error message should contain truncated line
		errorMsg := jsonErr.Error()
		if !strings.Contains(errorMsg, strings.Repeat("x", 100)) {
			t.Error("Error message should contain truncated line")
		}

		// But the Line field should contain the full line
		if jsonErr.Line != longLine {
			t.Error("Line field should contain the full untruncated line")
		}
	})
}

// TestMessageParseError verifies MessageParseError functionality.
func TestMessageParseError(t *testing.T) {
	testData := map[string]interface{}{"invalid": "data"}
	parseErr := NewMessageParseError("invalid message format", testData)

	// Test Type() method
	if parseErr.Type() != "message_parse_error" {
		t.Errorf("Expected Type() to return 'message_parse_error', got %q", parseErr.Type())
	}

	// Test Error() message
	errorMsg := parseErr.Error()
	if errorMsg != "invalid message format" {
		t.Errorf("Expected error message 'invalid message format', got %q", errorMsg)
	}

	// Test field access - compare by value since maps can't be compared directly
	if parseErr.Data == nil {
		t.Error("Expected Data field to not be nil")
	} else {
		dataMap, ok := parseErr.Data.(map[string]interface{})
		if !ok {
			t.Errorf("Expected Data to be map[string]interface{}, got %T", parseErr.Data)
		} else if dataMap["invalid"] != "data" {
			t.Errorf("Expected Data['invalid'] to be 'data', got %v", dataMap["invalid"])
		}
	}

	// Test with nil data
	nilParseErr := NewMessageParseError("test with nil data", nil)
	if nilParseErr.Data != nil {
		t.Errorf("Expected Data field to be nil, got %v", nilParseErr.Data)
	}
}
