// Package main demonstrates the Hook System for lifecycle events.
//
// This example shows how to use hooks to intercept and respond to
// lifecycle events during Claude Code CLI execution. Hooks enable:
// - Logging and auditing of tool usage (PreToolUse, PostToolUse)
// - Blocking dangerous commands before execution
// - Adding context to tool responses
// - Monitoring session lifecycle events
//
// Hook events supported:
// - PreToolUse: Before a tool executes (can block or modify input)
// - PostToolUse: After a tool executes (can add context)
// - UserPromptSubmit: When user submits a prompt
// - Stop: When session is stopping
// - SubagentStop: When a subagent is stopping
// - PreCompact: Before context compaction
//
// NOTE: Hooks are invoked when the CLI sends hook callback requests
// to the SDK. The callbacks demonstrate the correct API usage pattern
// for handling these lifecycle events.
//
// Run: go run main.go
package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	claudecode "github.com/severity1/claude-agent-sdk-go"
)

func main() {
	fmt.Println("Claude Code SDK - Hook System Example")
	fmt.Println("======================================")
	fmt.Println()

	// Example 1: Basic tool logging with PreToolUse and PostToolUse hooks
	fmt.Println("--- Example 1: Tool Logging Hooks ---")
	fmt.Println("Hook: Log all tool usage before and after execution")
	fmt.Println()
	runToolLoggingExample()

	// Example 2: Blocking dangerous commands
	fmt.Println()
	fmt.Println("--- Example 2: Command Blocking Hook ---")
	fmt.Println("Hook: Block dangerous bash commands before execution")
	fmt.Println()
	runBlockingExample()

	// Example 3: Adding context to tool responses
	fmt.Println()
	fmt.Println("--- Example 3: Context Injection Hook ---")
	fmt.Println("Hook: Add timing information after tool execution")
	fmt.Println()
	runContextInjectionExample()

	fmt.Println()
	fmt.Println("Hook system examples completed!")
}

// runToolLoggingExample demonstrates logging tool usage with hooks
func runToolLoggingExample() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Thread-safe log storage
	var toolLog []ToolLogEntry
	var logMu sync.Mutex

	// PreToolUse hook - log before execution
	preToolHook := claudecode.WithPreToolUseHook("", func(
		_ context.Context,
		input any,
		_ *string,
		_ claudecode.HookContext,
	) (claudecode.HookJSONOutput, error) {
		preInput, ok := input.(*claudecode.PreToolUseHookInput)
		if !ok {
			return claudecode.HookJSONOutput{}, nil
		}

		entry := ToolLogEntry{
			Timestamp: time.Now(),
			Tool:      preInput.ToolName,
			Phase:     "PRE",
		}

		logMu.Lock()
		toolLog = append(toolLog, entry)
		logMu.Unlock()

		fmt.Printf("  [PRE]  Tool: %-10s | Session: %s\n",
			preInput.ToolName, truncate(preInput.SessionID, 12))

		return claudecode.HookJSONOutput{}, nil
	})

	// PostToolUse hook - log after execution
	postToolHook := claudecode.WithPostToolUseHook("", func(
		_ context.Context,
		input any,
		_ *string,
		_ claudecode.HookContext,
	) (claudecode.HookJSONOutput, error) {
		postInput, ok := input.(*claudecode.PostToolUseHookInput)
		if !ok {
			return claudecode.HookJSONOutput{}, nil
		}

		entry := ToolLogEntry{
			Timestamp: time.Now(),
			Tool:      postInput.ToolName,
			Phase:     "POST",
		}

		logMu.Lock()
		toolLog = append(toolLog, entry)
		logMu.Unlock()

		fmt.Printf("  [POST] Tool: %-10s | Response type: %T\n",
			postInput.ToolName, postInput.ToolResponse)

		return claudecode.HookJSONOutput{}, nil
	})

	fmt.Println("Asking Claude to read a file...")

	err := claudecode.WithClient(ctx, func(client claudecode.Client) error {
		if err := client.Query(ctx, "Read the file demo/sample.txt and tell me what it contains in one sentence."); err != nil {
			return err
		}
		return streamResponse(ctx, client)
	}, preToolHook, postToolHook, claudecode.WithMaxTurns(3))

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Print log summary
	fmt.Println("\n--- Tool Log Summary ---")
	logMu.Lock()
	for i, entry := range toolLog {
		fmt.Printf("  %d. [%s] %s at %s\n",
			i+1, entry.Phase, entry.Tool, entry.Timestamp.Format("15:04:05.000"))
	}
	fmt.Printf("Total hook invocations: %d\n", len(toolLog))
	logMu.Unlock()
}

