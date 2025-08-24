package parser

import (
	"strings"
	"testing"

	"github.com/severity1/claude-code-sdk-go"
)

// TestParseValidUserMessage tests T035: Parse Valid User Message with text content
func TestParseValidUserMessage(t *testing.T) {
	parser := New()
	data := map[string]any{
		"type": "user",
		"message": map[string]any{
			"content": []any{
				map[string]any{"type": "text", "text": "Hello"},
			},
		},
	}

	message, err := parser.ParseMessage(data)
	if err != nil {
		t.Fatalf("Failed to parse valid user message: %v", err)
	}

	userMsg, ok := message.(*claudecode.UserMessage)
	if !ok {
		t.Fatalf("Expected UserMessage, got %T", message)
	}

	blocks, ok := userMsg.Content.([]claudecode.ContentBlock)
	if !ok {
		t.Fatalf("Expected Content to be []ContentBlock, got %T", userMsg.Content)
	}

	if len(blocks) != 1 {
		t.Fatalf("Expected 1 content block, got %d", len(blocks))
	}

	textBlock, ok := blocks[0].(*claudecode.TextBlock)
	if !ok {
		t.Fatalf("Expected TextBlock, got %T", blocks[0])
	}

	if textBlock.Text != "Hello" {
		t.Errorf("Expected text 'Hello', got '%s'", textBlock.Text)
	}
}

// TestParseUserMessageWithToolUse tests T036: Parse User Message with Tool Use content block
func TestParseUserMessageWithToolUse(t *testing.T) {
	parser := New()
	data := map[string]any{
		"type": "user",
		"message": map[string]any{
			"content": []any{
				map[string]any{"type": "text", "text": "Let me read this file"},
				map[string]any{
					"type":  "tool_use",
					"id":    "tool_456",
					"name":  "Read",
					"input": map[string]any{"file_path": "/example.txt"},
				},
			},
		},
	}

	message, err := parser.ParseMessage(data)
	if err != nil {
		t.Fatalf("Failed to parse user message with tool use: %v", err)
	}

	userMsg, ok := message.(*claudecode.UserMessage)
	if !ok {
		t.Fatalf("Expected UserMessage, got %T", message)
	}

	blocks, ok := userMsg.Content.([]claudecode.ContentBlock)
	if !ok {
		t.Fatalf("Expected Content to be []ContentBlock, got %T", userMsg.Content)
	}

	if len(blocks) != 2 {
		t.Fatalf("Expected 2 content blocks, got %d", len(blocks))
	}

	// Check text block
	textBlock, ok := blocks[0].(*claudecode.TextBlock)
	if !ok {
		t.Fatalf("Expected first block to be TextBlock, got %T", blocks[0])
	}
	if textBlock.Text != "Let me read this file" {
		t.Errorf("Expected text 'Let me read this file', got '%s'", textBlock.Text)
	}

	// Check tool use block
	toolUseBlock, ok := blocks[1].(*claudecode.ToolUseBlock)
	if !ok {
		t.Fatalf("Expected second block to be ToolUseBlock, got %T", blocks[1])
	}
	if toolUseBlock.ID != "tool_456" {
		t.Errorf("Expected tool use ID 'tool_456', got '%s'", toolUseBlock.ID)
	}
	if toolUseBlock.Name != "Read" {
		t.Errorf("Expected tool name 'Read', got '%s'", toolUseBlock.Name)
	}

	filePath, ok := toolUseBlock.Input["file_path"].(string)
	if !ok || filePath != "/example.txt" {
		t.Errorf("Expected input file_path '/example.txt', got %v", toolUseBlock.Input["file_path"])
	}
}

