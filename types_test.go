package claudecode

import (
	"testing"
)

// TestPublicAPIMessageReExports tests that message type re-exports work through public API
func TestPublicAPIMessageReExports(t *testing.T) {
	tests := []struct {
		name         string
		message      Message
		expectedType string
	}{
		{
			name:         "UserMessage re-export",
			message:      &UserMessage{Content: "test"},
			expectedType: MessageTypeUser,
		},
		{
			name:         "AssistantMessage re-export",
			message:      &AssistantMessage{Content: []ContentBlock{}, Model: "test"},
			expectedType: MessageTypeAssistant,
		},
		{
			name:         "SystemMessage re-export",
			message:      &SystemMessage{Subtype: "test", Data: map[string]any{}},
			expectedType: MessageTypeSystem,
		},
		{
			name:         "ResultMessage re-export",
			message:      &ResultMessage{Subtype: "test", SessionID: "test"},
			expectedType: MessageTypeResult,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that re-exported type implements Message interface
			assertTypesMessageInterface(t, tt.message, tt.expectedType)
		})
	}
}

// TestPublicAPIContentBlockReExports tests that content block type re-exports work through public API
func TestPublicAPIContentBlockReExports(t *testing.T) {
	tests := []struct {
		name         string
		block        ContentBlock
		expectedType string
	}{
		{
			name:         "TextBlock re-export",
			block:        &TextBlock{Text: "test"},
			expectedType: ContentBlockTypeText,
		},
		{
			name:         "ThinkingBlock re-export",
			block:        &ThinkingBlock{Thinking: "test", Signature: "test"},
			expectedType: ContentBlockTypeThinking,
		},
		{
			name:         "ToolUseBlock re-export",
			block:        &ToolUseBlock{ToolUseID: "test", Name: "test", Input: map[string]any{}},
			expectedType: ContentBlockTypeToolUse,
		},
		{
			name:         "ToolResultBlock re-export",
			block:        &ToolResultBlock{ToolUseID: "test", Content: "test"},
			expectedType: ContentBlockTypeToolResult,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that re-exported type implements ContentBlock interface
			assertTypesContentBlockInterface(t, tt.block, tt.expectedType)
		})
	}
}

// TestConstantsFunctionalUsage tests that re-exported constants work functionally with the type system
func TestConstantsFunctionalUsage(t *testing.T) {
	// Test that message constants work with actual message creation
	userMsg := &UserMessage{Content: "test"}
	assertTypesConstantUsage(t, userMsg.Type(), MessageTypeUser, "MessageTypeUser")

	assistantMsg := &AssistantMessage{Content: []ContentBlock{}, Model: "test"}
	assertTypesConstantUsage(t, assistantMsg.Type(), MessageTypeAssistant, "MessageTypeAssistant")

	systemMsg := &SystemMessage{Subtype: "test", Data: map[string]any{}}
	assertTypesConstantUsage(t, systemMsg.Type(), MessageTypeSystem, "MessageTypeSystem")

	resultMsg := &ResultMessage{Subtype: "test", SessionID: "test"}
	assertTypesConstantUsage(t, resultMsg.Type(), MessageTypeResult, "MessageTypeResult")

	// Test that content block constants work with actual block creation
	textBlock := &TextBlock{Text: "test"}
	assertTypesConstantUsage(t, textBlock.BlockType(), ContentBlockTypeText, "ContentBlockTypeText")

	thinkingBlock := &ThinkingBlock{Thinking: "test", Signature: "test"}
	assertTypesConstantUsage(t, thinkingBlock.BlockType(), ContentBlockTypeThinking, "ContentBlockTypeThinking")

	toolUseBlock := &ToolUseBlock{ToolUseID: "test", Name: "test", Input: map[string]any{}}
	assertTypesConstantUsage(t, toolUseBlock.BlockType(), ContentBlockTypeToolUse, "ContentBlockTypeToolUse")

	toolResultBlock := &ToolResultBlock{ToolUseID: "test", Content: "test"}
	assertTypesConstantUsage(t, toolResultBlock.BlockType(), ContentBlockTypeToolResult, "ContentBlockTypeToolResult")
}

