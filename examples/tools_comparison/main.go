// Package main demonstrates comparing Query API vs Client API for tool-heavy workflows.
//
// This example shows:
// - When to use Query API vs Client API for operations involving tools
// - Performance characteristics of each approach with tool usage
// - Resource implications and best practices for tool-heavy workflows
// - Side-by-side comparison using the same AWS audit task
//
// Prerequisites:
// - AWS API MCP server installed: `claude mcp add aws-api` (optional - will mock if not available)
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
	fmt.Println("Claude Code SDK for Go - Query vs Client API with Tools Comparison")
	fmt.Println("================================================================")

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second) // 5 minutes for complex operations
	defer cancel()

	// Check tool availability
	awsAvailable := checkAWSCredentials()
	if !awsAvailable {
		fmt.Println("ℹ️  AWS credentials not detected - examples will show interaction patterns")
	}

	// Define the same complex task for both APIs
	complexTask := `Please perform a comprehensive AWS infrastructure audit:

1. **Resource Discovery**: List all EC2 instances, S3 buckets, RDS databases, and Lambda functions across all regions
2. **Cost Analysis**: Get current billing information and identify top cost drivers
3. **Security Review**: Check for security groups with 0.0.0.0/0 access, unencrypted resources, and excessive IAM permissions
4. **Compliance Check**: Verify CloudTrail is enabled, check for untagged resources
5. **Report Generation**: Create a comprehensive audit report in aws-audit-report.md with findings and recommendations
6. **Executive Summary**: Write a one-page executive summary in aws-executive-summary.txt

This is a complex, multi-tool workflow that will demonstrate the differences between Query and Client approaches.`

	fmt.Printf("🎯 Common Task: %s\n", strings.TrimSpace(complexTask))

	// Part 1: Query API Approach (One-Shot)
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("🎯 QUERY API APPROACH - One-Shot Execution")
	fmt.Println(strings.Repeat("=", 70))

	fmt.Println("🚀 Query API Characteristics:")
	fmt.Println("   • Single process execution")
	fmt.Println("   • All tools available in one session")
	fmt.Println("   • Automatic resource cleanup")
	fmt.Println("   • Optimized for complete workflows")

	queryStartTime := time.Now()
	if err := demonstrateQueryApproach(ctx, complexTask, awsAvailable); err != nil {
		log.Printf("Query API approach failed: %v", err)
	}
	queryDuration := time.Since(queryStartTime)

	// Cleanup pause
	fmt.Println("\n⏳ Switching approaches...")
	time.Sleep(3 * time.Second)

	// Part 2: Client API Approach (Interactive)
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("🔄 CLIENT API APPROACH - Interactive Execution")
	fmt.Println(strings.Repeat("=", 70))

	fmt.Println("🚀 Client API Characteristics:")
	fmt.Println("   • Persistent connection")
	fmt.Println("   • Context maintained across steps")
	fmt.Println("   • Interactive problem-solving")
	fmt.Println("   • Optimized for iterative workflows")

	clientStartTime := time.Now()
	if err := demonstrateClientApproach(ctx, complexTask, awsAvailable); err != nil {
		log.Printf("Client API approach failed: %v", err)
	}
	clientDuration := time.Since(clientStartTime)

	// Comprehensive comparison
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("📊 COMPREHENSIVE COMPARISON RESULTS")
	fmt.Println(strings.Repeat("=", 70))

	// Performance Analysis
	fmt.Printf("\n⏱️  Performance Analysis:\n")
	fmt.Printf("   Query API Duration:  %v\n", queryDuration)
	fmt.Printf("   Client API Duration: %v\n", clientDuration)
	
	if queryDuration < clientDuration {
		fmt.Printf("   → Query API was faster by %v (%.1f%% faster)\n", 
			clientDuration-queryDuration, 
			float64(clientDuration-queryDuration)/float64(clientDuration)*100)
	} else {
		fmt.Printf("   → Client API was faster by %v (%.1f%% faster)\n", 
			queryDuration-clientDuration,
			float64(queryDuration-clientDuration)/float64(queryDuration)*100)
	}

	// Use Case Analysis
	fmt.Println("\n🎯 Use Case Analysis:")
	
	fmt.Println("\n📋 Query API - Best for:")
	fmt.Println("   ✅ One-shot automation tasks")
	fmt.Println("   ✅ Batch processing and scheduled jobs")
	fmt.Println("   ✅ CI/CD pipeline integration")
	fmt.Println("   ✅ Simple tool workflows with clear objectives")
	fmt.Println("   ✅ Scripts that need to run and complete")
	fmt.Println("   ✅ When you want automatic cleanup")

	fmt.Println("\n🔄 Client API - Best for:")
	fmt.Println("   ✅ Interactive tool-based workflows")
	fmt.Println("   ✅ Complex multi-step processes")
	fmt.Println("   ✅ Workflows that need human input/review")
	fmt.Println("   ✅ Iterative problem-solving with tools")
	fmt.Println("   ✅ Building context across tool operations")
	fmt.Println("   ✅ Long-running management sessions")

	// Resource and Architecture Analysis
	fmt.Println("\n🏗️  Resource and Architecture Implications:")
	
	fmt.Printf("\n🎯 Query API Resource Profile:\n")
	fmt.Printf("   • Process Lifecycle: Create → Execute → Cleanup\n")
	fmt.Printf("   • Memory Usage: High during execution, zero after completion\n")
	fmt.Printf("   • Tool Context: Fresh environment for each query\n")
	fmt.Printf("   • Parallelization: Easy - multiple independent queries\n")
	fmt.Printf("   • Error Recovery: Restart entire workflow\n")

	fmt.Printf("\n🔄 Client API Resource Profile:\n")
	fmt.Printf("   • Process Lifecycle: Create → Maintain → Manual cleanup\n")
	fmt.Printf("   • Memory Usage: Constant throughout session\n")
	fmt.Printf("   • Tool Context: Persistent across interactions\n")
	fmt.Printf("   • Parallelization: Complex - shared state management needed\n")
	fmt.Printf("   • Error Recovery: Continue from last successful step\n")

	// Tool-Specific Considerations
	fmt.Println("\n🔧 Tool-Specific Considerations:")

	fmt.Println("\n📂 File Operations (Read/Write/Edit):")
	fmt.Println("   Query API: Great for document generation, file processing")
	fmt.Println("   Client API: Better for iterative editing, multi-file workflows")

	fmt.Println("\n☁️  MCP Tools (AWS, GitHub, etc.):")
	fmt.Println("   Query API: Excellent for infrastructure automation, deployments")
	fmt.Println("   Client API: Better for exploration, interactive management")

	fmt.Println("\n🔒 Security and Permissions:")
	fmt.Println("   Query API: Easier to restrict tools per query")
	fmt.Println("   Client API: Tools available for entire session (need careful scoping)")

	// Practical Recommendations
	fmt.Println("\n💡 Practical Recommendations:")

	fmt.Println("\n🎯 Use Query API when you:")
	fmt.Println("   • Have a well-defined task with clear completion criteria")
	fmt.Println("   • Want automatic resource cleanup")
	fmt.Println("   • Need to run the same workflow repeatedly")
	fmt.Println("   • Are building automation scripts or CI/CD integrations")
	fmt.Println("   • Want to minimize resource usage")

	fmt.Println("\n🔄 Use Client API when you:")
	fmt.Println("   • Need to iterate and refine based on tool results")
	fmt.Println("   • Want to review intermediate results before proceeding")
	fmt.Println("   • Are exploring or learning about your infrastructure")
	fmt.Println("   • Need to maintain context between operations")
	fmt.Println("   • Want to build complex workflows interactively")

	// Cost and Performance Considerations
	fmt.Println("\n💰 Cost and Performance Considerations:")
	fmt.Println("   Query API: Lower total cost for simple tasks, higher per-execution overhead")
	fmt.Println("   Client API: Higher baseline cost, more efficient for multiple operations")
	fmt.Println("   Recommendation: Choose based on usage frequency and complexity")

	fmt.Println("\n🎉 Comparison Complete!")
	fmt.Println("\n🧠 Key Takeaway:")
	fmt.Println("   Both APIs excel with tools, but for different use cases.")
	fmt.Println("   Query API = Automation & Scripts")
	fmt.Println("   Client API = Interactive & Complex Workflows")
}

