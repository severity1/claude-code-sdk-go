package claudecode

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// T121: Simple Query Execution 🔴 RED
// Python Reference: test_client.py::TestQueryFunction::test_query_single_prompt
func TestSimpleQueryExecution(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create mock CLI for testing
	mockCLI := createMockQueryCLI(t)

	// Create a mock transport for testing
	transport := &mockTransport{
		mockCLI: mockCLI,
	}

	// Execute simple text query with QueryWithTransport function
	iter, err := QueryWithTransport(ctx, "What is 2+2?", transport, WithCLIPath(mockCLI))
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	defer iter.Close()

	// Must return MessageIterator for one-shot queries
	if iter == nil {
		t.Fatal("Expected MessageIterator, got nil")
	}

	// Collect all messages
	var messages []Message
	for {
		msg, err := iter.Next(ctx)
		if err != nil {
			if err == ErrNoMoreMessages {
				break
			}
			t.Fatalf("Iterator error: %v", err)
		}
		if msg != nil { // Only append non-nil messages
			messages = append(messages, msg)
		}
	}

	// Verify we got exactly one assistant message (matching Python test)
	if len(messages) != 1 {
		t.Logf("Messages received: %d", len(messages))
		for i, msg := range messages {
			t.Logf("Message %d: %T", i, msg)
		}
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	assistantMsg, ok := messages[0].(*AssistantMessage)
	if !ok {
		t.Fatalf("Expected AssistantMessage, got %T", messages[0])
	}

	// Verify message structure matches Python test expectations
	if len(assistantMsg.Content) != 1 {
		t.Fatalf("Expected 1 content block, got %d", len(assistantMsg.Content))
	}

	textBlock, ok := assistantMsg.Content[0].(*TextBlock)
	if !ok {
		t.Fatalf("Expected TextBlock, got %T", assistantMsg.Content[0])
	}

	if textBlock.Text != "4" {
		t.Errorf("Expected text '4', got '%s'", textBlock.Text)
	}

	if assistantMsg.Model != "claude-opus-4-1-20250805" {
		t.Errorf("Expected model 'claude-opus-4-1-20250805', got '%s'", assistantMsg.Model)
	}
}

// Helper to create mock CLI that mimics Python test behavior
func createMockQueryCLI(t *testing.T) string {
	script := `#!/bin/bash
# Mock Claude CLI that returns the response expected by Python test
echo '{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"4"}],"model":"claude-opus-4-1-20250805"}}'
echo '{"type":"result","subtype":"success","duration_ms":1000,"duration_api_ms":800,"is_error":false,"num_turns":1,"session_id":"test-session","total_cost_usd":0.01}'
sleep 0.1
`
	return createTempScript(t, script)
}

func createTempScript(t *testing.T, script string) string {
	tempDir := t.TempDir()
	scriptPath := filepath.Join(tempDir, "mock-claude")

	err := os.WriteFile(scriptPath, []byte(script), 0755)
	if err != nil {
		t.Fatalf("Failed to create mock CLI script: %v", err)
	}

	return scriptPath
}

// mockTransport implements Transport interface for testing
type mockTransport struct {
	mockCLI   string
	connected bool
	msgChan   chan Message
	errChan   chan error
}

func (m *mockTransport) Connect(ctx context.Context) error {
	m.connected = true
	m.msgChan = make(chan Message, 10)
	m.errChan = make(chan error, 10)

	// Simulate assistant response
	assistantMsg := &AssistantMessage{
		Content: []ContentBlock{
			&TextBlock{Text: "4"},
		},
		Model: "claude-opus-4-1-20250805",
	}

	// Send the message asynchronously
	go func() {
		defer close(m.msgChan)
		defer close(m.errChan)

		// Only send the assistant message, matching Python test behavior
		m.msgChan <- assistantMsg
	}()

	return nil
}

func (m *mockTransport) SendMessage(ctx context.Context, message StreamMessage) error {
	if !m.connected {
		return fmt.Errorf("not connected")
	}
	// Mock transport accepts any message
	return nil
}

func (m *mockTransport) ReceiveMessages(ctx context.Context) (<-chan Message, <-chan error) {
	return m.msgChan, m.errChan
}

func (m *mockTransport) Interrupt(ctx context.Context) error {
	return nil
}

func (m *mockTransport) Close() error {
	m.connected = false
	return nil
}

// T126: Query Stream Input - Test processStreamQuery and streamIterator
func TestQueryStreamInput(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create mock transport for streaming
	transport := &mockStreamTransport{}

	// Create a message channel with test messages
	messages := make(chan StreamMessage, 3)
	messages <- StreamMessage{Type: "request", Message: &UserMessage{Content: "First message"}}
	messages <- StreamMessage{Type: "request", Message: &UserMessage{Content: "Second message"}}
	close(messages)

	// Execute stream query with QueryStreamWithTransport function
	iter, err := QueryStreamWithTransport(ctx, messages, transport)
	if err != nil {
		t.Fatalf("QueryStream failed: %v", err)
	}
	defer iter.Close()

	// Must return MessageIterator for stream queries
	if iter == nil {
		t.Fatal("Expected MessageIterator, got nil")
	}

	// Collect all messages
	var receivedMessages []Message
	for {
		msg, err := iter.Next(ctx)
		if err != nil {
			if err == ErrNoMoreMessages {
				break
			}
			t.Fatalf("Iterator error: %v", err)
		}
		if msg != nil {
			receivedMessages = append(receivedMessages, msg)
		}
	}

	// Verify we got the expected response
	if len(receivedMessages) != 1 {
		t.Fatalf("Expected 1 response message, got %d", len(receivedMessages))
	}

	assistantMsg, ok := receivedMessages[0].(*AssistantMessage)
	if !ok {
		t.Fatalf("Expected AssistantMessage, got %T", receivedMessages[0])
	}

	if len(assistantMsg.Content) != 1 {
		t.Fatalf("Expected 1 content block, got %d", len(assistantMsg.Content))
	}

	textBlock, ok := assistantMsg.Content[0].(*TextBlock)
	if !ok {
		t.Fatalf("Expected TextBlock, got %T", assistantMsg.Content[0])
	}

	if textBlock.Text != "Stream response" {
		t.Errorf("Expected text 'Stream response', got '%s'", textBlock.Text)
	}

	// Verify the transport received both messages
	if transport.ReceivedMessageCount() != 2 {
		t.Errorf("Expected transport to receive 2 messages, got %d", transport.ReceivedMessageCount())
	}
}

// Test context cancellation during streaming
func TestQueryStreamContextCancellation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	transport := &mockStreamTransport{
		delay: 200 * time.Millisecond, // Longer than context timeout
	}

	// Create a message channel
	messages := make(chan StreamMessage, 1)
	messages <- StreamMessage{Type: "request", Message: &UserMessage{Content: "Test message"}}
	close(messages)

	iter, err := QueryStreamWithTransport(ctx, messages, transport)
	if err != nil {
		t.Fatalf("QueryStream failed: %v", err)
	}
	defer iter.Close()

	// This should timeout due to context cancellation
	_, err = iter.Next(ctx)
	if err == nil {
		t.Fatal("Expected context cancellation error, got nil")
	}
	if err != context.DeadlineExceeded {
		t.Errorf("Expected context.DeadlineExceeded, got %v", err)
	}
}

