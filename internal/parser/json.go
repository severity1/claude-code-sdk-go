// Package parser provides JSON message parsing functionality with speculative parsing and buffer management.
package parser

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/severity1/claude-code-sdk-go"
)

const (
	// MaxBufferSize is the maximum buffer size to prevent memory exhaustion (1MB).
	MaxBufferSize = 1024 * 1024
)

// Parser handles JSON message parsing with speculative parsing and buffer management.
// It implements the same speculative parsing strategy as the Python SDK.
type Parser struct {
	buffer        strings.Builder
	maxBufferSize int
	mu            sync.Mutex // Thread safety
}

// New creates a new JSON parser with default buffer size.
func New() *Parser {
	return &Parser{
		maxBufferSize: MaxBufferSize,
	}
}

// ProcessLine processes a line of JSON input with speculative parsing.
// Handles multiple JSON objects on single line and embedded newlines.
func (p *Parser) ProcessLine(line string) ([]claudecode.Message, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	line = strings.TrimSpace(line)
	if line == "" {
		return nil, nil
	}

	var messages []claudecode.Message

	// Handle multiple JSON objects on single line by splitting on newlines
	jsonLines := strings.Split(line, "\n")
	for _, jsonLine := range jsonLines {
		jsonLine = strings.TrimSpace(jsonLine)
		if jsonLine == "" {
			continue
		}

		// Process each JSON line with speculative parsing (unlocked version)
		msg, err := p.processJSONLineUnlocked(jsonLine)
		if err != nil {
			return messages, err
		}
		if msg != nil {
			messages = append(messages, msg)
		}
	}

	return messages, nil
}

// ParseMessage parses a raw JSON object into the appropriate Message type.
// Implements type discrimination based on the "type" field.
func (p *Parser) ParseMessage(data map[string]any) (claudecode.Message, error) {
	msgType, ok := data["type"].(string)
	if !ok {
		return nil, claudecode.NewMessageParseError("missing or invalid type field", data)
	}

	switch msgType {
	case claudecode.MessageTypeUser:
		return p.parseUserMessage(data)
	case claudecode.MessageTypeAssistant:
		return p.parseAssistantMessage(data)
	case claudecode.MessageTypeSystem:
		return p.parseSystemMessage(data)
	case claudecode.MessageTypeResult:
		return p.parseResultMessage(data)
	default:
		return nil, claudecode.NewMessageParseError(
			fmt.Sprintf("unknown message type: %s", msgType),
			data,
		)
	}
}

// Reset clears the internal buffer.
func (p *Parser) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.buffer.Reset()
}

// BufferSize returns the current buffer size.
func (p *Parser) BufferSize() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.buffer.Len()
}

// processJSONLine attempts to parse accumulated buffer as JSON using speculative parsing.
// This is the core of the speculative parsing strategy from the Python SDK.
func (p *Parser) processJSONLine(jsonLine string) (claudecode.Message, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	return p.processJSONLineUnlocked(jsonLine)
}

// processJSONLineUnlocked is the unlocked version of processJSONLine.
// Must be called with mutex already held.
func (p *Parser) processJSONLineUnlocked(jsonLine string) (claudecode.Message, error) {
	p.buffer.WriteString(jsonLine)

	// Check buffer size limit
	if p.buffer.Len() > p.maxBufferSize {
		bufferSize := p.buffer.Len()
		p.buffer.Reset()
		return nil, claudecode.NewJSONDecodeError(
			"buffer overflow",
			0,
			fmt.Errorf("buffer size %d exceeds limit %d", bufferSize, p.maxBufferSize),
		)
	}

	// Attempt speculative JSON parsing
	var rawData map[string]any
	bufferContent := p.buffer.String()

	if err := json.Unmarshal([]byte(bufferContent), &rawData); err != nil {
		// JSON is incomplete - continue accumulating
		// This is NOT an error condition in speculative parsing!
		return nil, nil
	}

	// Successfully parsed complete JSON - reset buffer and parse message
	p.buffer.Reset()
	return p.ParseMessage(rawData)
}

// parseUserMessage parses a user message from raw JSON data.
func (p *Parser) parseUserMessage(data map[string]any) (*claudecode.UserMessage, error) {
	messageData, ok := data["message"].(map[string]any)
	if !ok {
		return nil, claudecode.NewMessageParseError("user message missing message field", data)
	}

	content := messageData["content"]
	if content == nil {
		return nil, claudecode.NewMessageParseError("user message missing content field", data)
	}

	// Handle both string content and array of content blocks
	switch c := content.(type) {
	case string:
		// String content - create directly
		return &claudecode.UserMessage{
			Content: c,
		}, nil
	case []any:
		// Array of content blocks
		blocks := make([]claudecode.ContentBlock, len(c))
		for i, blockData := range c {
			block, err := p.parseContentBlock(blockData)
			if err != nil {
				return nil, fmt.Errorf("failed to parse content block %d: %w", i, err)
			}
			blocks[i] = block
		}
		return &claudecode.UserMessage{
			Content: blocks,
		}, nil
	default:
		return nil, claudecode.NewMessageParseError("invalid user message content type", data)
	}
}

