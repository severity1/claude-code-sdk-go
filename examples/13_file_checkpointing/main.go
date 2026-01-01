// Package main demonstrates the File Checkpointing and Rewind System.
//
// This example shows how to use file checkpointing to track file changes
// during a Claude Code session and rewind them to a previous state.
// File checkpointing enables:
// - Safe experimentation with file modifications
// - Undo/rollback capability for file changes
// - Recovery from unwanted modifications
// - Testing different approaches without permanent changes
//
// Key concepts:
// - WithFileCheckpointing(): Enables file change tracking
// - UserMessage.UUID: Checkpoint identifier for each user message
// - RewindFiles(): Reverts files to their state at a specific checkpoint
//
// NOTE: File checkpointing only works with the Client API (streaming mode).
// It is not available with the Query API (one-shot mode) because the control
// protocol required for rewind operations needs a persistent connection.
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
	fmt.Println("Claude Code SDK - File Checkpointing Example")
	fmt.Println("=============================================")
	fmt.Println()

	// Example 1: Basic file checkpointing setup and UUID capture
	fmt.Println("--- Example 1: Capturing Checkpoints ---")
	fmt.Println("Setup: Enable checkpointing and capture UserMessage UUIDs")
	fmt.Println()
	runCheckpointCaptureExample()

	// Example 2: File modification tracking
	fmt.Println()
	fmt.Println("--- Example 2: File Modification Tracking ---")
	fmt.Println("Setup: Track file changes during a session")
	fmt.Println()
	runModificationTrackingExample()

	// Example 3: Rewind demonstration (conceptual)
	fmt.Println()
	fmt.Println("--- Example 3: Rewind Workflow ---")
	fmt.Println("Setup: Demonstrate the rewind pattern")
	fmt.Println()
	runRewindWorkflowExample()

	fmt.Println()
	fmt.Println("File checkpointing examples completed!")
}

// CheckpointEntry represents a captured checkpoint
type CheckpointEntry struct {
	Timestamp time.Time
	UUID      string
	Query     string
}

// runCheckpointCaptureExample demonstrates capturing UserMessage UUIDs
func runCheckpointCaptureExample() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Thread-safe checkpoint storage
	var checkpoints []CheckpointEntry
	var mu sync.Mutex

	fmt.Println("Asking Claude to read a file (capturing checkpoint)...")

	err := claudecode.WithClient(ctx, func(client claudecode.Client) error {
		// Send a query
		query := "Read the file demo/notes.txt and tell me what version it is."
		if err := client.Query(ctx, query); err != nil {
			return err
		}

		// Process messages and capture UserMessage UUIDs
		return streamWithCheckpoints(ctx, client, func(uuid, q string) {
			mu.Lock()
			checkpoints = append(checkpoints, CheckpointEntry{
				Timestamp: time.Now(),
				UUID:      uuid,
				Query:     q,
			})
			mu.Unlock()
		}, query)
	}, claudecode.WithFileCheckpointing(), claudecode.WithMaxTurns(3), claudecode.WithCwd(exampleDir()))

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Print captured checkpoints
	fmt.Println("\n--- Captured Checkpoints ---")
	mu.Lock()
	for i, cp := range checkpoints {
		fmt.Printf("  %d. UUID: %s\n", i+1, truncate(cp.UUID, 36))
		fmt.Printf("     Query: %s\n", truncate(cp.Query, 50))
		fmt.Printf("     Time: %s\n", cp.Timestamp.Format("15:04:05"))
	}
	if len(checkpoints) == 0 {
		fmt.Println("  (No checkpoints captured - CLI may not have sent UserMessage)")
	}
	mu.Unlock()
}

// runModificationTrackingExample demonstrates file modification tracking
func runModificationTrackingExample() {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// Track modifications
	var modifications []string
	var mu sync.Mutex

	fmt.Println("Asking Claude to describe a file (read-only operation)...")

	err := claudecode.WithClient(ctx, func(client claudecode.Client) error {
		// First query - read the file
		if err := client.Query(ctx, "Read demo/notes.txt and describe its contents briefly."); err != nil {
			return err
		}
		if err := streamWithModifications(ctx, client, &modifications, &mu); err != nil {
			return err
		}

		return nil
	}, claudecode.WithFileCheckpointing(), claudecode.WithMaxTurns(5), claudecode.WithCwd(exampleDir()))

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Print modification summary
	fmt.Println("\n--- Modification Summary ---")
	mu.Lock()
	if len(modifications) > 0 {
		for i, mod := range modifications {
			fmt.Printf("  %d. %s\n", i+1, mod)
		}
	} else {
		fmt.Println("  No file modifications detected (read-only operation)")
	}
	mu.Unlock()
}

