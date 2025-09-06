// Package main demonstrates basic streaming usage of the Claude Code SDK Client API.
//
// This example shows how to:
// - Create and connect a Claude Code streaming client
// - Send a query and process streaming responses in real-time
// - Handle different message types and content blocks
// - Properly manage client connection lifecycle
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
	fmt.Println("Claude Code SDK for Go - Client Streaming Example")
	fmt.Println("================================================")

	// Create context with reasonable timeout for streaming
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Create a new streaming client
	fmt.Println("🔗 Creating streaming client...")
	client := claudecode.NewClient()

	// Connect to Claude Code CLI
	fmt.Println("📡 Connecting to Claude Code...")
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect client: %v", err)
	}
	
	// Always ensure proper cleanup
	defer func() {
		fmt.Println("\n🧹 Cleaning up connection...")
		if err := client.Disconnect(); err != nil {
			log.Printf("Warning: Failed to disconnect cleanly: %v", err)
		}
		fmt.Println("✅ Connection closed")
	}()

	fmt.Println("✅ Connected successfully!")

	// Send a query to demonstrate streaming
	question := "Explain what Go goroutines are and show a simple example"
	fmt.Printf("\n🤖 Asking Claude: %s\n", question)
	
	if err := client.Query(ctx, question); err != nil {
		log.Fatalf("Failed to send query: %v", err)
	}

	fmt.Println("\n📥 Streaming response:")
	fmt.Println(strings.Repeat("-", 50))

	// Process streaming messages as they arrive
	responseReceived := false
	msgChan := client.ReceiveMessages(ctx)
	for {
		select {
		case message := <-msgChan:
			if message == nil {
				fmt.Println("\n📡 Stream ended")
				goto streamDone
			}

			// Handle different message types
			switch msg := message.(type) {
			case *claudecode.AssistantMessage:
				responseReceived = true
				// Process content blocks in real-time
				for _, block := range msg.Content {
					switch b := block.(type) {
					case *claudecode.TextBlock:
						// Stream text content as it arrives
						fmt.Print(b.Text)
					case *claudecode.ThinkingBlock:
						// Show Claude's thinking process
						fmt.Printf("\n💭 [Thinking: %s]\n", b.Thinking)
					}
				}
			case *claudecode.SystemMessage:
				// System message (usually initialization)
				fmt.Println("⚙️  System initialized")
			case *claudecode.ResultMessage:
				// Final result message
				if msg.IsError {
					fmt.Printf("\n❌ Error: %s\n", msg.Result)
				} else {
					fmt.Printf("\n✅ Stream completed successfully\n")
				}
			default:
				fmt.Printf("\n📦 Received message type: %T\n", message)
			}
		case <-ctx.Done():
			fmt.Println("\n⏰ Context timeout")
			goto streamDone
		}
	}

streamDone:

	if !responseReceived {
		fmt.Println("⚠️  No response received - check your Claude Code installation")
	} else {
		fmt.Println(strings.Repeat("-", 50))
		fmt.Println("\n🎉 Streaming example completed!")
		fmt.Println("\n💡 Key features demonstrated:")
		fmt.Println("   • Real-time streaming responses")
		fmt.Println("   • Proper connection management")
		fmt.Println("   • Message type handling")
		fmt.Println("   • Resource cleanup")
	}
}