// TestParseUserMessageWithToolResult tests T037: Parse User Message with Tool Result content block
func TestParseUserMessageWithToolResult(t *testing.T) {
	parser := New()
	data := map[string]any{
		"type": "user",
		"message": map[string]any{
			"content": []any{
				map[string]any{
					"type":        "tool_result",
					"tool_use_id": "tool_456",
					"content":     "File content here",
				},
			},
		},
	}

	message, err := parser.ParseMessage(data)
	if err != nil {
		t.Fatalf("Failed to parse user message with tool result: %v", err)
	}

	userMsg, ok := message.(*claudecode.UserMessage)
	if !ok {
		t.Fatalf("Expected UserMessage, got %T", message)
	}

	blocks, ok := userMsg.Content.([]claudecode.ContentBlock)
	if !ok {
		t.Fatalf("Expected Content to be []ContentBlock, got %T", userMsg.Content)
	}

	if len(blocks) != 1 {
		t.Fatalf("Expected 1 content block, got %d", len(blocks))
	}

	toolResultBlock, ok := blocks[0].(*claudecode.ToolResultBlock)
	if !ok {
		t.Fatalf("Expected ToolResultBlock, got %T", blocks[0])
	}

	if toolResultBlock.ToolUseID != "tool_456" {
		t.Errorf("Expected tool_use_id 'tool_456', got '%s'", toolResultBlock.ToolUseID)
	}

	content, ok := toolResultBlock.Content.(string)
	if !ok || content != "File content here" {
		t.Errorf("Expected content 'File content here', got %v", toolResultBlock.Content)
	}
}

// TestParseUserMessageWithToolResultError tests T038: Parse User Message with Tool Result Error
func TestParseUserMessageWithToolResultError(t *testing.T) {
	parser := New()
	data := map[string]any{
		"type": "user",
		"message": map[string]any{
			"content": []any{
				map[string]any{
					"type":        "tool_result",
					"tool_use_id": "tool_456",
					"content":     "Error: File not found",
					"is_error":    true,
				},
			},
		},
	}

	message, err := parser.ParseMessage(data)
	if err != nil {
		t.Fatalf("Failed to parse user message with tool result error: %v", err)
	}

	userMsg, ok := message.(*claudecode.UserMessage)
	if !ok {
		t.Fatalf("Expected UserMessage, got %T", message)
	}

	blocks, ok := userMsg.Content.([]claudecode.ContentBlock)
	if !ok {
		t.Fatalf("Expected Content to be []ContentBlock, got %T", userMsg.Content)
	}

	toolResultBlock, ok := blocks[0].(*claudecode.ToolResultBlock)
	if !ok {
		t.Fatalf("Expected ToolResultBlock, got %T", blocks[0])
	}

	if toolResultBlock.IsError == nil || !*toolResultBlock.IsError {
		t.Errorf("Expected is_error to be true, got %v", toolResultBlock.IsError)
	}
}

// TestParseUserMessageMixedContent tests T039: Parse User Message with Mixed Content
func TestParseUserMessageMixedContent(t *testing.T) {
	parser := New()
	data := map[string]any{
		"type": "user",
		"message": map[string]any{
			"content": []any{
				map[string]any{"type": "text", "text": "First analyze this:"},
				map[string]any{
					"type":  "tool_use",
					"id":    "tool_1",
					"name":  "Read",
					"input": map[string]any{"file_path": "/data.txt"},
				},
				map[string]any{
					"type":        "tool_result",
					"tool_use_id": "tool_1",
					"content":     "Data: 42",
				},
				map[string]any{"type": "text", "text": "Now process it"},
			},
		},
	}

	message, err := parser.ParseMessage(data)
	if err != nil {
		t.Fatalf("Failed to parse user message with mixed content: %v", err)
	}

	userMsg, ok := message.(*claudecode.UserMessage)
	if !ok {
		t.Fatalf("Expected UserMessage, got %T", message)
	}

	blocks, ok := userMsg.Content.([]claudecode.ContentBlock)
	if !ok {
		t.Fatalf("Expected Content to be []ContentBlock, got %T", userMsg.Content)
	}

	if len(blocks) != 4 {
		t.Fatalf("Expected 4 content blocks, got %d", len(blocks))
	}

	// Verify block types in order
	expectedTypes := []string{"text", "tool_use", "tool_result", "text"}
	for i, block := range blocks {
		var blockType string
		switch block.(type) {
		case *claudecode.TextBlock:
			blockType = "text"
		case *claudecode.ToolUseBlock:
			blockType = "tool_use"
		case *claudecode.ToolResultBlock:
			blockType = "tool_result"
		default:
			blockType = "unknown"
		}

		if blockType != expectedTypes[i] {
			t.Errorf("Block %d: expected type '%s', got '%s'", i, expectedTypes[i], blockType)
		}
	}
}