// runRewindWorkflowExample demonstrates the rewind workflow pattern
func runRewindWorkflowExample() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Capture the checkpoint UUID
	var capturedUUID string
	var mu sync.Mutex

	fmt.Println("Demonstrating the rewind workflow pattern...")
	fmt.Println()
	fmt.Println("Step 1: Enable checkpointing and start session")

	err := claudecode.WithClient(ctx, func(client claudecode.Client) error {
		// Step 2: Capture checkpoint from first interaction
		fmt.Println("Step 2: Send query and capture checkpoint UUID")

		if err := client.Query(ctx, "What is 2 + 2? Answer in one word."); err != nil {
			return err
		}

		// Capture UUID from UserMessage
		// Process ALL messages until channel returns nil to avoid race condition
		// where UserMessage might still be buffered after ResultMessage
		msgChan := client.ReceiveMessages(ctx)
		var streamErr error
		for {
			select {
			case message := <-msgChan:
				if message == nil {
					// Channel closed - all messages processed
					if streamErr != nil {
						return streamErr
					}
					goto done
				}

				switch msg := message.(type) {
				case *claudecode.UserMessage:
					if msg.UUID != nil {
						mu.Lock()
						capturedUUID = *msg.UUID
						mu.Unlock()
						fmt.Printf("         Captured UUID: %s\n", truncate(*msg.UUID, 36))
					}
				case *claudecode.AssistantMessage:
					for _, block := range msg.Content {
						if textBlock, ok := block.(*claudecode.TextBlock); ok {
							text := textBlock.Text
							if len(text) > 100 {
								text = text[:100] + "..."
							}
							fmt.Printf("         Response: %s\n", strings.ReplaceAll(text, "\n", " "))
						}
					}
				case *claudecode.ResultMessage:
					if msg.IsError {
						if msg.Result != nil {
							streamErr = fmt.Errorf("error: %s", *msg.Result)
						} else {
							streamErr = fmt.Errorf("error: unknown error")
						}
					}
					// Don't exit here - continue draining to capture UserMessage
				}
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	done:

		// Step 3: Show how RewindFiles would be called
		fmt.Println()
		fmt.Println("Step 3: RewindFiles usage pattern")

		mu.Lock()
		uuid := capturedUUID
		mu.Unlock()

		if uuid != "" {
			fmt.Println("         To rewind files to this checkpoint:")
			fmt.Printf("         err := client.RewindFiles(ctx, %q)\n", truncate(uuid, 36))
			fmt.Println()
			fmt.Println("         This would revert all file changes made after this point.")

			// Note: We don't actually call RewindFiles here because:
			// 1. No files were modified in this example
			// 2. The CLI needs actual file changes to rewind
			// Uncomment the following to actually call rewind:
			// if err := client.RewindFiles(ctx, uuid); err != nil {
			//     fmt.Printf("         Rewind error: %v\n", err)
			// }
		} else {
			fmt.Println("         (No UUID captured - CLI may not have sent UserMessage)")
		}

		return nil
	}, claudecode.WithFileCheckpointing(), claudecode.WithMaxTurns(3), claudecode.WithCwd(exampleDir()))

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	fmt.Println()
	fmt.Println("Workflow complete. In a real scenario, you would:")
	fmt.Println("  1. Enable checkpointing with WithFileCheckpointing()")
	fmt.Println("  2. Capture UUIDs from UserMessage during streaming")
	fmt.Println("  3. Call RewindFiles(ctx, uuid) to revert file changes")
}

// streamWithCheckpoints processes messages and captures UserMessage UUIDs
func streamWithCheckpoints(
	ctx context.Context,
	client claudecode.Client,
	onCheckpoint func(uuid, query string),
	currentQuery string,
) error {
	msgChan := client.ReceiveMessages(ctx)

	for {
		select {
		case message := <-msgChan:
			if message == nil {
				return nil
			}

			switch msg := message.(type) {
			case *claudecode.UserMessage:
				// Capture the UUID for this user message
				if msg.UUID != nil {
					onCheckpoint(*msg.UUID, currentQuery)
					fmt.Printf("  [CHECKPOINT] UUID captured: %s\n", truncate(*msg.UUID, 24))
				}
			case *claudecode.AssistantMessage:
				for _, block := range msg.Content {
					if textBlock, ok := block.(*claudecode.TextBlock); ok {
						text := textBlock.Text
						if len(text) > 150 {
							text = text[:150] + "..."
						}
						fmt.Printf("  Response: %s\n", strings.ReplaceAll(text, "\n", " "))
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

// streamWithModifications processes messages and tracks tool use for modifications
func streamWithModifications(
	ctx context.Context,
	client claudecode.Client,
	modifications *[]string,
	mu *sync.Mutex,
) error {
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
					switch b := block.(type) {
					case *claudecode.TextBlock:
						text := b.Text
						if len(text) > 150 {
							text = text[:150] + "..."
						}
						fmt.Printf("  Response: %s\n", strings.ReplaceAll(text, "\n", " "))
					case *claudecode.ToolUseBlock:
						// Track potential file modifications
						if b.Name == "Write" || b.Name == "Edit" {
							mu.Lock()
							*modifications = append(*modifications, fmt.Sprintf("%s tool used", b.Name))
							mu.Unlock()
							fmt.Printf("  [MODIFY] %s tool invoked\n", b.Name)
						}
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
