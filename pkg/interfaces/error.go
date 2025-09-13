package interfaces

import "fmt"

// SDKError represents errors specific to the Claude Code SDK.
// Embeds the standard error interface and adds type information.
type SDKError interface {
	error
	Type() string
}

// BaseError provides common error functionality for SDK errors.
type BaseError struct {
	message string
	cause   error
}

func (e *BaseError) Error() string {
	if e.cause != nil {
		return e.message + ": " + e.cause.Error()
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

// DiscoveryError represents service discovery failures.
type DiscoveryError struct {
	BaseError
	Service string
}

// Type returns the error type for DiscoveryError.
func (e *DiscoveryError) Type() string {
	return "discovery_error"
}

// NewDiscoveryError creates a new DiscoveryError.
func NewDiscoveryError(service, message string, cause error) *DiscoveryError {
	return &DiscoveryError{
		BaseError: BaseError{message: message, cause: cause},
		Service:   service,
	}
}

// ValidationError represents validation failures.
type ValidationError struct {
	BaseError
	Field string
	Value interface{}
}

// Type returns the error type for ValidationError.
func (e *ValidationError) Type() string {
	return "validation_error"
}

// NewValidationError creates a new ValidationError.
func NewValidationError(field, message string, value interface{}) *ValidationError {
	return &ValidationError{
		BaseError: BaseError{message: message},
		Field:     field,
		Value:     value,
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

// Unwrap returns the original error for error chain compatibility.
func (e *JSONDecodeError) Unwrap() error {
	return e.OriginalError
}

// NewJSONDecodeError creates a new JSONDecodeError.
func NewJSONDecodeError(line string, position int, cause error) *JSONDecodeError {
	const maxLineDisplayLength = 100
	// Match Python behavior: truncate line to maxLineDisplayLength chars
	truncatedLine := line
	if len(line) > maxLineDisplayLength {
		truncatedLine = line[:maxLineDisplayLength]
	}
	message := fmt.Sprintf("Failed to decode JSON: %s...", truncatedLine)

	return &JSONDecodeError{
		BaseError:     BaseError{message: message},
		Line:          line,
		Position:      position,
		OriginalError: cause,
	}
}

// MessageParseError represents message structure parsing failures.
type MessageParseError struct {
	BaseError
	Data interface{}
}

// Type returns the error type for MessageParseError.
func (e *MessageParseError) Type() string {
	return "message_parse_error"
}

// NewMessageParseError creates a new MessageParseError.
func NewMessageParseError(message string, data interface{}) *MessageParseError {
	return &MessageParseError{
		BaseError: BaseError{message: message},
		Data:      data,
	}
}
