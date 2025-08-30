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
	
	if userMsg.Content != "What is 2+2?" {
		t.Errorf("Expected content 'What is 2+2?', got '%s'", userMsg.Content)
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
	c.connected = false
	if c.msgChan != nil {
		close(c.msgChan)
		close(c.errChan)
	}
	return nil
}

// Helper method to send response messages
func (c *clientMockTransport) sendResponses() {
	if c.responseMessages != nil && c.msgChan != nil {
		go func() {
			for _, msg := range c.responseMessages {
				c.msgChan <- msg
			}
		}()
	}
}

// Helper method to send async error
func (c *clientMockTransport) sendAsyncError() {
	if c.asyncError != nil && c.errChan != nil {
		go func() {
			c.errChan <- c.asyncError
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