// Test resource cleanup in stream iterator
func TestQueryStreamResourceCleanup(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	transport := &mockStreamTransport{}

	messages := make(chan StreamMessage, 1)
	messages <- StreamMessage{Type: "request", Message: &UserMessage{Content: "Test"}}
	close(messages)

	iter, err := QueryStreamWithTransport(ctx, messages, transport)
	if err != nil {
		t.Fatalf("QueryStream failed: %v", err)
	}

	// Close the iterator before reading all messages
	err = iter.Close()
	if err != nil {
		t.Errorf("Close() failed: %v", err)
	}

	// Verify transport was closed
	if transport.connected {
		t.Error("Expected transport to be closed")
	}

	// Further calls to Next should return ErrNoMoreMessages
	_, err = iter.Next(ctx)
	if err != ErrNoMoreMessages {
		t.Errorf("Expected ErrNoMoreMessages after Close(), got %v", err)
	}
}

// mockStreamTransport implements Transport interface for stream testing
type mockStreamTransport struct {
	mu               sync.RWMutex
	connected        bool
	msgChan          chan Message
	errChan          chan error
	receivedMessages []StreamMessage
	delay            time.Duration
}

func (m *mockStreamTransport) Connect(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.connected = true
	m.msgChan = make(chan Message, 10)
	m.errChan = make(chan error, 10)
	m.receivedMessages = make([]StreamMessage, 0)

	// Send response after a delay (for testing timeouts)
	go func() {
		defer close(m.msgChan)
		defer close(m.errChan)

		if m.delay > 0 {
			time.Sleep(m.delay)
		}

		// Send a single assistant response
		assistantMsg := &AssistantMessage{
			Content: []ContentBlock{
				&TextBlock{Text: "Stream response"},
			},
			Model: "claude-opus-4-1-20250805",
		}
		m.msgChan <- assistantMsg
	}()

	return nil
}

