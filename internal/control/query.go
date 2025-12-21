package control

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// Query handles bidirectional control protocol communication with the CLI.
// This mirrors the Python SDK's Query class in _internal/query.py.
type Query struct {
	transport Transport

	// Request tracking (matches Python pending_control_responses/results)
	requestCounter atomic.Int64
	pendingReqs    map[string]chan ResponsePayload
	pendingMu      sync.RWMutex

	// Hook callbacks (matches Python hook_callbacks)
	hookCallbacks  map[string]HookHandler
	nextCallbackID atomic.Int64
	hookMu         sync.RWMutex

	// Permission callback (matches Python can_use_tool)
	canUseTool CanUseToolHandler

	// State
	initialized  atomic.Bool
	initResponse *InitializeResponse
	initTimeout  time.Duration
	closed       atomic.Bool

	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc
}

// Option configures a Query instance.
type Option func(*Query)

// WithInitTimeout sets the initialization handshake timeout.
func WithInitTimeout(d time.Duration) Option {
	return func(q *Query) {
		q.initTimeout = d
	}
}

// WithCanUseTool sets the handler for can_use_tool permission requests.
func WithCanUseTool(h CanUseToolHandler) Option {
	return func(q *Query) {
		q.canUseTool = h
	}
}

// New creates a new Query instance for control protocol communication.
// This mirrors Python Query.__init__.
func New(transport Transport, opts ...Option) *Query {
	q := &Query{
		transport:     transport,
		pendingReqs:   make(map[string]chan ResponsePayload),
		hookCallbacks: make(map[string]HookHandler),
		initTimeout:   getDefaultInitTimeout(),
	}

	for _, opt := range opts {
		opt(q)
	}

	return q
}

// getDefaultInitTimeout returns the default init timeout from environment or default.
// Matches Python: CLAUDE_CODE_STREAM_CLOSE_TIMEOUT env var (in ms), minimum 60s.
func getDefaultInitTimeout() time.Duration {
	if ms := os.Getenv("CLAUDE_CODE_STREAM_CLOSE_TIMEOUT"); ms != "" {
		if v, err := strconv.Atoi(ms); err == nil {
			timeout := time.Duration(v) * time.Millisecond
			if timeout < DefaultInitTimeout {
				timeout = DefaultInitTimeout
			}
			return timeout
		}
	}
	return DefaultInitTimeout
}

// Start begins reading messages from the transport.
// This mirrors Python Query.start.
func (q *Query) Start(ctx context.Context) error {
	if q.closed.Load() {
		return fmt.Errorf("query already closed")
	}

	q.ctx, q.cancel = context.WithCancel(ctx)
	return nil
}

// Initialize performs the initialization handshake with the CLI.
// This mirrors Python Query.initialize.
func (q *Query) Initialize(ctx context.Context, hooks map[string][]HookMatcher) (*InitializeResponse, error) {
	if q.initialized.Load() {
		return q.initResponse, nil
	}

	// Build hooks configuration for initialization
	hooksConfig := make(map[string]any)
	for event, matchers := range hooks {
		if len(matchers) > 0 {
			matcherConfigs := make([]map[string]any, 0, len(matchers))
			for _, matcher := range matchers {
				matcherConfig := map[string]any{
					"matcher":         matcher.Matcher,
					"hookCallbackIds": matcher.HookCallbackIDs,
				}
				if matcher.Timeout != nil {
					matcherConfig["timeout"] = *matcher.Timeout
				}
				matcherConfigs = append(matcherConfigs, matcherConfig)
			}
			hooksConfig[event] = matcherConfigs
		}
	}

	// Send initialize request
	request := map[string]any{
		"subtype": SubtypeInitialize,
	}
	if len(hooksConfig) > 0 {
		request["hooks"] = hooksConfig
	}

	// Use initialization timeout
	initCtx, cancel := context.WithTimeout(ctx, q.initTimeout)
	defer cancel()

	response, err := q.SendRequest(initCtx, request)
	if err != nil {
		return nil, fmt.Errorf("initialization handshake failed: %w", err)
	}

	// Parse response
	initResp := &InitializeResponse{}
	if commands, ok := response["commands"].([]any); ok {
		for _, cmd := range commands {
			if cmdStr, ok := cmd.(string); ok {
				initResp.Commands = append(initResp.Commands, cmdStr)
			}
		}
	}
	if style, ok := response["output_style"].(string); ok {
		initResp.OutputStyle = style
	}

	q.initResponse = initResp
	q.initialized.Store(true)

	return initResp, nil
}

// Interrupt sends an interrupt control request.
// This mirrors Python Query.interrupt.
func (q *Query) Interrupt(ctx context.Context) error {
	request := map[string]any{
		"subtype": SubtypeInterrupt,
	}
	_, err := q.SendRequest(ctx, request)
	return err
}

