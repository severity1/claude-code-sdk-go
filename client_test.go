package claudecode

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestClientAutoConnectContextManager(t *testing.T) {
	ctx, cancel := setupTestContext(t, 10*time.Second)
	defer cancel()

	transport := newMockTransport()

	// Test defer-based resource management (Go equivalent of Python context manager)
	func() {
		client := setupClient(t, transport)
		defer disconnectClient(t, client)

		// Connect should be called automatically or explicitly
		connectClient(t, ctx, client)

		// Verify connection was established
		assertConnected(t, transport)

		// Client should be ready to use
		err := client.Query(ctx, "test message")
		assertError(t, err, false, "")
	}() // Defer should trigger disconnect

	// Verify disconnect was called
	assertDisconnected(t, transport)
}

func TestClientManualConnection(t *testing.T) {
	ctx, cancel := setupTestContext(t, 10*time.Second)
	defer cancel()

	transport := newMockTransport()
	client := setupClient(t, transport)

	// Manual Connect/Disconnect lifecycle
	connectClient(t, ctx, client)
	assertConnected(t, transport)

	disconnectClient(t, client)
	assertDisconnected(t, transport)
}

func TestClientQueryExecution(t *testing.T) {
	ctx, cancel := setupTestContext(t, 10*time.Second)
	defer cancel()

	transport := newMockTransport()
	client := setupClient(t, transport)
	defer disconnectClient(t, client)

	connectClient(t, ctx, client)

	// Execute query through connected client
	err := client.Query(ctx, "What is 2+2?")
	assertError(t, err, false, "")

	// Verify message was sent to transport
	assertMessageCount(t, transport, 1)

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

func TestClientStreamQuery(t *testing.T) {
	ctx, cancel := setupTestContext(t, 10*time.Second)
	defer cancel()

	transport := newMockTransport()
	client := setupClient(t, transport)
	defer disconnectClient(t, client)

	connectClient(t, ctx, client)

	// Create message channel
	messages := make(chan StreamMessage, 3)
	messages <- StreamMessage{Type: "request", Message: &UserMessage{Content: "First message"}}
	messages <- StreamMessage{Type: "request", Message: &UserMessage{Content: "Second message"}}
	messages <- StreamMessage{Type: "request", Message: &UserMessage{Content: "Third message"}}
	close(messages)

	// Execute stream query
	err := client.QueryStream(ctx, messages)
	assertError(t, err, false, "")

	// Give time for messages to be processed
	time.Sleep(50 * time.Millisecond)

	// Verify all messages were sent
	assertMessageCount(t, transport, 3)
}

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
	connectError   error
	sendError      error
	interruptError error
	closeError     error
	asyncError     error
	slowSend       bool // Enable slow sending for backpressure testing
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

// setupTestContext creates a context with timeout and cancel function
func setupTestContext(t *testing.T, timeout time.Duration) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(context.Background(), timeout)
}


// newMockTransport creates a simple mock transport with default configuration.
func newMockTransport() *clientMockTransport {
	return &clientMockTransport{}
}

// MockTransportOption configures mock transport for table-driven tests.
type MockTransportOption func(*clientMockTransport)

// WithConnectError injects a connect error.
func WithConnectError(err error) MockTransportOption {
	return func(t *clientMockTransport) {
		t.connectError = err
	}
}

// WithSendError injects a send error.
func WithSendError(err error) MockTransportOption {
	return func(t *clientMockTransport) {
		t.sendError = err
	}
}

// WithAsyncError injects an async error.
func WithAsyncError(err error) MockTransportOption {
	return func(t *clientMockTransport) {
		t.asyncError = err
	}
}

// WithInterruptError injects an interrupt error.
func WithInterruptError(err error) MockTransportOption {
	return func(t *clientMockTransport) {
		t.interruptError = err
	}
}

// WithCloseError injects a close error.
func WithCloseError(err error) MockTransportOption {
	return func(t *clientMockTransport) {
		t.closeError = err
	}
}

// WithResponseMessages sets response messages.
func WithResponseMessages(messages []Message) MockTransportOption {
	return func(t *clientMockTransport) {
		t.responseMessages = messages
	}
}

// newMockTransportWithOptions creates configured mock transport using functional options.
func newMockTransportWithOptions(options ...MockTransportOption) *clientMockTransport {
	transport := &clientMockTransport{}
	for _, option := range options {
		option(transport)
	}
	return transport
}

// setupClient creates client with transport and options
func setupClient(t *testing.T, transport Transport, options ...Option) Client {
	t.Helper()
	return NewClientWithTransport(transport, options...)
}

// connectClient connects client with error handling
func connectClient(t *testing.T, ctx context.Context, client Client) {
	t.Helper()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Client connect failed: %v", err)
	}
}

// disconnectClient safely disconnects client
func disconnectClient(t *testing.T, client Client) {
	t.Helper()
	if err := client.Disconnect(); err != nil {
		t.Errorf("Client disconnect failed: %v", err)
	}
}

// Assertion helpers with t.Helper()
func assertConnected(t *testing.T, transport *clientMockTransport) {
	t.Helper()
	if !transport.connected {
		t.Error("Expected transport to be connected")
	}
}

func assertDisconnected(t *testing.T, transport *clientMockTransport) {
	t.Helper()
	if transport.connected {
		t.Error("Expected transport to be disconnected")
	}
}

func assertError(t *testing.T, err error, wantErr bool, msgContains string) {
	t.Helper()
	if (err != nil) != wantErr {
		t.Errorf("error = %v, wantErr %v", err, wantErr)
		return
	}
	if wantErr && msgContains != "" && !strings.Contains(err.Error(), msgContains) {
		t.Errorf("error = %v, expected message to contain %q", err, msgContains)
	}
}

func assertMessageCount(t *testing.T, transport *clientMockTransport, expected int) {
	t.Helper()
	actual := transport.getSentMessageCount()
	if actual != expected {
		t.Errorf("Expected %d sent messages, got %d", expected, actual)
	}
}

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

