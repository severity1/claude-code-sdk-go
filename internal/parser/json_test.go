package parser

import (
	"fmt"
	"strings"
	"sync"
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

// TestBufferResetOnSuccess tests T065: Buffer Reset on Success
func TestBufferResetOnSuccess(t *testing.T) {
	parser := New()

	// Parse a complete JSON message
	validJSON := `{"type": "system", "subtype": "status"}`

	msg, err := parser.processJSONLine(validJSON)
	if err != nil {
		t.Fatalf("Failed to parse valid JSON: %v", err)
	}

	if msg == nil {
		t.Fatal("Expected message, got nil")
	}

	// Buffer should be reset after successful parse
	if parser.BufferSize() != 0 {
		t.Errorf("Expected buffer to be reset after successful parse, but size is %d", parser.BufferSize())
	}

	// Test multiple successful parses to ensure consistent reset behavior
	for i := 0; i < 5; i++ {
		testJSON := fmt.Sprintf(`{"type": "system", "subtype": "test_%d"}`, i)

		msg, err := parser.processJSONLine(testJSON)
		if err != nil {
			t.Fatalf("Failed to parse JSON on iteration %d: %v", i, err)
		}

		if msg == nil {
			t.Fatalf("Expected message on iteration %d, got nil", i)
		}

		// Buffer should be reset after each successful parse
		if parser.BufferSize() != 0 {
			t.Errorf("Iteration %d: Expected buffer to be reset, but size is %d", i, parser.BufferSize())
		}
	}
}

// TestConcurrentBufferAccess tests T066: Concurrent Buffer Access
func TestConcurrentBufferAccess(t *testing.T) {
	parser := New()
	const numGoroutines = 10
	const messagesPerGoroutine = 100

	var wg sync.WaitGroup
	results := make(chan error, numGoroutines)

	// Launch multiple goroutines to parse messages concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < messagesPerGoroutine; j++ {
				// Each goroutine parses different messages to avoid confusion
				testJSON := fmt.Sprintf(`{"type": "system", "subtype": "goroutine_%d_msg_%d"}`, goroutineID, j)

				msg, err := parser.processJSONLine(testJSON)
				if err != nil {
					results <- fmt.Errorf("goroutine %d, message %d: failed to parse JSON: %v", goroutineID, j, err)
					return
				}

				if msg == nil {
					results <- fmt.Errorf("goroutine %d, message %d: expected message, got nil", goroutineID, j)
					return
				}

				// Verify the parsed message
				systemMsg, ok := msg.(*claudecode.SystemMessage)
				if !ok {
					results <- fmt.Errorf("goroutine %d, message %d: expected SystemMessage, got %T", goroutineID, j, msg)
					return
				}

				expectedSubtype := fmt.Sprintf("goroutine_%d_msg_%d", goroutineID, j)
				if systemMsg.Subtype != expectedSubtype {
					results <- fmt.Errorf("goroutine %d, message %d: expected subtype %s, got %s", goroutineID, j, expectedSubtype, systemMsg.Subtype)
					return
				}
			}
		}(i)
	}

	// Wait for all goroutines to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Check for any errors
	for err := range results {
		t.Errorf("Concurrent access error: %v", err)
	}
}

// TestBufferStateManagement tests T067: Buffer State Management
func TestBufferStateManagement(t *testing.T) {
	parser := New()

	// Test 1: Buffer state after partial JSON (should accumulate)
	partialJSON := `{"type": "user", "message"`

	msg, err := parser.processJSONLine(partialJSON)
	if err != nil {
		t.Fatalf("Unexpected error for partial JSON: %v", err)
	}

	if msg != nil {
		t.Fatalf("Expected no message for partial JSON, got: %v", msg)
	}

	// Buffer should contain the partial JSON
	expectedSize := len(partialJSON)
	if parser.BufferSize() != expectedSize {
		t.Errorf("Expected buffer size %d after partial JSON, got %d", expectedSize, parser.BufferSize())
	}

	// Test 2: Complete the JSON (should parse successfully and reset buffer)
	completingJSON := `: {"content": [{"type": "text", "text": "hello"}]}}`

	msg, err = parser.processJSONLine(completingJSON)
	if err != nil {
		t.Fatalf("Failed to parse completed JSON: %v", err)
	}

	if msg == nil {
		t.Fatal("Expected message after completing JSON, got nil")
	}

	// Buffer should be reset after successful parse
	if parser.BufferSize() != 0 {
		t.Errorf("Expected buffer to be reset after successful parse, but size is %d", parser.BufferSize())
	}

	// Test 3: Multiple failed parses should accumulate correctly
	parser.Reset() // Start fresh

	partials := []string{
		`{"type":`,
		` "system",`,
		` "subtype":`,
		` "test"}`,
	}

	for i, partial := range partials {
		msg, err := parser.processJSONLine(partial)
		if err != nil {
			t.Fatalf("Unexpected error for partial %d: %v", i, err)
		}

		if i < len(partials)-1 {
			// Should not have message yet
			if msg != nil {
				t.Fatalf("Partial %d: Expected no message, got: %v", i, msg)
			}

			// Buffer should be growing
			if parser.BufferSize() == 0 {
				t.Errorf("Partial %d: Expected buffer to contain data", i)
			}
		} else {
			// Final part should complete the message
			if msg == nil {
				t.Fatal("Expected message after final partial, got nil")
			}

			// Buffer should be reset
			if parser.BufferSize() != 0 {
				t.Errorf("Expected buffer to be reset after completing partials, but size is %d", parser.BufferSize())
			}
		}
	}

	// Test 4: Error conditions should maintain consistent state
	parser.Reset()

	// Try to trigger a buffer overflow error
	largeString := strings.Repeat("x", MaxBufferSize+100)

	_, err = parser.processJSONLine(largeString)
	if err == nil {
		t.Fatal("Expected error for oversized buffer, got nil")
	}

	// Buffer should be reset after overflow error
	if parser.BufferSize() != 0 {
		t.Errorf("Expected buffer to be reset after overflow error, but size is %d", parser.BufferSize())
	}

	// Parser should still work after error
	validJSON := `{"type": "system", "subtype": "recovery_test"}`
	msg, err = parser.processJSONLine(validJSON)
	if err != nil {
		t.Fatalf("Parser should work after error, but got: %v", err)
	}

	if msg == nil {
		t.Fatal("Expected message after error recovery, got nil")
	}
}

