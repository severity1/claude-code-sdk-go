package claudecode

import (
	"context"
	"fmt"
	"sync"
)

// InternalClient handles the internal logic for query and client operations
// This follows the Python SDK's InternalClient pattern
type InternalClient struct{}

// ProcessQuery handles one-shot query execution
func (ic *InternalClient) ProcessQuery(ctx context.Context, prompt string, options *Options, transport Transport) (MessageIterator, error) {
	// Use provided transport or create a mock one for testing
	if transport == nil {
		// For now, return an error. Users must provide a transport.
		// This avoids the import cycle while maintaining the interface.
		return nil, fmt.Errorf("transport is required - use QueryWithTransport or provide a transport")
	}

	// Pass options to transport if it supports configuration
	if configurableTransport, ok := transport.(interface {
		SetOptions(options *Options)
	}); ok {
		configurableTransport.SetOptions(options)
	}

	// Create iterator that manages the transport lifecycle
	return &queryIterator{
		transport: transport,
		prompt:    prompt,
		ctx:       ctx,
		options:   options,
	}, nil
}

// ProcessStreamQuery handles streaming query execution
func (ic *InternalClient) ProcessStreamQuery(ctx context.Context, messages <-chan StreamMessage, options *Options, transport Transport) (MessageIterator, error) {
	// Use provided transport or create a mock one for testing
	if transport == nil {
		// For now, return an error. Users must provide a transport.
		// This avoids the import cycle while maintaining the interface.
		return nil, fmt.Errorf("transport is required - use QueryStreamWithTransport or provide a transport")
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
	mu        sync.RWMutex
	closed    bool
	closeOnce sync.Once
}

func (qi *queryIterator) Next(ctx context.Context) (Message, error) {
	qi.mu.RLock()
	if qi.closed {
		qi.mu.RUnlock()
		return nil, ErrNoMoreMessages
	}
	qi.mu.RUnlock()

	if !qi.started {
		// Start the transport and send the query
		if err := qi.start(); err != nil {
			return nil, err
		}
		qi.started = true
	}

	// Read from message channel
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
	case <-ctx.Done():
		qi.mu.Lock()
		qi.closed = true
		qi.mu.Unlock()
		return nil, ctx.Err()
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

	// Send the prompt as a StreamMessage
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
	mu        sync.RWMutex
	closed    bool
	closeOnce sync.Once
}

func (si *streamIterator) Next(ctx context.Context) (Message, error) {
	si.mu.RLock()
	if si.closed {
		si.mu.RUnlock()
		return nil, ErrNoMoreMessages
	}
	si.mu.RUnlock()

	if !si.started {
		// Start the transport
		if err := si.start(); err != nil {
			return nil, err
		}
		si.started = true
	}

	// Read from message channel
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
	case <-ctx.Done():
		si.mu.Lock()
		si.closed = true
		si.mu.Unlock()
		return nil, ctx.Err()
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
		defer func() {
			// Close the transport when we're done sending messages
			si.Close()
		}()

		for {
			select {
			case msg, ok := <-si.messages:
				if !ok {
					return // Channel closed
				}
				if err := si.transport.SendMessage(si.ctx, msg); err != nil {
					// Error sending - let the error channel handle it
					return
				}
			case <-si.ctx.Done():
				return
			}
		}
	}()

	return nil
}
