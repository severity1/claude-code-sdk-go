package claudecode

import (
	"context"

	"github.com/severity1/claude-code-sdk-go/internal/control"
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

// AssistantMessageError represents error types in assistant messages.
type AssistantMessageError = shared.AssistantMessageError

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

	// Control protocol message types
	MessageTypeControlRequest  = shared.MessageTypeControlRequest
	MessageTypeControlResponse = shared.MessageTypeControlResponse

	// Partial message streaming type
	MessageTypeStreamEvent = shared.MessageTypeStreamEvent
)

// Re-export content block type constants
const (
	ContentBlockTypeText       = shared.ContentBlockTypeText
	ContentBlockTypeThinking   = shared.ContentBlockTypeThinking
	ContentBlockTypeToolUse    = shared.ContentBlockTypeToolUse
	ContentBlockTypeToolResult = shared.ContentBlockTypeToolResult
)

// Re-export stream event type constants for Event["type"] discrimination.
const (
	StreamEventTypeContentBlockStart = shared.StreamEventTypeContentBlockStart
	StreamEventTypeContentBlockDelta = shared.StreamEventTypeContentBlockDelta
	StreamEventTypeContentBlockStop  = shared.StreamEventTypeContentBlockStop
	StreamEventTypeMessageStart      = shared.StreamEventTypeMessageStart
	StreamEventTypeMessageDelta      = shared.StreamEventTypeMessageDelta
	StreamEventTypeMessageStop       = shared.StreamEventTypeMessageStop
)

// Re-export AssistantMessageError constants
const (
	AssistantMessageErrorAuthFailed     = shared.AssistantMessageErrorAuthFailed
	AssistantMessageErrorBilling        = shared.AssistantMessageErrorBilling
	AssistantMessageErrorRateLimit      = shared.AssistantMessageErrorRateLimit
	AssistantMessageErrorInvalidRequest = shared.AssistantMessageErrorInvalidRequest
	AssistantMessageErrorServer         = shared.AssistantMessageErrorServer
	AssistantMessageErrorUnknown        = shared.AssistantMessageErrorUnknown
)

// AgentModel represents the model to use for an agent.
type AgentModel = shared.AgentModel

// AgentDefinition defines a programmatic subagent.
type AgentDefinition = shared.AgentDefinition

// Re-export agent model constants
const (
	AgentModelSonnet  = shared.AgentModelSonnet
	AgentModelOpus    = shared.AgentModelOpus
	AgentModelHaiku   = shared.AgentModelHaiku
	AgentModelInherit = shared.AgentModelInherit
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

// RawControlMessage wraps raw control protocol messages for passthrough.
type RawControlMessage = shared.RawControlMessage

// StreamEvent represents a partial message update during streaming.
type StreamEvent = shared.StreamEvent

// Control protocol types for SDK-CLI bidirectional communication.

// SDKControlRequest represents a control request sent to the CLI.
type SDKControlRequest = control.SDKControlRequest

// SDKControlResponse represents a control response received from the CLI.
type SDKControlResponse = control.SDKControlResponse

// ControlResponse is the inner response structure.
type ControlResponse = control.Response

// InitializeRequest for control protocol handshake.
type InitializeRequest = control.InitializeRequest

// InitializeResponse from CLI with supported capabilities.
type InitializeResponse = control.InitializeResponse

// InterruptRequest to interrupt current operation via control protocol.
type InterruptRequest = control.InterruptRequest

// ControlProtocol manages bidirectional control communication with CLI.
type ControlProtocol = control.Protocol

// Re-export control protocol subtype constants
const (
	// Control request subtypes
	SubtypeInterrupt         = control.SubtypeInterrupt
	SubtypeCanUseTool        = control.SubtypeCanUseTool
	SubtypeInitialize        = control.SubtypeInitialize
	SubtypeSetPermissionMode = control.SubtypeSetPermissionMode
	SubtypeHookCallback      = control.SubtypeHookCallback
	SubtypeMcpMessage        = control.SubtypeMcpMessage

	// Control response subtypes
	ResponseSubtypeSuccess = control.ResponseSubtypeSuccess
	ResponseSubtypeError   = control.ResponseSubtypeError
)
