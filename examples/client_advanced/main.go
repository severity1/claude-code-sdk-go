// Package main demonstrates advanced features of the Claude Code SDK Client API.
//
// This example shows how to:
// - Use client options (system prompts, model selection, etc.)
// - Handle various error conditions gracefully
// - Implement robust connection management with retries
// - Use context for timeout and cancellation
// - Demonstrate best practices for production usage
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/severity1/claude-code-sdk-go"
)

func main() {
	fmt.Println("Claude Code SDK for Go - Advanced Client Features Example")
	fmt.Println("=========================================================")

	// Demonstrate advanced client configuration
	fmt.Println("🔧 Creating client with advanced options...")
	
	client := claudecode.NewClient(
		claudecode.WithSystemPrompt("You are a senior Go developer and technical mentor. Provide detailed, practical advice with code examples when appropriate. Keep responses focused and actionable."),
		// claudecode.WithModel("claude-opus-4"), // Uncomment if you have access to specific models
		// claudecode.WithAllowedTools("read_file", "write_file"), // Uncomment to restrict tools
	)

	// Demonstrate robust connection handling with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	fmt.Println("📡 Connecting with error handling and retries...")
	
	// Implement connection with retry logic
	maxRetries := 3
	var connectionErr error
	
	for attempt := 1; attempt <= maxRetries; attempt++ {
		fmt.Printf("   Attempt %d/%d...\n", attempt, maxRetries)
		
		connectionErr = client.Connect(ctx)
		if connectionErr == nil {
			fmt.Println("✅ Connected successfully!")
			break
		}
		
		// Handle specific error types
		var cliError *claudecode.CLINotFoundError
		if errors.As(connectionErr, &cliError) {
			fmt.Printf("❌ Claude CLI not found: %v\n", cliError)
			fmt.Println("💡 Please install: npm install -g @anthropic-ai/claude-code")
			return
		}
		
		var connError *claudecode.ConnectionError
		if errors.As(connectionErr, &connError) {
			fmt.Printf("⚠️  Connection failed (attempt %d): %v\n", attempt, connError)
			if attempt < maxRetries {
				fmt.Println("   Retrying in 2 seconds...")
				time.Sleep(2 * time.Second)
				continue
			}
		}
		
		// Unknown error - don't retry
		fmt.Printf("❌ Unknown connection error: %v\n", connectionErr)
		return
	}
	
	if connectionErr != nil {
		log.Fatalf("Failed to connect after %d attempts: %v", maxRetries, connectionErr)
	}

	// Ensure proper cleanup with error handling
	defer func() {
		fmt.Println("\n🧹 Cleaning up connection...")
		if err := client.Disconnect(); err != nil {
			log.Printf("⚠️  Cleanup warning: %v", err)
		} else {
			fmt.Println("✅ Disconnected cleanly")
		}
	}()

	// Demonstrate advanced interaction patterns
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🚀 Advanced Usage Patterns")
	fmt.Println(strings.Repeat("=", 60))

	// Pattern 1: Technical question with context
	technicalQuestion := `I'm building a concurrent web crawler in Go. I'm concerned about managing goroutines and avoiding race conditions. What patterns should I follow for:
1. Limiting concurrent requests
2. Sharing state safely between goroutines
3. Graceful shutdown`

	if err := runAdvancedQuery(ctx, client, "Technical Architecture Question", technicalQuestion); err != nil {
		log.Printf("Query 1 failed: %v", err)
	}

	// Brief pause between queries
	time.Sleep(2 * time.Second)

	// Pattern 2: Code review request
	codeReview := `Can you review this Go code for potential issues?

func processItems(items []Item) error {
    var wg sync.WaitGroup
    errCh := make(chan error, len(items))
    
    for _, item := range items {
        wg.Add(1)
        go func(item Item) {
            defer wg.Done()
            if err := processItem(item); err != nil {
                errCh <- err
            }
        }(item)
    }
    
    wg.Wait()
    close(errCh)
    
    for err := range errCh {
        if err != nil {
            return err
        }
    }
    return nil
}`

	if err := runAdvancedQuery(ctx, client, "Code Review", codeReview); err != nil {
		log.Printf("Query 2 failed: %v", err)
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎉 Advanced features demonstration completed!")
	fmt.Println("\n✨ Features demonstrated:")
	fmt.Println("   • Client configuration with system prompts")
	fmt.Println("   • Robust error handling with specific error types")
	fmt.Println("   • Connection retry logic")
	fmt.Println("   • Context-based timeout management")
	fmt.Println("   • Production-ready patterns")
	fmt.Println("   • Complex multi-part queries")
	fmt.Println("   • Graceful resource cleanup")
}

// runAdvancedQuery demonstrates a reusable pattern for handling queries with full error handling
func runAdvancedQuery(ctx context.Context, client claudecode.Client, title, question string) error {
	fmt.Printf("\n📋 %s\n", title)
	fmt.Printf("❓ %s\n", strings.TrimSpace(question))
	fmt.Println(strings.Repeat("-", 50))

	if err := client.Query(ctx, question); err != nil {
		return fmt.Errorf("failed to send query: %w", err)
	}

	fmt.Println("🤖 Response:")
	responseReceived := false
	
	msgChan := client.ReceiveMessages(ctx)
	for {
		select {
		case message := <-msgChan:
			if message == nil {
				goto queryComplete // Stream ended
			}

			switch msg := message.(type) {
			case *claudecode.AssistantMessage:
				responseReceived = true
				for _, block := range msg.Content {
					switch b := block.(type) {
					case *claudecode.TextBlock:
						fmt.Print(b.Text)
					case *claudecode.ThinkingBlock:
						fmt.Printf("\n💭 [Analysis: %s]\n", b.Thinking)
					}
				}
			case *claudecode.SystemMessage:
				// System messages (can be safely ignored in most cases)
			case *claudecode.ResultMessage:
				if msg.IsError {
					return fmt.Errorf("claude returned error: %s", msg.Result)
				}
				goto queryComplete
			default:
				fmt.Printf("\n📦 Unexpected message type: %T\n", message)
			}
		case <-ctx.Done():
			return fmt.Errorf("query timed out: %w", ctx.Err())
		}
	}

queryComplete:

	if !responseReceived {
		return fmt.Errorf("no response received")
	}

	fmt.Println("\n✅ Query completed successfully")
	return nil
}