// TestParseValidAssistantMessage tests T040: Parse Valid Assistant Message
func TestParseValidAssistantMessage(t *testing.T) {
	parser := New()
	data := map[string]any{
		"type": "assistant",
		"message": map[string]any{
			"content": []any{
				map[string]any{"type": "text", "text": "I'll help you with that"},
				map[string]any{
					"type":  "tool_use",
					"id":    "tool_123",
					"name":  "Calculate",
					"input": map[string]any{"expression": "2 + 2"},
				},
			},
			"model": "claude-3-5-sonnet-20241022",
		},
	}

	message, err := parser.ParseMessage(data)
	if err != nil {
		t.Fatalf("Failed to parse assistant message: %v", err)
	}

	assistantMsg, ok := message.(*claudecode.AssistantMessage)
	if !ok {
		t.Fatalf("Expected AssistantMessage, got %T", message)
	}

	if len(assistantMsg.Content) != 2 {
		t.Fatalf("Expected 2 content blocks, got %d", len(assistantMsg.Content))
	}

	if assistantMsg.Model != "claude-3-5-sonnet-20241022" {
		t.Errorf("Expected model 'claude-3-5-sonnet-20241022', got '%s'", assistantMsg.Model)
	}

	// Check text block
	textBlock, ok := assistantMsg.Content[0].(*claudecode.TextBlock)
	if !ok {
		t.Fatalf("Expected first block to be TextBlock, got %T", assistantMsg.Content[0])
	}
	if textBlock.Text != "I'll help you with that" {
		t.Errorf("Expected text 'I'll help you with that', got '%s'", textBlock.Text)
	}

	// Check tool use block
	toolUseBlock, ok := assistantMsg.Content[1].(*claudecode.ToolUseBlock)
	if !ok {
		t.Fatalf("Expected second block to be ToolUseBlock, got %T", assistantMsg.Content[1])
	}
	if toolUseBlock.Name != "Calculate" {
		t.Errorf("Expected tool name 'Calculate', got '%s'", toolUseBlock.Name)
	}
}

// TestParseAssistantMessageWithThinking tests T041: Parse Assistant Message with Thinking block
func TestParseAssistantMessageWithThinking(t *testing.T) {
	parser := New()
	data := map[string]any{
		"type": "assistant",
		"message": map[string]any{
			"content": []any{
				map[string]any{
					"type":      "thinking",
					"thinking":  "Let me think about this problem step by step...",
					"signature": "thinking_block_sig_123",
				},
				map[string]any{"type": "text", "text": "Here's my analysis:"},
			},
			"model": "claude-3-5-sonnet-20241022",
		},
	}

	message, err := parser.ParseMessage(data)
	if err != nil {
		t.Fatalf("Failed to parse assistant message with thinking: %v", err)
	}

	assistantMsg, ok := message.(*claudecode.AssistantMessage)
	if !ok {
		t.Fatalf("Expected AssistantMessage, got %T", message)
	}

	if len(assistantMsg.Content) != 2 {
		t.Fatalf("Expected 2 content blocks, got %d", len(assistantMsg.Content))
	}

	// Check thinking block
	thinkingBlock, ok := assistantMsg.Content[0].(*claudecode.ThinkingBlock)
	if !ok {
		t.Fatalf("Expected first block to be ThinkingBlock, got %T", assistantMsg.Content[0])
	}

	if thinkingBlock.Thinking != "Let me think about this problem step by step..." {
		t.Errorf("Expected thinking text, got '%s'", thinkingBlock.Thinking)
	}

	if thinkingBlock.Signature != "thinking_block_sig_123" {
		t.Errorf("Expected signature 'thinking_block_sig_123', got '%s'", thinkingBlock.Signature)
	}
}

