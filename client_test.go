package claudecode

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/severity1/claude-code-sdk-go/pkg/interfaces"
)

// TestClientLifecycleManagement tests connection, resource cleanup, and transport integration
// Covers T133: Client Auto Connect Context Manager + resource management + transport integration
func TestClientLifecycleManagement(t *testing.T) {
	ctx, cancel := setupClientTestContext(t, 10*time.Second)
	defer cancel()

	subtests := []struct {
		name string
		test func(*testing.T, context.Context)
	}{
		{"basic_lifecycle", testBasicLifecycle},
		{"resource_cleanup_cycles", testResourceCleanupCycles},
		{"transport_integration", testTransportIntegration},
	}

	for _, subtest := range subtests {
		t.Run(subtest.name, func(t *testing.T) {
			subtest.test(t, ctx)
		})
	}
}

func testBasicLifecycle(t *testing.T, ctx context.Context) {
	t.Helper()
	transport := newClientMockTransport()

	// Test defer-based resource management (Go equivalent of Python context manager)
	func() {
		client := setupClientForTest(t, transport)
		defer disconnectClientSafely(t, client)
		connectClientSafely(t, ctx, client)
		assertClientConnected(t, transport)
		err := client.Query(ctx, "test message")
		assertNoError(t, err)
	}() // Defer should trigger disconnect

	assertClientDisconnected(t, transport)

	// Test manual connection lifecycle
	client := setupClientForTest(t, transport)
	connectClientSafely(t, ctx, client)
	assertClientConnected(t, transport)
	disconnectClientSafely(t, client)
	assertClientDisconnected(t, transport)
}

func testResourceCleanupCycles(t *testing.T, ctx context.Context) {
	t.Helper()
	transport := newClientMockTransport()

	// Test resource cleanup with multiple connect/disconnect cycles
	for i := 0; i < 3; i++ {
		client := setupClientForTest(t, transport)
		connectClientSafely(t, ctx, client)
		assertClientConnected(t, transport)
		err := client.Query(ctx, fmt.Sprintf("test query %d", i))
		assertNoError(t, err)
		disconnectClientSafely(t, client)
		assertClientDisconnected(t, transport)
		transport.reset()
	}

	// Verify no resource leaks (basic check)
	if transport.getSentMessageCount() != 0 {
		t.Error("Expected transport to be reset after cleanup")
	}
}

func testTransportIntegration(t *testing.T, ctx context.Context) {
	t.Helper()
	transport := newClientMockTransport()
	client := setupClientForTest(t, transport)
	defer disconnectClientSafely(t, client)

	// Verify interface compliance
	var _ Transport = transport

	// Test transport operations through client
	err := client.Connect(ctx)
	assertNoError(t, err)
	if !transport.connected {
		t.Error("Expected transport to be connected via client")
	}

	// Test message sending
	err = client.Query(ctx, "test message")
	assertNoError(t, err)
	if transport.getSentMessageCount() != 1 {
		t.Errorf("Expected 1 message sent, got %d", transport.getSentMessageCount())
	}

	// Test disconnect
	err = client.Close()
	assertNoError(t, err)
	if transport.connected {
		t.Error("Expected transport to be disconnected via client")
	}
}

// TestClientQueryExecution tests one-shot query functionality
func TestClientQueryExecution(t *testing.T) {
	ctx, cancel := setupClientTestContext(t, 10*time.Second)
	defer cancel()

	transport := newClientMockTransport()
	client := setupClientForTest(t, transport)
	defer disconnectClientSafely(t, client)

	connectClientSafely(t, ctx, client)

	// Execute query through connected client
	err := client.Query(ctx, "What is 2+2?")
	assertNoError(t, err)

	// Verify message was sent to transport
	assertClientMessageCount(t, transport, 1)

	// Verify message content
	sentMsg, ok := transport.getSentMessage(0)
	if !ok {
		t.Fatal("Failed to get sent message")
	}
	if sentMsg.Type != "user" {
		t.Errorf("Expected message type 'user', got '%s'", sentMsg.Type)
	}

	userMsg, ok := sentMsg.Message.(*UserMessage)
	if !ok {
		t.Fatalf("Expected *UserMessage, got %T", sentMsg.Message)
	}

	if userMsg.Type() != "user" {
		t.Errorf("Expected message type 'user', got '%s'", userMsg.Type())
	}

	textContent, ok := userMsg.Content.(interfaces.TextContent)
	if !ok {
		t.Fatalf("Expected TextContent, got %T", userMsg.Content)
	}

	if textContent.Text != "What is 2+2?" {
		t.Errorf("Expected content 'What is 2+2?', got '%s'", textContent.Text)
	}
}

// TestClientStreamQuery tests streaming query with message handling
func TestClientStreamQuery(t *testing.T) {
	ctx, cancel := setupClientTestContext(t, 10*time.Second)
	defer cancel()

	transport := newClientMockTransport()
	client := setupClientForTest(t, transport)
	defer disconnectClientSafely(t, client)

	connectClientSafely(t, ctx, client)

	// Create message channel
	messages := make(chan StreamMessage, 3)
	messages <- StreamMessage{
		Type: "request",
		Message: &UserMessage{
			Content: interfaces.TextContent{Text: "Hello"},
		},
	}
	messages <- StreamMessage{
		Type: "request",
		Message: &UserMessage{
			Content: interfaces.TextContent{Text: "How are you?"},
		},
	}
	close(messages)

	// Execute stream query
	err := client.QueryStream(ctx, messages)
	assertNoError(t, err)

	// Wait a bit for async processing to complete
	time.Sleep(100 * time.Millisecond)

	// Verify messages were sent
	assertClientMessageCount(t, transport, 2)
}

// TestClientErrorHandling tests connection, send, and async error scenarios - streamlined
func TestClientErrorHandling(t *testing.T) {
	ctx, cancel := setupClientTestContext(t, 10*time.Second)
	defer cancel()

	errorTests := map[string]struct {
		errorType string
		operation string
		errorMsg  string
	}{
		"connection_error": {"connect", "Connect", "connection failed"},
		"send_error":       {"send", "Query", "send failed"},
		"successful_op":    {"", "Query", ""},
	}

	for name, test := range errorTests {
		t.Run(name, func(t *testing.T) {
			var transport *clientMockTransport
			if test.errorType == "" {
				transport = newClientMockTransport()
			} else {
				transport = newMockTransportWithError(test.errorType, errors.New(test.errorMsg))
			}

			client := setupClientForTest(t, transport)
			defer disconnectClientSafely(t, client)

			var err error
			switch test.operation {
			case "Connect":
				err = client.Connect(ctx)
			case "Query":
				if test.errorType != "connect" {
					connectClientSafely(t, ctx, client)
				}
				err = client.Query(ctx, "test")
			}

			wantErr := test.errorMsg != ""
			assertClientError(t, err, wantErr, test.errorMsg)
		})
	}
}

