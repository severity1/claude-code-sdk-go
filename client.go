package claudecode

import (
	"context"
	"fmt"
	"sync"
)

// TransportFactory is a function type that creates a default transport.
type TransportFactory func(options *Options, closeStdin bool) (Transport, error)

// defaultTransportFactory is the default factory function for creating transports.
// This can be set by the init() function in a separate file to avoid import cycles.
var defaultTransportFactory TransportFactory

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
	mu              sync.RWMutex
	transport       Transport
	customTransport Transport // For testing with WithTransport
	options         *Options
	connected       bool
	msgChan         <-chan Message
	errChan         <-chan error
}

// NewClient creates a new Client with the given options.
func NewClient(opts ...Option) Client {
	options := NewOptions(opts...)
	client := &ClientImpl{
		options: options,
	}
	return client
}

// NewClientWithTransport creates a new Client with a custom transport (for testing).
func NewClientWithTransport(transport Transport, opts ...Option) Client {
	options := NewOptions(opts...)
	return &ClientImpl{
		customTransport: transport,
		options:         options,
	}
}

// Connect establishes a connection to the Claude Code CLI.
func (c *ClientImpl) Connect(ctx context.Context, prompt ...StreamMessage) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Use custom transport if provided, otherwise create default
	if c.customTransport != nil {
		c.transport = c.customTransport
	} else {
		// Create default transport using the factory
		if defaultTransportFactory == nil {
			return fmt.Errorf("no default transport factory available - please install github.com/severity1/claude-code-sdk-go/transport package or use NewClientWithTransport")
		}
		
		transport, err := defaultTransportFactory(c.options, false) // false = streaming mode
		if err != nil {
			return fmt.Errorf("failed to create default transport: %w", err)
		}
		c.transport = transport
	}
	
	// Connect the transport
	if err := c.transport.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect transport: %w", err)
	}
	
	// Get message channels
	c.msgChan, c.errChan = c.transport.ReceiveMessages(ctx)
	
	c.connected = true
	return nil
}

// Disconnect closes the connection to the Claude Code CLI.
func (c *ClientImpl) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.transport != nil && c.connected {
		if err := c.transport.Close(); err != nil {
			return fmt.Errorf("failed to close transport: %w", err)
		}
	}
	c.connected = false
	c.transport = nil
	c.msgChan = nil
	c.errChan = nil
	return nil
}

// Query sends a simple text query.
func (c *ClientImpl) Query(ctx context.Context, prompt string, sessionID ...string) error {
	// Check connection status with read lock
	c.mu.RLock()
	connected := c.connected
	transport := c.transport
	c.mu.RUnlock()
	
	if !connected || transport == nil {
		return fmt.Errorf("client not connected")
	}
	
	// Determine session ID - use first provided, otherwise default to "default"
	sid := "default"
	if len(sessionID) > 0 && sessionID[0] != "" {
		sid = sessionID[0]
	}
	
	// Create user message
	userMsg := &UserMessage{Content: prompt}
	streamMsg := StreamMessage{
		Type:      "request",
		Message:   userMsg,
		SessionID: sid,
	}
	
	// Send message via transport (without holding mutex to avoid blocking other operations)
	return transport.SendMessage(ctx, streamMsg)
}

// QueryStream sends a stream of messages.
func (c *ClientImpl) QueryStream(ctx context.Context, messages <-chan StreamMessage) error {
	// Check connection status with read lock
	c.mu.RLock()
	connected := c.connected
	transport := c.transport
	c.mu.RUnlock()
	
	if !connected || transport == nil {
		return fmt.Errorf("client not connected")
	}
	
	// Send messages from channel in a goroutine
	go func() {
		for {
			select {
			case msg, ok := <-messages:
				if !ok {
					return // Channel closed
				}
				if err := transport.SendMessage(ctx, msg); err != nil {
					// Log error but continue processing
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	
	return nil
}

// ReceiveMessages returns a channel of incoming messages.
func (c *ClientImpl) ReceiveMessages(ctx context.Context) <-chan Message {
	// Check connection status with read lock
	c.mu.RLock()
	connected := c.connected
	msgChan := c.msgChan
	c.mu.RUnlock()
	
	if !connected || msgChan == nil {
		// Return closed channel if not connected
		closedChan := make(chan Message)
		close(closedChan)
		return closedChan
	}
	
	// Return the transport's message channel directly
	return msgChan
}

// ReceiveResponse returns an iterator for the response messages.
func (c *ClientImpl) ReceiveResponse(ctx context.Context) MessageIterator {
	// Check connection status with read lock
	c.mu.RLock()
	connected := c.connected
	msgChan := c.msgChan
	errChan := c.errChan
	c.mu.RUnlock()
	
	if !connected || msgChan == nil {
		return nil
	}
	
	// Create a simple iterator over the message channel
	return &clientIterator{
		msgChan: msgChan,
		errChan: errChan,
	}
}

// Interrupt sends an interrupt signal to stop the current operation.
func (c *ClientImpl) Interrupt(ctx context.Context) error {
	// Check connection status with read lock
	c.mu.RLock()
	connected := c.connected
	transport := c.transport
	c.mu.RUnlock()
	
	if !connected || transport == nil {
		return fmt.Errorf("client not connected")
	}
	
	return transport.Interrupt(ctx)
}


// clientIterator implements MessageIterator for client message reception
type clientIterator struct {
	msgChan <-chan Message
	errChan <-chan error
	closed  bool
}

func (ci *clientIterator) Next(ctx context.Context) (Message, error) {
	if ci.closed {
		return nil, ErrNoMoreMessages
	}
	
	select {
	case msg, ok := <-ci.msgChan:
		if !ok {
			ci.closed = true
			return nil, ErrNoMoreMessages
		}
		return msg, nil
	case err := <-ci.errChan:
		ci.closed = true
		return nil, err
	case <-ctx.Done():
		ci.closed = true
		return nil, ctx.Err()
	}
}

func (ci *clientIterator) Close() error {
	ci.closed = true
	return nil
}