func (m *mockStreamTransport) SendMessage(ctx context.Context, message StreamMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.connected {
		return fmt.Errorf("not connected")
	}
	// Record the message
	m.receivedMessages = append(m.receivedMessages, message)
	return nil
}

func (m *mockStreamTransport) ReceiveMessages(ctx context.Context) (<-chan Message, <-chan error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.msgChan, m.errChan
}

func (m *mockStreamTransport) Interrupt(ctx context.Context) error {
	return nil
}

func (m *mockStreamTransport) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.connected = false
	return nil
}

func (m *mockStreamTransport) ReceivedMessageCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.receivedMessages)
}

// Test transport error handling during query
func TestQueryTransportError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test with nil transport (should return error)
	iter, err := QueryWithTransport(ctx, "test", nil)
	if err == nil {
		t.Fatal("Expected error with nil transport, got nil")
	}
	if iter != nil {
		t.Fatal("Expected nil iterator with error, got non-nil")
	}

	expectedErr := "transport is required"
	if !contains(err.Error(), expectedErr) {
		t.Errorf("Expected error containing '%s', got '%s'", expectedErr, err.Error())
	}
}

// Test stream transport error handling
func TestQueryStreamTransportError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	messages := make(chan StreamMessage, 1)
	messages <- StreamMessage{Type: "request", Message: &UserMessage{Content: "Test"}}
	close(messages)

	// Test with nil transport (should return error)
	iter, err := QueryStreamWithTransport(ctx, messages, nil)
	if err == nil {
		t.Fatal("Expected error with nil transport, got nil")
	}
	if iter != nil {
		t.Fatal("Expected nil iterator with error, got non-nil")
	}

	expectedErr := "transport is required"
	if !contains(err.Error(), expectedErr) {
		t.Errorf("Expected error containing '%s', got '%s'", expectedErr, err.Error())
	}
}

// Test transport connection failure
func TestQueryTransportConnectionFailure(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a transport that fails to connect
	transport := &failingTransport{
		connectError: fmt.Errorf("connection failed"),
	}

	iter, err := QueryWithTransport(ctx, "test", transport)
	if err != nil {
		t.Fatalf("QueryWithTransport should not fail immediately: %v", err)
	}
	defer iter.Close()

	// The error should occur when we try to get the next message
	_, err = iter.Next(ctx)
	if err == nil {
		t.Fatal("Expected connection error, got nil")
	}

	expectedErr := "failed to connect transport"
	if !contains(err.Error(), expectedErr) {
		t.Errorf("Expected error containing '%s', got '%s'", expectedErr, err.Error())
	}
}

// Test transport send message failure during streaming
func TestQueryStreamSendMessageFailure(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	transport := &failingTransport{
		sendError: fmt.Errorf("send failed"),
	}

	messages := make(chan StreamMessage, 1)
	messages <- StreamMessage{Type: "request", Message: &UserMessage{Content: "Test"}}
	close(messages)

	iter, err := QueryStreamWithTransport(ctx, messages, transport)
	if err != nil {
		t.Fatalf("QueryStreamWithTransport should not fail immediately: %v", err)
	}
	defer iter.Close()

	// Start the iterator to trigger the send
	_, err = iter.Next(ctx)

	// The send error should be handled gracefully (goroutine will exit)
	// The iterator should eventually return ErrNoMoreMessages when channels close
	if err != ErrNoMoreMessages {
		// Give some time for the goroutine to handle the error
		time.Sleep(100 * time.Millisecond)
		_, err = iter.Next(ctx)
	}

	if err != ErrNoMoreMessages {
		t.Errorf("Expected ErrNoMoreMessages after send failure, got %v", err)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || (len(s) > len(substr) &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
				containsMiddle(s, substr))))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// failingTransport implements Transport interface but fails in specific ways for testing
type failingTransport struct {
	connectError error
	sendError    error
	connected    bool
	msgChan      chan Message
	errChan      chan error
}

func (f *failingTransport) Connect(ctx context.Context) error {
	if f.connectError != nil {
		return f.connectError
	}
	f.connected = true
	f.msgChan = make(chan Message, 1)
	f.errChan = make(chan error, 1)

	// Close channels immediately to simulate no responses
	close(f.msgChan)
	close(f.errChan)

	return nil
}

func (f *failingTransport) SendMessage(ctx context.Context, message StreamMessage) error {
	if f.sendError != nil {
		return f.sendError
	}
	return nil
}

func (f *failingTransport) ReceiveMessages(ctx context.Context) (<-chan Message, <-chan error) {
	return f.msgChan, f.errChan
}

