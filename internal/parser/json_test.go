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
			name: "user_message_string_content",
			data: map[string]any{
				"type":    "user",
				"message": map[string]any{"content": "Hello world"},
			},
			expectedType: shared.MessageTypeUser,
		},
		{
			name: "user_message_block_content",
			data: map[string]any{
				"type": "user",
				"message": map[string]any{
					"content": []any{
						map[string]any{"type": "text", "text": "Hello"},
						map[string]any{"type": "tool_use", "id": "t1", "name": "Read"},
					},
				},
			},
			expectedType: shared.MessageTypeUser,
		},
		// Issue #24: UUID and ParentToolUseID field tests
		{
			name: "user_message_with_uuid",
			data: map[string]any{
				"type":    "user",
				"uuid":    "msg-123-abc",
				"message": map[string]any{"content": "Hello"},
			},
			expectedType: shared.MessageTypeUser,
			validate: func(t *testing.T, msg shared.Message) {
				t.Helper()
				um := msg.(*shared.UserMessage)
				if um.UUID == nil || *um.UUID != "msg-123-abc" {
					t.Errorf("expected UUID 'msg-123-abc', got %v", um.UUID)
				}
			},
		},
		{
			name: "user_message_with_parent_tool_use_id",
			data: map[string]any{
				"type":               "user",
				"parent_tool_use_id": "tool-456",
				"message":            map[string]any{"content": "Tool response"},
			},
			expectedType: shared.MessageTypeUser,
			validate: func(t *testing.T, msg shared.Message) {
				t.Helper()
				um := msg.(*shared.UserMessage)
				if um.ParentToolUseID == nil || *um.ParentToolUseID != "tool-456" {
					t.Errorf("expected ParentToolUseID 'tool-456', got %v", um.ParentToolUseID)
				}
			},
		},
		{
			name: "user_message_with_uuid_and_parent_tool_use_id",
			data: map[string]any{
				"type":               "user",
				"uuid":               "msg-789",
				"parent_tool_use_id": "tool-012",
				"message":            map[string]any{"content": "Both fields"},
			},
			expectedType: shared.MessageTypeUser,
			validate: func(t *testing.T, msg shared.Message) {
				t.Helper()
				um := msg.(*shared.UserMessage)
				if um.UUID == nil || *um.UUID != "msg-789" {
					t.Errorf("expected UUID 'msg-789', got %v", um.UUID)
				}
				if um.ParentToolUseID == nil || *um.ParentToolUseID != "tool-012" {
					t.Errorf("expected ParentToolUseID 'tool-012', got %v", um.ParentToolUseID)
				}
			},
		},
		{
			name: "user_message_without_optional_fields",
			data: map[string]any{
				"type":    "user",
				"message": map[string]any{"content": "No optional fields"},
			},
			expectedType: shared.MessageTypeUser,
			validate: func(t *testing.T, msg shared.Message) {
				t.Helper()
				um := msg.(*shared.UserMessage)
				if um.UUID != nil {
					t.Errorf("expected UUID nil, got %v", um.UUID)
				}
				if um.ParentToolUseID != nil {
					t.Errorf("expected ParentToolUseID nil, got %v", um.ParentToolUseID)
				}
			},
		},
		{
			name: "assistant_message",
			data: map[string]any{
				"type": "assistant",
				"message": map[string]any{
					"content": []any{map[string]any{"type": "text", "text": "Hi"}},
					"model":   "claude-3-sonnet",
				},
			},
			expectedType: shared.MessageTypeAssistant,
		},
		// Issue #23: AssistantMessage error field tests
		{
			name: "assistant_message_with_rate_limit_error",
			data: map[string]any{
				"type": "assistant",
				"message": map[string]any{
					"content": []any{map[string]any{"type": "text", "text": "Rate limited"}},
					"model":   "claude-3-sonnet",
					"error":   "rate_limit",
				},
			},
			expectedType: shared.MessageTypeAssistant,
			validate: func(t *testing.T, msg shared.Message) {
				t.Helper()
				am := msg.(*shared.AssistantMessage)
				if am.Error == nil {
					t.Fatal("expected Error to be set, got nil")
				}
				if *am.Error != shared.AssistantMessageErrorRateLimit {
					t.Errorf("expected Error 'rate_limit', got %v", *am.Error)
				}
				if !am.IsRateLimited() {
					t.Error("expected IsRateLimited() to return true")
				}
			},
		},
		{
			name: "assistant_message_with_auth_error",
			data: map[string]any{
				"type": "assistant",
				"message": map[string]any{
					"content": []any{map[string]any{"type": "text", "text": "Auth failed"}},
					"model":   "claude-3-sonnet",
					"error":   "authentication_failed",
				},
			},
			expectedType: shared.MessageTypeAssistant,
			validate: func(t *testing.T, msg shared.Message) {
				t.Helper()
				am := msg.(*shared.AssistantMessage)
				if am.Error == nil {
					t.Fatal("expected Error to be set, got nil")
				}
				if *am.Error != shared.AssistantMessageErrorAuthFailed {
					t.Errorf("expected Error 'authentication_failed', got %v", *am.Error)
				}
				if !am.HasError() {
					t.Error("expected HasError() to return true")
				}
				if am.IsRateLimited() {
					t.Error("expected IsRateLimited() to return false for auth error")
				}
			},
		},
		{
			name: "assistant_message_without_error",
			data: map[string]any{
				"type": "assistant",
				"message": map[string]any{
					"content": []any{map[string]any{"type": "text", "text": "Success"}},
					"model":   "claude-3-sonnet",
				},
			},
			expectedType: shared.MessageTypeAssistant,
			validate: func(t *testing.T, msg shared.Message) {
				t.Helper()
				am := msg.(*shared.AssistantMessage)
				if am.Error != nil {
					t.Errorf("expected Error to be nil, got %v", am.Error)
				}
				if am.HasError() {
					t.Error("expected HasError() to return false")
				}
			},
		},
		{
			name:         "system_message",
			data:         map[string]any{"type": "system", "subtype": "status"},
			expectedType: shared.MessageTypeSystem,
		},
		{
			name: "result_message",
			data: map[string]any{
				"type":            "result",
				"subtype":         "completed",
				"duration_ms":     1500.0,
				"duration_api_ms": 800.0,
				"is_error":        false,
				"num_turns":       2.0,
				"session_id":      "s123",
			},
			expectedType: shared.MessageTypeResult,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parser := setupParserTest(t)
			message, err := parser.ParseMessage(test.data)
			assertParseSuccess(t, err, message)
			assertMessageType(t, message, test.expectedType)
			if test.validate != nil {
				test.validate(t, message)
			}
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
	if systemMsg.Subtype != "status" {
		t.Errorf("Expected subtype 'status', got %q", systemMsg.Subtype)
	}
}

