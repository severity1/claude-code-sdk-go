package interfaces

import (
	"reflect"
	"testing"
)

// TestMessageInterfaceExistence verifies that the message interfaces exist and have correct method signatures.
func TestMessageInterfaceExistence(t *testing.T) {
	tests := []struct {
		name            string
		interfaceType   reflect.Type
		expectedMethods []string
	}{
		{
			name:            "Message interface",
			interfaceType:   reflect.TypeOf((*Message)(nil)).Elem(),
			expectedMethods: []string{"Type"},
		},
		{
			name:            "ContentBlock interface",
			interfaceType:   reflect.TypeOf((*ContentBlock)(nil)).Elem(),
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

// TestTypeMethodSignature verifies Type() method has correct signature.
func TestTypeMethodSignature(t *testing.T) {
	testCases := []struct {
		name          string
		interfaceType reflect.Type
	}{
		{"Message", reflect.TypeOf((*Message)(nil)).Elem()},
		{"ContentBlock", reflect.TypeOf((*ContentBlock)(nil)).Elem()},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			method, found := tc.interfaceType.MethodByName("Type")
			if !found {
				t.Errorf("%s interface must have Type method", tc.name)
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
		})
	}
}

// TestMethodNamingConsistency verifies consistent naming across interfaces.
func TestMethodNamingConsistency(t *testing.T) {
	// Both Message and ContentBlock should use Type(), not BlockType() or GetType()
	messageType := reflect.TypeOf((*Message)(nil)).Elem()
	contentBlockType := reflect.TypeOf((*ContentBlock)(nil)).Elem()

	// Message should have Type() method
	messageMethod, messageHasType := messageType.MethodByName("Type")
	if !messageHasType {
		t.Error("Message interface must have Type() method")
	}

	// ContentBlock should have Type() method (not BlockType())
	blockMethod, blockHasType := contentBlockType.MethodByName("Type")
	if !blockHasType {
		t.Error("ContentBlock interface must have Type() method (not BlockType())")
	}

	// Verify methods have identical signatures
	if messageHasType && blockHasType {
		if !methodSignaturesEqual(messageMethod.Type, blockMethod.Type) {
			t.Error("Message.Type() and ContentBlock.Type() should have identical signatures")
		}
	}

	// Verify no legacy method names exist
	if _, hasBlockType := contentBlockType.MethodByName("BlockType"); hasBlockType {
		t.Error("ContentBlock interface should not have BlockType() method - use Type() instead")
	}
}

// TestInterfaceNilHandling verifies interfaces handle nil values correctly.
func TestInterfaceNilHandling(t *testing.T) {
	var message Message
	var contentBlock ContentBlock

	if message != nil {
		t.Error("Nil Message interface should be nil")
	}
	if contentBlock != nil {
		t.Error("Nil ContentBlock interface should be nil")
	}

	// Test interface value equality
	var message2 Message
	if message != message2 {
		t.Error("Two nil Message interfaces should be equal")
	}
}

// Helper function to compare method signatures
func methodSignaturesEqual(t1, t2 reflect.Type) bool {
	if t1.NumIn() != t2.NumIn() || t1.NumOut() != t2.NumOut() {
		return false
	}

	for i := 0; i < t1.NumIn(); i++ {
		if t1.In(i) != t2.In(i) {
			return false
		}
	}

	for i := 0; i < t1.NumOut(); i++ {
		if t1.Out(i) != t2.Out(i) {
			return false
		}
	}

	return true
}
