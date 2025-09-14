package parser

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/severity1/claude-code-sdk-go/internal/shared"
)

// TestParseValidMessages tests parsing of valid message types
func TestParseValidMessages(t *testing.T) {
	tests := []struct {
		name         string
		data         map[string]any
		expectedType string
		validate     func(*testing.T, shared.Message)
	}{
		{
			name: "user_message_with_text",
			data: map[string]any{
				"type": "user",
				"message": map[string]any{
					"content": []any{
						map[string]any{"type": "text", "text": "Hello"},
					},
				},
			},
			expectedType: shared.MessageTypeUser,
			validate: func(t *testing.T, msg shared.Message) {
				userMsg, ok := msg.(*shared.UserMessage)
				if !ok {
					t.Fatalf("Expected UserMessage, got %T", msg)
				}
				blocks, ok := userMsg.Content.([]shared.ContentBlock)
				if !ok {
					t.Fatalf("Expected []ContentBlock, got %T", userMsg.Content)
				}
				assertContentBlockCount(t, blocks, 1)
				assertTextBlockContent(t, blocks[0], "Hello")
			},
		},
		{
			name: "user_message_with_tool_use",
			data: map[string]any{
				"type": "user",
				"message": map[string]any{
					"content": []any{
						map[string]any{
							"type":  "tool_use",
							"id":    "tool_123",
							"name":  "Read",
							"input": map[string]any{"file_path": "/example.txt"},
						},
					},
				},
			},
			expectedType: shared.MessageTypeUser,
			validate: func(t *testing.T, msg shared.Message) {
				userMsg, ok := msg.(*shared.UserMessage)
				if !ok {
					t.Fatalf("Expected UserMessage, got %T", msg)
				}
				blocks, ok := userMsg.Content.([]shared.ContentBlock)
				if !ok {
					t.Fatalf("Expected []ContentBlock, got %T", userMsg.Content)
				}
				assertContentBlockCount(t, blocks, 1)
				assertToolUseBlock(t, blocks[0], "tool_123", "Read")
			},
		},
		{
			name: "user_message_with_tool_result",
			data: map[string]any{
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
			},
			expectedType: shared.MessageTypeUser,
			validate: func(t *testing.T, msg shared.Message) {
				userMsg, ok := msg.(*shared.UserMessage)
				if !ok {
					t.Fatalf("Expected UserMessage, got %T", msg)
				}
				blocks, ok := userMsg.Content.([]shared.ContentBlock)
				if !ok {
					t.Fatalf("Expected []ContentBlock, got %T", userMsg.Content)
				}
				assertContentBlockCount(t, blocks, 1)
				assertToolResultBlock(t, blocks[0], "tool_456", "File content here", false)
			},
		},
		{
			name: "user_message_with_mixed_content",
			data: map[string]any{
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
			},
			expectedType: shared.MessageTypeUser,
			validate: func(t *testing.T, msg shared.Message) {
				userMsg, ok := msg.(*shared.UserMessage)
				if !ok {
					t.Fatalf("Expected UserMessage, got %T", msg)
				}
				blocks, ok := userMsg.Content.([]shared.ContentBlock)
				if !ok {
					t.Fatalf("Expected []ContentBlock, got %T", userMsg.Content)
				}
				assertContentBlockCount(t, blocks, 4)
				assertMixedContentBlocks(t, blocks)
			},
		},
		{
			name: "assistant_message",
			data: map[string]any{
				"type": "assistant",
				"message": map[string]any{
					"content": []any{
						map[string]any{"type": "text", "text": "I'll help you with that"},
					},
					"model": "claude-3-5-sonnet-20241022",
				},
			},
			expectedType: shared.MessageTypeAssistant,
			validate: func(t *testing.T, msg shared.Message) {
				assistantMsg, ok := msg.(*shared.AssistantMessage)
				if !ok {
					t.Fatalf("Expected AssistantMessage, got %T", msg)
				}
				assertContentBlockCount(t, assistantMsg.Content, 1)
				assertTextBlockContent(t, assistantMsg.Content[0], "I'll help you with that")
				assertAssistantModel(t, assistantMsg, "claude-3-5-sonnet-20241022")
			},
		},
		{
			name: "assistant_message_with_thinking",
			data: map[string]any{
				"type": "assistant",
				"message": map[string]any{
					"content": []any{
						map[string]any{
							"type":      "thinking",
							"thinking":  "Let me think about this step by step...",
							"signature": "thinking_block_sig_123",
						},
						map[string]any{"type": "text", "text": "Here's my analysis:"},
					},
					"model": "claude-3-5-sonnet-20241022",
				},
			},
			expectedType: shared.MessageTypeAssistant,
			validate: func(t *testing.T, msg shared.Message) {
				assistantMsg, ok := msg.(*shared.AssistantMessage)
				if !ok {
					t.Fatalf("Expected AssistantMessage, got %T", msg)
				}
				assertContentBlockCount(t, assistantMsg.Content, 2)
				assertThinkingBlock(t, assistantMsg.Content[0], "Let me think about this step by step...")
			},
		},
		{
			name: "system_message",
			data: map[string]any{
				"type":      "system",
				"subtype":   "tool_output",
				"data":      map[string]any{"output": "System ready"},
				"timestamp": "2024-01-01T12:00:00Z",
			},
			expectedType: shared.MessageTypeSystem,
			validate: func(t *testing.T, msg shared.Message) {
				systemMsg, ok := msg.(*shared.SystemMessage)
				if !ok {
					t.Fatalf("Expected SystemMessage, got %T", msg)
				}
				assertSystemSubtype(t, systemMsg, "tool_output")
				assertSystemData(t, systemMsg, "timestamp", "2024-01-01T12:00:00Z")
			},
		},
		{
			name: "result_message",
			data: map[string]any{
				"type":            "result",
				"subtype":         "query_completed",
				"duration_ms":     1500.0,
				"duration_api_ms": 800.0,
				"is_error":        false,
				"num_turns":       2.0,
				"session_id":      "session_123",
				"total_cost_usd":  0.05,
			},
			expectedType: shared.MessageTypeResult,
			validate: func(t *testing.T, msg shared.Message) {
				resultMsg, ok := msg.(*shared.ResultMessage)
				if !ok {
					t.Fatalf("Expected ResultMessage, got %T", msg)
				}
				assertResultFields(t, resultMsg, "query_completed", 1500, 800, false, 2, "session_123")
				assertResultCost(t, resultMsg, 0.05)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parser := setupParserTest(t)

			message, err := parser.ParseMessage(test.data)
			assertParseSuccess(t, err, message)
			assertMessageType(t, message, test.expectedType)

			test.validate(t, message)
		})
	}
}

