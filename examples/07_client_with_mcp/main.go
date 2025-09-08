// Package main demonstrates Client API with AWS MCP tools using WithClient for security analysis.
package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/severity1/claude-code-sdk-go"
)

func main() {
	fmt.Println("Claude Code SDK - Client API with AWS MCP Tools Example")
	fmt.Println("S3 security analysis with context preservation")

	ctx := context.Background()

	// Two-step AWS security workflow using WithClient
	steps := []string{
		"List all S3 buckets with their names, creation dates, and regions",
		"Analyze the S3 buckets from step 1 for public access and security risks. Provide recommendations.",
	}

	// WithClient maintains context between AWS operations automatically
	err := claudecode.WithClient(ctx, func(client claudecode.Client) error {
		fmt.Println("\nConnected! Starting S3 security analysis workflow...")

		for i, step := range steps {
			fmt.Printf("\n--- Step %d ---\n", i+1)
			fmt.Printf("Query: %s\n", step)

			if err := client.Query(ctx, step); err != nil {
				return fmt.Errorf("step %d failed: %w", i+1, err)
			}

			if err := streamAWSResponse(ctx, client); err != nil {
				return fmt.Errorf("step %d response failed: %w", i+1, err)
			}
		}

		fmt.Println("\nS3 security analysis completed!")
		fmt.Println("Context was preserved - step 2 referenced buckets from step 1")
		return nil
	}, claudecode.WithAllowedTools(
		"mcp__aws-api-mcp__call_aws",
		"mcp__aws-api-mcp__suggest_aws_commands"),
		claudecode.WithSystemPrompt("You are an AWS security expert. Analyze S3 buckets for security risks."))
	if err != nil {
		log.Fatalf("S3 analysis failed: %v", err)
	}
}

func streamAWSResponse(ctx context.Context, client claudecode.Client) error {
	fmt.Println("\nResponse:")

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
						fmt.Print(textBlock.Text)
					}
				}
			case *claudecode.UserMessage:
				if blocks, ok := msg.Content.([]claudecode.ContentBlock); ok {
					for _, block := range blocks {
						if toolResult, ok := block.(*claudecode.ToolResultBlock); ok {
							if content, ok := toolResult.Content.(string); ok {
								if strings.Contains(content, "tool_use_error") {
									fmt.Printf("⚠️ AWS Error: %s\n", content)
								} else if len(content) > 150 {
									fmt.Printf("🔧 AWS: %s...\n", content[:150])
								} else {
									fmt.Printf("🔧 AWS: %s\n", content)
								}
							}
						}
					}
				}
			case *claudecode.ResultMessage:
				if msg.IsError {
					return fmt.Errorf("error: %s", msg.Result)
				}
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