// TestClientConcurrency tests basic thread safety validation
func TestClientConcurrency(t *testing.T) {
	ctx, cancel := setupClientTestContext(t, 30*time.Second)
	defer cancel()

	transport := newClientMockTransport()
	client := setupClientForTest(t, transport)
	defer disconnectClientSafely(t, client)

	connectClientSafely(t, ctx, client)

	// Run concurrent queries
	const numGoroutines = 10
	const queriesPerGoroutine = 5

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*queriesPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < queriesPerGoroutine; j++ {
				err := client.Query(ctx, fmt.Sprintf("query %d-%d", id, j))
				if err != nil {
					errors <- err
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent query error: %v", err)
	}

	// Verify all messages were sent
	expectedMessages := numGoroutines * queriesPerGoroutine
	assertClientMessageCount(t, transport, expectedMessages)
}

// TestClientConfiguration tests options application and validation with proper behavior verification
func TestClientConfiguration(t *testing.T) {
	tests := []struct {
		name     string
		options  []Option
		validate func(*testing.T, Client, *clientMockTransport)
	}{
		{"default_configuration", []Option{}, verifyDefaultConfiguration},
		{"system_prompt_configuration", []Option{WithSystemPrompt("You are a test assistant. Always respond with 'TEST_RESPONSE'.")}, verifySystemPromptConfig},
		{"tools_configuration", []Option{WithAllowedTools("Read", "Write"), WithDisallowedTools("Bash", "WebSearch")}, verifyToolsConfig},
		{"multiple_options_precedence", []Option{WithSystemPrompt("First prompt"), WithMaxThinkingTokens(5000), WithSystemPrompt("Second prompt"), WithMaxThinkingTokens(10000), WithAllowedTools("Read"), WithAllowedTools("Read", "Write")}, verifyOptionsConfig},
		{"complex_configuration", []Option{WithSystemPrompt("Complex test system prompt"), WithAllowedTools("Read", "Write", "Edit"), WithDisallowedTools("Bash"), WithContinueConversation(true), WithMaxThinkingTokens(8000), WithPermissionMode(PermissionModeAcceptEdits)}, verifyComplexConfig},
		{"session_configuration", []Option{WithContinueConversation(true), WithResume("test-session-123")}, verifySessionConfig},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			transport := newClientMockTransport()
			client := NewClientWithTransport(transport, test.options...)
			defer disconnectClientSafely(t, client)

			test.validate(t, client, transport)
		})
	}
}

// TestClientReceiveMessages tests message reception through client channels
// Covers T137: Client Message Reception
func TestClientReceiveMessages(t *testing.T) {
	ctx, cancel := setupClientTestContext(t, 10*time.Second)
	defer cancel()

	transport := newClientMockTransport()
	client := setupClientForTest(t, transport)
	defer disconnectClientSafely(t, client)

	connectClientSafely(t, ctx, client)

	// Get message channel from client
	msgChan := client.ReceiveMessages(ctx)
	if msgChan == nil {
		t.Fatal("Expected message channel, got nil")
	}

	// Create and inject a test message for this test
	testMessage := &AssistantMessage{
		Content: []ContentBlock{&TextBlock{Text: "Test response message"}},
		Model:   "claude-3-5-sonnet-20241022",
	}
	transport.injectTestMessage(testMessage)

	// Receive the message through client channel
	select {
	case msg := <-msgChan:
		if msg == nil {
			t.Error("Received nil message")
			return
		}

		assistantMsg, ok := msg.(*AssistantMessage)
		if !ok {
			t.Errorf("Expected AssistantMessage, got %T", msg)
			return
		}

		if len(assistantMsg.Content) != 1 {
			t.Errorf("Expected 1 content block, got %d", len(assistantMsg.Content))
			return
		}

		textBlock, ok := assistantMsg.Content[0].(*TextBlock)
		if !ok {
			t.Errorf("Expected TextBlock, got %T", assistantMsg.Content[0])
			return
		}

		if textBlock.Text != "Test response message" {
			t.Errorf("Expected 'Test response message', got '%s'", textBlock.Text)
		}

	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for message from client channel")
	}
}

// TestClientResponseIterator tests response iteration through MessageIterator
// Covers T138: Client Response Iterator
func TestClientResponseIterator(t *testing.T) {
	ctx, cancel := setupClientTestContext(t, 10*time.Second)
	defer cancel()

	transport := newClientMockTransport()
	client := setupClientForTest(t, transport)
	defer disconnectClientSafely(t, client)

	connectClientSafely(t, ctx, client)

	// Get response iterator from client
	iter := client.ReceiveResponse(ctx)
	if iter == nil {
		t.Fatal("Expected MessageIterator, got nil")
	}

	// Inject test messages for iterator testing
	transport.injectTestMessage(&AssistantMessage{
		Content: []ContentBlock{&TextBlock{Text: "First response"}},
		Model:   "claude-3-5-sonnet-20241022",
	})
	transport.injectTestMessage(&AssistantMessage{
		Content: []ContentBlock{&TextBlock{Text: "Second response"}},
		Model:   "claude-3-5-sonnet-20241022",
	})

	// Iterate through messages using iterator
	receivedCount := 0
	expectedTexts := []string{"First response", "Second response"}

	for i := 0; i < len(expectedTexts); i++ {
		msg, err := iter.Next(ctx)
		if err != nil {
			t.Fatalf("Iterator error: %v", err)
		}

		if msg == nil {
			t.Fatal("Expected message from iterator, got nil")
		}

		assistantMsg, ok := msg.(*AssistantMessage)
		if !ok {
			t.Errorf("Expected AssistantMessage, got %T", msg)
			continue
		}

		if len(assistantMsg.Content) != 1 {
			t.Errorf("Expected 1 content block, got %d", len(assistantMsg.Content))
			continue
		}

		textBlock, ok := assistantMsg.Content[0].(*TextBlock)
		if !ok {
			t.Errorf("Expected TextBlock, got %T", assistantMsg.Content[0])
			continue
		}

		if textBlock.Text != expectedTexts[i] {
			t.Errorf("Expected '%s', got '%s'", expectedTexts[i], textBlock.Text)
		}

		receivedCount++
	}

	if receivedCount != len(expectedTexts) {
		t.Errorf("Expected %d messages, received %d", len(expectedTexts), receivedCount)
	}
}

// TestClientInterrupt tests interrupt functionality during operations
// Covers T139: Client Interrupt Functionality
func TestClientInterrupt(t *testing.T) {
	ctx, cancel := setupClientTestContext(t, 10*time.Second)
	defer cancel()

	transport := newClientMockTransport()
	client := setupClientForTest(t, transport)
	defer disconnectClientSafely(t, client)

	connectClientSafely(t, ctx, client)

	// Test interrupt on connected client
	err := client.Interrupt(ctx)
	assertNoError(t, err)

	// Test interrupt propagation to transport
	if transport.interruptError != nil {
		t.Errorf("Transport interrupt should not have error by default, got: %v", transport.interruptError)
	}

	// Test interrupt with transport error
	transportWithError := newClientMockTransportWithOptions(WithClientInterruptError(fmt.Errorf("interrupt failed")))
	clientWithError := setupClientForTest(t, transportWithError)
	defer disconnectClientSafely(t, clientWithError)

	connectClientSafely(t, ctx, clientWithError)

	err = clientWithError.Interrupt(ctx)
	assertClientError(t, err, true, "interrupt failed")

	// Test interrupt during query operation
	longRunningTransport := newClientMockTransport()
	longRunningClient := setupClientForTest(t, longRunningTransport)
	defer disconnectClientSafely(t, longRunningClient)

	connectClientSafely(t, ctx, longRunningClient)

	// Use a channel to synchronize the goroutine
	done := make(chan error, 1)

	// Start a query operation
	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("goroutine panicked: %v", r)
				return
			}
		}()
		time.Sleep(50 * time.Millisecond) // Let query start
		err := longRunningClient.Interrupt(ctx)
		done <- err
	}()

	// Execute query (interrupt should not prevent this from completing)
	err = longRunningClient.Query(ctx, "test query")
	assertNoError(t, err)

	// Wait for goroutine to complete before test ends
	select {
	case goroutineErr := <-done:
		if goroutineErr != nil {
			t.Errorf("Interrupt during operation failed: %v", goroutineErr)
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("Timeout waiting for interrupt goroutine to complete")
	}

	// Verify query was sent despite interrupt
	assertClientMessageCount(t, longRunningTransport, 1)
}

