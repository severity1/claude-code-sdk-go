// Package main demonstrates the differences between Query API and Client API.
//
// This example shows:
// - When to use Query API vs Client API
// - Performance and resource implications
// - Use case scenarios for each approach
// - Side-by-side comparison of the same task using both APIs
package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/severity1/claude-code-sdk-go"
)

func main() {
	fmt.Println("Claude Code SDK for Go - Query vs Client API Comparison")
	fmt.Println("=======================================================")

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Demonstrate the same task using both APIs
	question := "What are the key differences between channels and mutexes in Go concurrency?"

	fmt.Println("📊 Comparing Query API vs Client API for the same task:")
	fmt.Printf("❓ Question: %s\n\n", question)

	// Part 1: Using Query API (one-shot)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("🎯 QUERY API - One-shot approach")
	fmt.Println(strings.Repeat("=", 60))
	
	startTime := time.Now()
	if err := demonstrateQueryAPI(ctx, question); err != nil {
		log.Printf("Query API demo failed: %v", err)
	}
	queryDuration := time.Since(startTime)

	fmt.Printf("\n⏱️  Query API completed in: %v\n", queryDuration)
	
	// Brief pause between demonstrations
	fmt.Println("\n⏳ Switching to Client API...")
	time.Sleep(2 * time.Second)

	// Part 2: Using Client API (streaming)
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🔄 CLIENT API - Streaming approach")
	fmt.Println(strings.Repeat("=", 60))
	
	startTime = time.Now()
	if err := demonstrateClientAPI(ctx, question); err != nil {
		log.Printf("Client API demo failed: %v", err)
	}
	clientDuration := time.Since(startTime)

	fmt.Printf("\n⏱️  Client API completed in: %v\n", clientDuration)

	// Comprehensive comparison
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📈 COMPREHENSIVE COMPARISON")
	fmt.Println(strings.Repeat("=", 60))

	fmt.Printf("\n⏱️  Performance Comparison:\n")
	fmt.Printf("   Query API: %v\n", queryDuration)
	fmt.Printf("   Client API: %v\n", clientDuration)
	if queryDuration < clientDuration {
		fmt.Printf("   → Query API was faster by %v\n", clientDuration-queryDuration)
	} else {
		fmt.Printf("   → Client API was faster by %v\n", queryDuration-clientDuration)
	}

	fmt.Println("\n🎯 When to use Query API:")
	fmt.Println("   ✅ One-shot questions or commands")
	fmt.Println("   ✅ Batch processing")
	fmt.Println("   ✅ CI/CD scripts")
	fmt.Println("   ✅ Simple automation tasks")
	fmt.Println("   ✅ Lower resource overhead")
	fmt.Println("   ✅ Fire-and-forget operations")

	fmt.Println("\n🔄 When to use Client API:")
	fmt.Println("   ✅ Multi-turn conversations")
	fmt.Println("   ✅ Interactive applications")
	fmt.Println("   ✅ Context-dependent workflows")
	fmt.Println("   ✅ Real-time streaming needs")
	fmt.Println("   ✅ Complex state management")
	fmt.Println("   ✅ Long-running sessions")

	fmt.Println("\n💡 Resource Considerations:")
	fmt.Println("   Query API: Creates new process per query")
	fmt.Println("   Client API: Maintains persistent connection")
	fmt.Println("   → Use Query for occasional calls")
	fmt.Println("   → Use Client for frequent interactions")

	fmt.Println("\n🏗️  Architecture Patterns:")
	fmt.Println("   Query API: Stateless, functional approach")
	fmt.Println("   Client API: Stateful, object-oriented approach")

	// Demonstrate multi-turn scenario where Client API shines
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🌟 CLIENT API ADVANTAGE - Multi-turn scenario")
	fmt.Println(strings.Repeat("=", 60))
	
	if err := demonstrateClientAdvantage(ctx); err != nil {
		log.Printf("Client advantage demo failed: %v", err)
	}

	fmt.Println("\n🎉 Comparison complete!")
	fmt.Println("\n🧠 Key Takeaways:")
	fmt.Println("   • Both APIs are production-ready with 100% Python SDK compatibility")
	fmt.Println("   • Choose based on your use case, not just preference")
	fmt.Println("   • Query API: Simple, stateless, efficient for one-shots")
	fmt.Println("   • Client API: Powerful, stateful, ideal for conversations")
	fmt.Println("   • Consider resource usage patterns in your application")
}

