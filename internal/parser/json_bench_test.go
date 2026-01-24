package parser

import (
	"fmt"
	"strings"
	"testing"
)

// sink prevents dead code elimination by the compiler.
var sink any

// BenchmarkProcessLine measures the performance of ProcessLine with various message types.
func BenchmarkProcessLine(b *testing.B) {
	tests := []struct {
		name string
		line string
	}{
		{
			name: "user_simple",
			line: `{"type":"user","message":{"content":"Hello, world!"}}`,
		},
		{
			name: "user_with_uuid",
			line: `{"type":"user","uuid":"msg-123","message":{"content":"Hello with UUID"}}`,
		},
		{
			name: "assistant_text_block",
			line: `{"type":"assistant","message":{"model":"claude-sonnet-4-5","content":[{"type":"text","text":"Hello from assistant"}]}}`,
		},
		{
			name: "assistant_multiple_blocks",
			line: `{"type":"assistant","message":{"model":"claude-sonnet-4-5","content":[{"type":"text","text":"Let me help you"},{"type":"tool_use","id":"tool_123","name":"Read","input":{"path":"/test.go"}}]}}`,
		},
		{
			name: "system_init",
			line: `{"type":"system","subtype":"init","data":{"session_id":"sess_123"}}`,
		},
		{
			name: "result_success",
			line: `{"type":"result","subtype":"success","duration_ms":1000,"duration_api_ms":800,"is_error":false,"num_turns":1,"session_id":"sess_123"}`,
		},
		// Issue #98: tool_use_result benchmark (Python SDK v0.1.22 parity)
		{
			name: "user_with_tool_use_result",
			line: `{"type":"user","uuid":"msg-123","message":{"content":[{"tool_use_id":"t1","type":"tool_result","content":"Updated"}]},"tool_use_result":{"filePath":"/test.py","oldString":"old","newString":"new","originalFile":"contents","structuredPatch":[{"oldStart":1,"oldLines":3,"newStart":1,"newLines":3,"lines":["-old","+new"]}],"userModified":false,"replaceAll":false}}`,
		},
	}

	for _, tc := range tests {
		b.Run(tc.name, func(b *testing.B) {
			p := New()

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				result, _ := p.ProcessLine(tc.line)
				sink = result
				p.Reset()
			}
		})
	}
}

// BenchmarkProcessLine_LargePayload measures parsing performance with large content.
func BenchmarkProcessLine_LargePayload(b *testing.B) {
	sizes := []int{1024, 10 * 1024, 100 * 1024}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("%dKB", size/1024), func(b *testing.B) {
			content := strings.Repeat("x", size)
			line := fmt.Sprintf(`{"type":"user","message":{"content":"%s"}}`, content)
			p := New()

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				result, _ := p.ProcessLine(line)
				sink = result
				p.Reset()
			}
		})
	}
}

// BenchmarkParseMessage measures type discrimination performance.
func BenchmarkParseMessage(b *testing.B) {
	tests := []struct {
		name string
		data map[string]any
	}{
		{
			name: "user",
			data: map[string]any{
				"type":    "user",
				"message": map[string]any{"content": "Hello"},
			},
		},
		{
			name: "assistant",
			data: map[string]any{
				"type": "assistant",
				"message": map[string]any{
					"model": "claude-sonnet-4-5",
					"content": []any{
						map[string]any{"type": "text", "text": "Hello"},
					},
				},
			},
		},
		{
			name: "system",
			data: map[string]any{
				"type":    "system",
				"subtype": "init",
			},
		},
		{
			name: "result",
			data: map[string]any{
				"type":            "result",
				"subtype":         "success",
				"duration_ms":     float64(1000),
				"duration_api_ms": float64(800),
				"is_error":        false,
				"num_turns":       float64(1),
				"session_id":      "sess_123",
			},
		},
	}

	for _, tc := range tests {
		b.Run(tc.name, func(b *testing.B) {
			p := New()

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				result, _ := p.ParseMessage(tc.data)
				sink = result
			}
		})
	}
}

// BenchmarkProcessLine_MultipleMessages measures parsing multiple JSON objects.
func BenchmarkProcessLine_MultipleMessages(b *testing.B) {
	lines := []string{
		`{"type":"user","message":{"content":"First"}}`,
		`{"type":"assistant","message":{"model":"claude-sonnet-4-5","content":[{"type":"text","text":"Response"}]}}`,
		`{"type":"result","subtype":"success","duration_ms":100,"duration_api_ms":80,"is_error":false,"num_turns":1,"session_id":"s1"}`,
	}
	combined := strings.Join(lines, "\n")

	p := New()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result, _ := p.ProcessLine(combined)
		sink = result
		p.Reset()
	}
}
