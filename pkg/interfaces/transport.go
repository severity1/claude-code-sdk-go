package interfaces

import "context"

// Transport abstracts the communication layer with Claude Code CLI.
// Follows Go-native concurrency patterns with context-first design.
type Transport interface {
	Connect(ctx context.Context) error
	SendMessage(ctx context.Context, message StreamMessage) error
	ReceiveMessages(ctx context.Context) (<-chan Message, <-chan error)
	Interrupt(ctx context.Context) error
	Close() error
}

// MessageIterator provides an iterator pattern for streaming messages.
// Follows standard Go iterator patterns with context support.
type MessageIterator interface {
	Next(ctx context.Context) (Message, error)
	Close() error
}

// StreamMessage represents messages sent to the CLI for streaming communication.
// This is a placeholder - will be properly typed in later phases.
type StreamMessage interface {
	// Placeholder interface for now
}
