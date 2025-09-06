package claudecode

import (
	"github.com/severity1/claude-code-sdk-go/internal/shared"
)

// Re-export shared error types for public API compatibility
type SDKError = shared.SDKError
type BaseError = shared.BaseError
type ConnectionError = shared.ConnectionError
type CLINotFoundError = shared.CLINotFoundError
type ProcessError = shared.ProcessError
type JSONDecodeError = shared.JSONDecodeError
type MessageParseError = shared.MessageParseError

// Re-export constructor functions
var NewConnectionError = shared.NewConnectionError
var NewCLINotFoundError = shared.NewCLINotFoundError
var NewProcessError = shared.NewProcessError
var NewJSONDecodeError = shared.NewJSONDecodeError
var NewMessageParseError = shared.NewMessageParseError