// SetPermissionMode changes the permission mode during conversation.
// This mirrors Python Query.set_permission_mode.
func (q *Query) SetPermissionMode(ctx context.Context, mode string) error {
	request := map[string]any{
		"subtype": SubtypeSetPermissionMode,
		"mode":    mode,
	}
	_, err := q.SendRequest(ctx, request)
	return err
}

// SetModel changes the AI model during conversation.
// This mirrors Python Query.set_model.
func (q *Query) SetModel(ctx context.Context, model *string) error {
	request := map[string]any{
		"subtype": SubtypeSetModel,
	}
	if model != nil {
		request["model"] = *model
	}
	_, err := q.SendRequest(ctx, request)
	return err
}

// RewindFiles rewinds tracked files to their state at a specific user message.
// This mirrors Python Query.rewind_files.
func (q *Query) RewindFiles(ctx context.Context, userMessageID string) error {
	request := map[string]any{
		"subtype":         SubtypeRewindFiles,
		"user_message_id": userMessageID,
	}
	_, err := q.SendRequest(ctx, request)
	return err
}

// SendRequest sends a control request and waits for response.
// This mirrors Python Query._send_control_request.
func (q *Query) SendRequest(ctx context.Context, request map[string]any) (map[string]any, error) {
	if q.closed.Load() {
		return nil, fmt.Errorf("query is closed")
	}

	// Generate unique request ID matching Python format: req_{counter}_{hex}
	reqID := q.generateRequestID()

	// Create response channel
	respChan := make(chan ResponsePayload, 1)

	// Register pending request
	q.pendingMu.Lock()
	q.pendingReqs[reqID] = respChan
	q.pendingMu.Unlock()

	defer func() {
		q.pendingMu.Lock()
		delete(q.pendingReqs, reqID)
		q.pendingMu.Unlock()
	}()

	// Build and send control request
	ctrlReq := Request{
		Type:      TypeControlRequest,
		RequestID: reqID,
		Request:   request,
	}

	data, err := json.Marshal(ctrlReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal control request: %w", err)
	}

	// Append newline for JSON line protocol
	data = append(data, '\n')

	if err := q.transport.Write(ctx, data); err != nil {
		return nil, fmt.Errorf("failed to send control request: %w", err)
	}

	// Wait for response with context timeout
	select {
	case resp := <-respChan:
		if resp.Subtype == SubtypeError {
			return nil, fmt.Errorf("control request failed: %s", resp.Error)
		}
		return resp.Response, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("control request timeout: %w", ctx.Err())
	}
}

// HandleIncoming routes an incoming control message from the CLI.
// This handles both control_response (responses to our requests) and
// control_request (incoming requests from CLI like can_use_tool, hook_callback).
func (q *Query) HandleIncoming(ctx context.Context, msg map[string]any) error {
	msgType, _ := msg["type"].(string)

	switch msgType {
	case TypeControlResponse:
		return q.handleControlResponse(msg)
	case TypeControlRequest:
		return q.handleControlRequest(ctx, msg)
	default:
		// Not a control message, ignore
		return nil
	}
}

// handleControlResponse routes a response to the waiting request.
func (q *Query) handleControlResponse(msg map[string]any) error {
	respData, ok := msg["response"].(map[string]any)
	if !ok {
		return fmt.Errorf("invalid control response format")
	}

	requestID, _ := respData["request_id"].(string)
	if requestID == "" {
		return fmt.Errorf("control response missing request_id")
	}

	// Find pending request
	q.pendingMu.RLock()
	respChan, exists := q.pendingReqs[requestID]
	q.pendingMu.RUnlock()

	if !exists {
		// No pending request, might be stale or already handled
		return nil
	}

	// Build response payload
	payload := ResponsePayload{
		Subtype:   respData["subtype"].(string),
		RequestID: requestID,
	}

	if resp, ok := respData["response"].(map[string]any); ok {
		payload.Response = resp
	}
	if errMsg, ok := respData["error"].(string); ok {
		payload.Error = errMsg
	}

	// Send to waiting goroutine (non-blocking)
	select {
	case respChan <- payload:
	default:
		// Channel full or closed, response will be dropped
	}

	return nil
}

// handleControlRequest handles incoming requests from CLI (can_use_tool, hook_callback, etc).
func (q *Query) handleControlRequest(ctx context.Context, msg map[string]any) error {
	requestID, _ := msg["request_id"].(string)
	reqData, ok := msg["request"].(map[string]any)
	if !ok {
		return q.sendErrorResponse(ctx, requestID, "invalid control request format")
	}

	subtype, _ := reqData["subtype"].(string)

	var response map[string]any
	var handleErr error

	switch subtype {
	case SubtypeCanUseTool:
		response, handleErr = q.handleCanUseTool(ctx, reqData)
	case SubtypeHookCallback:
		response, handleErr = q.handleHookCallback(ctx, reqData)
	default:
		handleErr = fmt.Errorf("unsupported control request subtype: %s", subtype)
	}

	if handleErr != nil {
		return q.sendErrorResponse(ctx, requestID, handleErr.Error())
	}

	return q.sendSuccessResponse(ctx, requestID, response)
}

