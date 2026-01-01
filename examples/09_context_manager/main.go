// Package main demonstrates WithClient context manager pattern.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/severity1/claude-agent-sdk-go"
)

func main() {
	fmt.Println("Claude Code SDK - WithClient Context Manager")
	fmt.Println("Automatic resource management vs manual pattern")

	ctx := context.Background()
	question := "What are the benefits of using context managers in programming?"

	// WithClient pattern (recommended)
	fmt.Println("\n--- WithClient Pattern (Recommended) ---")
	fmt.Println("[+] Automatic connect/disconnect")
	fmt.Println("[+] Guaranteed cleanup on errors")

	if err := demonstrateWithClient(ctx, question); err != nil {
		log.Printf("WithClient failed: %v", err)
	}

	// Manual pattern (still supported)
	fmt.Println("\n--- Manual Pattern (Still Supported) ---")
	fmt.Println("[!] Manual connect/disconnect required")
	fmt.Println("[!] Easy to forget cleanup")

	if err := demonstrateManualPattern(ctx, question); err != nil {
		log.Printf("Manual pattern failed: %v", err)
	}

	// Error handling demonstration
	fmt.Println("\n--- Error Handling ---")
	if err := demonstrateErrorScenarios(ctx); err != nil {
		log.Printf("Error demo failed: %v", err)
	}

	fmt.Println("\nRecommendation: Use WithClient for automatic resource management")
}

func demonstrateWithClient(ctx context.Context, question string) error {
	fmt.Println("Using WithClient for automatic resource management...")

	return claudecode.WithClient(ctx, func(client claudecode.Client) error {
		fmt.Println("Connected! Client managed automatically")

		if err := client.Query(ctx, question); err != nil {
			return fmt.Errorf("query failed: %w", err)
		}

		fmt.Println("\nResponse (first lines):")
		if err := showFirstLines(ctx, client, 3, 80); err != nil {
			return err
		}
		fmt.Println("[+] WithClient will handle cleanup automatically")
		return nil
	})
}

func demonstrateManualPattern(ctx context.Context, question string) error {
	fmt.Println("Using manual Connect/Disconnect pattern...")

	client := claudecode.NewClient()

	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("connect failed: %w", err)
	}

	defer func() {
		fmt.Println("Manual cleanup...")
		if err := client.Disconnect(); err != nil {
			log.Printf("Disconnect warning: %v", err)
		}
		fmt.Println("[+] Manual cleanup completed")
	}()

	fmt.Println("Connected manually")

	if err := client.Query(ctx, question); err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	fmt.Println("\nResponse (first lines):")
	if err := showFirstLines(ctx, client, 3, 80); err != nil {
		return err
	}
	return nil
}

func demonstrateErrorScenarios(ctx context.Context) error {
	fmt.Println("Testing WithClient error handling...")

	// Test context cancellation
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel() // Cancel immediately

	err := claudecode.WithClient(cancelCtx, func(client claudecode.Client) error {
		return client.Query(cancelCtx, "This will be cancelled")
	})
	if err != nil {
		fmt.Printf("[+] WithClient handled cancellation: %v\n", err)
	}

	// Test function error
	err = claudecode.WithClient(ctx, func(client claudecode.Client) error {
		return fmt.Errorf("simulated application error")
	})
	if err != nil {
		fmt.Printf("[+] WithClient propagated error: %v\n", err)
		fmt.Println("   Connection was still cleaned up automatically")
	}

	return nil
}

// showFirstLines displays first lines of response from client
func showFirstLines(ctx context.Context, client claudecode.Client, maxLines, maxWidth int) error {
	msgChan := client.ReceiveMessages(ctx)
	linesShown := 0

	for linesShown < maxLines {
		select {
		case message := <-msgChan:
			if message == nil {
				return nil
			}

			switch msg := message.(type) {
			case *claudecode.AssistantMessage:
				for _, block := range msg.Content {
					if textBlock, ok := block.(*claudecode.TextBlock); ok {
						if linesShown < maxLines {
							text := textBlock.Text
							if len(text) > maxWidth {
								text = text[:maxWidth] + "..."
							}
							fmt.Printf("  %s\n", text)
							linesShown++
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

	// Drain remaining messages
	drainMessages(msgChan)
	return nil
}

// drainMessages consumes remaining messages from a channel
func drainMessages(msgChan <-chan claudecode.Message) {
	for {
		select {
		case message := <-msgChan:
			if message == nil {
				return
			}
		default:
			return
		}
	}
}
