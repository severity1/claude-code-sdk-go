package control

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// DefaultInitTimeout is the default timeout for the Initialize handshake.
const DefaultInitTimeout = 60 * time.Second

// Transport abstracts the I/O operations for the control protocol.
// This allows testing with mock transports.
type Transport interface {
	// Write sends data to the CLI stdin.
	Write(ctx context.Context, data []byte) error
	// Read returns a channel that receives data from CLI stdout.
	Read(ctx context.Context) <-chan []byte
	// Close closes the transport.
	Close() error
}

// Protocol manages the bidirectional control protocol with Claude CLI.
// It handles request/response correlation, message routing, and initialization.
type Protocol struct {
	mu        sync.Mutex
	transport Transport

	// Request correlation
	pendingRequests map[string]chan *Response
	requestCounter  int64

	// Message routing
	messageStream chan map[string]any

	// State
	initialized  bool
	initResponse *InitializeResponse
	closed       bool
	started      bool

	// Configuration
	initTimeout time.Duration

	// Background goroutine management
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// ProtocolOption configures Protocol behavior.
type ProtocolOption func(*Protocol)

// WithInitTimeout sets the initialization timeout.
func WithInitTimeout(timeout time.Duration) ProtocolOption {
	return func(p *Protocol) {
		p.initTimeout = timeout
	}
}

// NewProtocol creates a new control protocol handler.
func NewProtocol(transport Transport, opts ...ProtocolOption) *Protocol {
	p := &Protocol{
		transport:       transport,
		pendingRequests: make(map[string]chan *Response),
		messageStream:   make(chan map[string]any, 100),
		initTimeout:     DefaultInitTimeout,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Start begins the message reading goroutine.
// This must be called before sending any control requests.
func (p *Protocol) Start(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.started {
		return nil
	}

	p.ctx, p.cancel = context.WithCancel(ctx)
	p.started = true

	// Start background message reader
	p.wg.Add(1)
	go p.readLoop()

	return nil
}

// readLoop continuously reads from transport and routes messages.
func (p *Protocol) readLoop() {
	defer p.wg.Done()

	readChan := p.transport.Read(p.ctx)

	for {
		select {
		case <-p.ctx.Done():
			return
		case data, ok := <-readChan:
			if !ok {
				return
			}

			// Parse the incoming message
			var msg map[string]any
			if err := json.Unmarshal(data, &msg); err != nil {
				// Log parse error but continue
				continue
			}

			// Route the message
			if err := p.HandleIncomingMessage(p.ctx, msg); err != nil {
				// Log routing error but continue
				continue
			}
		}
	}
}

// generateRequestID creates a unique request ID matching Python SDK format.
// Format: req_{counter}_{random_hex}
func (p *Protocol) generateRequestID() string {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.requestCounter++

	// Generate 4 random bytes as hex
	randomBytes := make([]byte, 4)
	_, _ = rand.Read(randomBytes)

	return fmt.Sprintf("req_%d_%x", p.requestCounter, randomBytes)
}

// SendControlRequest sends a control request and waits for the response.
// It uses the request ID for correlation with the matching response.
func (p *Protocol) SendControlRequest(ctx context.Context, request any, timeout time.Duration) (any, error) {
	requestID := p.generateRequestID()

	// Create response channel
	responseChan := make(chan *Response, 1)

	p.mu.Lock()
	p.pendingRequests[requestID] = responseChan
	p.mu.Unlock()

	// Cleanup on exit
	defer func() {
		p.mu.Lock()
		delete(p.pendingRequests, requestID)
		p.mu.Unlock()
	}()

	// Build control request envelope
	controlReq := SDKControlRequest{
		Type:      MessageTypeControlRequest,
		RequestID: requestID,
		Request:   request,
	}

	// Serialize and send
	data, err := json.Marshal(controlReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal control request: %w", err)
	}

	// Add newline for JSON lines protocol
	data = append(data, '\n')

	if err := p.transport.Write(ctx, data); err != nil {
		return nil, fmt.Errorf("failed to send control request: %w", err)
	}

	// Wait for response with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	select {
	case response := <-responseChan:
		if response.Subtype == ResponseSubtypeError {
			return nil, fmt.Errorf("control request error: %s", response.Error)
		}
		return response.Response, nil

	case <-timeoutCtx.Done():
		return nil, fmt.Errorf("control request timeout: %w", timeoutCtx.Err())
	}
}

// HandleIncomingMessage routes incoming messages based on their type.
// Control messages are handled internally, regular messages are forwarded to the stream.
func (p *Protocol) HandleIncomingMessage(ctx context.Context, msg map[string]any) error {
	msgType, ok := msg["type"].(string)
	if !ok {
		// No type field - forward to stream for compatibility
		return p.forwardToStream(ctx, msg)
	}

	switch msgType {
	case MessageTypeControlResponse:
		return p.handleControlResponse(ctx, msg)
	case MessageTypeControlRequest:
		// Incoming control request from CLI (e.g., hook callback, permission check)
		// For now, just log - handlers will be added in issues #8, #9
		return nil
	default:
		// Regular SDK message - forward to stream
		return p.forwardToStream(ctx, msg)
	}
}

// handleControlResponse routes a control response to the waiting request.
func (p *Protocol) handleControlResponse(_ context.Context, msg map[string]any) error {
	responseData, ok := msg["response"].(map[string]any)
	if !ok {
		return fmt.Errorf("invalid control response: missing response field")
	}

	requestID, ok := responseData["request_id"].(string)
	if !ok {
		return fmt.Errorf("invalid control response: missing request_id")
	}

	p.mu.Lock()
	responseChan, exists := p.pendingRequests[requestID]
	p.mu.Unlock()

	if !exists {
		// Response for unknown request - ignore (could be stale or from another session)
		return nil
	}

	response := &Response{
		RequestID: requestID,
	}

	if subtype, ok := responseData["subtype"].(string); ok {
		response.Subtype = subtype
	}

	if response.Subtype == ResponseSubtypeError {
		if errMsg, ok := responseData["error"].(string); ok {
			response.Error = errMsg
		}
	} else {
		response.Response = responseData["response"]
	}

	// Send response to waiting goroutine (non-blocking)
	select {
	case responseChan <- response:
	default:
		// Channel full or closed - ignore
	}

	return nil
}

// forwardToStream sends a message to the regular message stream.
func (p *Protocol) forwardToStream(ctx context.Context, msg map[string]any) error {
	select {
	case p.messageStream <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Initialize performs the control protocol handshake with the CLI.
// This must be called in streaming mode before other control operations.
// The result is cached - subsequent calls return the cached response.
func (p *Protocol) Initialize(ctx context.Context) (*InitializeResponse, error) {
	p.mu.Lock()
	if p.initialized {
		resp := p.initResponse
		p.mu.Unlock()
		return resp, nil
	}
	p.mu.Unlock()

	// Send initialize request
	result, err := p.SendControlRequest(ctx, InitializeRequest{
		Subtype: SubtypeInitialize,
	}, p.initTimeout)

	if err != nil {
		return nil, fmt.Errorf("initialize failed: %w", err)
	}

	// Parse response
	var initResp InitializeResponse
	if resultMap, ok := result.(map[string]any); ok {
		if cmds, ok := resultMap["supported_commands"].([]any); ok {
			for _, cmd := range cmds {
				if cmdStr, ok := cmd.(string); ok {
					initResp.SupportedCommands = append(initResp.SupportedCommands, cmdStr)
				}
			}
		}
	}

	p.mu.Lock()
	p.initialized = true
	p.initResponse = &initResp
	p.mu.Unlock()

	return &initResp, nil
}

// Interrupt sends an interrupt control request to the CLI.
func (p *Protocol) Interrupt(ctx context.Context) error {
	_, err := p.SendControlRequest(ctx, InterruptRequest{
		Subtype: SubtypeInterrupt,
	}, 5*time.Second)

	return err
}

// SetModel changes the AI model during a streaming session.
// Pass nil to reset to the default model.
// Returns error if the control request fails or times out.
func (p *Protocol) SetModel(ctx context.Context, model *string) error {
	_, err := p.SendControlRequest(ctx, SetModelRequest{
		Subtype: SubtypeSetModel,
		Model:   model,
	}, 5*time.Second)

	return err
}

// SetPermissionMode changes the permission mode during a streaming session.
// Valid modes: "default", "accept_edits", "plan", "bypass_permissions"
// Returns error if the control request fails or times out.
func (p *Protocol) SetPermissionMode(ctx context.Context, mode string) error {
	_, err := p.SendControlRequest(ctx, SetPermissionModeRequest{
		Subtype: SubtypeSetPermissionMode,
		Mode:    mode,
	}, 5*time.Second)

	return err
}

// ReceiveMessages returns a channel for receiving regular (non-control) messages.
func (p *Protocol) ReceiveMessages() <-chan map[string]any {
	return p.messageStream
}

// IsClosed returns whether the protocol has been closed.
func (p *Protocol) IsClosed() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.closed
}

// Close shuts down the protocol handler.
func (p *Protocol) Close() error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true
	p.mu.Unlock()

	// Cancel background goroutines
	if p.cancel != nil {
		p.cancel()
	}

	// Wait for goroutines to finish
	p.wg.Wait()

	// Close message stream
	close(p.messageStream)

	return nil
}

// setPendingRequest adds a pending request for testing purposes.
func (p *Protocol) setPendingRequest(requestID string, responseChan chan *Response) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pendingRequests[requestID] = responseChan
}
