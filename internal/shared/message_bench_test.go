package shared

import (
	"encoding/json"
	"testing"
)

// sink prevents dead code elimination by the compiler.
var sink any

// BenchmarkUserMessage_Marshal measures UserMessage JSON serialization.
func BenchmarkUserMessage_Marshal(b *testing.B) {
	tests := []struct {
		name string
		msg  *UserMessage
	}{
		{
			name: "string_content",
			msg: &UserMessage{
				MessageType: MessageTypeUser,
				Content:     "Hello, world!",
			},
		},
		{
			name: "with_uuid",
			msg: func() *UserMessage {
				uuid := "msg-123-abc"
				return &UserMessage{
					MessageType: MessageTypeUser,
					Content:     "Hello with UUID",
					UUID:        &uuid,
				}
			}(),
		},
		{
			name: "with_content_blocks",
			msg: &UserMessage{
				MessageType: MessageTypeUser,
				Content: []ContentBlock{
					&TextBlock{Text: "Some text"},
					&ToolUseBlock{ToolUseID: "tool_1", Name: "Read", Input: map[string]any{"path": "/test"}},
				},
			},
		},
	}

	for _, tc := range tests {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				result, _ := json.Marshal(tc.msg)
				sink = result
			}
		})
	}
}

// BenchmarkAssistantMessage_Marshal measures AssistantMessage JSON serialization.
func BenchmarkAssistantMessage_Marshal(b *testing.B) {
	tests := []struct {
		name string
		msg  *AssistantMessage
	}{
		{
			name: "single_text_block",
			msg: &AssistantMessage{
				MessageType: MessageTypeAssistant,
				Model:       "claude-sonnet-4-5",
				Content: []ContentBlock{
					&TextBlock{Text: "Hello from assistant"},
				},
			},
		},
		{
			name: "multiple_blocks",
			msg: &AssistantMessage{
				MessageType: MessageTypeAssistant,
				Model:       "claude-sonnet-4-5",
				Content: []ContentBlock{
					&TextBlock{Text: "Let me help you with that"},
					&ToolUseBlock{
						ToolUseID: "tool_123",
						Name:      "Read",
						Input:     map[string]any{"path": "/src/main.go"},
					},
				},
			},
		},
		{
			name: "with_thinking",
			msg: &AssistantMessage{
				MessageType: MessageTypeAssistant,
				Model:       "claude-sonnet-4-5",
				Content: []ContentBlock{
					&ThinkingBlock{Thinking: "Let me analyze this request...", Signature: "sig123"},
					&TextBlock{Text: "Based on my analysis..."},
				},
			},
		},
	}

	for _, tc := range tests {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				result, _ := json.Marshal(tc.msg)
				sink = result
			}
		})
	}
}

// BenchmarkResultMessage_Marshal measures ResultMessage JSON serialization.
func BenchmarkResultMessage_Marshal(b *testing.B) {
	cost := 0.01
	usage := map[string]any{
		"input_tokens":  100,
		"output_tokens": 50,
	}
	result := "Task completed"

	msg := &ResultMessage{
		MessageType:   MessageTypeResult,
		Subtype:       "success",
		DurationMs:    1000,
		DurationAPIMs: 800,
		IsError:       false,
		NumTurns:      3,
		SessionID:     "sess_123456",
		TotalCostUSD:  &cost,
		Usage:         &usage,
		Result:        &result,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result, _ := json.Marshal(msg)
		sink = result
	}
}

// BenchmarkMessageDiscrimination measures type switch performance on raw JSON.
func BenchmarkMessageDiscrimination(b *testing.B) {
	testData := [][]byte{
		[]byte(`{"type":"user","message":{"content":"test"}}`),
		[]byte(`{"type":"assistant","message":{"model":"claude","content":[]}}`),
		[]byte(`{"type":"system","subtype":"init"}`),
		[]byte(`{"type":"result","subtype":"success","duration_ms":100,"duration_api_ms":80,"is_error":false,"num_turns":1,"session_id":"s1"}`),
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		data := testData[i%len(testData)]
		var raw map[string]any
		_ = json.Unmarshal(data, &raw)
		msgType := raw["type"].(string)
		sink = msgType
	}
}

// BenchmarkContentBlock_Interface measures interface method call overhead.
func BenchmarkContentBlock_Interface(b *testing.B) {
	blocks := []ContentBlock{
		&TextBlock{Text: "Hello"},
		&ThinkingBlock{Thinking: "Analyzing...", Signature: "sig"},
		&ToolUseBlock{ToolUseID: "t1", Name: "Read", Input: map[string]any{}},
		&ToolResultBlock{ToolUseID: "t1", Content: "result"},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		block := blocks[i%len(blocks)]
		sink = block.BlockType()
	}
}