// TestUnicodeAndEscapeHandling tests Unicode and JSON escape sequences
func TestUnicodeAndEscapeHandling(t *testing.T) {
	parser := setupParserTest(t)

	jsonString := `{"type": "user", "message": {"content": [{"type": "text", "text": "Hello 🌍\nEscaped\"Quote"}]}}`
	messages, err := parser.ProcessLine(jsonString)
	assertNoParseError(t, err)
	assertMessageCount(t, messages, 1)

	userMsg := messages[0].(*shared.UserMessage)
	blocks := userMsg.Content.([]shared.ContentBlock)
	assertTextBlockContent(t, blocks[0], "Hello 🌍\nEscaped\"Quote")
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

	emptyInputs := []string{"", "   ", "\t\n"}
	for _, input := range emptyInputs {
		messages, err := parser.ProcessLine(input)
		assertNoParseError(t, err)
		assertMessageCount(t, messages, 0)
	}
}

// TestParseMessages tests the convenience function
func TestParseMessages(t *testing.T) {
	// Test successful parsing
	lines := []string{
		`{"type": "user", "message": {"content": "Hello"}}`,
		`{"type": "system", "subtype": "status"}`,
	}

	messages, err := ParseMessages(lines)
	assertNoParseError(t, err)
	assertMessageCount(t, messages, 2)

	// Test error handling
	errorLines := []string{
		`{"type": "user", "message": {"content": "Valid"}}`,
		`{"type": "invalid"}`, // This should cause an error
	}

	_, err = ParseMessages(errorLines)
	if err == nil {
		t.Error("Expected error for invalid message type")
	}
	if !strings.Contains(err.Error(), "error parsing line 1") {
		t.Errorf("Expected line number in error, got: %v", err)
	}
}

