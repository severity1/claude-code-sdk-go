package interfaces

import (
	"errors"
	"reflect"
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