// TestContentBlockDiscrimination tests parsing of different content block types
func TestContentBlockDiscrimination(t *testing.T) {
	parser := setupParserTest(t)

	tests := []struct {
		name      string
		blockData map[string]any
		blockType string
		validate  func(*testing.T, shared.ContentBlock)
	}{
		{
			name: "text_block",
			blockData: map[string]any{
				"type": "text",
				"text": "Hello world",
			},
			blockType: "text",
			validate: func(t *testing.T, block shared.ContentBlock) {
				assertTextBlockContent(t, block, "Hello world")
			},
		},
		{
			name: "thinking_block",
			blockData: map[string]any{
				"type":      "thinking",
				"thinking":  "Let me think...",
				"signature": "sig123",
			},
			blockType: "thinking",
			validate: func(t *testing.T, block shared.ContentBlock) {
				assertThinkingBlock(t, block, "Let me think...")
			},
		},
		{
			name: "tool_use_block",
			blockData: map[string]any{
				"type":  "tool_use",
				"id":    "tool_1",
				"name":  "Calculator",
				"input": map[string]any{"expr": "1+1"},
			},
			blockType: "tool_use",
			validate: func(t *testing.T, block shared.ContentBlock) {
				assertToolUseBlock(t, block, "tool_1", "Calculator")
			},
		},
		{
			name: "tool_result_block",
			blockData: map[string]any{
				"type":        "tool_result",
				"tool_use_id": "tool_1",
				"content":     "2",
			},
			blockType: "tool_result",
			validate: func(t *testing.T, block shared.ContentBlock) {
				assertToolResultBlock(t, block, "tool_1", "2", false)
			},
		},
		{
			name: "tool_result_error_block",
			blockData: map[string]any{
				"type":        "tool_result",
				"tool_use_id": "tool_2",
				"content":     "Error: File not found",
				"is_error":    true,
			},
			blockType: "tool_result",
			validate: func(t *testing.T, block shared.ContentBlock) {
				assertToolResultBlock(t, block, "tool_2", "Error: File not found", true)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			block, err := parser.parseContentBlock(test.blockData)
			assertParseSuccess(t, err, block)

			actualType := getContentBlockType(block)
			if actualType != test.blockType {
				t.Errorf("Expected block type %s, got %s", test.blockType, actualType)
			}

			test.validate(t, block)
		})
	}
}

