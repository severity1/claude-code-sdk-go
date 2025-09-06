// Package main demonstrates advanced features of the Claude Code SDK Client API.
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

	fmt.Println("🔧 Creating client with advanced options...")
	
	client := claudecode.NewClient(
		claudecode.WithSystemPrompt("You are a senior Go developer and technical mentor. Provide detailed, practical advice with code examples when appropriate. Keep responses focused and actionable."),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	fmt.Println("📡 Connecting with error handling and retries...")
	
	maxRetries := 3
	var connectionErr error
	
	for attempt := 1; attempt <= maxRetries; attempt++ {
		fmt.Printf("   Attempt %d/%d...\n", attempt, maxRetries)
		
		connectionErr = client.Connect(ctx)
		if connectionErr == nil {
			fmt.Println("✅ Connected successfully!")
			break
		}
		
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
		
		fmt.Printf("❌ Unknown connection error: %v\n", connectionErr)
		return
	}
	
	if connectionErr != nil {
		log.Fatalf("Failed to connect after %d attempts: %v", maxRetries, connectionErr)
	}

	defer func() {
		fmt.Println("\n🧹 Cleaning up connection...")
		if err := client.Disconnect(); err != nil {
			log.Printf("⚠️  Cleanup warning: %v", err)
		} else {
			fmt.Println("✅ Disconnected cleanly")
		}
	}()

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🚀 Advanced Usage Patterns")
	fmt.Println(strings.Repeat("=", 60))

	technicalQuestion := `I'm building a concurrent web crawler in Go. I'm concerned about managing goroutines and avoiding race conditions. What patterns should I follow for:
1. Limiting concurrent requests
2. Sharing state safely between goroutines
3. Graceful shutdown`

	if err := runAdvancedQuery(ctx, client, "Technical Architecture Question", technicalQuestion); err != nil {
		log.Printf("Query 1 failed: %v", err)
	}

	time.Sleep(2 * time.Second)

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
}

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
				goto queryComplete
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
			case *claudecode.UserMessage:
				if blocks, ok := msg.Content.([]claudecode.ContentBlock); ok {
					for _, block := range blocks {
						if textBlock, ok := block.(*claudecode.TextBlock); ok {
							fmt.Printf("📤 User: %s\n", textBlock.Text)
						}
					}
				} else if contentStr, ok := msg.Content.(string); ok {
					fmt.Printf("📤 User: %s\n", contentStr)
				}
			case *claudecode.SystemMessage:
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