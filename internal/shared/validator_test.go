package shared

import (
	"sync"
	"testing"
)

// Test functions first (primary purpose)

func TestStreamValidator_CompleteStream(t *testing.T) {
	validator := NewStreamValidator()

	// Simulate complete stream: request → result → final result
	assistantMsg := &AssistantMessage{
		Content: []ContentBlock{
			&ToolUseBlock{
				ToolUseID: "tool_123",
				Name:      "read_file",
			},
		},
	}

	userMsg := &UserMessage{
		Content: []ContentBlock{
			&ToolResultBlock{
				ToolUseID: "tool_123",
				Content:   "file contents",
			},
		},
	}

	resultMsg := &ResultMessage{}

	validator.TrackMessage(assistantMsg)
	validator.TrackMessage(userMsg)
	validator.TrackMessage(resultMsg)
	validator.MarkStreamEnd()

	// Validate no issues
	issues := validator.GetIssues()
	if len(issues) != 0 {
		t.Errorf("Expected no issues, got %d: %+v", len(issues), issues)
	}

	// Validate stats
	stats := validator.GetStats()
	if stats.ToolsRequested != 1 {
		t.Errorf("Expected 1 tool requested, got %d", stats.ToolsRequested)
	}
	if stats.ToolsReceived != 1 {
		t.Errorf("Expected 1 tool received, got %d", stats.ToolsReceived)
	}
	if len(stats.PendingTools) != 0 {
		t.Errorf("Expected no pending tools, got %d", len(stats.PendingTools))
	}
	if !stats.HasResult {
		t.Error("Expected result message to be tracked")
	}
}

func TestStreamValidator_MissingToolResult(t *testing.T) {
	validator := NewStreamValidator()

	// Simulate stream with missing tool result
	assistantMsg := &AssistantMessage{
		Content: []ContentBlock{
			&ToolUseBlock{
				ToolUseID: "tool_123",
				Name:      "read_file",
			},
		},
	}

	validator.TrackMessage(assistantMsg)
	validator.MarkStreamEnd()

	// Check for missing tool result issue
	issues := validator.GetIssues()
	if len(issues) != 2 {
		t.Errorf("Expected 2 issues (missing tool result + missing result message), got %d", len(issues))
	}

	foundMissingTool := false
	for _, issue := range issues {
		if issue.Type == "missing_tool_result" && issue.ToolUseID == "tool_123" {
			foundMissingTool = true
		}
	}
	if !foundMissingTool {
		t.Error("Expected missing_tool_result issue for tool_123")
	}

	// Validate stats
	stats := validator.GetStats()
	if stats.ToolsRequested != 1 {
		t.Errorf("Expected 1 tool requested, got %d", stats.ToolsRequested)
	}
	if stats.ToolsReceived != 0 {
		t.Errorf("Expected 0 tools received, got %d", stats.ToolsReceived)
	}
	if len(stats.PendingTools) != 1 {
		t.Errorf("Expected 1 pending tool, got %d", len(stats.PendingTools))
	}
}

func TestStreamValidator_ExtraToolResult(t *testing.T) {
	validator := NewStreamValidator()

	// Simulate tool result without request
	userMsg := &UserMessage{
		Content: []ContentBlock{
			&ToolResultBlock{
				ToolUseID: "tool_999",
				Content:   "unexpected result",
			},
		},
	}

	validator.TrackMessage(userMsg)
	validator.MarkStreamEnd()

	// Check for extra tool result issue
	issues := validator.GetIssues()
	if len(issues) != 1 {
		t.Errorf("Expected 1 issue, got %d", len(issues))
	}

	if issues[0].Type != "extra_tool_result" {
		t.Errorf("Expected extra_tool_result issue, got %s", issues[0].Type)
	}
	if issues[0].ToolUseID != "tool_999" {
		t.Errorf("Expected tool_999, got %s", issues[0].ToolUseID)
	}
}

func TestStreamValidator_MultipleTools(t *testing.T) {
	validator := NewStreamValidator()

	// Track multiple tools
	assistantMsg := &AssistantMessage{
		Content: []ContentBlock{
			&ToolUseBlock{ToolUseID: "tool_1", Name: "read_file"},
			&ToolUseBlock{ToolUseID: "tool_2", Name: "write_file"},
			&ToolUseBlock{ToolUseID: "tool_3", Name: "list_files"},
		},
	}

	userMsg := &UserMessage{
		Content: []ContentBlock{
			&ToolResultBlock{ToolUseID: "tool_1", Content: "result1"},
			&ToolResultBlock{ToolUseID: "tool_2", Content: "result2"},
			&ToolResultBlock{ToolUseID: "tool_3", Content: "result3"},
		},
	}

	resultMsg := &ResultMessage{}

	validator.TrackMessage(assistantMsg)
	validator.TrackMessage(userMsg)
	validator.TrackMessage(resultMsg)
	validator.MarkStreamEnd()

	// Validate complete stream
	issues := validator.GetIssues()
	if len(issues) != 0 {
		t.Errorf("Expected no issues, got %d: %+v", len(issues), issues)
	}

	stats := validator.GetStats()
	if stats.ToolsRequested != 3 {
		t.Errorf("Expected 3 tools requested, got %d", stats.ToolsRequested)
	}
	if stats.ToolsReceived != 3 {
		t.Errorf("Expected 3 tools received, got %d", stats.ToolsReceived)
	}
	if len(stats.PendingTools) != 0 {
		t.Errorf("Expected no pending tools, got %d", len(stats.PendingTools))
	}
}

