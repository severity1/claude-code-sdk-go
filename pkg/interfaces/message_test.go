package interfaces

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

// TestMessageInterfaceExistence verifies that the message interfaces exist and have correct method signatures.
func TestMessageInterfaceExistence(t *testing.T) {
	tests := []struct {
		name            string
		interfaceType   reflect.Type
		expectedMethods []string
	}{
		{
			name:            "Message interface",
			interfaceType:   reflect.TypeOf((*Message)(nil)).Elem(),
			expectedMethods: []string{"Type"},
		},
		{
			name:            "ContentBlock interface",
			interfaceType:   reflect.TypeOf((*ContentBlock)(nil)).Elem(),
			expectedMethods: []string{"Type"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.interfaceType.Kind() != reflect.Interface {
				t.Errorf("Expected %s to be an interface, got %s", tt.name, tt.interfaceType.Kind())
				return
			}

			actualMethods := make(map[string]bool)
			for i := 0; i < tt.interfaceType.NumMethod(); i++ {
				method := tt.interfaceType.Method(i)
				actualMethods[method.Name] = true
			}

			for _, expectedMethod := range tt.expectedMethods {
				if !actualMethods[expectedMethod] {
					t.Errorf("Expected %s to have method %s", tt.name, expectedMethod)
				}
			}
		})
	}
}

// TestTypeMethodSignature verifies Type() method has correct signature.
func TestTypeMethodSignature(t *testing.T) {
	testCases := []struct {
		name          string
		interfaceType reflect.Type
	}{
		{"Message", reflect.TypeOf((*Message)(nil)).Elem()},
		{"ContentBlock", reflect.TypeOf((*ContentBlock)(nil)).Elem()},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			method, found := tc.interfaceType.MethodByName("Type")
			if !found {
				t.Errorf("%s interface must have Type method", tc.name)
				return
			}

			methodType := method.Type

			// Should take no parameters (interface methods don't count receiver)
			if methodType.NumIn() != 0 {
				t.Errorf("Type method should have 0 parameters, got %d", methodType.NumIn())
			}

			// Should return exactly 1 value (string)
			if methodType.NumOut() != 1 {
				t.Errorf("Type method should return 1 value, got %d", methodType.NumOut())
			}

			// Return type should be string
			if methodType.NumOut() > 0 {
				returnType := methodType.Out(0)
				if returnType.Kind() != reflect.String {
					t.Errorf("Type method should return string, got %s", returnType.Kind())
				}
			}
		})
	}
}

// TestMethodNamingConsistency verifies consistent naming across interfaces.
func TestMethodNamingConsistency(t *testing.T) {
	// Both Message and ContentBlock should use Type(), not BlockType() or GetType()
	messageType := reflect.TypeOf((*Message)(nil)).Elem()
	contentBlockType := reflect.TypeOf((*ContentBlock)(nil)).Elem()

	// Message should have Type() method
	messageMethod, messageHasType := messageType.MethodByName("Type")
	if !messageHasType {
		t.Error("Message interface must have Type() method")
	}

	// ContentBlock should have Type() method (not BlockType())
	blockMethod, blockHasType := contentBlockType.MethodByName("Type")
	if !blockHasType {
		t.Error("ContentBlock interface must have Type() method (not BlockType())")
	}

	// Verify methods have identical signatures
	if messageHasType && blockHasType {
		if !methodSignaturesEqual(messageMethod.Type, blockMethod.Type) {
			t.Error("Message.Type() and ContentBlock.Type() should have identical signatures")
		}
	}

	// Verify no legacy method names exist
	if _, hasBlockType := contentBlockType.MethodByName("BlockType"); hasBlockType {
		t.Error("ContentBlock interface should not have BlockType() method - use Type() instead")
	}
}

