package claudecode

import (
	"context"
	"encoding/json"
)

// Message type constants
const (
	MessageTypeUser      = "user"
	MessageTypeAssistant = "assistant"
	MessageTypeSystem    = "system"
	MessageTypeResult    = "result"
)

// Content block type constants
const (
	ContentBlockTypeText       = "text"
	ContentBlockTypeThinking   = "thinking"
	ContentBlockTypeToolUse    = "tool_use"
	ContentBlockTypeToolResult = "tool_result"
)

// Message represents any message type in the Claude Code protocol.
type Message interface {
	Type() string
}

// ContentBlock represents any content block within a message.
type ContentBlock interface {
	BlockType() string
}

// UserMessage represents a message from the user.
type UserMessage struct {
	Type_   string      `json:"type"`
	Content interface{} `json:"content"` // string or []ContentBlock
}

func (m *UserMessage) Type() string {
	return MessageTypeUser
}

// MarshalJSON implements custom JSON marshaling for UserMessage
func (m *UserMessage) MarshalJSON() ([]byte, error) {
	type userMessage UserMessage
	temp := struct {
		Type string `json:"type"`
		*userMessage
	}{
		Type:        MessageTypeUser,
		userMessage: (*userMessage)(m),
	}
	return json.Marshal(temp)
}

// AssistantMessage represents a message from the assistant.
type AssistantMessage struct {
	Type_   string         `json:"type"`
	Content []ContentBlock `json:"content"`
	Model   string         `json:"model"`
}

func (m *AssistantMessage) Type() string {
	return MessageTypeAssistant
}

// MarshalJSON implements custom JSON marshaling for AssistantMessage
func (m *AssistantMessage) MarshalJSON() ([]byte, error) {
	type assistantMessage AssistantMessage
	temp := struct {
		Type string `json:"type"`
		*assistantMessage
	}{
		Type:             MessageTypeAssistant,
		assistantMessage: (*assistantMessage)(m),
	}
	return json.Marshal(temp)
}

// SystemMessage represents a system message.
type SystemMessage struct {
	Type_   string         `json:"type"`
	Subtype string         `json:"subtype"`
	Data    map[string]any `json:"-"` // Preserve all original data
}

func (m *SystemMessage) Type() string {
	return MessageTypeSystem
}

// MarshalJSON implements custom JSON marshaling for SystemMessage
func (m *SystemMessage) MarshalJSON() ([]byte, error) {
	data := make(map[string]any)
	for k, v := range m.Data {
		data[k] = v
	}
	data["type"] = MessageTypeSystem
	data["subtype"] = m.Subtype
	return json.Marshal(data)
}

// ResultMessage represents a result message with timing and session info.
type ResultMessage struct {
	Type_         string         `json:"type"`
	Subtype       string         `json:"subtype"`
	DurationMs    int            `json:"duration_ms"`
	DurationAPIMs int            `json:"duration_api_ms"`
	IsError       bool           `json:"is_error"`
	NumTurns      int            `json:"num_turns"`
	SessionID     string         `json:"session_id"`
	TotalCostUSD  *float64       `json:"total_cost_usd,omitempty"`
	Usage         map[string]any `json:"usage,omitempty"`
	Result        *string        `json:"result,omitempty"`
}

func (m *ResultMessage) Type() string {
	return MessageTypeResult
}

// MarshalJSON implements custom JSON marshaling for ResultMessage
func (m *ResultMessage) MarshalJSON() ([]byte, error) {
	type resultMessage ResultMessage
	temp := struct {
		Type string `json:"type"`
		*resultMessage
	}{
		Type:          MessageTypeResult,
		resultMessage: (*resultMessage)(m),
	}
	return json.Marshal(temp)
}

// TextBlock represents a text content block.
type TextBlock struct {
	Type_ string `json:"type"`
	Text  string `json:"text"`
}

func (b *TextBlock) BlockType() string {
	return ContentBlockTypeText
}

// MarshalJSON implements custom JSON marshaling for TextBlock
func (b *TextBlock) MarshalJSON() ([]byte, error) {
	type textBlock TextBlock
	temp := struct {
		Type string `json:"type"`
		*textBlock
	}{
		Type:      ContentBlockTypeText,
		textBlock: (*textBlock)(b),
	}
	return json.Marshal(temp)
}

// ThinkingBlock represents a thinking content block.
type ThinkingBlock struct {
	Type_     string `json:"type"`
	Thinking  string `json:"thinking"`
	Signature string `json:"signature"`
}

func (b *ThinkingBlock) BlockType() string {
	return ContentBlockTypeThinking
}

// MarshalJSON implements custom JSON marshaling for ThinkingBlock
func (b *ThinkingBlock) MarshalJSON() ([]byte, error) {
	type thinkingBlock ThinkingBlock
	temp := struct {
		Type string `json:"type"`
		*thinkingBlock
	}{
		Type:          ContentBlockTypeThinking,
		thinkingBlock: (*thinkingBlock)(b),
	}
	return json.Marshal(temp)
}

// ToolUseBlock represents a tool use content block.
type ToolUseBlock struct {
	Type_ string         `json:"type"`
	ID    string         `json:"id"`
	Name  string         `json:"name"`
	Input map[string]any `json:"input"`
}

func (b *ToolUseBlock) BlockType() string {
	return ContentBlockTypeToolUse
}

// MarshalJSON implements custom JSON marshaling for ToolUseBlock
func (b *ToolUseBlock) MarshalJSON() ([]byte, error) {
	type toolUseBlock ToolUseBlock
	temp := struct {
		Type string `json:"type"`
		*toolUseBlock
	}{
		Type:         ContentBlockTypeToolUse,
		toolUseBlock: (*toolUseBlock)(b),
	}
	return json.Marshal(temp)
}

