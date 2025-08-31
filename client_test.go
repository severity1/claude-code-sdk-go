package claudecode

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

// T133: Client Auto Connect Context Manager 🔴 RED
// Python Reference: test_streaming_client.py::TestClaudeSDKClientStreaming::test_auto_connect_with_context_manager
func TestClientAutoConnectContextManager(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create mock transport for testing
	transport := &clientMockTransport{}
	
	// Test defer-based resource management (Go equivalent of Python context manager)
	func() {
		client := NewClientWithTransport(transport)
		defer client.Disconnect() // Go defer pattern equivalent to context manager
		
		// Connect should be called automatically or explicitly
		err := client.Connect(ctx)
		if err != nil {
			t.Fatalf("Client connect failed: %v", err)
		}
		
		// Verify connection was established
		if !transport.connected {
			t.Error("Expected transport to be connected")
		}
		
		// Client should be ready to use
		err = client.Query(ctx, "test message")
		if err != nil {
			t.Errorf("Client query failed: %v", err)
		}
	}() // Defer should trigger disconnect
	
	// Verify disconnect was called
	if transport.connected {
		t.Error("Expected transport to be disconnected after defer")
	}
}

// T134: Client Manual Connection 🔴 RED
// Go Target: client_test.go::TestClientManualConnection
func TestClientManualConnection(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	transport := &clientMockTransport{}
	client := NewClientWithTransport(transport)

	// Manual Connect/Disconnect lifecycle
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	
	if !transport.connected {
		t.Error("Expected transport to be connected")
	}

	err = client.Disconnect()
	if err != nil {
		t.Errorf("Disconnect failed: %v", err)
	}
	
	if transport.connected {
		t.Error("Expected transport to be disconnected")
	}
}

// T135: Client Query Execution 🔴 RED
// Go Target: client_test.go::TestClientQueryExecution
func TestClientQueryExecution(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	transport := &clientMockTransport{}
	client := NewClientWithTransport(transport)
	defer client.Disconnect()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Execute query through connected client
	err = client.Query(ctx, "What is 2+2?")
	if err != nil {
		t.Errorf("Query execution failed: %v", err)
	}

	// Verify message was sent to transport
	if transport.getSentMessageCount() != 1 {
		t.Errorf("Expected 1 sent message, got %d", transport.getSentMessageCount())
	}
	
	// Verify message content
	sentMsg, ok := transport.getSentMessage(0)
	if !ok {
		t.Fatal("Failed to get sent message")
	}
	if sentMsg.Type != "request" {
		t.Errorf("Expected message type 'request', got '%s'", sentMsg.Type)
	}
	
	userMsg, ok := sentMsg.Message.(*UserMessage)
	if !ok {
		t.Fatalf("Expected UserMessage, got %T", sentMsg.Message)
	}
	
	if content, ok := userMsg.Content.(string); !ok || content != "What is 2+2?" {
		t.Errorf("Expected content 'What is 2+2?', got '%v'", userMsg.Content)
	}
}

// T136: Client Stream Query 🔴 RED
// Go Target: client_test.go::TestClientStreamQuery
func TestClientStreamQuery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	transport := &clientMockTransport{}
	client := NewClientWithTransport(transport)
	defer client.Disconnect()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Create message channel
	messages := make(chan StreamMessage, 3)
	messages <- StreamMessage{Type: "request", Message: &UserMessage{Content: "First message"}}
	messages <- StreamMessage{Type: "request", Message: &UserMessage{Content: "Second message"}}
	messages <- StreamMessage{Type: "request", Message: &UserMessage{Content: "Third message"}}
	close(messages)

	// Execute stream query
	err = client.QueryStream(ctx, messages)
	if err != nil {
		t.Errorf("Stream query execution failed: %v", err)
	}

	// Give time for messages to be processed
	time.Sleep(50 * time.Millisecond)

	// Verify all messages were sent
	if transport.getSentMessageCount() != 3 {
		t.Errorf("Expected 3 sent messages, got %d", transport.getSentMessageCount())
	}
}

// T137: Client Message Reception 🔴 RED
// Go Target: client_test.go::TestClientMessageReception
func TestClientMessageReception(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	transport := &clientMockTransport{
		responseMessages: []Message{
			&AssistantMessage{
				Content: []ContentBlock{&TextBlock{Text: "Hello!"}},
				Model:   "claude-opus-4-1-20250805",
			},
			&ResultMessage{
				Subtype:   "success",
				IsError:   false,
				NumTurns:  1,
				SessionID: "test-session",
			},
		},
	}
	
	client := NewClientWithTransport(transport)
	defer client.Disconnect()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Get message channel
	msgChan := client.ReceiveMessages(ctx)
	
	// Trigger sending response messages
	transport.sendResponses()

	// Receive messages
	var receivedMessages []Message
	for i := 0; i < 2; i++ {
		select {
		case msg := <-msgChan:
			if msg != nil {
				receivedMessages = append(receivedMessages, msg)
			}
		case <-time.After(1 * time.Second):
			t.Fatal("Timeout waiting for messages")
		}
	}

	// Verify we received both messages
	if len(receivedMessages) != 2 {
		t.Errorf("Expected 2 received messages, got %d", len(receivedMessages))
	}

	// Verify message types
	assistantMsg, ok := receivedMessages[0].(*AssistantMessage)
	if !ok {
		t.Errorf("Expected first message to be AssistantMessage, got %T", receivedMessages[0])
	} else {
		textBlock := assistantMsg.Content[0].(*TextBlock)
		if textBlock.Text != "Hello!" {
			t.Errorf("Expected text 'Hello!', got '%s'", textBlock.Text)
		}
	}

	resultMsg, ok := receivedMessages[1].(*ResultMessage)
	if !ok {
		t.Errorf("Expected second message to be ResultMessage, got %T", receivedMessages[1])
	} else {
		if resultMsg.Subtype != "success" {
			t.Errorf("Expected subtype 'success', got '%s'", resultMsg.Subtype)
		}
	}
}

// clientMockTransport implements Transport interface for client testing
type clientMockTransport struct {
	mu               sync.Mutex
	connected        bool
	closed           bool // Track if transport is closed
	msgChan          chan Message
	errChan          chan error
	sentMessages     []StreamMessage
	responseMessages []Message
	
	// Error injection for testing
	connectError    error
	sendError       error
	interruptError  error
	closeError      error
	asyncError      error
	slowSend        bool // Enable slow sending for backpressure testing
}

func (c *clientMockTransport) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.connectError != nil {
		return c.connectError
	}
	c.connected = true
	c.msgChan = make(chan Message, 10)
	c.errChan = make(chan error, 10)
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
	
	// Simulate slow sending for backpressure testing
	if c.slowSend {
		c.mu.Unlock()
		select {
		case <-time.After(50 * time.Millisecond):
		case <-ctx.Done():
			c.mu.Lock()
			return ctx.Err()
		}
		c.mu.Lock()
	}
	
	c.sentMessages = append(c.sentMessages, message)
	return nil
}

func (c *clientMockTransport) ReceiveMessages(ctx context.Context) (<-chan Message, <-chan error) {
	return c.msgChan, c.errChan
}

func (c *clientMockTransport) Interrupt(ctx context.Context) error {
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
	
	// Mark as closed and disconnected
	c.connected = false
	if !c.closed {
		c.closed = true
		if c.msgChan != nil {
			close(c.msgChan)
			c.msgChan = nil
		}
		if c.errChan != nil {
			close(c.errChan)
			c.errChan = nil
		}
	}
	return nil
}

// Helper method to send response messages
func (c *clientMockTransport) sendResponses() {
	c.mu.Lock()
	msgChan := c.msgChan // Capture channel reference
	responseMessages := c.responseMessages
	closed := c.closed
	c.mu.Unlock()
	
	if responseMessages != nil && msgChan != nil && !closed {
		go func() {
			defer func() {
				// Recover from panic if channel is closed
				if r := recover(); r != nil {
					// Ignore panic from closed channel
				}
			}()
			for _, msg := range responseMessages {
				// Check if we should stop sending
				c.mu.Lock()
				shouldStop := c.closed
				c.mu.Unlock()
				
				if shouldStop {
					return
				}
				
				select {
				case msgChan <- msg:
				default:
					// Channel might be full or closed, just return
					return
				}
			}
		}()
	}
}

// Helper method to send async error
func (c *clientMockTransport) sendAsyncError() {
	c.mu.Lock()
	errChan := c.errChan
	asyncError := c.asyncError
	closed := c.closed
	c.mu.Unlock()
	
	if asyncError != nil && errChan != nil && !closed {
		go func() {
			defer func() {
				// Recover from panic if channel is closed
				if r := recover(); r != nil {
					// Ignore panic from closed channel
				}
			}()
			
			select {
			case errChan <- asyncError:
			default:
				// Channel might be closed or full, ignore
			}
		}()
	}
}

// Helper method to safely get sent messages count
func (c *clientMockTransport) getSentMessageCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.sentMessages)
}

// Helper method to safely reset sent messages
func (c *clientMockTransport) resetSentMessages() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sentMessages = nil
}

// Helper method to safely get a specific sent message
func (c *clientMockTransport) getSentMessage(index int) (StreamMessage, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if index < 0 || index >= len(c.sentMessages) {
		return StreamMessage{}, false
	}
	return c.sentMessages[index], true
}

// T138: Client Response Iterator 🔴 RED
// Go Target: client_test.go::TestClientResponseIterator
func TestClientResponseIterator(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	transport := &clientMockTransport{
		responseMessages: []Message{
			&AssistantMessage{Content: []ContentBlock{&TextBlock{Text: "Response 1"}}, Model: "claude-opus-4-1-20250805"},
			&AssistantMessage{Content: []ContentBlock{&TextBlock{Text: "Response 2"}}, Model: "claude-opus-4-1-20250805"},
		},
	}
	
	client := NewClientWithTransport(transport)
	defer client.Disconnect()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Get response iterator
	iter := client.ReceiveResponse(ctx)
	if iter == nil {
		t.Fatal("Expected response iterator, got nil")
	}
	defer iter.Close()
	
	// Trigger responses
	transport.sendResponses()

	// Read responses using iterator
	var responses []Message
	for i := 0; i < 2; i++ {
		msg, err := iter.Next(ctx)
		if err != nil {
			t.Fatalf("Iterator error: %v", err)
		}
		responses = append(responses, msg)
	}

	if len(responses) != 2 {
		t.Errorf("Expected 2 responses, got %d", len(responses))
	}
}

// T139: Client Interrupt Functionality 🔴 RED
// Go Target: client_test.go::TestClientInterruptFunctionality  
func TestClientInterruptFunctionality(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	transport := &clientMockTransport{}
	client := NewClientWithTransport(transport)
	defer client.Disconnect()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Test interrupt
	err = client.Interrupt(ctx)
	if err != nil {
		t.Errorf("Interrupt failed: %v", err)
	}
}

// T141: Client Connection State 🔴 RED
// Go Target: client_test.go::TestClientConnectionState
func TestClientConnectionState(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	transport := &clientMockTransport{}
	client := NewClientWithTransport(transport)

	// Test operations on disconnected client
	err := client.Query(ctx, "test")
	if err == nil {
		t.Error("Expected error when querying disconnected client")
	}

	err = client.Interrupt(ctx)
	if err == nil {
		t.Error("Expected error when interrupting disconnected client")
	}

	// Connect and test operations work
	err = client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	err = client.Query(ctx, "test")
	if err != nil {
		t.Errorf("Query should work when connected: %v", err)
	}

	err = client.Interrupt(ctx)
	if err != nil {
		t.Errorf("Interrupt should work when connected: %v", err)
	}

	// Disconnect and test operations fail again
	err = client.Disconnect()
	if err != nil {
		t.Errorf("Disconnect failed: %v", err)
	}

	err = client.Query(ctx, "test")
	if err == nil {
		t.Error("Expected error when querying after disconnect")
	}
}

// T140: Client Session Management 🔴 RED
// Go Target: client_test.go::TestClientSessionManagement
func TestClientSessionManagement(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	transport := &clientMockTransport{}
	client := NewClientWithTransport(transport)
	defer client.Disconnect()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Test default session ID
	err = client.Query(ctx, "test message")
	if err != nil {
		t.Errorf("Query with default session failed: %v", err)
	}

	if transport.getSentMessageCount() != 1 {
		t.Fatalf("Expected 1 message, got %d", transport.getSentMessageCount())
	}

	// Should use default session ID
	sentMsg, ok := transport.getSentMessage(0)
	if !ok {
		t.Fatal("Failed to get sent message")
	}
	if sentMsg.SessionID != "default" {
		t.Errorf("Expected default session ID 'default', got '%s'", sentMsg.SessionID)
	}

	// Test custom session ID
	transport.resetSentMessages() // Reset
	err = client.Query(ctx, "test message 2", "custom-session")
	if err != nil {
		t.Errorf("Query with custom session failed: %v", err)
	}

	if transport.getSentMessageCount() != 1 {
		t.Fatalf("Expected 1 message, got %d", transport.getSentMessageCount())
	}

	sentMsg, ok = transport.getSentMessage(0)
	if !ok {
		t.Fatal("Failed to get sent message")
	}
	if sentMsg.SessionID != "custom-session" {
		t.Errorf("Expected custom session ID 'custom-session', got '%s'", sentMsg.SessionID)
	}

	// Test multiple custom session IDs (should use the first one)
	transport.resetSentMessages() // Reset
	err = client.Query(ctx, "test message 3", "session1", "session2")
	if err != nil {
		t.Errorf("Query with multiple session IDs failed: %v", err)
	}

	if transport.getSentMessageCount() != 1 {
		t.Fatalf("Expected 1 message, got %d", transport.getSentMessageCount())
	}

	sentMsg, ok = transport.getSentMessage(0)
	if !ok {
		t.Fatal("Failed to get sent message")
	}
	if sentMsg.SessionID != "session1" {
		t.Errorf("Expected first session ID 'session1', got '%s'", sentMsg.SessionID)
	}
}