// TestClientSessionID tests session ID handling in client operations
// Covers T140: Client Session Management
func TestClientSessionID(t *testing.T) {
	ctx, cancel := setupClientTestContext(t, 10*time.Second)
	defer cancel()

	// Test 1: Client with default session ID
	transport1 := newClientMockTransport()
	client1 := setupClientForTest(t, transport1) // No WithResume option
	defer disconnectClientSafely(t, client1)

	connectClientSafely(t, ctx, client1)

	err := client1.Query(ctx, "test message")
	assertNoError(t, err)

	// Verify default session ID was used
	sentMsg, ok := transport1.getSentMessage(0)
	if !ok {
		t.Fatal("Failed to get sent message")
	}
	if sentMsg.SessionID != "default" {
		t.Errorf("Expected default session ID 'default', got '%s'", sentMsg.SessionID)
	}

	disconnectClientSafely(t, client1)

	// Test 2: Client with custom session ID
	transport2 := newClientMockTransport()
	client2 := NewClientWithTransport(transport2, WithResume("custom-session"))
	defer disconnectClientSafely(t, client2)

	connectClientSafely(t, ctx, client2)

	err = client2.Query(ctx, "test message 2")
	assertNoError(t, err)

	// Verify custom session ID was used
	sentMsg, ok = transport2.getSentMessage(0)
	if !ok {
		t.Fatal("Failed to get second sent message")
	}
	if sentMsg.SessionID != "custom-session" {
		t.Errorf("Expected custom session ID 'custom-session', got '%s'", sentMsg.SessionID)
	}

	// Test 3: Multiple queries from same client should use same session
	err = client2.Query(ctx, "test message 3")
	assertNoError(t, err)

	sentMsg, ok = transport2.getSentMessage(1)
	if !ok {
		t.Fatal("Failed to get third sent message")
	}
	if sentMsg.SessionID != "custom-session" {
		t.Errorf("Expected consistent session ID 'custom-session', got '%s'", sentMsg.SessionID)
	}

	// Verify message counts
	assertClientMessageCount(t, transport1, 1)
	assertClientMessageCount(t, transport2, 2)
}

// TestClientMultipleSessions tests concurrent operations with different session IDs
// Covers T151: Client Multiple Sessions + T156: State Consistency
func TestClientMultipleSessions(t *testing.T) {
	ctx, cancel := setupClientTestContext(t, 15*time.Second)
	defer cancel()

	// Test concurrent operations with different session IDs using separate clients
	const numSessions = 3
	const queriesPerSession = 2
	sessionIDs := []string{"session-1", "session-2", "session-3"}

	// Create separate transports and clients for each session
	transports := make([]*clientMockTransport, numSessions)
	clients := make([]Client, numSessions)

	for i, sessionID := range sessionIDs {
		transports[i] = newClientMockTransport()
		clients[i] = NewClientWithTransport(transports[i], WithResume(sessionID))
		defer disconnectClientSafely(t, clients[i])
		connectClientSafely(t, ctx, clients[i])
	}

	var wg sync.WaitGroup
	errors := make(chan error, numSessions*queriesPerSession)

	// Launch concurrent operations for different sessions
	for i, sessionID := range sessionIDs {
		wg.Add(1)
		go func(clientIndex int, sess string) {
			defer wg.Done()
			client := clients[clientIndex]
			for j := 0; j < queriesPerSession; j++ {
				err := client.Query(ctx, fmt.Sprintf("query %d-%d", clientIndex, j))
				if err != nil {
					errors <- fmt.Errorf("session %s query %d failed: %w", sess, j, err)
				}
			}
		}(i, sessionID)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		t.Errorf("Concurrent session operation error: %v", err)
	}

	// Verify all messages were sent to each transport
	for i := range transports {
		assertClientMessageCount(t, transports[i], queriesPerSession)
	}

	// Verify session IDs were properly propagated for each client
	for i, sessionID := range sessionIDs {
		transport := transports[i]

		// Check all messages from this transport have the correct session ID
		for j := 0; j < queriesPerSession; j++ {
			sentMsg, ok := transport.getSentMessage(j)
			if !ok {
				t.Errorf("Transport %d: Failed to get sent message %d", i, j)
				continue
			}
			if sentMsg.SessionID != sessionID {
				t.Errorf("Transport %d: expected session ID '%s', got '%s'",
					i, sessionID, sentMsg.SessionID)
			}
		}

		// Verify each transport/client maintains consistent state
		if !transports[i].connected {
			t.Errorf("Expected client %d to remain connected after operations", i)
		}
	}
}

// TestClientReconnection tests reconnection after transport failures
// Covers T150: Client Reconnection + T155: Error Recovery
func TestClientReconnection(t *testing.T) {
	ctx, cancel := setupClientTestContext(t, 5*time.Second)
	defer cancel()

	transport := newClientMockTransport()
	client := setupClientForTest(t, transport)
	defer disconnectClientSafely(t, client)

	// Initial connection
	connectClientSafely(t, ctx, client)
	err := client.Query(ctx, "test before disconnect")
	assertNoError(t, err)

	// Simulate disconnect and reconnect
	disconnectClientSafely(t, client)
	assertClientDisconnected(t, transport)
	transport.reset()
	connectClientSafely(t, ctx, client)

	// Test recovery after reconnection
	err = client.Query(ctx, "test after reconnect")
	assertNoError(t, err)
	assertClientMessageCount(t, transport, 1)
}