// handleCanUseTool processes a can_use_tool request from CLI.
func (q *Query) handleCanUseTool(ctx context.Context, reqData map[string]any) (map[string]any, error) {
	if q.canUseTool == nil {
		return nil, fmt.Errorf("canUseTool callback is not provided")
	}

	// Parse request
	req := CanUseToolRequest{
		Subtype:  SubtypeCanUseTool,
		ToolName: reqData["tool_name"].(string),
	}
	if input, ok := reqData["input"].(map[string]any); ok {
		req.Input = input
	}
	if suggestions, ok := reqData["permission_suggestions"].([]any); ok {
		req.PermissionSuggestions = suggestions
	}
	if blocked, ok := reqData["blocked_path"].(string); ok {
		req.BlockedPath = &blocked
	}

	// Call handler
	resp, err := q.canUseTool(ctx, req)
	if err != nil {
		return nil, err
	}

	// Build response
	result := map[string]any{
		"behavior": resp.Behavior,
	}
	if resp.UpdatedInput != nil {
		result["updatedInput"] = resp.UpdatedInput
	}
	if resp.UpdatedPermissions != nil {
		result["updatedPermissions"] = resp.UpdatedPermissions
	}
	if resp.Message != "" {
		result["message"] = resp.Message
	}
	if resp.Interrupt {
		result["interrupt"] = resp.Interrupt
	}

	return result, nil
}

// handleHookCallback processes a hook_callback request from CLI.
func (q *Query) handleHookCallback(ctx context.Context, reqData map[string]any) (map[string]any, error) {
	callbackID, _ := reqData["callback_id"].(string)

	q.hookMu.RLock()
	handler, exists := q.hookCallbacks[callbackID]
	q.hookMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no hook callback found for ID: %s", callbackID)
	}

	input := reqData["input"]
	var toolUseID *string
	if id, ok := reqData["tool_use_id"].(string); ok {
		toolUseID = &id
	}

	response, err := handler(ctx, input, toolUseID)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// sendSuccessResponse sends a success response back to CLI.
func (q *Query) sendSuccessResponse(ctx context.Context, requestID string, response map[string]any) error {
	resp := Response{
		Type: TypeControlResponse,
		Response: ResponsePayload{
			Subtype:   SubtypeSuccess,
			RequestID: requestID,
			Response:  response,
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("failed to marshal success response: %w", err)
	}

	data = append(data, '\n')
	return q.transport.Write(ctx, data)
}

// sendErrorResponse sends an error response back to CLI.
func (q *Query) sendErrorResponse(ctx context.Context, requestID string, errMsg string) error {
	resp := Response{
		Type: TypeControlResponse,
		Response: ResponsePayload{
			Subtype:   SubtypeError,
			RequestID: requestID,
			Error:     errMsg,
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("failed to marshal error response: %w", err)
	}

	data = append(data, '\n')
	return q.transport.Write(ctx, data)
}

// RegisterHookCallback registers a hook callback handler.
func (q *Query) RegisterHookCallback(callbackID string, handler HookHandler) {
	q.hookMu.Lock()
	defer q.hookMu.Unlock()
	q.hookCallbacks[callbackID] = handler
}

// GenerateCallbackID generates a new unique callback ID.
func (q *Query) GenerateCallbackID() string {
	n := q.nextCallbackID.Add(1)
	return fmt.Sprintf("hook_%d", n-1)
}

// Close shuts down the query and cancels pending operations.
// This mirrors Python Query.close.
func (q *Query) Close() error {
	if q.closed.Swap(true) {
		return nil // Already closed
	}

	if q.cancel != nil {
		q.cancel()
	}

	// Signal all pending requests to fail
	q.pendingMu.Lock()
	for _, ch := range q.pendingReqs {
		close(ch)
	}
	q.pendingReqs = make(map[string]chan ResponsePayload)
	q.pendingMu.Unlock()

	return nil
}

// IsInitialized returns whether the query has completed initialization.
func (q *Query) IsInitialized() bool {
	return q.initialized.Load()
}

// GetInitResponse returns the initialization response, or nil if not initialized.
func (q *Query) GetInitResponse() *InitializeResponse {
	if !q.initialized.Load() {
		return nil
	}
	return q.initResponse
}

// generateRequestID generates a unique request ID in Python SDK format.
// Format: req_{counter}_{random_hex}
func (q *Query) generateRequestID() string {
	n := q.requestCounter.Add(1)
	hexBytes := make([]byte, 4)
	_, _ = rand.Read(hexBytes)
	return fmt.Sprintf("req_%d_%x", n, hexBytes)
}
