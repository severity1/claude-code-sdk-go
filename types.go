package claudecode

import (
	"context"

	"github.com/severity1/claude-code-sdk-go/internal/shared"
)

// Message represents any message type in the conversation.
type Message = shared.Message

// ContentBlock represents a content block within a message.
type ContentBlock = shared.ContentBlock

// UserMessage represents a message from the user.
type UserMessage = shared.UserMessage

// AssistantMessage represents a message from the assistant.
type AssistantMessage = shared.AssistantMessage

// SystemMessage represents a system prompt message.
type SystemMessage = shared.SystemMessage

// ResultMessage represents a result or status message.
type ResultMessage = shared.ResultMessage

// TextBlock represents a text content block.
type TextBlock = shared.TextBlock

// ThinkingBlock represents a thinking content block.
type ThinkingBlock = shared.ThinkingBlock

// ToolUseBlock represents a tool usage content block.
type ToolUseBlock = shared.ToolUseBlock

// ToolResultBlock represents a tool result content block.
type ToolResultBlock = shared.ToolResultBlock

// StreamMessage represents a message in the streaming protocol.
type StreamMessage = shared.StreamMessage

// MessageIterator provides iteration over messages.
type MessageIterator = shared.MessageIterator

// StreamValidator tracks tool requests and results to detect incomplete streams.
type StreamValidator = shared.StreamValidator

// StreamIssue represents a validation issue found in the stream.
type StreamIssue = shared.StreamIssue

// StreamStats provides statistics about the message stream.
type StreamStats = shared.StreamStats

// Re-export message type constants
const (
	MessageTypeUser      = shared.MessageTypeUser
	MessageTypeAssistant = shared.MessageTypeAssistant
	MessageTypeSystem    = shared.MessageTypeSystem
	MessageTypeResult    = shared.MessageTypeResult
)

// Re-export content block type constants
const (
	ContentBlockTypeText       = shared.ContentBlockTypeText
	ContentBlockTypeThinking   = shared.ContentBlockTypeThinking
	ContentBlockTypeToolUse    = shared.ContentBlockTypeToolUse
	ContentBlockTypeToolResult = shared.ContentBlockTypeToolResult
)

// Transport abstracts the communication layer with Claude Code CLI.
// This interface stays in main package because it's used by client code.
type Transport interface {
	Connect(ctx context.Context) error
	SendMessage(ctx context.Context, message StreamMessage) error
	ReceiveMessages(ctx context.Context) (<-chan Message, <-chan error)
	Interrupt(ctx context.Context) error
	Close() error
	GetValidator() *StreamValidator
}