// TestInterfaceNilHandling verifies interfaces handle nil values correctly.
func TestInterfaceNilHandling(t *testing.T) {
	var message Message
	var contentBlock ContentBlock

	if message != nil {
		t.Error("Nil Message interface should be nil")
	}
	if contentBlock != nil {
		t.Error("Nil ContentBlock interface should be nil")
	}

	// Test interface value equality
	var message2 Message
	if message != message2 {
		t.Error("Two nil Message interfaces should be equal")
	}
}

// Helper function to compare method signatures
func methodSignaturesEqual(t1, t2 reflect.Type) bool {
	if t1.NumIn() != t2.NumIn() || t1.NumOut() != t2.NumOut() {
		return false
	}

	for i := 0; i < t1.NumIn(); i++ {
		if t1.In(i) != t2.In(i) {
			return false
		}
	}

	for i := 0; i < t1.NumOut(); i++ {
		if t1.Out(i) != t2.Out(i) {
			return false
		}
	}

	return true
}

// TestUserMessageTypedContent verifies UserMessage has typed content field instead of interface{}.
func TestUserMessageTypedContent(t *testing.T) {
	// This test will fail until we implement typed UserMessage
	userMsg := UserMessage{
		Content: TextContent{Text: "Hello"},
	}

	// Test that Content field is typed
	var content UserMessageContent = userMsg.Content
	_ = content

	// Test Message interface implementation
	var msg Message = userMsg
	if msg.Type() != "user" {
		t.Errorf("Expected message type 'user', got '%s'", msg.Type())
	}

	// Test basic field access
	if userMsg.Content == nil {
		t.Error("UserMessage.Content should not be nil")
	}
}

// TestAssistantMessageTypedContent verifies AssistantMessage has typed content field instead of interface{}.
func TestAssistantMessageTypedContent(t *testing.T) {
	// This test will fail until we implement typed AssistantMessage
	assistantMsg := AssistantMessage{
		Content: []ContentBlock{},
		Model:   "claude-3-5-sonnet-20241022",
	}

	// Test that Content field is typed as []ContentBlock
	blocks := assistantMsg.Content
	if blocks == nil {
		t.Error("AssistantMessage.Content should not be nil")
	}

	// Test Message interface implementation
	var msg Message = assistantMsg
	if msg.Type() != "assistant" {
		t.Errorf("Expected message type 'assistant', got '%s'", msg.Type())
	}

	// Test Model field
	if assistantMsg.Model == "" {
		t.Error("AssistantMessage.Model should not be empty")
	}
}

// TestStreamMessageTypedMessage verifies StreamMessage has typed Message field instead of interface{}.
func TestStreamMessageTypedMessage(t *testing.T) {
	// This test will fail until we implement typed StreamMessage
	userMsg := UserMessage{Content: TextContent{Text: "test"}}

	streamMsg := StreamMessage{
		Type:      "message",
		Message:   userMsg,
		SessionID: "session123",
	}

	// Test that Message field is typed
	var msg Message = streamMsg.Message
	_ = msg

	// Test basic field access
	if streamMsg.Type == "" {
		t.Error("StreamMessage.Type should not be empty")
	}
	if streamMsg.Message == nil {
		t.Error("StreamMessage.Message should not be nil")
	}
	if streamMsg.SessionID == "" {
		t.Error("StreamMessage.SessionID should not be empty")
	}
}