func (f *failingTransport) Interrupt(ctx context.Context) error {
	return nil
}

func (f *failingTransport) Close() error {
	f.connected = false
	return nil
}

// T122: Query with Options 🔴 RED
// Python Reference: test_client.py::TestQueryFunction::test_query_with_options
func TestQueryWithOptions(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create mock transport that can verify options
	transport := &optionsTestTransport{}

	// Test various configuration options matching Python test
	options := []Option{
		WithAllowedTools("Read", "Write"),
		WithSystemPrompt("You are helpful"),
		WithPermissionMode("acceptEdits"),
		WithMaxTurns(5),
	}

	// Execute query with options
	iter, err := QueryWithTransport(ctx, "Hi", transport, options...)
	if err != nil {
		t.Fatalf("Query with options failed: %v", err)
	}
	defer iter.Close()

	// Collect messages
	var messages []Message
	for {
		msg, err := iter.Next(ctx)
		if err != nil {
			if err == ErrNoMoreMessages {
				break
			}
			t.Fatalf("Iterator error: %v", err)
		}
		if msg != nil {
			messages = append(messages, msg)
		}
	}

	// Verify we got the expected response
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	assistantMsg, ok := messages[0].(*AssistantMessage)
	if !ok {
		t.Fatalf("Expected AssistantMessage, got %T", messages[0])
	}

	if len(assistantMsg.Content) != 1 {
		t.Fatalf("Expected 1 content block, got %d", len(assistantMsg.Content))
	}

	textBlock, ok := assistantMsg.Content[0].(*TextBlock)
	if !ok {
		t.Fatalf("Expected TextBlock, got %T", assistantMsg.Content[0])
	}

	if textBlock.Text != "Hello!" {
		t.Errorf("Expected text 'Hello!', got '%s'", textBlock.Text)
	}

	// Verify options were applied correctly to transport
	if transport.receivedOptions == nil {
		t.Fatal("Expected options to be passed to transport")
	}

	// Verify specific option values
	if len(transport.receivedOptions.AllowedTools) != 2 {
		t.Errorf("Expected 2 allowed tools, got %d", len(transport.receivedOptions.AllowedTools))
	}
	if transport.receivedOptions.AllowedTools[0] != "Read" || transport.receivedOptions.AllowedTools[1] != "Write" {
		t.Errorf("Expected allowed tools [Read, Write], got %v", transport.receivedOptions.AllowedTools)
	}

	if transport.receivedOptions.SystemPrompt == nil || *transport.receivedOptions.SystemPrompt != "You are helpful" {
		var actualValue string
		if transport.receivedOptions.SystemPrompt == nil {
			actualValue = "nil"
		} else {
			actualValue = *transport.receivedOptions.SystemPrompt
		}
		t.Errorf("Expected system prompt 'You are helpful', got '%s'", actualValue)
	}

	expectedPermissionMode := PermissionModeAcceptEdits
	if transport.receivedOptions.PermissionMode == nil || *transport.receivedOptions.PermissionMode != expectedPermissionMode {
		var actualValue string
		if transport.receivedOptions.PermissionMode == nil {
			actualValue = "nil"
		} else {
			actualValue = string(*transport.receivedOptions.PermissionMode)
		}
		t.Errorf("Expected permission mode 'acceptEdits', got '%s'", actualValue)
	}

	if transport.receivedOptions.MaxTurns != 5 {
		t.Errorf("Expected max turns 5, got %d", transport.receivedOptions.MaxTurns)
	}
}

// optionsTestTransport implements Transport to verify options are passed correctly
type optionsTestTransport struct {
	connected       bool
	msgChan         chan Message
	errChan         chan error
	receivedOptions *Options
	receivedPrompt  string
}

func (o *optionsTestTransport) SetOptions(options *Options) {
	o.receivedOptions = options
}

func (o *optionsTestTransport) Connect(ctx context.Context) error {
	o.connected = true
	o.msgChan = make(chan Message, 10)
	o.errChan = make(chan error, 10)

	// Send response matching Python test expectation
	assistantMsg := &AssistantMessage{
		Content: []ContentBlock{
			&TextBlock{Text: "Hello!"},
		},
		Model: "claude-opus-4-1-20250805",
	}

	go func() {
		defer close(o.msgChan)
		defer close(o.errChan)
		o.msgChan <- assistantMsg
	}()

	return nil
}

func (o *optionsTestTransport) SendMessage(ctx context.Context, message StreamMessage) error {
	if !o.connected {
		return fmt.Errorf("not connected")
	}
	return nil
}

