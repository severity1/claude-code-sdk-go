// Package protocol provides control request/response handling for interrupts.
package protocol

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

// Controller manages control protocol requests and responses.
type Controller struct {
	requestCounter  int64
	pendingRequests map[string]chan ControlResponse
}

// ControlRequest represents a control request message.
type ControlRequest struct {
	Type      string                 `json:"type"`
	RequestID string                 `json:"request_id"`
	Request   map[string]interface{} `json:"request"`
}

// ControlResponse represents a control response message.
type ControlResponse struct {
	Type      string                 `json:"type"`
	RequestID string                 `json:"request_id"`
	Response  map[string]interface{} `json:"response"`
	Subtype   string                 `json:"subtype,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// New creates a new protocol controller.
func New() *Controller {
	return &Controller{
		pendingRequests: make(map[string]chan ControlResponse),
	}
}

// SendInterrupt sends an interrupt control request.
func (c *Controller) SendInterrupt(ctx context.Context) error {
	requestID := c.generateRequestID()

	request := ControlRequest{
		Type:      "control_request",
		RequestID: requestID,
		Request:   map[string]interface{}{"subtype": "interrupt"},
	}

	// TODO: Send request through transport
	_ = request // Suppress unused variable warning

	// Wait for response with polling
	responseChan := make(chan ControlResponse, 1)
	c.pendingRequests[requestID] = responseChan
	defer delete(c.pendingRequests, requestID)

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case response := <-responseChan:
			if response.Subtype == "error" {
				return fmt.Errorf("interrupt failed: %s", response.Error)
			}
			return nil
		case <-ticker.C:
			// Continue polling
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// HandleResponse processes an incoming control response.
func (c *Controller) HandleResponse(response ControlResponse) {
	if responseChan, exists := c.pendingRequests[response.RequestID]; exists {
		select {
		case responseChan <- response:
		default:
			// Channel full, drop response
		}
	}
}

// generateRequestID creates a unique request ID.
func (c *Controller) generateRequestID() string {
	counter := atomic.AddInt64(&c.requestCounter, 1)
	// TODO: Add random component for uniqueness
	return fmt.Sprintf("req_%d", counter)
}