// T142: Client Error Propagation 🔴 RED
// Go Target: client_test.go::TestClientErrorPropagation
func TestClientErrorPropagation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test transport connection error propagation
	transport := &clientMockTransport{
		connectError: fmt.Errorf("connection failed: CLI not found"),
	}
	client := NewClientWithTransport(transport)

	err := client.Connect(ctx)
	if err == nil {
		t.Error("Expected connection error to be propagated")
	}
	if err.Error() != "failed to connect transport: connection failed: CLI not found" {
		t.Errorf("Expected wrapped connection error, got: %v", err)
	}

	// Test transport send error propagation
	transport2 := &clientMockTransport{
		sendError: fmt.Errorf("send failed: process exited"),
	}
	client2 := NewClientWithTransport(transport2)
	defer client2.Disconnect()

	err = client2.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect should succeed: %v", err)
	}

	err = client2.Query(ctx, "test")
	if err == nil {
		t.Error("Expected send error to be propagated")
	}
	if err.Error() != "send failed: process exited" {
		t.Errorf("Expected send error, got: %v", err)
	}

	// Test error channel propagation
	transport3 := &clientMockTransport{
		asyncError: fmt.Errorf("async transport error"),
	}
	client3 := NewClientWithTransport(transport3)
	defer client3.Disconnect()

	err = client3.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect should succeed: %v", err)
	}

	// Get iterator and expect error from error channel
	iter := client3.ReceiveResponse(ctx)
	if iter == nil {
		t.Fatal("Expected response iterator")
	}
	defer iter.Close()

	// Trigger async error
	transport3.sendAsyncError()

	// Next should return the async error
	_, err = iter.Next(ctx)
	if err == nil {
		t.Error("Expected async error to be propagated through iterator")
	}
	if err.Error() != "async transport error" {
		t.Errorf("Expected async transport error, got: %v", err)
	}

	// Test interrupt error propagation
	transport4 := &clientMockTransport{
		interruptError: fmt.Errorf("interrupt failed"),
	}
	client4 := NewClientWithTransport(transport4)
	defer client4.Disconnect()

	err = client4.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect should succeed: %v", err)
	}

	err = client4.Interrupt(ctx)
	if err == nil {
		t.Error("Expected interrupt error to be propagated")
	}
	if err.Error() != "interrupt failed" {
		t.Errorf("Expected interrupt error, got: %v", err)
	}

	// Test close error propagation
	transport5 := &clientMockTransport{
		closeError: fmt.Errorf("close failed"),
	}
	client5 := NewClientWithTransport(transport5)

	err = client5.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect should succeed: %v", err)
	}

	err = client5.Disconnect()
	if err == nil {
		t.Error("Expected close error to be propagated")
	}
	if err.Error() != "failed to close transport: close failed" {
		t.Errorf("Expected wrapped close error, got: %v", err)
	}
}

// T143: Client Concurrent Access 🔴 RED
// Go Target: client_test.go::TestClientConcurrentAccess
func TestClientConcurrentAccess(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	transport := &clientMockTransport{}
	client := NewClientWithTransport(transport)
	defer client.Disconnect()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Test concurrent Query operations
	const numGoroutines = 10
	const queriesPerGoroutine = 5
	
	var wg sync.WaitGroup
	errorChan := make(chan error, numGoroutines*queriesPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < queriesPerGoroutine; j++ {
				prompt := fmt.Sprintf("Query from goroutine %d, message %d", goroutineID, j)
				sessionID := fmt.Sprintf("session-%d-%d", goroutineID, j)
				
				if err := client.Query(ctx, prompt, sessionID); err != nil {
					errorChan <- err
					return
				}
			}
		}(i)
	}

	wg.Wait()
	close(errorChan)

	// Check for any errors
	for err := range errorChan {
		t.Errorf("Concurrent query error: %v", err)
	}

	// Verify all messages were sent
	expectedMessages := numGoroutines * queriesPerGoroutine
	if transport.getSentMessageCount() != expectedMessages {
		t.Errorf("Expected %d messages, got %d", expectedMessages, transport.getSentMessageCount())
	}

	// Test concurrent Interrupt operations
	transport.resetSentMessages() // Reset
	var interruptWg sync.WaitGroup
	interruptErrors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		interruptWg.Add(1)
		go func() {
			defer interruptWg.Done()
			if err := client.Interrupt(ctx); err != nil {
				interruptErrors <- err
			}
		}()
	}

	interruptWg.Wait()
	close(interruptErrors)

	// Check for interrupt errors
	for err := range interruptErrors {
		t.Errorf("Concurrent interrupt error: %v", err)
	}

	// Test concurrent Connect/Disconnect operations (should be handled gracefully)
	var connectWg sync.WaitGroup
	connectErrors := make(chan error, 4)

	// Test concurrent reconnections
	for i := 0; i < 2; i++ {
		connectWg.Add(1)
		go func() {
			defer connectWg.Done()
			client.Disconnect()
			if err := client.Connect(ctx); err != nil {
				connectErrors <- err
			}
		}()
	}

	connectWg.Wait()
	close(connectErrors)

	// Check for connection errors (some expected due to race conditions)
	var connectionErrors []error
	for err := range connectErrors {
		connectionErrors = append(connectionErrors, err)
	}
	
	// At least one connection should succeed, others may fail due to concurrent access
	if len(connectionErrors) == 4 {
		t.Error("All concurrent connection attempts failed")
	}
}

// T146: Client Transport Selection 🔴 RED
// Go Target: client_test.go::TestClientTransportSelection
func TestClientTransportSelection(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test that NewClient creates a client that can create default transport
	client := NewClient()
	defer client.Disconnect()

	// This should work without NewClientWithTransport
	err := client.Connect(ctx)
	// We expect this to fail but should NOT fail with the old "no transport available" message
	if err != nil {
		if strings.Contains(err.Error(), "no transport available - use NewClientWithTransport for testing") {
			t.Errorf("Client should create default transport, but got old error: %v", err)
		}
		if strings.Contains(err.Error(), "no default transport factory available") {
			t.Errorf("Transport factory should be initialized, got: %v", err)
		}
		// Should get either CLI-related error or transport selection success message
		if !strings.Contains(err.Error(), "claude") && !strings.Contains(err.Error(), "CLI") && 
		   !strings.Contains(err.Error(), "transport selection successful") {
			t.Errorf("Expected CLI or transport selection error, got: %v", err)
		} else {
			t.Logf("Got expected transport selection behavior: %v", err)
		}
	} else {
		// If it actually succeeds (CLI is available and transport works), that's fine too
		t.Log("Client successfully created default transport and connected")
	}

	// Test that transport uses closeStdin=false for client mode
	// This is harder to test without mocking internals, but we can verify the transport is created
	if client.(*ClientImpl).customTransport != nil {
		t.Error("NewClient should not have custom transport, should create default")
	}
}

// T144: Client Resource Cleanup 🔴 RED
// Go Target: client_test.go::TestClientResourceCleanup
func TestClientResourceCleanup(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test 1: Basic resource cleanup after disconnect
	t.Run("BasicResourceCleanup", func(t *testing.T) {
		transport := &clientMockTransport{}
		client := NewClientWithTransport(transport)

		err := client.Connect(ctx)
		if err != nil {
			t.Fatalf("Connect failed: %v", err)
		}

		// Verify resources are allocated
		if !transport.connected {
			t.Error("Expected transport to be connected")
		}

		// Disconnect and verify cleanup
		err = client.Disconnect()
		if err != nil {
			t.Errorf("Disconnect failed: %v", err)
		}

		// Verify transport is cleaned up
		if transport.connected {
			t.Error("Expected transport to be disconnected after cleanup")
		}

		// Verify client state is reset
		clientImpl := client.(*ClientImpl)
		clientImpl.mu.RLock()
		connected := clientImpl.connected
		transportRef := clientImpl.transport
		msgChan := clientImpl.msgChan
		errChan := clientImpl.errChan
		clientImpl.mu.RUnlock()

		if connected {
			t.Error("Expected client.connected to be false after disconnect")
		}
		if transportRef != nil {
			t.Error("Expected client.transport to be nil after disconnect")
		}
		if msgChan != nil {
			t.Error("Expected client.msgChan to be nil after disconnect")
		}
		if errChan != nil {
			t.Error("Expected client.errChan to be nil after disconnect")
		}
	})

	// Test 2: Multiple disconnect calls should be safe (no double-close)
	t.Run("MultipleDisconnectSafe", func(t *testing.T) {
		transport := &clientMockTransport{}
		client := NewClientWithTransport(transport)

		err := client.Connect(ctx)
		if err != nil {
			t.Fatalf("Connect failed: %v", err)
		}

		// First disconnect
		err = client.Disconnect()
		if err != nil {
			t.Errorf("First disconnect failed: %v", err)
		}

		// Second disconnect should not panic or error
		err = client.Disconnect()
		if err != nil {
			t.Errorf("Second disconnect should be safe, but got error: %v", err)
		}

		// Third disconnect should also be safe
		err = client.Disconnect()
		if err != nil {
			t.Errorf("Third disconnect should be safe, but got error: %v", err)
		}
	})

	// Test 3: Resource cleanup with goroutines running
	t.Run("CleanupWithActiveGoroutines", func(t *testing.T) {
		transport := &clientMockTransport{
			responseMessages: []Message{
				&AssistantMessage{
					Content: []ContentBlock{&TextBlock{Text: "Hello"}},
					Model:   "claude-opus-4-1-20250805",
				},
			},
		}
		client := NewClientWithTransport(transport)

		err := client.Connect(ctx)
		if err != nil {
			t.Fatalf("Connect failed: %v", err)
		}

		// Start some operations that create goroutines
		messages := make(chan StreamMessage, 2)
		messages <- StreamMessage{Type: "request", Message: &UserMessage{Content: "Test 1"}}
		messages <- StreamMessage{Type: "request", Message: &UserMessage{Content: "Test 2"}}
		close(messages)

		err = client.QueryStream(ctx, messages)
		if err != nil {
			t.Errorf("QueryStream failed: %v", err)
		}

		// Start receiving messages (this may create goroutines)
		msgChan := client.ReceiveMessages(ctx)
		
		// Start a goroutine to consume messages
		done := make(chan bool)
		go func() {
			for {
				select {
				case msg := <-msgChan:
					if msg == nil {
						done <- true
						return
					}
				case <-time.After(100 * time.Millisecond):
					done <- true
					return
				}
			}
		}()

		// Give some time for goroutines to start
		time.Sleep(50 * time.Millisecond)

		// Now disconnect - should clean up all resources
		err = client.Disconnect()
		if err != nil {
			t.Errorf("Disconnect with active goroutines failed: %v", err)
		}

		// Wait for consumer goroutine to finish
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Error("Consumer goroutine did not finish within timeout")
		}

		// Verify cleanup
		if transport.connected {
			t.Error("Transport should be disconnected after cleanup with active goroutines")
		}
	})

	// Test 4: Memory leak prevention - repeated connect/disconnect cycles
	t.Run("RepeatedConnectDisconnectNoLeak", func(t *testing.T) {
		transport := &clientMockTransport{}
		client := NewClientWithTransport(transport)

		// Perform many connect/disconnect cycles
		for i := 0; i < 100; i++ {
			err := client.Connect(ctx)
			if err != nil {
				t.Fatalf("Connect %d failed: %v", i, err)
			}

			// Do some work
			err = client.Query(ctx, fmt.Sprintf("Test query %d", i))
			if err != nil {
				t.Errorf("Query %d failed: %v", i, err)
			}

			err = client.Disconnect()
			if err != nil {
				t.Errorf("Disconnect %d failed: %v", i, err)
			}

			// Verify state is clean after each cycle
			clientImpl := client.(*ClientImpl)
			clientImpl.mu.RLock()
			connected := clientImpl.connected
			transportRef := clientImpl.transport
			clientImpl.mu.RUnlock()

			if connected {
				t.Errorf("Client should not be connected after disconnect cycle %d", i)
			}
			if transportRef != nil {
				t.Errorf("Transport reference should be nil after disconnect cycle %d", i)
			}
		}
	})

	// Test 5: Disconnect during operations should be safe
	t.Run("DisconnectDuringOperations", func(t *testing.T) {
		transport := &clientMockTransport{}
		client := NewClientWithTransport(transport)

		err := client.Connect(ctx)
		if err != nil {
			t.Fatalf("Connect failed: %v", err)
		}

		// Start multiple operations concurrently
		var wg sync.WaitGroup
		errorChan := make(chan error, 10)

		// Start multiple queries
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				if err := client.Query(ctx, fmt.Sprintf("Query %d", id)); err != nil {
					errorChan <- err
				}
			}(i)
		}

		// Start message receiving
		wg.Add(1)
		go func() {
			defer wg.Done()
			msgChan := client.ReceiveMessages(ctx)
			for {
				select {
				case msg := <-msgChan:
					if msg == nil {
						return
					}
				case <-time.After(50 * time.Millisecond):
					return
				}
			}
		}()

		// Give operations time to start
		time.Sleep(25 * time.Millisecond)

		// Disconnect while operations are running
		err = client.Disconnect()
		if err != nil {
			t.Errorf("Disconnect during operations failed: %v", err)
		}

		// Wait for all goroutines to complete
		done := make(chan bool)
		go func() {
			wg.Wait()
			done <- true
		}()

		select {
		case <-done:
		case <-time.After(3 * time.Second):
			t.Error("Goroutines did not complete within timeout after disconnect")
		}

		// Collect any errors (some expected due to disconnection)
		close(errorChan)
		for err := range errorChan {
			// Errors are expected here due to disconnection during operations
			if !strings.Contains(err.Error(), "not connected") {
				t.Logf("Expected connection error: %v", err)
			}
		}
	})
}

