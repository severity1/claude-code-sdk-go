package claudecode

import (
	"github.com/severity1/claude-code-sdk-go/internal/shared"
)

// SDKError represents the base interface for all SDK errors.
type SDKError = shared.SDKError

// BaseError provides common error functionality across the SDK.
type BaseError = shared.BaseError

// ConnectionError represents errors that occur during CLI connection.
type ConnectionError = shared.ConnectionError

// CLINotFoundError indicates that the Claude Code CLI was not found.
type CLINotFoundError = shared.CLINotFoundError

// ProcessError represents errors from the CLI process execution.
type ProcessError = shared.ProcessError

// JSONDecodeError represents JSON parsing errors from CLI responses.
type JSONDecodeError = shared.JSONDecodeError

// MessageParseError represents errors parsing message content.
type MessageParseError = shared.MessageParseError

// NewConnectionError creates a new connection error.
var NewConnectionError = shared.NewConnectionError

// NewCLINotFoundError creates a new CLI not found error.
var NewCLINotFoundError = shared.NewCLINotFoundError

// NewProcessError creates a new process error.
var NewProcessError = shared.NewProcessError

// NewJSONDecodeError creates a new JSON decode error.
var NewJSONDecodeError = shared.NewJSONDecodeError

// NewMessageParseError creates a new message parse error.
var NewMessageParseError = shared.NewMessageParseError
