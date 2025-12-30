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

	// Permission callback (Issue #8)
	canUseToolCallback CanUseToolCallback

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

// WithCanUseToolCallback sets the permission callback for tool usage requests.
// The callback is invoked when CLI requests permission to use a tool.
func WithCanUseToolCallback(callback CanUseToolCallback) ProtocolOption {
	return func(p *Protocol) {
		p.canUseToolCallback = callback
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
		return p.handleIncomingControlRequest(ctx, msg)
	default:
		// Regular SDK message - forward to stream
		return p.forwardToStream(ctx, msg)
	}
}

// handleIncomingControlRequest routes incoming control requests from CLI.
func (p *Protocol) handleIncomingControlRequest(ctx context.Context, msg map[string]any) error {
	request, ok := msg["request"].(map[string]any)
	if !ok {
		return fmt.Errorf("invalid control request: missing request field")
	}

	subtype, _ := request["subtype"].(string)
	requestID, _ := msg["request_id"].(string)

	switch subtype {
	case SubtypeCanUseTool:
		return p.handleCanUseToolRequest(ctx, requestID, request)
	case SubtypeHookCallback:
		// Issue #9 placeholder - hook callbacks
		return nil
	default:
		// Unknown subtype - ignore for forward compatibility
		return nil
	}
}

// handleCanUseToolRequest processes a permission check request from CLI.
// Follows StderrCallback pattern: synchronous with panic recovery.
func (p *Protocol) handleCanUseToolRequest(ctx context.Context, requestID string, request map[string]any) error {
	// Parse request fields
	toolName, _ := request["tool_name"].(string)
	if toolName == "" {
		return p.sendErrorResponse(ctx, requestID, "missing tool_name")
	}

	input, _ := request["input"].(map[string]any)
	if input == nil {
		input = make(map[string]any)
	}

	// Parse suggestions from context
	var permCtx ToolPermissionContext
	if suggestions, ok := request["permission_suggestions"].([]any); ok {
		permCtx.Suggestions = parsePermissionSuggestions(suggestions)
	}

	// Get callback (thread-safe read)
	p.mu.Lock()
	callback := p.canUseToolCallback
	p.mu.Unlock()

	// No callback = deny (secure default)
	if callback == nil {
		return p.sendPermissionResponse(ctx, requestID, NewPermissionResultDeny("no permission callback registered"))
	}

	// Invoke callback synchronously with panic recovery (matches StderrCallback pattern)
	var result PermissionResult
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("permission callback panicked: %v", r)
			}
		}()
		result, err = callback(ctx, toolName, input, permCtx)
	}()

	if err != nil {
		return p.sendErrorResponse(ctx, requestID, fmt.Sprintf("callback error: %v", err))
	}

	return p.sendPermissionResponse(ctx, requestID, result)
}

// sendPermissionResponse sends a permission result back to CLI.
func (p *Protocol) sendPermissionResponse(ctx context.Context, requestID string, result PermissionResult) error {
	// Build response based on result type
	var responseData map[string]any
	switch r := result.(type) {
	case PermissionResultAllow:
		responseData = map[string]any{"behavior": "allow"}
		if r.UpdatedInput != nil {
			responseData["updatedInput"] = r.UpdatedInput
		}
		if len(r.UpdatedPermissions) > 0 {
			responseData["updatedPermissions"] = r.UpdatedPermissions
		}
	case PermissionResultDeny:
		responseData = map[string]any{"behavior": "deny"}
		if r.Message != "" {
			responseData["message"] = r.Message
		}
		if r.Interrupt {
			responseData["interrupt"] = r.Interrupt
		}
	default:
		return fmt.Errorf("unknown permission result type: %T", result)
	}

	response := SDKControlResponse{
		Type: MessageTypeControlResponse,
		Response: Response{
			Subtype:   ResponseSubtypeSuccess,
			RequestID: requestID,
			Response:  responseData,
		},
	}

	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal permission response: %w", err)
	}

	return p.transport.Write(ctx, append(data, '\n'))
}

// sendErrorResponse sends an error response back to CLI.
func (p *Protocol) sendErrorResponse(ctx context.Context, requestID string, errMsg string) error {
	response := SDKControlResponse{
		Type: MessageTypeControlResponse,
		Response: Response{
			Subtype:   ResponseSubtypeError,
			RequestID: requestID,
			Error:     errMsg,
		},
	}

	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal error response: %w", err)
	}

	return p.transport.Write(ctx, append(data, '\n'))
}

// parsePermissionSuggestions converts raw JSON to PermissionUpdate slice.
func parsePermissionSuggestions(raw []any) []PermissionUpdate {
	var suggestions []PermissionUpdate
	for _, item := range raw {
		if m, ok := item.(map[string]any); ok {
			update := PermissionUpdate{}
			if t, ok := m["type"].(string); ok {
				update.Type = PermissionUpdateType(t)
			}
			if rules, ok := m["rules"].([]any); ok {
				for _, rule := range rules {
					if ruleMap, ok := rule.(map[string]any); ok {
						rv := PermissionRuleValue{}
						if tn, ok := ruleMap["toolName"].(string); ok {
							rv.ToolName = tn
						}
						if rc, ok := ruleMap["ruleContent"].(string); ok {
							rv.RuleContent = &rc
						}
						update.Rules = append(update.Rules, rv)
					}
				}
			}
			if b, ok := m["behavior"].(string); ok {
				update.Behavior = &b
			}
			if mode, ok := m["mode"].(string); ok {
				update.Mode = &mode
			}
			if dirs, ok := m["directories"].([]any); ok {
				for _, d := range dirs {
					if ds, ok := d.(string); ok {
						update.Directories = append(update.Directories, ds)
					}
				}
			}
			if dest, ok := m["destination"].(string); ok {
				update.Destination = &dest
			}
			suggestions = append(suggestions, update)
		}
	}
	return suggestions
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
