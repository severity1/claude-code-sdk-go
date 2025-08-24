package claudecode

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

// ConnectionError represents connection-related failures.
type ConnectionError struct {
	BaseError
}

func (e *ConnectionError) Type() string {
	return "connection_error"
}

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

func (e *CLINotFoundError) Type() string {
	return "cli_not_found_error"
}

func NewCLINotFoundError(path string, message string) *CLINotFoundError {
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

func (e *ProcessError) Type() string {
	return "process_error"
}

func (e *ProcessError) Error() string {
	return fmt.Sprintf("%s (exit code: %d, stderr: %s)", e.message, e.ExitCode, e.Stderr)
}

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
	Line     string
	Position int
}

func (e *JSONDecodeError) Type() string {
	return "json_decode_error"
}

func NewJSONDecodeError(line string, position int, cause error) *JSONDecodeError {
	return &JSONDecodeError{
		BaseError: BaseError{message: "Failed to decode JSON", cause: cause},
		Line:      line,
		Position:  position,
	}
}

// MessageParseError represents message structure parsing failures.
type MessageParseError struct {
	BaseError
	Data interface{}
}

func (e *MessageParseError) Type() string {
	return "message_parse_error"
}

func NewMessageParseError(message string, data interface{}) *MessageParseError {
	return &MessageParseError{
		BaseError: BaseError{message: message},
		Data:      data,
	}
}