// ToolResultBlock represents a tool result content block.
type ToolResultBlock struct {
	Type_     string `json:"type"`
	ToolUseID string `json:"tool_use_id"`
	Content   any    `json:"content,omitempty"` // *string, []map[string]any, or nil
	IsError   *bool  `json:"is_error,omitempty"`
}

func (b *ToolResultBlock) BlockType() string {
	return ContentBlockTypeToolResult
}

// MarshalJSON implements custom JSON marshaling for ToolResultBlock
func (b *ToolResultBlock) MarshalJSON() ([]byte, error) {
	type toolResultBlock ToolResultBlock
	temp := struct {
		Type string `json:"type"`
		*toolResultBlock
	}{
		Type:            ContentBlockTypeToolResult,
		toolResultBlock: (*toolResultBlock)(b),
	}
	return json.Marshal(temp)
}

// UnmarshalMessage unmarshals JSON into the appropriate Message type based on "type" field.
func UnmarshalMessage(data []byte) (Message, error) {
	var typeCheck struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal(data, &typeCheck); err != nil {
		return nil, err
	}

	switch typeCheck.Type {
	case MessageTypeUser:
		var msg UserMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, err
		}
		return &msg, nil
	case MessageTypeAssistant:
		var msg AssistantMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, err
		}
		return &msg, nil
	case MessageTypeSystem:
		// SystemMessage needs special handling to preserve all data
		var rawData map[string]any
		if err := json.Unmarshal(data, &rawData); err != nil {
			return nil, err
		}

		subtype, ok := rawData["subtype"].(string)
		if !ok {
			return nil, NewMessageParseError("missing or invalid subtype field in system message", rawData)
		}

		msg := &SystemMessage{
			Subtype: subtype,
			Data:    rawData,
		}
		return msg, nil
	case MessageTypeResult:
		var msg ResultMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, err
		}
		return &msg, nil
	default:
		return nil, NewMessageParseError("unknown message type: "+typeCheck.Type, map[string]any{"type": typeCheck.Type})
	}
}

// UnmarshalContentBlock unmarshals JSON into the appropriate ContentBlock type based on "type" field.
func UnmarshalContentBlock(data []byte) (ContentBlock, error) {
	var typeCheck struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal(data, &typeCheck); err != nil {
		return nil, err
	}

	switch typeCheck.Type {
	case ContentBlockTypeText:
		var block TextBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, err
		}
		return &block, nil
	case ContentBlockTypeThinking:
		var block ThinkingBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, err
		}
		return &block, nil
	case ContentBlockTypeToolUse:
		var block ToolUseBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, err
		}
		return &block, nil
	case ContentBlockTypeToolResult:
		var block ToolResultBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, err
		}
		return &block, nil
	default:
		return nil, NewMessageParseError("unknown content block type: "+typeCheck.Type, map[string]any{"type": typeCheck.Type})
	}
}

// Custom UnmarshalJSON for UserMessage to handle flexible Content field (string or []ContentBlock)
func (m *UserMessage) UnmarshalJSON(data []byte) error {
	// First try to unmarshal as string content
	var stringContent struct {
		Content string `json:"content"`
	}

	if err := json.Unmarshal(data, &stringContent); err == nil && stringContent.Content != "" {
		m.Content = stringContent.Content
		return nil
	}

	// If that fails, try as content block array
	var blockContent struct {
		Content []json.RawMessage `json:"content"`
	}

	if err := json.Unmarshal(data, &blockContent); err != nil {
		return err
	}

	// Unmarshal each content block
	contentBlocks := make([]ContentBlock, len(blockContent.Content))
	for i, rawBlock := range blockContent.Content {
		block, err := UnmarshalContentBlock(rawBlock)
		if err != nil {
			return err
		}
		contentBlocks[i] = block
	}

	m.Content = contentBlocks
	return nil
}

// Custom UnmarshalJSON for AssistantMessage to handle ContentBlock arrays
func (m *AssistantMessage) UnmarshalJSON(data []byte) error {
	var temp struct {
		Content []json.RawMessage `json:"content"`
		Model   string            `json:"model"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	m.Model = temp.Model

	// Unmarshal each content block
	m.Content = make([]ContentBlock, len(temp.Content))
	for i, rawBlock := range temp.Content {
		block, err := UnmarshalContentBlock(rawBlock)
		if err != nil {
			return err
		}
		m.Content[i] = block
	}

	return nil
}

// Transport abstracts the communication layer with Claude Code CLI.
type Transport interface {
	Connect(ctx context.Context) error
	SendMessage(ctx context.Context, message StreamMessage) error
	ReceiveMessages(ctx context.Context) (<-chan Message, <-chan error)
	Interrupt(ctx context.Context) error
	Close() error
}

// StreamMessage represents messages sent to the CLI for streaming communication.
type StreamMessage struct {
	Type            string                 `json:"type"`
	Message         interface{}            `json:"message,omitempty"`
	ParentToolUseID *string                `json:"parent_tool_use_id,omitempty"`
	SessionID       string                 `json:"session_id,omitempty"`
	RequestID       string                 `json:"request_id,omitempty"`
	Request         map[string]interface{} `json:"request,omitempty"`
	Response        map[string]interface{} `json:"response,omitempty"`
}

// MessageIterator provides an iterator pattern for streaming messages.
type MessageIterator interface {
	Next(ctx context.Context) (Message, error)
	Close() error
}