// TestParseValidSystemMessage tests T042: Parse Valid System Message
func TestParseValidSystemMessage(t *testing.T) {
	parser := New()
	data := map[string]any{
		"type":      "system",
		"subtype":   "tool_output",
		"data":      map[string]any{"output": "System ready"},
		"timestamp": "2024-01-01T12:00:00Z",
	}

	message, err := parser.ParseMessage(data)
	if err != nil {
		t.Fatalf("Failed to parse system message: %v", err)
	}

	systemMsg, ok := message.(*claudecode.SystemMessage)
	if !ok {
		t.Fatalf("Expected SystemMessage, got %T", message)
	}

	if systemMsg.Subtype != "tool_output" {
		t.Errorf("Expected subtype 'tool_output', got '%s'", systemMsg.Subtype)
	}

	// Verify all original data is preserved
	if timestamp, ok := systemMsg.Data["timestamp"].(string); !ok || timestamp != "2024-01-01T12:00:00Z" {
		t.Errorf("Expected timestamp to be preserved, got %v", systemMsg.Data["timestamp"])
	}
}

// TestParseValidResultMessage tests T043: Parse Valid Result Message
func TestParseValidResultMessage(t *testing.T) {
	parser := New()
	data := map[string]any{
		"type":            "result",
		"subtype":         "query_completed",
		"duration_ms":     1500.0,
		"duration_api_ms": 800.0,
		"is_error":        false,
		"num_turns":       2.0,
		"session_id":      "session_123",
		"total_cost_usd":  0.05,
		"result":          "Task completed successfully",
	}

	message, err := parser.ParseMessage(data)
	if err != nil {
		t.Fatalf("Failed to parse result message: %v", err)
	}

	resultMsg, ok := message.(*claudecode.ResultMessage)
	if !ok {
		t.Fatalf("Expected ResultMessage, got %T", message)
	}

	if resultMsg.Subtype != "query_completed" {
		t.Errorf("Expected subtype 'query_completed', got '%s'", resultMsg.Subtype)
	}
	if resultMsg.DurationMs != 1500 {
		t.Errorf("Expected duration_ms 1500, got %d", resultMsg.DurationMs)
	}
	if resultMsg.DurationAPIMs != 800 {
		t.Errorf("Expected duration_api_ms 800, got %d", resultMsg.DurationAPIMs)
	}
	if resultMsg.IsError {
		t.Errorf("Expected is_error false, got %t", resultMsg.IsError)
	}
	if resultMsg.NumTurns != 2 {
		t.Errorf("Expected num_turns 2, got %d", resultMsg.NumTurns)
	}
	if resultMsg.SessionID != "session_123" {
		t.Errorf("Expected session_id 'session_123', got '%s'", resultMsg.SessionID)
	}

	// Check optional fields
	if resultMsg.TotalCostUSD == nil || *resultMsg.TotalCostUSD != 0.05 {
		t.Errorf("Expected total_cost_usd 0.05, got %v", resultMsg.TotalCostUSD)
	}
	if resultMsg.Result == nil || *resultMsg.Result != "Task completed successfully" {
		t.Errorf("Expected result 'Task completed successfully', got %v", resultMsg.Result)
	}
}

// TestParseInvalidDataTypeError tests T044: Parse Invalid Data Type Error
func TestParseInvalidDataTypeError(t *testing.T) {
	parser := New()

	// Test with non-dict input (already handled in ParseMessage since we expect map[string]any)
	_, err := parser.ParseMessage(nil)
	if err == nil {
		t.Fatal("Expected error for nil input, got nil")
	}

	msgParseErr, ok := err.(*claudecode.MessageParseError)
	if !ok {
		t.Fatalf("Expected MessageParseError, got %T", err)
	}

	if !strings.Contains(msgParseErr.Error(), "missing or invalid type field") {
		t.Errorf("Expected error about missing type field, got: %s", msgParseErr.Error())
	}
}

// TestParseMissingTypeFieldError tests T045: Parse Missing Type Field Error
func TestParseMissingTypeFieldError(t *testing.T) {
	parser := New()
	data := map[string]any{
		"message": map[string]any{"content": "test"},
	}

	_, err := parser.ParseMessage(data)
	if err == nil {
		t.Fatal("Expected error for missing type field, got nil")
	}

	msgParseErr, ok := err.(*claudecode.MessageParseError)
	if !ok {
		t.Fatalf("Expected MessageParseError, got %T", err)
	}

	if !strings.Contains(msgParseErr.Error(), "missing or invalid type field") {
		t.Errorf("Expected error about missing type field, got: %s", msgParseErr.Error())
	}
}

