//go:build go1.18
// +build go1.18

package shared

import (
	"encoding/json"
	"testing"
)

// FuzzMessage_UnmarshalJSON fuzzes message type unmarshaling.
// Tests each message  type's unmarshaling logic independently from the parser.
func FuzzMessage_UnmarshalJSON(f *testing.F) {
	// Seed corpus with valid JSON for each message type

	// UserMessage - minimal
	f.Add(`{"type": "user", "message": {"content": "test"}}`)
	// UserMessage - with content blocks
	f.Add(`{"type": "user", "message": {"content": [{"type": "text", "text": "test"}]}}`)
	// UserMessage - with UUID
	f.Add(`{"type": "user", "uuid": "msg-123", "message": {"content": "test"}}`)
	// UserMessage - with parent_tool_use_id
	f.Add(`{"type": "user", "parent_tool_use_id": "tool-456", "message": {"content": "test"}}`)

	// AssistantMessage - minimal
	f.Add(`{"type": "assistant", "message": {"content": [{"type": "text", "text": "response"}], "model": "claude-3"}}`)
	// AssistantMessage - with error
	f.Add(`{"type": "assistant", "message": {"content": [], "model": "claude-3", "error": "rate_limit"}}`)
	// AssistantMessage - with thinking block
	f.Add(`{"type": "assistant", "message": {"content": [{"type": "thinking", "thinking": "thought"}], "model": "claude-3"}}`)
	// AssistantMessage - with tool use
	f.Add(`{"type": "assistant", "message": {"content": [{"type": "tool_use", "id": "t1", "name": "read", "input": {}}], "model": "claude-3"}}`)

	// SystemMessage
	f.Add(`{"type": "system", "subtype": "status"}`)
	f.Add(`{"type": "system", "subtype": "error", "error": "timeout"}`)

	// ResultMessage - minimal
	f.Add(`{"type": "result", "subtype": "completed", "duration_ms": 100.0, "duration_api_ms": 50.0, "is_error": false, "num_turns": 1.0, "session_id": "s1"}`)
	// ResultMessage - with optional fields
	f.Add(`{"type": "result", "subtype": "completed", "duration_ms": 100.0, "duration_api_ms": 50.0, "is_error": false, "num_turns": 1.0, "session_id": "s1", "total_cost_usd": 0.05, "result": "success"}`)
	// ResultMessage - with structured_output
	f.Add(`{"type": "result", "subtype": "completed", "duration_ms": 100.0, "duration_api_ms": 50.0, "is_error": false, "num_turns": 1.0, "session_id": "s1", "structured_output": {"answer": 42}}`)

	// ControlRequest
	f.Add(`{"type": "control_request", "action": "pause"}`)

	// ControlResponse
	f.Add(`{"type": "control_response", "status": "ok"}`)

	// StreamEvent
	f.Add(`{"type": "stream_event", "uuid": "evt-1", "session_id": "s1", "event": {"type": "text"}}`)
	f.Add(`{"type": "stream_event", "uuid": "evt-2", "session_id": "s1", "event": {"type": "delta"}, "parent_tool_use_id": "t1"}`)

	// ContentBlock variations - TextBlock
	f.Add(`{"type": "text", "text": "simple text"}`)
	f.Add(`{"type": "text", "text": ""}`)                                   // empty text
	f.Add(`{"type": "text", "text": "` + string(make([]byte, 1000)) + `"}`) // long text

	// ContentBlock - ThinkingBlock
	f.Add(`{"type": "thinking", "thinking": "analyzing..."}`)
	f.Add(`{"type": "thinking", "thinking": "thought", "signature": "sig123"}`)

	// ContentBlock - ToolUseBlock
	f.Add(`{"type": "tool_use", "id": "t1", "name": "read"}`)
	f.Add(`{"type": "tool_use", "id": "t2", "name": "write", "input": {"path": "/test"}}`)

	// ContentBlock - ToolResultBlock
	f.Add(`{"type": "tool_result", "tool_use_id": "t1", "content": "result"}`)
	f.Add(`{"type": "tool_result", "tool_use_id": "t2", "content": {"data": "value"}, "is_error": false}`)
	f.Add(`{"type": "tool_result", "tool_use_id": "t3", "content": "error", "is_error": true}`)

	// Invalid/edge cases
	f.Add(`{}`)                                                      // empty object
	f.Add(`{"type": "unknown"}`)                                     // unknown type
	f.Add(`{"type": "user"}`)                                        // missing required fields
	f.Add(`{"type": "user", "message": {}}`)                         // incomplete message
	f.Add(`{"type": "assistant", "message": {"content": "string"}}`) // wrong content type
	f.Add(`{"type": "result"}`)                                      // missing all required fields
	f.Add(`{"type": "text"}`)                                        // content block missing text field
	f.Add(`{"type": "tool_use"}`)                                    // missing id and name

	// Type mismatches
	f.Add(`{"type": 123}`)                                  // type as number
	f.Add(`{"type": ["array"]}`)                            // type as array
	f.Add(`{"type": null}`)                                 // type as null
	f.Add(`{"type": "user", "message": "string"}`)          // message not object
	f.Add(`{"type": "user", "message": {"content": null}}`) // null content

	// Null values
	f.Add(`{"type": "user", "uuid": null, "message": {"content": "test"}}`)
	f.Add(`{"type": "assistant", "message": {"content": null, "model": "claude-3"}}`)
	f.Add(`{"type": "result", "subtype": null}`)

	f.Fuzz(func(t *testing.T, jsonInput string) {
		// Try to unmarshal to map first to understand the structure
		var rawData map[string]any
		if err := json.Unmarshal([]byte(jsonInput), &rawData); err != nil {
			// Invalid JSON - skip
			t.Skip("invalid JSON")
			return
		}

		// Get the type field to determine what to unmarshal to
		typeField, ok := rawData["type"]
		if !ok {
			// No type field - test with generic map
			return
		}

		typeStr, ok := typeField.(string)
		if !ok {
			// Type field is not a string
			return
		}

		// Test unmarshaling based on type
		switch typeStr {
		case MessageTypeUser:
			testUnmarshalUserMessage(t, jsonInput)
		case MessageTypeAssistant:
			testUnmarshalAssistantMessage(t, jsonInput)
		case MessageTypeSystem:
			testUnmarshalSystemMessage(t, jsonInput)
		case MessageTypeResult:
			testUnmarshalResultMessage(t, jsonInput)
		case MessageTypeControlRequest, MessageTypeControlResponse:
			testUnmarshalControlMessage(t, jsonInput)
		case MessageTypeStreamEvent:
			testUnmarshalStreamEvent(t, jsonInput)
		case ContentBlockTypeText:
			testUnmarshalTextBlock(t, jsonInput)
		case ContentBlockTypeThinking:
			testUnmarshalThinkingBlock(t, jsonInput)
		case ContentBlockTypeToolUse:
			testUnmarshalToolUseBlock(t, jsonInput)
		case ContentBlockTypeToolResult:
			testUnmarshalToolResultBlock(t, jsonInput)
		default:
			// Unknown type - still test for no panic
			var generic map[string]any
			_ = json.Unmarshal([]byte(jsonInput), &generic)
		}
	})
}

