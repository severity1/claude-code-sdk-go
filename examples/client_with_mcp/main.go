// Package main demonstrates using Claude Code SDK Client API with MCP tools (AWS API).
//
// This example shows how to:
// - Configure Client API to use MCP tools across multiple interactions
// - Execute multi-turn conversations that build AWS context progressively
// - Handle streaming responses with AWS API results
// - Demonstrate interactive AWS management workflows
//
// Prerequisites:
// - AWS API MCP server installed: `claude mcp add aws-api`
// - AWS credentials configured (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
// - Or AWS credentials in ~/.aws/credentials
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
	fmt.Println("Claude Code SDK for Go - Client API with MCP Tools (AWS API) Example")
	fmt.Println("====================================================================")

	// Check if AWS credentials are available
	awsAvailable := checkAWSCredentials()
	if !awsAvailable {
		fmt.Println("⚠️  AWS credentials not found.")
		fmt.Println("This example will demonstrate the interaction patterns with explanatory responses.")
		fmt.Println("For full functionality, configure AWS credentials.")
	} else {
		fmt.Println("✅ AWS credentials detected - full AWS integration available!")
	}

	// Create context with extended timeout for AWS operations
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	// Create client with AWS MCP tools and file operations
	fmt.Println("\n🔗 Creating client with AWS MCP tools and file operations...")
	client := claudecode.NewClient(
		claudecode.WithAllowedTools("mcp__aws-api-mcp__*", "Read", "Write", "Edit"),
		claudecode.WithSystemPrompt("You are an expert AWS cloud consultant and DevOps engineer. You can analyze AWS infrastructure, optimize costs, improve security, and automate cloud operations. Always explain your analysis clearly and provide actionable recommendations."),
	)

	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect client: %v", err)
	}

	defer func() {
		fmt.Println("\n🧹 Ending AWS management session...")
		if err := client.Disconnect(); err != nil {
			log.Printf("Warning: Failed to disconnect cleanly: %v", err)
		}
		fmt.Println("👋 AWS management session ended!")
	}()

	fmt.Println("✅ Connected! Starting interactive AWS management session...")

	// Interactive AWS management workflow
	workflow := []struct {
		step        int
		title       string
		description string
		query       string
	}{
		{
			step:        1,
			title:       "Infrastructure Discovery",
			description: "Discover and catalog all AWS resources across regions",
			query: `Let's start with a comprehensive discovery of my AWS infrastructure. Please:

1. List all EC2 instances across all regions with their status, types, and tags
2. Show all S3 buckets with their sizes and access policies
3. List RDS instances and their configurations
4. Check Lambda functions and their recent execution stats
5. Identify any Load Balancers and their health status

Please organize this information clearly and save a detailed inventory to aws-infrastructure-inventory.json for reference in our future discussions.`,
		},
		{
			step:        2,
			title:       "Cost Analysis",
			description: "Analyze costs and identify optimization opportunities based on discovered resources",
			query: `Based on the infrastructure you just discovered, let's dive into cost analysis:

1. Get my current AWS billing information and cost breakdown by service
2. Identify which resources are consuming the most costs
3. Look for idle or underutilized resources from the inventory we just created
4. Calculate potential savings from rightsizing or terminating unused resources
5. Check for resources that could benefit from Reserved Instances or Savings Plans

Create a detailed cost optimization report in aws-cost-optimization-plan.md with specific recommendations and estimated savings.`,
		},
		{
			step:        3,
			title:       "Security Assessment",
			description: "Perform security audit building on infrastructure knowledge",
			query: `Now let's perform a thorough security assessment of the resources we've cataloged:

1. Review security groups for overly permissive rules (especially 0.0.0.0/0 access)
2. Check IAM users, roles, and policies for excessive permissions
3. Verify MFA status for all IAM users
4. Analyze S3 bucket public access settings and encryption status
5. Review VPC configurations and network ACLs
6. Check CloudTrail and CloudWatch logging status

Document all security findings in aws-security-assessment.md with risk levels and remediation steps. Reference the specific resources from our infrastructure inventory.`,
		},
		{
			step:        4,
			title:       "Automated Improvements",
			description: "Implement improvements based on previous analysis",
			query: `Based on our cost analysis and security assessment, let's implement some improvements:

1. Apply proper tags to any untagged resources we identified
2. Enable encryption on any unencrypted S3 buckets
3. Set up CloudWatch alarms for critical resources
4. Create lifecycle policies for S3 buckets to optimize storage costs
5. Implement any low-risk security recommendations from our assessment

Document all changes made in aws-improvements-log.md and update our infrastructure inventory to reflect the changes.`,
		},
		{
			step:        5,
			title:       "Ongoing Monitoring Setup",
			description: "Set up monitoring and create maintenance procedures",
			query: `Let's establish ongoing monitoring and create procedures for maintaining our optimized infrastructure:

1. Set up CloudWatch dashboards for key metrics from our resource inventory
2. Create cost budgets and alerts based on our cost analysis
3. Establish security monitoring rules for the issues we identified
4. Create automated backup policies for critical resources
5. Draft a monthly review checklist for cost and security optimization

Create an aws-monitoring-playbook.md with all monitoring setup and maintenance procedures. This will help ensure our improvements are sustained over time.`,
		},
	}

	// Execute the interactive AWS management workflow
	for _, step := range workflow {
		fmt.Printf("\n%s\n", strings.Repeat("=", 80))
		fmt.Printf("🔧 Step %d: %s\n", step.step, step.title)
		fmt.Printf("📋 %s\n", step.description)
		fmt.Println(strings.Repeat("-", 80))

		if err := executeAWSStep(ctx, client, step.step, step.query); err != nil {
			log.Printf("Step %d failed: %v", step.step, err)
			
			// Ask if user wants to continue despite the error
			fmt.Printf("\n⚠️  Step %d encountered an error. Continuing to next step...\n", step.step)
			continue
		}

		// Brief pause between steps
		if step.step < len(workflow) {
			fmt.Printf("\n⏳ Preparing for next step...\n")
			time.Sleep(2 * time.Second)
		}
	}

	// Final Summary
	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
	fmt.Println("🎉 Interactive AWS Management Session Completed!")
	
	fmt.Println("\n📊 Workflow Summary:")
	for _, step := range workflow {
		fmt.Printf("   ✅ Step %d: %s\n", step.step, step.title)
	}

	fmt.Println("\n✨ What was demonstrated:")
	fmt.Println("   • Multi-step AWS infrastructure management")
	fmt.Println("   • Context preservation across AWS operations")
	fmt.Println("   • Progressive analysis building on previous results")
	fmt.Println("   • Automated AWS improvements with change tracking")
	fmt.Println("   • Comprehensive documentation generation")

	fmt.Println("\n📂 Generated Documentation:")
	fmt.Println("   • aws-infrastructure-inventory.json - Complete resource catalog")
	fmt.Println("   • aws-cost-optimization-plan.md - Cost analysis and savings plan")
	fmt.Println("   • aws-security-assessment.md - Security findings and remediation")
	fmt.Println("   • aws-improvements-log.md - Record of all changes made")
	fmt.Println("   • aws-monitoring-playbook.md - Ongoing maintenance procedures")

	fmt.Println("\n💡 Client API Advantages for AWS Management:")
	fmt.Println("   • Maintains context across complex multi-step workflows")
	fmt.Println("   • Can reference previous analysis in subsequent steps")
	fmt.Println("   • Perfect for interactive cloud management sessions")
	fmt.Println("   • Enables progressive infrastructure optimization")
	fmt.Println("   • Builds comprehensive documentation iteratively")

	if !awsAvailable {
		fmt.Println("\n🔧 To enable full AWS integration:")
		fmt.Println("   1. Configure AWS credentials (aws configure)")
		fmt.Println("   2. Install AWS MCP: claude mcp add aws-api")
		fmt.Println("   3. Re-run this example for live AWS operations")
	}
}

// executeAWSStep handles a single step in the AWS workflow
func executeAWSStep(ctx context.Context, client claudecode.Client, stepNum int, query string) error {
	fmt.Printf("❓ Query: %s\n", strings.TrimSpace(query))
	fmt.Println(strings.Repeat("-", 50))

	// Send the query
	if err := client.Query(ctx, query); err != nil {
		return fmt.Errorf("failed to send query: %w", err)
	}

	// Process streaming response
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
			case *claudecode.SystemMessage:
				// System messages (can be safely ignored)
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