// TestLargeMessageHandling tests T068: Large Message Handling
func TestLargeMessageHandling(t *testing.T) {
	parser := New()

	// Test 1: Handle large message close to but under the limit (950KB)
	largeContent := strings.Repeat("X", 950*1024) // 950KB
	largeJSON := fmt.Sprintf(`{"type": "user", "message": {"content": [{"type": "text", "text": "%s"}]}}`, largeContent)

	// Should be under 1MB total
	if len(largeJSON) >= MaxBufferSize {
		t.Fatalf("Test setup error: large JSON (%d bytes) exceeds MaxBufferSize (%d bytes)", len(largeJSON), MaxBufferSize)
	}

	msg, err := parser.processJSONLine(largeJSON)
	if err != nil {
		t.Fatalf("Failed to parse large message under limit: %v", err)
	}

	if msg == nil {
		t.Fatal("Expected message for large JSON under limit, got nil")
	}

	// Verify the parsed message contains the large content
	userMsg, ok := msg.(*claudecode.UserMessage)
	if !ok {
		t.Fatalf("Expected UserMessage, got %T", msg)
	}

	blocks, ok := userMsg.Content.([]claudecode.ContentBlock)
	if !ok {
		t.Fatalf("Expected Content to be []ContentBlock, got %T", userMsg.Content)
	}

	textBlock, ok := blocks[0].(*claudecode.TextBlock)
	if !ok {
		t.Fatalf("Expected TextBlock, got %T", blocks[0])
	}

	if len(textBlock.Text) != len(largeContent) {
		t.Errorf("Expected text length %d, got %d", len(largeContent), len(textBlock.Text))
	}

	// Buffer should be reset after successful parse
	if parser.BufferSize() != 0 {
		t.Errorf("Expected buffer to be reset after large message parse, but size is %d", parser.BufferSize())
	}

	// Test 2: Handle large message built incrementally
	parser.Reset()

	// Build a large message in smaller chunks
	baseJSON := `{"type": "system", "subtype": "large_test", "data": "`
	largeData := strings.Repeat("Y", 800*1024) // 800KB of data
	endJSON := `"}`

	// Send in chunks
	chunkSize := 50000 // 50KB chunks
	totalJSON := baseJSON + largeData + endJSON

	var finalMessage claudecode.Message
	for i := 0; i < len(totalJSON); i += chunkSize {
		end := i + chunkSize
		if end > len(totalJSON) {
			end = len(totalJSON)
		}

		chunk := totalJSON[i:end]
		msg, err := parser.processJSONLine(chunk)

		if err != nil {
			t.Fatalf("Error processing chunk at position %d: %v", i, err)
		}

		if i+chunkSize < len(totalJSON) {
			// Intermediate chunks should not produce a message
			if msg != nil {
				t.Fatalf("Unexpected message at chunk position %d", i)
			}
		} else {
			// Final chunk should complete the message
			if msg == nil {
				t.Fatal("Expected message after final chunk")
			}
			finalMessage = msg
		}
	}

	// Verify the final message
	systemMsg, ok := finalMessage.(*claudecode.SystemMessage)
	if !ok {
		t.Fatalf("Expected SystemMessage, got %T", finalMessage)
	}

	if systemMsg.Subtype != "large_test" {
		t.Errorf("Expected subtype 'large_test', got '%s'", systemMsg.Subtype)
	}

	// Buffer should be reset
	if parser.BufferSize() != 0 {
		t.Errorf("Expected buffer to be reset after incremental large message, but size is %d", parser.BufferSize())
	}

	// Test 3: Parser should still work efficiently after handling large messages
	parser.Reset()

	for i := 0; i < 10; i++ {
		smallJSON := fmt.Sprintf(`{"type": "system", "subtype": "post_large_%d"}`, i)

		msg, err := parser.processJSONLine(smallJSON)
		if err != nil {
			t.Fatalf("Parser failed on small message %d after large message handling: %v", i, err)
		}

		if msg == nil {
			t.Fatalf("Expected message for small JSON %d, got nil", i)
		}
	}
}

