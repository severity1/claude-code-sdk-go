package interfaces

import (
	"context"
	"reflect"
	"testing"
)

// TestTransportInterfaceExistence verifies that the transport interfaces exist and have correct method signatures.
func TestTransportInterfaceExistence(t *testing.T) {
	tests := []struct {
		name            string
		interfaceType   reflect.Type
		expectedMethods []string
	}{
		{
			name:            "Transport interface",
			interfaceType:   reflect.TypeOf((*Transport)(nil)).Elem(),
			expectedMethods: []string{"Connect", "SendMessage", "ReceiveMessages", "Interrupt", "Close"},
		},
		{
			name:            "MessageIterator interface",
			interfaceType:   reflect.TypeOf((*MessageIterator)(nil)).Elem(),
			expectedMethods: []string{"Next", "Close"},
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

// TestTransportMethodSignatures verifies Transport methods have correct signatures.
func TestTransportMethodSignatures(t *testing.T) {
	transportType := reflect.TypeOf((*Transport)(nil)).Elem()

	testCases := []struct {
		methodName          string
		expectedInputs      int // interface methods don't count receiver
		expectedOutputs     int
		firstParamIsContext bool
		lastReturnIsError   bool
	}{
		{"Connect", 1, 1, true, true},          // func(ctx context.Context) error
		{"SendMessage", 2, 1, true, true},      // func(ctx context.Context, message StreamMessage) error
		{"ReceiveMessages", 1, 2, true, false}, // func(ctx context.Context) (<-chan Message, <-chan error)
		{"Interrupt", 1, 1, true, true},        // func(ctx context.Context) error
		{"Close", 0, 1, false, true},           // func() error
	}

	for _, tc := range testCases {
		t.Run(tc.methodName, func(t *testing.T) {
			method, found := transportType.MethodByName(tc.methodName)
			if !found {
				t.Errorf("Transport interface must have %s method", tc.methodName)
				return
			}

			methodType := method.Type

			if methodType.NumIn() != tc.expectedInputs {
				t.Errorf("%s method should have %d inputs, got %d", tc.methodName, tc.expectedInputs, methodType.NumIn())
			}

			if methodType.NumOut() != tc.expectedOutputs {
				t.Errorf("%s method should have %d outputs, got %d", tc.methodName, tc.expectedOutputs, methodType.NumOut())
			}

			// Check context.Context as first parameter
			if tc.firstParamIsContext && methodType.NumIn() > 0 {
				contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
				if methodType.In(0) != contextType {
					t.Errorf("%s method first parameter should be context.Context, got %s", tc.methodName, methodType.In(0))
				}
			}

			// Check error as last return value
			if tc.lastReturnIsError && methodType.NumOut() > 0 {
				errorInterface := reflect.TypeOf((*error)(nil)).Elem()
				lastReturn := methodType.Out(methodType.NumOut() - 1)
				if lastReturn != errorInterface {
					t.Errorf("%s method should return error as last value, got %s", tc.methodName, lastReturn)
				}
			}
		})
	}
}

// TestMessageIteratorMethodSignatures verifies MessageIterator methods have correct signatures.
func TestMessageIteratorMethodSignatures(t *testing.T) {
	iteratorType := reflect.TypeOf((*MessageIterator)(nil)).Elem()

	testCases := []struct {
		methodName          string
		expectedInputs      int
		expectedOutputs     int
		firstParamIsContext bool
	}{
		{"Next", 1, 2, true},   // func(ctx context.Context) (Message, error)
		{"Close", 0, 1, false}, // func() error
	}

	for _, tc := range testCases {
		t.Run(tc.methodName, func(t *testing.T) {
			method, found := iteratorType.MethodByName(tc.methodName)
			if !found {
				t.Errorf("MessageIterator interface must have %s method", tc.methodName)
				return
			}

			methodType := method.Type

			if methodType.NumIn() != tc.expectedInputs {
				t.Errorf("%s method should have %d inputs, got %d", tc.methodName, tc.expectedInputs, methodType.NumIn())
			}

			if methodType.NumOut() != tc.expectedOutputs {
				t.Errorf("%s method should have %d outputs, got %d", tc.methodName, tc.expectedOutputs, methodType.NumOut())
			}

			// Check context.Context as first parameter
			if tc.firstParamIsContext && methodType.NumIn() > 0 {
				contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
				if methodType.In(0) != contextType {
					t.Errorf("%s method first parameter should be context.Context, got %s", tc.methodName, methodType.In(0))
				}
			}

			// All methods should return error as last value
			if methodType.NumOut() > 0 {
				errorInterface := reflect.TypeOf((*error)(nil)).Elem()
				lastReturn := methodType.Out(methodType.NumOut() - 1)
				if lastReturn != errorInterface {
					t.Errorf("%s method should return error as last value, got %s", tc.methodName, lastReturn)
				}
			}
		})
	}
}

// TestTransportInterfaceNilHandling verifies interfaces handle nil values correctly.
func TestTransportInterfaceNilHandling(t *testing.T) {
	var transport Transport
	var iterator MessageIterator

	if transport != nil {
		t.Error("Nil Transport interface should be nil")
	}
	if iterator != nil {
		t.Error("Nil MessageIterator interface should be nil")
	}
}

// TestContextFirstDesignPattern verifies context-first design pattern.
func TestContextFirstDesignPattern(t *testing.T) {
	transportType := reflect.TypeOf((*Transport)(nil)).Elem()
	iteratorType := reflect.TypeOf((*MessageIterator)(nil)).Elem()
	contextType := reflect.TypeOf((*context.Context)(nil)).Elem()

	// Methods that should have context as first parameter
	contextMethods := []struct {
		interfaceType reflect.Type
		methodName    string
	}{
		{transportType, "Connect"},
		{transportType, "SendMessage"},
		{transportType, "ReceiveMessages"},
		{transportType, "Interrupt"},
		{iteratorType, "Next"},
	}

	for _, cm := range contextMethods {
		method, found := cm.interfaceType.MethodByName(cm.methodName)
		if !found {
			continue // Other tests will catch missing methods
		}

		methodType := method.Type
		if methodType.NumIn() > 0 {
			firstParam := methodType.In(0) // Interface methods don't have receiver
			if firstParam != contextType {
				t.Errorf("%s.%s should have context.Context as first parameter, got %s",
					cm.interfaceType.Name(), cm.methodName, firstParam)
			}
		}
	}
}