// TestParseUnknownMessageTypeError tests T046: Parse Unknown Message Type Error
func TestParseUnknownMessageTypeError(t *testing.T) {
	parser := New()
	data := map[string]any{
		"type":    "unknown_type",
		"content": "test",
	}

	_, err := parser.ParseMessage(data)
	if err == nil {
		t.Fatal("Expected error for unknown message type, got nil")
	}

	msgParseErr, ok := err.(*claudecode.MessageParseError)
	if !ok {
		t.Fatalf("Expected MessageParseError, got %T", err)
	}

	if !strings.Contains(msgParseErr.Error(), "unknown message type: unknown_type") {
		t.Errorf("Expected error about unknown message type, got: %s", msgParseErr.Error())
	}
}

// TestParseUserMessageMissingFields tests T047: Parse User Message Missing Fields
func TestParseUserMessageMissingFields(t *testing.T) {
	parser := New()

	// Test missing message field
	data := map[string]any{
		"type": "user",
	}

	_, err := parser.ParseMessage(data)
	if err == nil {
		t.Fatal("Expected error for missing message field, got nil")
	}

	msgParseErr, ok := err.(*claudecode.MessageParseError)
	if !ok {
		t.Fatalf("Expected MessageParseError, got %T", err)
	}

	if !strings.Contains(msgParseErr.Error(), "user message missing message field") {
		t.Errorf("Expected error about missing message field, got: %s", msgParseErr.Error())
	}

	// Test missing content field
	data2 := map[string]any{
		"type":    "user",
		"message": map[string]any{},
	}

	_, err2 := parser.ParseMessage(data2)
	if err2 == nil {
		t.Fatal("Expected error for missing content field, got nil")
	}

	msgParseErr2, ok := err2.(*claudecode.MessageParseError)
	if !ok {
		t.Fatalf("Expected MessageParseError, got %T", err2)
	}

	if !strings.Contains(msgParseErr2.Error(), "user message missing content field") {
		t.Errorf("Expected error about missing content field, got: %s", msgParseErr2.Error())
	}
}

// TestMultipleJSONObjectsSingleLine tests T060: Multiple JSON Objects on Single Line
func TestMultipleJSONObjectsSingleLine(t *testing.T) {
	parser := New()

	// Simulate buffered output with multiple JSON objects on single line
	obj1 := `{"type": "user", "message": {"content": [{"type": "text", "text": "First"}]}}`
	obj2 := `{"type": "system", "subtype": "status", "message": "ok"}`

	line := obj1 + "\n" + obj2

	messages, err := parser.ProcessLine(line)
	if err != nil {
		t.Fatalf("Failed to process line with multiple JSON objects: %v", err)
	}

	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}

	// Check first message
	userMsg, ok := messages[0].(*claudecode.UserMessage)
	if !ok {
		t.Fatalf("Expected first message to be UserMessage, got %T", messages[0])
	}

	blocks, ok := userMsg.Content.([]claudecode.ContentBlock)
	if !ok {
		t.Fatalf("Expected Content to be []ContentBlock, got %T", userMsg.Content)
	}

	textBlock, ok := blocks[0].(*claudecode.TextBlock)
	if !ok || textBlock.Text != "First" {
		t.Errorf("Expected text 'First', got %v", textBlock)
	}

	// Check second message
	systemMsg, ok := messages[1].(*claudecode.SystemMessage)
	if !ok {
		t.Fatalf("Expected second message to be SystemMessage, got %T", messages[1])
	}

	if systemMsg.Subtype != "status" {
		t.Errorf("Expected subtype 'status', got '%s'", systemMsg.Subtype)
	}
}