// TestParseErrors tests various error conditions
func TestParseErrors(t *testing.T) {
	tests := []struct {
		name        string
		data        map[string]any
		expectError string
	}{
		{
			name:        "missing_type_field",
			data:        map[string]any{"message": map[string]any{"content": "test"}},
			expectError: "missing or invalid type field",
		},
		{
			name:        "unknown_message_type",
			data:        map[string]any{"type": "unknown_type", "content": "test"},
			expectError: "unknown message type: unknown_type",
		},
		{
			name:        "user_message_missing_message_field",
			data:        map[string]any{"type": "user"},
			expectError: "user message missing message field",
		},
		{
			name: "user_message_missing_content_field",
			data: map[string]any{
				"type":    "user",
				"message": map[string]any{},
			},
			expectError: "user message missing content field",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parser := setupParserTest(t)

			_, err := parser.ParseMessage(test.data)
			assertParseError(t, err, test.expectError)
		})
	}
}

// TestSpeculativeJSONParsing tests incomplete JSON handling
func TestSpeculativeJSONParsing(t *testing.T) {
	parser := setupParserTest(t)

	// Send incomplete JSON
	msg1, err1 := parser.processJSONLine(`{"type": "user", "message":`)
	assertNoParseError(t, err1)
	assertNoMessage(t, msg1)
	assertBufferNotEmpty(t, parser)

	// Complete the JSON
	msg2, err2 := parser.processJSONLine(` {"content": [{"type": "text", "text": "Hello"}]}}`)
	assertNoParseError(t, err2)
	assertMessageExists(t, msg2)
	assertBufferEmpty(t, parser)

	// Verify the parsed message
	userMsg, ok := msg2.(*shared.UserMessage)
	if !ok {
		t.Fatalf("Expected UserMessage, got %T", msg2)
	}

	blocks, ok := userMsg.Content.([]shared.ContentBlock)
	if !ok {
		t.Fatalf("Expected Content to be []ContentBlock, got %T", userMsg.Content)
	}
	assertContentBlockCount(t, blocks, 1)
	assertTextBlockContent(t, blocks[0], "Hello")
}

