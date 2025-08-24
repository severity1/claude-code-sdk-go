// Package parser provides JSON message parsing functionality.
package parser

import (
	"encoding/json"
	"strings"
)

const (
	// MaxBufferSize is the maximum buffer size to prevent memory exhaustion (1MB).
	MaxBufferSize = 1024 * 1024
)

// Parser handles JSON message parsing with speculative parsing and buffer management.
type Parser struct {
	buffer        strings.Builder
	maxBufferSize int
}

// New creates a new JSON parser.
func New() *Parser {
	return &Parser{
		maxBufferSize: MaxBufferSize,
	}
}

// ProcessLine processes a line of JSON input with speculative parsing.
func (p *Parser) ProcessLine(line string) ([]any, error) {
	// TODO: Implement speculative JSON parsing logic
	// Handle multiple JSON objects on single line
	// Handle embedded newlines in JSON strings
	// Implement buffer overflow protection
	return nil, nil
}

// ParseMessage parses a single JSON message into the appropriate type.
func (p *Parser) ParseMessage(data map[string]any) (any, error) {
	// TODO: Implement message type discrimination
	// Parse based on "type" field
	return nil, nil
}

// Reset clears the internal buffer.
func (p *Parser) Reset() {
	p.buffer.Reset()
}

// BufferSize returns the current buffer size.
func (p *Parser) BufferSize() int {
	return p.buffer.Len()
}

// processJSONLine attempts to parse accumulated buffer as JSON.
func (p *Parser) processJSONLine(line string) (any, error) {
	p.buffer.WriteString(line)

	// Check buffer size limit
	if p.buffer.Len() > p.maxBufferSize {
		p.buffer.Reset()
		return nil, &json.UnmarshalTypeError{} // Return buffer overflow error
	}

	// Try to parse JSON
	var data any
	if err := json.Unmarshal([]byte(p.buffer.String()), &data); err != nil {
		// Not complete JSON yet, continue accumulating
		return nil, nil
	}

	// Successfully parsed, reset buffer
	p.buffer.Reset()
	return data, nil
}
