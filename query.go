package claudecode

import (
	"context"
)

// Query executes a one-shot query with automatic cleanup.
func Query(ctx context.Context, prompt string, opts ...Option) (MessageIterator, error) {
	// TODO: Implement query logic with transport
	return nil, nil
}

// QueryStream executes a query with a stream of messages.
func QueryStream(ctx context.Context, messages <-chan StreamMessage, opts ...Option) (MessageIterator, error) {
	// TODO: Implement stream query logic
	return nil, nil
}