// TestBufferManagement tests buffer overflow protection and management
func TestBufferManagement(t *testing.T) {
	t.Run("buffer_overflow_protection", func(t *testing.T) {
		parser := setupParserTest(t)

		// Create a string larger than MaxBufferSize (1MB)
		largeString := strings.Repeat("x", MaxBufferSize+1000)

		_, err := parser.processJSONLine(largeString)
		assertBufferOverflowError(t, err)
		assertBufferEmpty(t, parser)
	})

	t.Run("buffer_reset_on_success", func(t *testing.T) {
		parser := setupParserTest(t)

		validJSON := `{"type": "system", "subtype": "status"}`
		msg, err := parser.processJSONLine(validJSON)

		assertNoParseError(t, err)
		assertMessageExists(t, msg)
		assertBufferEmpty(t, parser)
	})

	t.Run("partial_message_accumulation", func(t *testing.T) {
		parser := setupParserTest(t)

		parts := []string{
			`{"type": "user",`,
			` "message": {"content":`,
			` [{"type": "text",`,
			` "text": "Complete"}]}}`,
		}

		var finalMessage shared.Message
		for i, part := range parts {
			msg, err := parser.processJSONLine(part)
			assertNoParseError(t, err)

			if i < len(parts)-1 {
				assertNoMessage(t, msg)
				assertBufferNotEmpty(t, parser)
			} else {
				assertMessageExists(t, msg)
				assertBufferEmpty(t, parser)
				finalMessage = msg
			}
		}

		// Verify final message
		userMsg, ok := finalMessage.(*shared.UserMessage)
		if !ok {
			t.Fatalf("Expected UserMessage, got %T", finalMessage)
		}
		blocks, ok := userMsg.Content.([]shared.ContentBlock)
		if !ok {
			t.Fatalf("Expected Content to be []ContentBlock, got %T", userMsg.Content)
		}
		assertTextBlockContent(t, blocks[0], "Complete")
	})

	t.Run("explicit_buffer_reset", func(t *testing.T) {
		parser := setupParserTest(t)

		// Add content to buffer via partial JSON
		msg1, err1 := parser.processJSONLine(`{"type": "user", "message":`)
		assertNoParseError(t, err1)
		assertNoMessage(t, msg1)
		assertBufferNotEmpty(t, parser)

		// Explicit reset should clear buffer
		parser.Reset()
		assertBufferEmpty(t, parser)

		// Parser should work normally after reset
		validJSON := `{"type": "system", "subtype": "status"}`
		msg2, err2 := parser.processJSONLine(validJSON)
		assertNoParseError(t, err2)
		assertMessageExists(t, msg2)
		assertBufferEmpty(t, parser)
	})
}

// TestMultipleJSONObjects tests handling of multiple JSON objects
func TestMultipleJSONObjects(t *testing.T) {
	parser := setupParserTest(t)

	obj1 := `{"type": "user", "message": {"content": [{"type": "text", "text": "First"}]}}`
	obj2 := `{"type": "system", "subtype": "status", "message": "ok"}`
	line := obj1 + "\n" + obj2

	messages, err := parser.ProcessLine(line)
	assertNoParseError(t, err)
	assertMessageCount(t, messages, 2)

	// Verify first message
	userMsg, ok := messages[0].(*shared.UserMessage)
	if !ok {
		t.Fatalf("Expected UserMessage, got %T", messages[0])
	}
	blocks, ok := userMsg.Content.([]shared.ContentBlock)
	if !ok {
		t.Fatalf("Expected Content to be []ContentBlock, got %T", userMsg.Content)
	}
	assertTextBlockContent(t, blocks[0], "First")

	// Verify second message
	systemMsg, ok := messages[1].(*shared.SystemMessage)
	if !ok {
		t.Fatalf("Expected SystemMessage, got %T", messages[1])
	}
	assertSystemSubtype(t, systemMsg, "status")
}

// TestUnicodeAndEscapeHandling tests Unicode and JSON escape sequences
func TestUnicodeAndEscapeHandling(t *testing.T) {
	parser := setupParserTest(t)

	tests := []struct {
		name         string
		jsonString   string
		expectedText string
	}{
		{
			name:         "basic_escape_sequences",
			jsonString: `{"type": "user", "message": {"content": [` +
				`{"type": "text", "text": "Line1\nLine2\tTabbed\"Quoted\"\\Backslash"}]}}`,
			expectedText: "Line1\nLine2\tTabbed\"Quoted\"\\Backslash",
		},
		{
			name:         "unicode_characters",
			jsonString: `{"type": "user", "message": {"content": [` +
				`{"type": "text", "text": "Hello 世界! 🌍 Café naïve résumé"}]}}`,
			expectedText: "Hello 世界! 🌍 Café naïve résumé",
		},
		{
			name:         "mixed_unicode_and_escapes",
			jsonString:   `{"type": "user", "message": {"content": [{"type": "text", "text": "Mixed: 🎉\n中文\t日本語\r한국어"}]}}`,
			expectedText: "Mixed: 🎉\n中文\t日本語\r한국어",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			messages, err := parser.ProcessLine(test.jsonString)
			assertNoParseError(t, err)
			assertMessageCount(t, messages, 1)

			userMsg, ok := messages[0].(*shared.UserMessage)
			if !ok {
				t.Fatalf("Expected UserMessage, got %T", messages[0])
			}
			blocks, ok := userMsg.Content.([]shared.ContentBlock)
			if !ok {
				t.Fatalf("Expected Content to be []ContentBlock, got %T", userMsg.Content)
			}
			assertTextBlockContent(t, blocks[0], test.expectedText)
		})
	}
}