func (o *optionsTestTransport) ReceiveMessages(ctx context.Context) (<-chan Message, <-chan error) {
	return o.msgChan, o.errChan
}

func (o *optionsTestTransport) Interrupt(ctx context.Context) error {
	return nil
}

func (o *optionsTestTransport) Close() error {
	o.connected = false
	return nil
}

// T123: Query Response Processing 🔴 RED
// Go Target: query_test.go::TestQueryResponseProcessing
func TestQueryResponseProcessing(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create transport that sends multiple message types
	transport := &multiMessageTransport{}

	// Execute query
	iter, err := QueryWithTransport(ctx, "Test query", transport)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	defer iter.Close()

	// Collect all messages and process them according to their types
	var userMessages []*UserMessage
	var assistantMessages []*AssistantMessage
	var systemMessages []*SystemMessage
	var resultMessages []*ResultMessage

	messageCount := 0
	for {
		msg, err := iter.Next(ctx)
		if err != nil {
			if err == ErrNoMoreMessages {
				break
			}
			t.Fatalf("Iterator error: %v", err)
		}

		if msg != nil {
			messageCount++

			// Process messages by type (like Python SDK pattern)
			switch m := msg.(type) {
			case *UserMessage:
				userMessages = append(userMessages, m)
			case *AssistantMessage:
				assistantMessages = append(assistantMessages, m)
			case *SystemMessage:
				systemMessages = append(systemMessages, m)
			case *ResultMessage:
				resultMessages = append(resultMessages, m)
			default:
				t.Errorf("Unexpected message type: %T", msg)
			}
		}
	}

	// Verify we processed all message types correctly
	if messageCount != 4 {
		t.Fatalf("Expected 4 messages total, got %d", messageCount)
	}

	if len(userMessages) != 1 {
		t.Errorf("Expected 1 user message, got %d", len(userMessages))
	}

	if len(assistantMessages) != 1 {
		t.Errorf("Expected 1 assistant message, got %d", len(assistantMessages))
	}

	if len(systemMessages) != 1 {
		t.Errorf("Expected 1 system message, got %d", len(systemMessages))
	}

	if len(resultMessages) != 1 {
		t.Errorf("Expected 1 result message, got %d", len(resultMessages))
	}

	// Verify message content processing
	userContent, ok := userMessages[0].Content.(string)
	if !ok {
		t.Fatalf("Expected user content to be string, got %T", userMessages[0].Content)
	}
	if userContent != "User input" {
		t.Errorf("Expected user content 'User input', got '%s'", userContent)
	}

	// Verify assistant message content blocks
	if len(assistantMessages[0].Content) != 2 {
		t.Fatalf("Expected 2 content blocks in assistant message, got %d", len(assistantMessages[0].Content))
	}

	textBlock, ok := assistantMessages[0].Content[0].(*TextBlock)
	if !ok {
		t.Fatalf("Expected first content block to be TextBlock, got %T", assistantMessages[0].Content[0])
	}
	if textBlock.Text != "Assistant response" {
		t.Errorf("Expected text 'Assistant response', got '%s'", textBlock.Text)
	}

	thinkingBlock, ok := assistantMessages[0].Content[1].(*ThinkingBlock)
	if !ok {
		t.Fatalf("Expected second content block to be ThinkingBlock, got %T", assistantMessages[0].Content[1])
	}
	if thinkingBlock.Thinking != "Let me think..." {
		t.Errorf("Expected thinking 'Let me think...', got '%s'", thinkingBlock.Thinking)
	}

	// Verify result message details
	if !resultMessages[0].IsError {
		t.Error("Expected result message to have IsError=true")
	}
	if resultMessages[0].NumTurns != 3 {
		t.Errorf("Expected result message NumTurns=3, got %d", resultMessages[0].NumTurns)
	}
}

// multiMessageTransport sends various message types for comprehensive testing
type multiMessageTransport struct {
	connected bool
	msgChan   chan Message
	errChan   chan error
}

func (m *multiMessageTransport) Connect(ctx context.Context) error {
	m.connected = true
	m.msgChan = make(chan Message, 10)
	m.errChan = make(chan error, 10)

	go func() {
		defer close(m.msgChan)
		defer close(m.errChan)

		// Send user message
		m.msgChan <- &UserMessage{
			Content: "User input",
		}

		// Small delay to ensure delivery
		time.Sleep(1 * time.Millisecond)

		// Send assistant message with mixed content blocks
		m.msgChan <- &AssistantMessage{
			Content: []ContentBlock{
				&TextBlock{Text: "Assistant response"},
				&ThinkingBlock{Thinking: "Let me think...", Signature: "assistant"},
			},
			Model: "claude-opus-4-1-20250805",
		}

		time.Sleep(1 * time.Millisecond)

		// Send system message
		m.msgChan <- &SystemMessage{
			Subtype: "tool_use",
			Data:    map[string]interface{}{"tool": "Read", "file": "test.txt"},
		}

		time.Sleep(1 * time.Millisecond)

		// Send result message
		m.msgChan <- &ResultMessage{
			Subtype:       "error",
			DurationMs:    2500,
			DurationAPIMs: 2000,
			IsError:       true,
			NumTurns:      3,
			SessionID:     "test-session-123",
		}
	}()

	return nil
}

