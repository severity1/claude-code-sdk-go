package interfaces

import (
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