// Helper functions to test unmarshaling for each type

func testUnmarshalUserMessage(t *testing.T, jsonInput string) {
	t.Helper()

	var msg UserMessage
	err := json.Unmarshal([]byte(jsonInput), &msg)

	// Either unmarshal successfully or return error (no panics)
	// No validation - just ensuring no panic on unmarshal
	_ = err
}

func testUnmarshalAssistantMessage(t *testing.T, jsonInput string) {
	t.Helper()

	var msg AssistantMessage
	err := json.Unmarshal([]byte(jsonInput), &msg)

	// No validation - just ensuring no panic
	_ = err
}

func testUnmarshalSystemMessage(t *testing.T, jsonInput string) {
	t.Helper()

	var msg SystemMessage
	err := json.Unmarshal([]byte(jsonInput), &msg)

	_ = err
}

func testUnmarshalResultMessage(t *testing.T, jsonInput string) {
	t.Helper()

	var msg ResultMessage
	err := json.Unmarshal([]byte(jsonInput), &msg)

	_ = err
}

func testUnmarshalControlMessage(t *testing.T, jsonInput string) {
	t.Helper()

	// Control messages are just map[string]any
	var msg RawControlMessage
	err := json.Unmarshal([]byte(jsonInput), &msg.Data)

	_ = err // error is OK
}

func testUnmarshalStreamEvent(t *testing.T, jsonInput string) {
	t.Helper()

	var msg StreamEvent
	err := json.Unmarshal([]byte(jsonInput), &msg)

	_ = err
}

func testUnmarshalTextBlock(t *testing.T, jsonInput string) {
	t.Helper()

	var block TextBlock
	err := json.Unmarshal([]byte(jsonInput), &block)

	_ = err
}

func testUnmarshalThinkingBlock(t *testing.T, jsonInput string) {
	t.Helper()

	var block ThinkingBlock
	err := json.Unmarshal([]byte(jsonInput), &block)

	_ = err
}

func testUnmarshalToolUseBlock(t *testing.T, jsonInput string) {
	t.Helper()

	var block ToolUseBlock
	err := json.Unmarshal([]byte(jsonInput), &block)

	_ = err
}

func testUnmarshalToolResultBlock(t *testing.T, jsonInput string) {
	t.Helper()

	var block ToolResultBlock
	err := json.Unmarshal([]byte(jsonInput), &block)

	_ = err
}