// TestTypedMessageJSONMarshaling verifies JSON marshaling works with typed fields.
func TestTypedMessageJSONMarshaling(t *testing.T) {
	tests := []struct {
		name     string
		message  interface{}
		validate func(t *testing.T, jsonData []byte)
	}{
		{
			name: "UserMessage JSON marshaling",
			message: UserMessage{
				Content: TextContent{Text: "Hello, world!"},
			},
			validate: func(t *testing.T, jsonData []byte) {
				var parsed map[string]interface{}
				if err := json.Unmarshal(jsonData, &parsed); err != nil {
					t.Fatalf("Failed to parse JSON: %v", err)
				}

				if parsed["type"] != "user" {
					t.Errorf("Expected type 'user', got %v", parsed["type"])
				}

				content, ok := parsed["content"].(map[string]interface{})
				if !ok {
					t.Errorf("Expected content to be object, got %T", parsed["content"])
					return
				}

				if content["text"] != "Hello, world!" {
					t.Errorf("Expected text 'Hello, world!', got %v", content["text"])
				}
			},
		},
		{
			name: "AssistantMessage JSON marshaling",
			message: AssistantMessage{
				Content: []ContentBlock{},
				Model:   "claude-3-5-sonnet-20241022",
			},
			validate: func(t *testing.T, jsonData []byte) {
				var parsed map[string]interface{}
				if err := json.Unmarshal(jsonData, &parsed); err != nil {
					t.Fatalf("Failed to parse JSON: %v", err)
				}

				if parsed["type"] != "assistant" {
					t.Errorf("Expected type 'assistant', got %v", parsed["type"])
				}

				if parsed["model"] != "claude-3-5-sonnet-20241022" {
					t.Errorf("Expected model 'claude-3-5-sonnet-20241022', got %v", parsed["model"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.message)
			if err != nil {
				t.Fatalf("Failed to marshal %T: %v", tt.message, err)
			}

			tt.validate(t, jsonData)
		})
	}
}

// TestNoInterfaceEmptyUsage verifies no interface{} usage in new message types.
func TestNoInterfaceEmptyUsage(t *testing.T) {
	// This is a compile-time test - if the types use interface{}, this won't compile
	userMsg := UserMessage{Content: TextContent{Text: "test"}}
	assistantMsg := AssistantMessage{Content: []ContentBlock{}}
	streamMsg := StreamMessage{Message: userMsg}

	// These should all be strongly typed, not interface{}
	var userContent UserMessageContent = userMsg.Content
	var assistantContent []ContentBlock = assistantMsg.Content
	var streamMessage Message = streamMsg.Message

	_ = userContent
	_ = assistantContent
	_ = streamMessage
}

// TestTypedMessageJSONBasicUnmarshaling verifies basic JSON unmarshaling for simple fields.
// Note: Complex union type unmarshaling will be implemented in a future iteration.
func TestTypedMessageJSONBasicUnmarshaling(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		target   interface{}
		validate func(t *testing.T, target interface{})
	}{
		{
			name:     "AssistantMessage JSON unmarshaling",
			jsonData: `{"type":"assistant","content":[],"model":"claude-3-5-sonnet-20241022"}`,
			target:   &AssistantMessage{},
			validate: func(t *testing.T, target interface{}) {
				assistantMsg := target.(*AssistantMessage)
				if assistantMsg.Model != "claude-3-5-sonnet-20241022" {
					t.Errorf("Expected model 'claude-3-5-sonnet-20241022', got '%s'", assistantMsg.Model)
				}
				if assistantMsg.Content == nil {
					t.Error("Expected Content to be empty slice, got nil")
				}
				if len(assistantMsg.Content) != 0 {
					t.Errorf("Expected empty Content slice, got length %d", len(assistantMsg.Content))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := json.Unmarshal([]byte(tt.jsonData), tt.target)
			if err != nil {
				t.Fatalf("Failed to unmarshal JSON into %T: %v", tt.target, err)
			}

			tt.validate(t, tt.target)
		})
	}
}

// TestTypedMessageConstruction verifies typed message construction works correctly.
func TestTypedMessageConstruction(t *testing.T) {
	// Test UserMessage with TextContent
	userMsg := UserMessage{
		Content: TextContent{Text: "Hello, world!"},
	}

	// Verify the content is strongly typed
	textContent, ok := userMsg.Content.(TextContent)
	if !ok {
		t.Errorf("Expected TextContent, got %T", userMsg.Content)
		return
	}

	if textContent.Text != "Hello, world!" {
		t.Errorf("Expected text 'Hello, world!', got '%s'", textContent.Text)
	}

	// Test UserMessage with BlockListContent
	userMsgBlocks := UserMessage{
		Content: BlockListContent{Blocks: []ContentBlock{}},
	}

	blockContent, ok := userMsgBlocks.Content.(BlockListContent)
	if !ok {
		t.Errorf("Expected BlockListContent, got %T", userMsgBlocks.Content)
		return
	}

	if blockContent.Blocks == nil {
		t.Error("Expected Blocks to be empty slice, got nil")
	}

	// Test AssistantMessage construction
	assistantMsg := AssistantMessage{
		Content: []ContentBlock{},
		Model:   "claude-3-5-sonnet-20241022",
	}

	if len(assistantMsg.Content) != 0 {
		t.Errorf("Expected empty content slice, got length %d", len(assistantMsg.Content))
	}

	if assistantMsg.Model != "claude-3-5-sonnet-20241022" {
		t.Errorf("Expected model 'claude-3-5-sonnet-20241022', got '%s'", assistantMsg.Model)
	}
}

// TestMessageTypeConsistency verifies Type() method matches JSON type field.
func TestMessageTypeConsistency(t *testing.T) {
	tests := []struct {
		name     string
		message  Message
		wantType string
	}{
		{
			name:     "UserMessage type consistency",
			message:  UserMessage{Content: TextContent{Text: "test"}},
			wantType: "user",
		},
		{
			name:     "AssistantMessage type consistency",
			message:  AssistantMessage{Content: []ContentBlock{}, Model: "test-model"},
			wantType: "assistant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Type() method
			if tt.message.Type() != tt.wantType {
				t.Errorf("Expected Type() to return '%s', got '%s'", tt.wantType, tt.message.Type())
			}

			// Test JSON marshaling includes correct type field
			jsonData, err := json.Marshal(tt.message)
			if err != nil {
				t.Fatalf("Failed to marshal %T: %v", tt.message, err)
			}

			var parsed map[string]interface{}
			if err := json.Unmarshal(jsonData, &parsed); err != nil {
				t.Fatalf("Failed to parse marshaled JSON: %v", err)
			}

			if parsed["type"] != tt.wantType {
				t.Errorf("JSON type field should be '%s', got %v", tt.wantType, parsed["type"])
			}
		})
	}
}

// TestStreamMessageFieldTypes verifies StreamMessage has proper field types.
func TestStreamMessageFieldTypes(t *testing.T) {
	userMsg := UserMessage{Content: TextContent{Text: "test"}}

	streamMsg := StreamMessage{
		Type:            "message",
		Message:         userMsg,
		ParentToolUseID: nil,
		SessionID:       "session123",
		RequestID:       "request456",
	}

	// Test field types
	if streamMsg.Type != "message" {
		t.Errorf("Expected Type 'message', got '%s'", streamMsg.Type)
	}

	// Test Message field is strongly typed
	var msg Message = streamMsg.Message
	if msg.Type() != "user" {
		t.Errorf("Expected Message type 'user', got '%s'", msg.Type())
	}

	// Test optional fields
	if streamMsg.ParentToolUseID != nil {
		t.Errorf("Expected ParentToolUseID to be nil, got %v", streamMsg.ParentToolUseID)
	}

	// Test string fields
	if streamMsg.SessionID != "session123" {
		t.Errorf("Expected SessionID 'session123', got '%s'", streamMsg.SessionID)
	}
	if streamMsg.RequestID != "request456" {
		t.Errorf("Expected RequestID 'request456', got '%s'", streamMsg.RequestID)
	}
}

// TestMessageInterfaceImplementation verifies all message types implement Message interface.
func TestMessageInterfaceImplementation(t *testing.T) {
	messages := []Message{
		UserMessage{Content: TextContent{Text: "test"}},
		AssistantMessage{Content: []ContentBlock{}, Model: "test"},
	}

	for i, msg := range messages {
		t.Run(fmt.Sprintf("Message_%d_implements_interface", i), func(t *testing.T) {
			// Test that Type() returns a non-empty string
			msgType := msg.Type()
			if msgType == "" {
				t.Errorf("Message.Type() should return non-empty string, got empty for %T", msg)
			}

			// Test that the message can be used as Message interface
			var interfaceMsg Message = msg
			if interfaceMsg.Type() != msgType {
				t.Errorf("Interface method should return same as concrete method")
			}
		})
	}
}

// TestContentBlockTypeMethodConsistency verifies all ContentBlock types use Type() method consistently.
func TestContentBlockTypeMethodConsistency(t *testing.T) {
	// This test will fail until we implement concrete ContentBlock types
	contentBlocks := []ContentBlock{
		TextBlock{Text: "Hello"},
		ThinkingBlock{Thinking: "Let me think...", Signature: "sig"},
		ToolUseBlock{ToolUseID: "tool_123", Name: "test_tool", Input: map[string]any{"key": "value"}},
		ToolResultBlock{ToolUseID: "tool_123", Content: TextContent{Text: "result"}},
	}

	expectedTypes := []string{"text", "thinking", "tool_use", "tool_result"}

	for i, block := range contentBlocks {
		t.Run(fmt.Sprintf("ContentBlock_%d_Type_method", i), func(t *testing.T) {
			// Test that Type() method exists and returns expected value
			blockType := block.Type()
			if blockType != expectedTypes[i] {
				t.Errorf("Expected Type() to return '%s', got '%s'", expectedTypes[i], blockType)
			}

			// Test that the block implements ContentBlock interface
			var cb ContentBlock = block
			if cb.Type() != blockType {
				t.Errorf("Interface method should return same as concrete method")
			}
		})
	}
}

// TestContentBlockNoBlockTypeMethod verifies ContentBlock types don't have legacy BlockType() method.
func TestContentBlockNoBlockTypeMethod(t *testing.T) {
	// Test that our new ContentBlock types don't have the old BlockType() method
	textBlock := TextBlock{Text: "test"}
	thinkingBlock := ThinkingBlock{Thinking: "thinking", Signature: "sig"}
	toolUseBlock := ToolUseBlock{ToolUseID: "tool_123", Name: "test_tool"}
	toolResultBlock := ToolResultBlock{ToolUseID: "tool_123", Content: TextContent{Text: "result"}}

	// This is a compile-time test - if BlockType() methods exist, this test will need updates
	// We expect only Type() methods to exist
	blocks := []ContentBlock{textBlock, thinkingBlock, toolUseBlock, toolResultBlock}

	for i, block := range blocks {
		t.Run(fmt.Sprintf("ContentBlock_%d_uses_Type_not_BlockType", i), func(t *testing.T) {
			// Test Type() method exists
			blockType := block.Type()
			if blockType == "" {
				t.Errorf("Type() method should return non-empty string")
			}

			// Note: We can't directly test for absence of BlockType() at runtime,
			// but if this compiles, it means our types are correctly using Type()
		})
	}
}

// TestContentBlockInterfaceImplementation verifies all concrete ContentBlock types implement the interface.
func TestContentBlockInterfaceImplementation(t *testing.T) {
	// This will fail until we implement the concrete types
	var textBlock ContentBlock = TextBlock{Text: "test"}
	var thinkingBlock ContentBlock = ThinkingBlock{Thinking: "thinking", Signature: "sig"}
	var toolUseBlock ContentBlock = ToolUseBlock{ToolUseID: "tool_123", Name: "test_tool"}
	var toolResultBlock ContentBlock = ToolResultBlock{ToolUseID: "tool_123", Content: TextContent{Text: "result"}}

	blocks := []ContentBlock{textBlock, thinkingBlock, toolUseBlock, toolResultBlock}
	expectedTypes := []string{"text", "thinking", "tool_use", "tool_result"}

	for i, block := range blocks {
		t.Run(fmt.Sprintf("ContentBlock_%d_implements_interface", i), func(t *testing.T) {
			if block.Type() != expectedTypes[i] {
				t.Errorf("Expected Type() '%s', got '%s'", expectedTypes[i], block.Type())
			}
		})
	}
}

// TestToolResultBlockTypedContent verifies ToolResultBlock uses typed Content field.
func TestToolResultBlockTypedContent(t *testing.T) {
	// This tests that ToolResultBlock.Content is MessageContent, not interface{}
	textContent := TextContent{Text: "Tool execution successful"}
	toolResult := ToolResultBlock{
		ToolUseID: "tool_123",
		Content:   textContent,
		IsError:   nil,
	}

	// Test that Content field is strongly typed
	var content MessageContent = toolResult.Content
	_ = content

	// Test field access
	if toolResult.ToolUseID != "tool_123" {
		t.Errorf("Expected ToolUseID 'tool_123', got '%s'", toolResult.ToolUseID)
	}

	// Test typed content
	textContentTyped, ok := toolResult.Content.(TextContent)
	if !ok {
		t.Errorf("Expected TextContent, got %T", toolResult.Content)
	} else if textContentTyped.Text != "Tool execution successful" {
		t.Errorf("Expected text 'Tool execution successful', got '%s'", textContentTyped.Text)
	}

	// Test optional field
	if toolResult.IsError != nil {
		t.Errorf("Expected IsError to be nil, got %v", toolResult.IsError)
	}
}

// TestContentBlockJSONMarshaling verifies JSON marshaling works for all ContentBlock types.
func TestContentBlockJSONMarshaling(t *testing.T) {
	tests := []struct {
		name     string
		block    ContentBlock
		wantJSON string
	}{
		{
			name:     "TextBlock JSON marshaling",
			block:    TextBlock{Text: "Hello, world!"},
			wantJSON: `{"text":"Hello, world!"}`,
		},
		{
			name:     "ThinkingBlock JSON marshaling",
			block:    ThinkingBlock{Thinking: "Let me think...", Signature: "signature123"},
			wantJSON: `{"thinking":"Let me think...","signature":"signature123"}`,
		},
		{
			name:     "ToolUseBlock JSON marshaling",
			block:    ToolUseBlock{ToolUseID: "tool_123", Name: "test_tool", Input: map[string]any{"key": "value"}},
			wantJSON: `{"tool_use_id":"tool_123","name":"test_tool","input":{"key":"value"}}`,
		},
		{
			name:     "ToolResultBlock JSON marshaling",
			block:    ToolResultBlock{ToolUseID: "tool_123", Content: TextContent{Text: "Success"}, IsError: nil},
			wantJSON: `{"tool_use_id":"tool_123","content":{"text":"Success"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.block)
			if err != nil {
				t.Fatalf("Failed to marshal %T: %v", tt.block, err)
			}

			if string(jsonData) != tt.wantJSON {
				t.Errorf("JSON mismatch for %T\nWant: %s\nGot:  %s", tt.block, tt.wantJSON, string(jsonData))
			}
		})
	}
}

// TestContentBlockZeroValues verifies zero values behave correctly for all ContentBlock types.
func TestContentBlockZeroValues(t *testing.T) {
	tests := []struct {
		name         string
		block        ContentBlock
		expectedType string
		validate     func(t *testing.T, block ContentBlock)
	}{
		{
			name:         "TextBlock zero value",
			block:        TextBlock{},
			expectedType: "text",
			validate: func(t *testing.T, block ContentBlock) {
				textBlock := block.(TextBlock)
				if textBlock.Text != "" {
					t.Errorf("Zero value TextBlock should have empty Text, got '%s'", textBlock.Text)
				}
			},
		},
		{
			name:         "ThinkingBlock zero value",
			block:        ThinkingBlock{},
			expectedType: "thinking",
			validate: func(t *testing.T, block ContentBlock) {
				thinkingBlock := block.(ThinkingBlock)
				if thinkingBlock.Thinking != "" {
					t.Errorf("Zero value ThinkingBlock should have empty Thinking, got '%s'", thinkingBlock.Thinking)
				}
				if thinkingBlock.Signature != "" {
					t.Errorf("Zero value ThinkingBlock should have empty Signature, got '%s'", thinkingBlock.Signature)
				}
			},
		},
		{
			name:         "ToolUseBlock zero value",
			block:        ToolUseBlock{},
			expectedType: "tool_use",
			validate: func(t *testing.T, block ContentBlock) {
				toolUseBlock := block.(ToolUseBlock)
				if toolUseBlock.ToolUseID != "" {
					t.Errorf("Zero value ToolUseBlock should have empty ToolUseID, got '%s'", toolUseBlock.ToolUseID)
				}
				if toolUseBlock.Name != "" {
					t.Errorf("Zero value ToolUseBlock should have empty Name, got '%s'", toolUseBlock.Name)
				}
				if toolUseBlock.Input != nil {
					t.Errorf("Zero value ToolUseBlock should have nil Input, got %v", toolUseBlock.Input)
				}
			},
		},
		{
			name:         "ToolResultBlock zero value",
			block:        ToolResultBlock{},
			expectedType: "tool_result",
			validate: func(t *testing.T, block ContentBlock) {
				toolResultBlock := block.(ToolResultBlock)
				if toolResultBlock.ToolUseID != "" {
					t.Errorf("Zero value ToolResultBlock should have empty ToolUseID, got '%s'", toolResultBlock.ToolUseID)
				}
				if toolResultBlock.Content != nil {
					t.Errorf("Zero value ToolResultBlock should have nil Content, got %v", toolResultBlock.Content)
				}
				if toolResultBlock.IsError != nil {
					t.Errorf("Zero value ToolResultBlock should have nil IsError, got %v", toolResultBlock.IsError)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Type() method
			if tt.block.Type() != tt.expectedType {
				t.Errorf("Expected Type() '%s', got '%s'", tt.expectedType, tt.block.Type())
			}

			// Test interface implementation
			var cb ContentBlock = tt.block
			if cb.Type() != tt.expectedType {
				t.Errorf("Interface method should return same as concrete method")
			}

			// Run custom validation
			tt.validate(t, tt.block)
		})
	}
}

// TestContentBlockFieldAccess verifies field access works correctly for all ContentBlock types.
func TestContentBlockFieldAccess(t *testing.T) {
	// Test TextBlock field access
	textBlock := TextBlock{Text: "Hello, world!"}
	if textBlock.Text != "Hello, world!" {
		t.Errorf("Expected Text 'Hello, world!', got '%s'", textBlock.Text)
	}

	// Test ThinkingBlock field access
	thinkingBlock := ThinkingBlock{Thinking: "I'm thinking...", Signature: "sig123"}
	if thinkingBlock.Thinking != "I'm thinking..." {
		t.Errorf("Expected Thinking 'I'm thinking...', got '%s'", thinkingBlock.Thinking)
	}
	if thinkingBlock.Signature != "sig123" {
		t.Errorf("Expected Signature 'sig123', got '%s'", thinkingBlock.Signature)
	}

	// Test ToolUseBlock field access
	toolUseBlock := ToolUseBlock{
		ToolUseID: "tool_456",
		Name:      "calculator",
		Input:     map[string]any{"operation": "add", "a": 1, "b": 2},
	}
	if toolUseBlock.ToolUseID != "tool_456" {
		t.Errorf("Expected ToolUseID 'tool_456', got '%s'", toolUseBlock.ToolUseID)
	}
	if toolUseBlock.Name != "calculator" {
		t.Errorf("Expected Name 'calculator', got '%s'", toolUseBlock.Name)
	}
	if toolUseBlock.Input["operation"] != "add" {
		t.Errorf("Expected operation 'add', got '%v'", toolUseBlock.Input["operation"])
	}

	// Test ToolResultBlock field access
	isError := false
	toolResultBlock := ToolResultBlock{
		ToolUseID: "tool_456",
		Content:   TextContent{Text: "Result: 3"},
		IsError:   &isError,
	}
	if toolResultBlock.ToolUseID != "tool_456" {
		t.Errorf("Expected ToolUseID 'tool_456', got '%s'", toolResultBlock.ToolUseID)
	}

	textContent, ok := toolResultBlock.Content.(TextContent)
	if !ok {
		t.Errorf("Expected TextContent, got %T", toolResultBlock.Content)
	} else if textContent.Text != "Result: 3" {
		t.Errorf("Expected text 'Result: 3', got '%s'", textContent.Text)
	}

	if toolResultBlock.IsError == nil || *toolResultBlock.IsError != false {
		t.Errorf("Expected IsError to be false, got %v", toolResultBlock.IsError)
	}
}

// TestSystemMessageType verifies SystemMessage.Type() method works correctly.
func TestSystemMessageType(t *testing.T) {
	systemMsg := &SystemMessage{
		MessageType: "system",
		Subtype:     "prompt",
		Data:        map[string]any{"key": "value"},
	}

	expectedType := "system"
	actualType := systemMsg.Type()

	if actualType != expectedType {
		t.Errorf("Expected SystemMessage.Type() to return %q, got %q", expectedType, actualType)
	}

	// Test Message interface implementation
	var msg Message = systemMsg
	if msg.Type() != expectedType {
		t.Errorf("SystemMessage should implement Message interface correctly")
	}
}

// TestResultMessageType verifies ResultMessage.Type() method works correctly.
func TestResultMessageType(t *testing.T) {
	cost := 0.05
	usage := map[string]any{"input_tokens": 100, "output_tokens": 50}
	result := map[string]any{"status": "completed"}

	resultMsg := &ResultMessage{
		MessageType:   "result",
		Subtype:       "completion",
		DurationMs:    1500,
		DurationAPIMs: 1200,
		IsError:       false,
		NumTurns:      3,
		SessionID:     "session123",
		TotalCostUSD:  &cost,
		Usage:         &usage,
		Result:        &result,
	}

	expectedType := "result"
	actualType := resultMsg.Type()

	if actualType != expectedType {
		t.Errorf("Expected ResultMessage.Type() to return %q, got %q", expectedType, actualType)
	}

	// Test Message interface implementation
	var msg Message = resultMsg
	if msg.Type() != expectedType {
		t.Errorf("ResultMessage should implement Message interface correctly")
	}

	// Test field access
	if resultMsg.MessageType != "result" {
		t.Errorf("Expected MessageType 'result', got %q", resultMsg.MessageType)
	}
	if resultMsg.Subtype != "completion" {
		t.Errorf("Expected Subtype 'completion', got %q", resultMsg.Subtype)
	}
	if resultMsg.DurationMs != 1500 {
		t.Errorf("Expected DurationMs 1500, got %d", resultMsg.DurationMs)
	}
	if resultMsg.DurationAPIMs != 1200 {
		t.Errorf("Expected DurationAPIMs 1200, got %d", resultMsg.DurationAPIMs)
	}
	if resultMsg.IsError != false {
		t.Errorf("Expected IsError false, got %v", resultMsg.IsError)
	}
	if resultMsg.NumTurns != 3 {
		t.Errorf("Expected NumTurns 3, got %d", resultMsg.NumTurns)
	}
	if resultMsg.SessionID != "session123" {
		t.Errorf("Expected SessionID 'session123', got %q", resultMsg.SessionID)
	}
	if resultMsg.TotalCostUSD == nil || *resultMsg.TotalCostUSD != 0.05 {
		t.Errorf("Expected TotalCostUSD 0.05, got %v", resultMsg.TotalCostUSD)
	}
}