// TestBufferOverflowProtection tests T062: Buffer Overflow Protection
func TestBufferOverflowProtection(t *testing.T) {
	parser := New()

	// Create a string larger than MaxBufferSize (1MB)
	largeString := strings.Repeat("x", MaxBufferSize+1000)

	_, err := parser.processJSONLine(largeString)
	if err == nil {
		t.Fatal("Expected error for buffer overflow, got nil")
	}

	jsonDecodeErr, ok := err.(*claudecode.JSONDecodeError)
	if !ok {
		t.Fatalf("Expected JSONDecodeError, got %T", err)
	}

	if !strings.Contains(jsonDecodeErr.Error(), "buffer overflow") {
		t.Errorf("Expected buffer overflow error, got: %s", jsonDecodeErr.Error())
	}

	// Verify buffer was reset
	if parser.BufferSize() != 0 {
		t.Errorf("Expected buffer to be reset after overflow, but size is %d", parser.BufferSize())
	}
}

// TestSpeculativeJSONParsing tests T063: Speculative JSON Parsing
func TestSpeculativeJSONParsing(t *testing.T) {
	parser := New()

	// First, send incomplete JSON
	msg1, err1 := parser.processJSONLine(`{"type": "user", "message":`)
	if err1 != nil {
		t.Fatalf("Expected no error for incomplete JSON, got %v", err1)
	}
	if msg1 != nil {
		t.Fatal("Expected no message for incomplete JSON, got message")
	}

	// Buffer should contain the partial JSON
	if parser.BufferSize() == 0 {
		t.Fatal("Expected buffer to contain partial JSON")
	}

	// Complete the JSON
	msg2, err2 := parser.processJSONLine(` {"content": [{"type": "text", "text": "Hello"}]}}`)
	if err2 != nil {
		t.Fatalf("Failed to parse completed JSON: %v", err2)
	}

	if msg2 == nil {
		t.Fatal("Expected message after completing JSON, got nil")
	}

	// Verify the parsed message
	userMsg, ok := msg2.(*claudecode.UserMessage)
	if !ok {
		t.Fatalf("Expected UserMessage, got %T", msg2)
	}

	blocks, ok := userMsg.Content.([]claudecode.ContentBlock)
	if !ok {
		t.Fatalf("Expected Content to be []ContentBlock, got %T", userMsg.Content)
	}

	textBlock, ok := blocks[0].(*claudecode.TextBlock)
	if !ok || textBlock.Text != "Hello" {
		t.Errorf("Expected text 'Hello', got %v", textBlock)
	}

	// Buffer should be reset after successful parse
	if parser.BufferSize() != 0 {
		t.Errorf("Expected buffer to be reset after successful parse, but size is %d", parser.BufferSize())
	}
}

// TestMessageParseErrorContainsData tests T051: Message Parse Error Contains Data
func TestMessageParseErrorContainsData(t *testing.T) {
	parser := New()
	originalData := map[string]any{
		"type": "user",
		"message": map[string]any{
			"invalid_field": "should cause error",
			// Missing content field
		},
	}

	_, err := parser.ParseMessage(originalData)
	if err == nil {
		t.Fatal("Expected error for invalid message, got nil")
	}

	msgParseErr, ok := err.(*claudecode.MessageParseError)
	if !ok {
		t.Fatalf("Expected MessageParseError, got %T", err)
	}

	// Verify original data is preserved in error
	if msgParseErr.Data == nil {
		t.Fatal("Expected error to contain original data, got nil")
	}

	data, ok := msgParseErr.Data.(map[string]any)
	if !ok {
		t.Fatalf("Expected error data to be map[string]any, got %T", msgParseErr.Data)
	}

	// Verify the original data fields are present
	if data["type"] != "user" {
		t.Errorf("Expected preserved data to have type 'user', got %v", data["type"])
	}
}