func (m *multiMessageTransport) SendMessage(ctx context.Context, message StreamMessage) error {
	if !m.connected {
		return fmt.Errorf("not connected")
	}
	return nil
}

func (m *multiMessageTransport) ReceiveMessages(ctx context.Context) (<-chan Message, <-chan error) {
	return m.msgChan, m.errChan
}

func (m *multiMessageTransport) Interrupt(ctx context.Context) error {
	return nil
}

func (m *multiMessageTransport) Close() error {
	m.connected = false
	return nil
}

// T124: Query Error Handling 🔴 RED
// Go Target: query_test.go::TestQueryErrorHandling
func TestQueryErrorHandling(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test 1: Transport connection error
	t.Run("TransportConnectionError", func(t *testing.T) {
		transport := &errorTransport{
			connectError: fmt.Errorf("connection failed: network unreachable"),
		}

		iter, err := QueryWithTransport(ctx, "test", transport)
		if err != nil {
			t.Fatalf("Query should not fail immediately: %v", err)
		}
		defer iter.Close()

		// Error should occur on first message fetch
		_, err = iter.Next(ctx)
		if err == nil {
			t.Fatal("Expected error from transport connection failure")
		}

		expectedErrSubstr := "failed to connect transport"
		if !contains(err.Error(), expectedErrSubstr) {
			t.Errorf("Expected error containing '%s', got '%s'", expectedErrSubstr, err.Error())
		}
	})

	// Test 2: Transport error during message receiving
	t.Run("TransportReceiveError", func(t *testing.T) {
		transport := &errorTransport{
			receiveError: fmt.Errorf("receive failed: stream interrupted"),
		}

		iter, err := QueryWithTransport(ctx, "test", transport)
		if err != nil {
			t.Fatalf("Query should not fail immediately: %v", err)
		}
		defer iter.Close()

		// Error should be propagated from transport
		_, err = iter.Next(ctx)
		if err == nil {
			t.Fatal("Expected error from transport receive failure")
		}

		expectedErrSubstr := "receive failed"
		if !contains(err.Error(), expectedErrSubstr) {
			t.Errorf("Expected error containing '%s', got '%s'", expectedErrSubstr, err.Error())
		}
	})

	// Test 3: Transport send message error (for streaming)
	t.Run("TransportSendError", func(t *testing.T) {
		transport := &errorTransport{
			sendError: fmt.Errorf("send failed: broken pipe"),
		}

		messages := make(chan StreamMessage, 1)
		messages <- StreamMessage{Type: "request", Message: &UserMessage{Content: "test"}}
		close(messages)

		iter, err := QueryStreamWithTransport(ctx, messages, transport)
		if err != nil {
			t.Fatalf("QueryStream should not fail immediately: %v", err)
		}
		defer iter.Close()

		// For send errors, the current implementation just exits the goroutine
		// The transport's channels close when the goroutine exits
		// This should eventually result in ErrNoMoreMessages
		var finalErr error
		for i := 0; i < 10; i++ { // Try multiple times
			_, finalErr = iter.Next(ctx)
			if finalErr != nil {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}

		if finalErr == nil {
			t.Fatal("Expected some error after send failure and channel closure")
		}

		if finalErr != ErrNoMoreMessages {
			t.Errorf("Expected ErrNoMoreMessages when channels close, got %v", finalErr)
		}
	})

	// Test 4: Multiple successive errors
	t.Run("MultipleErrors", func(t *testing.T) {
		transport := &errorTransport{
			multipleErrors: []error{
				fmt.Errorf("first error"),
				fmt.Errorf("second error"),
				fmt.Errorf("third error"),
			},
		}

		iter, err := QueryWithTransport(ctx, "test", transport)
		if err != nil {
			t.Fatalf("Query should not fail immediately: %v", err)
		}
		defer iter.Close()

		// Should get first error
		_, err = iter.Next(ctx)
		if err == nil || !contains(err.Error(), "first error") {
			t.Errorf("Expected first error, got %v", err)
		}

		// Iterator should be closed after first error
		_, err = iter.Next(ctx)
		if err != ErrNoMoreMessages {
			t.Errorf("Expected ErrNoMoreMessages after error, got %v", err)
		}
	})
}

// errorTransport implements Transport interface for error testing
type errorTransport struct {
	connected      bool
	msgChan        chan Message
	errChan        chan error
	connectError   error
	sendError      error
	receiveError   error
	multipleErrors []error
	errorIndex     int
}

func (e *errorTransport) Connect(ctx context.Context) error {
	if e.connectError != nil {
		return e.connectError
	}

	e.connected = true
	e.msgChan = make(chan Message, 1)
	e.errChan = make(chan error, 1)

	// Handle multiple errors scenario
	if len(e.multipleErrors) > 0 {
		go func() {
			defer close(e.msgChan)
			defer close(e.errChan)
			if e.errorIndex < len(e.multipleErrors) {
				e.errChan <- e.multipleErrors[e.errorIndex]
				e.errorIndex++
			}
		}()
		return nil
	}

	// Handle receive error scenario
	if e.receiveError != nil {
		go func() {
			defer close(e.msgChan)
			defer close(e.errChan)
			e.errChan <- e.receiveError
		}()
		return nil
	}

	// Normal connection - just close channels for no messages
	go func() {
		defer close(e.msgChan)
		defer close(e.errChan)
		// Give a small delay to allow iterator to start
		time.Sleep(10 * time.Millisecond)
	}()

	return nil
}

func (e *errorTransport) SendMessage(ctx context.Context, message StreamMessage) error {
	if e.sendError != nil {
		return e.sendError
	}
	return nil
}

func (e *errorTransport) ReceiveMessages(ctx context.Context) (<-chan Message, <-chan error) {
	return e.msgChan, e.errChan
}

func (e *errorTransport) Interrupt(ctx context.Context) error {
	return nil
}

func (e *errorTransport) Close() error {
	e.connected = false
	return nil
}

// T125: Query Context Cancellation 🔴 RED
// Go Target: query_test.go::TestQueryContextCancellation
func TestQueryContextCancellation(t *testing.T) {
	// Test 1: Context timeout during query
	t.Run("TimeoutDuringQuery", func(t *testing.T) {
		// Very short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		transport := &slowTransport{delay: 200 * time.Millisecond}

		iter, err := QueryWithTransport(ctx, "slow query", transport)
		if err != nil {
			t.Fatalf("Query should not fail immediately: %v", err)
		}
		defer iter.Close()

		// Should timeout before getting any message
		_, err = iter.Next(ctx)
		if err == nil {
			t.Fatal("Expected timeout error")
		}

		// The error might be wrapped, so check if it contains the timeout
		if !isContextError(err, context.DeadlineExceeded) {
			t.Errorf("Expected context.DeadlineExceeded (or wrapped), got %v", err)
		}
	})

	// Test 2: Manual context cancellation
	t.Run("ManualCancellation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		transport := &slowTransport{delay: 500 * time.Millisecond}

		iter, err := QueryWithTransport(ctx, "slow query", transport)
		if err != nil {
			t.Fatalf("Query should not fail immediately: %v", err)
		}
		defer iter.Close()

		// Cancel after 100ms
		go func() {
			time.Sleep(100 * time.Millisecond)
			cancel()
		}()

		// Should be cancelled before getting any message
		_, err = iter.Next(ctx)
		if err == nil {
			t.Fatal("Expected cancellation error")
		}

		// The error might be wrapped, so check if it contains cancellation
		if !isContextError(err, context.Canceled) {
			t.Errorf("Expected context.Canceled (or wrapped), got %v", err)
		}
	})

	// Test 3: Context cancellation during streaming
	t.Run("StreamingCancellation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		transport := &slowTransport{delay: 200 * time.Millisecond}

		messages := make(chan StreamMessage, 2)
		messages <- StreamMessage{Type: "request", Message: &UserMessage{Content: "first"}}
		messages <- StreamMessage{Type: "request", Message: &UserMessage{Content: "second"}}
		close(messages)

		iter, err := QueryStreamWithTransport(ctx, messages, transport)
		if err != nil {
			t.Fatalf("QueryStream should not fail immediately: %v", err)
		}
		defer iter.Close()

		// Should timeout during message processing
		_, err = iter.Next(ctx)
		if err == nil {
			t.Fatal("Expected timeout error during streaming")
		}

		if !isContextError(err, context.DeadlineExceeded) {
			t.Errorf("Expected context.DeadlineExceeded (or wrapped) during streaming, got %v", err)
		}
	})

	// Test 4: Context propagation to transport
	t.Run("TransportContextPropagation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
		defer cancel()

		transport := &contextAwareTransport{}

		iter, err := QueryWithTransport(ctx, "test", transport)
		if err != nil {
			t.Fatalf("Query should not fail immediately: %v", err)
		}
		defer iter.Close()

		// Transport should receive context cancellation
		_, err = iter.Next(ctx)
		if err == nil {
			t.Fatal("Expected context cancellation")
		}

		// Give time for transport to detect cancellation
		time.Sleep(50 * time.Millisecond)

		// Verify transport received the context cancellation
		if !transport.IsContextCancelled() {
			t.Error("Expected transport to detect context cancellation")
		}
	})

	// Test 5: Multiple iterator calls after cancellation
	t.Run("MultipleCallsAfterCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		transport := &mockTransport{mockCLI: ""}

		iter, err := QueryWithTransport(ctx, "test", transport)
		if err != nil {
			t.Fatalf("Query should not fail immediately: %v", err)
		}
		defer iter.Close()

		// First call should return context.Canceled
		_, err = iter.Next(ctx)
		if !isContextError(err, context.Canceled) {
			t.Errorf("First call: expected context.Canceled (or wrapped), got %v", err)
		}

		// Subsequent calls should return ErrNoMoreMessages (iterator closed after first error)
		for i := 1; i < 3; i++ {
			_, err = iter.Next(ctx)
			if err != ErrNoMoreMessages {
				t.Errorf("Call %d: expected ErrNoMoreMessages, got %v", i+1, err)
			}
		}
	})
}

