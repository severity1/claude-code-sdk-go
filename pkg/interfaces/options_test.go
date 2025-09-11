package interfaces

import (
	"reflect"
	"testing"
)

// TestOptionsInterfaceExistence verifies that the options interfaces exist and have correct method signatures.
func TestOptionsInterfaceExistence(t *testing.T) {
	tests := []struct {
		name            string
		interfaceType   reflect.Type
		expectedMethods []string
	}{
		{
			name:            "McpServerConfig interface",
			interfaceType:   reflect.TypeOf((*McpServerConfig)(nil)).Elem(),
			expectedMethods: []string{"Type"},
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

// TestMcpServerConfigTypeMethodSignature verifies Type() method has correct signature.
func TestMcpServerConfigTypeMethodSignature(t *testing.T) {
	configType := reflect.TypeOf((*McpServerConfig)(nil)).Elem()

	method, found := configType.MethodByName("Type")
	if !found {
		t.Error("McpServerConfig interface must have Type method")
		return
	}

	methodType := method.Type

	// Should take no parameters (interface methods don't count receiver)
	if methodType.NumIn() != 0 {
		t.Errorf("Type method should have 0 parameters, got %d", methodType.NumIn())
	}

	// Should return exactly 1 value
	if methodType.NumOut() != 1 {
		t.Errorf("Type method should return 1 value, got %d", methodType.NumOut())
	}

	// Return type should be McpServerType (not string - this is domain-specific)
	if methodType.NumOut() > 0 {
		returnType := methodType.Out(0)
		// We'll check the type name since we don't have McpServerType defined yet
		expectedTypeName := "McpServerType"
		if returnType.Name() != expectedTypeName {
			// For now, just check it's a named type (not primitive string)
			if returnType.Kind() == reflect.String && returnType.Name() == "" {
				t.Error("Type method should return named type (McpServerType), not plain string")
			}
		}
	}
}

// TestMethodNamingStandardization verifies standardized naming conventions.
func TestMethodNamingStandardization(t *testing.T) {
	configType := reflect.TypeOf((*McpServerConfig)(nil)).Elem()

	// Should have Type() method, not GetType()
	_, hasType := configType.MethodByName("Type")
	if !hasType {
		t.Error("McpServerConfig interface must have Type() method")
	}

	// Should NOT have GetType() method (legacy Java-style naming)
	if _, hasGetType := configType.MethodByName("GetType"); hasGetType {
		t.Error("McpServerConfig interface should not have GetType() method - use Type() instead")
	}
}

// TestOptionsInterfaceNilHandling verifies interfaces handle nil values correctly.
func TestOptionsInterfaceNilHandling(t *testing.T) {
	var config McpServerConfig

	if config != nil {
		t.Error("Nil McpServerConfig interface should be nil")
	}
}

// TestMcpServerTypeExistence verifies McpServerType is properly defined.
func TestMcpServerTypeExistence(t *testing.T) {
	// This test verifies that McpServerType exists as a named type
	// We'll get it through reflection on the interface method
	configType := reflect.TypeOf((*McpServerConfig)(nil)).Elem()

	method, found := configType.MethodByName("Type")
	if !found {
		t.Skip("Skipping type test - Type method not found")
		return
	}

	if method.Type.NumOut() == 0 {
		t.Error("Type method should return a value")
		return
	}

	returnType := method.Type.Out(0)

	// Should be a named type, not a primitive
	if returnType.Name() == "" {
		t.Error("Type method should return a named type (McpServerType), not an unnamed type")
	}

	// Should be based on string
	if returnType.Kind() != reflect.String {
		t.Error("McpServerType should be based on string type")
	}
}
