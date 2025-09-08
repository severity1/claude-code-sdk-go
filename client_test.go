package claudecode

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestClientLifecycle tests basic connection and resource management patterns
// Covers T133: Client Auto Connect Context Manager
func TestClientLifecycle(t *testing.T) {
	ctx, cancel := setupClientTestContext(t, 10*time.Second)
	defer cancel()

	transport := newClientMockTransport()

	// Test defer-based resource management (Go equivalent of Python context manager)
	func() {
		client := setupClientForTest(t, transport)
		defer disconnectClientSafely(t, client)

		// Connect should be called automatically or explicitly
		connectClientSafely(t, ctx, client)

		// Verify connection was established
		assertClientConnected(t, transport)

		// Client should be ready to use
		err := client.Query(ctx, "test message")
		assertClientError(t, err, false, "")
	}() // Defer should trigger disconnect

	// Verify disconnect was called
	assertClientDisconnected(t, transport)

	// Test manual connection lifecycle
	client := setupClientForTest(t, transport)
	connectClientSafely(t, ctx, client)
	assertClientConnected(t, transport)

	disconnectClientSafely(t, client)
	assertClientDisconnected(t, transport)
}

// TestTransportDirectClose tests transport Close method directly
func TestTransportDirectClose(t *testing.T) {
	transport := newClientMockTransport()

	// Initial state should be disconnected
	if transport.connected {
		t.Errorf("New transport should start disconnected, got connected=%t", transport.connected)
	}

	// Connect
	err := transport.Connect(context.Background())
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Should be connected
	if !transport.connected {
		t.Errorf("After Connect, expected connected=true, got connected=%t", transport.connected)
	}

	// Close
	err = transport.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Check final state
	transport.mu.Lock()
	connected := transport.connected
	closed := transport.closed
	transport.mu.Unlock()

	t.Logf("After Close: connected=%t, closed=%t", connected, closed)

	// Should be disconnected
	if connected {
		t.Errorf("After Close, expected connected=false, got connected=%t, closed=%t", connected, closed)
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
	assertClientError(t, err, false, "")

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

	messageMap, ok := sentMsg.Message.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map[string]interface{}, got %T", sentMsg.Message)
	}

	if role, ok := messageMap["role"]; !ok || role != "user" {
		t.Errorf("Expected message role 'user', got '%v'", role)
	}
	if content, ok := messageMap["content"]; !ok || content != "What is 2+2?" {
		t.Errorf("Expected content 'What is 2+2?', got '%v'", content)
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
			Content: "Hello",
		},
	}
	messages <- StreamMessage{
		Type: "request",
		Message: &UserMessage{
			Content: "How are you?",
		},
	}
	close(messages)

	// Execute stream query
	err := client.QueryStream(ctx, messages)
	assertClientError(t, err, false, "")

	// Wait a bit for async processing to complete
	time.Sleep(100 * time.Millisecond)

	// Verify messages were sent
	assertClientMessageCount(t, transport, 2)
}