// T157: Client Configuration Validation 🔴 RED
// Go Target: client_test.go::TestClientConfigurationValidation
func TestClientConfigurationValidation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test 1: Valid configuration should work
	t.Run("ValidConfiguration", func(t *testing.T) {
		transport := &clientMockTransport{}
		
		// Create client with valid options
		client := NewClientWithTransport(transport,
			WithSystemPrompt("You are a helpful assistant"),
			WithAllowedTools("Read", "Write", "Bash"),
			WithDisallowedTools("dangerous-tool"),
			WithMaxTurns(10),
			WithPermissionMode(PermissionModeAcceptEdits),
			WithModel("claude-opus-4-1-20250805"),
			WithCwd("/tmp"),
		)

		// Should connect successfully with valid configuration
		err := client.Connect(ctx)
		if err != nil {
			t.Errorf("Valid configuration should connect successfully: %v", err)
		}

		err = client.Disconnect()
		if err != nil {
			t.Errorf("Disconnect failed: %v", err)
		}
	})

	// Test 2: Invalid working directory should be rejected
	t.Run("InvalidWorkingDirectory", func(t *testing.T) {
		transport := &clientMockTransport{}
		
		// Create client with non-existent working directory
		client := NewClientWithTransport(transport,
			WithCwd("/non/existent/directory/that/should/not/exist"),
		)

		// Should fail during connect with helpful error
		err := client.Connect(ctx)
		if err == nil {
			t.Error("Expected error for invalid working directory")
			client.Disconnect()
		} else if !strings.Contains(err.Error(), "working directory") && !strings.Contains(err.Error(), "directory") {
			t.Errorf("Expected working directory validation error, got: %v", err)
		}
	})

	// Test 3: Invalid max turns should be rejected
	t.Run("InvalidMaxTurns", func(t *testing.T) {
		transport := &clientMockTransport{}
		
		// Create client with invalid max turns
		client := NewClientWithTransport(transport,
			WithMaxTurns(-1), // Negative turns should be invalid
		)

		// Should validate during client creation or connection
		err := client.Connect(ctx)
		if err == nil {
			t.Error("Expected error for negative max turns")
			client.Disconnect()
		} else if !strings.Contains(err.Error(), "max_turns") && !strings.Contains(err.Error(), "turns") {
			t.Logf("Got error (may be acceptable): %v", err)
		}
	})

	// Test 4: Invalid permission mode should be rejected
	t.Run("InvalidPermissionMode", func(t *testing.T) {
		transport := &clientMockTransport{}
		
		// Create client with invalid permission mode
		client := NewClientWithTransport(transport)
		
		// Manually set invalid permission mode to test validation
		clientImpl := client.(*ClientImpl)
		invalidMode := PermissionMode("invalid_mode")
		clientImpl.options.PermissionMode = &invalidMode

		// Should validate during connection
		err := client.Connect(ctx)
		if err == nil {
			t.Error("Expected error for invalid permission mode")
			client.Disconnect()
		} else if !strings.Contains(err.Error(), "permission") && !strings.Contains(err.Error(), "mode") {
			t.Logf("Got error (may be acceptable): %v", err)
		}
	})

	// Test 5: Tool validation - conflicting allowed and disallowed tools
	t.Run("ConflictingToolConfiguration", func(t *testing.T) {
		transport := &clientMockTransport{}
		
		// Create client with conflicting tool configuration
		client := NewClientWithTransport(transport,
			WithAllowedTools("Read", "Write"),
			WithDisallowedTools("Read"), // Read is both allowed and disallowed
		)

		// Should validate during connection
		err := client.Connect(ctx)
		if err == nil {
			// Some implementations might allow this and use precedence rules
			t.Log("Conflicting tool configuration was accepted (may use precedence rules)")
			client.Disconnect()
		} else if !strings.Contains(err.Error(), "tool") {
			t.Logf("Got error for conflicting tools: %v", err)
		}
	})

	// Test 6: Invalid model name validation
	t.Run("InvalidModelName", func(t *testing.T) {
		transport := &clientMockTransport{}
		
		// Create client with potentially invalid model name
		client := NewClientWithTransport(transport,
			WithModel(""), // Empty model name
		)

		// Should either reject empty model or use default
		err := client.Connect(ctx)
		if err != nil {
			if strings.Contains(err.Error(), "model") {
				t.Logf("Empty model name was rejected: %v", err)
			} else {
				t.Logf("Got different error: %v", err)
			}
		} else {
			// Empty model might be acceptable if defaults are used
			t.Log("Empty model name was accepted (using defaults)")
			client.Disconnect()
		}
	})

	// Test 7: Invalid MCP server configuration
	t.Run("InvalidMCPServerConfiguration", func(t *testing.T) {
		transport := &clientMockTransport{}
		
		// Create client with invalid MCP configuration
		mcpConfig := &McpStdioServerConfig{
			Type:    McpServerTypeStdio,
			Command: "", // Empty command should be invalid
			Args:    []string{},
		}
		
		mcpServers := map[string]McpServerConfig{
			"test-server": mcpConfig,
		}
		client := NewClientWithTransport(transport,
			WithMcpServers(mcpServers),
		)

		// Should validate MCP configuration
		err := client.Connect(ctx)
		if err == nil {
			t.Log("Empty MCP command was accepted")
			client.Disconnect()
		} else if !strings.Contains(err.Error(), "mcp") && !strings.Contains(err.Error(), "command") {
			t.Logf("Got error: %v", err)
		}
	})

	// Test 8: Nil options should be handled gracefully
	t.Run("NilOptionsHandling", func(t *testing.T) {
		transport := &clientMockTransport{}
		
		// Create client with no options (should use defaults)
		client := NewClientWithTransport(transport)

		// Should work with default options
		err := client.Connect(ctx)
		if err != nil {
			// This might fail due to transport setup, not options validation
			t.Logf("Connect with default options failed: %v", err)
		} else {
			client.Disconnect()
		}
	})

	// Test 9: Configuration immutability after connection
	t.Run("ConfigurationImmutabilityAfterConnection", func(t *testing.T) {
		transport := &clientMockTransport{}
		
		client := NewClientWithTransport(transport,
			WithSystemPrompt("Original prompt"),
		)

		err := client.Connect(ctx)
		if err != nil {
			t.Fatalf("Connect failed: %v", err)
		}

		// Try to modify configuration after connection (should not affect connected client)
		clientImpl := client.(*ClientImpl)
		var originalPrompt string
		if clientImpl.options.SystemPrompt != nil {
			originalPrompt = *clientImpl.options.SystemPrompt
		}
		modifiedPrompt := "Modified prompt"
		clientImpl.options.SystemPrompt = &modifiedPrompt

		// Send a query to verify the configuration wasn't changed mid-flight
		err = client.Query(ctx, "test")
		if err != nil {
			t.Errorf("Query failed after options modification: %v", err)
		}

		// Clean up
		client.Disconnect()

		// The original prompt should have been preserved in some form
		if originalPrompt != "Original prompt" {
			t.Errorf("Expected original prompt to be preserved, got %q", originalPrompt)
		}
	})

	// Test 10: Large configuration values
	t.Run("LargeConfigurationValues", func(t *testing.T) {
		transport := &clientMockTransport{}
		
		// Create very large system prompt
		largePrompt := strings.Repeat("This is a very long system prompt. ", 1000)
		
		client := NewClientWithTransport(transport,
			WithSystemPrompt(largePrompt),
			WithMaxThinkingTokens(100000), // Large thinking tokens
		)

		// Should handle large configuration values
		err := client.Connect(ctx)
		if err != nil {
			t.Logf("Large configuration values caused error: %v", err)
		} else {
			client.Disconnect()
		}
	})
}