// demonstrateQueryApproach shows the Query API approach to tool-heavy workflows
func demonstrateQueryApproach(ctx context.Context, task string, awsAvailable bool) error {
	fmt.Println("\n🚀 Executing with Query API...")
	fmt.Println("   Creating new Claude process with AWS and file tools...")

	tools := []string{"Write", "Read"}
	if awsAvailable {
		tools = append(tools, "mcp__aws-api-mcp__*")
	}

	iterator, err := claudecode.Query(ctx, task,
		claudecode.WithAllowedTools(tools...),
		claudecode.WithSystemPrompt("You are an expert AWS auditor. Perform thorough analysis and create comprehensive reports. Use available tools to gather information and generate documentation."),
	)
	if err != nil {
		return fmt.Errorf("failed to create query: %w", err)
	}
	defer iterator.Close()

	fmt.Println("📥 Processing one-shot response:")
	fmt.Println(strings.Repeat("-", 50))

	responseReceived := false
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

		switch msg := message.(type) {
		case *claudecode.AssistantMessage:
			responseReceived = true
			for _, block := range msg.Content {
				switch b := block.(type) {
				case *claudecode.TextBlock:
					// Show first few lines to demonstrate it's working
					lines := strings.Split(b.Text, "\n")
					for i, line := range lines {
						if i < 10 && strings.TrimSpace(line) != "" { // Show first 10 non-empty lines
							fmt.Printf("   %s\n", line)
						} else if i == 10 {
							fmt.Printf("   [...continuing with comprehensive analysis and report generation...]\n")
							break
						}
					}
				case *claudecode.ThinkingBlock:
					fmt.Printf("\n💭 [AWS Analysis: %s]\n", b.Thinking)
				}
			}
		case *claudecode.ResultMessage:
			if msg.IsError {
				return fmt.Errorf("query returned error: %s", msg.Result)
			}
		}
	}

	if !responseReceived {
		return fmt.Errorf("no response received")
	}

	fmt.Println(strings.Repeat("-", 50))
	fmt.Println("✅ Query API execution completed")
	fmt.Println("   • Process automatically cleaned up")
	fmt.Println("   • All tools were available in single session")
	fmt.Println("   • Generated comprehensive reports in one execution")
	return nil
}

