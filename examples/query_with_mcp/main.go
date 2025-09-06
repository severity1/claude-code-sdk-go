// Package main demonstrates using Claude Code SDK Query API with MCP tools (AWS API).
//
// This example shows how to:
// - Configure Query API to use MCP tools (AWS API in this case)
// - Execute queries that automatically invoke AWS operations
// - Handle responses that include both analysis and AWS API results
// - Combine MCP tools with file operations for comprehensive automation
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
	fmt.Println("Claude Code SDK for Go - Query API with MCP Tools (AWS API) Example")
	fmt.Println("===================================================================")

	// Check if AWS credentials are available
	if !checkAWSCredentials() {
		fmt.Println("⚠️  AWS credentials not found.")
		fmt.Println("Please configure AWS credentials using one of these methods:")
		fmt.Println("  • Environment variables: AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY")
		fmt.Println("  • AWS credentials file: ~/.aws/credentials")
		fmt.Println("  • AWS CLI: `aws configure`")
		fmt.Println("\n📚 This example will show how the code works with mock explanations.")
	}

	// Create context with timeout for AWS operations
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// Example 1: AWS Resource Discovery with Query API
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🔍 EXAMPLE 1: AWS Resource Discovery")
	fmt.Println(strings.Repeat("=", 60))

	query1 := `Please help me understand my current AWS infrastructure. 
Can you:
1. List my EC2 instances across all regions
2. Show my S3 buckets and their sizes  
3. Check my IAM users and their permissions
4. Create a summary report and save it to aws-infrastructure-report.txt

Please be thorough and explain what you find.`

	if err := runQueryWithMCP(ctx, "AWS Infrastructure Audit", query1, []string{"mcp__aws-api-mcp__*", "Write"}); err != nil {
		log.Printf("Query 1 failed: %v", err)
	}

	// Example 2: AWS Cost Analysis and Optimization
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💰 EXAMPLE 2: AWS Cost Analysis and Optimization")
	fmt.Println(strings.Repeat("=", 60))

	query2 := `Please analyze my AWS spending and suggest optimizations:
1. Get my current billing information and cost breakdown
2. Identify any idle or underutilized resources (EC2, RDS, etc.)
3. Check for unattached EBS volumes or unused Elastic IPs
4. Suggest cost optimization strategies
5. Create a detailed cost analysis report in aws-cost-analysis.md

Focus on actionable recommendations to reduce costs.`

	if err := runQueryWithMCP(ctx, "AWS Cost Optimization", query2, []string{"mcp__aws-api-mcp__*", "Write"}); err != nil {
		log.Printf("Query 2 failed: %v", err)
	}

	// Example 3: Security Audit with AWS MCP
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🔒 EXAMPLE 3: AWS Security Audit")
	fmt.Println(strings.Repeat("=", 60))

	query3 := `Please perform a security audit of my AWS account:
1. Check security groups for overly permissive rules (0.0.0.0/0)
2. Review IAM policies for excessive permissions
3. Verify MFA is enabled for IAM users
4. Check S3 bucket public access settings
5. Review CloudTrail logging status
6. Generate a security findings report in aws-security-audit.json

Prioritize findings by risk level and provide remediation steps.`

	if err := runQueryWithMCP(ctx, "AWS Security Audit", query3, []string{"mcp__aws-api-mcp__*", "Write"}); err != nil {
		log.Printf("Query 3 failed: %v", err)
	}

	// Example 4: Automated AWS Resource Management
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🤖 EXAMPLE 4: Automated AWS Resource Management")
	fmt.Println(strings.Repeat("=", 60))

	query4 := `Help me manage my AWS resources automatically:
1. Create a new S3 bucket for project backups with appropriate encryption
2. Set up lifecycle policies for cost optimization
3. Tag all my untagged resources with Environment=production
4. Create CloudWatch alarms for high CPU usage on my EC2 instances
5. Document all changes made in aws-automation-log.txt

Please explain each step and confirm before making changes.`

	if err := runQueryWithMCP(ctx, "AWS Automation", query4, []string{"mcp__aws-api-mcp__*", "Write", "Read"}); err != nil {
		log.Printf("Query 4 failed: %v", err)
	}

	// Example 5: Restricted MCP Access (Security Best Practice)
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🛡️  EXAMPLE 5: Read-Only AWS Analysis (Security Restricted)")
	fmt.Println(strings.Repeat("=", 60))

	query5 := `Please give me a comprehensive overview of my AWS environment, but only read/list information - do not make any changes:
1. List all resources across services (EC2, RDS, Lambda, etc.)
2. Analyze resource utilization and health
3. Check compliance with AWS best practices  
4. Create an executive summary of my cloud infrastructure
5. Save the analysis to aws-readonly-analysis.md

This should be a read-only analysis with no modifications to AWS resources.`

	fmt.Println("Note: This query only allows AWS read operations and file writing for security")
	if err := runQueryWithMCP(ctx, "Read-Only AWS Analysis", query5, []string{"mcp__aws-api-mcp__describe*", "mcp__aws-api-mcp__list*", "mcp__aws-api-mcp__get*", "Write"}); err != nil {
		log.Printf("Query 5 failed: %v", err)
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎉 Query with MCP Tools Examples Completed!")
	fmt.Println("\n✨ What was demonstrated:")
	fmt.Println("   • AWS API integration through MCP protocol")
	fmt.Println("   • Automated AWS resource discovery and analysis")
	fmt.Println("   • Cost optimization and security auditing")
	fmt.Println("   • Automated resource management with AWS APIs")
	fmt.Println("   • Security through tool restriction (read-only operations)")
	fmt.Println("   • Combining MCP tools with file operations")

	fmt.Println("\n🔧 MCP Tool Integration Benefits:")
	fmt.Println("   • One-shot AWS operations with natural language")
	fmt.Println("   • Automatic AWS credential management")
	fmt.Println("   • Multi-service AWS workflows in single queries")
	fmt.Println("   • Built-in error handling and retry logic")
	fmt.Println("   • Security through granular tool permissions")

	fmt.Println("\n📂 Generated reports (if AWS credentials available):")
	fmt.Println("   • aws-infrastructure-report.txt")
	fmt.Println("   • aws-cost-analysis.md") 
	fmt.Println("   • aws-security-audit.json")
	fmt.Println("   • aws-automation-log.txt")
	fmt.Println("   • aws-readonly-analysis.md")
}

