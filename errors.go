package claudecode

import (
	"github.com/severity1/claude-code-sdk-go/pkg/interfaces"
)

// SDKError represents the base interface for all SDK errors.
type SDKError = interfaces.SDKError

// BaseError provides common error functionality across the SDK.
type BaseError = interfaces.BaseError

// ConnectionError represents errors that occur during CLI connection.
type ConnectionError = interfaces.ConnectionError

// CLINotFoundError indicates that the Claude Code CLI was not found.
type CLINotFoundError = interfaces.CLINotFoundError

// ProcessError represents errors from the CLI process execution.
type ProcessError = interfaces.ProcessError

// JSONDecodeError represents JSON parsing errors from CLI responses.
type JSONDecodeError = interfaces.JSONDecodeError

// MessageParseError represents errors parsing message content.
type MessageParseError = interfaces.MessageParseError

// DiscoveryError represents service discovery failures.
type DiscoveryError = interfaces.DiscoveryError

// ValidationError represents validation failures.
type ValidationError = interfaces.ValidationError

// NewConnectionError creates a new connection error.
var NewConnectionError = interfaces.NewConnectionError

// NewCLINotFoundError creates a new CLI not found error.
var NewCLINotFoundError = interfaces.NewCLINotFoundError

// NewProcessError creates a new process error.
var NewProcessError = interfaces.NewProcessError

// NewJSONDecodeError creates a new JSON decode error.
var NewJSONDecodeError = interfaces.NewJSONDecodeError

// NewMessageParseError creates a new message parse error.
var NewMessageParseError = interfaces.NewMessageParseError

// NewDiscoveryError creates a new discovery error.
var NewDiscoveryError = interfaces.NewDiscoveryError

// NewValidationError creates a new validation error.
var NewValidationError = interfaces.NewValidationError