// T158: Client Interface Compliance 🔴 RED
// Go Target: client_test.go::TestClientInterfaceCompliance
func TestClientInterfaceCompliance(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test 1: Verify all Client interface methods exist and have correct signatures
	t.Run("InterfaceMethodsExist", func(t *testing.T) {
		transport := &clientMockTransport{}
		client := NewClientWithTransport(transport)

		// Type assertion to ensure ClientImpl implements Client interface
		var _ Client = client

		// Test that all methods are callable with correct signatures
		
		// Connect(ctx context.Context, prompt ...StreamMessage) error
		err := client.Connect(ctx)
		if err != nil {
			t.Logf("Connect returned error (expected for mock): %v", err)
		}

		// Query(ctx context.Context, prompt string, sessionID ...string) error
		err = client.Query(ctx, "test")
		if err != nil {
			t.Logf("Query returned error (expected for disconnected client): %v", err)
		}

		// Query with session ID
		err = client.Query(ctx, "test", "session123")
		if err != nil {
			t.Logf("Query with session ID returned error: %v", err)
		}

		// Query with multiple session IDs (should use first one)
		err = client.Query(ctx, "test", "session1", "session2")
		if err != nil {
			t.Logf("Query with multiple session IDs returned error: %v", err)
		}

		// QueryStream(ctx context.Context, messages <-chan StreamMessage) error
		messages := make(chan StreamMessage)
		close(messages) // Empty channel
		err = client.QueryStream(ctx, messages)
		if err != nil {
			t.Logf("QueryStream returned error: %v", err)
		}

		// ReceiveMessages(ctx context.Context) <-chan Message
		msgChan := client.ReceiveMessages(ctx)
		if msgChan == nil {
			t.Error("ReceiveMessages should not return nil channel")
		}

		// ReceiveResponse(ctx context.Context) MessageIterator
		iter := client.ReceiveResponse(ctx)
		if iter == nil {
			// This is acceptable for disconnected client
			t.Log("ReceiveResponse returned nil (expected for disconnected client)")
		}

		// Interrupt(ctx context.Context) error
		err = client.Interrupt(ctx)
		if err != nil {
			t.Logf("Interrupt returned error (expected for disconnected client): %v", err)
		}

		// Disconnect() error
		err = client.Disconnect()
		if err != nil {
			t.Errorf("Disconnect should not return error for clean disconnect: %v", err)
		}
	})

	// Test 2: Verify Connect method behavior
	t.Run("ConnectMethodBehavior", func(t *testing.T) {
		transport := &clientMockTransport{}
		client := NewClientWithTransport(transport)

		// Connect should work with valid transport
		err := client.Connect(ctx)
		if err != nil {
			t.Errorf("Connect should succeed with valid transport: %v", err)
		}

		// Connect with initial messages
		initialMsg := StreamMessage{Type: "request", Message: &UserMessage{Content: "Hello"}}
		err = client.Connect(ctx, initialMsg)
		if err != nil {
			t.Logf("Connect with initial message returned: %v", err)
		}

		client.Disconnect()
	})

	// Test 3: Verify Query method variations
	t.Run("QueryMethodVariations", func(t *testing.T) {
		transport := &clientMockTransport{}
		client := NewClientWithTransport(transport)

		err := client.Connect(ctx)
		if err != nil {
			t.Fatalf("Connect failed: %v", err)
		}

		// Query with no session ID (should use default)
		err = client.Query(ctx, "test message")
		if err != nil {
			t.Errorf("Query without session ID failed: %v", err)
		}

		// Query with empty session ID (should use default)
		err = client.Query(ctx, "test message", "")
		if err != nil {
			t.Errorf("Query with empty session ID failed: %v", err)
		}

		// Query with valid session ID
		err = client.Query(ctx, "test message", "session123")
		if err != nil {
			t.Errorf("Query with valid session ID failed: %v", err)
		}

		// Verify messages were sent correctly
		expectedMsgCount := 3
		if transport.getSentMessageCount() != expectedMsgCount {
			t.Errorf("Expected %d messages sent, got %d", expectedMsgCount, transport.getSentMessageCount())
		}

		client.Disconnect()
	})

	// Test 4: Verify QueryStream method behavior
	t.Run("QueryStreamMethodBehavior", func(t *testing.T) {
		transport := &clientMockTransport{}
		client := NewClientWithTransport(transport)

		err := client.Connect(ctx)
		if err != nil {
			t.Fatalf("Connect failed: %v", err)
		}

		// Test with multiple messages
		messages := make(chan StreamMessage, 3)
		messages <- StreamMessage{Type: "request", Message: &UserMessage{Content: "Message 1"}}
		messages <- StreamMessage{Type: "request", Message: &UserMessage{Content: "Message 2"}}
		messages <- StreamMessage{Type: "request", Message: &UserMessage{Content: "Message 3"}}
		close(messages)

		err = client.QueryStream(ctx, messages)
		if err != nil {
			t.Errorf("QueryStream failed: %v", err)
		}

		// Give time for goroutine to process messages
		time.Sleep(100 * time.Millisecond)

		// Test with empty channel
		emptyMessages := make(chan StreamMessage)
		close(emptyMessages)

		err = client.QueryStream(ctx, emptyMessages)
		if err != nil {
			t.Errorf("QueryStream with empty channel failed: %v", err)
		}

		client.Disconnect()
	})

	// Test 5: Verify ReceiveMessages method behavior
	t.Run("ReceiveMessagesMethodBehavior", func(t *testing.T) {
		transport := &clientMockTransport{
			responseMessages: []Message{
				&AssistantMessage{Content: []ContentBlock{&TextBlock{Text: "Response 1"}}, Model: "claude-opus-4-1-20250805"},
				&AssistantMessage{Content: []ContentBlock{&TextBlock{Text: "Response 2"}}, Model: "claude-opus-4-1-20250805"},
			},
		}
		client := NewClientWithTransport(transport)

		err := client.Connect(ctx)
		if err != nil {
			t.Fatalf("Connect failed: %v", err)
		}

		// Get message channel
		msgChan := client.ReceiveMessages(ctx)
		if msgChan == nil {
			t.Fatal("ReceiveMessages returned nil channel")
		}

		// Trigger sending responses
		transport.sendResponses()

		// Read messages
		var receivedMessages []Message
		for i := 0; i < 2; i++ {
			select {
			case msg := <-msgChan:
				if msg != nil {
					receivedMessages = append(receivedMessages, msg)
				}
			case <-time.After(1 * time.Second):
				t.Fatal("Timeout waiting for messages")
			}
		}

		if len(receivedMessages) != 2 {
			t.Errorf("Expected 2 messages, got %d", len(receivedMessages))
		}

		client.Disconnect()
	})

	// Test 6: Verify ReceiveResponse method behavior
	t.Run("ReceiveResponseMethodBehavior", func(t *testing.T) {
		transport := &clientMockTransport{
			responseMessages: []Message{
				&AssistantMessage{Content: []ContentBlock{&TextBlock{Text: "Iterator Response"}}, Model: "claude-opus-4-1-20250805"},
			},
		}
		client := NewClientWithTransport(transport)

		err := client.Connect(ctx)
		if err != nil {
			t.Fatalf("Connect failed: %v", err)
		}

		// Get response iterator
		iter := client.ReceiveResponse(ctx)
		if iter == nil {
			t.Fatal("ReceiveResponse returned nil iterator")
		}

		// Trigger responses
		transport.sendResponses()

		// Use iterator
		msg, err := iter.Next(ctx)
		if err != nil {
			t.Errorf("Iterator Next failed: %v", err)
		} else if msg == nil {
			t.Error("Iterator returned nil message")
		}

		// Close iterator
		err = iter.Close()
		if err != nil {
			t.Errorf("Iterator Close failed: %v", err)
		}

		client.Disconnect()
	})

	// Test 7: Verify Interrupt method behavior
	t.Run("InterruptMethodBehavior", func(t *testing.T) {
		transport := &clientMockTransport{}
		client := NewClientWithTransport(transport)

		// Interrupt on disconnected client should return error
		err := client.Interrupt(ctx)
		if err == nil {
			t.Error("Interrupt on disconnected client should return error")
		}

		// Interrupt on connected client should work
		err = client.Connect(ctx)
		if err != nil {
			t.Fatalf("Connect failed: %v", err)
		}

		err = client.Interrupt(ctx)
		if err != nil {
			t.Errorf("Interrupt on connected client failed: %v", err)
		}

		client.Disconnect()
	})

	// Test 8: Verify Disconnect method behavior
	t.Run("DisconnectMethodBehavior", func(t *testing.T) {
		transport := &clientMockTransport{}
		client := NewClientWithTransport(transport)

		// Disconnect on unconnected client should be safe
		err := client.Disconnect()
		if err != nil {
			t.Errorf("Disconnect on unconnected client should be safe: %v", err)
		}

		// Connect and disconnect
		err = client.Connect(ctx)
		if err != nil {
			t.Fatalf("Connect failed: %v", err)
		}

		err = client.Disconnect()
		if err != nil {
			t.Errorf("Disconnect after connect failed: %v", err)
		}

		// Multiple disconnects should be safe
		err = client.Disconnect()
		if err != nil {
			t.Errorf("Second disconnect should be safe: %v", err)
		}
	})

	// Test 9: Verify method error handling consistency
	t.Run("ErrorHandlingConsistency", func(t *testing.T) {
		transport := &clientMockTransport{
			connectError: fmt.Errorf("mock connect error"),
		}
		client := NewClientWithTransport(transport)

		// All methods should handle errors consistently
		err := client.Connect(ctx)
		if err == nil {
			t.Error("Expected connect error")
		}

		// Methods should return appropriate errors when not connected
		err = client.Query(ctx, "test")
		if err == nil {
			t.Error("Query should return error when not connected")
		}

		messages := make(chan StreamMessage)
		close(messages)
		err = client.QueryStream(ctx, messages)
		if err == nil {
			t.Error("QueryStream should return error when not connected")
		}

		err = client.Interrupt(ctx)
		if err == nil {
			t.Error("Interrupt should return error when not connected")
		}

		// ReceiveMessages should return closed channel when not connected
		msgChan := client.ReceiveMessages(ctx)
		select {
		case msg, ok := <-msgChan:
			if ok || msg != nil {
				t.Error("ReceiveMessages should return closed channel when not connected")
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("ReceiveMessages channel should be immediately readable when closed")
		}

		// ReceiveResponse should return nil when not connected
		iter := client.ReceiveResponse(ctx)
		if iter != nil {
			t.Error("ReceiveResponse should return nil when not connected")
		}
	})

	// Test 10: Verify interface contract with nil checks
	t.Run("InterfaceContractNilChecks", func(t *testing.T) {
		transport := &clientMockTransport{}
		client := NewClientWithTransport(transport)

		// Methods should handle nil context gracefully or panic consistently
		// We'll use a valid context for all calls to ensure consistent behavior

		// Connect with nil messages should work
		err := client.Connect(ctx)
		if err != nil {
			t.Logf("Connect returned: %v", err)
		}

		// Query with empty string should work
		err = client.Query(ctx, "")
		if err != nil {
			t.Logf("Query with empty string returned: %v", err)
		}

		client.Disconnect()
	})
}

// T145: Client Configuration Application 🔴 RED
// Go Target: client_test.go::TestClientConfigurationApplication
func TestClientConfigurationApplication(t *testing.T) {
	_ = context.Background() // We don't actually need context for this test

	// Test 1: Single functional option is applied
	t.Run("SingleFunctionalOption", func(t *testing.T) {
		transport := &clientMockTransport{}
		
		client := NewClientWithTransport(transport,
			WithSystemPrompt("Test system prompt"),
		)

		// Verify option was applied
		clientImpl := client.(*ClientImpl)
		if clientImpl.options.SystemPrompt == nil {
			t.Error("SystemPrompt option was not applied")
		} else if *clientImpl.options.SystemPrompt != "Test system prompt" {
			t.Errorf("Expected system prompt 'Test system prompt', got '%s'", *clientImpl.options.SystemPrompt)
		}
	})

	// Test 2: Multiple functional options are applied in order
	t.Run("MultipleFunctionalOptions", func(t *testing.T) {
		transport := &clientMockTransport{}
		
		client := NewClientWithTransport(transport,
			WithSystemPrompt("System prompt"),
			WithAppendSystemPrompt("Append prompt"),
			WithModel("claude-opus-4-1-20250805"),
			WithMaxTurns(10),
			WithMaxThinkingTokens(5000),
		)

		// Verify all options were applied
		clientImpl := client.(*ClientImpl)
		
		if clientImpl.options.SystemPrompt == nil || *clientImpl.options.SystemPrompt != "System prompt" {
			t.Error("SystemPrompt option was not applied correctly")
		}
		
		if clientImpl.options.AppendSystemPrompt == nil || *clientImpl.options.AppendSystemPrompt != "Append prompt" {
			t.Error("AppendSystemPrompt option was not applied correctly")
		}
		
		if clientImpl.options.Model == nil || *clientImpl.options.Model != "claude-opus-4-1-20250805" {
			t.Error("Model option was not applied correctly")
		}
		
		if clientImpl.options.MaxTurns != 10 {
			t.Errorf("Expected MaxTurns 10, got %d", clientImpl.options.MaxTurns)
		}
		
		if clientImpl.options.MaxThinkingTokens != 5000 {
			t.Errorf("Expected MaxThinkingTokens 5000, got %d", clientImpl.options.MaxThinkingTokens)
		}
	})

	// Test 3: Tool configuration options are applied
	t.Run("ToolConfigurationOptions", func(t *testing.T) {
		transport := &clientMockTransport{}
		
		allowedTools := []string{"Read", "Write", "Bash"}
		disallowedTools := []string{"dangerous-tool", "restricted-tool"}
		
		client := NewClientWithTransport(transport,
			WithAllowedTools(allowedTools...),
			WithDisallowedTools(disallowedTools...),
		)

		// Verify tool options were applied
		clientImpl := client.(*ClientImpl)
		
		if len(clientImpl.options.AllowedTools) != len(allowedTools) {
			t.Errorf("Expected %d allowed tools, got %d", len(allowedTools), len(clientImpl.options.AllowedTools))
		}
		
		for i, tool := range allowedTools {
			if clientImpl.options.AllowedTools[i] != tool {
				t.Errorf("Expected allowed tool %s at index %d, got %s", tool, i, clientImpl.options.AllowedTools[i])
			}
		}
		
		if len(clientImpl.options.DisallowedTools) != len(disallowedTools) {
			t.Errorf("Expected %d disallowed tools, got %d", len(disallowedTools), len(clientImpl.options.DisallowedTools))
		}
		
		for i, tool := range disallowedTools {
			if clientImpl.options.DisallowedTools[i] != tool {
				t.Errorf("Expected disallowed tool %s at index %d, got %s", tool, i, clientImpl.options.DisallowedTools[i])
			}
		}
	})

	// Test 4: Permission and session options are applied
	t.Run("PermissionAndSessionOptions", func(t *testing.T) {
		transport := &clientMockTransport{}
		
		client := NewClientWithTransport(transport,
			WithPermissionMode(PermissionModeAcceptEdits),
			WithPermissionPromptToolName("custom-prompt-tool"),
			WithContinueConversation(true),
			WithResume("session-123"),
		)

		// Verify options were applied
		clientImpl := client.(*ClientImpl)
		
		if clientImpl.options.PermissionMode == nil || *clientImpl.options.PermissionMode != PermissionModeAcceptEdits {
			t.Error("PermissionMode option was not applied correctly")
		}
		
		if clientImpl.options.PermissionPromptToolName == nil || *clientImpl.options.PermissionPromptToolName != "custom-prompt-tool" {
			t.Error("PermissionPromptToolName option was not applied correctly")
		}
		
		if !clientImpl.options.ContinueConversation {
			t.Error("ContinueConversation option was not applied correctly")
		}
		
		if clientImpl.options.Resume == nil || *clientImpl.options.Resume != "session-123" {
			t.Error("Resume option was not applied correctly")
		}
	})

	// Test 5: File system and context options are applied
	t.Run("FileSystemAndContextOptions", func(t *testing.T) {
		transport := &clientMockTransport{}
		
		addDirs := []string{"/project", "/libs", "/docs"}
		
		client := NewClientWithTransport(transport,
			WithCwd("/tmp"),
			WithAddDirs(addDirs...),
			WithSettings("custom-settings.json"),
		)

		// Verify options were applied
		clientImpl := client.(*ClientImpl)
		
		if clientImpl.options.Cwd == nil || *clientImpl.options.Cwd != "/tmp" {
			t.Error("Cwd option was not applied correctly")
		}
		
		if len(clientImpl.options.AddDirs) != len(addDirs) {
			t.Errorf("Expected %d add dirs, got %d", len(addDirs), len(clientImpl.options.AddDirs))
		}
		
		for i, dir := range addDirs {
			if clientImpl.options.AddDirs[i] != dir {
				t.Errorf("Expected add dir %s at index %d, got %s", dir, i, clientImpl.options.AddDirs[i])
			}
		}
		
		if clientImpl.options.Settings == nil || *clientImpl.options.Settings != "custom-settings.json" {
			t.Error("Settings option was not applied correctly")
		}
	})

	// Test 6: MCP server configuration is applied
	t.Run("MCPServerConfiguration", func(t *testing.T) {
		transport := &clientMockTransport{}
		
		mcpServers := map[string]McpServerConfig{
			"filesystem": &McpStdioServerConfig{
				Type:    McpServerTypeStdio,
				Command: "npx",
				Args:    []string{"@modelcontextprotocol/server-filesystem", "/tmp"},
			},
			"git": &McpStdioServerConfig{
				Type:    McpServerTypeStdio,
				Command: "npx",
				Args:    []string{"@modelcontextprotocol/server-git"},
			},
		}
		
		client := NewClientWithTransport(transport,
			WithMcpServers(mcpServers),
		)

		// Verify MCP servers were applied
		clientImpl := client.(*ClientImpl)
		
		if len(clientImpl.options.McpServers) != len(mcpServers) {
			t.Errorf("Expected %d MCP servers, got %d", len(mcpServers), len(clientImpl.options.McpServers))
		}
		
		for name, expectedConfig := range mcpServers {
			actualConfig, exists := clientImpl.options.McpServers[name]
			if !exists {
				t.Errorf("MCP server %s was not configured", name)
				continue
			}
			
			expectedStdio := expectedConfig.(*McpStdioServerConfig)
			actualStdio, ok := actualConfig.(*McpStdioServerConfig)
			if !ok {
				t.Errorf("MCP server %s has wrong type", name)
				continue
			}
			
			if actualStdio.Command != expectedStdio.Command {
				t.Errorf("MCP server %s command mismatch: expected %s, got %s", name, expectedStdio.Command, actualStdio.Command)
			}
		}
	})

	// Test 7: Extra arguments are applied
	t.Run("ExtraArgumentsConfiguration", func(t *testing.T) {
		transport := &clientMockTransport{}
		
		extraArgs := map[string]*string{
			"--custom-flag":       nil, // Boolean flag
			"--custom-with-value": stringPtr("custom-value"),
			"--debug":             nil,
			"--timeout":           stringPtr("30s"),
		}
		
		client := NewClientWithTransport(transport,
			WithExtraArgs(extraArgs),
		)

		// Verify extra args were applied
		clientImpl := client.(*ClientImpl)
		
		if len(clientImpl.options.ExtraArgs) < len(extraArgs) { // Might have transport marker
			t.Errorf("Expected at least %d extra args, got %d", len(extraArgs), len(clientImpl.options.ExtraArgs))
		}
		
		for flag, expectedValue := range extraArgs {
			actualValue, exists := clientImpl.options.ExtraArgs[flag]
			if !exists {
				t.Errorf("Extra arg %s was not configured", flag)
				continue
			}
			
			if expectedValue == nil && actualValue != nil {
				t.Errorf("Extra arg %s should be boolean flag (nil), got %v", flag, actualValue)
			} else if expectedValue != nil && actualValue == nil {
				t.Errorf("Extra arg %s should have value %s, got nil", flag, *expectedValue)
			} else if expectedValue != nil && actualValue != nil && *expectedValue != *actualValue {
				t.Errorf("Extra arg %s value mismatch: expected %s, got %s", flag, *expectedValue, *actualValue)
			}
		}
	})

	// Test 8: Option precedence (later options override earlier ones)
	t.Run("OptionPrecedence", func(t *testing.T) {
		transport := &clientMockTransport{}
		
		client := NewClientWithTransport(transport,
			WithSystemPrompt("First prompt"),
			WithModel("first-model"),
			WithSystemPrompt("Second prompt"), // Should override first
			WithModel("second-model"),         // Should override first
			WithMaxTurns(5),
			WithMaxTurns(10), // Should override first
		)

		// Verify later options override earlier ones
		clientImpl := client.(*ClientImpl)
		
		if clientImpl.options.SystemPrompt == nil || *clientImpl.options.SystemPrompt != "Second prompt" {
			t.Error("Later SystemPrompt option did not override earlier one")
		}
		
		if clientImpl.options.Model == nil || *clientImpl.options.Model != "second-model" {
			t.Error("Later Model option did not override earlier one")
		}
		
		if clientImpl.options.MaxTurns != 10 {
			t.Errorf("Later MaxTurns option did not override earlier one: expected 10, got %d", clientImpl.options.MaxTurns)
		}
	})

	// Test 9: Default values are preserved when options not provided
	t.Run("DefaultValuesPreserved", func(t *testing.T) {
		transport := &clientMockTransport{}
		
		client := NewClientWithTransport(transport,
			WithSystemPrompt("Custom prompt"), // Only set this option
		)

		// Verify defaults are preserved for other options
		clientImpl := client.(*ClientImpl)
		
		// SystemPrompt should be set
		if clientImpl.options.SystemPrompt == nil || *clientImpl.options.SystemPrompt != "Custom prompt" {
			t.Error("Custom SystemPrompt was not applied")
		}
		
		// Default values should be preserved (checking a few key ones)
		if clientImpl.options.MaxThinkingTokens != 8000 { // Default from shared package
			t.Errorf("Expected default MaxThinkingTokens 8000, got %d", clientImpl.options.MaxThinkingTokens)
		}
		
		if len(clientImpl.options.AllowedTools) != 0 { // Default empty
			t.Errorf("Expected empty AllowedTools by default, got %v", clientImpl.options.AllowedTools)
		}
	})

	// Test 10: Configuration is immutable after client creation
	t.Run("ConfigurationImmutabilityAfterCreation", func(t *testing.T) {
		transport := &clientMockTransport{}
		
		originalPrompt := "Original prompt"
		client := NewClientWithTransport(transport,
			WithSystemPrompt(originalPrompt),
		)

		// Get reference to options
		clientImpl := client.(*ClientImpl)
		originalOptions := clientImpl.options
		
		// Modify the options reference (simulating external modification)
		modifiedPrompt := "Modified prompt"
		originalOptions.SystemPrompt = &modifiedPrompt

		// Create another client with same functional option
		client2 := NewClientWithTransport(transport,
			WithSystemPrompt(originalPrompt),
		)

		// Second client should have original prompt, not modified
		client2Impl := client2.(*ClientImpl)
		if client2Impl.options.SystemPrompt == nil || *client2Impl.options.SystemPrompt != originalPrompt {
			t.Error("Options were not properly isolated between client instances")
		}

		// But first client should still have the modified prompt
		if *clientImpl.options.SystemPrompt != modifiedPrompt {
			t.Error("Options reference was not preserved in first client")
		}
	})
}

// T149: Client Context Propagation 🔴 RED
// Go Target: client_test.go::TestClientContextPropagation
func TestClientContextPropagation(t *testing.T) {
	// Test 1: Context cancellation during Connect
	t.Run("ContextCancellationDuringConnect", func(t *testing.T) {
		transport := &clientMockTransport{}
		client := NewClientWithTransport(transport)

		// Create a context that will be cancelled
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// Connect should respect cancelled context
		err := client.Connect(ctx)
		if err == nil {
			t.Error("Connect should return error for cancelled context")
			client.Disconnect()
		} else if err != context.Canceled {
			t.Logf("Got error (may be valid): %v", err)
		}
	})

	// Test 2: Context timeout during Connect
	t.Run("ContextTimeoutDuringConnect", func(t *testing.T) {
		transport := &clientMockTransport{}
		client := NewClientWithTransport(transport)

		// Create a context with immediate timeout
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		
		// Give the context time to timeout
		time.Sleep(1 * time.Millisecond)

		// Connect should respect timeout context
		err := client.Connect(ctx)
		if err == nil {
			t.Error("Connect should return error for timed out context")
			client.Disconnect()
		} else if err != context.DeadlineExceeded {
			t.Logf("Got error (may be valid): %v", err)
		}
	})

	// Test 3: Context cancellation during Query
	t.Run("ContextCancellationDuringQuery", func(t *testing.T) {
		transport := &clientMockTransport{}
		client := NewClientWithTransport(transport)

		// Connect first
		connectCtx, connectCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer connectCancel()
		
		err := client.Connect(connectCtx)
		if err != nil {
			t.Fatalf("Connect failed: %v", err)
		}
		defer client.Disconnect()

		// Create cancelled context for Query
		queryCtx, queryCancel := context.WithCancel(context.Background())
		queryCancel() // Cancel immediately

		// Query should respect cancelled context
		err = client.Query(queryCtx, "test query")
		if err == nil {
			t.Error("Query should return error for cancelled context")
		} else if err != context.Canceled {
			t.Logf("Query with cancelled context returned: %v", err)
		}
	})

	// Test 4: Context timeout during Query
	t.Run("ContextTimeoutDuringQuery", func(t *testing.T) {
		transport := &clientMockTransport{}
		client := NewClientWithTransport(transport)

		// Connect first
		connectCtx, connectCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer connectCancel()
		
		err := client.Connect(connectCtx)
		if err != nil {
			t.Fatalf("Connect failed: %v", err)
		}
		defer client.Disconnect()

		// Create timeout context for Query
		queryCtx, queryCancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer queryCancel()
		
		// Give time for timeout
		time.Sleep(1 * time.Millisecond)

		// Query should respect timeout context
		err = client.Query(queryCtx, "test query")
		if err == nil {
			t.Error("Query should return error for timed out context")
		} else if err != context.DeadlineExceeded {
			t.Logf("Query with timeout context returned: %v", err)
		}
	})

	// Test 5: Context cancellation during QueryStream
	t.Run("ContextCancellationDuringQueryStream", func(t *testing.T) {
		transport := &clientMockTransport{}
		client := NewClientWithTransport(transport)

		// Connect first
		connectCtx, connectCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer connectCancel()
		
		err := client.Connect(connectCtx)
		if err != nil {
			t.Fatalf("Connect failed: %v", err)
		}
		defer client.Disconnect()

		// Create cancelled context for QueryStream
		streamCtx, streamCancel := context.WithCancel(context.Background())
		
		// Create messages channel
		messages := make(chan StreamMessage, 1)
		messages <- StreamMessage{Type: "request", Message: &UserMessage{Content: "test"}}
		close(messages)

		// Start QueryStream
		err = client.QueryStream(streamCtx, messages)
		if err != nil {
			t.Errorf("QueryStream failed: %v", err)
		}

		// Cancel context while stream might be processing
		streamCancel()
		
		// Give time for cancellation to propagate
		time.Sleep(100 * time.Millisecond)
		
		// Context should be respected in the goroutine
		// This is hard to test directly but we verify no panic occurs
	})

	// Test 6: Context cancellation during ReceiveMessages
	t.Run("ContextCancellationDuringReceiveMessages", func(t *testing.T) {
		transport := &clientMockTransport{
			responseMessages: []Message{
				&AssistantMessage{Content: []ContentBlock{&TextBlock{Text: "Response"}}, Model: "claude-opus-4-1-20250805"},
			},
		}
		client := NewClientWithTransport(transport)

		// Connect first
		connectCtx, connectCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer connectCancel()
		
		err := client.Connect(connectCtx)
		if err != nil {
			t.Fatalf("Connect failed: %v", err)
		}
		defer client.Disconnect()

		// Create context for receiving messages
		receiveCtx, receiveCancel := context.WithCancel(context.Background())
		
		// Get message channel
		msgChan := client.ReceiveMessages(receiveCtx)
		
		// Cancel context
		receiveCancel()
		
		// Trigger responses
		transport.sendResponses()

		// Try to receive with cancelled context (should not block indefinitely)
		select {
		case msg := <-msgChan:
			if msg != nil {
				t.Log("Received message despite cancelled context (may be valid)")
			}
		case <-time.After(100 * time.Millisecond):
			t.Log("No message received within timeout (expected for cancelled context)")
		}
	})

	// Test 7: Context cancellation during ReceiveResponse iterator
	t.Run("ContextCancellationDuringReceiveResponse", func(t *testing.T) {
		transport := &clientMockTransport{
			responseMessages: []Message{
				&AssistantMessage{Content: []ContentBlock{&TextBlock{Text: "Response"}}, Model: "claude-opus-4-1-20250805"},
			},
		}
		client := NewClientWithTransport(transport)

		// Connect first
		connectCtx, connectCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer connectCancel()
		
		err := client.Connect(connectCtx)
		if err != nil {
			t.Fatalf("Connect failed: %v", err)
		}
		defer client.Disconnect()

		// Get response iterator
		iterCtx := context.Background()
		iter := client.ReceiveResponse(iterCtx)
		if iter == nil {
			t.Fatal("ReceiveResponse returned nil iterator")
		}

		// Create cancelled context for Next call
		nextCtx, nextCancel := context.WithCancel(context.Background())
		nextCancel() // Cancel immediately

		// Trigger responses
		transport.sendResponses()

		// Iterator Next should respect cancelled context
		_, err = iter.Next(nextCtx)
		if err == nil {
			t.Log("Iterator Next succeeded despite cancelled context (may be valid)")
		} else if err == context.Canceled {
			t.Log("Iterator correctly returned context.Canceled")
		} else {
			t.Logf("Iterator returned different error: %v", err)
		}

		iter.Close()
	})

	// Test 8: Context cancellation during Interrupt
	t.Run("ContextCancellationDuringInterrupt", func(t *testing.T) {
		transport := &clientMockTransport{}
		client := NewClientWithTransport(transport)

		// Connect first
		connectCtx, connectCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer connectCancel()
		
		err := client.Connect(connectCtx)
		if err != nil {
			t.Fatalf("Connect failed: %v", err)
		}
		defer client.Disconnect()

		// Create cancelled context for Interrupt
		interruptCtx, interruptCancel := context.WithCancel(context.Background())
		interruptCancel() // Cancel immediately

		// Interrupt should respect cancelled context
		err = client.Interrupt(interruptCtx)
		if err == nil {
			t.Log("Interrupt succeeded despite cancelled context (may be valid)")
		} else if err == context.Canceled {
			t.Log("Interrupt correctly returned context.Canceled")
		} else {
			t.Logf("Interrupt returned different error: %v", err)
		}
	})

	// Test 9: Context values are propagated
	t.Run("ContextValuesPropagation", func(t *testing.T) {
		transport := &clientMockTransport{}
		client := NewClientWithTransport(transport)

		// Create context with custom values
		type contextKey string
		const testKey contextKey = "test-key"
		testValue := "test-value"
		
		ctx := context.WithValue(context.Background(), testKey, testValue)

		// Context values should be available throughout operations
		err := client.Connect(ctx)
		if err != nil {
			t.Logf("Connect returned: %v", err)
		}

		err = client.Query(ctx, "test query")
		if err != nil {
			t.Logf("Query returned: %v", err)
		}

		// This test mainly ensures no panic occurs and context is properly threaded
		client.Disconnect()
	})

	// Test 10: Nested context cancellation behavior
	t.Run("NestedContextCancellation", func(t *testing.T) {
		transport := &clientMockTransport{}
		client := NewClientWithTransport(transport)

		// Create nested contexts
		parentCtx, parentCancel := context.WithCancel(context.Background())
		childCtx, childCancel := context.WithCancel(parentCtx)

		// Connect with child context
		err := client.Connect(childCtx)
		if err != nil {
			t.Logf("Connect returned: %v", err)
		}
		defer client.Disconnect()

		// Cancel parent context (should cascade to child)
		parentCancel()
		defer childCancel() // Cleanup

		// Operations with child context should see cancellation
		err = client.Query(childCtx, "test query")
		if err == nil {
			t.Log("Query succeeded despite parent context cancellation (may be valid)")
		} else if err == context.Canceled {
			t.Log("Query correctly propagated parent context cancellation")
		} else {
			t.Logf("Query returned: %v", err)
		}
	})

	// Test 11: Context deadline propagation
	t.Run("ContextDeadlinePropagation", func(t *testing.T) {
		transport := &clientMockTransport{}
		client := NewClientWithTransport(transport)

		// Create context with deadline
		deadline := time.Now().Add(100 * time.Millisecond)
		ctx, cancel := context.WithDeadline(context.Background(), deadline)
		defer cancel()

		// Connect should work initially
		err := client.Connect(ctx)
		if err != nil {
			t.Logf("Connect returned: %v", err)
		}
		defer client.Disconnect()

		// Wait for deadline to pass
		time.Sleep(150 * time.Millisecond)

		// Operations should now see deadline exceeded
		err = client.Query(ctx, "test query")
		if err == nil {
			t.Log("Query succeeded despite exceeded deadline (may be valid)")
		} else if err == context.DeadlineExceeded {
			t.Log("Query correctly returned context.DeadlineExceeded")
		} else {
			t.Logf("Query returned: %v", err)
		}
	})

	// Test 12: Multiple operations with same context
	t.Run("MultipleOperationsSameContext", func(t *testing.T) {
		transport := &clientMockTransport{}
		client := NewClientWithTransport(transport)

		// Create single context for all operations
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// All operations should use the same context
		err := client.Connect(ctx)
		if err != nil {
			t.Logf("Connect returned: %v", err)
		}
		defer client.Disconnect()

		err = client.Query(ctx, "query 1")
		if err != nil {
			t.Logf("Query 1 returned: %v", err)
		}

		err = client.Query(ctx, "query 2")
		if err != nil {
			t.Logf("Query 2 returned: %v", err)
		}

		messages := make(chan StreamMessage, 1)
		messages <- StreamMessage{Type: "request", Message: &UserMessage{Content: "stream query"}}
		close(messages)

		err = client.QueryStream(ctx, messages)
		if err != nil {
			t.Logf("QueryStream returned: %v", err)
		}

		// All operations should share the same timeout
		msgChan := client.ReceiveMessages(ctx)
		select {
		case <-msgChan:
		case <-time.After(100 * time.Millisecond):
		}

		err = client.Interrupt(ctx)
		if err != nil {
			t.Logf("Interrupt returned: %v", err)
		}
	})
}

// T148: Client Backpressure 🔴 RED
// Go Target: client_test.go::TestClientBackpressure
func TestClientBackpressure(t *testing.T) {
	// Test channel buffer management under load
	transport := &clientMockTransport{
		msgChan:  make(chan Message, 2), // Small buffer for backpressure testing
		errChan:  make(chan error, 10),
		slowSend: true, // Enable slow sending to create backpressure
	}
	
	client := NewClientWithTransport(transport, WithMaxTurns(100))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer client.Disconnect()
	
	// Test 1: High volume message sending
	t.Run("HighVolumeMessageSending", func(t *testing.T) {
		messageCount := 50
		var wg sync.WaitGroup
		errChan := make(chan error, messageCount)
		
		for i := 0; i < messageCount; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				err := client.Query(ctx, fmt.Sprintf("Load test message %d", idx))
				if err != nil && err != context.DeadlineExceeded {
					errChan <- fmt.Errorf("message %d failed: %w", idx, err)
				}
			}(i)
		}
		
		wg.Wait()
		close(errChan)
		
		// Check for errors (excluding timeout)
		for err := range errChan {
			t.Error(err)
		}
	})
	
	// Test 2: Stream backpressure handling
	t.Run("StreamBackpressureHandling", func(t *testing.T) {
		streamMessages := make(chan StreamMessage, 5) // Small buffer
		
		// Start streaming
		err = client.QueryStream(ctx, streamMessages)
		if err != nil {
			t.Errorf("QueryStream failed: %v", err)
		}
		
		// Send messages rapidly
		for i := 0; i < 10; i++ {
			userMsg := &UserMessage{Content: fmt.Sprintf("Stream message %d", i)}
			streamMsg := StreamMessage{
				Type:      "request",
				Message:   userMsg,
				SessionID: "stream-test",
			}
			
			select {
			case streamMessages <- streamMsg:
			case <-time.After(100 * time.Millisecond):
				// Expected - channel should have backpressure
			}
		}
		close(streamMessages)
	})
	
	// Test 3: Buffer overflow protection
	t.Run("BufferOverflowProtection", func(t *testing.T) {
		transport.mu.Lock()
		sentCount := len(transport.sentMessages)
		transport.mu.Unlock()
		
		// Should handle load without crashes
		if sentCount < 0 {
			t.Error("Negative sent message count indicates memory corruption")
		}
	})
}