// slowTransport introduces delays for timeout testing
type slowTransport struct {
	delay     time.Duration
	connected bool
	msgChan   chan Message
	errChan   chan error
}

func (s *slowTransport) Connect(ctx context.Context) error {
	s.connected = true
	s.msgChan = make(chan Message, 1)
	s.errChan = make(chan error, 1)

	go func() {
		defer close(s.msgChan)
		defer close(s.errChan)

		// Respect context cancellation during delay
		select {
		case <-time.After(s.delay):
			// Send message after delay
			s.msgChan <- &AssistantMessage{
				Content: []ContentBlock{&TextBlock{Text: "Slow response"}},
				Model:   "claude-opus-4-1-20250805",
			}
		case <-ctx.Done():
			// Context was cancelled during delay
			s.errChan <- ctx.Err()
		}
	}()

	return nil
}

func (s *slowTransport) SendMessage(ctx context.Context, message StreamMessage) error {
	// Simulate slow send
	select {
	case <-time.After(s.delay):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *slowTransport) ReceiveMessages(ctx context.Context) (<-chan Message, <-chan error) {
	return s.msgChan, s.errChan
}

func (s *slowTransport) Interrupt(ctx context.Context) error {
	return nil
}

func (s *slowTransport) Close() error {
	s.connected = false
	return nil
}

// contextAwareTransport tracks context cancellation
type contextAwareTransport struct {
	mu               sync.RWMutex
	connected        bool
	contextCancelled bool
	msgChan          chan Message
	errChan          chan error
}

func (c *contextAwareTransport) Connect(ctx context.Context) error {
	c.mu.Lock()
	c.connected = true
	c.msgChan = make(chan Message, 1)
	c.errChan = make(chan error, 1)
	c.mu.Unlock()

	go func() {
		defer close(c.msgChan)
		defer close(c.errChan)

		select {
		case <-time.After(100 * time.Millisecond):
			// This shouldn't happen in our test
		case <-ctx.Done():
			c.mu.Lock()
			c.contextCancelled = true
			c.mu.Unlock()
			c.errChan <- ctx.Err()
		}
	}()

	return nil
}

func (c *contextAwareTransport) SendMessage(ctx context.Context, message StreamMessage) error {
	select {
	case <-ctx.Done():
		c.mu.Lock()
		c.contextCancelled = true
		c.mu.Unlock()
		return ctx.Err()
	default:
		return nil
	}
}

func (c *contextAwareTransport) ReceiveMessages(ctx context.Context) (<-chan Message, <-chan error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.msgChan, c.errChan
}

func (c *contextAwareTransport) Interrupt(ctx context.Context) error {
	return nil
}

func (c *contextAwareTransport) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.connected = false
	return nil
}

func (c *contextAwareTransport) IsContextCancelled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.contextCancelled
}

// Helper function to check if an error is or contains a specific context error
func isContextError(err error, target error) bool {
	if err == target {
		return true
	}

	// Check if the error message contains the target error
	if err != nil && target != nil {
		return contains(err.Error(), target.Error())
	}

	return false
}
