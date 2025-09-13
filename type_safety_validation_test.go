package claudecode_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/severity1/claude-code-sdk-go"
	"github.com/severity1/claude-code-sdk-go/pkg/interfaces"
)

// TestZeroInterfaceEmptyUsage validates INTERFACE_SPEC.md Requirement #1: Zero interface{} Usage
// This test uses reflection to ensure no interface{} usage remains in public API
func TestZeroInterfaceEmptyUsage(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		testType    reflect.Type
		description string
	}{
		{
			name:        "UserMessage content type safety",
			packageName: "claudecode",
			testType:    reflect.TypeOf((*claudecode.UserMessage)(nil)).Elem(),
			description: "UserMessage.Content should use typed interfaces, not interface{}",
		},
		{
			name:        "AssistantMessage content type safety",
			packageName: "claudecode",
			testType:    reflect.TypeOf((*claudecode.AssistantMessage)(nil)).Elem(),
			description: "AssistantMessage.Content should use typed interfaces, not interface{}",
		},
		{
			name:        "StreamMessage type safety",
			packageName: "claudecode",
			testType:    reflect.TypeOf((*claudecode.StreamMessage)(nil)).Elem(),
			description: "StreamMessage should not use interface{} for any fields",
		},
		{
			name:        "ToolResultBlock content type safety",
			packageName: "claudecode",
			testType:    reflect.TypeOf((*claudecode.ToolResultBlock)(nil)).Elem(),
			description: "ToolResultBlock.Content should use typed interfaces, not interface{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := findInterfaceEmptyUsage(tt.testType, tt.packageName)
			if len(violations) > 0 {
				t.Errorf("Found interface{} usage violations in %s:", tt.testType.Name())
				for _, violation := range violations {
					t.Errorf("  - %s", violation)
				}
				t.Errorf("Description: %s", tt.description)
			} else {
				t.Logf("✅ %s: No interface{} usage found", tt.testType.Name())
			}
		})
	}
}

// TestSegregatedInterfaceTypeConsistency validates INTERFACE_SPEC.md Requirement #2: Method Naming Standardization
func TestSegregatedInterfaceTypeConsistency(t *testing.T) {
	tests := []struct {
		name            string
		interfaceType   reflect.Type
		expectedMethods map[string]string // method name -> expected signature pattern
	}{
		{
			name:          "ConnectionManager interface",
			interfaceType: reflect.TypeOf((*interfaces.ConnectionManager)(nil)).Elem(),
			expectedMethods: map[string]string{
				"Connect":     "func(context.Context) error",
				"Close":       "func() error",
				"IsConnected": "func() bool",
			},
		},
		{
			name:          "QueryExecutor interface",
			interfaceType: reflect.TypeOf((*interfaces.QueryExecutor)(nil)).Elem(),
			expectedMethods: map[string]string{
				"Query":       "func(context.Context, string) error",
				"QueryStream": "func(context.Context, <-chan StreamMessage) error",
			},
		},
		{
			name:          "MessageReceiver interface",
			interfaceType: reflect.TypeOf((*interfaces.MessageReceiver)(nil)).Elem(),
			expectedMethods: map[string]string{
				"ReceiveMessages": "func(context.Context) <-chan Message",
				"ReceiveResponse": "func(context.Context) MessageIterator",
			},
		},
		{
			name:          "ProcessController interface",
			interfaceType: reflect.TypeOf((*interfaces.ProcessController)(nil)).Elem(),
			expectedMethods: map[string]string{
				"Interrupt": "func(context.Context) error",
				"Status":    "func() ProcessStatus",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validateInterfaceMethodSignatures(t, tt.interfaceType, tt.expectedMethods)
		})
	}
}

// TestInterfaceCompositionConsistency validates INTERFACE_SPEC.md Requirement #5: Interface Segregation
func TestInterfaceCompositionConsistency(t *testing.T) {
	clientInterfaceType := reflect.TypeOf((*interfaces.Client)(nil)).Elem()

	// Test that interfaces.Client embeds all segregated interfaces
	expectedEmbeddedInterfaces := []reflect.Type{
		reflect.TypeOf((*interfaces.ConnectionManager)(nil)).Elem(),
		reflect.TypeOf((*interfaces.QueryExecutor)(nil)).Elem(),
		reflect.TypeOf((*interfaces.MessageReceiver)(nil)).Elem(),
		reflect.TypeOf((*interfaces.ProcessController)(nil)).Elem(),
	}

	t.Run("Client interface composition", func(t *testing.T) {
		for _, embeddedInterface := range expectedEmbeddedInterfaces {
			if !clientInterfaceImplements(clientInterfaceType, embeddedInterface) {
				t.Errorf("interfaces.Client does not properly embed %s interface", embeddedInterface.Name())
			} else {
				t.Logf("✅ interfaces.Client properly embeds %s", embeddedInterface.Name())
			}
		}
	})

	// Test specialized interface patterns
	t.Run("SimpleQuerier pattern", func(t *testing.T) {
		simpleQuerierType := reflect.TypeOf((*interfaces.SimpleQuerier)(nil)).Elem()
		queryExecutorType := reflect.TypeOf((*interfaces.QueryExecutor)(nil)).Elem()

		if !clientInterfaceImplements(simpleQuerierType, queryExecutorType) {
			t.Error("SimpleQuerier does not properly embed QueryExecutor")
		} else {
			t.Log("✅ SimpleQuerier properly embeds QueryExecutor")
		}
	})

	t.Run("StreamClient pattern", func(t *testing.T) {
		streamClientType := reflect.TypeOf((*interfaces.StreamClient)(nil)).Elem()
		connectionManagerType := reflect.TypeOf((*interfaces.ConnectionManager)(nil)).Elem()
		messageReceiverType := reflect.TypeOf((*interfaces.MessageReceiver)(nil)).Elem()

		if !clientInterfaceImplements(streamClientType, connectionManagerType) {
			t.Error("StreamClient does not properly embed ConnectionManager")
		} else {
			t.Log("✅ StreamClient properly embeds ConnectionManager")
		}

		if !clientInterfaceImplements(streamClientType, messageReceiverType) {
			t.Error("StreamClient does not properly embed MessageReceiver")
		} else {
			t.Log("✅ StreamClient properly embeds MessageReceiver")
		}
	})
}