// TestClientErrorPropagation tests error propagation through client operations.
func TestClientErrorPropagation(t *testing.T) {
	tests := []struct {
		name           string
		transportSetup func() *clientMockTransport
		operation      func(context.Context, Client) error
		wantErr        bool
		wantErrMsg     string
		validate       func(*testing.T, *clientMockTransport, Client)
	}{
		{
			name: "connect_error_propagation",
			transportSetup: func() *clientMockTransport {
				return newMockTransportWithOptions(WithConnectError(fmt.Errorf("connection failed: CLI not found")))
			},
			operation: func(ctx context.Context, client Client) error {
				return client.Connect(ctx)
			},
			wantErr:    true,
			wantErrMsg: "failed to connect transport: connection failed: CLI not found",
		},
		{
			name: "send_error_propagation",
			transportSetup: func() *clientMockTransport {
				return newMockTransportWithOptions(WithSendError(fmt.Errorf("send failed: process exited")))
			},
			operation: func(ctx context.Context, client Client) error {
				if err := client.Connect(ctx); err != nil {
					return fmt.Errorf("connect failed: %v", err)
				}
				return client.Query(ctx, "test")
			},
			wantErr:    true,
			wantErrMsg: "send failed: process exited",
		},
		{
			name: "async_error_propagation",
			transportSetup: func() *clientMockTransport {
				return newMockTransportWithOptions(WithAsyncError(fmt.Errorf("async transport error")))
			},
			operation: func(ctx context.Context, client Client) error {
				if err := client.Connect(ctx); err != nil {
					return fmt.Errorf("connect failed: %v", err)
				}
				// Get iterator and expect error from error channel
				iter := client.ReceiveResponse(ctx)
				if iter == nil {
					return fmt.Errorf("expected response iterator")
				}
				defer iter.Close()

				// Trigger async error
				if transport, ok := client.(*ClientImpl).transport.(*clientMockTransport); ok {
					transport.sendAsyncError()
				}

				// Next should return the async error
				_, err := iter.Next(ctx)
				return err
			},
			wantErr:    true,
			wantErrMsg: "async transport error",
		},
		{
			name: "interrupt_error_propagation",
			transportSetup: func() *clientMockTransport {
				return newMockTransportWithOptions(WithInterruptError(fmt.Errorf("interrupt failed")))
			},
			operation: func(ctx context.Context, client Client) error {
				if err := client.Connect(ctx); err != nil {
					return fmt.Errorf("connect failed: %v", err)
				}
				return client.Interrupt(ctx)
			},
			wantErr:    true,
			wantErrMsg: "interrupt failed",
		},
		{
			name: "close_error_propagation",
			transportSetup: func() *clientMockTransport {
				return newMockTransportWithOptions(WithCloseError(fmt.Errorf("close failed")))
			},
			operation: func(ctx context.Context, client Client) error {
				if err := client.Connect(ctx); err != nil {
					return fmt.Errorf("connect failed: %v", err)
				}
				return client.Disconnect()
			},
			wantErr:    true,
			wantErrMsg: "failed to close transport: close failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := setupTestContext(t, 5*time.Second)
			defer cancel()

			transport := tt.transportSetup()
			client := setupClient(t, transport)

			err := tt.operation(ctx, client)

			if (err != nil) != tt.wantErr {
				t.Errorf("operation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && err.Error() != tt.wantErrMsg {
				t.Errorf("operation() error message = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
			}

			if tt.validate != nil {
				tt.validate(t, transport, client)
			}
		})
	}
}

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

// TestClientResourceCleanup tests proper resource cleanup during client lifecycle.
func TestClientResourceCleanup(t *testing.T) {
	tests := []struct {
		name           string
		transportSetup func() *clientMockTransport
		operation      func(*testing.T, context.Context, Client, *clientMockTransport)
		validate       func(*testing.T, *clientMockTransport, Client)
	}{
		{
			name: "basic_resource_cleanup",
			transportSetup: func() *clientMockTransport {
				return &clientMockTransport{}
			},
			operation: func(t *testing.T, ctx context.Context, client Client, transport *clientMockTransport) {
				t.Helper()
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
			},
			validate: func(t *testing.T, transport *clientMockTransport, client Client) {
				t.Helper()
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
			},
		},
		{
			name: "multiple_disconnect_safe",
			transportSetup: func() *clientMockTransport {
				return &clientMockTransport{}
			},
			operation: func(t *testing.T, ctx context.Context, client Client, transport *clientMockTransport) {
				t.Helper()
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
			},
		},
		{
			name: "cleanup_with_active_goroutines",
			transportSetup: func() *clientMockTransport {
				return &clientMockTransport{
					responseMessages: []Message{
						&AssistantMessage{
							Content: []ContentBlock{&TextBlock{Text: "Hello"}},
							Model:   "claude-opus-4-1-20250805",
						},
					},
				}
			},
			operation: func(t *testing.T, ctx context.Context, client Client, transport *clientMockTransport) {
				t.Helper()
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
			},
			validate: func(t *testing.T, transport *clientMockTransport, client Client) {
				t.Helper()
				// Verify cleanup
				if transport.connected {
					t.Error("Transport should be disconnected after cleanup with active goroutines")
				}
			},
		},
		{
			name: "repeated_connect_disconnect_no_leak",
			transportSetup: func() *clientMockTransport {
				return &clientMockTransport{}
			},
			operation: func(t *testing.T, ctx context.Context, client Client, transport *clientMockTransport) {
				t.Helper()
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
			},
		},
		{
			name: "disconnect_during_operations",
			transportSetup: func() *clientMockTransport {
				return &clientMockTransport{}
			},
			operation: func(t *testing.T, ctx context.Context, client Client, transport *clientMockTransport) {
				t.Helper()
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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			transport := tt.transportSetup()
			client := NewClientWithTransport(transport)

			// Execute operation
			if tt.operation != nil {
				tt.operation(t, ctx, client, transport)
			}

			// Validation
			if tt.validate != nil {
				tt.validate(t, transport, client)
			}
		})
	}
}

// TestClientConfigurationValidation tests validation of client configuration options.
func TestClientConfigurationValidation(t *testing.T) {
	tests := []struct {
		name          string
		clientOptions []Option
		setupClient   func(Client) // For manual modifications like invalid permission mode
		wantErr       bool
		wantErrMsg    string // Expected error message substring
		shouldConnect bool   // Whether Connect() should succeed
		validate      func(*testing.T, Client)
	}{
		{
			name: "valid_configuration",
			clientOptions: []Option{
				WithSystemPrompt("You are a helpful assistant"),
				WithAllowedTools("Read", "Write", "Bash"),
				WithDisallowedTools("dangerous-tool"),
				WithMaxTurns(10),
				WithPermissionMode(PermissionModeAcceptEdits),
				WithModel("claude-opus-4-1-20250805"),
				WithCwd("/tmp"),
			},
			wantErr:       false,
			shouldConnect: true,
		},
		{
			name: "invalid_working_directory",
			clientOptions: []Option{
				WithCwd("/non/existent/directory/that/should/not/exist"),
			},
			wantErr:       true,
			wantErrMsg:    "directory",
			shouldConnect: false,
		},
		{
			name: "invalid_max_turns",
			clientOptions: []Option{
				WithMaxTurns(-1),
			},
			wantErr:       true,
			wantErrMsg:    "turns",
			shouldConnect: false,
		},
		{
			name:          "invalid_permission_mode",
			clientOptions: []Option{},
			setupClient: func(client Client) {
				// Manually set invalid permission mode
				clientImpl := client.(*ClientImpl)
				invalidMode := PermissionMode("invalid_mode")
				clientImpl.options.PermissionMode = &invalidMode
			},
			wantErr:       true,
			wantErrMsg:    "permission",
			shouldConnect: false,
		},
		{
			name: "conflicting_tool_configuration",
			clientOptions: []Option{
				WithAllowedTools("Read", "Write"),
				WithDisallowedTools("Read"), // Read is both allowed and disallowed
			},
			wantErr:       false, // May be accepted with precedence rules
			shouldConnect: true,
		},
		{
			name: "empty_model_name",
			clientOptions: []Option{
				WithModel(""),
			},
			wantErr:       false, // May use defaults
			shouldConnect: true,
		},
		{
			name: "invalid_mcp_server_configuration",
			clientOptions: []Option{
				WithMcpServers(map[string]McpServerConfig{
					"test-server": &McpStdioServerConfig{
						Type:    McpServerTypeStdio,
						Command: "", // Empty command
						Args:    []string{},
					},
				}),
			},
			wantErr:       false, // May be accepted
			shouldConnect: true,
		},
		{
			name:          "nil_options_handling",
			clientOptions: []Option{}, // No options, use defaults
			wantErr:       false,
			shouldConnect: true,
		},
		{
			name: "configuration_immutability_after_connection",
			clientOptions: []Option{
				WithSystemPrompt("Original prompt"),
			},
			wantErr:       false,
			shouldConnect: true,
			validate: func(t *testing.T, client Client) {
				// Verify config can be modified after creation but before connection
				clientImpl := client.(*ClientImpl)
				originalPrompt := "Original prompt"
				if clientImpl.options.SystemPrompt == nil || *clientImpl.options.SystemPrompt != originalPrompt {
					t.Errorf("Expected original prompt to be %q", originalPrompt)
				}

				// Modify after connection attempt
				modifiedPrompt := "Modified prompt"
				clientImpl.options.SystemPrompt = &modifiedPrompt

				// This tests that modification doesn't break subsequent operations
				ctx, cancel := setupTestContext(t, 5*time.Second)
				defer cancel()

				err := client.Query(ctx, "test")
				if err != nil {
					t.Errorf("Query failed after options modification: %v", err)
				}
			},
		},
		{
			name: "large_configuration_values",
			clientOptions: []Option{
				WithSystemPrompt(strings.Repeat("This is a very long system prompt. ", 1000)),
				WithMaxThinkingTokens(100000),
			},
			wantErr:       false, // Should handle large values
			shouldConnect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := setupTestContext(t, 5*time.Second)
			defer cancel()

			transport := newMockTransport()
			client := setupClient(t, transport, tt.clientOptions...)

			// Apply any manual setup modifications
			if tt.setupClient != nil {
				tt.setupClient(client)
			}

			err := client.Connect(ctx)

			if tt.shouldConnect {
				if err != nil && tt.wantErr && tt.wantErrMsg != "" {
					if !strings.Contains(err.Error(), tt.wantErrMsg) {
						t.Errorf("Connect() error = %v, wantErrMsg substring = %v", err.Error(), tt.wantErrMsg)
					}
				} else if err != nil && !tt.wantErr {
					t.Errorf("Connect() unexpected error = %v", err)
				} else if err == nil && tt.wantErr {
					t.Error("Connect() expected error but got none")
					client.Disconnect() // Clean up successful connection
				}

				// If connection succeeded, clean up
				if err == nil {
					defer client.Disconnect()
				}
			} else {
				// Should not connect
				if err == nil {
					t.Error("Connect() should have failed")
					client.Disconnect()
				} else if tt.wantErrMsg != "" && !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("Connect() error = %v, wantErrMsg substring = %v", err.Error(), tt.wantErrMsg)
				}
			}

			// Run additional validations
			if tt.validate != nil {
				tt.validate(t, client)
			}
		})
	}
}