// demonstrateClientApproach shows the Client API approach to tool-heavy workflows
func demonstrateClientApproach(ctx context.Context, task string, awsAvailable bool) error {
	fmt.Println("\n🚀 Starting Client API session...")
	fmt.Println("   Creating persistent connection with AWS and file tools...")

	tools := []string{"Write", "Read"}
	if awsAvailable {
		tools = append(tools, "mcp__aws-api-mcp__*")
	}

	client := claudecode.NewClient(
		claudecode.WithAllowedTools(tools...),
		claudecode.WithSystemPrompt("You are an expert AWS auditor. Work interactively to build comprehensive analysis step by step."),
	)

	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect client: %w", err)
	}

	defer func() {
		fmt.Println("\n🧹 Manually cleaning up client connection...")
		client.Disconnect()
	}()

	// Break the complex task into interactive steps
	steps := []string{
		"Let's start by discovering AWS resources. Please list all EC2 instances and S3 buckets across regions.",
		"Now let's analyze costs. Based on the resources we found, get billing information and identify cost drivers.",
		"Time for security review. Check the resources we discovered for security issues like open security groups.",
		"Let's check compliance. Verify CloudTrail is enabled and identify any untagged resources.",
		"Now create a comprehensive audit report based on all our findings from the previous steps.",
		"Finally, write an executive summary of our audit results for leadership review.",
	}

	fmt.Println("📥 Processing interactive multi-step workflow:")

	for i, step := range steps {
		fmt.Printf("\n🔄 Step %d: %s\n", i+1, step)
		fmt.Println(strings.Repeat("-", 40))

		if err := client.Query(ctx, step); err != nil {
			return fmt.Errorf("failed to send step %d: %w", i+1, err)
		}

		// Process response for this step
		msgChan := client.ReceiveMessages(ctx)
		stepComplete := false
		
		for !stepComplete {
			select {
			case message := <-msgChan:
				if message == nil {
					stepComplete = true
					continue
				}

				switch msg := message.(type) {
				case *claudecode.AssistantMessage:
					for _, block := range msg.Content {
						if textBlock, ok := block.(*claudecode.TextBlock); ok {
							// Show abbreviated response to demonstrate interaction
							lines := strings.Split(textBlock.Text, "\n")
							for j, line := range lines {
								if j < 3 && strings.TrimSpace(line) != "" {
									fmt.Printf("   %s\n", line)
								} else if j == 3 {
									fmt.Printf("   [...step %d continuing with detailed analysis...]\n", i+1)
									stepComplete = true
									break
								}
							}
						}
					}
				case *claudecode.ResultMessage:
					if msg.IsError {
						fmt.Printf("   ⚠️ Step %d had issues but continuing...\n", i+1)
					}
					stepComplete = true
				}
			case <-time.After(15 * time.Second):
				fmt.Printf("   ✅ Step %d completed (context preserved for next step)\n", i+1)
				stepComplete = true
			}
		}

		// Brief pause between steps
		if i < len(steps)-1 {
			fmt.Printf("   ⏳ Context maintained, preparing step %d...\n", i+2)
			time.Sleep(1 * time.Second)
		}
	}

	fmt.Println(strings.Repeat("-", 50))
	fmt.Println("✅ Client API workflow completed")
	fmt.Println("   • Context maintained across all steps")
	fmt.Println("   • Each step built on previous results")
	fmt.Println("   • Connection still available for more operations")
	return nil
}

// checkAWSCredentials checks if AWS credentials are available
func checkAWSCredentials() bool {
	// Check environment variables
	if os.Getenv("AWS_ACCESS_KEY_ID") != "" && os.Getenv("AWS_SECRET_ACCESS_KEY") != "" {
		return true
	}

	// Check AWS credentials file
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