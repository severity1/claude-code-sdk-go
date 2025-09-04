package claudecode

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// ErrNoMoreMessages indicates the message iterator has no more messages.
var ErrNoMoreMessages = errors.New("no more messages")

// Query executes a one-shot query with automatic cleanup.
// This follows the Python SDK pattern but uses dependency injection for transport.
func Query(ctx context.Context, prompt string, opts ...Option) (MessageIterator, error) {
	options := NewOptions(opts...)

	// Create default transport
	if defaultTransportFactory == nil {
		return nil, fmt.Errorf("no default transport factory available - please install transport package")
	}
	transport, err := defaultTransportFactory(options, true) // true = close stdin for one-shot query
	if err != nil {
		return nil, fmt.Errorf("failed to create default transport: %w", err)
	}

	return queryWithTransportAndOptions(ctx, prompt, transport, options)
}

// QueryStream executes a query with a stream of messages.
func QueryStream(ctx context.Context, messages <-chan StreamMessage, opts ...Option) (MessageIterator, error) {
	options := NewOptions(opts...)

	// Create default transport
	if defaultTransportFactory == nil {
		return nil, fmt.Errorf("no default transport factory available - please install transport package")
	}
	transport, err := defaultTransportFactory(options, false) // false = streaming mode, keep stdin open
	if err != nil {
		return nil, fmt.Errorf("failed to create default transport: %w", err)
	}

	return queryStreamWithTransportAndOptions(ctx, messages, transport, options)
}

// QueryWithTransport executes a query with a custom transport.
// The transport parameter is required and must not be nil.
func QueryWithTransport(ctx context.Context, prompt string, transport Transport, opts ...Option) (MessageIterator, error) {
	if transport == nil {
		return nil, fmt.Errorf("transport is required")
	}

	options := NewOptions(opts...)
	return queryWithTransportAndOptions(ctx, prompt, transport, options)
}

// QueryStreamWithTransport executes a stream query with a custom transport.
// The transport parameter is required and must not be nil.
func QueryStreamWithTransport(ctx context.Context, messages <-chan StreamMessage, transport Transport, opts ...Option) (MessageIterator, error) {
	if transport == nil {
		return nil, fmt.Errorf("transport is required")
	}

	options := NewOptions(opts...)
	return queryStreamWithTransportAndOptions(ctx, messages, transport, options)
}

// Internal helper functions
func queryWithTransportAndOptions(ctx context.Context, prompt string, transport Transport, options *Options) (MessageIterator, error) {
	if transport == nil {
		return nil, fmt.Errorf("transport is required")
	}

	// Create iterator that manages the transport lifecycle
	return &queryIterator{
		transport: transport,
		prompt:    prompt,
		ctx:       ctx,
		options:   options,
	}, nil
}

func queryStreamWithTransportAndOptions(ctx context.Context, messages <-chan StreamMessage, transport Transport, options *Options) (MessageIterator, error) {
	if transport == nil {
		return nil, fmt.Errorf("transport is required")
	}

	// Create iterator that manages the transport lifecycle
	return &streamIterator{
		transport: transport,
		messages:  messages,
		ctx:       ctx,
		options:   options,
	}, nil
}

// queryIterator implements MessageIterator for simple queries
type queryIterator struct {
	transport Transport
	prompt    string
	ctx       context.Context
	options   *Options
	started   bool
	msgChan   <-chan Message
	errChan   <-chan error
	mu        sync.Mutex
	closed    bool
	closeOnce sync.Once
}

func (qi *queryIterator) Next(ctx context.Context) (Message, error) {
	qi.mu.Lock()
	if qi.closed {
		qi.mu.Unlock()
		return nil, ErrNoMoreMessages
	}

	// Initialize on first call
	if !qi.started {
		if err := qi.start(); err != nil {
			qi.mu.Unlock()
			return nil, err
		}
		qi.started = true
	}
	qi.mu.Unlock()

	// Read from message channels
	select {
	case msg, ok := <-qi.msgChan:
		if !ok {
			qi.mu.Lock()
			qi.closed = true
			qi.mu.Unlock()
			return nil, ErrNoMoreMessages
		}
		return msg, nil
	case err := <-qi.errChan:
		qi.mu.Lock()
		qi.closed = true
		qi.mu.Unlock()
		return nil, err
	case <-qi.ctx.Done():
		qi.mu.Lock()
		qi.closed = true
		qi.mu.Unlock()
		return nil, qi.ctx.Err()
	}
}

