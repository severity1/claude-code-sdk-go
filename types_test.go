package claudecode

import (
	"encoding/json"
	"testing"
)

// T001: User Message Creation
func TestUserMessageCreation(t *testing.T) {
	// Test UserMessage with string content
	msg := &UserMessage{
		Content: "Hello Claude",
	}

	if msg.Type() != MessageTypeUser {
		t.Errorf("Expected Type() = %q, got %q", MessageTypeUser, msg.Type())
	}

	// Test JSON marshaling
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal UserMessage: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if parsed["type"] != MessageTypeUser {
		t.Errorf("Expected JSON type = %q, got %q", MessageTypeUser, parsed["type"])
	}

	if parsed["content"] != "Hello Claude" {
		t.Errorf("Expected JSON content = %q, got %q", "Hello Claude", parsed["content"])
	}
}

// T002: Assistant Message with Text
func TestAssistantMessageWithText(t *testing.T) {
	msg := &AssistantMessage{
		Content: []ContentBlock{
			&TextBlock{Text: "Hello there!"},
		},
		Model: "claude-opus-4-1-20250805",
	}

	if msg.Type() != MessageTypeAssistant {
		t.Errorf("Expected Type() = %q, got %q", MessageTypeAssistant, msg.Type())
	}

	if msg.Model != "claude-opus-4-1-20250805" {
		t.Errorf("Expected Model = %q, got %q", "claude-opus-4-1-20250805", msg.Model)
	}

	if len(msg.Content) != 1 {
		t.Errorf("Expected 1 content block, got %d", len(msg.Content))
	}

	textBlock, ok := msg.Content[0].(*TextBlock)
	if !ok {
		t.Errorf("Expected TextBlock, got %T", msg.Content[0])
	}

	if textBlock.Text != "Hello there!" {
		t.Errorf("Expected text = %q, got %q", "Hello there!", textBlock.Text)
	}
}

// T003: Assistant Message with Thinking
func TestAssistantMessageWithThinking(t *testing.T) {
	msg := &AssistantMessage{
		Content: []ContentBlock{
			&ThinkingBlock{
				Thinking:  "Let me think about this...",
				Signature: "claude-opus-4",
			},
		},
		Model: "claude-opus-4-1-20250805",
	}

	if msg.Type() != MessageTypeAssistant {
		t.Errorf("Expected Type() = %q, got %q", MessageTypeAssistant, msg.Type())
	}

	thinkingBlock, ok := msg.Content[0].(*ThinkingBlock)
	if !ok {
		t.Errorf("Expected ThinkingBlock, got %T", msg.Content[0])
	}

	if thinkingBlock.Thinking != "Let me think about this..." {
		t.Errorf("Expected thinking = %q, got %q", "Let me think about this...", thinkingBlock.Thinking)
	}

	if thinkingBlock.Signature != "claude-opus-4" {
		t.Errorf("Expected signature = %q, got %q", "claude-opus-4", thinkingBlock.Signature)
	}
}

// T004: Tool Use Block Creation
func TestToolUseBlockCreation(t *testing.T) {
	block := &ToolUseBlock{
		ID:   "toolu_123",
		Name: "calculate",
		Input: map[string]any{
			"expression": "2 + 2",
		},
	}

	if block.BlockType() != ContentBlockTypeToolUse {
		t.Errorf("Expected BlockType() = %q, got %q", ContentBlockTypeToolUse, block.BlockType())
	}

	if block.ID != "toolu_123" {
		t.Errorf("Expected ID = %q, got %q", "toolu_123", block.ID)
	}

	if block.Name != "calculate" {
		t.Errorf("Expected Name = %q, got %q", "calculate", block.Name)
	}

	expr, ok := block.Input["expression"].(string)
	if !ok || expr != "2 + 2" {
		t.Errorf("Expected Input.expression = %q, got %v", "2 + 2", block.Input["expression"])
	}
}

