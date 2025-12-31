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

	// Hook callbacks (Issue #9)
	hooks            map[HookEvent][]HookMatcher
	hookCallbacks    map[string]HookCallback
	hookCallbacksMu  sync.RWMutex
	nextHookCallback int64

	// SDK MCP servers for in-process tool handling (Issue #7)
	sdkMcpServers map[string]McpServer

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

// WithHooks sets the hook configuration for lifecycle events.
// Hooks are registered during initialization and invoked by the CLI.
func WithHooks(hooks map[HookEvent][]HookMatcher) ProtocolOption {
	return func(p *Protocol) {
		p.hooks = hooks
	}
}

// WithHookCallbacks sets pre-registered hook callbacks by ID.
// This is primarily used for testing.
func WithHookCallbacks(callbacks map[string]HookCallback) ProtocolOption {
	return func(p *Protocol) {
		p.hookCallbacks = callbacks
	}
}

// WithSdkMcpServers configures SDK MCP servers for in-process tool handling.
// The servers map is keyed by server name.
func WithSdkMcpServers(servers map[string]McpServer) ProtocolOption {
	return func(p *Protocol) {
		p.sdkMcpServers = servers
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
		return p.handleHookCallbackRequest(ctx, requestID, request)
	case SubtypeMcpMessage:
		return p.handleMcpMessageRequest(ctx, requestID, request)
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

// handleHookCallbackRequest processes a hook callback request from CLI.
// Follows the same pattern as handleCanUseToolRequest with panic recovery.
func (p *Protocol) handleHookCallbackRequest(ctx context.Context, requestID string, request map[string]any) error {
	// Parse callback ID
	callbackID, _ := request["callback_id"].(string)
	if callbackID == "" {
		return p.sendErrorResponse(ctx, requestID, "missing callback_id")
	}

	// Parse hook event name from input
	inputData, _ := request["input"].(map[string]any)
	if inputData == nil {
		inputData = make(map[string]any)
	}

	eventName, _ := inputData["hook_event_name"].(string)
	event := HookEvent(eventName)

	// Parse input based on event type
	input := p.parseHookInput(event, inputData)

	// Parse tool_use_id if present
	var toolUseID *string
	if id, ok := request["tool_use_id"].(string); ok {
		toolUseID = &id
	}

	// Get callback (thread-safe read)
	p.hookCallbacksMu.RLock()
	callback, exists := p.hookCallbacks[callbackID]
	p.hookCallbacksMu.RUnlock()

	if !exists {
		return p.sendErrorResponse(ctx, requestID, fmt.Sprintf("callback not found: %s", callbackID))
	}

	// Create hook context
	hookCtx := HookContext{Signal: ctx}

	// Invoke callback with panic recovery (matches permission callback pattern)
	var result HookJSONOutput
	var callbackErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				callbackErr = fmt.Errorf("hook callback panicked: %v", r)
			}
		}()
		result, callbackErr = callback(ctx, input, toolUseID, hookCtx)
	}()

	if callbackErr != nil {
		return p.sendErrorResponse(ctx, requestID, fmt.Sprintf("callback error: %v", callbackErr))
	}

	return p.sendHookResponse(ctx, requestID, result)
}

// parseHookInput creates the appropriate typed input based on event type.
// Returns the strongly-typed input struct for the callback.
func (p *Protocol) parseHookInput(event HookEvent, inputData map[string]any) any {
	// Parse base fields
	base := BaseHookInput{
		SessionID:      getString(inputData, "session_id"),
		TranscriptPath: getString(inputData, "transcript_path"),
		Cwd:            getString(inputData, "cwd"),
		PermissionMode: getString(inputData, "permission_mode"),
	}

	switch event {
	case HookEventPreToolUse:
		return &PreToolUseHookInput{
			BaseHookInput: base,
			HookEventName: "PreToolUse",
			ToolName:      getString(inputData, "tool_name"),
			ToolInput:     getMap(inputData, "tool_input"),
		}
	case HookEventPostToolUse:
		return &PostToolUseHookInput{
			BaseHookInput: base,
			HookEventName: "PostToolUse",
			ToolName:      getString(inputData, "tool_name"),
			ToolInput:     getMap(inputData, "tool_input"),
			ToolResponse:  inputData["tool_response"],
		}
	case HookEventUserPromptSubmit:
		return &UserPromptSubmitHookInput{
			BaseHookInput: base,
			HookEventName: "UserPromptSubmit",
			Prompt:        getString(inputData, "prompt"),
		}
	case HookEventStop:
		return &StopHookInput{
			BaseHookInput:  base,
			HookEventName:  "Stop",
			StopHookActive: getBool(inputData, "stop_hook_active"),
		}
	case HookEventSubagentStop:
		return &SubagentStopHookInput{
			BaseHookInput:  base,
			HookEventName:  "SubagentStop",
			StopHookActive: getBool(inputData, "stop_hook_active"),
		}
	case HookEventPreCompact:
		return &PreCompactHookInput{
			BaseHookInput:      base,
			HookEventName:      "PreCompact",
			Trigger:            getString(inputData, "trigger"),
			CustomInstructions: getStringPtr(inputData, "custom_instructions"),
		}
	default:
		// Forward compatibility - return raw input for unknown events
		return inputData
	}
}