// TestParseErrorConditions tests comprehensive error scenarios
func TestParseErrorConditions(t *testing.T) {
	parser := setupParserTest(t)

	tests := []struct {
		name        string
		data        map[string]any
		expectError string
	}{
		{
			name: "user_message_invalid_content_type",
			data: map[string]any{
				"type":    "user",
				"message": map[string]any{"content": 123}, // Invalid type
			},
			expectError: "invalid user message content type",
		},
		{
			name: "user_message_content_block_parse_error",
			data: map[string]any{
				"type": "user",
				"message": map[string]any{
					"content": []any{
						map[string]any{"type": "text"}, // Missing text field
					},
				},
			},
			expectError: "failed to parse content block 0",
		},
		{
			name:        "assistant_message_missing_message",
			data:        map[string]any{"type": "assistant"},
			expectError: "assistant message missing message field",
		},
		{
			name: "assistant_message_content_not_array",
			data: map[string]any{
				"type": "assistant",
				"message": map[string]any{
					"content": "not an array",
					"model":   "claude-3",
				},
			},
			expectError: "assistant message content must be array",
		},
		{
			name: "assistant_message_missing_model",
			data: map[string]any{
				"type": "assistant",
				"message": map[string]any{
					"content": []any{},
				},
			},
			expectError: "assistant message missing model field",
		},
		{
			name: "assistant_message_content_block_error",
			data: map[string]any{
				"type": "assistant",
				"message": map[string]any{
					"content": []any{
						map[string]any{"type": "unknown_block"},
					},
					"model": "claude-3",
				},
			},
			expectError: "failed to parse content block 0",
		},
		{
			name:        "system_message_missing_subtype",
			data:        map[string]any{"type": "system"},
			expectError: "system message missing subtype field",
		},
		{
			name:        "system_message_invalid_subtype",
			data:        map[string]any{"type": "system", "subtype": 123},
			expectError: "system message missing subtype field",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := parser.ParseMessage(test.data)
			assertParseError(t, err, test.expectError)
		})
	}
}

// TestResultMessageErrorConditions tests uncovered result message parsing paths
func TestResultMessageErrorConditions(t *testing.T) {
	parser := setupParserTest(t)

	tests := []struct {
		name        string
		data        map[string]any
		expectError string
	}{
		{
			name:        "missing_subtype",
			data:        map[string]any{"type": "result"},
			expectError: "result message missing subtype field",
		},
		{
			name:        "invalid_subtype_type",
			data:        map[string]any{"type": "result", "subtype": 123},
			expectError: "result message missing subtype field",
		},
		{
			name: "missing_duration_ms",
			data: map[string]any{
				"type":    "result",
				"subtype": "test",
			},
			expectError: "result message missing or invalid duration_ms field",
		},
		{
			name: "invalid_duration_ms_type",
			data: map[string]any{
				"type":        "result",
				"subtype":     "test",
				"duration_ms": "not a number",
			},
			expectError: "result message missing or invalid duration_ms field",
		},
		{
			name: "missing_duration_api_ms",
			data: map[string]any{
				"type":        "result",
				"subtype":     "test",
				"duration_ms": 100.0,
			},
			expectError: "result message missing or invalid duration_api_ms field",
		},
		{
			name: "invalid_is_error_type",
			data: map[string]any{
				"type":            "result",
				"subtype":         "test",
				"duration_ms":     100.0,
				"duration_api_ms": 50.0,
				"is_error":        "not a boolean",
			},
			expectError: "result message missing or invalid is_error field",
		},
		{
			name: "invalid_num_turns_type",
			data: map[string]any{
				"type":            "result",
				"subtype":         "test",
				"duration_ms":     100.0,
				"duration_api_ms": 50.0,
				"is_error":        false,
				"num_turns":       "not a number",
			},
			expectError: "result message missing or invalid num_turns field",
		},
		{
			name: "missing_session_id",
			data: map[string]any{
				"type":            "result",
				"subtype":         "test",
				"duration_ms":     100.0,
				"duration_api_ms": 50.0,
				"is_error":        false,
				"num_turns":       1.0,
			},
			expectError: "result message missing session_id field",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := parser.ParseMessage(test.data)
			assertParseError(t, err, test.expectError)
		})
	}
}

