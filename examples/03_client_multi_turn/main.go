// Package main demonstrates multi-turn conversation using the Claude Code SDK Client API.
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

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

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

	conversation := []struct {
		turn     int
		question string
	}{
		{1, "What is a binary search tree?"},
		{2, "Can you show me a Go implementation of inserting a node?"},
		{3, "What would be the time complexity of that insertion?"},
		{4, "How would I implement a search function for the same tree?"},
	}

	for _, turn := range conversation {
		fmt.Printf("\n%s\n", strings.Repeat("=", 60))
		fmt.Printf("🗣️  Turn %d\n", turn.turn)
		fmt.Printf("❓ Question: %s\n", turn.question)
		fmt.Println(strings.Repeat("-", 60))

		if err := client.Query(ctx, turn.question); err != nil {
			log.Fatalf("Failed to send question for turn %d: %v", turn.turn, err)
		}

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
						fmt.Printf("\n❌ Issue: %s\n", msg.Result)
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

		if turn.turn < len(conversation) {
			fmt.Printf("\n\n⏳ Preparing next question...\n")
			time.Sleep(1 * time.Second)
		}
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 60))
	fmt.Println("🎉 Multi-turn conversation completed!")
	fmt.Println("\n✨ Context was preserved across all questions")
	fmt.Println("💡 Each follow-up built on previous responses automatically")
}