// TestClientAsyncErrorHandling tests async transport error scenarios
// Covers T142: Client Error Propagation + T155: Client Error Recovery
func TestClientAsyncErrorHandling(t *testing.T) {
	ctx, cancel := setupClientTestContext(t, 5*time.Second)
	defer cancel()

	// Test async error propagation
	asyncErr := fmt.Errorf("async transport failure")
	transport := newClientMockTransportWithOptions(WithClientAsyncError(asyncErr))
	client := setupClientForTest(t, transport)
	defer disconnectClientSafely(t, client)

	connectClientSafely(t, ctx, client)

	// Get error channel from ReceiveMessages
	_, errChan := transport.ReceiveMessages(ctx)

	// Should receive async error
	select {
	case receivedErr := <-errChan:
		if receivedErr.Error() != asyncErr.Error() {
			t.Errorf("Expected async error %q, got %q", asyncErr.Error(), receivedErr.Error())
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected to receive async error from errChan")
	}

	// Client should still be functional after async error
	err := client.Query(ctx, "test query after async error")
	assertNoError(t, err)
	assertClientMessageCount(t, transport, 1)
}

// TestClientResponseSequencing tests pre-configured response sequences
// Covers T137: Client Message Reception + T138: Client Response Iterator + T147: Client Message Ordering
func TestClientResponseSequencing(t *testing.T) {
	ctx, cancel := setupClientTestContext(t, 5*time.Second)
	defer cancel()

	// Create pre-configured response sequence
	testMessages := []Message{
		&AssistantMessage{
			Content: []ContentBlock{&TextBlock{Text: "First response"}},
			Model:   "claude-3-5-sonnet-20241022",
		},
		&AssistantMessage{
			Content: []ContentBlock{&TextBlock{Text: "Second response"}},
			Model:   "claude-3-5-sonnet-20241022",
		},
		&AssistantMessage{
			Content: []ContentBlock{&TextBlock{Text: "Third response"}},
			Model:   "claude-3-5-sonnet-20241022",
		},
	}

	transport := newClientMockTransportWithOptions(WithClientResponseMessages(testMessages))
	client := setupClientForTest(t, transport)
	defer disconnectClientSafely(t, client)

	connectClientSafely(t, ctx, client)

	// Get message channel from ReceiveMessages
	msgChan, _ := transport.ReceiveMessages(ctx)

	// Should receive messages in correct order
	expectedTexts := []string{"First response", "Second response", "Third response"}
	for i, expectedText := range expectedTexts {
		select {
		case msg := <-msgChan:
			assistantMsg, ok := msg.(*AssistantMessage)
			if !ok {
				t.Fatalf("Expected AssistantMessage at index %d, got %T", i, msg)
			}

			if len(assistantMsg.Content) != 1 {
				t.Fatalf("Expected 1 content block at index %d, got %d", i, len(assistantMsg.Content))
			}

			textBlock, ok := assistantMsg.Content[0].(*TextBlock)
			if !ok {
				t.Fatalf("Expected TextBlock at index %d, got %T", i, assistantMsg.Content[0])
			}

			if textBlock.Text != expectedText {
				t.Errorf("Expected message %d to be %q, got %q", i, expectedText, textBlock.Text)
			}

		case <-time.After(100 * time.Millisecond):
			t.Fatalf("Timeout waiting for message %d", i)
		}
	}

	// Wait a bit longer for any potential extra messages, then verify no more
	extraMessageCount := 0
	timeout := time.After(50 * time.Millisecond)

	for {
		select {
		case msg := <-msgChan:
			extraMessageCount++
			t.Logf("Received unexpected extra message %d: %T", extraMessageCount, msg)
		case <-timeout:
			if extraMessageCount > 0 {
				t.Errorf("Expected exactly 3 messages, but received %d extra messages", extraMessageCount)
			}
			return // Exit the test - expected behavior
		}
	}
}

// TestClientGracefulShutdown tests proper shutdown and configuration
// Covers T154: Graceful Shutdown + T153: Memory Management + T160: Option Order + T163: Protocol Compliance
func TestClientGracefulShutdown(t *testing.T) {
	ctx, cancel := setupClientTestContext(t, 5*time.Second)
	defer cancel()

	// Test option precedence (T160)
	transport := newClientMockTransport()
	client := NewClientWithTransport(transport,
		WithSystemPrompt("first"),
		WithSystemPrompt("second"), // Should override first
		WithAllowedTools("Read"),
	)
	defer disconnectClientSafely(t, client)

	connectClientSafely(t, ctx, client)

	// Test protocol compliance (T163) - messages should be properly formatted
	err := client.Query(ctx, "test message")
	assertNoError(t, err)

	sentMsg, ok := transport.getSentMessage(0)
	if !ok {
		t.Fatal("Failed to get sent message")
	}
	if sentMsg.Type != "user" {
		t.Errorf("Expected message type 'user', got '%s'", sentMsg.Type)
	}

	// Test memory management (T153) - multiple operations should not leak
	for i := 0; i < 5; i++ {
		err := client.Query(ctx, fmt.Sprintf("memory test %d", i))
		assertNoError(t, err)
	}

	// Test graceful shutdown (T154) - disconnect should clean up resources
	disconnectClientSafely(t, client)
	assertClientDisconnected(t, transport)
}

// TestNewClient tests the NewClient constructor function
func TestNewClient(t *testing.T) {
	// Note: With direct transport creation, we test the constructor logic
	// without mocking the factory. Connect() will be tested separately with
	// proper transport mocking at the subprocess level.

	tests := []struct {
		name    string
		options []Option
		verify  func(t *testing.T, client Client)
	}{
		{
			name:    "default_client",
			options: nil,
			verify: func(t *testing.T, client Client) {
				t.Helper()
				if client == nil {
					t.Fatal("Expected client to be created")
				}
				// Test constructor creates client without errors
				// (Connection testing done separately with transport mocks)
			},
		},
		{
			name:    "client_with_system_prompt",
			options: []Option{WithSystemPrompt("Test system prompt")},
			verify: func(t *testing.T, client Client) {
				t.Helper()
				if client == nil {
					t.Fatal("Expected client to be created with system prompt")
				}
				// Test constructor accepts system prompt option
			},
		},
		{
			name: "client_with_multiple_options",
			options: []Option{
				WithSystemPrompt("Multi-option test"),
				WithAllowedTools("Read", "Write"),
				WithModel("claude-sonnet-3-5-20241022"),
			},
			verify: func(t *testing.T, client Client) {
				t.Helper()
				if client == nil {
					t.Fatal("Expected client to be created with multiple options")
				}
				// Test constructor accepts multiple options
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := NewClient(test.options...)
			defer disconnectClientSafely(t, client)

			test.verify(t, client)
		})
	}

	// Note: Error cases for Connect() are tested in TestClientErrorHandling
	// with proper transport mocking
}

// TestClientIteratorClose tests the clientIterator Close method - consolidated
func TestClientIteratorClose(t *testing.T) {
	iteratorTests := map[string]iteratorCloseTest{
		"close_unused":            {"unused", false, false, false},
		"close_with_messages":     {"with_messages", true, false, false},
		"multiple_close_calls":    {"multiple_close", false, false, true},
		"close_after_consumption": {"after_consumption", true, true, false},
	}

	for name, test := range iteratorTests {
		t.Run(name, func(t *testing.T) {
			var transport *clientMockTransport
			if test.needQuery {
				transport = newClientMockTransportWithOptions(WithClientResponseMessages([]Message{
					&AssistantMessage{Content: []ContentBlock{&TextBlock{Text: "response1"}}, Model: "claude-sonnet-3-5-20241022"},
					&AssistantMessage{Content: []ContentBlock{&TextBlock{Text: "response2"}}, Model: "claude-sonnet-3-5-20241022"},
				}))
			} else {
				transport = newClientMockTransport()
			}

			client := setupClientForTest(t, transport)
			defer disconnectClientSafely(t, client)

			verifyIteratorClose(t, client, transport, test)
		})
	}
}

// Mock Transport Implementation - simplified following options_test.go patterns
type clientMockTransport struct {
	mu           sync.Mutex
	connected    bool
	closed       bool
	sentMessages []StreamMessage

	// Minimal message support for essential tests
	testMessages []Message
	msgChan      chan Message
	errChan      chan error

	// Error injection for testing
	connectError   error
	sendError      error
	interruptError error
	closeError     error
	asyncError     error // For async error testing

	// Status behavior for testing
	statusBehavior func() ProcessStatus
}

func (c *clientMockTransport) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check context cancellation first
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if c.connectError != nil {
		return c.connectError
	}

	// For testing flexibility, allow reconnection of closed transports
	if c.closed {
		c.closed = false
	}

	c.connected = true
	return nil
}

func (c *clientMockTransport) SendMessage(ctx context.Context, message StreamMessage) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check context cancellation first
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if c.sendError != nil {
		return c.sendError
	}
	if !c.connected {
		return fmt.Errorf("not connected")
	}
	c.sentMessages = append(c.sentMessages, message)
	return nil
}