// runQueryWithMCP executes a query with MCP tool restrictions
func runQueryWithMCP(ctx context.Context, title, question string, allowedTools []string) error {
	fmt.Printf("\n🎯 %s\n", title)
	fmt.Printf("🔧 Allowed tools: %v\n", allowedTools)
	fmt.Printf("❓ Query: %s\n", strings.TrimSpace(question))
	fmt.Println(strings.Repeat("-", 50))

	// Create query with MCP tool restrictions
	iterator, err := claudecode.Query(ctx, question,
		claudecode.WithAllowedTools(allowedTools...),
		claudecode.WithSystemPrompt("You are an AWS cloud expert that can help with infrastructure management, cost optimization, and security. Use AWS MCP tools to gather information and make changes as requested. Always explain your actions clearly and provide detailed analysis."),
	)
	if err != nil {
		return fmt.Errorf("failed to create query: %w", err)
	}
	defer iterator.Close()

	fmt.Println("🤖 Claude's Response:")
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
					fmt.Print(b.Text)
				case *claudecode.ThinkingBlock:
					fmt.Printf("\n💭 [Analyzing AWS: %s]\n", b.Thinking)
				}
			}
		case *claudecode.SystemMessage:
			// System initialization (can be ignored)
		case *claudecode.ResultMessage:
			if msg.IsError {
				return fmt.Errorf("claude returned error: %s", msg.Result)
			}
		default:
			fmt.Printf("\n📦 Unexpected message type: %T\n", message)
		}
	}

	if !responseReceived {
		return fmt.Errorf("no response received")
	}

	fmt.Println("\n✅ Query completed successfully")
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