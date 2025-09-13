package interfaces

import "context"

// ConnectionManager handles connection lifecycle operations.
// Follows Interface Segregation Principle for focused responsibility.
type ConnectionManager interface {
	Connect(ctx context.Context) error
	Close() error
	IsConnected() bool
}

// QueryExecutor handles query execution operations.
// Focused interface for query-related functionality.
type QueryExecutor interface {
	Query(ctx context.Context, prompt string) error
	QueryStream(ctx context.Context, messages <-chan StreamMessage) error
}

// MessageReceiver handles message receiving operations.
// Focused interface for message reception functionality.
type MessageReceiver interface {
	ReceiveMessages(ctx context.Context) <-chan Message
	ReceiveResponse(ctx context.Context) MessageIterator
}

// ProcessStatus represents the current status of the Claude Code CLI process.
type ProcessStatus struct {
	Running   bool   `json:"running"`
	PID       int    `json:"pid,omitempty"`
	StartTime string `json:"start_time,omitempty"`
}

// ProcessController handles process control operations.
// Focused interface for interruption and control.
type ProcessController interface {
	Interrupt(ctx context.Context) error
	Status() ProcessStatus
}

// Client represents the full client interface composed of focused interfaces.
// Uses Go's interface embedding for composition.
type Client interface {
	ConnectionManager
	QueryExecutor
	MessageReceiver
	ProcessController
}

// SimpleQuerier provides a minimal interface for users who only need query functionality.
type SimpleQuerier interface {
	QueryExecutor
}

// StreamClient provides interfaces for users building custom streaming functionality.
type StreamClient interface {
	ConnectionManager
	MessageReceiver
}