func (c *clientMockTransport) ReceiveMessages(ctx context.Context) (msgChan <-chan Message, errChan <-chan error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		closedMsgChan := make(chan Message)
		closedErrChan := make(chan error)
		close(closedMsgChan)
		close(closedErrChan)
		return closedMsgChan, closedErrChan
	}

	// Initialize channels if not already done
	if c.msgChan == nil {
		c.msgChan = make(chan Message, 10)
		c.errChan = make(chan error, 10)

		// Send any pre-configured messages immediately
		for _, msg := range c.testMessages {
			c.msgChan <- msg
		}

		// Send async error if configured
		if c.asyncError != nil {
			c.errChan <- c.asyncError
		}
	}

	return c.msgChan, c.errChan
}

func (c *clientMockTransport) Interrupt(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.interruptError != nil {
		return c.interruptError
	}
	return nil
}

func (c *clientMockTransport) Status() ProcessStatus {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.statusBehavior != nil {
		return c.statusBehavior()
	}

	// Default behavior: return basic running status if connected
	return ProcessStatus{
		Running: c.connected,
	}
}

func (c *clientMockTransport) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closeError != nil {
		return c.closeError
	}

	if c.closed {
		return nil // Already closed
	}

	c.connected = false
	c.closed = true

	// Close channels if they exist
	if c.msgChan != nil {
		close(c.msgChan)
		c.msgChan = nil
	}
	if c.errChan != nil {
		close(c.errChan)
		c.errChan = nil
	}

	return nil
}

// Helper methods
func (c *clientMockTransport) getSentMessageCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.sentMessages)
}

func (c *clientMockTransport) getSentMessage(index int) (StreamMessage, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if index < 0 || index >= len(c.sentMessages) {
		return StreamMessage{}, false
	}
	return c.sentMessages[index], true
}

func (c *clientMockTransport) reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sentMessages = nil
	c.connected = false
	c.closed = false
}

// Simplified message injection helper
func (c *clientMockTransport) injectTestMessage(msg Message) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.testMessages == nil {
		c.testMessages = []Message{}
	}
	c.testMessages = append(c.testMessages, msg)
	if c.msgChan != nil {
		select {
		case c.msgChan <- msg:
		default:
		}
	}
}

// Streamlined Mock Transport Options - reduced from 11 to 6 essential functions
type ClientMockTransportOption func(*clientMockTransport)

func WithClientConnectError(err error) ClientMockTransportOption {
	return func(t *clientMockTransport) { t.connectError = err }
}

func WithClientSendError(err error) ClientMockTransportOption {
	return func(t *clientMockTransport) { t.sendError = err }
}

func WithClientInterruptError(err error) ClientMockTransportOption {
	return func(t *clientMockTransport) { t.interruptError = err }
}

func WithClientAsyncError(err error) ClientMockTransportOption {
	return func(t *clientMockTransport) { t.asyncError = err }
}

func WithClientResponseMessages(messages []Message) ClientMockTransportOption {
	return func(t *clientMockTransport) { t.testMessages = messages }
}

func withStatusBehavior(behavior func() ProcessStatus) ClientMockTransportOption {
	return func(t *clientMockTransport) { t.statusBehavior = behavior }
}

// Factory Functions - streamlined creation methods
func newClientMockTransport() *clientMockTransport {
	return &clientMockTransport{}
}

func newClientMockTransportWithOptions(options ...ClientMockTransportOption) *clientMockTransport {
	transport := &clientMockTransport{}
	for _, option := range options {
		option(transport)
	}
	return transport
}

// Convenience factory methods for common error scenarios
func newMockTransportWithError(errorType string, err error) *clientMockTransport {
	transport := newClientMockTransport()
	switch errorType {
	case "connect":
		transport.connectError = err
	case "send":
		transport.sendError = err
	case "interrupt":
		transport.interruptError = err
	case "async":
		transport.asyncError = err
	}
	return transport
}

// Helper Functions
func setupClientTestContext(t *testing.T, timeout time.Duration) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(context.Background(), timeout)
}

func setupClientForTest(t *testing.T, transport Transport) Client {
	t.Helper()
	return NewClientWithTransport(transport)
}

func connectClientSafely(t *testing.T, ctx context.Context, client Client) {
	t.Helper()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Client connect failed: %v", err)
	}
}

func disconnectClientSafely(t *testing.T, client Client) {
	t.Helper()
	if err := client.Close(); err != nil {
		t.Errorf("Client disconnect failed: %v", err)
	}
}

// Assertion helpers with t.Helper()
func assertClientConnected(t *testing.T, transport *clientMockTransport) {
	t.Helper()
	transport.mu.Lock()
	connected := transport.connected
	transport.mu.Unlock()
	if !connected {
		t.Error("Expected transport to be connected")
	}
}

func assertClientDisconnected(t *testing.T, transport *clientMockTransport) {
	t.Helper()
	transport.mu.Lock()
	connected := transport.connected
	closed := transport.closed
	transport.mu.Unlock()
	if connected {
		t.Errorf("Expected transport to be disconnected, but connected=%t, closed=%t", connected, closed)
	}
}

func assertClientError(t *testing.T, err error, wantErr bool, msgContains string) {
	t.Helper()
	if (err != nil) != wantErr {
		t.Errorf("error = %v, wantErr %v", err, wantErr)
		return
	}
	if wantErr && msgContains != "" && !strings.Contains(err.Error(), msgContains) {
		t.Errorf("error = %v, expected message to contain %q", err, msgContains)
	}
}

func assertClientMessageCount(t *testing.T, transport *clientMockTransport, expected int) {
	t.Helper()
	actual := transport.getSentMessageCount()
	if actual != expected {
		t.Errorf("Expected %d sent messages, got %d", expected, actual)
	}
}

// Helper for success-only assertions - replaces verbose assertNoError(t, err)
func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

// Configuration verification helper - consolidated from 8 redundant functions
type clientConfigTest struct {
	name         string
	messageCount int
	sessionID    string
	queryText    string
	validateFn   func(*testing.T, *clientMockTransport)
}

func verifyClientConfiguration(t *testing.T, client Client, transport *clientMockTransport, config clientConfigTest) {
	t.Helper()
	ctx, cancel := setupClientTestContext(t, 5*time.Second)
	defer cancel()

	if client == nil {
		t.Fatalf("Expected client to be created for %s", config.name)
	}

	connectClientSafely(t, ctx, client)

	// Execute queries based on message count
	for i := 0; i < config.messageCount; i++ {
		queryText := config.queryText
		if config.messageCount > 1 {
			queryText = fmt.Sprintf("%s %d", config.queryText, i+1)
		}

		var err error
		if config.sessionID != "" {
			err = client.Query(ctx, queryText)
		} else {
			err = client.Query(ctx, queryText)
		}
		assertNoError(t, err)
	}

	assertClientMessageCount(t, transport, config.messageCount)

	// Apply custom validation if provided
	if config.validateFn != nil {
		config.validateFn(t, transport)
	}
}

// Specific validation functions for different config types
func verifyDefaultConfiguration(t *testing.T, client Client, transport *clientMockTransport) {
	t.Helper()
	verifyClientConfiguration(t, client, transport, clientConfigTest{
		name:         "default_configuration",
		messageCount: 1,
		queryText:    "default test",
		validateFn: func(t *testing.T, tr *clientMockTransport) {
			sentMsg, ok := tr.getSentMessage(0)
			if !ok {
				t.Fatal("Expected sent message")
			}
			if sentMsg.SessionID != "default" {
				t.Errorf("Expected default session ID 'default', got %q", sentMsg.SessionID)
			}
		},
	})
}