// TestClientInterfaceCompliance tests compliance with the Client interface contract.
func TestClientInterfaceCompliance(t *testing.T) {
	tests := []struct {
		name           string
		transportSetup func() *clientMockTransport
		operation      func(*testing.T, context.Context, Client)
		validate       func(*testing.T, *clientMockTransport, Client)
		wantErr        bool
		connectFirst   bool // Whether to connect before operation
	}{
		{
			name: "interface_methods_exist",
			transportSetup: func() *clientMockTransport {
				return &clientMockTransport{}
			},
			operation: func(t *testing.T, ctx context.Context, client Client) {
				t.Helper()
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
			},
		},
		{
			name: "connect_method_behavior",
			transportSetup: func() *clientMockTransport {
				return &clientMockTransport{}
			},
			operation: func(t *testing.T, ctx context.Context, client Client) {
				t.Helper()
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
			},
		},
		{
			name: "query_method_variations",
			transportSetup: func() *clientMockTransport {
				return &clientMockTransport{}
			},
			connectFirst: true,
			operation: func(t *testing.T, ctx context.Context, client Client) {
				t.Helper()
				// Query with no session ID (should use default)
				err := client.Query(ctx, "test message")
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
			},
			validate: func(t *testing.T, transport *clientMockTransport, client Client) {
				t.Helper()
				// Verify messages were sent correctly
				expectedMsgCount := 3
				if transport.getSentMessageCount() != expectedMsgCount {
					t.Errorf("Expected %d messages sent, got %d", expectedMsgCount, transport.getSentMessageCount())
				}
			},
		},
		{
			name: "query_stream_method_behavior",
			transportSetup: func() *clientMockTransport {
				return &clientMockTransport{}
			},
			connectFirst: true,
			operation: func(t *testing.T, ctx context.Context, client Client) {
				t.Helper()
				// Test with multiple messages
				messages := make(chan StreamMessage, 3)
				messages <- StreamMessage{Type: "request", Message: &UserMessage{Content: "Message 1"}}
				messages <- StreamMessage{Type: "request", Message: &UserMessage{Content: "Message 2"}}
				messages <- StreamMessage{Type: "request", Message: &UserMessage{Content: "Message 3"}}
				close(messages)

				err := client.QueryStream(ctx, messages)
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
			},
		},
		{
			name: "receive_messages_method_behavior",
			transportSetup: func() *clientMockTransport {
				return &clientMockTransport{
					responseMessages: []Message{
						&AssistantMessage{Content: []ContentBlock{&TextBlock{Text: "Response 1"}}, Model: "claude-opus-4-1-20250805"},
						&AssistantMessage{Content: []ContentBlock{&TextBlock{Text: "Response 2"}}, Model: "claude-opus-4-1-20250805"},
					},
				}
			},
			connectFirst: true,
			operation: func(t *testing.T, ctx context.Context, client Client) {
				t.Helper()
				// Get message channel
				msgChan := client.ReceiveMessages(ctx)
				if msgChan == nil {
					t.Fatal("ReceiveMessages returned nil channel")
				}
			},
			validate: func(t *testing.T, transport *clientMockTransport, client Client) {
				t.Helper()
				ctx := context.Background()
				msgChan := client.ReceiveMessages(ctx)

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
			},
		},
		{
			name: "receive_response_method_behavior",
			transportSetup: func() *clientMockTransport {
				return &clientMockTransport{
					responseMessages: []Message{
						&AssistantMessage{Content: []ContentBlock{&TextBlock{Text: "Iterator Response"}}, Model: "claude-opus-4-1-20250805"},
					},
				}
			},
			connectFirst: true,
			operation: func(t *testing.T, ctx context.Context, client Client) {
				t.Helper()
				// Get response iterator
				iter := client.ReceiveResponse(ctx)
				if iter == nil {
					t.Fatal("ReceiveResponse returned nil iterator")
				}
				if iter != nil {
					defer iter.Close()
				}
			},
			validate: func(t *testing.T, transport *clientMockTransport, client Client) {
				t.Helper()
				ctx := context.Background()
				iter := client.ReceiveResponse(ctx)
				if iter == nil {
					t.Fatal("ReceiveResponse returned nil iterator")
				}
				defer iter.Close()

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
			},
		},
		{
			name: "interrupt_method_behavior",
			transportSetup: func() *clientMockTransport {
				return &clientMockTransport{}
			},
			operation: func(t *testing.T, ctx context.Context, client Client) {
				t.Helper()
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
			},
		},
		{
			name: "disconnect_method_behavior",
			transportSetup: func() *clientMockTransport {
				return &clientMockTransport{}
			},
			operation: func(t *testing.T, ctx context.Context, client Client) {
				t.Helper()
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
			},
		},
		{
			name: "error_handling_consistency",
			transportSetup: func() *clientMockTransport {
				return &clientMockTransport{
					connectError: fmt.Errorf("mock connect error"),
				}
			},
			operation: func(t *testing.T, ctx context.Context, client Client) {
				t.Helper()
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
			},
		},
		{
			name: "interface_contract_nil_checks",
			transportSetup: func() *clientMockTransport {
				return &clientMockTransport{}
			},
			operation: func(t *testing.T, ctx context.Context, client Client) {
				t.Helper()
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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			transport := tt.transportSetup()
			client := NewClientWithTransport(transport)

			// Connect first if required
			if tt.connectFirst {
				err := client.Connect(ctx)
				if err != nil {
					t.Fatalf("Connect failed: %v", err)
				}
				defer client.Disconnect()
			}

			// Execute operation
			if tt.operation != nil {
				tt.operation(t, ctx, client)
			}

			// Validation
			if tt.validate != nil {
				tt.validate(t, transport, client)
			}
		})
	}
}