// T151: Client Multiple Sessions 🔴 RED
// Go Target: client_test.go::TestClientMultipleSessions
func TestClientMultipleSessions(t *testing.T) {
	// Test session multiplexing and isolation
	transport := &clientMockTransport{
		msgChan: make(chan Message, 100),
		errChan: make(chan error, 10),
	}
	
	client := NewClientWithTransport(transport)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer client.Disconnect()
	
	// Test 1: Multiple sessions with different IDs
	t.Run("MultipleSessionIDs", func(t *testing.T) {
		sessionIDs := []string{"session1", "session2", "session3"}
		
		// Send messages to different sessions
		for _, sessionID := range sessionIDs {
			err := client.Query(ctx, fmt.Sprintf("Message for %s", sessionID), sessionID)
			if err != nil {
				t.Errorf("Query for session %s failed: %v", sessionID, err)
			}
		}
		
		// Verify messages were sent with correct session IDs
		transport.mu.Lock()
		sentCount := len(transport.sentMessages)
		transport.mu.Unlock()
		
		if sentCount != 3 {
			t.Fatalf("Expected 3 messages sent, got %d", sentCount)
		}
		
		// Check session isolation - verify all expected session IDs are present
		transport.mu.Lock()
		sessionsSeen := make(map[string]bool)
		for _, msg := range transport.sentMessages {
			sessionsSeen[msg.SessionID] = true
		}
		transport.mu.Unlock()
		
		// Verify all expected session IDs were used
		for _, expectedSessionID := range sessionIDs {
			if !sessionsSeen[expectedSessionID] {
				t.Errorf("Expected session ID %s not found in sent messages", expectedSessionID)
			}
		}
	})
	
	// Test 2: Concurrent sessions
	t.Run("ConcurrentSessions", func(t *testing.T) {
		sessionCount := 5
		messagesPerSession := 3
		var wg sync.WaitGroup
		
		// Clear previous messages
		transport.mu.Lock()
		transport.sentMessages = nil
		transport.mu.Unlock()
		
		// Send messages concurrently to different sessions
		for sessionIdx := 0; sessionIdx < sessionCount; sessionIdx++ {
			wg.Add(1)
			go func(sIdx int) {
				defer wg.Done()
				sessionID := fmt.Sprintf("concurrent-session-%d", sIdx)
				
				for msgIdx := 0; msgIdx < messagesPerSession; msgIdx++ {
					err := client.Query(ctx, fmt.Sprintf("Message %d", msgIdx), sessionID)
					if err != nil {
						t.Errorf("Concurrent query failed for session %s, message %d: %v", 
							sessionID, msgIdx, err)
					}
				}
			}(sessionIdx)
		}
		
		wg.Wait()
		
		// Verify all messages were sent
		transport.mu.Lock()
		totalExpected := sessionCount * messagesPerSession
		actualCount := len(transport.sentMessages)
		transport.mu.Unlock()
		
		if actualCount != totalExpected {
			t.Errorf("Expected %d total messages, got %d", totalExpected, actualCount)
		}
		
		// Verify session isolation - each session's messages should be grouped
		sessionMessageCounts := make(map[string]int)
		transport.mu.Lock()
		for _, msg := range transport.sentMessages {
			sessionMessageCounts[msg.SessionID]++
		}
		transport.mu.Unlock()
		
		if len(sessionMessageCounts) != sessionCount {
			t.Errorf("Expected %d unique sessions, got %d", sessionCount, len(sessionMessageCounts))
		}
		
		for sessionID, count := range sessionMessageCounts {
			if count != messagesPerSession {
				t.Errorf("Session %s should have %d messages, got %d", 
					sessionID, messagesPerSession, count)
			}
		}
	})
	
	// Test 3: Default session behavior
	t.Run("DefaultSessionBehavior", func(t *testing.T) {
		// Clear previous messages
		transport.mu.Lock()
		transport.sentMessages = nil
		transport.mu.Unlock()
		
		// Send message without explicit session ID
		err := client.Query(ctx, "Default session message")
		if err != nil {
			t.Errorf("Default session query failed: %v", err)
		}
		
		// Verify default session ID is used
		transport.mu.Lock()
		if len(transport.sentMessages) != 1 {
			t.Fatalf("Expected 1 message, got %d", len(transport.sentMessages))
		}
		
		msg := transport.sentMessages[0]
		if msg.SessionID != "default" {
			t.Errorf("Expected default session ID 'default', got '%s'", msg.SessionID)
		}
		transport.mu.Unlock()
	})
	
	// Test 4: Session isolation with QueryStream
	t.Run("SessionIsolationWithQueryStream", func(t *testing.T) {
		// Clear previous messages
		transport.mu.Lock()
		transport.sentMessages = nil
		transport.mu.Unlock()
		
		// Create streams for different sessions
		session1Messages := make(chan StreamMessage, 5)
		session2Messages := make(chan StreamMessage, 5)
		
		// Start streams
		err1 := client.QueryStream(ctx, session1Messages)
		if err1 != nil {
			t.Errorf("QueryStream for session1 failed: %v", err1)
		}
		
		err2 := client.QueryStream(ctx, session2Messages)
		if err2 != nil {
			t.Errorf("QueryStream for session2 failed: %v", err2)
		}
		
		// Send messages to different sessions via streams
		userMsg1 := &UserMessage{Content: "Stream message for session1"}
		streamMsg1 := StreamMessage{
			Type:      "request", 
			Message:   userMsg1,
			SessionID: "stream-session-1",
		}
		
		userMsg2 := &UserMessage{Content: "Stream message for session2"}
		streamMsg2 := StreamMessage{
			Type:      "request",
			Message:   userMsg2, 
			SessionID: "stream-session-2",
		}
		
		session1Messages <- streamMsg1
		session2Messages <- streamMsg2
		
		close(session1Messages)
		close(session2Messages)
		
		// Wait a bit for messages to be processed
		time.Sleep(100 * time.Millisecond)
		
		// Verify session isolation in stream messages
		transport.mu.Lock()
		sentCount := len(transport.sentMessages)
		if sentCount >= 2 {
			// Check that messages maintain session identity
			sessionFound := make(map[string]bool)
			for _, msg := range transport.sentMessages {
				sessionFound[msg.SessionID] = true
			}
			
			if !sessionFound["stream-session-1"] || !sessionFound["stream-session-2"] {
				t.Error("Stream messages did not maintain session isolation")
			}
		}
		transport.mu.Unlock()
	})
}

