package interfaces

import (
	"encoding/json"
	"reflect"
	"testing"
)

// TestContentInterfaceExistence verifies that the content interfaces exist and have correct method signatures.
func TestContentInterfaceExistence(t *testing.T) {
	tests := []struct {
		name            string
		interfaceType   reflect.Type
		expectedMethods []string
	}{
		{
			name:            "MessageContent interface",
			interfaceType:   reflect.TypeOf((*MessageContent)(nil)).Elem(),
			expectedMethods: []string{"messageContent"},
		},
		{
			name:            "UserMessageContent interface",
			interfaceType:   reflect.TypeOf((*UserMessageContent)(nil)).Elem(),
			expectedMethods: []string{"messageContent", "userMessageContent"},
		},
		{
			name:            "AssistantMessageContent interface",
			interfaceType:   reflect.TypeOf((*AssistantMessageContent)(nil)).Elem(),
			expectedMethods: []string{"messageContent", "assistantMessageContent"},
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

// TestSealedInterfacePattern verifies that sealed interfaces work correctly.
func TestSealedInterfacePattern(t *testing.T) {
	// This test will verify that our sealed interface pattern prevents external implementations
	// We'll test this by ensuring the sealing methods exist and are unexported
	messageContentType := reflect.TypeOf((*MessageContent)(nil)).Elem()

	// Check that the sealing method exists
	sealingMethod, found := messageContentType.MethodByName("messageContent")
	if !found {
		t.Error("MessageContent interface must have messageContent sealing method")
		return
	}

	// Verify method signature - should take no parameters and return nothing
	methodType := sealingMethod.Type
	if methodType.NumIn() != 0 { // interface methods don't count receiver
		t.Errorf("messageContent method should have 0 parameters, got %d", methodType.NumIn())
	}
	if methodType.NumOut() != 0 {
		t.Errorf("messageContent method should return nothing, got %d return values", methodType.NumOut())
	}
}

// TestContentInterfaceEmbedding verifies interface embedding works correctly.
func TestContentInterfaceEmbedding(t *testing.T) {
	userContentType := reflect.TypeOf((*UserMessageContent)(nil)).Elem()
	assistantContentType := reflect.TypeOf((*AssistantMessageContent)(nil)).Elem()

	// Check that UserMessageContent embeds MessageContent
	hasMessageContentMethod := false
	for i := 0; i < userContentType.NumMethod(); i++ {
		if userContentType.Method(i).Name == "messageContent" {
			hasMessageContentMethod = true
			break
		}
	}
	if !hasMessageContentMethod {
		t.Error("UserMessageContent should embed MessageContent interface")
	}

	// Check that AssistantMessageContent embeds MessageContent
	hasMessageContentMethod = false
	for i := 0; i < assistantContentType.NumMethod(); i++ {
		if assistantContentType.Method(i).Name == "messageContent" {
			hasMessageContentMethod = true
			break
		}
	}
	if !hasMessageContentMethod {
		t.Error("AssistantMessageContent should embed MessageContent interface")
	}
}

// TestContentInterfaceNilHandling verifies interfaces handle nil values correctly.
func TestContentInterfaceNilHandling(t *testing.T) {
	var messageContent MessageContent
	var userContent UserMessageContent
	var assistantContent AssistantMessageContent

	// Verify nil interfaces are properly typed
	if messageContent != nil {
		t.Error("Nil MessageContent interface should be nil")
	}
	if userContent != nil {
		t.Error("Nil UserMessageContent interface should be nil")
	}
	if assistantContent != nil {
		t.Error("Nil AssistantMessageContent interface should be nil")
	}

	// Test interface assignment compatibility
	messageContent = userContent // Should compile - UserMessageContent embeds MessageContent
	_ = messageContent
}

// TestTextContentExistence verifies TextContent struct exists and implements required interfaces.
func TestTextContentExistence(t *testing.T) {
	// This test will fail until we implement TextContent
	var textContent TextContent

	// Test that TextContent implements MessageContent
	var messageContent MessageContent = textContent
	_ = messageContent

	// Test that TextContent implements UserMessageContent
	var userContent UserMessageContent = textContent
	_ = userContent

	// Test basic field access
	if textContent.Text != "" {
		t.Error("Zero value TextContent should have empty Text field")
	}
}

// TestBlockListContentExistence verifies BlockListContent struct exists and implements required interfaces.
func TestBlockListContentExistence(t *testing.T) {
	// This test will fail until we implement BlockListContent
	var blockContent BlockListContent

	// Test that BlockListContent implements MessageContent
	var messageContent MessageContent = blockContent
	_ = messageContent

	// Test that BlockListContent implements UserMessageContent
	var userContent UserMessageContent = blockContent
	_ = userContent

	// Test basic field access
	if blockContent.Blocks != nil {
		t.Error("Zero value BlockListContent should have nil Blocks field")
	}
}

// TestThinkingContentExistence verifies ThinkingContent struct exists and implements required interfaces.
func TestThinkingContentExistence(t *testing.T) {
	// This test will fail until we implement ThinkingContent
	var thinkingContent ThinkingContent

	// Test that ThinkingContent implements MessageContent
	var messageContent MessageContent = thinkingContent
	_ = messageContent

	// Test that ThinkingContent implements AssistantMessageContent
	var assistantContent AssistantMessageContent = thinkingContent
	_ = assistantContent

	// Test basic field access
	if thinkingContent.Thinking != "" {
		t.Error("Zero value ThinkingContent should have empty Thinking field")
	}
	if thinkingContent.Signature != "" {
		t.Error("Zero value ThinkingContent should have empty Signature field")
	}
}

// TestConcreteContentTypeCompilation verifies all concrete types compile and have proper sealing methods.
func TestConcreteContentTypeCompilation(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func()
	}{
		{
			name: "TextContent sealing methods",
			testFunc: func() {
				var content TextContent
				content.messageContent()
				content.userMessageContent()
			},
		},
		{
			name: "BlockListContent sealing methods",
			testFunc: func() {
				var content BlockListContent
				content.messageContent()
				content.userMessageContent()
			},
		},
		{
			name: "ThinkingContent sealing methods",
			testFunc: func() {
				var content ThinkingContent
				content.messageContent()
				content.assistantMessageContent()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// If this compiles, the sealing methods exist
			tt.testFunc()
		})
	}
}

// TestConcreteContentTypeJSONMarshaling verifies JSON marshaling works correctly for all concrete types.
func TestConcreteContentTypeJSONMarshaling(t *testing.T) {
	tests := []struct {
		name     string
		content  interface{}
		wantJSON string
	}{
		{
			name:     "TextContent JSON marshaling",
			content:  TextContent{Text: "Hello, world!"},
			wantJSON: `{"text":"Hello, world!"}`,
		},
		{
			name:     "ThinkingContent JSON marshaling",
			content:  ThinkingContent{Thinking: "Let me think...", Signature: "signature123"},
			wantJSON: `{"thinking":"Let me think...","signature":"signature123"}`,
		},
		{
			name:     "BlockListContent empty blocks",
			content:  BlockListContent{Blocks: []ContentBlock{}},
			wantJSON: `{"blocks":[]}`,
		},
		{
			name:     "BlockListContent nil blocks",
			content:  BlockListContent{Blocks: nil},
			wantJSON: `{"blocks":null}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.content)
			if err != nil {
				t.Fatalf("Failed to marshal %T: %v", tt.content, err)
			}

			if string(jsonData) != tt.wantJSON {
				t.Errorf("JSON mismatch for %T\nWant: %s\nGot:  %s", tt.content, tt.wantJSON, string(jsonData))
			}
		})
	}
}

// TestConcreteContentTypeJSONUnmarshaling verifies JSON unmarshaling works correctly.
func TestConcreteContentTypeJSONUnmarshaling(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		target   interface{}
		validate func(t *testing.T, target interface{})
	}{
		{
			name:     "TextContent JSON unmarshaling",
			jsonData: `{"text":"Hello, world!"}`,
			target:   &TextContent{},
			validate: func(t *testing.T, target interface{}) {
				content := target.(*TextContent)
				if content.Text != "Hello, world!" {
					t.Errorf("Expected Text 'Hello, world!', got '%s'", content.Text)
				}
			},
		},
		{
			name:     "ThinkingContent JSON unmarshaling",
			jsonData: `{"thinking":"Let me think...","signature":"signature123"}`,
			target:   &ThinkingContent{},
			validate: func(t *testing.T, target interface{}) {
				content := target.(*ThinkingContent)
				if content.Thinking != "Let me think..." {
					t.Errorf("Expected Thinking 'Let me think...', got '%s'", content.Thinking)
				}
				if content.Signature != "signature123" {
					t.Errorf("Expected Signature 'signature123', got '%s'", content.Signature)
				}
			},
		},
		{
			name:     "BlockListContent JSON unmarshaling",
			jsonData: `{"blocks":[]}`,
			target:   &BlockListContent{},
			validate: func(t *testing.T, target interface{}) {
				content := target.(*BlockListContent)
				if content.Blocks == nil {
					t.Error("Expected Blocks to be empty slice, got nil")
				}
				if len(content.Blocks) != 0 {
					t.Errorf("Expected empty Blocks slice, got length %d", len(content.Blocks))
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

// TestSealedInterfaceAssignability verifies interface assignment works correctly.
func TestSealedInterfaceAssignability(t *testing.T) {
	// Test TextContent assignability
	textContent := TextContent{Text: "test"}

	var messageContent MessageContent = textContent
	var userContent UserMessageContent = textContent

	// Test assignment chain
	messageContent = userContent // UserMessageContent -> MessageContent
	_ = messageContent

	// Test ThinkingContent assignability
	thinkingContent := ThinkingContent{Thinking: "thinking", Signature: "sig"}

	var messageContent2 MessageContent = thinkingContent
	var assistantContent AssistantMessageContent = thinkingContent

	// Test assignment chain
	messageContent2 = assistantContent // AssistantMessageContent -> MessageContent
	_ = messageContent2

	// Test BlockListContent assignability
	blockContent := BlockListContent{Blocks: []ContentBlock{}}

	var messageContent3 MessageContent = blockContent
	var userContent2 UserMessageContent = blockContent

	// Test assignment chain
	messageContent3 = userContent2 // UserMessageContent -> MessageContent
	_ = messageContent3
}

// TestInterfaceEmbeddingBehavior verifies that interface embedding works as expected.
func TestInterfaceEmbeddingBehavior(t *testing.T) {
	// Test that UserMessageContent methods include MessageContent methods
	userContentType := reflect.TypeOf((*UserMessageContent)(nil)).Elem()

	methods := make(map[string]bool)
	for i := 0; i < userContentType.NumMethod(); i++ {
		method := userContentType.Method(i)
		methods[method.Name] = true
	}

	expectedMethods := []string{"messageContent", "userMessageContent"}
	for _, expectedMethod := range expectedMethods {
		if !methods[expectedMethod] {
			t.Errorf("UserMessageContent should have method %s", expectedMethod)
		}
	}

	// Test that AssistantMessageContent methods include MessageContent methods
	assistantContentType := reflect.TypeOf((*AssistantMessageContent)(nil)).Elem()

	methods = make(map[string]bool)
	for i := 0; i < assistantContentType.NumMethod(); i++ {
		method := assistantContentType.Method(i)
		methods[method.Name] = true
	}

	expectedMethods = []string{"messageContent", "assistantMessageContent"}
	for _, expectedMethod := range expectedMethods {
		if !methods[expectedMethod] {
			t.Errorf("AssistantMessageContent should have method %s", expectedMethod)
		}
	}
}

// TestConcreteTypeZeroValues verifies zero values behave correctly.
func TestConcreteTypeZeroValues(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "TextContent zero value",
			testFunc: func(t *testing.T) {
				var content TextContent
				if content.Text != "" {
					t.Errorf("Zero value TextContent should have empty Text, got '%s'", content.Text)
				}

				// Should still implement interfaces
				var _ MessageContent = content
				var _ UserMessageContent = content
			},
		},
		{
			name: "ThinkingContent zero value",
			testFunc: func(t *testing.T) {
				var content ThinkingContent
				if content.Thinking != "" {
					t.Errorf("Zero value ThinkingContent should have empty Thinking, got '%s'", content.Thinking)
				}
				if content.Signature != "" {
					t.Errorf("Zero value ThinkingContent should have empty Signature, got '%s'", content.Signature)
				}

				// Should still implement interfaces
				var _ MessageContent = content
				var _ AssistantMessageContent = content
			},
		},
		{
			name: "BlockListContent zero value",
			testFunc: func(t *testing.T) {
				var content BlockListContent
				if content.Blocks != nil {
					t.Errorf("Zero value BlockListContent should have nil Blocks, got %v", content.Blocks)
				}

				// Should still implement interfaces
				var _ MessageContent = content
				var _ UserMessageContent = content
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.testFunc)
	}
}
