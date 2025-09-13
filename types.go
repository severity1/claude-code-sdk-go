package claudecode

import (
	"context"

	"github.com/severity1/claude-code-sdk-go/internal/shared"
	"github.com/severity1/claude-code-sdk-go/pkg/interfaces"
)

// Message represents any message type in the conversation.
// NOTE: Phase 3 partial migration - keeping compatibility with subprocess
type Message = shared.Message

// ContentBlock represents a content block within a message.
// NOTE: Phase 3 partial migration - keeping compatibility with subprocess
type ContentBlock = shared.ContentBlock

// Content types for messages
type TextContent = interfaces.TextContent
type BlockListContent = interfaces.BlockListContent
type ThinkingContent = interfaces.ThinkingContent

// UserMessage represents a message from the user.
// NOTE: Phase 3 partial migration - keeping compatibility with subprocess
type UserMessage = shared.UserMessage

// AssistantMessage represents a message from the assistant.
// NOTE: Phase 3 partial migration - keeping compatibility with subprocess
type AssistantMessage = shared.AssistantMessage

// SystemMessage represents a system prompt message.
// Note: SystemMessage and ResultMessage will be added to interfaces package in future iterations
type SystemMessage struct {
	MessageType string         `json:"type"`
	Subtype     string         `json:"subtype"`
	Data        map[string]any `json:"-"`
}

// Type returns the message type for SystemMessage.
func (m *SystemMessage) Type() string {
	return "system"
}

// ResultMessage represents a result or status message.
// Note: ResultMessage will be added to interfaces package in future iterations
type ResultMessage struct {
	MessageType   string          `json:"type"`
	Subtype       string          `json:"subtype"`
	DurationMs    int             `json:"duration_ms"`
	DurationAPIMs int             `json:"duration_api_ms"`
	IsError       bool            `json:"is_error"`
	NumTurns      int             `json:"num_turns"`
	SessionID     string          `json:"session_id"`
	TotalCostUSD  *float64        `json:"total_cost_usd,omitempty"`
	Usage         *map[string]any `json:"usage,omitempty"`
	Result        *map[string]any `json:"result,omitempty"`
}

// Type returns the message type for ResultMessage.
func (m *ResultMessage) Type() string {
	return "result"
}

// TextBlock represents a text content block.
// NOTE: Phase 3 partial migration - keeping compatibility with subprocess
type TextBlock = shared.TextBlock

// ThinkingBlock represents a thinking content block.
// NOTE: Phase 3 partial migration - keeping compatibility with subprocess
type ThinkingBlock = shared.ThinkingBlock

// ToolUseBlock represents a tool usage content block.
// NOTE: Phase 3 partial migration - keeping compatibility with subprocess
type ToolUseBlock = shared.ToolUseBlock

// ToolResultBlock represents a tool result content block.
// NOTE: Phase 3 partial migration - keeping compatibility with subprocess
type ToolResultBlock = shared.ToolResultBlock

// StreamMessage represents a message in the streaming protocol.
// NOTE: Phase 3 partial migration - keeping compatibility with subprocess
type StreamMessage = shared.StreamMessage

// MessageIterator provides iteration over messages.
// NOTE: Phase 3 partial migration - keeping compatibility with subprocess
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
// NOTE: For Phase 3, keeping compatible with subprocess implementation
type Transport interface {
	Connect(ctx context.Context) error
	SendMessage(ctx context.Context, message StreamMessage) error
	ReceiveMessages(ctx context.Context) (<-chan Message, <-chan error)
	Interrupt(ctx context.Context) error
	Close() error
}

// Phase 3 Migration: Create compatibility aliases to bridge old and new interfaces
// These will be removed in Phase 3 Day 6 when we complete the migration

// NewMessage creates a new interface message from old shared message (compatibility)
type NewMessage = interfaces.Message

// NewContentBlock creates a new interface content block from old shared block (compatibility)
type NewContentBlock = interfaces.ContentBlock

// NewStreamMessage creates a new interface stream message (compatibility)
type NewStreamMessage = interfaces.StreamMessage
