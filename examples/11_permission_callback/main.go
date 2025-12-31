// Package main demonstrates the Permission Callback System.
//
// This example shows how to use permission callbacks to control
// tool usage at runtime. Permission callbacks enable:
// - Security policy enforcement (allow/deny specific tools)
// - Path-based access control (restrict file access to certain directories)
// - Audit logging of all tool usage requests
// - Dynamic permission decisions based on context
//
// NOTE: Permission callbacks are invoked when the CLI sends "can_use_tool"
// requests to the SDK. This happens when the CLI is configured to check
// with the SDK before using tools. The callbacks demonstrate the correct
// API usage pattern for handling these permission check requests.
//
// Run: go run main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	claudecode "github.com/severity1/claude-agent-sdk-go"
)

func main() {
	fmt.Println("Claude Code SDK - Permission Callback Example")
	fmt.Println("==============================================")
	fmt.Println()

	// Get absolute path to demo directory for path-based restrictions
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	demoDir := filepath.Join(wd, "demo")

	// Example 1: Basic tool filtering (allow Read, deny Write)
	fmt.Println("--- Example 1: Tool-Based Permission Control ---")
	fmt.Println("Policy: Allow Read tool, deny Write/Edit tools")
	fmt.Println()
	runToolFilterExample()

	// Example 2: Path-based access control
	fmt.Println()
	fmt.Println("--- Example 2: Path-Based Access Control ---")
	fmt.Printf("Policy: Only allow reading files in %s\n", demoDir)
	fmt.Println("         Block access to 'sensitive' files")
	fmt.Println()
	runPathBasedExample(demoDir)

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
		// Allow only Read tool, deny everything else
		switch toolName {
		case "Read":
			fmt.Printf("  [ALLOW] Tool: %s\n", toolName)
			return claudecode.NewPermissionResultAllow(), nil
		case "Write", "Edit":
			fmt.Printf("  [DENY]  Tool: %s - Write operations not permitted\n", toolName)
			return claudecode.NewPermissionResultDeny("Write operations are not allowed in read-only mode"), nil
		default:
			fmt.Printf("  [DENY]  Tool: %s - Not in allowlist\n", toolName)
			return claudecode.NewPermissionResultDeny(fmt.Sprintf("Tool %s is not permitted", toolName)), nil
		}
	})

	// Ask Claude to read a file (should be allowed)
	fmt.Println("Asking Claude to read demo/public_data.txt...")

	err := claudecode.WithClient(ctx, func(client claudecode.Client) error {
		if err := client.Query(ctx, "Read the file demo/public_data.txt and summarize its contents in one sentence."); err != nil {
			return err
		}

		return streamResponse(ctx, client)
	}, permissionCallback, claudecode.WithMaxTurns(3))

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

// runPathBasedExample demonstrates path-based access control
func runPathBasedExample(allowedDir string) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// Permission callback with path-based restrictions
	permissionCallback := claudecode.WithCanUseTool(func(
		_ context.Context,
		toolName string,
		input map[string]any,
		_ claudecode.ToolPermissionContext,
	) (claudecode.PermissionResult, error) {
		// Only filter Read tool for path-based access
		if toolName != "Read" {
			return claudecode.NewPermissionResultDeny("Only Read tool is allowed"), nil
		}

		// Extract file path from input
		filePath, ok := input["file_path"].(string)
		if !ok {
			return claudecode.NewPermissionResultDeny("Missing file_path parameter"), nil
		}

		// Resolve to absolute path for comparison
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			return claudecode.NewPermissionResultDeny("Invalid file path"), nil
		}

		// Check 1: Must be within allowed directory
		if !strings.HasPrefix(absPath, allowedDir) {
			fmt.Printf("  [DENY]  Path outside allowed directory: %s\n", filePath)
			return claudecode.NewPermissionResultDeny(
				fmt.Sprintf("Access denied: file must be within %s", allowedDir),
			), nil
		}

		// Check 2: Block sensitive files
		if strings.Contains(strings.ToLower(filepath.Base(absPath)), "sensitive") {
			fmt.Printf("  [DENY]  Sensitive file blocked: %s\n", filePath)
			return claudecode.NewPermissionResultDeny("Access to sensitive files is prohibited"), nil
		}

		fmt.Printf("  [ALLOW] Path within allowed directory: %s\n", filePath)
		return claudecode.NewPermissionResultAllow(), nil
	})

	err := claudecode.WithClient(ctx, func(client claudecode.Client) error {
		// First, try to read the public file (should succeed)
		fmt.Println("Attempt 1: Reading demo/public_data.txt (should succeed)...")
		if err := client.Query(ctx, "Read demo/public_data.txt and tell me what it says in one line."); err != nil {
			return err
		}
		if err := streamResponse(ctx, client); err != nil {
			return err
		}

		// Then, try to read sensitive file (should be blocked)
		fmt.Println("\nAttempt 2: Reading demo/sensitive_data.txt (should be blocked)...")
		if err := client.Query(ctx, "Read demo/sensitive_data.txt"); err != nil {
			return err
		}
		return streamResponse(ctx, client)
	}, permissionCallback, claudecode.WithMaxTurns(5))

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

	fmt.Println("Asking Claude to explore the demo directory...")

	err := claudecode.WithClient(ctx, func(client claudecode.Client) error {
		if err := client.Query(ctx, "List what files exist in the demo directory, then read public_data.txt and summarize it briefly."); err != nil {
			return err
		}
		return streamResponse(ctx, client)
	}, permissionCallback, claudecode.WithMaxTurns(5))

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
