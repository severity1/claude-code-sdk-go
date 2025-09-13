package claudecode

import (
	"github.com/severity1/claude-code-sdk-go/pkg/interfaces"
)

// Message represents any message type in the conversation.
type Message = interfaces.Message

// ContentBlock represents a content block within a message.
type ContentBlock = interfaces.ContentBlock

// Content types for messages
type TextContent = interfaces.TextContent
type BlockListContent = interfaces.BlockListContent
type ThinkingContent = interfaces.ThinkingContent

// UserMessage represents a message from the user.
type UserMessage = interfaces.UserMessage

// AssistantMessage represents a message from the assistant.
type AssistantMessage = interfaces.AssistantMessage

// SystemMessage represents a system prompt message.
type SystemMessage = interfaces.SystemMessage

// ResultMessage represents a result or status message.
type ResultMessage = interfaces.ResultMessage

// TextBlock represents a text content block.
type TextBlock = interfaces.TextBlock

// ThinkingBlock represents a thinking content block.
type ThinkingBlock = interfaces.ThinkingBlock

// ToolUseBlock represents a tool usage content block.
type ToolUseBlock = interfaces.ToolUseBlock

// ToolResultBlock represents a tool result content block.
type ToolResultBlock = interfaces.ToolResultBlock

// StreamMessage represents a message in the streaming protocol.
type StreamMessage = interfaces.StreamMessage

// MessageIterator provides iteration over messages.
type MessageIterator = interfaces.MessageIterator

// Re-export message type constants
const (
	MessageTypeUser      = interfaces.MessageTypeUser
	MessageTypeAssistant = interfaces.MessageTypeAssistant
	MessageTypeSystem    = interfaces.MessageTypeSystem
	MessageTypeResult    = interfaces.MessageTypeResult
)

// Re-export content block type constants
const (
	ContentBlockTypeText       = interfaces.ContentBlockTypeText
	ContentBlockTypeThinking   = interfaces.ContentBlockTypeThinking
	ContentBlockTypeToolUse    = interfaces.ContentBlockTypeToolUse
	ContentBlockTypeToolResult = interfaces.ContentBlockTypeToolResult
)

// Transport abstracts the communication layer with Claude Code CLI.
type Transport = interfaces.Transport

// ProcessStatus represents the current status of the Claude Code CLI process.
type ProcessStatus = interfaces.ProcessStatus
