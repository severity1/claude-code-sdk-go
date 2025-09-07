// Package main demonstrates using Claude Code SDK Client API with MCP tools (AWS API).
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/severity1/claude-code-sdk-go"
)

func main() {
	fmt.Println("Claude Code SDK for Go - Client API with MCP Tools (S3 Security Analysis) Example")
	fmt.Println("==========================================================================")

	awsAvailable := checkAWSCredentials()
	if !awsAvailable {
		fmt.Println("⚠️  AWS credentials not found.")
		fmt.Println("This example will demonstrate S3 security analysis patterns with explanatory responses.")
		fmt.Println("For full functionality, configure AWS credentials.")
	} else {
		fmt.Println("✅ AWS credentials detected - full S3 security analysis available!")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	fmt.Println("\n🔗 Creating client with AWS MCP tools for S3 security analysis...")
	client := claudecode.NewClient(
		claudecode.WithAllowedTools(
			"mcp__aws-api-mcp__call_aws",
			"mcp__aws-api-mcp__suggest_aws_commands",
			"Read", "Write", "Edit", "TodoWrite"),
		claudecode.WithSystemPrompt("You are an expert AWS security specialist focusing on S3 storage security. You can discover S3 buckets, analyze their public access configurations, and identify security recommendations."),
	)

	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect client: %v", err)
	}

	defer func() {
		fmt.Println("\n🧹 Ending S3 security analysis session...")
		if err := client.Disconnect(); err != nil {
			log.Printf("Warning: Failed to disconnect cleanly: %v", err)
		}
		fmt.Println("👋 S3 security analysis session ended!")
	}()

	fmt.Println("✅ Connected! Starting interactive S3 security analysis session...")

	workflow := []struct {
		step        int
		title       string
		description string
		query       string
	}{
		{
			step:        1,
			title:       "S3 Bucket Discovery",
			description: "List all S3 buckets in the AWS account",
			query:       `Please list all S3 buckets in my AWS account. Show me the bucket names, creation dates, and regions where they are located.`,
		},
		{
			step:        2,
			title:       "S3 Public Access Analysis",
			description: "Check if any S3 buckets are publicly accessible",
			query:       `Now, please analyze the S3 buckets we just discovered and check if any of them are publicly accessible. Look for buckets with public read or write access through bucket policies, ACLs, or block public access settings. Please identify any security risks and provide recommendations.`,
		},
	}

	for _, step := range workflow {
		fmt.Printf("\n%s\n", strings.Repeat("=", 80))
		fmt.Printf("🔧 Step %d: %s\n", step.step, step.title)
		fmt.Printf("📋 %s\n", step.description)
		fmt.Println(strings.Repeat("-", 80))

		if err := executeAWSStep(ctx, client, step.step, step.query); err != nil {
			log.Printf("Step %d failed: %v", step.step, err)

			fmt.Printf("\n⚠️  Step %d encountered an error. Continuing to next step...\n", step.step)
			continue
		}

		if step.step < len(workflow) {
			fmt.Printf("\n⏳ Preparing for next step...\n")
			time.Sleep(2 * time.Second)
		}
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
	fmt.Println("🎉 S3 Security Analysis Session Completed!")

	fmt.Println("\n📊 Workflow Summary:")
	for _, step := range workflow {
		fmt.Printf("   ✅ Step %d: %s\n", step.step, step.title)
	}

	fmt.Println("\n✨ What was demonstrated:")
	fmt.Println("   • S3 bucket discovery using AWS MCP tools")
	fmt.Println("   • Context preservation across S3 analysis operations")
	fmt.Println("   • Progressive S3 security analysis building on previous results")
	fmt.Println("   • Two-step AWS security workflow with Client API")

	fmt.Println("\n💡 Client API Advantages for AWS Security Analysis:")
	fmt.Println("   • Maintains context between bucket discovery and security analysis")
	fmt.Println("   • Can reference bucket list from step 1 when analyzing access in step 2")
	fmt.Println("   • Perfect for interactive security assessment workflows")
	fmt.Println("   • Enables progressive S3 security analysis")

	if !awsAvailable {
		fmt.Println("\n🔧 To enable full AWS integration:")
		fmt.Println("   1. Configure AWS credentials (aws configure)")
		fmt.Println("   2. Install AWS MCP: claude mcp add aws-api")
		fmt.Println("   3. Re-run this example for live AWS operations")
	}
}

func executeAWSStep(ctx context.Context, client claudecode.Client, stepNum int, query string) error {
	fmt.Printf("❓ Query: %s\n", strings.TrimSpace(query))
	fmt.Println(strings.Repeat("-", 50))

	if err := client.Query(ctx, query); err != nil {
		return fmt.Errorf("failed to send query: %w", err)
	}

	fmt.Println("🤖 Claude's Response:")
	responseReceived := false
	msgChan := client.ReceiveMessages(ctx)

	for {
		select {
		case message := <-msgChan:
			if message == nil {
				goto stepComplete
			}

			switch msg := message.(type) {
			case *claudecode.AssistantMessage:
				responseReceived = true
				for _, block := range msg.Content {
					switch b := block.(type) {
					case *claudecode.TextBlock:
						fmt.Print(b.Text)
					case *claudecode.ThinkingBlock:
						fmt.Printf("\n💭 [AWS Analysis: %s]\n", b.Thinking)
					}
				}
			case *claudecode.UserMessage:
				if blocks, ok := msg.Content.([]claudecode.ContentBlock); ok {
					for _, block := range blocks {
						switch b := block.(type) {
						case *claudecode.TextBlock:
							fmt.Printf("📤 AWS Tool: %s\n", b.Text)
						case *claudecode.ToolResultBlock:
							if contentStr, ok := b.Content.(string); ok {
								displayContent := contentStr
								if strings.Contains(contentStr, "<tool_use_error>") {
									displayContent = strings.ReplaceAll(contentStr, "<tool_use_error>", "⚠️ ")
									displayContent = strings.ReplaceAll(displayContent, "</tool_use_error>", "")
									fmt.Printf("🔧 AWS Tool Issue (id: %s): %s\n", b.ToolUseID[:8]+"...", strings.TrimSpace(displayContent))
								} else if len(displayContent) > 150 {
									fmt.Printf("🔧 AWS Tool Result (id: %s): %s...\n", b.ToolUseID[:8]+"...", displayContent[:150])
								} else {
									fmt.Printf("🔧 AWS Tool Result (id: %s): %s\n", b.ToolUseID[:8]+"...", displayContent)
								}
							} else {
								fmt.Printf("🔧 AWS Tool Result (id: %s): <structured data>\n", b.ToolUseID[:8]+"...")
							}
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
				goto stepComplete
			default:
				fmt.Printf("\n📦 Received message type: %T\n", message)
			}
		case <-time.After(60 * time.Second):
			fmt.Printf("\n⏰ Step %d timed out, but may have completed\n", stepNum)
			goto stepComplete
		}
	}

stepComplete:
	if !responseReceived {
		return fmt.Errorf("no response received for step %d", stepNum)
	}

	fmt.Printf("\n✅ Step %d completed successfully\n", stepNum)
	return nil
}

func checkAWSCredentials() bool {
	if os.Getenv("AWS_ACCESS_KEY_ID") != "" && os.Getenv("AWS_SECRET_ACCESS_KEY") != "" {
		return true
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	credentialsPath := homeDir + "/.aws/credentials"
	if _, err := os.Stat(credentialsPath); err == nil {
		return true
	}

	return false
}
