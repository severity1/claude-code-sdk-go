package claudecode

import (
	"context"
)

// Client provides bidirectional streaming communication with Claude Code CLI.
type Client interface {
	Connect(ctx context.Context, prompt ...StreamMessage) error
	Disconnect() error
	Query(ctx context.Context, prompt string, sessionID ...string) error
	QueryStream(ctx context.Context, messages <-chan StreamMessage) error
	ReceiveMessages(ctx context.Context) <-chan Message
	ReceiveResponse(ctx context.Context) MessageIterator
	Interrupt(ctx context.Context) error
}

// ClientImpl implements the Client interface.
type ClientImpl struct {
	transport Transport
	options   *Options
	connected bool
}

// NewClient creates a new Client with the given options.
func NewClient(opts ...Option) Client {
	options := NewOptions(opts...)
	return &ClientImpl{
		options: options,
	}
}

// Connect establishes a connection to the Claude Code CLI.
func (c *ClientImpl) Connect(ctx context.Context, prompt ...StreamMessage) error {
	// TODO: Implement connection logic with transport
	c.connected = true
	return nil
}

// Disconnect closes the connection to the Claude Code CLI.
func (c *ClientImpl) Disconnect() error {
	// TODO: Implement disconnection logic
	c.connected = false
	return nil
}

// Query sends a simple text query.
func (c *ClientImpl) Query(ctx context.Context, prompt string, sessionID ...string) error {
	// TODO: Implement query logic
	return nil
}

// QueryStream sends a stream of messages.
func (c *ClientImpl) QueryStream(ctx context.Context, messages <-chan StreamMessage) error {
	// TODO: Implement stream query logic
	return nil
}

// ReceiveMessages returns a channel of incoming messages.
func (c *ClientImpl) ReceiveMessages(ctx context.Context) <-chan Message {
	// TODO: Implement message receiving
	msgChan := make(chan Message)
	close(msgChan) // Temporary: close empty channel
	return msgChan
}

// ReceiveResponse returns an iterator for the response messages.
func (c *ClientImpl) ReceiveResponse(ctx context.Context) MessageIterator {
	// TODO: Implement response receiving
	return nil
}

// Interrupt sends an interrupt signal to stop the current operation.
func (c *ClientImpl) Interrupt(ctx context.Context) error {
	// TODO: Implement interrupt logic
	return nil
}