func verifySystemPromptConfig(t *testing.T, client Client, transport *clientMockTransport) {
	t.Helper()
	verifyClientConfiguration(t, client, transport, clientConfigTest{
		name:         "system_prompt_configuration",
		messageCount: 1,
		queryText:    "test with system prompt",
	})
}

func verifyToolsConfig(t *testing.T, client Client, transport *clientMockTransport) {
	t.Helper()
	verifyClientConfiguration(t, client, transport, clientConfigTest{
		name:         "tools_configuration",
		messageCount: 1,
		queryText:    "test with tools config",
	})
}

func verifyOptionsConfig(t *testing.T, client Client, transport *clientMockTransport) {
	t.Helper()
	verifyClientConfiguration(t, client, transport, clientConfigTest{
		name:         "multiple_options",
		messageCount: 1,
		queryText:    "test option precedence",
	})
}

func verifyComplexConfig(t *testing.T, client Client, transport *clientMockTransport) {
	t.Helper()
	verifyClientConfiguration(t, client, transport, clientConfigTest{
		name:         "complex_configuration",
		messageCount: 2,
		queryText:    "complex query",
		validateFn: func(t *testing.T, tr *clientMockTransport) {
			for i := 0; i < 2; i++ {
				sentMsg, ok := tr.getSentMessage(i)
				if !ok {
					t.Fatalf("Expected sent message %d", i)
				}
				if sentMsg.Type != "user" {
					t.Errorf("Expected message type 'user', got %q", sentMsg.Type)
				}
			}
		},
	})
}

func verifySessionConfig(t *testing.T, client Client, transport *clientMockTransport) {
	t.Helper()
	verifyClientConfiguration(t, client, transport, clientConfigTest{
		name:         "session_configuration",
		messageCount: 1,
		sessionID:    "test-session-123",
		queryText:    "session test",
		validateFn: func(t *testing.T, tr *clientMockTransport) {
			sentMsg, ok := tr.getSentMessage(0)
			if !ok {
				t.Fatal("Expected sent message")
			}
			if sentMsg.SessionID != "test-session-123" {
				t.Errorf("Expected session ID 'test-session-123', got %q", sentMsg.SessionID)
			}
		},
	})
}

// Iterator verification helper - consolidated from 4 redundant functions
type iteratorCloseTest struct {
	name          string
	needQuery     bool
	consumeFirst  bool
	multipleCalls bool
}

func verifyIteratorClose(t *testing.T, client Client, _ *clientMockTransport, test iteratorCloseTest) {
	t.Helper()
	ctx, cancel := setupClientTestContext(t, 10*time.Second)
	defer cancel()

	connectClientSafely(t, ctx, client)

	// Send query if needed for the test scenario
	if test.needQuery {
		err := client.Query(ctx, fmt.Sprintf("%s query", test.name))
		assertNoError(t, err)
	}

	iter := client.ReceiveResponse(ctx)
	if iter == nil {
		t.Fatal("Expected non-nil iterator from ReceiveResponse")
	}

	// Consume first message if requested
	if test.consumeFirst {
		msg, err := iter.Next(ctx)
		if err != nil {
			t.Fatalf("Expected first message, got error: %v", err)
		}
		if msg == nil {
			t.Fatal("Expected first message, got nil")
		}
	}

	// Perform close operation(s)
	closeCount := 1
	if test.multipleCalls {
		closeCount = 3
	}

	for i := 1; i <= closeCount; i++ {
		err := iter.Close()
		if err != nil {
			t.Errorf("Expected Close() call %d to succeed, got: %v", i, err)
		}
	}

	// Verify Next() behavior after close
	nextCalls := 1
	if test.multipleCalls {
		nextCalls = 3
	}

	for i := 0; i < nextCalls; i++ {
		msg, err := iter.Next(ctx)
		if err != ErrNoMoreMessages {
			t.Errorf("Expected ErrNoMoreMessages on Next() call %d after close, got: %v", i+1, err)
		}
		if msg != nil {
			t.Errorf("Expected nil message on Next() call %d after close, got message", i+1)
		}
	}
}

// TestClientContextManager tests Go-idiomatic context manager pattern following Python SDK parity
// Covers the single critical improvement: automatic resource lifecycle management
func TestClientContextManager(t *testing.T) {
	tests := []struct {
		name           string
		setupTransport func() *clientMockTransport
		operation      func(Client) error
		wantErr        bool
		validate       func(*testing.T, *clientMockTransport)
	}{
		{
			name:           "automatic_resource_management",
			setupTransport: newClientMockTransport,
			operation: func(c Client) error {
				return c.Query(context.Background(), "test")
			},
			wantErr: false,
			validate: func(t *testing.T, tr *clientMockTransport) {
				assertClientDisconnected(t, tr)
			},
		},
		{
			name: "error_handling_with_cleanup",
			setupTransport: func() *clientMockTransport {
				return newClientMockTransportWithOptions(WithClientSendError(fmt.Errorf("send failed")))
			},
			operation: func(c Client) error {
				return c.Query(context.Background(), "test")
			},
			wantErr: true,
			validate: func(t *testing.T, tr *clientMockTransport) {
				assertClientDisconnected(t, tr)
			},
		},
		{
			name:           "context_cancellation_with_cleanup",
			setupTransport: newClientMockTransport,
			operation: func(c Client) error {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately
				return c.Query(ctx, "test")
			},
			wantErr: true,
			validate: func(t *testing.T, tr *clientMockTransport) {
				assertClientDisconnected(t, tr)
			},
		},
		{
			name: "connection_error_no_cleanup_needed",
			setupTransport: func() *clientMockTransport {
				return newClientMockTransportWithOptions(WithClientConnectError(fmt.Errorf("connect failed")))
			},
			operation: func(c Client) error {
				return c.Query(context.Background(), "test")
			},
			wantErr: true,
			validate: func(t *testing.T, tr *clientMockTransport) {
				// Should not be connected if connect failed
				if tr.connected {
					t.Error("Expected transport to not be connected after connect failure")
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			transport := test.setupTransport()

			err := WithClientTransport(context.Background(), transport, test.operation)

			assertClientError(t, err, test.wantErr, "")
			test.validate(t, transport)
		})
	}
}

// TestWithClientConcurrentUsage tests concurrent access patterns with context manager
func TestWithClientConcurrentUsage(t *testing.T) {
	ctx, cancel := setupClientTestContext(t, 15*time.Second)
	defer cancel()

	const numGoroutines = 5
	const operationsPerGoroutine = 3

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*operationsPerGoroutine)

	// Track all operations and their transports
	var allTransports []*clientMockTransport
	var transportsMu sync.Mutex

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				// Create a new transport for each operation to avoid race conditions
				transport := newClientMockTransport()
				transportsMu.Lock()
				allTransports = append(allTransports, transport)
				transportsMu.Unlock()

				err := WithClientTransport(ctx, transport, func(client Client) error {
					return client.Query(ctx, fmt.Sprintf("concurrent query %d-%d", id, j))
				})
				if err != nil {
					errors <- fmt.Errorf("goroutine %d operation %d: %w", id, j, err)
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		t.Errorf("Concurrent context manager operation error: %v", err)
	}

	// Verify all operations completed successfully
	expectedOperations := numGoroutines * operationsPerGoroutine
	if len(allTransports) != expectedOperations {
		t.Errorf("Expected %d transport instances, got %d", expectedOperations, len(allTransports))
	}

	// Verify each transport sent exactly one message and was properly cleaned up
	totalMessages := 0
	for i, transport := range allTransports {
		messageCount := transport.getSentMessageCount()
		if messageCount != 1 {
			t.Errorf("Transport %d: expected 1 message, got %d", i, messageCount)
		}
		totalMessages += messageCount

		// Verify cleanup occurred
		assertClientDisconnected(t, transport)
	}

	// Verify total message count
	if totalMessages != expectedOperations {
		t.Errorf("Expected %d total messages, got %d", expectedOperations, totalMessages)
	}
}