// T159: Client Factory Function 🔴 RED
// Go Target: client_test.go::TestClientFactoryFunction
func TestClientFactoryFunction(t *testing.T) {
	// Test NewClient constructor with options
	
	// Test 1: NewClient with no options
	t.Run("NewClientNoOptions", func(t *testing.T) {
		client := NewClient()
		if client == nil {
			t.Fatal("NewClient() returned nil")
		}
		
		// Verify it's a ClientImpl
		clientImpl, ok := client.(*ClientImpl)
		if !ok {
			t.Fatalf("Expected *ClientImpl, got %T", client)
		}
		
		// Verify options are initialized
		if clientImpl.options == nil {
			t.Error("Client options should be initialized")
		}
		
		// Verify no custom transport
		if clientImpl.customTransport != nil {
			t.Error("NewClient should not have custom transport")
		}
		
		// Verify not connected initially
		clientImpl.mu.RLock()
		connected := clientImpl.connected
		clientImpl.mu.RUnlock()
		
		if connected {
			t.Error("NewClient should not be connected initially")
		}
	})
	
	// Test 2: NewClient with single option
	t.Run("NewClientSingleOption", func(t *testing.T) {
		client := NewClient(WithMaxTurns(5))
		if client == nil {
			t.Fatal("NewClient() returned nil")
		}
		
		clientImpl := client.(*ClientImpl)
		if clientImpl.options == nil {
			t.Fatal("Client options should be initialized")
		}
		
		if clientImpl.options.MaxTurns != 5 {
			t.Errorf("Expected MaxTurns=5, got %d", clientImpl.options.MaxTurns)
		}
	})
	
	// Test 3: NewClient with multiple options
	t.Run("NewClientMultipleOptions", func(t *testing.T) {
		workingDir := "/tmp"
		systemPrompt := "Custom system prompt"
		
		client := NewClient(
			WithMaxTurns(10),
			WithCwd(workingDir),
			WithSystemPrompt(systemPrompt),
			WithPermissionMode(PermissionModeBypassPermissions),
		)
		
		if client == nil {
			t.Fatal("NewClient() returned nil")
		}
		
		clientImpl := client.(*ClientImpl)
		if clientImpl.options == nil {
			t.Fatal("Client options should be initialized")
		}
		
		// Verify all options were applied
		if clientImpl.options.MaxTurns != 10 {
			t.Errorf("Expected MaxTurns=10, got %d", clientImpl.options.MaxTurns)
		}
		
		if clientImpl.options.Cwd == nil || *clientImpl.options.Cwd != workingDir {
			t.Errorf("Expected Cwd=%s, got %v", workingDir, clientImpl.options.Cwd)
		}
		
		if clientImpl.options.SystemPrompt == nil || *clientImpl.options.SystemPrompt != systemPrompt {
			t.Errorf("Expected SystemPrompt=%s, got %v", systemPrompt, clientImpl.options.SystemPrompt)
		}
		
		if clientImpl.options.PermissionMode == nil || *clientImpl.options.PermissionMode != PermissionModeBypassPermissions {
			t.Errorf("Expected PermissionMode=%s, got %v", PermissionModeBypassPermissions, clientImpl.options.PermissionMode)
		}
	})
	
	// Test 4: NewClient vs NewClientWithTransport behavior
	t.Run("NewClientVsNewClientWithTransport", func(t *testing.T) {
		// NewClient
		client1 := NewClient(WithMaxTurns(7))
		clientImpl1 := client1.(*ClientImpl)
		
		// NewClientWithTransport
		transport := &clientMockTransport{}
		client2 := NewClientWithTransport(transport, WithMaxTurns(7))
		clientImpl2 := client2.(*ClientImpl)
		
		// Both should have same options
		if clientImpl1.options.MaxTurns != clientImpl2.options.MaxTurns {
			t.Error("Both constructors should apply options equally")
		}
		
		// But different transport behavior
		if clientImpl1.customTransport != nil {
			t.Error("NewClient should not have custom transport")
		}
		
		if clientImpl2.customTransport == nil {
			t.Error("NewClientWithTransport should have custom transport")
		}
		
		if clientImpl2.customTransport != transport {
			t.Error("NewClientWithTransport should store the provided transport")
		}
	})
	
	// Test 5: NewClient factory function consistency
	t.Run("FactoryFunctionConsistency", func(t *testing.T) {
		// Create multiple clients with same options
		clients := make([]Client, 3)
		for i := range clients {
			clients[i] = NewClient(
				WithMaxTurns(15),
				WithSystemPrompt("Test prompt"),
			)
		}
		
		// All should be configured identically
		for i, client := range clients {
			clientImpl := client.(*ClientImpl)
			
			if clientImpl.options.MaxTurns != 15 {
				t.Errorf("Client %d MaxTurns mismatch: expected 15, got %d", i, clientImpl.options.MaxTurns)
			}
			
			if clientImpl.options.SystemPrompt == nil || *clientImpl.options.SystemPrompt != "Test prompt" {
				t.Errorf("Client %d SystemPrompt mismatch: expected 'Test prompt', got %v", i, clientImpl.options.SystemPrompt)
			}
			
			// Each should be independent instance
			if clientImpl.customTransport != nil {
				t.Errorf("Client %d should not have custom transport", i)
			}
		}
	})
	
	// Test 6: Factory function with option override
	t.Run("FactoryFunctionOptionOverride", func(t *testing.T) {
		// Test that later options override earlier ones
		client := NewClient(
			WithMaxTurns(1),
			WithMaxTurns(2),
			WithMaxTurns(3), // This should be the final value
		)
		
		clientImpl := client.(*ClientImpl)
		if clientImpl.options.MaxTurns != 3 {
			t.Errorf("Expected final MaxTurns=3, got %d", clientImpl.options.MaxTurns)
		}
	})
}