// runBlockingExample demonstrates blocking dangerous commands with PreToolUse hooks
func runBlockingExample() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Dangerous command patterns to block
	dangerousPatterns := []string{"rm -rf", "sudo", "chmod 777", "> /dev/"}

	// PreToolUse hook that blocks dangerous Bash commands
	blockingHook := claudecode.WithPreToolUseHook("Bash", func(
		_ context.Context,
		input any,
		_ *string,
		_ claudecode.HookContext,
	) (claudecode.HookJSONOutput, error) {
		preInput, ok := input.(*claudecode.PreToolUseHookInput)
		if !ok {
			return claudecode.HookJSONOutput{}, nil
		}

		// Extract command from tool input
		command, ok := preInput.ToolInput["command"].(string)
		if !ok {
			return claudecode.HookJSONOutput{}, nil
		}

		// Check for dangerous patterns
		for _, pattern := range dangerousPatterns {
			if strings.Contains(strings.ToLower(command), strings.ToLower(pattern)) {
				fmt.Printf("  [BLOCK] Dangerous command detected: %q\n", truncate(command, 50))

				// Block the command
				decision := "block"
				reason := fmt.Sprintf("Command blocked: contains dangerous pattern '%s'", pattern)
				return claudecode.HookJSONOutput{
					Decision: &decision,
					Reason:   &reason,
				}, nil
			}
		}

		fmt.Printf("  [ALLOW] Safe command: %q\n", truncate(command, 50))
		return claudecode.HookJSONOutput{}, nil
	})

	fmt.Println("Asking Claude to run some commands (dangerous ones will be blocked)...")

	err := claudecode.WithClient(ctx, func(client claudecode.Client) error {
		if err := client.Query(ctx, "Run 'ls -la demo/' to list files, then run 'echo hello' to test."); err != nil {
			return err
		}
		return streamResponse(ctx, client)
	}, blockingHook, claudecode.WithMaxTurns(5))

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

// runContextInjectionExample demonstrates adding context after tool execution
func runContextInjectionExample() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Track tool execution timing
	toolStartTimes := make(map[string]time.Time)
	var timingMu sync.Mutex

	// PreToolUse hook - record start time
	preHook := claudecode.WithPreToolUseHook("", func(
		_ context.Context,
		_ any,
		toolUseID *string,
		_ claudecode.HookContext,
	) (claudecode.HookJSONOutput, error) {
		if toolUseID == nil {
			return claudecode.HookJSONOutput{}, nil
		}

		timingMu.Lock()
		toolStartTimes[*toolUseID] = time.Now()
		timingMu.Unlock()

		return claudecode.HookJSONOutput{}, nil
	})

	// PostToolUse hook - add timing context
	postHook := claudecode.WithPostToolUseHook("", func(
		_ context.Context,
		input any,
		toolUseID *string,
		_ claudecode.HookContext,
	) (claudecode.HookJSONOutput, error) {
		postInput, ok := input.(*claudecode.PostToolUseHookInput)
		if !ok || toolUseID == nil {
			return claudecode.HookJSONOutput{}, nil
		}

		timingMu.Lock()
		startTime, exists := toolStartTimes[*toolUseID]
		delete(toolStartTimes, *toolUseID)
		timingMu.Unlock()

		if !exists {
			return claudecode.HookJSONOutput{}, nil
		}

		duration := time.Since(startTime)
		context := fmt.Sprintf("Tool %s completed in %v", postInput.ToolName, duration)
		fmt.Printf("  [TIMING] %s\n", context)

		// Add timing context for Claude
		return claudecode.HookJSONOutput{
			HookSpecificOutput: claudecode.PostToolUseHookSpecificOutput{
				HookEventName:     "PostToolUse",
				AdditionalContext: &context,
			},
		}, nil
	})

	fmt.Println("Asking Claude to perform file operations (timing will be tracked)...")

	err := claudecode.WithClient(ctx, func(client claudecode.Client) error {
		if err := client.Query(ctx, "List the files in demo/ directory and read sample.txt if it exists."); err != nil {
			return err
		}
		return streamResponse(ctx, client)
	}, preHook, postHook, claudecode.WithMaxTurns(5))

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

// ToolLogEntry represents a logged tool usage event
type ToolLogEntry struct {
	Timestamp time.Time
	Tool      string
	Phase     string // "PRE" or "POST"
}

// streamResponse reads and displays messages from the client
func streamResponse(ctx context.Context, client claudecode.Client) error {
	msgChan := client.ReceiveMessages(ctx)

	for {
		select {
		case message := <-msgChan:
			if message == nil {
				return nil
			}

			switch msg := message.(type) {
			case *claudecode.AssistantMessage:
				for _, block := range msg.Content {
					if textBlock, ok := block.(*claudecode.TextBlock); ok {
						text := textBlock.Text
						if len(text) > 150 {
							text = text[:150] + "..."
						}
						fmt.Printf("Response: %s\n", strings.ReplaceAll(text, "\n", " "))
					}
				}
			case *claudecode.ResultMessage:
				if msg.IsError {
					if msg.Result != nil {
						return fmt.Errorf("error: %s", *msg.Result)
					}
					return fmt.Errorf("error: unknown error")
				}
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// truncate shortens a string to maxLen characters
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
