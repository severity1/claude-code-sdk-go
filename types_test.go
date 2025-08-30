package claudecode

import (
	"testing"
)

// TestMessageTypesReExport tests that message type re-exports work properly
func TestMessageTypesReExport(t *testing.T) {
	// Test UserMessage re-export
	userMsg := &UserMessage{Content: "test"}
	if userMsg.Type() != MessageTypeUser {
		t.Errorf("Expected UserMessage Type() = %s, got %s", MessageTypeUser, userMsg.Type())
	}

	// Test AssistantMessage re-export  
	assistantMsg := &AssistantMessage{Content: []ContentBlock{}, Model: "test"}
	if assistantMsg.Type() != MessageTypeAssistant {
		t.Errorf("Expected AssistantMessage Type() = %s, got %s", MessageTypeAssistant, assistantMsg.Type())
	}

	// Test SystemMessage re-export
	systemMsg := &SystemMessage{Subtype: "test", Data: map[string]any{}}
	if systemMsg.Type() != MessageTypeSystem {
		t.Errorf("Expected SystemMessage Type() = %s, got %s", MessageTypeSystem, systemMsg.Type())
	}

	// Test that all implement Message interface
	var messages []Message = []Message{userMsg, assistantMsg, systemMsg}
	expectedTypes := []string{MessageTypeUser, MessageTypeAssistant, MessageTypeSystem}

	for i, msg := range messages {
		if msg.Type() != expectedTypes[i] {
			t.Errorf("Message %d: expected type %s, got %s", i, expectedTypes[i], msg.Type())
		}
	}
}

// TestContentBlockTypesReExport tests that content block type re-exports work
func TestContentBlockTypesReExport(t *testing.T) {
	// Test TextBlock re-export
	textBlock := &TextBlock{Text: "test"}
	if textBlock.BlockType() != ContentBlockTypeText {
		t.Errorf("Expected TextBlock BlockType() = %s, got %s", ContentBlockTypeText, textBlock.BlockType())
	}

	// Test ThinkingBlock re-export
	thinkingBlock := &ThinkingBlock{Thinking: "test", Signature: "test"}
	if thinkingBlock.BlockType() != ContentBlockTypeThinking {
		t.Errorf("Expected ThinkingBlock BlockType() = %s, got %s", ContentBlockTypeThinking, thinkingBlock.BlockType())
	}

	// Test ToolUseBlock re-export
	toolUseBlock := &ToolUseBlock{ToolUseID: "test", Name: "test", Input: map[string]any{}}
	if toolUseBlock.BlockType() != ContentBlockTypeToolUse {
		t.Errorf("Expected ToolUseBlock BlockType() = %s, got %s", ContentBlockTypeToolUse, toolUseBlock.BlockType())
	}

	// Test ToolResultBlock re-export
	toolResultBlock := &ToolResultBlock{ToolUseID: "test", Content: "test"}
	if toolResultBlock.BlockType() != ContentBlockTypeToolResult {
		t.Errorf("Expected ToolResultBlock BlockType() = %s, got %s", ContentBlockTypeToolResult, toolResultBlock.BlockType())
	}

	// Test that all implement ContentBlock interface
	var blocks []ContentBlock = []ContentBlock{textBlock, thinkingBlock, toolUseBlock, toolResultBlock}
	expectedTypes := []string{ContentBlockTypeText, ContentBlockTypeThinking, ContentBlockTypeToolUse, ContentBlockTypeToolResult}

	for i, block := range blocks {
		if block.BlockType() != expectedTypes[i] {
			t.Errorf("Block %d: expected type %s, got %s", i, expectedTypes[i], block.BlockType())
		}
	}
}

// TestConstantsReExport tests that constants are re-exported correctly
func TestConstantsReExport(t *testing.T) {
	// Test message type constants
	if MessageTypeUser != "user" {
		t.Errorf("Expected MessageTypeUser = 'user', got %s", MessageTypeUser)
	}
	if MessageTypeAssistant != "assistant" {
		t.Errorf("Expected MessageTypeAssistant = 'assistant', got %s", MessageTypeAssistant)
	}
	if MessageTypeSystem != "system" {
		t.Errorf("Expected MessageTypeSystem = 'system', got %s", MessageTypeSystem)
	}
	if MessageTypeResult != "result" {
		t.Errorf("Expected MessageTypeResult = 'result', got %s", MessageTypeResult)
	}

	// Test content block type constants
	if ContentBlockTypeText != "text" {
		t.Errorf("Expected ContentBlockTypeText = 'text', got %s", ContentBlockTypeText)
	}
	if ContentBlockTypeThinking != "thinking" {
		t.Errorf("Expected ContentBlockTypeThinking = 'thinking', got %s", ContentBlockTypeThinking)
	}
	if ContentBlockTypeToolUse != "tool_use" {
		t.Errorf("Expected ContentBlockTypeToolUse = 'tool_use', got %s", ContentBlockTypeToolUse)
	}
	if ContentBlockTypeToolResult != "tool_result" {
		t.Errorf("Expected ContentBlockTypeToolResult = 'tool_result', got %s", ContentBlockTypeToolResult)
	}
}

// TestTransportInterfaceStaysLocal tests that Transport interface is still in main package
func TestTransportInterfaceStaysLocal(t *testing.T) {
	// This is a compile-time test - if Transport interface is available, this will compile
	var transport Transport
	if transport != nil {
		// This won't execute, just testing that interface is available
		t.Log("Transport interface is available in main package")
	}
}

// TestStreamMessageReExport tests that StreamMessage is re-exported correctly
func TestStreamMessageReExport(t *testing.T) {
	msg := &StreamMessage{
		Type:      "request",
		Message:   "test",
		SessionID: "session123",
	}

	if msg.Type != "request" {
		t.Errorf("Expected StreamMessage Type = 'request', got %s", msg.Type)
	}
	if msg.SessionID != "session123" {
		t.Errorf("Expected StreamMessage SessionID = 'session123', got %s", msg.SessionID)
	}
}