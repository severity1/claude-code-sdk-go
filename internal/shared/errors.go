// Package shared provides shared types and interfaces used across internal packages.
package shared

import "fmt"

// SDKError is the base interface for all Claude Code SDK errors.
type SDKError interface {
	error
	Type() string
}

// BaseError provides common error functionality.
type BaseError struct {
	message string
	cause   error
}

func (e *BaseError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.message, e.cause)
	}
	return e.message
}

func (e *BaseError) Unwrap() error {
	return e.cause
}

// Type returns the error type for BaseError.
func (e *BaseError) Type() string {
	return "base_error"
}

// ConnectionError represents connection-related failures.
type ConnectionError struct {
	BaseError
}

// Type returns the error type for ConnectionError.
func (e *ConnectionError) Type() string {
	return "connection_error"
}

// NewConnectionError creates a new ConnectionError.
func NewConnectionError(message string, cause error) *ConnectionError {
	return &ConnectionError{
		BaseError: BaseError{message: message, cause: cause},
	}
}

// CLINotFoundError indicates the Claude CLI was not found.
type CLINotFoundError struct {
	BaseError
	Path string
}

// Type returns the error type for CLINotFoundError.
func (e *CLINotFoundError) Type() string {
	return "cli_not_found_error"
}

// NewCLINotFoundError creates a new CLINotFoundError.
func NewCLINotFoundError(path, message string) *CLINotFoundError {
	// Match Python behavior: if path provided, format as "message: path"
	if path != "" {
		message = fmt.Sprintf("%s: %s", message, path)
	}
	return &CLINotFoundError{
		BaseError: BaseError{message: message},
		Path:      path,
	}
}

// ProcessError represents subprocess execution failures.
type ProcessError struct {
	BaseError
	ExitCode int
	Stderr   string
}

// Type returns the error type for ProcessError.
func (e *ProcessError) Type() string {
	return "process_error"
}

func (e *ProcessError) Error() string {
	message := e.message
	if e.ExitCode != 0 {
		message = fmt.Sprintf("%s (exit code: %d)", message, e.ExitCode)
	}
	if e.Stderr != "" {
		message = fmt.Sprintf("%s\nError output: %s", message, e.Stderr)
	}
	return message
}

// NewProcessError creates a new ProcessError.
func NewProcessError(message string, exitCode int, stderr string) *ProcessError {
	return &ProcessError{
		BaseError: BaseError{message: message},
		ExitCode:  exitCode,
		Stderr:    stderr,
	}
}

// JSONDecodeError represents JSON parsing failures.
type JSONDecodeError struct {
	BaseError
	Line          string
	Position      int
	OriginalError error
}

// Type returns the error type for JSONDecodeError.
func (e *JSONDecodeError) Type() string {
	return "json_decode_error"
}

const maxLineDisplayLength = 100

// NewJSONDecodeError creates a new JSONDecodeError.
func NewJSONDecodeError(line string, position int, cause error) *JSONDecodeError {
	// Match Python behavior: truncate line to maxLineDisplayLength chars and add ...
	truncatedLine := line
	if len(line) > maxLineDisplayLength {
		truncatedLine = line[:maxLineDisplayLength]
	}
	message := fmt.Sprintf("Failed to decode JSON: %s...", truncatedLine)

	return &JSONDecodeError{
		BaseError:     BaseError{message: message}, // Don't include cause in message
		Line:          line,
		Position:      position,
		OriginalError: cause, // Store separately like Python
	}
}

func (e *JSONDecodeError) Unwrap() error {
	return e.OriginalError
}

// MessageParseError represents message structure parsing failures.
type MessageParseError struct {
	BaseError
	Data any
}

// Type returns the error type for MessageParseError.
func (e *MessageParseError) Type() string {
	return "message_parse_error"
}

// NewMessageParseError creates a new MessageParseError.
func NewMessageParseError(message string, data any) *MessageParseError {
	return &MessageParseError{
		BaseError: BaseError{message: message},
		Data:      data,
	}
}