// TestContentBlockTypeDiscrimination tests T052: Content Block Type Discrimination
func TestContentBlockTypeDiscrimination(t *testing.T) {
	parser := New()

	testCases := []struct {
		name      string
		blockData map[string]any
		expected  string
	}{
		{
			name: "text block",
			blockData: map[string]any{
				"type": "text",
				"text": "Hello world",
			},
			expected: "text",
		},
		{
			name: "thinking block",
			blockData: map[string]any{
				"type":      "thinking",
				"thinking":  "Let me think...",
				"signature": "sig123",
			},
			expected: "thinking",
		},
		{
			name: "tool_use block",
			blockData: map[string]any{
				"type":  "tool_use",
				"id":    "tool_1",
				"name":  "Calculator",
				"input": map[string]any{"expr": "1+1"},
			},
			expected: "tool_use",
		},
		{
			name: "tool_result block",
			blockData: map[string]any{
				"type":        "tool_result",
				"tool_use_id": "tool_1",
				"content":     "2",
			},
			expected: "tool_result",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			block, err := parser.parseContentBlock(tc.blockData)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", tc.name, err)
			}

			var blockType string
			switch block.(type) {
			case *claudecode.TextBlock:
				blockType = "text"
			case *claudecode.ThinkingBlock:
				blockType = "thinking"
			case *claudecode.ToolUseBlock:
				blockType = "tool_use"
			case *claudecode.ToolResultBlock:
				blockType = "tool_result"
			default:
				blockType = "unknown"
			}

			if blockType != tc.expected {
				t.Errorf("Expected block type '%s', got '%s'", tc.expected, blockType)
			}
		})
	}
}

// TestOptionalFieldHandling tests T054: Optional Field Handling
func TestOptionalFieldHandling(t *testing.T) {
	parser := New()

	// Test result message with only required fields
	data := map[string]any{
		"type":            "result",
		"subtype":         "test",
		"duration_ms":     100.0,
		"duration_api_ms": 50.0,
		"is_error":        false,
		"num_turns":       1.0,
		"session_id":      "s1",
		// Optional fields omitted: total_cost_usd, usage, result
	}

	message, err := parser.ParseMessage(data)
	if err != nil {
		t.Fatalf("Failed to parse result message without optional fields: %v", err)
	}

	resultMsg, ok := message.(*claudecode.ResultMessage)
	if !ok {
		t.Fatalf("Expected ResultMessage, got %T", message)
	}

	// Verify optional fields are nil when not provided
	if resultMsg.TotalCostUSD != nil {
		t.Errorf("Expected TotalCostUSD to be nil, got %v", *resultMsg.TotalCostUSD)
	}
	if resultMsg.Result != nil {
		t.Errorf("Expected Result to be nil, got %v", *resultMsg.Result)
	}
	if resultMsg.Usage != nil {
		t.Errorf("Expected Usage to be nil, got %v", resultMsg.Usage)
	}
}

// TestEmbeddedNewlinesInJSONStrings tests T061: Embedded Newlines in JSON Strings
func TestEmbeddedNewlinesInJSONStrings(t *testing.T) {
	parser := New()

	// Test JSON with embedded newlines in string values
	jsonWithNewlines := `{"type": "user", "message": {"content": [{"type": "text", "text": "Line 1\nLine 2\nLine 3"}]}}`

	messages, err := parser.ProcessLine(jsonWithNewlines)
	if err != nil {
		t.Fatalf("Failed to parse JSON with embedded newlines: %v", err)
	}

	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	userMsg, ok := messages[0].(*claudecode.UserMessage)
	if !ok {
		t.Fatalf("Expected UserMessage, got %T", messages[0])
	}

	blocks, ok := userMsg.Content.([]claudecode.ContentBlock)
	if !ok {
		t.Fatalf("Expected Content to be []ContentBlock, got %T", userMsg.Content)
	}

	textBlock, ok := blocks[0].(*claudecode.TextBlock)
	if !ok {
		t.Fatalf("Expected TextBlock, got %T", blocks[0])
	}

	expectedText := "Line 1\nLine 2\nLine 3"
	if textBlock.Text != expectedText {
		t.Errorf("Expected text with newlines '%s', got '%s'", expectedText, textBlock.Text)
	}
}