func demonstrateQueryAPI(ctx context.Context, question string) error {
	fmt.Println("🚀 Starting Query API...")
	fmt.Println("   • Creates new Claude process")
	fmt.Println("   • Sends question")
	fmt.Println("   • Receives complete response")
	fmt.Println("   • Automatically cleans up")
	fmt.Println("\n📤 Sending query...")

	iterator, err := claudecode.Query(ctx, question)
	if err != nil {
		return fmt.Errorf("failed to create query: %w", err)
	}
	defer iterator.Close()

	fmt.Println("📥 Response:")
	fmt.Println(strings.Repeat("-", 40))

	for {
		message, err := iterator.Next(ctx)
		if err != nil {
			if err.Error() == "no more messages" {
				break
			}
			return fmt.Errorf("failed to get next message: %w", err)
		}

		if message == nil {
			break
		}

		if assistantMsg, ok := message.(*claudecode.AssistantMessage); ok {
			for _, block := range assistantMsg.Content {
				if textBlock, ok := block.(*claudecode.TextBlock); ok {
					fmt.Print(textBlock.Text)
				}
			}
		}
	}

	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("✅ Query API completed - process automatically cleaned up")
	return nil
}

func demonstrateClientAPI(ctx context.Context, question string) error {
	fmt.Println("🚀 Starting Client API...")
	fmt.Println("   • Creates persistent connection")
	fmt.Println("   • Maintains session state")
	fmt.Println("   • Streams responses in real-time")
	fmt.Println("   • Keeps connection for future queries")

	client := claudecode.NewClient()
	
	fmt.Println("\n🔗 Connecting...")
	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	
	defer func() {
		fmt.Println("🧹 Manually cleaning up connection...")
		if err := client.Disconnect(); err != nil {
			log.Printf("Warning: %v", err)
		}
	}()

	fmt.Println("📤 Sending query...")
	if err := client.Query(ctx, question); err != nil {
		return fmt.Errorf("failed to send query: %w", err)
	}

	fmt.Println("📥 Streaming response:")
	fmt.Println(strings.Repeat("-", 40))

	msgChan := client.ReceiveMessages(ctx)
	for {
		select {
		case message := <-msgChan:
			if message == nil {
				goto clientDone
			}

			if assistantMsg, ok := message.(*claudecode.AssistantMessage); ok {
				for _, block := range assistantMsg.Content {
					if textBlock, ok := block.(*claudecode.TextBlock); ok {
						fmt.Print(textBlock.Text)
					}
				}
			}
		case <-ctx.Done():
			goto clientDone
		}
	}

clientDone:

	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("✅ Client API response completed - connection still available for more queries")
	return nil
}

func demonstrateClientAdvantage(ctx context.Context) error {
	fmt.Println("🎭 Demonstrating why Client API is superior for conversations...")
	fmt.Println("Scenario: Building on previous answers (context preservation)")

	client := claudecode.NewClient()
	if err := client.Connect(ctx); err != nil {
		return err
	}
	defer client.Disconnect()

	// Multi-turn conversation that builds context
	turns := []string{
		"What is dependency injection in software development?",
		"Show me how to implement it in Go using interfaces",
		"What are the testing advantages of the approach you just showed?",
	}

	for i, question := range turns {
		fmt.Printf("\n🗣️  Turn %d: %s\n", i+1, question)
		fmt.Println(strings.Repeat("-", 30))
		
		if err := client.Query(ctx, question); err != nil {
			return err
		}

		// Brief response processing (truncated for demo)
		responseLines := 0
		msgChan := client.ReceiveMessages(ctx)
		for responseLines < 5 { // Just show first few lines
			select {
			case message := <-msgChan:
				if message == nil {
					goto turnDone
				}
				if assistantMsg, ok := message.(*claudecode.AssistantMessage); ok {
					for _, block := range assistantMsg.Content {
						if textBlock, ok := block.(*claudecode.TextBlock); ok {
							lines := strings.Split(textBlock.Text, "\n")
							for _, line := range lines {
								if responseLines < 5 && strings.TrimSpace(line) != "" {
									fmt.Printf("   %s\n", line)
									responseLines++
								}
							}
						}
					}
				}
			case <-time.After(10 * time.Second):
				goto turnDone
			}
		}
		if responseLines >= 5 {
			fmt.Println("   [... response continues with full context from previous turns ...]")
		}
		
	turnDone:
		// Drain remaining messages
		for {
			select {
			case message := <-msgChan:
				if message == nil {
					goto drained
				}
			case <-time.After(1 * time.Second):
				goto drained
			}
		}
	drained:
	}

	fmt.Println("\n💡 Notice: Each follow-up question built on the previous context!")
	fmt.Println("   • Turn 2 referenced 'it' (dependency injection)")
	fmt.Println("   • Turn 3 referenced 'the approach you just showed'")
	fmt.Println("   • Client API maintained full conversation history")
	fmt.Println("   • Query API would require repeating full context each time")

	return nil
}