// T005: Tool Result Block Creation
func TestToolResultBlockCreation(t *testing.T) {
	content := "4"
	isError := false

	block := &ToolResultBlock{
		ToolUseID: "toolu_123",
		Content:   &content,
		IsError:   &isError,
	}

	if block.BlockType() != ContentBlockTypeToolResult {
		t.Errorf("Expected BlockType() = %q, got %q", ContentBlockTypeToolResult, block.BlockType())
	}

	if block.ToolUseID != "toolu_123" {
		t.Errorf("Expected ToolUseID = %q, got %q", "toolu_123", block.ToolUseID)
	}

	contentStr, ok := block.Content.(*string)
	if !ok || *contentStr != "4" {
		t.Errorf("Expected Content = %q, got %v", "4", block.Content)
	}

	if *block.IsError != false {
		t.Errorf("Expected IsError = false, got %v", *block.IsError)
	}
}

// T006: Result Message Creation
func TestResultMessageCreation(t *testing.T) {
	totalCost := 0.05
	result := "Calculation completed"

	msg := &ResultMessage{
		Subtype:       "query_completed",
		DurationMs:    1500,
		DurationAPIMs: 800,
		IsError:       false,
		NumTurns:      2,
		SessionID:     "session_123",
		TotalCostUSD:  &totalCost,
		Result:        &result,
	}

	if msg.Type() != MessageTypeResult {
		t.Errorf("Expected Type() = %q, got %q", MessageTypeResult, msg.Type())
	}

	if msg.Subtype != "query_completed" {
		t.Errorf("Expected Subtype = %q, got %q", "query_completed", msg.Subtype)
	}

	if msg.DurationMs != 1500 {
		t.Errorf("Expected DurationMs = 1500, got %d", msg.DurationMs)
	}

	if *msg.TotalCostUSD != 0.05 {
		t.Errorf("Expected TotalCostUSD = 0.05, got %v", *msg.TotalCostUSD)
	}
}