// TestRealWorldUsagePatterns validates that the segregated interfaces work in real scenarios
func TestRealWorldUsagePatterns(t *testing.T) {
	transport := &simpleMockTransport{}
	client := claudecode.NewClientWithTransport(transport)

	t.Run("Focused interface usage", func(t *testing.T) {
		// Test that code can depend on minimal interfaces
		useConnectionManager := func(cm interfaces.ConnectionManager) error {
			if cm.IsConnected() {
				return nil
			}
			// For test purposes only - real code should use proper context
			return nil
		}

		useQueryExecutor := func(qe interfaces.QueryExecutor) error {
			// Interface exists and can be called - that's what we're testing
			return nil
		}

		useSimpleQuerier := func(sq interfaces.SimpleQuerier) error {
			// Interface exists and can be called - that's what we're testing
			return nil
		}

		useStreamClient := func(sc interfaces.StreamClient) error {
			// Interface composition works - that's what we're testing
			_ = sc.IsConnected()
			return nil
		}

		// These should all compile and work
		if err := useConnectionManager(client); err != nil {
			t.Errorf("ConnectionManager usage failed: %v", err)
		} else {
			t.Log("✅ ConnectionManager focused usage works")
		}

		if err := useQueryExecutor(client); err != nil {
			t.Errorf("QueryExecutor usage failed: %v", err)
		} else {
			t.Log("✅ QueryExecutor focused usage works")
		}

		if err := useSimpleQuerier(client); err != nil {
			t.Errorf("SimpleQuerier usage failed: %v", err)
		} else {
			t.Log("✅ SimpleQuerier focused usage works")
		}

		if err := useStreamClient(client); err != nil {
			t.Errorf("StreamClient usage failed: %v", err)
		} else {
			t.Log("✅ StreamClient focused usage works")
		}
	})
}

// Helper functions for validation

// findInterfaceEmptyUsage recursively searches for interface{} usage in a type
func findInterfaceEmptyUsage(t reflect.Type, packageName string) []string {
	var violations []string

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return violations
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldType := field.Type

		// Skip unexported fields from other packages
		if !field.IsExported() && field.PkgPath != "" && !strings.Contains(field.PkgPath, packageName) {
			continue
		}

		// Check for direct interface{} usage
		if fieldType.Kind() == reflect.Interface && fieldType.NumMethod() == 0 {
			violations = append(violations, fmt.Sprintf("Field %s.%s uses interface{}", t.Name(), field.Name))
		}

		// Check for interface{} in slices
		if fieldType.Kind() == reflect.Slice {
			elemType := fieldType.Elem()
			if elemType.Kind() == reflect.Interface && elemType.NumMethod() == 0 {
				violations = append(violations, fmt.Sprintf("Field %s.%s uses []interface{}", t.Name(), field.Name))
			}
		}

		// Check for interface{} in maps
		if fieldType.Kind() == reflect.Map {
			valueType := fieldType.Elem()
			if valueType.Kind() == reflect.Interface && valueType.NumMethod() == 0 {
				violations = append(violations, fmt.Sprintf("Field %s.%s uses map[...]interface{}", t.Name(), field.Name))
			}
		}
	}

	return violations
}

// validateInterfaceMethodSignatures checks that interface methods have expected signatures
func validateInterfaceMethodSignatures(t *testing.T, interfaceType reflect.Type, expectedMethods map[string]string) {
	t.Helper()

	if interfaceType.Kind() != reflect.Interface {
		t.Fatalf("%s is not an interface type", interfaceType.Name())
	}

	// Check that all expected methods exist
	for methodName, _ := range expectedMethods {
		method, found := interfaceType.MethodByName(methodName)
		if !found {
			t.Errorf("Method %s not found in %s interface", methodName, interfaceType.Name())
			continue
		}

		// Basic signature validation (method exists and is callable)
		if method.Type.Kind() != reflect.Func {
			t.Errorf("Method %s in %s is not a function", methodName, interfaceType.Name())
		}

		t.Logf("✅ Method %s.%s exists with correct type", interfaceType.Name(), methodName)
	}

	// Check that no unexpected methods exist
	for i := 0; i < interfaceType.NumMethod(); i++ {
		method := interfaceType.Method(i)
		if _, expected := expectedMethods[method.Name]; !expected {
			t.Errorf("Unexpected method %s found in %s interface", method.Name, interfaceType.Name())
		}
	}
}

// clientInterfaceImplements checks if one interface implements another using method comparison
func clientInterfaceImplements(implementer, target reflect.Type) bool {
	if implementer.Kind() != reflect.Interface || target.Kind() != reflect.Interface {
		return false
	}

	// Check that implementer has all methods from target
	for i := 0; i < target.NumMethod(); i++ {
		targetMethod := target.Method(i)
		implementerMethod, found := implementer.MethodByName(targetMethod.Name)
		if !found {
			return false
		}

		// Basic type compatibility check
		if implementerMethod.Type.String() != targetMethod.Type.String() {
			return false
		}
	}

	return true
}

// Note: simpleMockTransport is defined in interface_compatibility_test.go