// TestConcurrentAccess tests thread safety
func TestConcurrentAccess(t *testing.T) {
	parser := setupParserTest(t)
	const numGoroutines = 5
	const messagesPerGoroutine = 10

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*messagesPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < messagesPerGoroutine; j++ {
				testJSON := fmt.Sprintf(`{"type": "system", "subtype": "goroutine_%d_msg_%d"}`, goroutineID, j)

				msg, err := parser.processJSONLine(testJSON)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d, message %d: %v", goroutineID, j, err)
					return
				}
				if msg == nil {
					errors <- fmt.Errorf("goroutine %d, message %d: expected message", goroutineID, j)
					return
				}

				systemMsg, ok := msg.(*shared.SystemMessage)
				if !ok {
					errors <- fmt.Errorf("goroutine %d, message %d: wrong type %T", goroutineID, j, msg)
					return
				}

				expectedSubtype := fmt.Sprintf("goroutine_%d_msg_%d", goroutineID, j)
				if systemMsg.Subtype != expectedSubtype {
					errors <- fmt.Errorf("goroutine %d, message %d: expected %s, got %s",
						goroutineID, j, expectedSubtype, systemMsg.Subtype)
					return
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Error(err)
	}
}

// TestLargeMessageHandling tests handling of large messages
func TestLargeMessageHandling(t *testing.T) {
	parser := setupParserTest(t)

	// Test large message under limit (950KB)
	largeContent := strings.Repeat("X", 950*1024)
	largeJSON := fmt.Sprintf(`{"type": "user", "message": {"content": [{"type": "text", "text": %q}]}}`, largeContent)

	if len(largeJSON) >= MaxBufferSize {
		t.Fatalf("Test setup error: large JSON exceeds MaxBufferSize")
	}

	msg, err := parser.processJSONLine(largeJSON)
	assertNoParseError(t, err)
	assertMessageExists(t, msg)

	userMsg, ok := msg.(*shared.UserMessage)
	if !ok {
		t.Fatalf("Expected UserMessage, got %T", msg)
	}
	blocks, ok := userMsg.Content.([]shared.ContentBlock)
	if !ok {
		t.Fatalf("Expected Content to be []ContentBlock, got %T", userMsg.Content)
	}
	textBlock, ok := blocks[0].(*shared.TextBlock)
	if !ok {
		t.Fatalf("Expected TextBlock, got %T", blocks[0])
	}

	if len(textBlock.Text) != len(largeContent) {
		t.Errorf("Expected text length %d, got %d", len(largeContent), len(textBlock.Text))
	}

	assertBufferEmpty(t, parser)
}

// TestEmptyAndWhitespaceHandling tests handling of empty lines
func TestEmptyAndWhitespaceHandling(t *testing.T) {
	parser := setupParserTest(t)

	emptyInputs := []string{"", "   ", "\t", "\n", " \t \n ", "\r\n"}

	for i, input := range emptyInputs {
		t.Run(fmt.Sprintf("empty_input_%d", i), func(t *testing.T) {
			messages, err := parser.ProcessLine(input)
			assertNoParseError(t, err)
			assertMessageCount(t, messages, 0)
		})
	}
}

// TestParseMessages tests the convenience function
func TestParseMessages(t *testing.T) {
	lines := []string{
		`{"type": "user", "message": {"content": [{"type": "text", "text": "Hello"}]}}`,
		`{"type": "system", "subtype": "status"}`,
		`{"type": "result", "subtype": "test", "duration_ms": 100, ` +
			`"duration_api_ms": 50, "is_error": false, "num_turns": 1, "session_id": "s1"}`,
	}

	messages, err := ParseMessages(lines)
	assertNoParseError(t, err)
	assertMessageCount(t, messages, 3)

	expectedTypes := []string{
		shared.MessageTypeUser,
		shared.MessageTypeSystem,
		shared.MessageTypeResult,
	}

	for i, msg := range messages {
		assertMessageType(t, msg, expectedTypes[i])
	}
}

