//go:build integration

package claudecode_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/severity1/claude-code-sdk-go"
	"github.com/severity1/claude-code-sdk-go/pkg/interfaces"
)

// TestClientInterfaceSegregationIntegration tests that ClientImpl implements all segregated interfaces
// This is Phase 6 Day 11-12: Integration Tests for new interface compatibility
// RED: These tests will fail initially since ClientImpl doesn't implement new interfaces
func TestClientInterfaceSegregationIntegration(t *testing.T) {
	ctx, cancel := setupIntegrationTestContext(t, 5*time.Second)
	defer cancel()

	tests := []struct {
		name          string
		interfaceType reflect.Type
		testFn        func(*testing.T, context.Context, claudecode.Client)
	}{
		{
			name:          "ConnectionManager interface compliance",
			interfaceType: reflect.TypeOf((*interfaces.ConnectionManager)(nil)).Elem(),
			testFn:        testConnectionManagerCompliance,
		},
		{
			name:          "QueryExecutor interface compliance",
			interfaceType: reflect.TypeOf((*interfaces.QueryExecutor)(nil)).Elem(),
			testFn:        testQueryExecutorCompliance,
		},
		{
			name:          "MessageReceiver interface compliance",
			interfaceType: reflect.TypeOf((*interfaces.MessageReceiver)(nil)).Elem(),
			testFn:        testMessageReceiverCompliance,
		},
		{
			name:          "ProcessController interface compliance",
			interfaceType: reflect.TypeOf((*interfaces.ProcessController)(nil)).Elem(),
			testFn:        testProcessControllerCompliance,
		},
		{
			name:          "Full Client interface composition",
			interfaceType: reflect.TypeOf((*interfaces.Client)(nil)).Elem(),
			testFn:        testClientInterfaceComposition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create client with mock transport
			transport := newIntegrationMockTransport()
			client := claudecode.NewClientWithTransport(transport)

			// Test interface compliance using reflection
			clientType := reflect.TypeOf(client)
			if !clientType.Implements(tt.interfaceType) {
				t.Errorf("ClientImpl does not implement %s interface", tt.interfaceType.Name())
				return
			}

			// Run specific interface tests
			tt.testFn(t, ctx, client)
		})
	}
}

// testConnectionManagerCompliance tests ConnectionManager interface methods
func testConnectionManagerCompliance(t *testing.T, ctx context.Context, client claudecode.Client) {
	t.Helper()

	// Cast to ConnectionManager interface
	connectionMgr, ok := client.(interfaces.ConnectionManager)
	if !ok {
		t.Fatal("Client does not implement ConnectionManager interface")
	}

	// Test IsConnected before connection
	if connectionMgr.IsConnected() {
		t.Error("Expected IsConnected() to return false before connection")
	}

	// Test Connect
	if err := connectionMgr.Connect(ctx); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Test IsConnected after connection
	if !connectionMgr.IsConnected() {
		t.Error("Expected IsConnected() to return true after connection")
	}

	// Test Close (not Disconnect)
	if err := connectionMgr.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Test IsConnected after close
	if connectionMgr.IsConnected() {
		t.Error("Expected IsConnected() to return false after close")
	}
}

// testQueryExecutorCompliance tests QueryExecutor interface methods
func testQueryExecutorCompliance(t *testing.T, ctx context.Context, client claudecode.Client) {
	t.Helper()

	// Cast to QueryExecutor interface
	queryExec, ok := client.(interfaces.QueryExecutor)
	if !ok {
		t.Fatal("Client does not implement QueryExecutor interface")
	}

	// Connect first (required for queries)
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Test Query with simplified signature (no sessionID parameter)
	if err := queryExec.Query(ctx, "What is 2+2?"); err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	// Test QueryStream
	msgChan := make(chan interfaces.StreamMessage, 1)
	msgChan <- interfaces.StreamMessage{
		Type:    "user",
		Message: &claudecode.UserMessage{Content: interfaces.TextContent{Text: "Stream test"}},
	}
	close(msgChan)

	if err := queryExec.QueryStream(ctx, msgChan); err != nil {
		t.Fatalf("QueryStream failed: %v", err)
	}
}

// testMessageReceiverCompliance tests MessageReceiver interface methods
func testMessageReceiverCompliance(t *testing.T, ctx context.Context, client claudecode.Client) {
	t.Helper()

	// Cast to MessageReceiver interface
	msgReceiver, ok := client.(interfaces.MessageReceiver)
	if !ok {
		t.Fatal("Client does not implement MessageReceiver interface")
	}

	// Connect first
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Test ReceiveMessages returns a channel
	msgChan := msgReceiver.ReceiveMessages(ctx)
	if msgChan == nil {
		t.Fatal("ReceiveMessages returned nil channel")
	}

	// Test ReceiveResponse returns an iterator
	iter := msgReceiver.ReceiveResponse(ctx)
	if iter == nil {
		t.Fatal("ReceiveResponse returned nil iterator")
	}
}