// TestResultMessageOptionalFields tests optional field handling
func TestResultMessageOptionalFields(t *testing.T) {
	parser := setupParserTest(t)

	baseData := map[string]any{
		"type":            "result",
		"subtype":         "test",
		"duration_ms":     100.0,
		"duration_api_ms": 50.0,
		"is_error":        false,
		"num_turns":       1.0,
		"session_id":      "s123",
	}

	// Test with all optional fields
	dataWithOptionals := make(map[string]any)
	for k, v := range baseData {
		dataWithOptionals[k] = v
	}
	dataWithOptionals["total_cost_usd"] = 0.05
	dataWithOptionals["usage"] = map[string]any{"input_tokens": 100}
	dataWithOptionals["result"] = "The answer is 42"

	msg, err := parser.ParseMessage(dataWithOptionals)
	assertNoParseError(t, err)

	resultMsg := msg.(*shared.ResultMessage)
	if resultMsg.TotalCostUSD == nil || *resultMsg.TotalCostUSD != 0.05 {
		t.Errorf("Expected total_cost_usd = 0.05, got %v", resultMsg.TotalCostUSD)
	}
	if resultMsg.Usage == nil {
		t.Error("Expected usage field to be set")
	}
	if resultMsg.Result == nil {
		t.Error("Expected result field to be set")
	}
	if *resultMsg.Result != "The answer is 42" {
		t.Errorf("Expected result = 'The answer is 42', got %v", *resultMsg.Result)
	}

	// Test with invalid result type (not string)
	dataWithInvalidResult := make(map[string]any)
	for k, v := range baseData {
		dataWithInvalidResult[k] = v
	}
	dataWithInvalidResult["result"] = map[string]any{"not": "a string"}

	msg2, err2 := parser.ParseMessage(dataWithInvalidResult)
	assertNoParseError(t, err2)
	resultMsg2 := msg2.(*shared.ResultMessage)
	if resultMsg2.Result != nil {
		t.Error("Expected result field to be nil for invalid type")
	}
}

