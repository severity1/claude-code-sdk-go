package shared

import (
	"encoding/json"
	"testing"
)

// TestUserMessageCreation tests creating UserMessage with string content
func TestUserMessageCreation(t *testing.T) {
	content := "Hello, Claude!"
	msg := &UserMessage{
		Content: content,
	}

	// Test that Type() returns correct value
	if msg.Type() != MessageTypeUser {
		t.Errorf("Expected Type() = %q, got %q", MessageTypeUser, msg.Type())
	}

	// Test JSON marshaling includes type
	jsonData, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal UserMessage: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if result["type"] != MessageTypeUser {
		t.Errorf("Expected JSON type = %q, got %v", MessageTypeUser, result["type"])
	}

	if result["content"] != content {
		t.Errorf("Expected JSON content = %q, got %v", content, result["content"])
	}
}

// TestAssistantMessageWithText tests AssistantMessage with TextBlock content and model field
func TestAssistantMessageWithText(t *testing.T) {
	model := "claude-3-sonnet"
	textBlock := &TextBlock{
		Text: "Hello! How can I help you?",
	}

	msg := &AssistantMessage{
		Content: []ContentBlock{textBlock},
		Model:   model,
	}

	// Test Type() method
	if msg.Type() != MessageTypeAssistant {
		t.Errorf("Expected Type() = %q, got %q", MessageTypeAssistant, msg.Type())
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal AssistantMessage: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if result["type"] != MessageTypeAssistant {
		t.Errorf("Expected JSON type = %q, got %v", MessageTypeAssistant, result["type"])
	}

	if result["model"] != model {
		t.Errorf("Expected JSON model = %q, got %v", model, result["model"])
	}
}

// TestTextBlock tests TextBlock content type
func TestTextBlock(t *testing.T) {
	text := "This is a text block"
	block := &TextBlock{
		Text: text,
	}

	// Test BlockType() method
	if block.BlockType() != ContentBlockTypeText {
		t.Errorf("Expected BlockType() = %q, got %q", ContentBlockTypeText, block.BlockType())
	}

	// Test that it implements ContentBlock interface
	var cb ContentBlock = block
	if cb.BlockType() != ContentBlockTypeText {
		t.Errorf("Expected ContentBlock.BlockType() = %q, got %q", ContentBlockTypeText, cb.BlockType())
	}

	if block.Text != text {
		t.Errorf("Expected Text = %q, got %q", text, block.Text)
	}
}

// TestThinkingBlock tests ThinkingBlock content type
func TestThinkingBlock(t *testing.T) {
	thinking := "Let me think about this..."
	signature := "claude-3-sonnet-20240229"
	block := &ThinkingBlock{
		Thinking:  thinking,
		Signature: signature,
	}

	// Test BlockType() method
	if block.BlockType() != ContentBlockTypeThinking {
		t.Errorf("Expected BlockType() = %q, got %q", ContentBlockTypeThinking, block.BlockType())
	}

	// Test that it implements ContentBlock interface
	var cb ContentBlock = block
	if cb.BlockType() != ContentBlockTypeThinking {
		t.Errorf("Expected ContentBlock.BlockType() = %q, got %q", ContentBlockTypeThinking, cb.BlockType())
	}

	if block.Thinking != thinking {
		t.Errorf("Expected Thinking = %q, got %q", thinking, block.Thinking)
	}

	if block.Signature != signature {
		t.Errorf("Expected Signature = %q, got %q", signature, block.Signature)
	}
}

// TestSystemMessage tests SystemMessage with subtype and data preservation
func TestSystemMessage(t *testing.T) {
	subtype := "user_confirmation"
	data := map[string]any{
		"type":     MessageTypeSystem,
		"subtype":  subtype,
		"question": "Do you want to proceed?",
		"options":  []string{"yes", "no"},
	}

	msg := &SystemMessage{
		Subtype: subtype,
		Data:    data,
	}

	// Test Type() method
	if msg.Type() != MessageTypeSystem {
		t.Errorf("Expected Type() = %q, got %q", MessageTypeSystem, msg.Type())
	}

	// Test JSON marshaling preserves all data
	jsonData, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal SystemMessage: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(jsonData, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if result["type"] != MessageTypeSystem {
		t.Errorf("Expected JSON type = %q, got %v", MessageTypeSystem, result["type"])
	}

	if result["subtype"] != subtype {
		t.Errorf("Expected JSON subtype = %q, got %v", subtype, result["subtype"])
	}

	// Check that original data fields are preserved
	if result["question"] != "Do you want to proceed?" {
		t.Errorf("Expected question field to be preserved")
	}
}

// TestMessageInterface tests all message types implement Message interface
func TestMessageInterface(t *testing.T) {
	// Test all message types implement Message interface correctly
	var messages []Message = []Message{
		&UserMessage{Content: "test"},
		&AssistantMessage{Content: []ContentBlock{}, Model: "test"},
		&SystemMessage{Subtype: "test", Data: map[string]any{}},
		&ResultMessage{Subtype: "completion", SessionID: "test"},
	}

	expectedTypes := []string{
		MessageTypeUser,
		MessageTypeAssistant,
		MessageTypeSystem,
		MessageTypeResult,
	}

	for i, msg := range messages {
		if msg.Type() != expectedTypes[i] {
			t.Errorf("Message %d: expected type %q, got %q", i, expectedTypes[i], msg.Type())
		}
	}
}

// TestContentBlockInterface tests all content blocks implement ContentBlock interface
func TestContentBlockInterface(t *testing.T) {
	// Test all content block types implement ContentBlock interface correctly
	var blocks []ContentBlock = []ContentBlock{
		&TextBlock{Text: "test"},
		&ThinkingBlock{Thinking: "test", Signature: "test"},
		&ToolUseBlock{ToolUseID: "test", Name: "test", Input: map[string]any{}},
		&ToolResultBlock{ToolUseID: "test", Content: "test"},
	}

	expectedTypes := []string{
		ContentBlockTypeText,
		ContentBlockTypeThinking,
		ContentBlockTypeToolUse,
		ContentBlockTypeToolResult,
	}

	for i, block := range blocks {
		if block.BlockType() != expectedTypes[i] {
			t.Errorf("Block %d: expected type %q, got %q", i, expectedTypes[i], block.BlockType())
		}
	}
}

// TestMessageTypeConstants tests message type string constants
func TestMessageTypeConstants(t *testing.T) {
	// Test that all message type constants have expected values
	expectedTypes := map[string]string{
		MessageTypeUser:      "user",
		MessageTypeAssistant: "assistant",
		MessageTypeSystem:    "system",
		MessageTypeResult:    "result",
	}

	for constant, expectedValue := range expectedTypes {
		if constant != expectedValue {
			t.Errorf("Expected message type constant %q to equal %q, got %q", constant, expectedValue, constant)
		}
	}

	// Test content block type constants
	expectedBlockTypes := map[string]string{
		ContentBlockTypeText:       "text",
		ContentBlockTypeThinking:   "thinking",
		ContentBlockTypeToolUse:    "tool_use",
		ContentBlockTypeToolResult: "tool_result",
	}

	for constant, expectedValue := range expectedBlockTypes {
		if constant != expectedValue {
			t.Errorf("Expected content block type constant %q to equal %q, got %q", constant, expectedValue, constant)
		}
	}
}

// TestToolUseBlockCreation tests ToolUseBlock with ID, name, and input parameters
func TestToolUseBlockCreation(t *testing.T) {
	toolUseID := "tool_456"
	name := "Read"
	input := map[string]any{
		"file_path": "/home/user/document.txt",
		"limit":     100,
	}

	block := &ToolUseBlock{
		ToolUseID: toolUseID,
		Name:      name,
		Input:     input,
	}

	// Test BlockType() method
	if block.BlockType() != ContentBlockTypeToolUse {
		t.Errorf("Expected BlockType() = %q, got %q", ContentBlockTypeToolUse, block.BlockType())
	}

	// Test fields
	if block.ToolUseID != toolUseID {
		t.Errorf("Expected ToolUseID = %q, got %q", toolUseID, block.ToolUseID)
	}
	if block.Name != name {
		t.Errorf("Expected Name = %q, got %q", name, block.Name)
	}

	// Test that it implements ContentBlock interface
	var cb ContentBlock = block
	if cb.BlockType() != ContentBlockTypeToolUse {
		t.Errorf("Expected ContentBlock.BlockType() = %q, got %q", ContentBlockTypeToolUse, cb.BlockType())
	}
}

// TestToolResultBlockCreation tests ToolResultBlock with content and error flag
func TestToolResultBlockCreation(t *testing.T) {
	toolUseID := "tool_456"
	content := "File content here..."
	isError := false

	block := &ToolResultBlock{
		ToolUseID: toolUseID,
		Content:   content,
		IsError:   &isError,
	}

	// Test BlockType() method
	if block.BlockType() != ContentBlockTypeToolResult {
		t.Errorf("Expected BlockType() = %q, got %q", ContentBlockTypeToolResult, block.BlockType())
	}

	// Test fields
	if block.ToolUseID != toolUseID {
		t.Errorf("Expected ToolUseID = %q, got %q", toolUseID, block.ToolUseID)
	}
	if block.Content != content {
		t.Errorf("Expected Content = %q, got %v", content, block.Content)
	}
	if block.IsError == nil || *block.IsError != isError {
		t.Errorf("Expected IsError = %v, got %v", isError, block.IsError)
	}

	// Test that it implements ContentBlock interface
	var cb ContentBlock = block
	if cb.BlockType() != ContentBlockTypeToolResult {
		t.Errorf("Expected ContentBlock.BlockType() = %q, got %q", ContentBlockTypeToolResult, cb.BlockType())
	}
}