// testProcessControllerCompliance tests ProcessController interface methods
func testProcessControllerCompliance(t *testing.T, ctx context.Context, client claudecode.Client) {
	t.Helper()

	// Cast to ProcessController interface
	procController, ok := client.(interfaces.ProcessController)
	if !ok {
		t.Fatal("Client does not implement ProcessController interface")
	}

	// Connect first
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Test Status returns proper ProcessStatus
	status := procController.Status()
	if !status.Running {
		t.Error("Expected Status().Running to be true when connected")
	}

	// Test Interrupt
	if err := procController.Interrupt(ctx); err != nil {
		t.Fatalf("Interrupt failed: %v", err)
	}
}

// testClientInterfaceComposition tests that Client interface properly composes all interfaces
func testClientInterfaceComposition(t *testing.T, ctx context.Context, client claudecode.Client) {
	t.Helper()

	// Test that Client interface embeds all the segregated interfaces
	clientType := reflect.TypeOf(client)

	expectedInterfaces := []reflect.Type{
		reflect.TypeOf((*interfaces.ConnectionManager)(nil)).Elem(),
		reflect.TypeOf((*interfaces.QueryExecutor)(nil)).Elem(),
		reflect.TypeOf((*interfaces.MessageReceiver)(nil)).Elem(),
		reflect.TypeOf((*interfaces.ProcessController)(nil)).Elem(),
		reflect.TypeOf((*interfaces.Client)(nil)).Elem(),
	}

	for _, iface := range expectedInterfaces {
		if !clientType.Implements(iface) {
			t.Errorf("Client does not implement %s interface", iface.Name())
		}
	}

	// Test that client can be used as any of the segregated interfaces
	var connectionMgr interfaces.ConnectionManager = client
	var queryExec interfaces.QueryExecutor = client
	var msgReceiver interfaces.MessageReceiver = client
	var procController interfaces.ProcessController = client
	var fullClient interfaces.Client = client

	// Use the interfaces (basic smoke test)
	_ = connectionMgr
	_ = queryExec
	_ = msgReceiver
	_ = procController
	_ = fullClient
}

// TestSpecializedInterfacePatterns tests SimpleQuerier and StreamClient patterns
func TestSpecializedInterfacePatterns(t *testing.T) {
	ctx, cancel := setupIntegrationTestContext(t, 5*time.Second)
	defer cancel()

	tests := []struct {
		name          string
		interfaceType reflect.Type
		testFn        func(*testing.T, context.Context, claudecode.Client)
	}{
		{
			name:          "SimpleQuerier pattern",
			interfaceType: reflect.TypeOf((*interfaces.SimpleQuerier)(nil)).Elem(),
			testFn:        testSimpleQuerierPattern,
		},
		{
			name:          "StreamClient pattern",
			interfaceType: reflect.TypeOf((*interfaces.StreamClient)(nil)).Elem(),
			testFn:        testStreamClientPattern,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := newIntegrationMockTransport()
			client := claudecode.NewClientWithTransport(transport)

			// Test interface compliance
			clientType := reflect.TypeOf(client)
			if !clientType.Implements(tt.interfaceType) {
				t.Errorf("ClientImpl does not implement %s interface", tt.interfaceType.Name())
				return
			}

			tt.testFn(t, ctx, client)
		})
	}
}

// testSimpleQuerierPattern tests the SimpleQuerier interface pattern
func testSimpleQuerierPattern(t *testing.T, ctx context.Context, client claudecode.Client) {
	t.Helper()

	// Cast to SimpleQuerier (should only have QueryExecutor methods)
	simpleQuerier, ok := client.(interfaces.SimpleQuerier)
	if !ok {
		t.Fatal("Client does not implement SimpleQuerier interface")
	}

	// Connect first (still need connection for queries)
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Test Query method (should work through SimpleQuerier interface)
	if err := simpleQuerier.Query(ctx, "Simple query test"); err != nil {
		t.Fatalf("SimpleQuerier.Query failed: %v", err)
	}
}

// testStreamClientPattern tests the StreamClient interface pattern
func testStreamClientPattern(t *testing.T, ctx context.Context, client claudecode.Client) {
	t.Helper()

	// Cast to StreamClient (should have ConnectionManager + MessageReceiver)
	streamClient, ok := client.(interfaces.StreamClient)
	if !ok {
		t.Fatal("Client does not implement StreamClient interface")
	}

	// Test connection methods
	if streamClient.IsConnected() {
		t.Error("Expected IsConnected() to return false initially")
	}

	if err := streamClient.Connect(ctx); err != nil {
		t.Fatalf("StreamClient.Connect failed: %v", err)
	}

	if !streamClient.IsConnected() {
		t.Error("Expected IsConnected() to return true after connect")
	}

	// Test message receiving methods
	msgChan := streamClient.ReceiveMessages(ctx)
	if msgChan == nil {
		t.Fatal("ReceiveMessages returned nil channel")
	}

	iter := streamClient.ReceiveResponse(ctx)
	if iter == nil {
		t.Fatal("ReceiveResponse returned nil iterator")
	}

	// Cleanup
	if err := streamClient.Close(); err != nil {
		t.Fatalf("StreamClient.Close failed: %v", err)
	}
}

// Helper function to setup integration test context
func setupIntegrationTestContext(t *testing.T, timeout time.Duration) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(context.Background(), timeout)
}