// parseAssistantMessage parses an assistant message from raw JSON data.
func (p *Parser) parseAssistantMessage(data map[string]any) (*claudecode.AssistantMessage, error) {
	messageData, ok := data["message"].(map[string]any)
	if !ok {
		return nil, claudecode.NewMessageParseError("assistant message missing message field", data)
	}

	contentArray, ok := messageData["content"].([]any)
	if !ok {
		return nil, claudecode.NewMessageParseError("assistant message content must be array", data)
	}

	model, ok := messageData["model"].(string)
	if !ok {
		return nil, claudecode.NewMessageParseError("assistant message missing model field", data)
	}

	blocks := make([]claudecode.ContentBlock, len(contentArray))
	for i, blockData := range contentArray {
		block, err := p.parseContentBlock(blockData)
		if err != nil {
			return nil, fmt.Errorf("failed to parse content block %d: %w", i, err)
		}
		blocks[i] = block
	}

	return &claudecode.AssistantMessage{
		Content: blocks,
		Model:   model,
	}, nil
}

// parseSystemMessage parses a system message from raw JSON data.
func (p *Parser) parseSystemMessage(data map[string]any) (*claudecode.SystemMessage, error) {
	subtype, ok := data["subtype"].(string)
	if !ok {
		return nil, claudecode.NewMessageParseError("system message missing subtype field", data)
	}

	return &claudecode.SystemMessage{
		Subtype: subtype,
		Data:    data, // Preserve all original data
	}, nil
}

// parseResultMessage parses a result message from raw JSON data.
func (p *Parser) parseResultMessage(data map[string]any) (*claudecode.ResultMessage, error) {
	result := &claudecode.ResultMessage{}

	// Required fields with validation
	if subtype, ok := data["subtype"].(string); ok {
		result.Subtype = subtype
	} else {
		return nil, claudecode.NewMessageParseError("result message missing subtype field", data)
	}

	if durationMS, ok := data["duration_ms"].(float64); ok {
		result.DurationMs = int(durationMS)
	} else {
		return nil, claudecode.NewMessageParseError("result message missing or invalid duration_ms field", data)
	}

	if durationAPIMS, ok := data["duration_api_ms"].(float64); ok {
		result.DurationAPIMs = int(durationAPIMS)
	} else {
		return nil, claudecode.NewMessageParseError("result message missing or invalid duration_api_ms field", data)
	}

	if isError, ok := data["is_error"].(bool); ok {
		result.IsError = isError
	} else {
		return nil, claudecode.NewMessageParseError("result message missing or invalid is_error field", data)
	}

	if numTurns, ok := data["num_turns"].(float64); ok {
		result.NumTurns = int(numTurns)
	} else {
		return nil, claudecode.NewMessageParseError("result message missing or invalid num_turns field", data)
	}

	if sessionID, ok := data["session_id"].(string); ok {
		result.SessionID = sessionID
	} else {
		return nil, claudecode.NewMessageParseError("result message missing session_id field", data)
	}

	// Optional fields (no validation errors if missing)
	if totalCostUSD, ok := data["total_cost_usd"].(float64); ok {
		result.TotalCostUSD = &totalCostUSD
	}

	if usage, ok := data["usage"].(map[string]any); ok {
		result.Usage = usage
	}

	if resultStr, ok := data["result"].(string); ok {
		result.Result = &resultStr
	}

	return result, nil
}

// parseContentBlock parses a content block based on its type field.
func (p *Parser) parseContentBlock(blockData any) (claudecode.ContentBlock, error) {
	data, ok := blockData.(map[string]any)
	if !ok {
		return nil, claudecode.NewMessageParseError("content block must be an object", blockData)
	}

	blockType, ok := data["type"].(string)
	if !ok {
		return nil, claudecode.NewMessageParseError("content block missing type field", data)
	}

	switch blockType {
	case claudecode.ContentBlockTypeText:
		text, ok := data["text"].(string)
		if !ok {
			return nil, claudecode.NewMessageParseError("text block missing text field", data)
		}
		return &claudecode.TextBlock{Text: text}, nil

	case claudecode.ContentBlockTypeThinking:
		thinking, ok := data["thinking"].(string)
		if !ok {
			return nil, claudecode.NewMessageParseError("thinking block missing thinking field", data)
		}
		signature, _ := data["signature"].(string) // Optional field
		return &claudecode.ThinkingBlock{
			Thinking:  thinking,
			Signature: signature,
		}, nil

	case claudecode.ContentBlockTypeToolUse:
		id, ok := data["id"].(string)
		if !ok {
			return nil, claudecode.NewMessageParseError("tool_use block missing id field", data)
		}
		name, ok := data["name"].(string)
		if !ok {
			return nil, claudecode.NewMessageParseError("tool_use block missing name field", data)
		}
		input, _ := data["input"].(map[string]any)
		if input == nil {
			input = make(map[string]any)
		}
		return &claudecode.ToolUseBlock{
			ID:    id,
			Name:  name,
			Input: input,
		}, nil

	case claudecode.ContentBlockTypeToolResult:
		toolUseID, ok := data["tool_use_id"].(string)
		if !ok {
			return nil, claudecode.NewMessageParseError("tool_result block missing tool_use_id field", data)
		}

		var isError *bool
		if isErrorValue, exists := data["is_error"]; exists {
			if b, ok := isErrorValue.(bool); ok {
				isError = &b
			}
		}

		return &claudecode.ToolResultBlock{
			ToolUseID: toolUseID,
			Content:   data["content"],
			IsError:   isError,
		}, nil

	default:
		return nil, claudecode.NewMessageParseError(
			fmt.Sprintf("unknown content block type: %s", blockType),
			data,
		)
	}
}

// ParseMessages is a convenience function to parse multiple JSON lines.
func ParseMessages(lines []string) ([]claudecode.Message, error) {
	parser := New()
	var allMessages []claudecode.Message

	for i, line := range lines {
		messages, err := parser.ProcessLine(line)
		if err != nil {
			return allMessages, fmt.Errorf("error parsing line %d: %w", i, err)
		}
		allMessages = append(allMessages, messages...)
	}

	return allMessages, nil
}