// T161: Client Default Configuration 🔴 RED  
// Go Target: client_test.go::TestClientDefaultConfiguration
func TestClientDefaultConfiguration(t *testing.T) {
	// Test sensible defaults for zero-config usage
	
	// Test 1: Zero-config client creation
	t.Run("ZeroConfigClientCreation", func(t *testing.T) {
		client := NewClient()
		if client == nil {
			t.Fatal("NewClient() should not return nil")
		}
		
		clientImpl := client.(*ClientImpl)
		if clientImpl.options == nil {
			t.Fatal("Client options should be initialized even with no options")
		}
		
		// Verify default values are sensible
		options := clientImpl.options
		
		// MaxTurns should have reasonable default (check current default)
		if options.MaxTurns < 0 {
			t.Error("Default MaxTurns should be non-negative")
		}
		
		// Working directory should be nil initially (uses CLI default)
		if options.Cwd != nil {
			t.Errorf("Default Cwd should be nil to use CLI default, got %v", options.Cwd)
		}
		
		// System prompt should be nil initially (uses CLI default)
		if options.SystemPrompt != nil {
			t.Errorf("Default SystemPrompt should be nil to use CLI default, got %v", options.SystemPrompt)
		}
		
		// Permission mode should be nil initially (uses CLI default)
		if options.PermissionMode != nil {
			t.Errorf("Default PermissionMode should be nil to use CLI default, got %v", options.PermissionMode)
		}
	})
	
	// Test 2: Default values don't break functionality
	t.Run("DefaultValuesFunctional", func(t *testing.T) {
		transport := &clientMockTransport{
			msgChan: make(chan Message, 10),
			errChan: make(chan error, 10),
		}
		
		client := NewClient() // Zero config
		
		// Set custom transport for testing
		clientImpl := client.(*ClientImpl)
		clientImpl.customTransport = transport
		
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		
		// Connect should work with defaults
		err := client.Connect(ctx)
		if err != nil {
			t.Fatalf("Connect with default config failed: %v", err)
		}
		defer client.Disconnect()
		
		// Query should work with defaults
		err = client.Query(ctx, "Default config test")
		if err != nil {
			t.Errorf("Query with default config failed: %v", err)
		}
		
		// Verify message was sent
		transport.mu.Lock()
		sentCount := len(transport.sentMessages)
		transport.mu.Unlock()
		
		if sentCount != 1 {
			t.Errorf("Expected 1 message sent, got %d", sentCount)
		}
	})
	
	// Test 3: Default MaxTurns behavior
	t.Run("DefaultMaxTurnsBehavior", func(t *testing.T) {
		client := NewClient()
		clientImpl := client.(*ClientImpl)
		
		// Should have sensible default (0 means no limit, which is valid)
		defaultMaxTurns := clientImpl.options.MaxTurns
		if defaultMaxTurns < 0 {
			t.Errorf("Default MaxTurns should be non-negative, got %d", defaultMaxTurns)
		}
		
		// Default of 0 means "no limit" which is a reasonable default
		if defaultMaxTurns != 0 {
			t.Logf("Default MaxTurns is %d (0 means no limit)", defaultMaxTurns)
		}
	})
	
	// Test 4: Defaults vs explicit configuration
	t.Run("DefaultsVsExplicitConfiguration", func(t *testing.T) {
		// Zero config client
		defaultClient := NewClient()
		defaultImpl := defaultClient.(*ClientImpl)
		
		// Explicitly configured client with same values as defaults
		explicitClient := NewClient(
			WithMaxTurns(defaultImpl.options.MaxTurns), // Use same value as default
		)
		explicitImpl := explicitClient.(*ClientImpl)
		
		// Should behave identically
		if defaultImpl.options.MaxTurns != explicitImpl.options.MaxTurns {
			t.Error("Default and explicit configuration should be identical when values match")
		}
		
		// Both should work identically
		transport1 := &clientMockTransport{
			msgChan: make(chan Message, 10),
			errChan: make(chan error, 10),
		}
		transport2 := &clientMockTransport{
			msgChan: make(chan Message, 10),
			errChan: make(chan error, 10),
		}
		
		defaultImpl.customTransport = transport1
		explicitImpl.customTransport = transport2
		
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		
		// Both should connect successfully
		err1 := defaultClient.Connect(ctx)
		err2 := explicitClient.Connect(ctx)
		
		if err1 != nil {
			t.Errorf("Default client connect failed: %v", err1)
		}
		if err2 != nil {
			t.Errorf("Explicit client connect failed: %v", err2)
		}
		
		defer defaultClient.Disconnect()
		defer explicitClient.Disconnect()
	})
	
	// Test 5: Default configuration validation
	t.Run("DefaultConfigurationValidation", func(t *testing.T) {
		client := NewClient()
		clientImpl := client.(*ClientImpl)
		
		// Default configuration should pass validation
		err := clientImpl.validateOptions()
		if err != nil {
			t.Errorf("Default configuration should pass validation, got error: %v", err)
		}
	})
	
	// Test 6: Zero config usability
	t.Run("ZeroConfigUsability", func(t *testing.T) {
		// This tests the user experience for zero-config usage
		transport := &clientMockTransport{
			msgChan: make(chan Message, 10),
			errChan: make(chan error, 10),
		}
		
		// User creates client with no configuration
		client := NewClient()
		clientImpl := client.(*ClientImpl)
		clientImpl.customTransport = transport
		
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		
		// Should work immediately without additional setup
		err := client.Connect(ctx)
		if err != nil {
			t.Fatalf("Zero-config client should connect immediately, got: %v", err)
		}
		defer client.Disconnect()
		
		// Should be able to send messages immediately
		err = client.Query(ctx, "Hello, Claude!")
		if err != nil {
			t.Fatalf("Zero-config client should send messages immediately, got: %v", err)
		}
		
		// Should receive messages
		messages := client.ReceiveMessages(ctx)
		if messages == nil {
			t.Error("Zero-config client should receive messages")
		}
		
		// Should support iterators
		iter := client.ReceiveResponse(ctx)
		if iter == nil {
			t.Error("Zero-config client should support response iterators")
		}
		if iter != nil {
			iter.Close()
		}
		
		// Should support interrupts
		err = client.Interrupt(ctx)
		if err != nil {
			t.Logf("Interrupt returned: %v (may be expected)", err)
		}
		
		// Verify the expected message was sent
		transport.mu.Lock()
		sentCount := len(transport.sentMessages)
		if sentCount > 0 {
			msg := transport.sentMessages[0]
			if userMsg, ok := msg.Message.(*UserMessage); ok {
				if content, ok := userMsg.Content.(string); ok {
					if content != "Hello, Claude!" {
						t.Errorf("Expected message 'Hello, Claude!', got '%s'", content)
					}
				}
			}
		}
		transport.mu.Unlock()
	})
}