// T007: Text Block Implementation
func TestTextBlock(t *testing.T) {
	block := &TextBlock{Text: "This is text"}

	if block.BlockType() != ContentBlockTypeText {
		t.Errorf("Expected BlockType() = %q, got %q", ContentBlockTypeText, block.BlockType())
	}

	if block.Text != "This is text" {
		t.Errorf("Expected Text = %q, got %q", "This is text", block.Text)
	}

	// Test JSON marshaling
	data, err := json.Marshal(block)
	if err != nil {
		t.Fatalf("Failed to marshal TextBlock: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if parsed["type"] != ContentBlockTypeText {
		t.Errorf("Expected JSON type = %q, got %q", ContentBlockTypeText, parsed["type"])
	}
}

// T008: Thinking Block Implementation
func TestThinkingBlock(t *testing.T) {
	block := &ThinkingBlock{
		Thinking:  "I need to analyze this carefully",
		Signature: "claude-opus-4",
	}

	if block.BlockType() != ContentBlockTypeThinking {
		t.Errorf("Expected BlockType() = %q, got %q", ContentBlockTypeThinking, block.BlockType())
	}

	// Test JSON marshaling
	data, err := json.Marshal(block)
	if err != nil {
		t.Fatalf("Failed to marshal ThinkingBlock: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if parsed["type"] != ContentBlockTypeThinking {
		t.Errorf("Expected JSON type = %q, got %q", ContentBlockTypeThinking, parsed["type"])
	}
}

// T009: System Message Implementation
func TestSystemMessage(t *testing.T) {
	data := map[string]any{
		"type":    MessageTypeSystem,
		"subtype": "session_start",
		"extra":   "some data",
		"number":  42,
	}

	msg := &SystemMessage{
		Subtype: "session_start",
		Data:    data,
	}

	if msg.Type() != MessageTypeSystem {
		t.Errorf("Expected Type() = %q, got %q", MessageTypeSystem, msg.Type())
	}

	if msg.Subtype != "session_start" {
		t.Errorf("Expected Subtype = %q, got %q", "session_start", msg.Subtype)
	}

	// Verify data preservation
	if msg.Data["extra"] != "some data" {
		t.Errorf("Expected Data.extra = %q, got %v", "some data", msg.Data["extra"])
	}

	if msg.Data["number"] != 42 {
		t.Errorf("Expected Data.number = 42, got %v", msg.Data["number"])
	}
}

// T010: User Message Mixed Content
func TestUserMessageMixedContent(t *testing.T) {
	msg := &UserMessage{
		Content: []ContentBlock{
			&TextBlock{Text: "Please calculate"},
			&ToolUseBlock{
				ID:    "tool_1",
				Name:  "calc",
				Input: map[string]any{"expr": "5+5"},
			},
		},
	}

	contentBlocks, ok := msg.Content.([]ContentBlock)
	if !ok {
		t.Errorf("Expected Content to be []ContentBlock, got %T", msg.Content)
	}

	if len(contentBlocks) != 2 {
		t.Errorf("Expected 2 content blocks, got %d", len(contentBlocks))
	}

	// Verify first block is TextBlock
	_, ok = contentBlocks[0].(*TextBlock)
	if !ok {
		t.Errorf("Expected first block to be TextBlock, got %T", contentBlocks[0])
	}

	// Verify second block is ToolUseBlock
	_, ok = contentBlocks[1].(*ToolUseBlock)
	if !ok {
		t.Errorf("Expected second block to be ToolUseBlock, got %T", contentBlocks[1])
	}
}

// T011: Assistant Message Mixed Content
func TestAssistantMessageMixedContent(t *testing.T) {
	msg := &AssistantMessage{
		Content: []ContentBlock{
			&ThinkingBlock{
				Thinking:  "Let me calculate this",
				Signature: "claude-opus-4",
			},
			&TextBlock{Text: "The answer is:"},
			&ToolUseBlock{
				ID:    "tool_2",
				Name:  "calculate",
				Input: map[string]any{"expression": "10 * 5"},
			},
		},
		Model: "claude-opus-4-1-20250805",
	}

	if len(msg.Content) != 3 {
		t.Errorf("Expected 3 content blocks, got %d", len(msg.Content))
	}

	// Verify block types
	_, ok := msg.Content[0].(*ThinkingBlock)
	if !ok {
		t.Errorf("Expected first block to be ThinkingBlock, got %T", msg.Content[0])
	}

	_, ok = msg.Content[1].(*TextBlock)
	if !ok {
		t.Errorf("Expected second block to be TextBlock, got %T", msg.Content[1])
	}

	_, ok = msg.Content[2].(*ToolUseBlock)
	if !ok {
		t.Errorf("Expected third block to be ToolUseBlock, got %T", msg.Content[2])
	}
}

// T012: Message Interface Compliance
func TestMessageInterface(t *testing.T) {
	messages := []Message{
		&UserMessage{Content: "test"},
		&AssistantMessage{Content: []ContentBlock{}, Model: "claude"},
		&SystemMessage{Subtype: "test", Data: make(map[string]any)},
		&ResultMessage{Subtype: "test", DurationMs: 100, DurationAPIMs: 50, IsError: false, NumTurns: 1, SessionID: "s1"},
	}

	expectedTypes := []string{
		MessageTypeUser,
		MessageTypeAssistant,
		MessageTypeSystem,
		MessageTypeResult,
	}

	for i, msg := range messages {
		if msg.Type() != expectedTypes[i] {
			t.Errorf("Message %d: expected Type() = %q, got %q", i, expectedTypes[i], msg.Type())
		}
	}
}

// T013: Content Block Interface Compliance
func TestContentBlockInterface(t *testing.T) {
	blocks := []ContentBlock{
		&TextBlock{Text: "test"},
		&ThinkingBlock{Thinking: "test", Signature: "test"},
		&ToolUseBlock{ID: "1", Name: "test", Input: make(map[string]any)},
		&ToolResultBlock{ToolUseID: "1"},
	}

	expectedTypes := []string{
		ContentBlockTypeText,
		ContentBlockTypeThinking,
		ContentBlockTypeToolUse,
		ContentBlockTypeToolResult,
	}

	for i, block := range blocks {
		if block.BlockType() != expectedTypes[i] {
			t.Errorf("Block %d: expected BlockType() = %q, got %q", i, expectedTypes[i], block.BlockType())
		}
	}
}

// T014: Message Type Constants
func TestMessageTypeConstants(t *testing.T) {
	// Verify message type constants match Python SDK
	if MessageTypeUser != "user" {
		t.Errorf("Expected MessageTypeUser = %q, got %q", "user", MessageTypeUser)
	}

	if MessageTypeAssistant != "assistant" {
		t.Errorf("Expected MessageTypeAssistant = %q, got %q", "assistant", MessageTypeAssistant)
	}

	if MessageTypeSystem != "system" {
		t.Errorf("Expected MessageTypeSystem = %q, got %q", "system", MessageTypeSystem)
	}

	if MessageTypeResult != "result" {
		t.Errorf("Expected MessageTypeResult = %q, got %q", "result", MessageTypeResult)
	}

	// Verify content block type constants match Python SDK
	if ContentBlockTypeText != "text" {
		t.Errorf("Expected ContentBlockTypeText = %q, got %q", "text", ContentBlockTypeText)
	}

	if ContentBlockTypeThinking != "thinking" {
		t.Errorf("Expected ContentBlockTypeThinking = %q, got %q", "thinking", ContentBlockTypeThinking)
	}

	if ContentBlockTypeToolUse != "tool_use" {
		t.Errorf("Expected ContentBlockTypeToolUse = %q, got %q", "tool_use", ContentBlockTypeToolUse)
	}

	if ContentBlockTypeToolResult != "tool_result" {
		t.Errorf("Expected ContentBlockTypeToolResult = %q, got %q", "tool_result", ContentBlockTypeToolResult)
	}
}

// Additional test: JSON unmarshaling of UserMessage with string content
func TestUserMessageUnmarshalString(t *testing.T) {
	jsonData := `{"type": "user", "content": "Hello world"}`

	msg, err := UnmarshalMessage([]byte(jsonData))
	if err != nil {
		t.Fatalf("Failed to unmarshal UserMessage: %v", err)
	}

	userMsg, ok := msg.(*UserMessage)
	if !ok {
		t.Fatalf("Expected *UserMessage, got %T", msg)
	}

	content, ok := userMsg.Content.(string)
	if !ok {
		t.Fatalf("Expected string content, got %T", userMsg.Content)
	}

	if content != "Hello world" {
		t.Errorf("Expected content = %q, got %q", "Hello world", content)
	}
}

// Additional test: JSON unmarshaling of UserMessage with ContentBlock array
func TestUserMessageUnmarshalBlocks(t *testing.T) {
	jsonData := `{
		"type": "user",
		"content": [
			{"type": "text", "text": "Hello"},
			{"type": "tool_use", "id": "tool1", "name": "calc", "input": {"x": 5}}
		]
	}`

	msg, err := UnmarshalMessage([]byte(jsonData))
	if err != nil {
		t.Fatalf("Failed to unmarshal UserMessage: %v", err)
	}

	userMsg, ok := msg.(*UserMessage)
	if !ok {
		t.Fatalf("Expected *UserMessage, got %T", msg)
	}

	blocks, ok := userMsg.Content.([]ContentBlock)
	if !ok {
		t.Fatalf("Expected []ContentBlock content, got %T", userMsg.Content)
	}

	if len(blocks) != 2 {
		t.Errorf("Expected 2 blocks, got %d", len(blocks))
	}

	textBlock, ok := blocks[0].(*TextBlock)
	if !ok || textBlock.Text != "Hello" {
		t.Errorf("Expected first block to be TextBlock with 'Hello', got %T: %v", blocks[0], blocks[0])
	}

	toolBlock, ok := blocks[1].(*ToolUseBlock)
	if !ok || toolBlock.Name != "calc" {
		t.Errorf("Expected second block to be ToolUseBlock with name 'calc', got %T: %v", blocks[1], blocks[1])
	}
}