func TestStreamValidator_MissingResultMessage(t *testing.T) {
	validator := NewStreamValidator()

	// Complete tool cycle but no result message
	assistantMsg := &AssistantMessage{
		Content: []ContentBlock{
			&ToolUseBlock{ToolUseID: "tool_1", Name: "read_file"},
		},
	}

	userMsg := &UserMessage{
		Content: []ContentBlock{
			&ToolResultBlock{ToolUseID: "tool_1", Content: "result"},
		},
	}

	validator.TrackMessage(assistantMsg)
	validator.TrackMessage(userMsg)
	validator.MarkStreamEnd()

	// Check for missing result message
	issues := validator.GetIssues()
	if len(issues) != 1 {
		t.Errorf("Expected 1 issue, got %d", len(issues))
	}

	if issues[0].Type != "missing_result_message" {
		t.Errorf("Expected missing_result_message, got %s", issues[0].Type)
	}

	stats := validator.GetStats()
	if stats.HasResult {
		t.Error("Expected HasResult to be false")
	}
}

func TestStreamValidator_ThreadSafety(t *testing.T) {
	validator := NewStreamValidator()
	var wg sync.WaitGroup

	// Simulate concurrent message tracking
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Each goroutine tracks a tool request and result
			assistantMsg := &AssistantMessage{
				Content: []ContentBlock{
					&ToolUseBlock{
						ToolUseID: "tool_" + string(rune('0'+id)),
						Name:      "test",
					},
				},
			}

			userMsg := &UserMessage{
				Content: []ContentBlock{
					&ToolResultBlock{
						ToolUseID: "tool_" + string(rune('0'+id)),
						Content:   "result",
					},
				},
			}

			validator.TrackMessage(assistantMsg)
			validator.TrackMessage(userMsg)

			// Also read stats concurrently
			_ = validator.GetStats()
			_ = validator.GetIssues()
		}(i)
	}

	wg.Wait()
	validator.MarkStreamEnd()

	// Verify all tools tracked correctly
	stats := validator.GetStats()
	if stats.ToolsRequested != 10 {
		t.Errorf("Expected 10 tools requested, got %d", stats.ToolsRequested)
	}
	if stats.ToolsReceived != 10 {
		t.Errorf("Expected 10 tools received, got %d", stats.ToolsReceived)
	}
}

func TestStreamValidator_PartialCompletion(t *testing.T) {
	validator := NewStreamValidator()

	// Request 3 tools, receive 2
	assistantMsg := &AssistantMessage{
		Content: []ContentBlock{
			&ToolUseBlock{ToolUseID: "tool_1", Name: "read_file"},
			&ToolUseBlock{ToolUseID: "tool_2", Name: "write_file"},
			&ToolUseBlock{ToolUseID: "tool_3", Name: "list_files"},
		},
	}

	userMsg := &UserMessage{
		Content: []ContentBlock{
			&ToolResultBlock{ToolUseID: "tool_1", Content: "result1"},
			&ToolResultBlock{ToolUseID: "tool_2", Content: "result2"},
			// tool_3 result missing
		},
	}

	validator.TrackMessage(assistantMsg)
	validator.TrackMessage(userMsg)
	validator.MarkStreamEnd()

	// Check for missing tool result
	issues := validator.GetIssues()
	if len(issues) != 2 {
		t.Errorf("Expected 2 issues, got %d", len(issues))
	}

	foundMissingTool := false
	for _, issue := range issues {
		if issue.Type == "missing_tool_result" && issue.ToolUseID == "tool_3" {
			foundMissingTool = true
		}
	}
	if !foundMissingTool {
		t.Error("Expected missing_tool_result issue for tool_3")
	}

	stats := validator.GetStats()
	if stats.ToolsRequested != 3 {
		t.Errorf("Expected 3 tools requested, got %d", stats.ToolsRequested)
	}
	if stats.ToolsReceived != 2 {
		t.Errorf("Expected 2 tools received, got %d", stats.ToolsReceived)
	}
	if len(stats.PendingTools) != 1 {
		t.Errorf("Expected 1 pending tool, got %d", len(stats.PendingTools))
	}
}