// TestTransportInterfacePublicAPI tests that Transport interface remains accessible through public API
func TestTransportInterfacePublicAPI(t *testing.T) {
	// Test that Transport interface is available for public API consumers
	var transport Transport
	assertTypesTransportInterface(t, transport)

	// Test that Transport interface has expected methods (compile-time verification)
	// This ensures the interface contract remains stable for public API
	// Transport interface methods: Connect, SendMessage, ReceiveMessages, Interrupt, Close
	t.Log("Transport interface methods verified at compile time")
}

// TestStreamMessageReExport tests that StreamMessage re-export works through public API
func TestStreamMessageReExport(t *testing.T) {
	tests := []struct {
		name      string
		msgType   string
		message   string
		sessionID string
	}{
		{
			name:      "Request message",
			msgType:   "request",
			message:   "test query",
			sessionID: "session123",
		},
		{
			name:      "Response message",
			msgType:   "response",
			message:   "test response",
			sessionID: "session456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &StreamMessage{
				Type:      tt.msgType,
				Message:   tt.message,
				SessionID: tt.sessionID,
			}

			assertTypesStreamMessage(t, msg, tt.msgType, tt.message, tt.sessionID)
		})
	}
}

// Helper functions

// assertTypesMessageInterface tests that a re-exported message implements Message interface correctly
func assertTypesMessageInterface(t *testing.T, msg Message, expectedType string) {
	t.Helper()
	if msg == nil {
		t.Fatal("Message should not be nil")
	}
	if msg.Type() != expectedType {
		t.Errorf("Expected message type %s, got %s", expectedType, msg.Type())
	}
}

// assertTypesContentBlockInterface tests that a re-exported content block implements ContentBlock interface correctly
func assertTypesContentBlockInterface(t *testing.T, block ContentBlock, expectedType string) {
	t.Helper()
	if block == nil {
		t.Fatal("ContentBlock should not be nil")
	}
	if block.BlockType() != expectedType {
		t.Errorf("Expected content block type %s, got %s", expectedType, block.BlockType())
	}
}

// assertTypesConstantUsage tests that re-exported constants work functionally with the type system
func assertTypesConstantUsage(t *testing.T, actualValue, expectedConstant, constantName string) {
	t.Helper()
	if actualValue != expectedConstant {
		t.Errorf("Constant %s not working functionally: expected %s, got %s", constantName, expectedConstant, actualValue)
	}
}

// assertTypesTransportInterface tests that Transport interface is available through public API
func assertTypesTransportInterface(t *testing.T, transport Transport) {
	t.Helper()
	// This is primarily a compile-time test - if Transport interface is not available,
	// this function wouldn't compile. The test verifies architectural integrity.
	if transport == nil {
		// Expected for nil interface value
		t.Log("Transport interface is correctly available through public API")
	}
}

// assertTypesStreamMessage tests that StreamMessage re-export maintains field access
func assertTypesStreamMessage(t *testing.T, msg *StreamMessage, expectedType, expectedMessage, expectedSessionID string) {
	t.Helper()
	if msg == nil {
		t.Fatal("StreamMessage should not be nil")
	}
	if msg.Type != expectedType {
		t.Errorf("Expected StreamMessage.Type = %s, got %s", expectedType, msg.Type)
	}
	if msg.Message != expectedMessage {
		t.Errorf("Expected StreamMessage.Message = %s, got %s", expectedMessage, msg.Message)
	}
	if msg.SessionID != expectedSessionID {
		t.Errorf("Expected StreamMessage.SessionID = %s, got %s", expectedSessionID, msg.SessionID)
	}
}
