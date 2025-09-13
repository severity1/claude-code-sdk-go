package interfaces

import "encoding/json"

// Message represents any message type in the Claude Code protocol.
type Message interface {
	Type() string
}

// ContentBlock represents any content block within a message.
// Uses consistent Type() method naming instead of BlockType().
type ContentBlock interface {
	Type() string
}

// UserMessage represents a message from the user with strongly-typed content.
type UserMessage struct {
	Content UserMessageContent `json:"content"`
}

// Type returns the message type for UserMessage.
func (m UserMessage) Type() string {
	return "user"
}

// MarshalJSON implements custom JSON marshaling for UserMessage.
func (m UserMessage) MarshalJSON() ([]byte, error) {
	type userMessage UserMessage
	temp := struct {
		Type string `json:"type"`
		*userMessage
	}{
		Type:        "user",
		userMessage: (*userMessage)(&m),
	}
	return json.Marshal(temp)
}

// AssistantMessage represents a message from the assistant with strongly-typed content.
type AssistantMessage struct {
	Content []ContentBlock `json:"content"`
	Model   string         `json:"model"`
}

// Type returns the message type for AssistantMessage.
func (m AssistantMessage) Type() string {
	return "assistant"
}

// MarshalJSON implements custom JSON marshaling for AssistantMessage.
func (m AssistantMessage) MarshalJSON() ([]byte, error) {
	type assistantMessage AssistantMessage
	temp := struct {
		Type string `json:"type"`
		*assistantMessage
	}{
		Type:             "assistant",
		assistantMessage: (*assistantMessage)(&m),
	}
	return json.Marshal(temp)
}

// StreamMessage represents messages sent to the CLI for streaming communication with typed Message field.
type StreamMessage struct {
	Type            string  `json:"type"`
	Message         Message `json:"message,omitempty"`
	ParentToolUseID *string `json:"parent_tool_use_id,omitempty"`
	SessionID       string  `json:"session_id,omitempty"`
	RequestID       string  `json:"request_id,omitempty"`
	// Note: Request and Response fields will be handled in a future iteration with proper typing
}

// TextBlock represents text content within a message.
type TextBlock struct {
	Text string `json:"text"`
}

// Type returns the content block type for TextBlock.
func (b TextBlock) Type() string {
	return "text"
}

// ThinkingBlock represents thinking content with signature.
type ThinkingBlock struct {
	Thinking  string `json:"thinking"`
	Signature string `json:"signature"`
}

// Type returns the content block type for ThinkingBlock.
func (b ThinkingBlock) Type() string {
	return "thinking"
}

// ToolUseBlock represents a tool use request.
type ToolUseBlock struct {
	ToolUseID string         `json:"tool_use_id"`
	Name      string         `json:"name"`
	Input     map[string]any `json:"input"`
}

// Type returns the content block type for ToolUseBlock.
func (b ToolUseBlock) Type() string {
	return "tool_use"
}

// ToolResultBlock represents the result of a tool use with strongly-typed content.
type ToolResultBlock struct {
	ToolUseID string         `json:"tool_use_id"`
	Content   MessageContent `json:"content"` // Strongly typed, not interface{}
	IsError   *bool          `json:"is_error,omitempty"`
}

// Type returns the content block type for ToolResultBlock.
func (b ToolResultBlock) Type() string {
	return "tool_result"
}
