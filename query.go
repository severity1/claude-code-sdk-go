package claudecode

import (
	"context"
	"errors"
)

// ErrNoMoreMessages indicates the message iterator has no more messages.
var ErrNoMoreMessages = errors.New("no more messages")

// Query executes a one-shot query with automatic cleanup.
// This follows the Python SDK pattern but uses dependency injection for transport.
func Query(ctx context.Context, prompt string, opts ...Option) (MessageIterator, error) {
	return QueryWithTransport(ctx, prompt, nil, opts...)
}

// QueryStream executes a query with a stream of messages.
func QueryStream(ctx context.Context, messages <-chan StreamMessage, opts ...Option) (MessageIterator, error) {
	return QueryStreamWithTransport(ctx, messages, nil, opts...)
}

// QueryWithTransport executes a query with a custom transport.
// If transport is nil, a default transport will be created.
func QueryWithTransport(ctx context.Context, prompt string, transport Transport, opts ...Option) (MessageIterator, error) {
	options := NewOptions(opts...)

	// Create an internal client to handle the query
	internalClient := &InternalClient{}

	return internalClient.ProcessQuery(ctx, prompt, options, transport)
}

// QueryStreamWithTransport executes a stream query with a custom transport.
// If transport is nil, a default transport will be created.
func QueryStreamWithTransport(ctx context.Context, messages <-chan StreamMessage, transport Transport, opts ...Option) (MessageIterator, error) {
	options := NewOptions(opts...)

	// Create an internal client to handle the stream query
	internalClient := &InternalClient{}

	return internalClient.ProcessStreamQuery(ctx, messages, options, transport)
}
