package interfaces

// SDKError represents errors specific to the Claude Code SDK.
// Embeds the standard error interface and adds type information.
type SDKError interface {
	error
	Type() string
}
