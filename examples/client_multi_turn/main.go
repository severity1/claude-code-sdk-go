// Package main demonstrates multi-turn conversation using the Claude Code SDK Client API.
//
// This example shows how to:
// - Maintain context across multiple questions and responses
// - Build on previous conversation history
// - Handle streaming responses in a conversational flow
// - Demonstrate the power of Client API vs Query API for interactive scenarios
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
	fmt.Println("Claude Code SDK for Go - Multi-Turn Conversation Example")
	fmt.Println("========================================================")

	// Longer timeout for multi-turn conversation
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Create and connect client
	fmt.Println("🔗 Setting up streaming client for conversation...")
	client := claudecode.NewClient()

	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect client: %v", err)
	}
	
	defer func() {
		fmt.Println("\n🧹 Ending conversation...")
		if err := client.Disconnect(); err != nil {
			log.Printf("Warning: Failed to disconnect cleanly: %v", err)
		}
		fmt.Println("👋 Goodbye!")
	}()

	fmt.Println("✅ Connected! Starting multi-turn conversation...")

	// Conversation flow demonstrating context preservation
	conversation := []struct {
		turn     int
		question string
		context  string
	}{
		{
			turn:     1,
			question: "What is a binary search tree?",
			context:  "Initial question about data structures",
		},
		{
			turn:     2,
			question: "Can you show me a Go implementation of inserting a node?",
			context:  "Follow-up asking for code (builds on previous answer)",
		},
		{
			turn:     3,
			question: "What would be the time complexity of that insertion?",
			context:  "Analysis question (references the code from turn 2)",
		},
		{
			turn:     4,
			question: "How would I implement a search function for the same tree?",
			context:  "Extension question (builds on the entire conversation)",
		},
	}

	// Execute the conversation
	for _, turn := range conversation {
		fmt.Printf("\n%s\n", strings.Repeat("=", 60))
		fmt.Printf("🗣️  Turn %d: %s\n", turn.turn, turn.context)
		fmt.Printf("❓ Question: %s\n", turn.question)
		fmt.Println(strings.Repeat("-", 60))

		// Send question
		if err := client.Query(ctx, turn.question); err != nil {
			log.Fatalf("Failed to send question for turn %d: %v", turn.turn, err)
		}

		// Process streaming response
		fmt.Printf("🤖 Claude's Response:\n\n")
		responseReceived := false
		
		msgChan := client.ReceiveMessages(ctx)
		for {
			select {
			case message := <-msgChan:
				if message == nil {
					goto turnComplete
				}

				switch msg := message.(type) {
				case *claudecode.AssistantMessage:
					responseReceived = true
					for _, block := range msg.Content {
						switch b := block.(type) {
						case *claudecode.TextBlock:
							fmt.Print(b.Text)
						case *claudecode.ThinkingBlock:
							fmt.Printf("\n💭 [Claude is thinking: %s]\n", b.Thinking)
						}
					}
				case *claudecode.ResultMessage:
					if msg.IsError {
						fmt.Printf("\n❌ Error: %s\n", msg.Result)
					}
					goto turnComplete
				}
			case <-time.After(30 * time.Second):
				fmt.Printf("\n⏰ Turn %d timed out\n", turn.turn)
				goto turnComplete
			}
		}

	turnComplete:

		if !responseReceived {
			fmt.Printf("⚠️  No response received for turn %d\n", turn.turn)
		}

		// Brief pause between turns for readability
		if turn.turn < len(conversation) {
			fmt.Printf("\n\n⏳ Preparing next question...\n")
			time.Sleep(1 * time.Second)
		}
	}

	// Demonstration summary
	fmt.Printf("\n%s\n", strings.Repeat("=", 60))
	fmt.Println("🎉 Multi-turn conversation completed!")
	fmt.Println("\n✨ What this example demonstrated:")
	fmt.Println("   • Context preservation across multiple questions")
	fmt.Println("   • Building on previous responses (BST → Go code → complexity → search)")
	fmt.Println("   • Streaming responses in conversational flow")
	fmt.Println("   • Client API maintaining session state automatically")
	fmt.Println("\n💡 Key advantage over Query API:")
	fmt.Println("   • Query API would require repeating full context each time")
	fmt.Println("   • Client API maintains conversation history automatically")
	fmt.Println("   • Perfect for interactive applications and complex workflows")
}