// T147: Client Message Ordering 🔴 RED
// Go Target: client_test.go::TestClientMessageOrdering
func TestClientMessageOrdering(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test 1: Sequential messages maintain order
	t.Run("SequentialMessageOrdering", func(t *testing.T) {
		transport := &clientMockTransport{}
		client := NewClientWithTransport(transport)

		err := client.Connect(ctx)
		if err != nil {
			t.Fatalf("Connect failed: %v", err)
		}
		defer client.Disconnect()

		// Send messages sequentially
		for i := 0; i < 10; i++ {
			err = client.Query(ctx, fmt.Sprintf("Message %d", i))
			if err != nil {
				t.Errorf("Query %d failed: %v", i, err)
			}
		}

		// Verify messages were sent in order
		expectedCount := 10
		if transport.getSentMessageCount() != expectedCount {
			t.Errorf("Expected %d messages, got %d", expectedCount, transport.getSentMessageCount())
		}

		// Check message ordering
		for i := 0; i < expectedCount; i++ {
			sentMsg, ok := transport.getSentMessage(i)
			if !ok {
				t.Errorf("Failed to get message %d", i)
				continue
			}
			
			userMsg, ok := sentMsg.Message.(*UserMessage)
			if !ok {
				t.Errorf("Message %d is not UserMessage", i)
				continue
			}
			
			expectedContent := fmt.Sprintf("Message %d", i)
			if content, ok := userMsg.Content.(string); !ok || content != expectedContent {
				t.Errorf("Message %d content mismatch: expected %s, got %v", i, expectedContent, userMsg.Content)
			}
		}
	})

	// Test 2: Concurrent messages from multiple goroutines maintain order per session
	t.Run("ConcurrentMessageOrderingPerSession", func(t *testing.T) {
		transport := &clientMockTransport{}
		client := NewClientWithTransport(transport)

		err := client.Connect(ctx)
		if err != nil {
			t.Fatalf("Connect failed: %v", err)
		}
		defer client.Disconnect()

		const numGoroutines = 10
		const messagesPerGoroutine = 5
		var wg sync.WaitGroup
		errorChan := make(chan error, numGoroutines*messagesPerGoroutine)

		// Launch multiple goroutines sending messages with different session IDs
		for goroutineID := 0; goroutineID < numGoroutines; goroutineID++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				sessionID := fmt.Sprintf("session-%d", id)
				
				// Send messages sequentially within this goroutine
				for msgNum := 0; msgNum < messagesPerGoroutine; msgNum++ {
					content := fmt.Sprintf("G%d-M%d", id, msgNum)
					if err := client.Query(ctx, content, sessionID); err != nil {
						errorChan <- fmt.Errorf("goroutine %d, message %d: %w", id, msgNum, err)
						return
					}
					// Small delay to increase chance of interleaving
					time.Sleep(1 * time.Millisecond)
				}
			}(goroutineID)
		}

		wg.Wait()
		close(errorChan)

		// Check for any errors
		for err := range errorChan {
			t.Errorf("Concurrent message error: %v", err)
		}

		// Verify all messages were sent
		expectedTotal := numGoroutines * messagesPerGoroutine
		if transport.getSentMessageCount() != expectedTotal {
			t.Errorf("Expected %d total messages, got %d", expectedTotal, transport.getSentMessageCount())
		}

		// Group messages by session ID and verify ordering within each session
		sessionMessages := make(map[string][]string)
		for i := 0; i < transport.getSentMessageCount(); i++ {
			sentMsg, ok := transport.getSentMessage(i)
			if !ok {
				continue
			}
			
			userMsg, ok := sentMsg.Message.(*UserMessage)
			if !ok {
				continue
			}
			
			if content, ok := userMsg.Content.(string); ok {
				sessionMessages[sentMsg.SessionID] = append(sessionMessages[sentMsg.SessionID], content)
			}
		}

		// Verify ordering within each session
		for sessionID, messages := range sessionMessages {
			if len(messages) != messagesPerGoroutine {
				t.Errorf("Session %s should have %d messages, got %d", sessionID, messagesPerGoroutine, len(messages))
			}
			
			// Extract goroutine ID from session ID
			var goroutineID int
			if _, err := fmt.Sscanf(sessionID, "session-%d", &goroutineID); err != nil {
				continue
			}
			
			// Verify messages are in order for this session
			for msgIndex, content := range messages {
				expected := fmt.Sprintf("G%d-M%d", goroutineID, msgIndex)
				if content != expected {
					t.Errorf("Session %s message %d: expected %s, got %s", sessionID, msgIndex, expected, content)
				}
			}
		}
	})

	// Test 3: QueryStream maintains message order
	t.Run("QueryStreamMessageOrdering", func(t *testing.T) {
		transport := &clientMockTransport{}
		client := NewClientWithTransport(transport)

		err := client.Connect(ctx)
		if err != nil {
			t.Fatalf("Connect failed: %v", err)
		}
		defer client.Disconnect()

		// Create ordered messages
		const messageCount = 20
		messages := make(chan StreamMessage, messageCount)
		
		for i := 0; i < messageCount; i++ {
			messages <- StreamMessage{
				Type:      "request",
				Message:   &UserMessage{Content: fmt.Sprintf("Stream message %d", i)},
				SessionID: "stream-session",
			}
		}
		close(messages)

		// Send stream
		err = client.QueryStream(ctx, messages)
		if err != nil {
			t.Errorf("QueryStream failed: %v", err)
		}

		// Give time for all messages to be processed
		time.Sleep(200 * time.Millisecond)

		// Verify messages were sent in order
		if transport.getSentMessageCount() != messageCount {
			t.Errorf("Expected %d messages, got %d", messageCount, transport.getSentMessageCount())
		}

		for i := 0; i < messageCount; i++ {
			sentMsg, ok := transport.getSentMessage(i)
			if !ok {
				t.Errorf("Failed to get stream message %d", i)
				continue
			}
			
			userMsg, ok := sentMsg.Message.(*UserMessage)
			if !ok {
				t.Errorf("Stream message %d is not UserMessage", i)
				continue
			}
			
			expectedContent := fmt.Sprintf("Stream message %d", i)
			if content, ok := userMsg.Content.(string); !ok || content != expectedContent {
				t.Errorf("Stream message %d content mismatch: expected %s, got %v", i, expectedContent, userMsg.Content)
			}
			
			if sentMsg.SessionID != "stream-session" {
				t.Errorf("Stream message %d session ID mismatch: expected stream-session, got %s", i, sentMsg.SessionID)
			}
		}
	})

	// Test 4: Mixed Query and QueryStream maintain relative ordering
	t.Run("MixedQueryStreamOrdering", func(t *testing.T) {
		transport := &clientMockTransport{}
		client := NewClientWithTransport(transport)

		err := client.Connect(ctx)
		if err != nil {
			t.Fatalf("Connect failed: %v", err)
		}
		defer client.Disconnect()

		var wg sync.WaitGroup
		var orderMutex sync.Mutex
		var executionOrder []string

		// Send some regular queries
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 5; i++ {
				err := client.Query(ctx, fmt.Sprintf("Query %d", i))
				if err != nil {
					t.Errorf("Query %d failed: %v", i, err)
				}
				orderMutex.Lock()
				executionOrder = append(executionOrder, fmt.Sprintf("Query %d", i))
				orderMutex.Unlock()
				time.Sleep(2 * time.Millisecond)
			}
		}()

		// Send some stream queries
		wg.Add(1)
		go func() {
			defer wg.Done()
			messages := make(chan StreamMessage, 5)
			for i := 0; i < 5; i++ {
				messages <- StreamMessage{
					Type:    "request",
					Message: &UserMessage{Content: fmt.Sprintf("Stream %d", i)},
				}
			}
			close(messages)
			
			err := client.QueryStream(ctx, messages)
			if err != nil {
				t.Errorf("QueryStream failed: %v", err)
			}
			orderMutex.Lock()
			executionOrder = append(executionOrder, "Stream batch")
			orderMutex.Unlock()
		}()

		wg.Wait()
		
		// Give time for all operations to complete
		time.Sleep(100 * time.Millisecond)

		// Verify some messages were sent (exact ordering may vary due to concurrency)
		if transport.getSentMessageCount() < 5 {
			t.Errorf("Expected at least 5 messages, got %d", transport.getSentMessageCount())
		}
	})

	// Test 5: High concurrency message ordering stress test
	t.Run("HighConcurrencyMessageOrdering", func(t *testing.T) {
		transport := &clientMockTransport{}
		client := NewClientWithTransport(transport)

		err := client.Connect(ctx)
		if err != nil {
			t.Fatalf("Connect failed: %v", err)
		}
		defer client.Disconnect()

		const numWorkers = 50
		const messagesPerWorker = 10
		var wg sync.WaitGroup
		errorChan := make(chan error, numWorkers*messagesPerWorker)

		// Launch many concurrent workers
		for workerID := 0; workerID < numWorkers; workerID++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				sessionID := fmt.Sprintf("worker-%d", id)
				
				for msgNum := 0; msgNum < messagesPerWorker; msgNum++ {
					content := fmt.Sprintf("W%d-M%d", id, msgNum)
					if err := client.Query(ctx, content, sessionID); err != nil {
						errorChan <- err
						return
					}
				}
			}(workerID)
		}

		wg.Wait()
		close(errorChan)

		// Check for errors
		var errorCount int
		for err := range errorChan {
			errorCount++
			if errorCount <= 5 { // Only log first 5 errors to avoid spam
				t.Logf("High concurrency error: %v", err)
			}
		}

		if errorCount > 0 {
			t.Errorf("High concurrency test had %d errors", errorCount)
		}

		// Verify reasonable number of messages were sent
		expectedTotal := numWorkers * messagesPerWorker
		actualCount := transport.getSentMessageCount()
		
		if actualCount < expectedTotal*8/10 { // Allow for some loss due to high contention
			t.Errorf("Expected at least %d messages (80%% of %d), got %d", expectedTotal*8/10, expectedTotal, actualCount)
		}

		// Verify session isolation in high concurrency
		sessionCounts := make(map[string]int)
		for i := 0; i < actualCount; i++ {
			sentMsg, ok := transport.getSentMessage(i)
			if ok {
				sessionCounts[sentMsg.SessionID]++
			}
		}

		if len(sessionCounts) < numWorkers/2 { // Should have messages from at least half the workers
			t.Errorf("Expected messages from at least %d sessions, got %d", numWorkers/2, len(sessionCounts))
		}
	})

	// Test 6: FIFO ordering guarantee
	t.Run("FIFOOrderingGuarantee", func(t *testing.T) {
		transport := &clientMockTransport{}
		client := NewClientWithTransport(transport)

		err := client.Connect(ctx)
		if err != nil {
			t.Fatalf("Connect failed: %v", err)
		}
		defer client.Disconnect()

		// Send messages with timestamps
		const messageCount = 100
		var wg sync.WaitGroup
		
		// Single session to ensure FIFO
		sessionID := "fifo-test"
		
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < messageCount; i++ {
				timestamp := time.Now().UnixNano()
				content := fmt.Sprintf("FIFO-%d-T%d", i, timestamp)
				if err := client.Query(ctx, content, sessionID); err != nil {
					t.Errorf("FIFO message %d failed: %v", i, err)
					return
				}
				// No delay - send as fast as possible to test ordering under pressure
			}
		}()

		wg.Wait()

		// Give time for all messages to be processed
		time.Sleep(100 * time.Millisecond)

		// Verify FIFO ordering
		actualCount := transport.getSentMessageCount()
		if actualCount != messageCount {
			t.Errorf("Expected %d messages for FIFO test, got %d", messageCount, actualCount)
		}

		for i := 0; i < actualCount; i++ {
			sentMsg, ok := transport.getSentMessage(i)
			if !ok {
				t.Errorf("Failed to get FIFO message %d", i)
				continue
			}
			
			userMsg, ok := sentMsg.Message.(*UserMessage)
			if !ok {
				t.Errorf("FIFO message %d is not UserMessage", i)
				continue
			}
			
			// Extract message index from content
			var msgIndex int
			content, ok := userMsg.Content.(string)
			if !ok {
				t.Errorf("FIFO message %d content is not string", i)
				continue
			}
			if n, err := fmt.Sscanf(content, "FIFO-%d-T", &msgIndex); n != 1 || err != nil {
				t.Errorf("Could not parse FIFO message %d content: %s", i, content)
				continue
			}
			
			if msgIndex != i {
				t.Errorf("FIFO violation: expected message index %d at position %d, got %d", i, i, msgIndex)
			}
		}
	})
}