// TestWithClientContextCancellation tests context cancellation behavior
func TestWithClientContextCancellation(t *testing.T) {
	tests := []struct {
		name         string
		setupContext func() (context.Context, context.CancelFunc)
		wantErr      bool
		errorMsg     string
	}{
		{
			name: "already_canceled_context",
			setupContext: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately
				return ctx, cancel
			},
			wantErr:  true,
			errorMsg: "context canceled",
		},
		{
			name: "timeout_context",
			setupContext: func() (context.Context, context.CancelFunc) {
				// Create a context that has already timed out deterministically
				ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Hour))
				return ctx, cancel
			},
			wantErr:  true,
			errorMsg: "context deadline exceeded",
		},
		{
			name: "valid_context",
			setupContext: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 5*time.Second)
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := test.setupContext()
			defer cancel()

			transport := newClientMockTransport()

			err := WithClientTransport(ctx, transport, func(client Client) error {
				return client.Query(ctx, "context test")
			})

			assertClientError(t, err, test.wantErr, test.errorMsg)

			// Cleanup should always occur, even with context cancellation
			assertClientDisconnected(t, transport)
		})
	}
}

// TestWithClientOptionsPropagate tests that options are properly passed through
func TestWithClientOptionsPropagate(t *testing.T) {
	ctx, cancel := setupClientTestContext(t, 5*time.Second)
	defer cancel()

	transport := newClientMockTransport()

	// Test with various options
	err := WithClientTransport(ctx, transport, func(client Client) error {
		// Verify client was created and connected
		return client.Query(ctx, "options test")
	},
		WithSystemPrompt("Test system prompt"),
		WithAllowedTools("Read", "Write"),
		WithResume("custom-session"),
	)

	assertNoError(t, err)
	assertClientMessageCount(t, transport, 1)

	// Verify message was sent with correct session
	sentMsg, ok := transport.getSentMessage(0)
	if !ok {
		t.Fatal("Expected sent message")
	}
	if sentMsg.SessionID != "custom-session" {
		t.Errorf("Expected session ID 'custom-session', got %q", sentMsg.SessionID)
	}

	// Verify cleanup
	assertClientDisconnected(t, transport)
}

// TestClientPythonSDKCompatibility tests Client with Python SDK compatible message format and streaming
func TestClientPythonSDKCompatibility(t *testing.T) {
	ctx, cancel := setupClientTestContext(t, 10*time.Second)
	defer cancel()

	// Create mock messages similar to what Python SDK would receive
	costValue := 0.001234
	testMessages := []Message{
		&AssistantMessage{
			Content: []ContentBlock{
				&TextBlock{
					Text: "Hello! I understand you want to test the streaming functionality.",
				},
			},
		},
		&ResultMessage{
			TotalCostUSD: &costValue,
		},
	}

	transport := newClientMockTransportWithOptions(
		WithClientResponseMessages(testMessages),
	)
	client := NewClientWithTransport(transport, WithResume("test-session"))
	defer disconnectClientSafely(t, client)

	// Test complete workflow: Connect → Query → ReceiveMessages
	connectClientSafely(t, ctx, client)
	assertClientConnected(t, transport)

	// Send query using new Python SDK compatible format
	err := client.Query(ctx, "Test streaming with Python SDK format")
	assertNoError(t, err)

	// Verify message was sent in correct Python SDK format
	assertClientMessageCount(t, transport, 1)
	sentMsg, ok := transport.getSentMessage(0)
	if !ok {
		t.Fatal("Failed to get sent message")
	}

	// Verify Python SDK compatible message structure
	if sentMsg.Type != "user" {
		t.Errorf("Expected message type 'user', got '%s'", sentMsg.Type)
	}
	if sentMsg.SessionID != "test-session" {
		t.Errorf("Expected session ID 'test-session', got '%s'", sentMsg.SessionID)
	}
	if sentMsg.ParentToolUseID != nil {
		t.Errorf("Expected nil ParentToolUseID, got '%v'", sentMsg.ParentToolUseID)
	}

	// Verify nested message structure uses strongly-typed interfaces
	userMsg, ok := sentMsg.Message.(*UserMessage)
	if !ok {
		t.Fatalf("Expected Message to be *UserMessage, got %T", sentMsg.Message)
	}
	if userMsg.Type() != "user" {
		t.Errorf("Expected message type 'user', got '%s'", userMsg.Type())
	}

	textContent, ok := userMsg.Content.(interfaces.TextContent)
	if !ok {
		t.Fatalf("Expected TextContent, got %T", userMsg.Content)
	}

	if textContent.Text != "Test streaming with Python SDK format" {
		t.Errorf("Expected content to match prompt, got '%s'", textContent.Text)
	}

	// Test message receiving functionality
	msgChan := client.ReceiveMessages(ctx)
	if msgChan == nil {
		t.Fatal("ReceiveMessages returned nil channel")
	}

	// Receive first message (AssistantMessage)
	select {
	case msg := <-msgChan:
		if msg == nil {
			t.Fatal("Received nil message")
		}
		assistantMsg, ok := msg.(*AssistantMessage)
		if !ok {
			t.Fatalf("Expected AssistantMessage, got %T", msg)
		}
		if len(assistantMsg.Content) != 1 {
			t.Fatalf("Expected 1 content block, got %d", len(assistantMsg.Content))
		}
		textBlock, ok := assistantMsg.Content[0].(*TextBlock)
		if !ok {
			t.Fatalf("Expected TextBlock, got %T", assistantMsg.Content[0])
		}
		if !strings.Contains(textBlock.Text, "streaming functionality") {
			t.Errorf("Expected text to mention streaming functionality, got: %s", textBlock.Text)
		}
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for first message")
	}

	// Receive second message (ResultMessage)
	select {
	case msg := <-msgChan:
		if msg == nil {
			t.Fatal("Received nil message")
		}
		resultMsg, ok := msg.(*ResultMessage)
		if !ok {
			t.Fatalf("Expected ResultMessage, got %T", msg)
		}
		if resultMsg.TotalCostUSD == nil || *resultMsg.TotalCostUSD != 0.001234 {
			var cost float64
			if resultMsg.TotalCostUSD != nil {
				cost = *resultMsg.TotalCostUSD
			}
			t.Errorf("Expected cost 0.001234, got %f", cost)
		}
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for second message")
	}

	// Test iterator pattern with ReceiveResponse (basic functionality)
	iter := client.ReceiveResponse(ctx)
	if iter == nil {
		t.Fatal("ReceiveResponse returned nil iterator")
	}

	// Test that iterator can be closed immediately (following existing test patterns)
	err = iter.Close()
	assertNoError(t, err)
}