// TestLineBoundaryEdgeCases tests T070: Line Boundary Edge Cases
func TestLineBoundaryEdgeCases(t *testing.T) {
	parser := New()

	// Test 1: JSON with multiple embedded newlines creating complex line boundaries
	complexJSON := `{"type": "user", "message": {"content": [
		{"type": "text", "text": "Line 1\nLine 2\nLine 3"},
		{
			"type": "tool_use",
			"id": "tool_123",
			"name": "MultiLineTool",
			"input": {
				"script": "function test() {\n  console.log('hello');\n  return 'world';\n}"
			}
		}
	]}}`

	messages, err := parser.ProcessLine(complexJSON)
	if err != nil {
		t.Fatalf("Failed to parse JSON with complex line boundaries: %v", err)
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

	// Verify first text block preserves newlines
	textBlock, ok := blocks[0].(*claudecode.TextBlock)
	if !ok {
		t.Fatalf("Expected first block to be TextBlock, got %T", blocks[0])
	}

	expectedText := "Line 1\nLine 2\nLine 3"
	if textBlock.Text != expectedText {
		t.Errorf("Expected text with newlines '%s', got '%s'", expectedText, textBlock.Text)
	}

	// Verify tool use block with multiline input
	toolBlock, ok := blocks[1].(*claudecode.ToolUseBlock)
	if !ok {
		t.Fatalf("Expected second block to be ToolUseBlock, got %T", blocks[1])
	}

	script, ok := toolBlock.Input["script"].(string)
	if !ok {
		t.Fatalf("Expected script to be string, got %T", toolBlock.Input["script"])
	}

	expectedScript := "function test() {\n  console.log('hello');\n  return 'world';\n}"
	if script != expectedScript {
		t.Errorf("Expected script with newlines, got '%s'", script)
	}

	// Test 2: Multiple JSON objects with complex line boundaries in single input
	complexMultiJSON := `{"type": "system", "subtype": "start", "data": "first\nsecond"}
	{"type": "user", "message": {"content": [{"type": "text", "text": "A\nB\nC"}]}}
	{"type": "system", "subtype": "end", "multiline": "X\nY\nZ"}`

	messages, err = parser.ProcessLine(complexMultiJSON)
	if err != nil {
		t.Fatalf("Failed to parse multiple JSON with line boundaries: %v", err)
	}

	if len(messages) != 3 {
		t.Fatalf("Expected 3 messages, got %d", len(messages))
	}

	// Verify first system message
	systemMsg1, ok := messages[0].(*claudecode.SystemMessage)
	if !ok {
		t.Fatalf("Expected first message to be SystemMessage, got %T", messages[0])
	}

	if systemMsg1.Subtype != "start" {
		t.Errorf("Expected subtype 'start', got '%s'", systemMsg1.Subtype)
	}

	// Verify user message in the middle
	userMsg2, ok := messages[1].(*claudecode.UserMessage)
	if !ok {
		t.Fatalf("Expected second message to be UserMessage, got %T", messages[1])
	}

	blocks2, ok := userMsg2.Content.([]claudecode.ContentBlock)
	if !ok {
		t.Fatalf("Expected Content to be []ContentBlock, got %T", userMsg2.Content)
	}

	textBlock2, ok := blocks2[0].(*claudecode.TextBlock)
	if !ok {
		t.Fatalf("Expected TextBlock, got %T", blocks2[0])
	}

	expectedText2 := "A\nB\nC"
	if textBlock2.Text != expectedText2 {
		t.Errorf("Expected text '%s', got '%s'", expectedText2, textBlock2.Text)
	}

	// Verify final system message
	systemMsg3, ok := messages[2].(*claudecode.SystemMessage)
	if !ok {
		t.Fatalf("Expected third message to be SystemMessage, got %T", messages[2])
	}

	if systemMsg3.Subtype != "end" {
		t.Errorf("Expected subtype 'end', got '%s'", systemMsg3.Subtype)
	}

	// Test 3: JSON spanning multiple actual lines (fed through ProcessLine)
	parser.Reset()

	// Simulate a JSON object that spans multiple "lines" as received from stream
	jsonParts := []string{
		`{"type": "assistant",`,
		` "message": {`,
		`   "content": [`,
		`     {"type": "text", "text": "Multi\nLine\nResponse"},`,
		`     {"type": "thinking", "thinking": "Step 1\nStep 2\nStep 3"}`,
		`   ],`,
		`   "model": "claude-3-5-sonnet-20241022"`,
		` }`,
		`}`,
	}

	var finalMessage claudecode.Message
	for i, part := range jsonParts {
		messages, err := parser.ProcessLine(part)
		if err != nil {
			t.Fatalf("Error processing part %d: %v", i, err)
		}

		if i < len(jsonParts)-1 {
			// Intermediate parts should not produce messages (except last one)
			if len(messages) != 0 {
				t.Fatalf("Unexpected message at part %d", i)
			}
		} else {
			// Final part should complete the message
			if len(messages) != 1 {
				t.Fatalf("Expected 1 message at final part, got %d", len(messages))
			}
			finalMessage = messages[0]
		}
	}

	// Verify the completed message
	assistantMsg, ok := finalMessage.(*claudecode.AssistantMessage)
	if !ok {
		t.Fatalf("Expected AssistantMessage, got %T", finalMessage)
	}

	if len(assistantMsg.Content) != 2 {
		t.Fatalf("Expected 2 content blocks, got %d", len(assistantMsg.Content))
	}

	// Verify text block preserves newlines
	textBlock3, ok := assistantMsg.Content[0].(*claudecode.TextBlock)
	if !ok {
		t.Fatalf("Expected first block to be TextBlock, got %T", assistantMsg.Content[0])
	}

	expectedText3 := "Multi\nLine\nResponse"
	if textBlock3.Text != expectedText3 {
		t.Errorf("Expected text '%s', got '%s'", expectedText3, textBlock3.Text)
	}

	// Verify thinking block preserves newlines
	thinkingBlock, ok := assistantMsg.Content[1].(*claudecode.ThinkingBlock)
	if !ok {
		t.Fatalf("Expected second block to be ThinkingBlock, got %T", assistantMsg.Content[1])
	}

	expectedThinking := "Step 1\nStep 2\nStep 3"
	if thinkingBlock.Thinking != expectedThinking {
		t.Errorf("Expected thinking '%s', got '%s'", expectedThinking, thinkingBlock.Thinking)
	}
}

// TestBufferSizeTracking tests T071: Buffer Size Tracking
func TestBufferSizeTracking(t *testing.T) {
	parser := New()

	// Test 1: Initial buffer size should be 0
	if parser.BufferSize() != 0 {
		t.Errorf("Expected initial buffer size 0, got %d", parser.BufferSize())
	}

	// Test 2: Buffer size should increase as partial JSON is added
	partialJSON := `{"type": "user"`

	msg, err := parser.processJSONLine(partialJSON)
	if err != nil {
		t.Fatalf("Unexpected error for partial JSON: %v", err)
	}

	if msg != nil {
		t.Fatal("Expected no message for partial JSON")
	}

	expectedSize := len(partialJSON)
	if parser.BufferSize() != expectedSize {
		t.Errorf("Expected buffer size %d after partial JSON, got %d", expectedSize, parser.BufferSize())
	}

	// Test 3: Buffer size should continue growing with more partial content
	additionalPartial := `, "message": {"content"`

	msg, err = parser.processJSONLine(additionalPartial)
	if err != nil {
		t.Fatalf("Unexpected error for additional partial JSON: %v", err)
	}

	if msg != nil {
		t.Fatal("Expected no message for additional partial JSON")
	}

	expectedSize = len(partialJSON + additionalPartial)
	if parser.BufferSize() != expectedSize {
		t.Errorf("Expected buffer size %d after additional partial, got %d", expectedSize, parser.BufferSize())
	}

	// Test 4: Buffer size should reset to 0 after successful parse
	completingJSON := `: [{"type": "text", "text": "hello"}]}}`

	msg, err = parser.processJSONLine(completingJSON)
	if err != nil {
		t.Fatalf("Failed to complete JSON: %v", err)
	}

	if msg == nil {
		t.Fatal("Expected message after completing JSON")
	}

	if parser.BufferSize() != 0 {
		t.Errorf("Expected buffer size 0 after successful parse, got %d", parser.BufferSize())
	}

	// Test 5: Buffer size tracking with Reset() method
	parser.Reset()

	if parser.BufferSize() != 0 {
		t.Errorf("Expected buffer size 0 after Reset(), got %d", parser.BufferSize())
	}

	// Add some content
	testJSON := `{"type": "incomplete`

	msg, err = parser.processJSONLine(testJSON)
	if err != nil {
		t.Fatalf("Unexpected error for test JSON: %v", err)
	}

	if msg != nil {
		t.Fatal("Expected no message for incomplete JSON")
	}

	if parser.BufferSize() != len(testJSON) {
		t.Errorf("Expected buffer size %d, got %d", len(testJSON), parser.BufferSize())
	}

	// Reset should clear it
	parser.Reset()

	if parser.BufferSize() != 0 {
		t.Errorf("Expected buffer size 0 after second Reset(), got %d", parser.BufferSize())
	}

	// Test 6: Buffer size tracking with overflow error
	parser.Reset()

	// Create a string that will exceed buffer limit
	largeString := strings.Repeat("x", MaxBufferSize+100)

	_, err = parser.processJSONLine(largeString)
	if err == nil {
		t.Fatal("Expected error for buffer overflow")
	}

	// Buffer should be reset to 0 after overflow error
	if parser.BufferSize() != 0 {
		t.Errorf("Expected buffer size 0 after overflow error, got %d", parser.BufferSize())
	}

	// Test 7: Accurate size tracking with various character encodings
	parser.Reset()

	unicodeJSON := `{"type": "user", "content": "Hello 🌍 世界"`
	unicodeSize := len(unicodeJSON) // byte length, not rune length

	msg, err = parser.processJSONLine(unicodeJSON)
	if err != nil {
		t.Fatalf("Unexpected error for unicode JSON: %v", err)
	}

	if msg != nil {
		t.Fatal("Expected no message for incomplete unicode JSON")
	}

	if parser.BufferSize() != unicodeSize {
		t.Errorf("Expected buffer size %d for unicode content, got %d", unicodeSize, parser.BufferSize())
	}

	// Test 8: Thread-safe buffer size tracking
	parser.Reset()

	var wg sync.WaitGroup
	const numGoroutines = 5

	// All goroutines will add to buffer concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			partialData := fmt.Sprintf(`{"part_%d": "data"`, id)

			// This will add to buffer but not complete JSON
			_, err := parser.processJSONLine(partialData)
			if err != nil {
				t.Errorf("Goroutine %d: unexpected error: %v", id, err)
			}
		}(i)
	}

	wg.Wait()

	// Buffer should contain all the partial data
	finalSize := parser.BufferSize()
	if finalSize == 0 {
		t.Error("Expected buffer to contain data from all goroutines, got size 0")
	}

	// Buffer should be consistent (no corruption from concurrent access)
	if finalSize < 0 {
		t.Errorf("Buffer size should not be negative, got %d", finalSize)
	}
}

