// Package main demonstrates the Permission Callback System.
//
// This example shows how to use permission callbacks to control
// tool usage at runtime. Permission callbacks enable:
// - Security policy enforcement (allow/deny specific tools)
// - Path-based access control (restrict file writes to certain directories)
// - Audit logging of all tool usage requests
// - Dynamic permission decisions based on context
//
// IMPORTANT: Permission callbacks are only invoked for tools that would
// normally prompt the user for permission. Read-only operations (Read, Glob,
// Grep) are auto-approved by the CLI and do NOT trigger callbacks.
// Write operations (Write, Edit, Bash) DO trigger callbacks.
//
// The permission flow is:
// PreToolUse Hook -> Deny Rules -> Allow Rules -> Ask Rules -> Permission Mode -> canUseTool Callback
//
// Run: go run main.go
package main

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	claudecode "github.com/severity1/claude-agent-sdk-go"
)

// exampleDir returns the directory containing this source file.
func exampleDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(file)
}

func main() {
	fmt.Println("Claude Code SDK - Permission Callback Example")
	fmt.Println("==============================================")
	fmt.Println()

	// Example 1: Basic tool filtering (allow Write/Bash, deny Edit)
	fmt.Println("--- Example 1: Tool-Based Permission Control ---")
	fmt.Println("Policy: Allow Write/Bash tools, deny Edit tool")
	fmt.Println("Note: Read/Glob are auto-approved by CLI and don't trigger callbacks")
	fmt.Println()
	runToolFilterExample()

	// Example 2: Path-based access control for writes
	fmt.Println()
	fmt.Println("--- Example 2: Path-Based Access Control ---")
	fmt.Println("Policy: Only allow writes to /tmp directory")
	fmt.Println("        Block filenames containing 'sensitive'")
	fmt.Println()
	runPathBasedExample()

	// Example 3: Audit logging
	fmt.Println()
	fmt.Println("--- Example 3: Audit Logging ---")
	fmt.Println("Policy: Log all tool requests for security auditing")
	fmt.Println()
	runAuditLoggingExample()

	fmt.Println()
	fmt.Println("Permission callback examples completed!")
}

// runToolFilterExample demonstrates basic tool allow/deny filtering
func runToolFilterExample() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Permission callback that filters by tool name
	permissionCallback := claudecode.WithCanUseTool(func(
		_ context.Context,
		toolName string,
		_ map[string]any,
		_ claudecode.ToolPermissionContext,
	) (claudecode.PermissionResult, error) {
		// Log all permission requests
		fmt.Printf("  [CALLBACK] Tool: %s\n", toolName)

		// Allow Bash and Write for demo, deny Edit
		switch toolName {
		case "Bash", "Write":
			fmt.Printf("  [ALLOW] Tool: %s\n", toolName)
			return claudecode.NewPermissionResultAllow(), nil
		case "Edit":
			fmt.Printf("  [DENY]  Tool: %s - Edit not permitted in this example\n", toolName)
			return claudecode.NewPermissionResultDeny("Edit operations are not allowed"), nil
		default:
			// Allow other tools (Read, Glob, etc. don't trigger callbacks)
			fmt.Printf("  [ALLOW] Tool: %s\n", toolName)
			return claudecode.NewPermissionResultAllow(), nil
		}
	})

	// Ask Claude to create a file (requires Write permission)
	fmt.Println("Asking Claude to create a test file in /tmp...")

	err := claudecode.WithClient(ctx, func(client claudecode.Client) error {
		if err := client.Query(ctx, "Create a file at /tmp/sdk_permission_test.txt with the text 'Hello from SDK permission callback test!'"); err != nil {
			return err
		}

		return streamResponse(ctx, client)
	}, permissionCallback,
		claudecode.WithPermissionMode(claudecode.PermissionModeAcceptEdits), // Auto-approve edits, callback controls access
		claudecode.WithMaxTurns(3),
		claudecode.WithCwd(exampleDir()))

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