// TestPartialMessageAccumulation tests T064: Partial Message Accumulation
func TestPartialMessageAccumulation(t *testing.T) {
	parser := New()

	// Send partial JSON in chunks
	parts := []string{
		`{"type": "user",`,
		` "message": {"content":`,
		` [{"type": "text",`,
		` "text": "Complete"}]}}`,
	}

	var finalMessage claudecode.Message
	var err error

	// Process each part
	for i, part := range parts {
		msg, parseErr := parser.processJSONLine(part)
		if i < len(parts)-1 {
			// Intermediate parts should not produce a message
			if parseErr != nil {
				t.Fatalf("Unexpected error on part %d: %v", i, parseErr)
			}
			if msg != nil {
				t.Fatalf("Unexpected message on partial part %d", i)
			}
		} else {
			// Final part should complete the message
			if parseErr != nil {
				t.Fatalf("Failed to parse completed message: %v", parseErr)
			}
			if msg == nil {
				t.Fatal("Expected message after completion, got nil")
			}
			finalMessage = msg
			err = parseErr
		}
	}

	if err != nil {
		t.Fatalf("Failed to parse accumulated message: %v", err)
	}

	userMsg, ok := finalMessage.(*claudecode.UserMessage)
	if !ok {
		t.Fatalf("Expected UserMessage, got %T", finalMessage)
	}

	blocks, ok := userMsg.Content.([]claudecode.ContentBlock)
	if !ok {
		t.Fatalf("Expected Content to be []ContentBlock, got %T", userMsg.Content)
	}

	textBlock, ok := blocks[0].(*claudecode.TextBlock)
	if !ok || textBlock.Text != "Complete" {
		t.Errorf("Expected text 'Complete', got %v", textBlock)
	}
}

// TestMalformedJSONRecovery tests T069: Malformed JSON Recovery
func TestMalformedJSONRecovery(t *testing.T) {
	parser := New()

	// Test recovery from buffer size exceeded scenario
	// In speculative parsing, malformed JSON just stays in buffer until
	// either it's completed successfully or buffer size is exceeded

	// Create a large malformed JSON that will exceed buffer size
	largePartialJSON := `{"type": "user", "content": "` + strings.Repeat("x", MaxBufferSize+100) + `invalid`

	// This should trigger buffer overflow error and reset the buffer
	_, err1 := parser.processJSONLine(largePartialJSON)
	if err1 == nil {
		t.Fatal("Expected buffer overflow error for large malformed JSON")
	}

	// Verify it's a buffer overflow error
	jsonDecodeErr, ok := err1.(*claudecode.JSONDecodeError)
	if !ok {
		t.Fatalf("Expected JSONDecodeError, got %T", err1)
	}

	if !strings.Contains(jsonDecodeErr.Error(), "buffer overflow") {
		t.Errorf("Expected buffer overflow error, got: %s", jsonDecodeErr.Error())
	}

	// Buffer should be reset after overflow
	if parser.BufferSize() != 0 {
		t.Errorf("Expected buffer to be reset after overflow, but size is %d", parser.BufferSize())
	}

	// Now parser should be able to handle new valid JSON
	validJSON := `{"type": "system", "subtype": "status"}`
	msg2, err2 := parser.processJSONLine(validJSON)
	if err2 != nil {
		t.Fatalf("Failed to parse valid JSON after recovery: %v", err2)
	}
	if msg2 == nil {
		t.Fatal("Expected valid message after recovery")
	}

	systemMsg, ok := msg2.(*claudecode.SystemMessage)
	if !ok || systemMsg.Subtype != "status" {
		t.Errorf("Expected valid system message with subtype 'status'")
	}
}

// TestParseMessages convenience function test
func TestParseMessages(t *testing.T) {
	lines := []string{
		`{"type": "user", "message": {"content": [{"type": "text", "text": "Hello"}]}}`,
		`{"type": "system", "subtype": "status"}`,
		`{"type": "result", "subtype": "test", "duration_ms": 100, "duration_api_ms": 50, "is_error": false, "num_turns": 1, "session_id": "s1"}`,
	}

	messages, err := ParseMessages(lines)
	if err != nil {
		t.Fatalf("Failed to parse messages: %v", err)
	}

	if len(messages) != 3 {
		t.Fatalf("Expected 3 messages, got %d", len(messages))
	}

	// Verify message types
	expectedTypes := []string{
		claudecode.MessageTypeUser,
		claudecode.MessageTypeSystem,
		claudecode.MessageTypeResult,
	}

	for i, msg := range messages {
		if msg.Type() != expectedTypes[i] {
			t.Errorf("Message %d: expected type '%s', got '%s'", i, expectedTypes[i], msg.Type())
		}
	}
}