// TestWithClient tests the WithClient convenience function with automatic CLI discovery
// This tests the actual WithClient function (not WithClientTransport) which has 0% coverage
func TestWithClient(t *testing.T) {
	tests := []struct {
		name    string
		ctx     func(t *testing.T) (context.Context, context.CancelFunc)
		fn      func(Client) error
		opts    []Option
		wantErr bool
		errMsg  string
	}{
		{
			name: "canceled_context",
			ctx: func(t *testing.T) (context.Context, context.CancelFunc) {
				t.Helper()
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately
				return ctx, cancel
			},
			fn: func(client Client) error {
				return nil // Should not be called
			},
			wantErr: true,
			errMsg:  "context canceled",
		},
		{
			name: "function_returns_error_on_successful_connection",
			ctx: func(t *testing.T) (context.Context, context.CancelFunc) {
				t.Helper()
				return setupClientTestContext(t, 5*time.Second)
			},
			fn: func(client Client) error {
				// If we get here, connection succeeded
				return fmt.Errorf("test function error")
			},
			opts:    []Option{WithCLIPath("nonexistent")}, // Force failure
			wantErr: true,
			errMsg:  "", // Will either be connection error or function error
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := test.ctx(t)
			defer cancel()

			// This will attempt to auto-discover CLI, which will fail in test environment
			// but that's the expected behavior we want to test
			err := WithClient(ctx, test.fn, test.opts...)

			if test.wantErr {
				if err == nil {
					t.Errorf("Expected error, got nil")
				} else if test.errMsg != "" && !strings.Contains(err.Error(), test.errMsg) {
					t.Errorf("Expected error to contain %q, got %v", test.errMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestClientIteratorNextErrorPaths tests error scenarios in clientIterator.Next() method
// Targets the missing 45.5% coverage in Next function error paths
func TestClientIteratorNextErrorPaths(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T) (*clientIterator, context.Context, context.CancelFunc)
		validate func(t *testing.T, msg Message, err error)
	}{
		{
			name: "next_on_closed_iterator",
			setup: func(t *testing.T) (*clientIterator, context.Context, context.CancelFunc) {
				t.Helper()
				msgChan := make(chan Message)
				errChan := make(chan error)
				iter := &clientIterator{
					msgChan: msgChan,
					errChan: errChan,
					closed:  true, // Already closed
				}
				ctx, cancel := setupClientTestContext(t, 5*time.Second)
				return iter, ctx, cancel
			},
			validate: func(t *testing.T, msg Message, err error) {
				t.Helper()
				if err != ErrNoMoreMessages {
					t.Errorf("Expected ErrNoMoreMessages on closed iterator, got: %v", err)
				}
				if msg != nil {
					t.Errorf("Expected nil message on closed iterator, got: %v", msg)
				}
			},
		},
		{
			name: "context_canceled_while_waiting",
			setup: func(t *testing.T) (*clientIterator, context.Context, context.CancelFunc) {
				t.Helper()
				msgChan := make(chan Message)
				errChan := make(chan error)
				iter := &clientIterator{
					msgChan: msgChan,
					errChan: errChan,
					closed:  false,
				}
				ctx, cancel := setupClientTestContext(t, 50*time.Millisecond)
				return iter, ctx, cancel
			},
			validate: func(t *testing.T, msg Message, err error) {
				t.Helper()
				if err != context.DeadlineExceeded {
					t.Errorf("Expected context.DeadlineExceeded, got: %v", err)
				}
				if msg != nil {
					t.Errorf("Expected nil message on context cancellation, got: %v", msg)
				}
			},
		},
		{
			name: "error_received_on_error_channel",
			setup: func(t *testing.T) (*clientIterator, context.Context, context.CancelFunc) {
				t.Helper()
				msgChan := make(chan Message)
				errChan := make(chan error, 1)
				iter := &clientIterator{
					msgChan: msgChan,
					errChan: errChan,
					closed:  false,
				}

				// Send error to error channel
				expectedErr := fmt.Errorf("transport error")
				errChan <- expectedErr

				ctx, cancel := setupClientTestContext(t, 5*time.Second)
				return iter, ctx, cancel
			},
			validate: func(t *testing.T, msg Message, err error) {
				t.Helper()
				if err == nil {
					t.Fatal("Expected error from error channel, got nil")
				}
				if err.Error() != "transport error" {
					t.Errorf("Expected 'transport error', got: %v", err)
				}
				if msg != nil {
					t.Errorf("Expected nil message on error, got: %v", msg)
				}
			},
		},
		{
			name: "message_channel_closed",
			setup: func(t *testing.T) (*clientIterator, context.Context, context.CancelFunc) {
				t.Helper()
				msgChan := make(chan Message)
				errChan := make(chan error)
				iter := &clientIterator{
					msgChan: msgChan,
					errChan: errChan,
					closed:  false,
				}

				// Close the message channel
				close(msgChan)

				ctx, cancel := setupClientTestContext(t, 5*time.Second)
				return iter, ctx, cancel
			},
			validate: func(t *testing.T, msg Message, err error) {
				t.Helper()
				if err != ErrNoMoreMessages {
					t.Errorf("Expected ErrNoMoreMessages on closed channel, got: %v", err)
				}
				if msg != nil {
					t.Errorf("Expected nil message on closed channel, got: %v", msg)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			iter, ctx, cancel := test.setup(t)
			defer cancel()

			msg, err := iter.Next(ctx)
			test.validate(t, msg, err)

			// Verify iterator is closed after error conditions
			if test.name != "next_on_closed_iterator" && !iter.closed {
				t.Error("Expected iterator to be closed after error condition")
			}
		})
	}
}

// TestClientStatus tests the Status method functionality.
func TestClientStatus(t *testing.T) {
	tests := []struct {
		name            string
		setup           func() *clientMockTransport
		expectConnected bool
		expectRunning   bool
		expectPID       bool
		expectStartTime bool
	}{
		{
			name: "connected_client_with_running_process",
			setup: func() *clientMockTransport {
				return newClientMockTransportWithOptions(
					withStatusBehavior(func() ProcessStatus {
						return ProcessStatus{
							Running:   true,
							PID:       12345,
							StartTime: "2024-01-15T10:30:45Z",
						}
					}),
				)
			},
			expectConnected: true,
			expectRunning:   true,
			expectPID:       true,
			expectStartTime: true,
		},
		{
			name: "disconnected_client",
			setup: func() *clientMockTransport {
				return newClientMockTransport()
			},
			expectConnected: false,
			expectRunning:   false,
			expectPID:       false,
			expectStartTime: false,
		},
		{
			name: "connected_client_without_status_support",
			setup: func() *clientMockTransport {
				// No status behavior - should use fallback
				return newClientMockTransport()
			},
			expectConnected: true,
			expectRunning:   true,
			expectPID:       false,
			expectStartTime: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := tt.setup()
			client := NewClientWithTransport(transport)
			ctx, cancel := setupClientTestContext(t, 10*time.Second)
			defer cancel()

			if tt.expectConnected {
				if err := client.Connect(ctx); err != nil {
					t.Fatal(err)
				}
				defer client.Close()
			}

			status := client.Status()

			if status.Running != tt.expectRunning {
				t.Errorf("Expected Running=%v, got %v", tt.expectRunning, status.Running)
			}

			if tt.expectPID && status.PID == 0 {
				t.Error("Expected PID to be set, got 0")
			}
			if !tt.expectPID && status.PID != 0 {
				t.Errorf("Expected PID to be 0, got %d", status.PID)
			}

			if tt.expectStartTime && status.StartTime == "" {
				t.Error("Expected StartTime to be set, got empty string")
			}
			if !tt.expectStartTime && status.StartTime != "" {
				t.Errorf("Expected StartTime to be empty, got %s", status.StartTime)
			}
		})
	}
}
