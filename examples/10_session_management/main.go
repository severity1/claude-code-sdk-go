// Package main demonstrates session management with the Claude Agent SDK for Go.
//
// This example illustrates two distinct session concepts:
//
//  1. Context Session ID - Used by QueryWithSession() to organize separate conversation
//     contexts WITHIN a single client connection. This is passed to the CLI via
//     StreamMessage.session_id and enables context isolation (e.g., "math" context
//     remembers 3+3, "default" context remembers 2+2). Not visible in Claude Code UI.
//
//  2. CLI Session UUID - The persistent conversation identifier returned in
//     ResultMessage.SessionID. This appears in Claude Code UI and is used with
//     WithResume() to continue conversations ACROSS client connections.
//
// The Python SDK follows the same pattern with ClaudeSDKClient.query(session_id="default").
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/severity1/claude-agent-sdk-go"
)

func main() {
	fmt.Println("Claude Agent SDK - Session Management Example")
	fmt.Println("==============================================")

	ctx := context.Background()

	if err := runExample(ctx); err != nil {
		log.Fatalf("Example failed: %v", err)
	}
}

func runExample(ctx context.Context) error {
	// Part 1: Context isolation within a single client connection
	// All queries here share the same CLI Session UUID but use different Context Session IDs.
	var cliSessionUUID string

	err := claudecode.WithClient(ctx, func(client claudecode.Client) error {
		// 1. Default context (uses Context Session ID = "default")
		fmt.Println("\n1. Default context with Query()")
		fmt.Println("   Asking: What's 2+2?")
		if err := client.Query(ctx, "Hello! What's 2+2? Reply briefly."); err != nil {
			return fmt.Errorf("default context query: %w", err)
		}
		result, err := streamResponse(ctx, client)
		if err != nil {
			return err
		}
		if result != nil {
			fmt.Printf("   [CLI Session UUID: %s]\n", result.SessionID)
		}

		// 2. Custom context (uses Context Session ID = "math-session")
		fmt.Println("\n2. Custom context with QueryWithSession()")
		fmt.Println("   Asking: What's 3+3?")
		if err := client.QueryWithSession(ctx, "Hello! What's 3+3? Reply briefly.", "math-session"); err != nil {
			return fmt.Errorf("math context query: %w", err)
		}
		mathResult, err := streamResponse(ctx, client)
		if err != nil {
			return err
		}
		if mathResult != nil {
			cliSessionUUID = mathResult.SessionID
			fmt.Printf("   [CLI Session UUID: %s]\n", cliSessionUUID)
		}

		// 3. Context isolation demonstration
		// Both contexts share the same CLI Session UUID but maintain separate conversation history.
		fmt.Println("\n3. Context isolation demonstration")

		fmt.Println("   Default context asking about previous question:")
		if err := client.Query(ctx, "What was my previous math question? Reply briefly."); err != nil {
			return fmt.Errorf("default context isolation test: %w", err)
		}
		if _, err := streamResponse(ctx, client); err != nil {
			return err
		}

		fmt.Println("\n   Math context remembers its own history:")
		if err := client.QueryWithSession(ctx, "What was my previous math question? Reply briefly.", "math-session"); err != nil {
			return fmt.Errorf("math context isolation test: %w", err)
		}
		if _, err := streamResponse(ctx, client); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Part 2: Session Resumption with WithResume()
	// This demonstrates using the CLI Session UUID to continue a conversation
	// in a NEW client connection. This is different from context isolation above.
	if cliSessionUUID != "" {
		fmt.Println("\n4. Session Resumption with WithResume()")
		fmt.Printf("   Using CLI Session UUID: %s\n", cliSessionUUID)

		err = claudecode.WithClient(ctx, func(client claudecode.Client) error {
			fmt.Println("   Asking resumed session about previous context:")
			// This new client connection has access to the full conversation history
			// from the previous connection because we're using the CLI Session UUID.
			if err := client.Query(ctx, "What math problem did we discuss earlier? Reply briefly."); err != nil {
				return fmt.Errorf("resumed session query: %w", err)
			}
			if _, err := streamResponse(ctx, client); err != nil {
				return err
			}
			return nil
		}, claudecode.WithResume(cliSessionUUID))
		if err != nil {
			return err
		}
	}

	fmt.Println("\nSession management demonstration completed!")
	return nil
}

// streamResponse streams a complete response from the client and returns the ResultMessage.
// This follows established SDK patterns for proper streaming output without messy duplicates.
func streamResponse(ctx context.Context, client claudecode.Client) (*claudecode.ResultMessage, error) {
	msgChan := client.ReceiveMessages(ctx)
	for {
		select {
		case message := <-msgChan:
			if message == nil {
				return nil, nil
			}

			switch msg := message.(type) {
			case *claudecode.AssistantMessage:
				for _, block := range msg.Content {
					if textBlock, ok := block.(*claudecode.TextBlock); ok {
						fmt.Print(textBlock.Text)
					}
				}
			case *claudecode.ResultMessage:
				fmt.Println() // Add newline after complete response
				if msg.IsError {
					if msg.Result != nil {
						return nil, fmt.Errorf("error: %s", *msg.Result)
					}
					return nil, fmt.Errorf("error: unknown error")
				}
				return msg, nil
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}