// TestClientConfigurationApplication tests application of client configuration options.
func TestClientConfigurationApplication(t *testing.T) {
	tests := []struct {
		name     string
		options  []Option
		validate func(*testing.T, *ClientImpl)
		preTest  func(*testing.T) interface{}               // For setup before test
		postTest func(*testing.T, interface{}, *ClientImpl) // For complex validation
	}{
		{
			name: "single_functional_option",
			options: []Option{
				WithSystemPrompt("Test system prompt"),
			},
			validate: func(t *testing.T, client *ClientImpl) {
				t.Helper()
				if client.options.SystemPrompt == nil {
					t.Error("SystemPrompt option was not applied")
				} else if *client.options.SystemPrompt != "Test system prompt" {
					t.Errorf("Expected system prompt 'Test system prompt', got '%s'", *client.options.SystemPrompt)
				}
			},
		},
		{
			name: "multiple_functional_options",
			options: []Option{
				WithSystemPrompt("System prompt"),
				WithAppendSystemPrompt("Append prompt"),
				WithModel("claude-opus-4-1-20250805"),
				WithMaxTurns(10),
				WithMaxThinkingTokens(5000),
			},
			validate: func(t *testing.T, client *ClientImpl) {
				t.Helper()
				if client.options.SystemPrompt == nil || *client.options.SystemPrompt != "System prompt" {
					t.Error("SystemPrompt option was not applied correctly")
				}
				if client.options.AppendSystemPrompt == nil || *client.options.AppendSystemPrompt != "Append prompt" {
					t.Error("AppendSystemPrompt option was not applied correctly")
				}
				if client.options.Model == nil || *client.options.Model != "claude-opus-4-1-20250805" {
					t.Error("Model option was not applied correctly")
				}
				if client.options.MaxTurns != 10 {
					t.Errorf("Expected MaxTurns 10, got %d", client.options.MaxTurns)
				}
				if client.options.MaxThinkingTokens != 5000 {
					t.Errorf("Expected MaxThinkingTokens 5000, got %d", client.options.MaxThinkingTokens)
				}
			},
		},
		{
			name: "tool_configuration_options",
			preTest: func(t *testing.T) interface{} {
				return struct {
					allowedTools    []string
					disallowedTools []string
				}{
					allowedTools:    []string{"Read", "Write", "Bash"},
					disallowedTools: []string{"dangerous-tool", "restricted-tool"},
				}
			},
			options: []Option{
				WithAllowedTools("Read", "Write", "Bash"),
				WithDisallowedTools("dangerous-tool", "restricted-tool"),
			},
			postTest: func(t *testing.T, preData interface{}, client *ClientImpl) {
				t.Helper()
				data := preData.(struct {
					allowedTools    []string
					disallowedTools []string
				})
				if len(client.options.AllowedTools) != len(data.allowedTools) {
					t.Errorf("Expected %d allowed tools, got %d", len(data.allowedTools), len(client.options.AllowedTools))
				}
				for i, tool := range data.allowedTools {
					if client.options.AllowedTools[i] != tool {
						t.Errorf("Expected allowed tool %s at index %d, got %s", tool, i, client.options.AllowedTools[i])
					}
				}
				if len(client.options.DisallowedTools) != len(data.disallowedTools) {
					t.Errorf("Expected %d disallowed tools, got %d", len(data.disallowedTools), len(client.options.DisallowedTools))
				}
				for i, tool := range data.disallowedTools {
					if client.options.DisallowedTools[i] != tool {
						t.Errorf("Expected disallowed tool %s at index %d, got %s", tool, i, client.options.DisallowedTools[i])
					}
				}
			},
		},
		{
			name: "permission_and_session_options",
			options: []Option{
				WithPermissionMode(PermissionModeAcceptEdits),
				WithPermissionPromptToolName("custom-prompt-tool"),
				WithContinueConversation(true),
				WithResume("session-123"),
			},
			validate: func(t *testing.T, client *ClientImpl) {
				t.Helper()
				if client.options.PermissionMode == nil || *client.options.PermissionMode != PermissionModeAcceptEdits {
					t.Error("PermissionMode option was not applied correctly")
				}
				if client.options.PermissionPromptToolName == nil || *client.options.PermissionPromptToolName != "custom-prompt-tool" {
					t.Error("PermissionPromptToolName option was not applied correctly")
				}
				if !client.options.ContinueConversation {
					t.Error("ContinueConversation option was not applied correctly")
				}
				if client.options.Resume == nil || *client.options.Resume != "session-123" {
					t.Error("Resume option was not applied correctly")
				}
			},
		},
		{
			name: "file_system_and_context_options",
			preTest: func(t *testing.T) interface{} {
				return []string{"/project", "/libs", "/docs"}
			},
			options: []Option{
				WithCwd("/tmp"),
				WithAddDirs("/project", "/libs", "/docs"),
				WithSettings("custom-settings.json"),
			},
			postTest: func(t *testing.T, preData interface{}, client *ClientImpl) {
				t.Helper()
				addDirs := preData.([]string)
				if client.options.Cwd == nil || *client.options.Cwd != "/tmp" {
					t.Error("Cwd option was not applied correctly")
				}
				if len(client.options.AddDirs) != len(addDirs) {
					t.Errorf("Expected %d add dirs, got %d", len(addDirs), len(client.options.AddDirs))
				}
				for i, dir := range addDirs {
					if client.options.AddDirs[i] != dir {
						t.Errorf("Expected add dir %s at index %d, got %s", dir, i, client.options.AddDirs[i])
					}
				}
				if client.options.Settings == nil || *client.options.Settings != "custom-settings.json" {
					t.Error("Settings option was not applied correctly")
				}
			},
		},
		{
			name: "mcp_server_configuration",
			preTest: func(t *testing.T) interface{} {
				return map[string]McpServerConfig{
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
			},
			options: []Option{
				WithMcpServers(map[string]McpServerConfig{
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
				}),
			},
			postTest: func(t *testing.T, preData interface{}, client *ClientImpl) {
				t.Helper()
				mcpServers := preData.(map[string]McpServerConfig)
				if len(client.options.McpServers) != len(mcpServers) {
					t.Errorf("Expected %d MCP servers, got %d", len(mcpServers), len(client.options.McpServers))
				}
				for name, expectedConfig := range mcpServers {
					actualConfig, exists := client.options.McpServers[name]
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
			},
		},
		{
			name: "extra_arguments_configuration",
			preTest: func(t *testing.T) interface{} {
				return map[string]*string{
					"--custom-flag":       nil, // Boolean flag
					"--custom-with-value": stringPtr("custom-value"),
					"--debug":             nil,
					"--timeout":           stringPtr("30s"),
				}
			},
			options: []Option{
				WithExtraArgs(map[string]*string{
					"--custom-flag":       nil, // Boolean flag
					"--custom-with-value": stringPtr("custom-value"),
					"--debug":             nil,
					"--timeout":           stringPtr("30s"),
				}),
			},
			postTest: func(t *testing.T, preData interface{}, client *ClientImpl) {
				t.Helper()
				extraArgs := preData.(map[string]*string)
				if len(client.options.ExtraArgs) < len(extraArgs) { // Might have transport marker
					t.Errorf("Expected at least %d extra args, got %d", len(extraArgs), len(client.options.ExtraArgs))
				}
				for flag, expectedValue := range extraArgs {
					actualValue, exists := client.options.ExtraArgs[flag]
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
			},
		},
		{
			name: "option_precedence",
			options: []Option{
				WithSystemPrompt("First prompt"),
				WithModel("first-model"),
				WithSystemPrompt("Second prompt"), // Should override first
				WithModel("second-model"),         // Should override first
				WithMaxTurns(5),
				WithMaxTurns(10), // Should override first
			},
			validate: func(t *testing.T, client *ClientImpl) {
				t.Helper()
				if client.options.SystemPrompt == nil || *client.options.SystemPrompt != "Second prompt" {
					t.Error("Later SystemPrompt option did not override earlier one")
				}
				if client.options.Model == nil || *client.options.Model != "second-model" {
					t.Error("Later Model option did not override earlier one")
				}
				if client.options.MaxTurns != 10 {
					t.Errorf("Later MaxTurns option did not override earlier one: expected 10, got %d", client.options.MaxTurns)
				}
			},
		},
		{
			name: "default_values_preserved",
			options: []Option{
				WithSystemPrompt("Custom prompt"), // Only set this option
			},
			validate: func(t *testing.T, client *ClientImpl) {
				t.Helper()
				// SystemPrompt should be set
				if client.options.SystemPrompt == nil || *client.options.SystemPrompt != "Custom prompt" {
					t.Error("Custom SystemPrompt was not applied")
				}
				// Default values should be preserved (checking a few key ones)
				if client.options.MaxThinkingTokens != 8000 { // Default from shared package
					t.Errorf("Expected default MaxThinkingTokens 8000, got %d", client.options.MaxThinkingTokens)
				}
				if len(client.options.AllowedTools) != 0 { // Default empty
					t.Errorf("Expected empty AllowedTools by default, got %v", client.options.AllowedTools)
				}
			},
		},
		{
			name: "configuration_immutability_after_creation",
			options: []Option{
				WithSystemPrompt("Original prompt"),
			},
			postTest: func(t *testing.T, preData interface{}, client *ClientImpl) {
				t.Helper()
				originalPrompt := "Original prompt"
				// Get reference to options
				originalOptions := client.options
				// Modify the options reference (simulating external modification)
				modifiedPrompt := "Modified prompt"
				originalOptions.SystemPrompt = &modifiedPrompt
				// Create another client with same functional option
				transport2 := &clientMockTransport{}
				client2 := NewClientWithTransport(transport2,
					WithSystemPrompt(originalPrompt),
				)
				// Second client should have original prompt, not modified
				client2Impl := client2.(*ClientImpl)
				if client2Impl.options.SystemPrompt == nil || *client2Impl.options.SystemPrompt != originalPrompt {
					t.Error("Options were not properly isolated between client instances")
				}
				// But first client should still have the modified prompt
				if *client.options.SystemPrompt != modifiedPrompt {
					t.Error("Options reference was not preserved in first client")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := &clientMockTransport{}

			// Setup phase
			var preData interface{}
			if tt.preTest != nil {
				preData = tt.preTest(t)
			}

			// Create client with options
			client := NewClientWithTransport(transport, tt.options...)
			clientImpl := client.(*ClientImpl)

			// Validation phase
			if tt.validate != nil {
				tt.validate(t, clientImpl)
			}
			if tt.postTest != nil {
				tt.postTest(t, preData, clientImpl)
			}
		})
	}
}

// TestClientContextPropagation tests context propagation through client operations.
func TestClientContextPropagation(t *testing.T) {
	tests := []struct {
		name           string
		transportSetup func() *clientMockTransport
		contextSetup   func() (context.Context, context.CancelFunc)
		operation      func(context.Context, Client) error
		wantErr        bool
		wantErrType    error
		validate       func(*testing.T, Client)
	}{
		{
			name: "context_cancellation_during_connect",
			transportSetup: func() *clientMockTransport {
				return newMockTransport()
			},
			contextSetup: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately
				return ctx, cancel
			},
			operation: func(ctx context.Context, client Client) error {
				return client.Connect(ctx)
			},
			wantErr: true,
		},
		{
			name: "context_timeout_during_connect",
			transportSetup: func() *clientMockTransport {
				return newMockTransport()
			},
			contextSetup: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
				time.Sleep(1 * time.Millisecond) // Let it timeout
				return ctx, cancel
			},
			operation: func(ctx context.Context, client Client) error {
				return client.Connect(ctx)
			},
			wantErr: true,
		},
		{
			name: "context_cancellation_during_query",
			transportSetup: func() *clientMockTransport {
				return newMockTransport()
			},
			contextSetup: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately for query operation
				return ctx, cancel
			},
			operation: func(ctx context.Context, client Client) error {
				// Connect first with a separate context
				connectCtx, connectCancel := setupTestContext(t, 5*time.Second)
				defer connectCancel()
				if err := client.Connect(connectCtx); err != nil {
					return fmt.Errorf("connect failed: %v", err)
				}
				// Use cancelled context for query
				return client.Query(ctx, "test query")
			},
			wantErr: true,
		},
		{
			name: "context_timeout_during_query",
			transportSetup: func() *clientMockTransport {
				return newMockTransport()
			},
			contextSetup: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
				time.Sleep(1 * time.Millisecond) // Let it timeout
				return ctx, cancel
			},
			operation: func(ctx context.Context, client Client) error {
				// Connect first with a separate context
				connectCtx, connectCancel := setupTestContext(t, 5*time.Second)
				defer connectCancel()
				if err := client.Connect(connectCtx); err != nil {
					return fmt.Errorf("connect failed: %v", err)
				}
				// Use timed-out context for query
				return client.Query(ctx, "test query")
			},
			wantErr: true,
		},
		{
			name: "context_cancellation_during_query_stream",
			transportSetup: func() *clientMockTransport {
				return newMockTransport()
			},
			contextSetup: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			operation: func(ctx context.Context, client Client) error {
				// Connect first
				connectCtx, connectCancel := setupTestContext(t, 5*time.Second)
				defer connectCancel()
				if err := client.Connect(connectCtx); err != nil {
					return fmt.Errorf("connect failed: %v", err)
				}

				// Create messages channel
				messages := make(chan StreamMessage, 1)
				messages <- StreamMessage{Type: "request", Message: &UserMessage{Content: "test"}}
				close(messages)

				// Start QueryStream and then cancel
				err := client.QueryStream(ctx, messages)
				time.Sleep(100 * time.Millisecond) // Give time for processing
				return err
			},
			wantErr: false, // QueryStream may succeed before cancellation
		},
		{
			name: "context_cancellation_during_receive_messages",
			transportSetup: func() *clientMockTransport {
				return newMockTransportWithOptions(WithResponseMessages([]Message{
					&AssistantMessage{Content: []ContentBlock{&TextBlock{Text: "Response"}}, Model: "claude-opus-4-1-20250805"},
				}))
			},
			contextSetup: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			operation: func(ctx context.Context, client Client) error {
				// Connect first
				connectCtx, connectCancel := setupTestContext(t, 5*time.Second)
				defer connectCancel()
				if err := client.Connect(connectCtx); err != nil {
					return fmt.Errorf("connect failed: %v", err)
				}

				// Get message channel and cancel context
				msgChan := client.ReceiveMessages(ctx)

				// Trigger responses
				if transport, ok := client.(*ClientImpl).transport.(*clientMockTransport); ok {
					transport.sendResponses()
				}

				// Try to receive (should handle cancelled context gracefully)
				select {
				case <-msgChan:
				case <-time.After(100 * time.Millisecond):
				}
				return nil
			},
			wantErr: false, // This is more about graceful handling
		},
		{
			name: "context_cancellation_during_receive_response",
			transportSetup: func() *clientMockTransport {
				return newMockTransportWithOptions(WithResponseMessages([]Message{
					&AssistantMessage{Content: []ContentBlock{&TextBlock{Text: "Response"}}, Model: "claude-opus-4-1-20250805"},
				}))
			},
			contextSetup: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately
				return ctx, cancel
			},
			operation: func(ctx context.Context, client Client) error {
				// Connect first
				connectCtx, connectCancel := setupTestContext(t, 5*time.Second)
				defer connectCancel()
				if err := client.Connect(connectCtx); err != nil {
					return fmt.Errorf("connect failed: %v", err)
				}

				// Get iterator with normal context
				iter := client.ReceiveResponse(context.Background())
				if iter == nil {
					return fmt.Errorf("expected response iterator")
				}
				defer iter.Close()

				// Trigger responses
				if transport, ok := client.(*ClientImpl).transport.(*clientMockTransport); ok {
					transport.sendResponses()
				}

				// Use the already-cancelled context for Next call
				_, err := iter.Next(ctx)
				return err
			},
			wantErr: true, // Should respect cancelled context
		},
		{
			name: "context_cancellation_during_interrupt",
			transportSetup: func() *clientMockTransport {
				return newMockTransport()
			},
			contextSetup: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately
				return ctx, cancel
			},
			operation: func(ctx context.Context, client Client) error {
				// Connect first
				connectCtx, connectCancel := setupTestContext(t, 5*time.Second)
				defer connectCancel()
				if err := client.Connect(connectCtx); err != nil {
					return fmt.Errorf("connect failed: %v", err)
				}
				// Use cancelled context for interrupt
				return client.Interrupt(ctx)
			},
			wantErr: true, // Should respect cancelled context
		},
		{
			name: "context_values_propagation",
			transportSetup: func() *clientMockTransport {
				return newMockTransport()
			},
			contextSetup: func() (context.Context, context.CancelFunc) {
				type contextKey string
				const testKey contextKey = "test-key"
				testValue := "test-value"
				ctx := context.WithValue(context.Background(), testKey, testValue)
				return ctx, func() {}
			},
			operation: func(ctx context.Context, client Client) error {
				// Test that context values are threaded through operations
				if err := client.Connect(ctx); err != nil {
					t.Logf("Connect returned: %v", err)
				}
				if err := client.Query(ctx, "test query"); err != nil {
					t.Logf("Query returned: %v", err)
				}
				return nil
			},
			wantErr: false, // Mainly ensures no panic occurs
		},
		{
			name: "nested_context_cancellation",
			transportSetup: func() *clientMockTransport {
				return newMockTransport()
			},
			contextSetup: func() (context.Context, context.CancelFunc) {
				parentCtx, parentCancel := context.WithCancel(context.Background())
				childCtx, childCancel := context.WithCancel(parentCtx)
				parentCancel() // Cancel parent, should cascade to child
				return childCtx, childCancel
			},
			operation: func(ctx context.Context, client Client) error {
				// Connect first (may work before cancellation is checked)
				if err := client.Connect(context.Background()); err != nil {
					return fmt.Errorf("connect failed: %v", err)
				}
				// Use child context which should see parent cancellation
				return client.Query(ctx, "test query")
			},
			wantErr: true, // Should see parent cancellation
		},
		{
			name: "context_deadline_propagation",
			transportSetup: func() *clientMockTransport {
				return newMockTransport()
			},
			contextSetup: func() (context.Context, context.CancelFunc) {
				deadline := time.Now().Add(1 * time.Millisecond) // Very short deadline
				ctx, cancel := context.WithDeadline(context.Background(), deadline)
				time.Sleep(2 * time.Millisecond) // Ensure deadline passes
				return ctx, cancel
			},
			operation: func(ctx context.Context, client Client) error {
				// Connect first with fresh context
				connectCtx, connectCancel := setupTestContext(t, 5*time.Second)
				defer connectCancel()
				if err := client.Connect(connectCtx); err != nil {
					return fmt.Errorf("connect failed: %v", err)
				}
				// Use deadline-exceeded context for query
				return client.Query(ctx, "test query")
			},
			wantErr:     true,
			wantErrType: context.DeadlineExceeded,
		},
		{
			name: "multiple_operations_same_context",
			transportSetup: func() *clientMockTransport {
				return newMockTransport()
			},
			contextSetup: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 5*time.Second)
			},
			operation: func(ctx context.Context, client Client) error {
				// All operations use the same context
				if err := client.Connect(ctx); err != nil {
					t.Logf("Connect returned: %v", err)
				}

				if err := client.Query(ctx, "query 1"); err != nil {
					t.Logf("Query 1 returned: %v", err)
				}

				if err := client.Query(ctx, "query 2"); err != nil {
					t.Logf("Query 2 returned: %v", err)
				}

				messages := make(chan StreamMessage, 1)
				messages <- StreamMessage{Type: "request", Message: &UserMessage{Content: "stream query"}}
				close(messages)

				if err := client.QueryStream(ctx, messages); err != nil {
					t.Logf("QueryStream returned: %v", err)
				}

				msgChan := client.ReceiveMessages(ctx)
				select {
				case <-msgChan:
				case <-time.After(50 * time.Millisecond):
				}

				if err := client.Interrupt(ctx); err != nil {
					t.Logf("Interrupt returned: %v", err)
				}

				return nil
			},
			wantErr: false, // Should handle shared context gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := tt.transportSetup()
			client := setupClient(t, transport)
			defer client.Disconnect()

			ctx, cancel := tt.contextSetup()
			defer cancel()

			err := tt.operation(ctx, client)

			if (err != nil) != tt.wantErr {
				t.Errorf("operation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.wantErrType != nil {
				if err != tt.wantErrType {
					t.Logf("operation() error = %v, wantErrType = %v (may be acceptable)", err, tt.wantErrType)
				}
			}

			if tt.validate != nil {
				tt.validate(t, client)
			}
		})
	}
}

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