// runPathBasedExample demonstrates path-based access control for Write operations
func runPathBasedExample() {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// Permission callback with path-based restrictions for Write
	permissionCallback := claudecode.WithCanUseTool(func(
		_ context.Context,
		toolName string,
		input map[string]any,
		_ claudecode.ToolPermissionContext,
	) (claudecode.PermissionResult, error) {
		fmt.Printf("  [CALLBACK] Tool: %s\n", toolName)

		// For Write operations, enforce path restrictions
		if toolName == "Write" {
			filePath, ok := input["file_path"].(string)
			if !ok {
				return claudecode.NewPermissionResultDeny("Missing file_path parameter"), nil
			}

			// Only allow writes to /tmp
			if !strings.HasPrefix(filePath, "/tmp/") {
				fmt.Printf("  [DENY]  Write outside /tmp: %s\n", filePath)
				return claudecode.NewPermissionResultDeny(
					fmt.Sprintf("Writes only allowed to /tmp, not: %s", filePath),
				), nil
			}

			// Block sensitive filenames
			if strings.Contains(strings.ToLower(filepath.Base(filePath)), "sensitive") {
				fmt.Printf("  [DENY]  Sensitive filename blocked: %s\n", filePath)
				return claudecode.NewPermissionResultDeny("Cannot create files with 'sensitive' in name"), nil
			}

			fmt.Printf("  [ALLOW] Write to allowed path: %s\n", filePath)
			return claudecode.NewPermissionResultAllow(), nil
		}

		// Allow Bash for verification
		if toolName == "Bash" {
			fmt.Printf("  [ALLOW] Bash command allowed\n")
			return claudecode.NewPermissionResultAllow(), nil
		}

		// Allow other tools
		return claudecode.NewPermissionResultAllow(), nil
	})

	err := claudecode.WithClient(ctx, func(client claudecode.Client) error {
		// First, write to /tmp (should succeed)
		fmt.Println("Attempt 1: Writing to /tmp/allowed_test.txt (should succeed)...")
		if err := client.Query(ctx, "Create a file at /tmp/allowed_test.txt with 'Test content'"); err != nil {
			return err
		}
		if err := streamResponse(ctx, client); err != nil {
			return err
		}

		// Then, try to write with 'sensitive' in name (should be blocked)
		fmt.Println("\nAttempt 2: Writing to /tmp/sensitive_data.txt (should be blocked)...")
		if err := client.Query(ctx, "Create a file at /tmp/sensitive_data.txt with 'Secret content'"); err != nil {
			return err
		}
		return streamResponse(ctx, client)
	}, permissionCallback,
		claudecode.WithPermissionMode(claudecode.PermissionModeAcceptEdits), // Auto-approve edits, callback controls access
		claudecode.WithMaxTurns(5),
		claudecode.WithCwd(exampleDir()))

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

// runAuditLoggingExample demonstrates audit logging of all tool requests
func runAuditLoggingExample() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Thread-safe audit log
	var auditLog []AuditEntry
	var auditMu sync.Mutex

	// Permission callback that logs all requests
	permissionCallback := claudecode.WithCanUseTool(func(
		_ context.Context,
		toolName string,
		input map[string]any,
		_ claudecode.ToolPermissionContext,
	) (claudecode.PermissionResult, error) {
		// Create audit entry
		entry := AuditEntry{
			Timestamp: time.Now(),
			Tool:      toolName,
			Input:     input,
			Allowed:   true, // We'll allow everything but log it
		}

		// Thread-safe append to audit log
		auditMu.Lock()
		auditLog = append(auditLog, entry)
		auditMu.Unlock()

		fmt.Printf("  [AUDIT] Tool: %-10s | Input keys: %v\n", toolName, mapKeys(input))

		// Allow all tools (audit-only mode)
		return claudecode.NewPermissionResultAllow(), nil
	})

	fmt.Println("Asking Claude to create a test file (all operations will be logged)...")

	err := claudecode.WithClient(ctx, func(client claudecode.Client) error {
		if err := client.Query(ctx, "Create a file at /tmp/sdk_audit_test.txt with the current timestamp, then run 'cat /tmp/sdk_audit_test.txt' to verify it."); err != nil {
			return err
		}
		return streamResponse(ctx, client)
	}, permissionCallback,
		claudecode.WithPermissionMode(claudecode.PermissionModeAcceptEdits), // Auto-approve edits, callback controls access
		claudecode.WithMaxTurns(5),
		claudecode.WithCwd(exampleDir()))

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Print audit summary
	fmt.Println("\n--- Audit Log Summary ---")
	auditMu.Lock()
	for i, entry := range auditLog {
		status := "ALLOWED"
		if !entry.Allowed {
			status = "DENIED"
		}
		fmt.Printf("  %d. [%s] %s at %s\n",
			i+1, status, entry.Tool, entry.Timestamp.Format("15:04:05"))
	}
	fmt.Printf("Total tool requests: %d\n", len(auditLog))
	auditMu.Unlock()
}

// AuditEntry represents a logged tool usage request
type AuditEntry struct {
	Timestamp time.Time
	Tool      string
	Input     map[string]any
	Allowed   bool
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
						// Show first 150 chars of response
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

// mapKeys returns the keys of a map as a slice
func mapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