// Mock and Helper Functions

// setupParserTest creates a new parser for testing
func setupParserTest(t *testing.T) *Parser {
	t.Helper()
	return New()
}

// Assertion helpers

func assertParseSuccess(t *testing.T, err error, result any) {
	t.Helper()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result == nil {
		t.Fatal("Expected parse result, got nil")
	}
}

func assertParseError(t *testing.T, err error, expectedMsg string) {
	t.Helper()
	if err == nil {
		t.Fatal("Expected parse error, got nil")
	}
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error containing %q, got %q", expectedMsg, err.Error())
	}
}

func assertNoParseError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Unexpected parse error: %v", err)
	}
}

func assertMessageType(t *testing.T, msg shared.Message, expectedType string) {
	t.Helper()
	if msg.Type() != expectedType {
		t.Errorf("Expected message type %s, got %s", expectedType, msg.Type())
	}
}

func assertMessageExists(t *testing.T, msg shared.Message) {
	t.Helper()
	if msg == nil {
		t.Fatal("Expected message, got nil")
	}
}

func assertNoMessage(t *testing.T, msg shared.Message) {
	t.Helper()
	if msg != nil {
		t.Fatalf("Expected no message, got %T", msg)
	}
}

func assertMessageCount(t *testing.T, messages []shared.Message, expected int) {
	t.Helper()
	if len(messages) != expected {
		t.Errorf("Expected %d messages, got %d", expected, len(messages))
	}
}

func assertContentBlockCount(t *testing.T, blocks []shared.ContentBlock, expected int) {
	t.Helper()
	if len(blocks) != expected {
		t.Errorf("Expected %d content blocks, got %d", expected, len(blocks))
	}
}

func assertTextBlockContent(t *testing.T, block shared.ContentBlock, expectedText string) {
	t.Helper()
	textBlock, ok := block.(*shared.TextBlock)
	if !ok {
		t.Fatalf("Expected TextBlock, got %T", block)
	}
	if textBlock.Text != expectedText {
		t.Errorf("Expected text %q, got %q", expectedText, textBlock.Text)
	}
}

func assertThinkingBlock(t *testing.T, block shared.ContentBlock, expectedThinking string) {
	t.Helper()
	thinkingBlock, ok := block.(*shared.ThinkingBlock)
	if !ok {
		t.Fatalf("Expected ThinkingBlock, got %T", block)
	}
	if thinkingBlock.Thinking != expectedThinking {
		t.Errorf("Expected thinking %q, got %q", expectedThinking, thinkingBlock.Thinking)
	}
}

func assertToolUseBlock(t *testing.T, block shared.ContentBlock, expectedID, expectedName string) {
	t.Helper()
	toolUseBlock, ok := block.(*shared.ToolUseBlock)
	if !ok {
		t.Fatalf("Expected ToolUseBlock, got %T", block)
	}
	if toolUseBlock.ToolUseID != expectedID {
		t.Errorf("Expected tool use ID %q, got %q", expectedID, toolUseBlock.ToolUseID)
	}
	if toolUseBlock.Name != expectedName {
		t.Errorf("Expected tool name %q, got %q", expectedName, toolUseBlock.Name)
	}
}

func assertToolResultBlock(t *testing.T, block shared.ContentBlock,
	expectedToolUseID, expectedContent string, expectedIsError bool) {
	t.Helper()
	toolResultBlock, ok := block.(*shared.ToolResultBlock)
	if !ok {
		t.Fatalf("Expected ToolResultBlock, got %T", block)
	}
	if toolResultBlock.ToolUseID != expectedToolUseID {
		t.Errorf("Expected tool_use_id %q, got %q", expectedToolUseID, toolResultBlock.ToolUseID)
	}
	if content, ok := toolResultBlock.Content.(string); !ok || content != expectedContent {
		t.Errorf("Expected content %q, got %v", expectedContent, toolResultBlock.Content)
	}
	if expectedIsError && (toolResultBlock.IsError == nil || !*toolResultBlock.IsError) {
		t.Error("Expected is_error to be true")
	} else if !expectedIsError && toolResultBlock.IsError != nil && *toolResultBlock.IsError {
		t.Error("Expected is_error to be false or nil")
	}
}