func (qi *queryIterator) Close() error {
	var err error
	qi.closeOnce.Do(func() {
		qi.mu.Lock()
		qi.closed = true
		qi.mu.Unlock()
		if qi.transport != nil {
			err = qi.transport.Close()
		}
	})
	return err
}

func (qi *queryIterator) start() error {
	// Connect to transport
	if err := qi.transport.Connect(qi.ctx); err != nil {
		return fmt.Errorf("failed to connect transport: %w", err)
	}

	// Get message channels
	msgChan, errChan := qi.transport.ReceiveMessages(qi.ctx)
	qi.msgChan = msgChan
	qi.errChan = errChan

	// Send the prompt
	userMsg := &UserMessage{Content: qi.prompt}
	streamMsg := StreamMessage{
		Type:    "request",
		Message: userMsg,
	}

	if err := qi.transport.SendMessage(qi.ctx, streamMsg); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// streamIterator implements MessageIterator for streaming queries
type streamIterator struct {
	transport Transport
	messages  <-chan StreamMessage
	ctx       context.Context
	options   *Options
	started   bool
	msgChan   <-chan Message
	errChan   <-chan error
	sendErr   error
	mu        sync.Mutex
	closed    bool
	closeOnce sync.Once
}

func (si *streamIterator) Next(ctx context.Context) (Message, error) {
	si.mu.Lock()
	if si.closed {
		si.mu.Unlock()
		return nil, ErrNoMoreMessages
	}

	// Check for send errors
	if si.sendErr != nil {
		err := si.sendErr
		si.closed = true
		si.mu.Unlock()
		return nil, fmt.Errorf("send error: %w", err)
	}

	// Initialize on first call
	if !si.started {
		if err := si.start(); err != nil {
			si.mu.Unlock()
			return nil, err
		}
		si.started = true
	}
	si.mu.Unlock()

	// Read from message channels (transport-driven completion)
	select {
	case msg, ok := <-si.msgChan:
		if !ok {
			si.mu.Lock()
			si.closed = true
			si.mu.Unlock()
			return nil, ErrNoMoreMessages
		}
		return msg, nil
	case err := <-si.errChan:
		si.mu.Lock()
		si.closed = true
		si.mu.Unlock()
		return nil, err
	case <-si.ctx.Done():
		si.mu.Lock()
		si.closed = true
		si.mu.Unlock()
		return nil, si.ctx.Err()
	}
}

func (si *streamIterator) Close() error {
	var err error
	si.closeOnce.Do(func() {
		si.mu.Lock()
		si.closed = true
		si.mu.Unlock()
		if si.transport != nil {
			err = si.transport.Close()
		}
	})
	return err
}

func (si *streamIterator) start() error {
	// Connect to transport
	if err := si.transport.Connect(si.ctx); err != nil {
		return fmt.Errorf("failed to connect transport: %w", err)
	}

	// Get message channels
	msgChan, errChan := si.transport.ReceiveMessages(si.ctx)
	si.msgChan = msgChan
	si.errChan = errChan

	// Start goroutine to send messages from the stream
	go func() {
		for {
			select {
			case msg, ok := <-si.messages:
				if !ok {
					return // Channel closed - sender completed normally
				}
				if err := si.transport.SendMessage(si.ctx, msg); err != nil {
					// Store send error in iterator state
					si.mu.Lock()
					si.sendErr = err
					si.mu.Unlock()
					return
				}
			case <-si.ctx.Done():
				return
			}
		}
	}()

	return nil
}
