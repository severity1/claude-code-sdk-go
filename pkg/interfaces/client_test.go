package interfaces

import (
	"context"
	"reflect"
	"testing"
)

// TestClientInterfaceExistence verifies that the client interfaces exist and have correct method signatures.
func TestClientInterfaceExistence(t *testing.T) {
	tests := []struct {
		name            string
		interfaceType   reflect.Type
		expectedMethods []string
	}{
		{
			name:            "ConnectionManager interface",
			interfaceType:   reflect.TypeOf((*ConnectionManager)(nil)).Elem(),
			expectedMethods: []string{"Connect", "Close", "IsConnected"},
		},
		{
			name:            "QueryExecutor interface",
			interfaceType:   reflect.TypeOf((*QueryExecutor)(nil)).Elem(),
			expectedMethods: []string{"Query", "QueryStream"},
		},
		{
			name:            "MessageReceiver interface",
			interfaceType:   reflect.TypeOf((*MessageReceiver)(nil)).Elem(),
			expectedMethods: []string{"ReceiveMessages", "ReceiveResponse"},
		},
		{
			name:            "ProcessController interface",
			interfaceType:   reflect.TypeOf((*ProcessController)(nil)).Elem(),
			expectedMethods: []string{"Interrupt"},
		},
		{
			name:            "Client interface",
			interfaceType:   reflect.TypeOf((*Client)(nil)).Elem(),
			expectedMethods: []string{"Connect", "Close", "IsConnected", "Query", "QueryStream", "ReceiveMessages", "ReceiveResponse", "Interrupt"},
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

// TestClientInterfaceEmbedding verifies Client interface embeds focused interfaces.
func TestClientInterfaceEmbedding(t *testing.T) {
	clientType := reflect.TypeOf((*Client)(nil)).Elem()

	// Client should have methods from all embedded interfaces
	expectedMethods := map[string]string{
		"Connect":         "ConnectionManager",
		"Close":           "ConnectionManager",
		"IsConnected":     "ConnectionManager",
		"Query":           "QueryExecutor",
		"QueryStream":     "QueryExecutor",
		"ReceiveMessages": "MessageReceiver",
		"ReceiveResponse": "MessageReceiver",
		"Interrupt":       "ProcessController",
	}

	for methodName, sourceInterface := range expectedMethods {
		_, found := clientType.MethodByName(methodName)
		if !found {
			t.Errorf("Client interface should have %s method from %s interface", methodName, sourceInterface)
		}
	}
}

// TestConnectionManagerMethodSignatures verifies ConnectionManager methods have correct signatures.
func TestConnectionManagerMethodSignatures(t *testing.T) {
	connType := reflect.TypeOf((*ConnectionManager)(nil)).Elem()
	contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
	errorInterface := reflect.TypeOf((*error)(nil)).Elem()

	testCases := []struct {
		methodName      string
		expectedInputs  int
		expectedOutputs int
		hasContext      bool
		returnsError    bool
	}{
		{"Connect", 1, 1, true, true},       // func(ctx context.Context) error
		{"Close", 0, 1, false, true},        // func() error
		{"IsConnected", 0, 1, false, false}, // func() bool
	}

	for _, tc := range testCases {
		t.Run(tc.methodName, func(t *testing.T) {
			method, found := connType.MethodByName(tc.methodName)
			if !found {
				t.Errorf("ConnectionManager must have %s method", tc.methodName)
				return
			}

			methodType := method.Type

			if methodType.NumIn() != tc.expectedInputs {
				t.Errorf("%s should have %d inputs, got %d", tc.methodName, tc.expectedInputs, methodType.NumIn())
			}

			if methodType.NumOut() != tc.expectedOutputs {
				t.Errorf("%s should have %d outputs, got %d", tc.methodName, tc.expectedOutputs, methodType.NumOut())
			}

			// Check context parameter
			if tc.hasContext && methodType.NumIn() > 0 {
				if methodType.In(0) != contextType {
					t.Errorf("%s first parameter should be context.Context", tc.methodName)
				}
			}

			// Check error return
			if tc.returnsError && methodType.NumOut() > 0 {
				lastReturn := methodType.Out(methodType.NumOut() - 1)
				if lastReturn != errorInterface {
					t.Errorf("%s should return error, got %s", tc.methodName, lastReturn)
				}
			}

			// IsConnected should return bool
			if tc.methodName == "IsConnected" && methodType.NumOut() > 0 {
				if methodType.Out(0).Kind() != reflect.Bool {
					t.Errorf("IsConnected should return bool, got %s", methodType.Out(0).Kind())
				}
			}
		})
	}
}

// TestQueryExecutorMethodSignatures verifies QueryExecutor methods have correct signatures.
func TestQueryExecutorMethodSignatures(t *testing.T) {
	queryType := reflect.TypeOf((*QueryExecutor)(nil)).Elem()
	contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
	errorInterface := reflect.TypeOf((*error)(nil)).Elem()

	testCases := []struct {
		methodName      string
		expectedInputs  int
		expectedOutputs int
	}{
		{"Query", 2, 1},       // func(ctx context.Context, prompt string) error
		{"QueryStream", 2, 1}, // func(ctx context.Context, messages <-chan StreamMessage) error
	}

	for _, tc := range testCases {
		t.Run(tc.methodName, func(t *testing.T) {
			method, found := queryType.MethodByName(tc.methodName)
			if !found {
				t.Errorf("QueryExecutor must have %s method", tc.methodName)
				return
			}

			methodType := method.Type

			// Basic signature check
			if methodType.NumOut() != tc.expectedOutputs {
				t.Errorf("%s should have %d outputs, got %d", tc.methodName, tc.expectedOutputs, methodType.NumOut())
			}

			// Should start with context
			if methodType.NumIn() > 0 && methodType.In(0) != contextType {
				t.Errorf("%s first parameter should be context.Context", tc.methodName)
			}

			// Should return error
			if methodType.NumOut() > 0 {
				lastReturn := methodType.Out(methodType.NumOut() - 1)
				if lastReturn != errorInterface {
					t.Errorf("%s should return error, got %s", tc.methodName, lastReturn)
				}
			}
		})
	}
}

// TestInterfaceSegregationPrinciple verifies interfaces follow SOLID principles.
func TestInterfaceSegregationPrinciple(t *testing.T) {
	// Each focused interface should have a single responsibility
	testCases := []struct {
		name           string
		interfaceType  reflect.Type
		maxMethods     int
		responsibility string
	}{
		{"ConnectionManager", reflect.TypeOf((*ConnectionManager)(nil)).Elem(), 3, "connection lifecycle"},
		{"QueryExecutor", reflect.TypeOf((*QueryExecutor)(nil)).Elem(), 2, "query execution"},
		{"MessageReceiver", reflect.TypeOf((*MessageReceiver)(nil)).Elem(), 2, "message receiving"},
		{"ProcessController", reflect.TypeOf((*ProcessController)(nil)).Elem(), 1, "process control"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			numMethods := tc.interfaceType.NumMethod()
			if numMethods > tc.maxMethods {
				t.Errorf("%s interface has %d methods, should have <= %d for single responsibility (%s)",
					tc.name, numMethods, tc.maxMethods, tc.responsibility)
			}
		})
	}
}

// TestClientInterfaceNilHandling verifies interfaces handle nil values correctly.
func TestClientInterfaceNilHandling(t *testing.T) {
	var connectionManager ConnectionManager
	var queryExecutor QueryExecutor
	var messageReceiver MessageReceiver
	var processController ProcessController
	var client Client

	nilInterfaces := []interface{}{
		connectionManager,
		queryExecutor,
		messageReceiver,
		processController,
		client,
	}

	for i, iface := range nilInterfaces {
		if iface != nil {
			t.Errorf("Nil interface %d should be nil", i)
		}
	}
}

// TestSpecializedInterfaceCombinations verifies specialized combinations work.
func TestSpecializedInterfaceCombinations(t *testing.T) {
	// Test that we can create specialized interface combinations
	var simpleQuerier SimpleQuerier
	var streamClient StreamClient

	if simpleQuerier != nil {
		t.Error("Nil SimpleQuerier should be nil")
	}
	if streamClient != nil {
		t.Error("Nil StreamClient should be nil")
	}

	// Verify SimpleQuerier has only query methods
	simpleQuerierType := reflect.TypeOf((*SimpleQuerier)(nil)).Elem()
	expectedMethods := map[string]bool{"Query": true, "QueryStream": true}

	for i := 0; i < simpleQuerierType.NumMethod(); i++ {
		method := simpleQuerierType.Method(i)
		if !expectedMethods[method.Name] {
			t.Errorf("SimpleQuerier should only have Query/QueryStream methods, found %s", method.Name)
		}
	}
}

// TestContextFirstDesignInClientInterfaces verifies context-first design throughout.
func TestContextFirstDesignInClientInterfaces(t *testing.T) {
	contextType := reflect.TypeOf((*context.Context)(nil)).Elem()

	interfaceMethodMap := map[reflect.Type][]string{
		reflect.TypeOf((*ConnectionManager)(nil)).Elem(): {"Connect"},
		reflect.TypeOf((*QueryExecutor)(nil)).Elem():     {"Query", "QueryStream"},
		reflect.TypeOf((*MessageReceiver)(nil)).Elem():   {"ReceiveMessages", "ReceiveResponse"},
		reflect.TypeOf((*ProcessController)(nil)).Elem(): {"Interrupt"},
	}

	for interfaceType, methods := range interfaceMethodMap {
		for _, methodName := range methods {
			method, found := interfaceType.MethodByName(methodName)
			if !found {
				continue // Other tests will catch missing methods
			}

			methodType := method.Type
			if methodType.NumIn() > 0 {
				firstParam := methodType.In(0) // Interface methods don't have receiver
				if firstParam != contextType {
					t.Errorf("%s.%s should have context.Context as first parameter, got %s",
						interfaceType.Name(), methodName, firstParam)
				}
			}
		}
	}
}