func assertMixedContentBlocks(t *testing.T, blocks []shared.ContentBlock) {
	t.Helper()
	expectedTypes := []string{"text", "tool_use", "tool_result", "text"}
	for i, block := range blocks {
		actualType := getContentBlockType(block)
		if actualType != expectedTypes[i] {
			t.Errorf("Block %d: expected type %s, got %s", i, expectedTypes[i], actualType)
		}
	}
}

func assertAssistantModel(t *testing.T, msg *shared.AssistantMessage, expectedModel string) {
	t.Helper()
	if msg.Model != expectedModel {
		t.Errorf("Expected model %q, got %q", expectedModel, msg.Model)
	}
}

func assertSystemSubtype(t *testing.T, msg *shared.SystemMessage, expectedSubtype string) {
	t.Helper()
	if msg.Subtype != expectedSubtype {
		t.Errorf("Expected subtype %q, got %q", expectedSubtype, msg.Subtype)
	}
}

func assertSystemData(t *testing.T, msg *shared.SystemMessage, key, expectedValue string) {
	t.Helper()
	if value, ok := msg.Data[key].(string); !ok || value != expectedValue {
		t.Errorf("Expected data[%s] = %q, got %v", key, expectedValue, msg.Data[key])
	}
}

func assertResultFields(t *testing.T, msg *shared.ResultMessage, expectedSubtype string,
	expectedDurationMs, expectedDurationAPIMs int, expectedIsError bool,
	expectedNumTurns int, expectedSessionID string) {
	t.Helper()
	if msg.Subtype != expectedSubtype {
		t.Errorf("Expected subtype %q, got %q", expectedSubtype, msg.Subtype)
	}
	if msg.DurationMs != expectedDurationMs {
		t.Errorf("Expected duration_ms %d, got %d", expectedDurationMs, msg.DurationMs)
	}
	if msg.DurationAPIMs != expectedDurationAPIMs {
		t.Errorf("Expected duration_api_ms %d, got %d", expectedDurationAPIMs, msg.DurationAPIMs)
	}
	if msg.IsError != expectedIsError {
		t.Errorf("Expected is_error %t, got %t", expectedIsError, msg.IsError)
	}
	if msg.NumTurns != expectedNumTurns {
		t.Errorf("Expected num_turns %d, got %d", expectedNumTurns, msg.NumTurns)
	}
	if msg.SessionID != expectedSessionID {
		t.Errorf("Expected session_id %q, got %q", expectedSessionID, msg.SessionID)
	}
}

func assertResultCost(t *testing.T, msg *shared.ResultMessage, expectedCost float64) {
	t.Helper()
	if msg.TotalCostUSD == nil || *msg.TotalCostUSD != expectedCost {
		t.Errorf("Expected total_cost_usd %f, got %v", expectedCost, msg.TotalCostUSD)
	}
}

func assertBufferEmpty(t *testing.T, parser *Parser) {
	t.Helper()
	if parser.BufferSize() != 0 {
		t.Errorf("Expected empty buffer, got size %d", parser.BufferSize())
	}
}

func assertBufferNotEmpty(t *testing.T, parser *Parser) {
	t.Helper()
	if parser.BufferSize() == 0 {
		t.Error("Expected non-empty buffer, got empty")
	}
}

func assertBufferOverflowError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("Expected buffer overflow error, got nil")
	}
	jsonDecodeErr, ok := err.(*shared.JSONDecodeError)
	if !ok {
		t.Fatalf("Expected JSONDecodeError, got %T", err)
	}
	if !strings.Contains(jsonDecodeErr.Error(), "buffer overflow") {
		t.Errorf("Expected buffer overflow error, got %q", jsonDecodeErr.Error())
	}
}

func getContentBlockType(block shared.ContentBlock) string {
	switch block.(type) {
	case *shared.TextBlock:
		return "text"
	case *shared.ThinkingBlock:
		return "thinking"
	case *shared.ToolUseBlock:
		return "tool_use"
	case *shared.ToolResultBlock:
		return "tool_result"
	default:
		return "unknown"
	}
}
