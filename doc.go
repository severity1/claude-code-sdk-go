// Package claudecode provides the Claude Code SDK for Go.
//
// This SDK enables programmatic interaction with Claude Code CLI through two main APIs:
// - Query() for one-shot requests with automatic cleanup
// - Client for bidirectional streaming conversations
//
// The SDK follows Go-native patterns with goroutines and channels instead of
// async/await, providing context-first design for cancellation and timeouts.
//
// Example usage:
//
//	import "github.com/severity1/claude-code-sdk-go"
//
//	// One-shot query
//	messages, err := claudecode.Query(ctx, "Hello, Claude!")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Streaming client
//	client := claudecode.NewClient(
//		claudecode.WithSystemPrompt("You are a helpful assistant"),
//	)
//	defer client.Close()
//
// The SDK provides 100% feature parity with the Python SDK while embracing
// Go idioms and patterns.
package claudecode

// Version represents the current version of the Claude Code SDK for Go.
const Version = "0.1.0"
