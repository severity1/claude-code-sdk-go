package shared

import (
	"sync"
)

// StreamValidator tracks tool requests and results to detect incomplete streams.
type StreamValidator struct {
	mu              sync.RWMutex
	toolsRequested  map[string]bool   // Set of all tool_use IDs requested
	toolsReceived   map[string]bool   // Set of all tool_result IDs received
	pendingToolsSet map[string]bool   // Set of tool IDs awaiting results
	hasResultMessage bool              // Whether we've seen a result message
	streamEnded     bool               // Whether stream has ended
	issues          []StreamIssue      // Validation issues found
}

// StreamIssue represents a validation issue found in the stream.
type StreamIssue struct {
	Type        string `json:"type"`        // "missing_tool_result", "extra_tool_result", etc.
	Description string `json:"description"` // Human-readable description
	ToolUseID   string `json:"tool_use_id,omitempty"` // Related tool use ID if applicable
}

// StreamStats provides statistics about the message stream.
type StreamStats struct {
	ToolsRequested int      `json:"tools_requested"` // Total tools requested
	ToolsReceived  int      `json:"tools_received"`  // Total tool results received
	PendingTools   []string `json:"pending_tools"`   // Tool IDs still awaiting results
	HasResult      bool     `json:"has_result"`      // Whether result message was seen
	StreamEnded    bool     `json:"stream_ended"`    // Whether stream has ended
}

// NewStreamValidator creates a new stream validator.
func NewStreamValidator() *StreamValidator {
	return &StreamValidator{
		toolsRequested:  make(map[string]bool),
		toolsReceived:   make(map[string]bool),
		pendingToolsSet: make(map[string]bool),
		issues:          []StreamIssue{},
	}
}

// TrackMessage processes a message and updates validation state.
func (v *StreamValidator) TrackMessage(msg Message) {
	v.mu.Lock()
	defer v.mu.Unlock()

	switch m := msg.(type) {
	case *AssistantMessage:
		// Track tool use requests
		for _, block := range m.Content {
			if toolUse, ok := block.(*ToolUseBlock); ok {
				v.toolsRequested[toolUse.ToolUseID] = true
				v.pendingToolsSet[toolUse.ToolUseID] = true
			}
		}

	case *UserMessage:
		// Track tool results
		if blocks, ok := m.Content.([]ContentBlock); ok {
			for _, block := range blocks {
				if toolResult, ok := block.(*ToolResultBlock); ok {
					v.toolsReceived[toolResult.ToolUseID] = true
					delete(v.pendingToolsSet, toolResult.ToolUseID)

					// Check for extra tool results (results without requests)
					if !v.toolsRequested[toolResult.ToolUseID] {
						v.issues = append(v.issues, StreamIssue{
							Type:        "extra_tool_result",
							Description: "Received tool result without corresponding tool request",
							ToolUseID:   toolResult.ToolUseID,
						})
					}
				}
			}
		}

	case *ResultMessage:
		v.hasResultMessage = true
	}
}

// MarkStreamEnd marks the stream as ended and performs final validation.
func (v *StreamValidator) MarkStreamEnd() {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.streamEnded = true

	// Check for missing tool results
	for toolID := range v.pendingToolsSet {
		v.issues = append(v.issues, StreamIssue{
			Type:        "missing_tool_result",
			Description: "Tool was requested but result was never received",
			ToolUseID:   toolID,
		})
	}

	// Check for missing result message
	if len(v.toolsRequested) > 0 && !v.hasResultMessage {
		v.issues = append(v.issues, StreamIssue{
			Type:        "missing_result_message",
			Description: "Stream ended without result message",
		})
	}
}

// GetIssues returns all validation issues found.
func (v *StreamValidator) GetIssues() []StreamIssue {
	v.mu.RLock()
	defer v.mu.RUnlock()

	// Return a copy to prevent external modification
	issues := make([]StreamIssue, len(v.issues))
	copy(issues, v.issues)
	return issues
}

// GetStats returns current stream statistics.
func (v *StreamValidator) GetStats() StreamStats {
	v.mu.RLock()
	defer v.mu.RUnlock()

	pendingTools := make([]string, 0, len(v.pendingToolsSet))
	for toolID := range v.pendingToolsSet {
		pendingTools = append(pendingTools, toolID)
	}

	return StreamStats{
		ToolsRequested: len(v.toolsRequested),
		ToolsReceived:  len(v.toolsReceived),
		PendingTools:   pendingTools,
		HasResult:      v.hasResultMessage,
		StreamEnded:    v.streamEnded,
	}
}

// HasIssues returns whether any validation issues were found.
func (v *StreamValidator) HasIssues() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return len(v.issues) > 0
}