// sendHookResponse sends a hook callback response back to CLI.
func (p *Protocol) sendHookResponse(ctx context.Context, requestID string, result HookJSONOutput) error {
	// Build response data from HookJSONOutput
	responseData := make(map[string]any)

	if result.Continue != nil {
		responseData["continue"] = *result.Continue
	}
	if result.SuppressOutput != nil {
		responseData["suppressOutput"] = *result.SuppressOutput
	}
	if result.StopReason != nil {
		responseData["stopReason"] = *result.StopReason
	}
	if result.Decision != nil {
		responseData["decision"] = *result.Decision
	}
	if result.SystemMessage != nil {
		responseData["systemMessage"] = *result.SystemMessage
	}
	if result.Reason != nil {
		responseData["reason"] = *result.Reason
	}
	if result.HookSpecificOutput != nil {
		responseData["hookSpecificOutput"] = result.HookSpecificOutput
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
		return fmt.Errorf("failed to marshal hook response: %w", err)
	}

	return p.transport.Write(ctx, append(data, '\n'))
}

// generateHookRegistrations creates hook registrations for initialization.
// This builds the hooks config to send to CLI during initialize.
func (p *Protocol) generateHookRegistrations() []HookRegistration {
	var registrations []HookRegistration

	if p.hooks == nil {
		return registrations
	}

	// Initialize callback map if needed
	p.hookCallbacksMu.Lock()
	if p.hookCallbacks == nil {
		p.hookCallbacks = make(map[string]HookCallback)
	}

	for _, matchers := range p.hooks {
		for _, matcher := range matchers {
			for _, callback := range matcher.Hooks {
				// Generate callback ID matching Python SDK format
				callbackID := fmt.Sprintf("hook_%d", p.nextHookCallback)
				p.nextHookCallback++

				// Store callback for later lookup
				p.hookCallbacks[callbackID] = callback

				registrations = append(registrations, HookRegistration{
					CallbackID: callbackID,
					Matcher:    matcher.Matcher,
					Timeout:    matcher.Timeout,
				})
			}
		}
	}
	p.hookCallbacksMu.Unlock()

	return registrations
}

// buildHooksConfig creates the hooks config for the initialize request.
// Format: {"PreToolUse": [{"matcher": "Bash", "hookCallbackIds": ["hook_0"]}], ...}
// This matches the Python SDK's format exactly for CLI compatibility.
func (p *Protocol) buildHooksConfig() map[string][]HookMatcherConfig {
	if p.hooks == nil {
		return nil
	}

	config := make(map[string][]HookMatcherConfig)

	// Initialize callback map if needed
	p.hookCallbacksMu.Lock()
	if p.hookCallbacks == nil {
		p.hookCallbacks = make(map[string]HookCallback)
	}

	for event, matchers := range p.hooks {
		eventName := string(event)
		var matcherConfigs []HookMatcherConfig

		for _, matcher := range matchers {
			// Generate callback IDs for each callback in this matcher
			var callbackIDs []string
			for _, callback := range matcher.Hooks {
				callbackID := fmt.Sprintf("hook_%d", p.nextHookCallback)
				p.nextHookCallback++

				// Store callback for later lookup
				p.hookCallbacks[callbackID] = callback
				callbackIDs = append(callbackIDs, callbackID)
			}

			matcherConfigs = append(matcherConfigs, HookMatcherConfig{
				Matcher:         matcher.Matcher,
				HookCallbackIDs: callbackIDs,
				Timeout:         matcher.Timeout,
			})
		}

		if len(matcherConfigs) > 0 {
			config[eventName] = matcherConfigs
		}
	}
	p.hookCallbacksMu.Unlock()

	return config
}

// Helper functions for parsing hook input fields