// TestJSONEscapeSequenceHandling tests T074: JSON Escape Sequence Handling
func TestJSONEscapeSequenceHandling(t *testing.T) {
	parser := New()

	// Test 1: Basic escape sequences in text content
	escapedJSON := `{"type": "user", "message": {"content": [{"type": "text", "text": "Line1\nLine2\tTabbed\"Quoted\"\\Backslash"}]}}`

	messages, err := parser.ProcessLine(escapedJSON)
	if err != nil {
		t.Fatalf("Failed to parse JSON with basic escape sequences: %v", err)
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

	expectedText := "Line1\nLine2\tTabbed\"Quoted\"\\Backslash"
	if textBlock.Text != expectedText {
		t.Errorf("Expected text with escape sequences '%s', got '%s'", expectedText, textBlock.Text)
	}

	// Test 2: Escape sequences in tool use input
	toolEscapeJSON := `{"type": "user", "message": {"content": [{"type": "tool_use", "id": "tool_123", "name": "Process", "input": {"script": "if (condition) {\n  console.log(\"Hello\\nWorld\");\n  return \"test\\ttab\";\n}"}}]}}`

	messages, err = parser.ProcessLine(toolEscapeJSON)
	if err != nil {
		t.Fatalf("Failed to parse JSON with tool escape sequences: %v", err)
	}

	userMsg2, ok := messages[0].(*claudecode.UserMessage)
	if !ok {
		t.Fatalf("Expected UserMessage, got %T", messages[0])
	}

	blocks2, ok := userMsg2.Content.([]claudecode.ContentBlock)
	if !ok {
		t.Fatalf("Expected Content to be []ContentBlock, got %T", userMsg2.Content)
	}

	toolBlock, ok := blocks2[0].(*claudecode.ToolUseBlock)
	if !ok {
		t.Fatalf("Expected ToolUseBlock, got %T", blocks2[0])
	}

	script, ok := toolBlock.Input["script"].(string)
	if !ok {
		t.Fatalf("Expected script to be string, got %T", toolBlock.Input["script"])
	}

	expectedScript := "if (condition) {\n  console.log(\"Hello\\nWorld\");\n  return \"test\\ttab\";\n}"
	if script != expectedScript {
		t.Errorf("Expected script with escape sequences, got '%s'", script)
	}

	// Test 3: Complex escape sequences including unicode escapes
	complexEscapeJSON := `{"type": "system", "subtype": "test", "data": "Unicode: \u00E9\u00F1\u00FC and escapes: \\n\\t\\r\\b\\f"}`

	messages, err = parser.ProcessLine(complexEscapeJSON)
	if err != nil {
		t.Fatalf("Failed to parse JSON with complex escape sequences: %v", err)
	}

	systemMsg, ok := messages[0].(*claudecode.SystemMessage)
	if !ok {
		t.Fatalf("Expected SystemMessage, got %T", messages[0])
	}

	dataValue, ok := systemMsg.Data["data"].(string)
	if !ok {
		t.Fatalf("Expected data to be string, got %T", systemMsg.Data["data"])
	}

	expectedData := "Unicode: éñü and escapes: \\n\\t\\r\\b\\f"
	if dataValue != expectedData {
		t.Errorf("Expected data with unicode escapes '%s', got '%s'", expectedData, dataValue)
	}

	// Test 4: Escape sequences in partial JSON (speculative parsing)
	parser.Reset()

	partialEscapeJSON1 := `{"type": "user", "message": {"content": [{"type": "text", "text": "Start\nMiddle`
	partialEscapeJSON2 := `\tEnd\"Quote"}]}}`

	// First part should not produce message
	msg1, err := parser.processJSONLine(partialEscapeJSON1)
	if err != nil {
		t.Fatalf("Unexpected error for first partial escape JSON: %v", err)
	}

	if msg1 != nil {
		t.Fatal("Expected no message for first partial escape JSON")
	}

	// Second part should complete the message
	msg2, err := parser.processJSONLine(partialEscapeJSON2)
	if err != nil {
		t.Fatalf("Failed to complete escape JSON: %v", err)
	}

	if msg2 == nil {
		t.Fatal("Expected message after completing escape JSON")
	}

	userMsg3, ok := msg2.(*claudecode.UserMessage)
	if !ok {
		t.Fatalf("Expected UserMessage, got %T", msg2)
	}

	blocks3, ok := userMsg3.Content.([]claudecode.ContentBlock)
	if !ok {
		t.Fatalf("Expected Content to be []ContentBlock, got %T", userMsg3.Content)
	}

	textBlock2, ok := blocks3[0].(*claudecode.TextBlock)
	if !ok {
		t.Fatalf("Expected TextBlock, got %T", blocks3[0])
	}

	expectedText2 := "Start\nMiddle\tEnd\"Quote"
	if textBlock2.Text != expectedText2 {
		t.Errorf("Expected text with partial escape sequences '%s', got '%s'", expectedText2, textBlock2.Text)
	}

	// Test 5: Malformed escape sequences should still parse (JSON parser handles it)
	malformedEscapeJSON := `{"type": "system", "subtype": "test", "invalid_escape": "This has \\z invalid escape"}`

	messages, err = parser.ProcessLine(malformedEscapeJSON)
	if err != nil {
		t.Fatalf("Failed to parse JSON with malformed escape (Go's JSON parser should handle): %v", err)
	}

	systemMsg2, ok := messages[0].(*claudecode.SystemMessage)
	if !ok {
		t.Fatalf("Expected SystemMessage, got %T", messages[0])
	}

	invalidEscape, ok := systemMsg2.Data["invalid_escape"].(string)
	if !ok {
		t.Fatalf("Expected invalid_escape to be string, got %T", systemMsg2.Data["invalid_escape"])
	}

	// Go's JSON parser will preserve the literal backslash for invalid escapes
	expectedInvalid := "This has \\z invalid escape"
	if invalidEscape != expectedInvalid {
		t.Errorf("Expected malformed escape to be preserved '%s', got '%s'", expectedInvalid, invalidEscape)
	}
}

// TestUnicodeStringHandling tests T075: Unicode String Handling
func TestUnicodeStringHandling(t *testing.T) {
	parser := New()

	// Test 1: Basic Unicode characters in text content
	unicodeJSON := `{"type": "user", "message": {"content": [{"type": "text", "text": "Hello 世界! 🌍 Café naïve résumé"}]}}`

	messages, err := parser.ProcessLine(unicodeJSON)
	if err != nil {
		t.Fatalf("Failed to parse JSON with Unicode characters: %v", err)
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

	expectedText := "Hello 世界! 🌍 Café naïve résumé"
	if textBlock.Text != expectedText {
		t.Errorf("Expected Unicode text '%s', got '%s'", expectedText, textBlock.Text)
	}

	// Test 2: Unicode in tool names and input
	unicodeToolJSON := `{"type": "user", "message": {"content": [{"type": "tool_use", "id": "tool_123", "name": "处理文件", "input": {"文件名": "测试.txt", "内容": "包含中文的内容"}}]}}`

	messages, err = parser.ProcessLine(unicodeToolJSON)
	if err != nil {
		t.Fatalf("Failed to parse JSON with Unicode in tool data: %v", err)
	}

	userMsg2, ok := messages[0].(*claudecode.UserMessage)
	if !ok {
		t.Fatalf("Expected UserMessage, got %T", messages[0])
	}

	blocks2, ok := userMsg2.Content.([]claudecode.ContentBlock)
	if !ok {
		t.Fatalf("Expected Content to be []ContentBlock, got %T", userMsg2.Content)
	}

	toolBlock, ok := blocks2[0].(*claudecode.ToolUseBlock)
	if !ok {
		t.Fatalf("Expected ToolUseBlock, got %T", blocks2[0])
	}

	if toolBlock.Name != "处理文件" {
		t.Errorf("Expected Unicode tool name '处理文件', got '%s'", toolBlock.Name)
	}

	fileName, ok := toolBlock.Input["文件名"].(string)
	if !ok || fileName != "测试.txt" {
		t.Errorf("Expected Unicode input key with value '测试.txt', got %v", fileName)
	}

	content, ok := toolBlock.Input["内容"].(string)
	if !ok || content != "包含中文的内容" {
		t.Errorf("Expected Unicode content '包含中文的内容', got %v", content)
	}

	// Test 3: Mixed Unicode with escape sequences
	mixedUnicodeJSON := `{"type": "system", "subtype": "test", "data": "Mixed: 🎉\n中文\t日本語\r한국어"}`

	messages, err = parser.ProcessLine(mixedUnicodeJSON)
	if err != nil {
		t.Fatalf("Failed to parse JSON with mixed Unicode and escapes: %v", err)
	}

	systemMsg, ok := messages[0].(*claudecode.SystemMessage)
	if !ok {
		t.Fatalf("Expected SystemMessage, got %T", messages[0])
	}

	dataValue, ok := systemMsg.Data["data"].(string)
	if !ok {
		t.Fatalf("Expected data to be string, got %T", systemMsg.Data["data"])
	}

	expectedData := "Mixed: 🎉\n中文\t日本語\r한국어"
	if dataValue != expectedData {
		t.Errorf("Expected mixed Unicode data '%s', got '%s'", expectedData, dataValue)
	}

	// Test 4: Unicode in partial JSON (speculative parsing)
	parser.Reset()

	partialUnicode1 := `{"type": "user", "message": {"content": [{"type": "text", "text": "Start 🌟 中文`
	partialUnicode2 := ` 继续 End 🏁"}]}}`

	// First part should not produce message
	msg1, err := parser.processJSONLine(partialUnicode1)
	if err != nil {
		t.Fatalf("Unexpected error for first partial Unicode JSON: %v", err)
	}

	if msg1 != nil {
		t.Fatal("Expected no message for first partial Unicode JSON")
	}

	// Second part should complete the message
	msg2, err := parser.processJSONLine(partialUnicode2)
	if err != nil {
		t.Fatalf("Failed to complete Unicode JSON: %v", err)
	}

	if msg2 == nil {
		t.Fatal("Expected message after completing Unicode JSON")
	}

	userMsg3, ok := msg2.(*claudecode.UserMessage)
	if !ok {
		t.Fatalf("Expected UserMessage, got %T", msg2)
	}

	blocks3, ok := userMsg3.Content.([]claudecode.ContentBlock)
	if !ok {
		t.Fatalf("Expected Content to be []ContentBlock, got %T", userMsg3.Content)
	}

	textBlock2, ok := blocks3[0].(*claudecode.TextBlock)
	if !ok {
		t.Fatalf("Expected TextBlock, got %T", blocks3[0])
	}

	expectedText2 := "Start 🌟 中文 继续 End 🏁"
	if textBlock2.Text != expectedText2 {
		t.Errorf("Expected partial Unicode text '%s', got '%s'", expectedText2, textBlock2.Text)
	}

	// Test 5: Unicode buffer size calculation (bytes vs runes)
	parser.Reset()

	// Test that buffer size is calculated in bytes, not runes
	unicodePartial := "Hello 世界" // This is more bytes than runes due to UTF-8 encoding
	unicodeSizeBytes := len(unicodePartial)
	unicodeSizeRunes := len([]rune(unicodePartial))

	// Should be different - bytes > runes for Unicode
	if unicodeSizeBytes <= unicodeSizeRunes {
		t.Errorf("Test setup error: expected byte length (%d) > rune length (%d)", unicodeSizeBytes, unicodeSizeRunes)
	}

	incompleteUnicodeJSON := fmt.Sprintf(`{"type": "user", "content": "%s`, unicodePartial)

	msg, err := parser.processJSONLine(incompleteUnicodeJSON)
	if err != nil {
		t.Fatalf("Unexpected error for incomplete Unicode JSON: %v", err)
	}

	if msg != nil {
		t.Fatal("Expected no message for incomplete Unicode JSON")
	}

	// Buffer size should be in bytes, not runes
	expectedBufferSize := len(incompleteUnicodeJSON)
	if parser.BufferSize() != expectedBufferSize {
		t.Errorf("Expected buffer size %d bytes, got %d", expectedBufferSize, parser.BufferSize())
	}

	// Test 6: Various Unicode ranges
	parser.Reset()

	variousUnicodeJSON := `{"type": "system", "subtype": "unicode_test", "data": {
		"latin": "àáâãäå",
		"greek": "αβγδε",
		"cyrillic": "абвгд",
		"cjk": "你好世界",
		"emoji": "😀😃😄😁",
		"symbols": "∞∑∫∂∆",
		"arrows": "←↑→↓↔"
	}}`

	messages, err = parser.ProcessLine(variousUnicodeJSON)
	if err != nil {
		t.Fatalf("Failed to parse JSON with various Unicode ranges: %v", err)
	}

	systemMsg2, ok := messages[0].(*claudecode.SystemMessage)
	if !ok {
		t.Fatalf("Expected SystemMessage, got %T", messages[0])
	}

	data, ok := systemMsg2.Data["data"].(map[string]any)
	if !ok {
		t.Fatalf("Expected data to be map, got %T", systemMsg2.Data["data"])
	}

	// Verify each Unicode range was preserved
	testCases := map[string]string{
		"latin":    "àáâãäå",
		"greek":    "αβγδε",
		"cyrillic": "абвгд",
		"cjk":      "你好世界",
		"emoji":    "😀😃😄😁",
		"symbols":  "∞∑∫∂∆",
		"arrows":   "←↑→↓↔",
	}

	for key, expected := range testCases {
		actual, ok := data[key].(string)
		if !ok {
			t.Errorf("Expected %s to be string, got %T", key, data[key])
			continue
		}
		if actual != expected {
			t.Errorf("Unicode range %s: expected '%s', got '%s'", key, expected, actual)
		}
	}
}

// TestEmptyMessageHandling tests T076: Empty Message Handling
func TestEmptyMessageHandling(t *testing.T) {
	parser := New()

	// Test 1: Empty line should return no messages
	messages, err := parser.ProcessLine("")
	if err != nil {
		t.Fatalf("Unexpected error for empty line: %v", err)
	}

	if len(messages) != 0 {
		t.Fatalf("Expected 0 messages for empty line, got %d", len(messages))
	}

	// Test 2: Whitespace-only line should return no messages
	whitespaceLines := []string{
		"   ",
		"\t",
		"\n",
		" \t \n ",
		"\r\n",
	}

	for i, line := range whitespaceLines {
		messages, err := parser.ProcessLine(line)
		if err != nil {
			t.Fatalf("Unexpected error for whitespace line %d: %v", i, err)
		}

		if len(messages) != 0 {
			t.Errorf("Whitespace line %d: expected 0 messages, got %d", i, len(messages))
		}
	}

	// Test 3: Mixed empty and valid lines
	mixedInput := `
	
	{"type": "system", "subtype": "start"}
	
	
	{"type": "user", "message": {"content": [{"type": "text", "text": "Hello"}]}}
	
	   
	{"type": "system", "subtype": "end"}
	
	`

	messages, err = parser.ProcessLine(mixedInput)
	if err != nil {
		t.Fatalf("Failed to parse mixed empty and valid lines: %v", err)
	}

	// Should only get the 3 valid messages, ignoring empty lines
	if len(messages) != 3 {
		t.Fatalf("Expected 3 messages from mixed input, got %d", len(messages))
	}

	// Test 4: Empty lines in middle of JSON should be handled (as part of split logic)
	multilineInput := `{"type": "system", "subtype": "test1"}

{"type": "system", "subtype": "test2"}`

	messages, err = parser.ProcessLine(multilineInput)
	if err != nil {
		t.Fatalf("Failed to parse multiline with empty line: %v", err)
	}

	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages from multiline input, got %d", len(messages))
	}

	systemMsg1, ok := messages[0].(*claudecode.SystemMessage)
	if !ok || systemMsg1.Subtype != "test1" {
		t.Errorf("Expected first message to be SystemMessage with subtype 'test1'")
	}

	systemMsg2, ok := messages[1].(*claudecode.SystemMessage)
	if !ok || systemMsg2.Subtype != "test2" {
		t.Errorf("Expected second message to be SystemMessage with subtype 'test2'")
	}

	// Test 5: Empty message followed by partial JSON
	parser.Reset()

	// First process an empty line
	messages, err = parser.ProcessLine("   ")
	if err != nil {
		t.Fatalf("Unexpected error for empty line: %v", err)
	}

	if len(messages) != 0 {
		t.Fatal("Expected no messages for empty line")
	}

	// Buffer should still be empty
	if parser.BufferSize() != 0 {
		t.Errorf("Expected buffer size 0 after empty line, got %d", parser.BufferSize())
	}

	// Then process a partial JSON
	partialJSON := `{"type": "user"`
	msg, err := parser.processJSONLine(partialJSON)
	if err != nil {
		t.Fatalf("Unexpected error for partial JSON: %v", err)
	}

	if msg != nil {
		t.Fatal("Expected no message for partial JSON")
	}

	// Buffer should contain the partial JSON
	if parser.BufferSize() != len(partialJSON) {
		t.Errorf("Expected buffer size %d after partial JSON, got %d", len(partialJSON), parser.BufferSize())
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
