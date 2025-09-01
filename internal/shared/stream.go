package shared

import "context"

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