// TestContentBlockErrorConditions tests uncovered content block parsing paths
func TestContentBlockErrorConditions(t *testing.T) {
	parser := setupParserTest(t)

	tests := []struct {
		name        string
		blockData   any
		expectError string
	}{
		{
			name:        "non_object_block",
			blockData:   "not an object",
			expectError: "content block must be an object",
		},
		{
			name:        "missing_type_field",
			blockData:   map[string]any{"text": "hello"},
			expectError: "content block missing type field",
		},
		{
			name:        "invalid_type_field",
			blockData:   map[string]any{"type": 123},
			expectError: "content block missing type field",
		},
		{
			name:        "unknown_block_type",
			blockData:   map[string]any{"type": "unknown_type"},
			expectError: "unknown content block type: unknown_type",
		},
		{
			name:        "text_block_missing_text",
			blockData:   map[string]any{"type": "text"},
			expectError: "text block missing text field",
		},
		{
			name:        "text_block_invalid_text_type",
			blockData:   map[string]any{"type": "text", "text": 123},
			expectError: "text block missing text field",
		},
		{
			name:        "thinking_block_missing_thinking",
			blockData:   map[string]any{"type": "thinking"},
			expectError: "thinking block missing thinking field",
		},
		{
			name:        "thinking_block_invalid_thinking_type",
			blockData:   map[string]any{"type": "thinking", "thinking": 123},
			expectError: "thinking block missing thinking field",
		},
		{
			name:        "tool_use_block_missing_id",
			blockData:   map[string]any{"type": "tool_use", "name": "Read"},
			expectError: "tool_use block missing id field",
		},
		{
			name:        "tool_use_block_invalid_id_type",
			blockData:   map[string]any{"type": "tool_use", "id": 123, "name": "Read"},
			expectError: "tool_use block missing id field",
		},
		{
			name:        "tool_use_block_missing_name",
			blockData:   map[string]any{"type": "tool_use", "id": "t1"},
			expectError: "tool_use block missing name field",
		},
		{
			name:        "tool_use_block_invalid_name_type",
			blockData:   map[string]any{"type": "tool_use", "id": "t1", "name": 123},
			expectError: "tool_use block missing name field",
		},
		{
			name:        "tool_result_block_missing_tool_use_id",
			blockData:   map[string]any{"type": "tool_result", "content": "result"},
			expectError: "tool_result block missing tool_use_id field",
		},
		{
			name:        "tool_result_block_invalid_tool_use_id_type",
			blockData:   map[string]any{"type": "tool_result", "tool_use_id": 123, "content": "result"},
			expectError: "tool_result block missing tool_use_id field",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := parser.parseContentBlock(test.blockData)
			assertParseError(t, err, test.expectError)
		})
	}
}

// TestContentBlockOptionalFields tests optional field handling
func TestContentBlockOptionalFields(t *testing.T) {
	parser := setupParserTest(t)

	// Test thinking block without signature
	thinkingBlock, err := parser.parseContentBlock(map[string]any{
		"type":     "thinking",
		"thinking": "I need to think...",
	})
	assertNoParseError(t, err)
	thinking := thinkingBlock.(*shared.ThinkingBlock)
	if thinking.Signature != "" {
		t.Errorf("Expected empty signature, got %q", thinking.Signature)
	}

	// Test tool use block without input
	toolUseBlock, err := parser.parseContentBlock(map[string]any{
		"type": "tool_use",
		"id":   "t1",
		"name": "Read",
	})
	assertNoParseError(t, err)
	toolUse := toolUseBlock.(*shared.ToolUseBlock)
	if toolUse.Input == nil {
		t.Error("Expected empty input map, got nil")
	}
	if len(toolUse.Input) != 0 {
		t.Errorf("Expected empty input map, got %v", toolUse.Input)
	}

	// Test tool result block with invalid is_error type
	toolResultBlock, err := parser.parseContentBlock(map[string]any{
		"type":        "tool_result",
		"tool_use_id": "t1",
		"content":     "result",
		"is_error":    "not a boolean",
	})
	assertNoParseError(t, err)
	toolResult := toolResultBlock.(*shared.ToolResultBlock)
	if toolResult.IsError != nil {
		t.Errorf("Expected nil IsError for invalid type, got %v", toolResult.IsError)
	}
}

// TestProcessLineEdgeCases tests uncovered ProcessLine scenarios
func TestProcessLineEdgeCases(t *testing.T) {
	parser := setupParserTest(t)

	// Test line with content block parse error
	invalidBlockLine := `{"type": "user", "message": {"content": [{"type": "unknown_block"}]}}`
	messages, err := parser.ProcessLine(invalidBlockLine)
	if err == nil {
		t.Error("Expected error for invalid content block")
	}
	if len(messages) != 0 {
		t.Errorf("Expected no messages on error, got %d", len(messages))
	}

	// Test multiple lines with one having an error
	mixedLine := `{"type": "system", "subtype": "ok"}` + "\n" + `{"type": "invalid"}`
	messages2, err2 := parser.ProcessLine(mixedLine)
	if err2 == nil {
		t.Error("Expected error for second invalid message")
	}
	// Should return the first valid message before error
	if len(messages2) != 1 {
		t.Errorf("Expected 1 message before error, got %d", len(messages2))
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