func getString(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getStringPtr(m map[string]any, key string) *string {
	if v, ok := m[key].(string); ok {
		return &v
	}
	return nil
}

func getBool(m map[string]any, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}

func getMap(m map[string]any, key string) map[string]any {
	if v, ok := m[key].(map[string]any); ok {
		return v
	}
	return make(map[string]any)
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
// Invalid or unrecognized items are silently skipped for forward compatibility
// with future CLI versions that may introduce new fields or formats.
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

	// Build initialize request with hooks configuration
	initReq := InitializeRequest{
		Subtype: SubtypeInitialize,
	}

	// Generate hook registrations and build hooks config
	if p.hooks != nil {
		initReq.Hooks = p.buildHooksConfig()
	}

	// Send initialize request
	result, err := p.SendControlRequest(ctx, initReq, p.initTimeout)

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

// RewindFiles reverts tracked files to their state at a specific user message.
// The userMessageID should be the UUID from a UserMessage received during the session.
// Requires EnableFileCheckpointing to be set when creating the client.
// Returns error if the control request fails or times out.
//
// This method matches Python SDK's rewind_files behavior exactly:
// - Uses "rewind_files" subtype
// - Sends user_message_id in the request
// - Uses standard 5-second timeout
func (p *Protocol) RewindFiles(ctx context.Context, userMessageID string) error {
	_, err := p.SendControlRequest(ctx, RewindFilesRequest{
		Subtype:       SubtypeRewindFiles,
		UserMessageID: userMessageID,
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

// =============================================================================
// MCP Message Handling (Issue #7)
// =============================================================================

// handleMcpMessageRequest routes MCP JSONRPC messages to SDK servers.
// Follows handleCanUseToolRequest pattern with panic recovery.
func (p *Protocol) handleMcpMessageRequest(ctx context.Context, requestID string, request map[string]any) error {
	serverName := getString(request, "server_name")
	if serverName == "" {
		return p.sendErrorResponse(ctx, requestID, "missing server_name")
	}

	message, _ := request["message"].(map[string]any)
	if message == nil {
		return p.sendErrorResponse(ctx, requestID, "missing message")
	}

	// Thread-safe server lookup
	p.mu.Lock()
	server, exists := p.sdkMcpServers[serverName]
	p.mu.Unlock()

	if !exists {
		return p.sendMcpErrorResponse(ctx, requestID, message, -32601,
			fmt.Sprintf("server '%s' not found", serverName))
	}

	// Route JSONRPC method with panic recovery
	var mcpResponse map[string]any
	var routeErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				routeErr = fmt.Errorf("MCP handler panicked: %v", r)
			}
		}()
		mcpResponse, routeErr = p.routeMcpMethod(ctx, server, message)
	}()

	if routeErr != nil {
		return p.sendMcpErrorResponse(ctx, requestID, message, -32603, routeErr.Error())
	}

	return p.sendMcpResponse(ctx, requestID, mcpResponse)
}

// routeMcpMethod dispatches JSONRPC methods to server handlers.
func (p *Protocol) routeMcpMethod(ctx context.Context, server McpServer, msg map[string]any) (map[string]any, error) {
	method := getString(msg, "method")
	params, _ := msg["params"].(map[string]any)
	msgID := msg["id"]

	switch method {
	case "initialize":
		return map[string]any{
			"jsonrpc": "2.0",
			"id":      msgID,
			"result": map[string]any{
				"protocolVersion": "2024-11-05",
				"capabilities":    map[string]any{"tools": map[string]any{}},
				"serverInfo": map[string]any{
					"name":    server.Name(),
					"version": server.Version(),
				},
			},
		}, nil

	case "tools/list":
		tools, err := server.ListTools(ctx)
		if err != nil {
			return nil, err
		}
		toolsData := make([]map[string]any, len(tools))
		for i, t := range tools {
			toolsData[i] = map[string]any{
				"name":        t.Name,
				"description": t.Description,
				"inputSchema": t.InputSchema,
			}
		}
		return map[string]any{
			"jsonrpc": "2.0",
			"id":      msgID,
			"result":  map[string]any{"tools": toolsData},
		}, nil

	case "tools/call":
		if params == nil {
			params = make(map[string]any)
		}
		name := getString(params, "name")
		args, _ := params["arguments"].(map[string]any)
		if args == nil {
			args = make(map[string]any)
		}

		result, err := server.CallTool(ctx, name, args)
		if err != nil {
			return nil, err
		}

		content := make([]map[string]any, len(result.Content))
		for i, c := range result.Content {
			item := map[string]any{"type": c.Type}
			switch c.Type {
			case "text":
				item["text"] = c.Text
			case "image":
				item["data"] = c.Data
				item["mimeType"] = c.MimeType
			}
			content[i] = item
		}

		respData := map[string]any{"content": content}
		if result.IsError {
			respData["isError"] = true
		}
		return map[string]any{
			"jsonrpc": "2.0",
			"id":      msgID,
			"result":  respData,
		}, nil

	case "notifications/initialized":
		// Notification - no response required per JSONRPC spec
		return map[string]any{"jsonrpc": "2.0", "result": map[string]any{}}, nil

	default:
		return nil, fmt.Errorf("method '%s' not found", method)
	}
}

// sendMcpResponse sends an MCP success response.
func (p *Protocol) sendMcpResponse(ctx context.Context, requestID string, mcpResp map[string]any) error {
	response := SDKControlResponse{
		Type: MessageTypeControlResponse,
		Response: Response{
			Subtype:   ResponseSubtypeSuccess,
			RequestID: requestID,
			Response:  map[string]any{"mcp_response": mcpResp},
		},
	}
	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal MCP response: %w", err)
	}
	return p.transport.Write(ctx, append(data, '\n'))
}

// sendMcpErrorResponse sends an MCP JSONRPC error response.
func (p *Protocol) sendMcpErrorResponse(ctx context.Context, requestID string, msg map[string]any, code int, message string) error {
	errorResp := map[string]any{
		"jsonrpc": "2.0",
		"id":      msg["id"],
		"error": map[string]any{
			"code":    code,
			"message": message,
		},
	}
	return p.sendMcpResponse(ctx, requestID, errorResp)
}
