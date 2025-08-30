package claudecode

import (
	"context"

	"github.com/severity1/claude-code-sdk-go/internal/shared"
)

// Re-export shared message types and constants for public API compatibility
type Message = shared.Message
type ContentBlock = shared.ContentBlock
type UserMessage = shared.UserMessage
type AssistantMessage = shared.AssistantMessage
type SystemMessage = shared.SystemMessage
type ResultMessage = shared.ResultMessage
type TextBlock = shared.TextBlock
type ThinkingBlock = shared.ThinkingBlock
type ToolUseBlock = shared.ToolUseBlock
type ToolResultBlock = shared.ToolResultBlock
type StreamMessage = shared.StreamMessage
type MessageIterator = shared.MessageIterator

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
}