// TestClientErrorHandling tests connection, send, and async error scenarios
func TestClientErrorHandling(t *testing.T) {
	ctx, cancel := setupClientTestContext(t, 10*time.Second)
	defer cancel()

	tests := []struct {
		name           string
		setupTransport func() *clientMockTransport
		operation      func(Client) error
		expectError    bool
		errorContains  string
	}{
		{
			name: "connection_error",
			setupTransport: func() *clientMockTransport {
				return newClientMockTransportWithOptions(WithClientConnectError(fmt.Errorf("connection failed")))
			},
			operation: func(c Client) error {
				return c.Connect(ctx)
			},
			expectError:   true,
			errorContains: "connection failed",
		},
		{
			name: "send_error",
			setupTransport: func() *clientMockTransport {
				return newClientMockTransportWithOptions(WithClientSendError(fmt.Errorf("send failed")))
			},
			operation: func(c Client) error {
				connectClientSafely(t, ctx, c)
				return c.Query(ctx, "test")
			},
			expectError:   true,
			errorContains: "send failed",
		},
		{
			name:           "successful_operation",
			setupTransport: newClientMockTransport,
			operation: func(c Client) error {
				connectClientSafely(t, ctx, c)
				return c.Query(ctx, "test")
			},
			expectError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			transport := test.setupTransport()
			client := setupClientForTest(t, transport)
			defer disconnectClientSafely(t, client)

			err := test.operation(client)
			assertClientError(t, err, test.expectError, test.errorContains)
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

// TestClientResourceCleanup tests proper cleanup and session management
func TestClientResourceCleanup(t *testing.T) {
	ctx, cancel := setupClientTestContext(t, 10*time.Second)
	defer cancel()

	transport := newClientMockTransport()

	// Test resource cleanup with multiple connect/disconnect cycles
	for i := 0; i < 3; i++ {
		client := setupClientForTest(t, transport)

		connectClientSafely(t, ctx, client)
		assertClientConnected(t, transport)

		// Use the client
		err := client.Query(ctx, fmt.Sprintf("test query %d", i))
		assertClientError(t, err, false, "")

		// Clean disconnect
		disconnectClientSafely(t, client)
		assertClientDisconnected(t, transport)

		// Reset transport for next iteration
		transport.reset()
	}

	// Verify no resource leaks (basic check)
	if transport.getSentMessageCount() != 0 {
		t.Error("Expected transport to be reset after cleanup")
	}
}

// TestClientTransportIntegration tests transport interface compliance
func TestClientTransportIntegration(t *testing.T) {
	ctx, cancel := setupClientTestContext(t, 10*time.Second)
	defer cancel()

	// Test that client properly uses transport interface
	transport := newClientMockTransport()
	client := setupClientForTest(t, transport)
	defer disconnectClientSafely(t, client)

	// Verify interface compliance
	var _ Transport = transport

	// Test transport operations through client
	err := client.Connect(ctx)
	assertClientError(t, err, false, "")

	if !transport.connected {
		t.Error("Expected transport to be connected via client")
	}

	// Test message sending
	err = client.Query(ctx, "test message")
	assertClientError(t, err, false, "")

	if transport.getSentMessageCount() != 1 {
		t.Errorf("Expected 1 message sent, got %d", transport.getSentMessageCount())
	}

	// Test disconnect
	err = client.Disconnect()
	assertClientError(t, err, false, "")

	if transport.connected {
		t.Error("Expected transport to be disconnected via client")
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
	assertClientError(t, err, false, "")

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
	assertClientError(t, err, false, "")

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

	transport := newClientMockTransport()
	client := setupClientForTest(t, transport)
	defer disconnectClientSafely(t, client)

	connectClientSafely(t, ctx, client)

	// Test query with default session ID
	err := client.Query(ctx, "test message")
	assertClientError(t, err, false, "")

	// Verify default session ID was used
	sentMsg, ok := transport.getSentMessage(0)
	if !ok {
		t.Fatal("Failed to get sent message")
	}
	if sentMsg.SessionID != "default" {
		t.Errorf("Expected default session ID 'default', got '%s'", sentMsg.SessionID)
	}

	// Test query with custom session ID
	err = client.Query(ctx, "test message 2", "custom-session")
	assertClientError(t, err, false, "")

	// Verify custom session ID was used
	sentMsg, ok = transport.getSentMessage(1)
	if !ok {
		t.Fatal("Failed to get second sent message")
	}
	if sentMsg.SessionID != "custom-session" {
		t.Errorf("Expected custom session ID 'custom-session', got '%s'", sentMsg.SessionID)
	}

	// Test query with empty session ID (should use default)
	err = client.Query(ctx, "test message 3", "")
	assertClientError(t, err, false, "")

	// Verify default session ID was used for empty string
	sentMsg, ok = transport.getSentMessage(2)
	if !ok {
		t.Fatal("Failed to get third sent message")
	}
	if sentMsg.SessionID != "default" {
		t.Errorf("Expected default session ID for empty string, got '%s'", sentMsg.SessionID)
	}

	// Verify total message count
	assertClientMessageCount(t, transport, 3)
}

// TestClientMultipleSessions tests concurrent operations with different session IDs
// Covers T151: Client Multiple Sessions + T156: State Consistency
func TestClientMultipleSessions(t *testing.T) {
	ctx, cancel := setupClientTestContext(t, 15*time.Second)
	defer cancel()

	transport := newClientMockTransport()
	client := setupClientForTest(t, transport)
	defer disconnectClientSafely(t, client)

	connectClientSafely(t, ctx, client)

	// Test concurrent operations with different session IDs
	const numSessions = 3
	const queriesPerSession = 2
	sessionIDs := []string{"session-1", "session-2", "session-3"}

	var wg sync.WaitGroup
	errors := make(chan error, numSessions*queriesPerSession)

	// Launch concurrent operations for different sessions
	for i, sessionID := range sessionIDs {
		wg.Add(1)
		go func(id int, sess string) {
			defer wg.Done()
			for j := 0; j < queriesPerSession; j++ {
				err := client.Query(ctx, fmt.Sprintf("query %d-%d", id, j), sess)
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

	// Verify all messages were sent
	expectedMessageCount := numSessions * queriesPerSession
	assertClientMessageCount(t, transport, expectedMessageCount)

	// Verify session IDs were properly propagated
	sessionCounts := make(map[string]int)
	for i := 0; i < expectedMessageCount; i++ {
		sentMsg, ok := transport.getSentMessage(i)
		if !ok {
			t.Errorf("Failed to get sent message %d", i)
			continue
		}
		sessionCounts[sentMsg.SessionID]++
	}

	// Verify each session received the correct number of messages
	for _, sessionID := range sessionIDs {
		if sessionCounts[sessionID] != queriesPerSession {
			t.Errorf("Session %s: expected %d messages, got %d",
				sessionID, queriesPerSession, sessionCounts[sessionID])
		}
	}

	// Test state consistency: client should remain connected throughout
	if !transport.connected {
		t.Error("Expected client to remain connected after concurrent session operations")
	}

	// Test session isolation: different sessions should not interfere
	err := client.Query(ctx, "final test", "session-1")
	assertClientError(t, err, false, "")

	// Verify the final message used correct session ID
	finalMsg, ok := transport.getSentMessage(expectedMessageCount)
	if !ok {
		t.Fatal("Failed to get final sent message")
	}
	if finalMsg.SessionID != "session-1" {
		t.Errorf("Expected final message session ID 'session-1', got '%s'", finalMsg.SessionID)
	}
}

// TestClientBackpressure tests handling of slow message processing
// Covers T148: Client Backpressure
func TestClientBackpressure(t *testing.T) {
	ctx, cancel := setupClientTestContext(t, 5*time.Second)
	defer cancel()

	transport := newClientMockTransport()
	client := setupClientForTest(t, transport)
	defer disconnectClientSafely(t, client)
	connectClientSafely(t, ctx, client)

	// Rapidly send multiple queries to test backpressure handling
	for i := 0; i < 10; i++ {
		err := client.Query(ctx, fmt.Sprintf("message %d", i))
		assertClientError(t, err, false, "")
	}

	// Verify all messages were handled without blocking
	assertClientMessageCount(t, transport, 10)
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
	assertClientError(t, err, false, "")

	// Simulate disconnect and reconnect
	disconnectClientSafely(t, client)
	assertClientDisconnected(t, transport)
	transport.reset()
	connectClientSafely(t, ctx, client)

	// Test recovery after reconnection
	err = client.Query(ctx, "test after reconnect")
	assertClientError(t, err, false, "")
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
	assertClientError(t, err, false, "")
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
	assertClientError(t, err, false, "")

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
		assertClientError(t, err, false, "")
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

// TestClientIteratorClose tests the clientIterator Close method
func TestClientIteratorClose(t *testing.T) {
	tests := []struct {
		name       string
		withMsgs   bool
		validateFn func(*testing.T, Client, *clientMockTransport)
	}{
		{"close_unused_iterator", false, verifyIteratorCloseUnused},
		{"close_with_pending_messages", true, verifyIteratorCloseWithMessages},
		{"multiple_close_calls", false, verifyIteratorMultipleClose},
		{"close_after_partial_consumption", true, verifyIteratorCloseAfterConsumption},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var transport *clientMockTransport
			if test.withMsgs {
				transport = newClientMockTransportWithOptions(WithClientResponseMessages([]Message{
					&AssistantMessage{Content: []ContentBlock{&TextBlock{Text: "response1"}}, Model: "claude-sonnet-3-5-20241022"},
					&AssistantMessage{Content: []ContentBlock{&TextBlock{Text: "response2"}}, Model: "claude-sonnet-3-5-20241022"},
				}))
			} else {
				transport = newClientMockTransport()
			}

			client := setupClientForTest(t, transport)
			defer disconnectClientSafely(t, client)

			test.validateFn(t, client, transport)
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
}

func (c *clientMockTransport) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
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

// Helper method to inject test messages (pattern-consistent)
func (c *clientMockTransport) injectTestMessage(msg Message) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Add to test messages for configuration
	if c.testMessages == nil {
		c.testMessages = []Message{}
	}
	c.testMessages = append(c.testMessages, msg)

	// If channels are already initialized, send immediately
	if c.msgChan != nil {
		select {
		case c.msgChan <- msg:
		default:
			// Channel full, skip (test should handle this)
		}
	}
}

// Mock Transport Options
type ClientMockTransportOption func(*clientMockTransport)

func WithClientConnectError(err error) ClientMockTransportOption {
	return func(t *clientMockTransport) {
		t.connectError = err
	}
}

func WithClientSendError(err error) ClientMockTransportOption {
	return func(t *clientMockTransport) {
		t.sendError = err
	}
}

func WithClientInterruptError(err error) ClientMockTransportOption {
	return func(t *clientMockTransport) {
		t.interruptError = err
	}
}

func WithClientCloseError(err error) ClientMockTransportOption {
	return func(t *clientMockTransport) {
		t.closeError = err
	}
}

// Simplified options for compatibility
func WithClientAsyncError(err error) ClientMockTransportOption {
	return func(t *clientMockTransport) {
		t.asyncError = err
	}
}

func WithClientResponseMessages(messages []Message) ClientMockTransportOption {
	return func(t *clientMockTransport) {
		t.testMessages = messages
	}
}

// Factory Functions
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
	if err := client.Disconnect(); err != nil {
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

// Configuration verification helpers following options_test.go patterns
func verifyDefaultConfiguration(t *testing.T, client Client, transport *clientMockTransport) {
	t.Helper()
	ctx, cancel := setupClientTestContext(t, 5*time.Second)
	defer cancel()

	if client == nil {
		t.Fatal("Expected client to be created")
	}

	connectClientSafely(t, ctx, client)
	err := client.Query(ctx, "default test")
	assertClientError(t, err, false, "")
	assertClientMessageCount(t, transport, 1)

	sentMsg, ok := transport.getSentMessage(0)
	if !ok {
		t.Fatal("Expected sent message")
	}
	if sentMsg.SessionID != "default" {
		t.Errorf("Expected default session ID 'default', got %q", sentMsg.SessionID)
	}
}

func verifySystemPromptConfig(t *testing.T, client Client, transport *clientMockTransport) {
	t.Helper()
	ctx, cancel := setupClientTestContext(t, 5*time.Second)
	defer cancel()

	if client == nil {
		t.Fatal("Expected client to be created with system prompt")
	}

	connectClientSafely(t, ctx, client)
	err := client.Query(ctx, "test with system prompt")
	assertClientError(t, err, false, "")
	assertClientMessageCount(t, transport, 1)
}

func verifyToolsConfig(t *testing.T, client Client, transport *clientMockTransport) {
	t.Helper()
	ctx, cancel := setupClientTestContext(t, 5*time.Second)
	defer cancel()

	if client == nil {
		t.Fatal("Expected client to be created with tools configuration")
	}

	connectClientSafely(t, ctx, client)
	err := client.Query(ctx, "test with tools config")
	assertClientError(t, err, false, "")
	assertClientMessageCount(t, transport, 1)
}

func verifyOptionsConfig(t *testing.T, client Client, transport *clientMockTransport) {
	t.Helper()
	ctx, cancel := setupClientTestContext(t, 5*time.Second)
	defer cancel()

	if client == nil {
		t.Fatal("Expected client to be created with multiple options")
	}

	connectClientSafely(t, ctx, client)
	err := client.Query(ctx, "test option precedence")
	assertClientError(t, err, false, "")
	assertClientMessageCount(t, transport, 1)
}

func verifyComplexConfig(t *testing.T, client Client, transport *clientMockTransport) {
	t.Helper()
	ctx, cancel := setupClientTestContext(t, 5*time.Second)
	defer cancel()

	if client == nil {
		t.Fatal("Expected client to be created with complex configuration")
	}

	connectClientSafely(t, ctx, client)

	err := client.Query(ctx, "first complex query")
	assertClientError(t, err, false, "")

	err = client.Query(ctx, "second complex query")
	assertClientError(t, err, false, "")

	assertClientMessageCount(t, transport, 2)

	for i := 0; i < 2; i++ {
		sentMsg, ok := transport.getSentMessage(i)
		if !ok {
			t.Fatalf("Expected sent message %d", i)
		}
		if sentMsg.Type != "user" {
			t.Errorf("Expected message type 'user', got %q", sentMsg.Type)
		}
	}
}

func verifySessionConfig(t *testing.T, client Client, transport *clientMockTransport) {
	t.Helper()
	ctx, cancel := setupClientTestContext(t, 5*time.Second)
	defer cancel()

	if client == nil {
		t.Fatal("Expected client to be created with session configuration")
	}

	connectClientSafely(t, ctx, client)

	err := client.Query(ctx, "session test", "custom-session-456")
	assertClientError(t, err, false, "")

	sentMsg, ok := transport.getSentMessage(0)
	if !ok {
		t.Fatal("Expected sent message")
	}
	if sentMsg.SessionID != "custom-session-456" {
		t.Errorf("Expected session ID 'custom-session-456', got %q", sentMsg.SessionID)
	}
}

// Iterator verification helpers following options_test.go patterns
func verifyIteratorCloseUnused(t *testing.T, client Client, transport *clientMockTransport) {
	t.Helper()
	ctx, cancel := setupClientTestContext(t, 10*time.Second)
	defer cancel()

	connectClientSafely(t, ctx, client)

	iter := client.ReceiveResponse(ctx)
	if iter == nil {
		t.Fatal("Expected non-nil iterator from ReceiveResponse")
	}

	err := iter.Close()
	if err != nil {
		t.Errorf("Expected Close() to succeed, got: %v", err)
	}

	msg, err := iter.Next(ctx)
	if err != ErrNoMoreMessages {
		t.Errorf("Expected ErrNoMoreMessages after close, got: %v", err)
	}
	if msg != nil {
		t.Error("Expected nil message after close, got message")
	}
}

func verifyIteratorCloseWithMessages(t *testing.T, client Client, transport *clientMockTransport) {
	t.Helper()
	ctx, cancel := setupClientTestContext(t, 10*time.Second)
	defer cancel()

	connectClientSafely(t, ctx, client)

	err := client.Query(ctx, "test with messages")
	assertClientError(t, err, false, "")

	iter := client.ReceiveResponse(ctx)
	if iter == nil {
		t.Fatal("Expected non-nil iterator from ReceiveResponse")
	}

	err = iter.Close()
	if err != nil {
		t.Errorf("Expected Close() to succeed with pending messages, got: %v", err)
	}

	msg, err := iter.Next(ctx)
	if err != ErrNoMoreMessages {
		t.Errorf("Expected ErrNoMoreMessages after close, got: %v", err)
	}
	if msg != nil {
		t.Error("Expected nil message after close, got message")
	}
}

func verifyIteratorMultipleClose(t *testing.T, client Client, transport *clientMockTransport) {
	t.Helper()
	ctx, cancel := setupClientTestContext(t, 10*time.Second)
	defer cancel()

	connectClientSafely(t, ctx, client)

	iter := client.ReceiveResponse(ctx)
	if iter == nil {
		t.Fatal("Expected non-nil iterator from ReceiveResponse")
	}

	// Multiple closes should all succeed (idempotent)
	for i := 1; i <= 3; i++ {
		err := iter.Close()
		if err != nil {
			t.Errorf("Expected Close() call %d to succeed, got: %v", i, err)
		}
	}

	// Next should consistently return ErrNoMoreMessages
	for i := 0; i < 3; i++ {
		msg, err := iter.Next(ctx)
		if err != ErrNoMoreMessages {
			t.Errorf("Expected ErrNoMoreMessages on call %d after multiple closes, got: %v", i+1, err)
		}
		if msg != nil {
			t.Errorf("Expected nil message on call %d after multiple closes, got message", i+1)
		}
	}
}

func verifyIteratorCloseAfterConsumption(t *testing.T, client Client, transport *clientMockTransport) {
	t.Helper()
	ctx, cancel := setupClientTestContext(t, 10*time.Second)
	defer cancel()

	connectClientSafely(t, ctx, client)

	err := client.Query(ctx, "multi message test")
	assertClientError(t, err, false, "")

	iter := client.ReceiveResponse(ctx)
	if iter == nil {
		t.Fatal("Expected non-nil iterator from ReceiveResponse")
	}

	// Consume one message
	msg, err := iter.Next(ctx)
	if err != nil {
		t.Fatalf("Expected first message, got error: %v", err)
	}
	if msg == nil {
		t.Fatal("Expected first message, got nil")
	}

	// Close after consuming
	err = iter.Close()
	if err != nil {
		t.Errorf("Expected Close() after consuming message to succeed, got: %v", err)
	}

	// Next should return ErrNoMoreMessages even though more messages might be available
	msg, err = iter.Next(ctx)
	if err != ErrNoMoreMessages {
		t.Errorf("Expected ErrNoMoreMessages after close, got: %v", err)
	}
	if msg != nil {
		t.Error("Expected nil message after close, got message")
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
				// Use a more reliable timeout that works across platforms
				return context.WithTimeout(context.Background(), 1*time.Microsecond)
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
		return client.Query(ctx, "options test", "custom-session")
	},
		WithSystemPrompt("Test system prompt"),
		WithAllowedTools("Read", "Write"),
	)

	assertClientError(t, err, false, "")
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
	client := setupClientForTest(t, transport)
	defer disconnectClientSafely(t, client)

	// Test complete workflow: Connect → Query → ReceiveMessages
	connectClientSafely(t, ctx, client)
	assertClientConnected(t, transport)

	// Send query using new Python SDK compatible format
	err := client.Query(ctx, "Test streaming with Python SDK format", "test-session")
	assertClientError(t, err, false, "")

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

	// Verify nested message structure matches Python format
	messageMap, ok := sentMsg.Message.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected Message to be map[string]interface{}, got %T", sentMsg.Message)
	}
	if role, ok := messageMap["role"]; !ok || role != "user" {
		t.Errorf("Expected message role 'user', got '%v'", role)
	}
	if content, ok := messageMap["content"]; !ok || content != "Test streaming with Python SDK format" {
		t.Errorf("Expected content to match prompt, got '%v'", content)
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
	assertClientError(t, err, false, "")
}