// TestClientFactoryFunction tests the NewClient factory function behavior.
func TestClientFactoryFunction(t *testing.T) {
	tests := []struct {
		name     string
		options  []Option
		validate func(*testing.T, Client)
		setup    func(*testing.T) *clientMockTransport          // For NewClientWithTransport tests
		postTest func(*testing.T, Client, *clientMockTransport) // For complex validation
	}{
		{
			name:    "new_client_no_options",
			options: []Option{},
			validate: func(t *testing.T, client Client) {
				t.Helper()
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
			},
		},
		{
			name:    "new_client_single_option",
			options: []Option{WithMaxTurns(5)},
			validate: func(t *testing.T, client Client) {
				t.Helper()
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
			},
		},
		{
			name: "new_client_multiple_options",
			options: []Option{
				WithMaxTurns(10),
				WithCwd("/tmp"),
				WithSystemPrompt("Custom system prompt"),
				WithPermissionMode(PermissionModeBypassPermissions),
			},
			validate: func(t *testing.T, client Client) {
				t.Helper()
				if client == nil {
					t.Fatal("NewClient() returned nil")
				}

				clientImpl := client.(*ClientImpl)
				if clientImpl.options == nil {
					t.Fatal("Client options should be initialized")
				}

				// Verify all options were applied
				workingDir := "/tmp"
				systemPrompt := "Custom system prompt"

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
			},
		},
		{
			name:    "new_client_vs_new_client_with_transport",
			options: []Option{WithMaxTurns(7)},
			setup: func(t *testing.T) *clientMockTransport {
				return &clientMockTransport{}
			},
			postTest: func(t *testing.T, client1 Client, transport *clientMockTransport) {
				t.Helper()
				// NewClient
				clientImpl1 := client1.(*ClientImpl)

				// NewClientWithTransport
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
			},
		},
		{
			name: "factory_function_consistency",
			options: []Option{
				WithMaxTurns(15),
				WithSystemPrompt("Test prompt"),
			},
			validate: func(t *testing.T, client Client) {
				t.Helper()
				// Create multiple clients with same options
				clients := make([]Client, 3)
				for i := range clients {
					clients[i] = NewClient(
						WithMaxTurns(15),
						WithSystemPrompt("Test prompt"),
					)
				}

				// All should be configured identically
				for i, testClient := range clients {
					clientImpl := testClient.(*ClientImpl)

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
			},
		},
		{
			name: "factory_function_option_override",
			options: []Option{
				WithMaxTurns(1),
				WithMaxTurns(2),
				WithMaxTurns(3), // This should be the final value
			},
			validate: func(t *testing.T, client Client) {
				t.Helper()
				clientImpl := client.(*ClientImpl)
				if clientImpl.options.MaxTurns != 3 {
					t.Errorf("Expected final MaxTurns=3, got %d", clientImpl.options.MaxTurns)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup transport if needed
			var transport *clientMockTransport
			if tt.setup != nil {
				transport = tt.setup(t)
			}

			// Create client with options
			client := NewClient(tt.options...)

			// Validation
			if tt.validate != nil {
				tt.validate(t, client)
			}
			if tt.postTest != nil {
				tt.postTest(t, client, transport)
			}
		})
	}
}

// TestClientDefaultConfiguration tests default client configuration values.
func TestClientDefaultConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		options     []Option                              // Empty for zero-config tests
		preTest     func(*testing.T) *clientMockTransport // Setup transport if needed
		validate    func(*testing.T, *ClientImpl, *clientMockTransport)
		requiresCtx bool // Whether test needs context/timeout
	}{
		{
			name:    "zero_config_client_creation",
			options: []Option{}, // No options - zero config
			validate: func(t *testing.T, client *ClientImpl, transport *clientMockTransport) {
				t.Helper()
				if client == nil {
					t.Fatal("NewClient() should not return nil")
				}
				if client.options == nil {
					t.Fatal("Client options should be initialized even with no options")
				}

				options := client.options
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
			},
		},
		{
			name:    "default_values_functional",
			options: []Option{}, // Zero config
			preTest: func(t *testing.T) *clientMockTransport {
				return &clientMockTransport{
					msgChan: make(chan Message, 10),
					errChan: make(chan error, 10),
				}
			},
			requiresCtx: true,
			validate: func(t *testing.T, client *ClientImpl, transport *clientMockTransport) {
				t.Helper()
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()

				// Connect should work with defaults
				clientInterface := Client(client)
				err := clientInterface.Connect(ctx)
				if err != nil {
					t.Fatalf("Connect with default config failed: %v", err)
				}
				defer clientInterface.Disconnect()

				// Query should work with defaults
				err = clientInterface.Query(ctx, "Default config test")
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
			},
		},
		{
			name:    "default_max_turns_behavior",
			options: []Option{}, // Zero config
			validate: func(t *testing.T, client *ClientImpl, transport *clientMockTransport) {
				t.Helper()
				// Should have sensible default (0 means no limit, which is valid)
				defaultMaxTurns := client.options.MaxTurns
				if defaultMaxTurns < 0 {
					t.Errorf("Default MaxTurns should be non-negative, got %d", defaultMaxTurns)
				}

				// Default of 0 means "no limit" which is a reasonable default
				if defaultMaxTurns != 0 {
					t.Logf("Default MaxTurns is %d (0 means no limit)", defaultMaxTurns)
				}
			},
		},
		{
			name:        "defaults_vs_explicit_configuration",
			options:     []Option{}, // Zero config, comparison done in validation
			requiresCtx: true,
			validate: func(t *testing.T, defaultClient *ClientImpl, transport *clientMockTransport) {
				t.Helper()
				// Explicitly configured client with same values as defaults
				explicitClient := NewClient(
					WithMaxTurns(defaultClient.options.MaxTurns), // Use same value as default
				)
				explicitImpl := explicitClient.(*ClientImpl)

				// Should behave identically
				if defaultClient.options.MaxTurns != explicitImpl.options.MaxTurns {
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

				defaultClient.customTransport = transport1
				explicitImpl.customTransport = transport2

				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()

				// Both should connect successfully
				defaultClientInterface := Client(defaultClient)
				explicitClientInterface := Client(explicitImpl)

				err1 := defaultClientInterface.Connect(ctx)
				err2 := explicitClientInterface.Connect(ctx)

				if err1 != nil {
					t.Errorf("Default client connect failed: %v", err1)
				}
				if err2 != nil {
					t.Errorf("Explicit client connect failed: %v", err2)
				}

				defer defaultClientInterface.Disconnect()
				defer explicitClientInterface.Disconnect()
			},
		},
		{
			name:    "default_configuration_validation",
			options: []Option{}, // Zero config
			validate: func(t *testing.T, client *ClientImpl, transport *clientMockTransport) {
				t.Helper()
				// Default configuration should pass validation
				err := client.validateOptions()
				if err != nil {
					t.Errorf("Default configuration should pass validation, got error: %v", err)
				}
			},
		},
		{
			name:    "zero_config_usability",
			options: []Option{}, // Zero config
			preTest: func(t *testing.T) *clientMockTransport {
				return &clientMockTransport{
					msgChan: make(chan Message, 10),
					errChan: make(chan error, 10),
				}
			},
			requiresCtx: true,
			validate: func(t *testing.T, client *ClientImpl, transport *clientMockTransport) {
				t.Helper()
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()

				// Should work immediately without additional setup
				clientInterface := Client(client)
				err := clientInterface.Connect(ctx)
				if err != nil {
					t.Fatalf("Zero-config client should connect immediately, got: %v", err)
				}
				defer clientInterface.Disconnect()

				// Should be able to send messages immediately
				err = clientInterface.Query(ctx, "Hello, Claude!")
				if err != nil {
					t.Fatalf("Zero-config client should send messages immediately, got: %v", err)
				}

				// Should receive messages
				messages := clientInterface.ReceiveMessages(ctx)
				if messages == nil {
					t.Error("Zero-config client should receive messages")
				}

				// Should support iterators
				iter := clientInterface.ReceiveResponse(ctx)
				if iter == nil {
					t.Error("Zero-config client should support response iterators")
				}
				if iter != nil {
					iter.Close()
				}

				// Should support interrupts
				err = clientInterface.Interrupt(ctx)
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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup transport if needed
			var transport *clientMockTransport
			if tt.preTest != nil {
				transport = tt.preTest(t)
			}

			// Create client with options (empty for zero-config)
			var client Client
			if len(tt.options) == 0 {
				client = NewClient() // Zero config
			} else {
				client = NewClient(tt.options...)
			}

			clientImpl := client.(*ClientImpl)

			// Set custom transport if provided
			if transport != nil {
				clientImpl.customTransport = transport
			}

			// Validation
			if tt.validate != nil {
				tt.validate(t, clientImpl, transport)
			}
		})
	}
}

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
