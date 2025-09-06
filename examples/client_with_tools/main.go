// Package main demonstrates using Claude Code SDK Client API with Read/Write tools.
//
// This example shows how to:
// - Configure Client API to use specific tools across multiple interactions
// - Execute multi-turn conversations that build context with file operations
// - Handle streaming responses that include both text and tool results
// - Demonstrate interactive file manipulation workflows
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/severity1/claude-code-sdk-go"
)

func main() {
	fmt.Println("Claude Code SDK for Go - Client API with Read/Write Tools Example")
	fmt.Println("================================================================")

	// Create context with extended timeout for interactive session
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Setup working directory and sample files
	workDir := "./client_tools_demo"
	if err := setupDemoFiles(workDir); err != nil {
		log.Fatalf("Failed to setup demo files: %v", err)
	}
	defer cleanupDemoFiles(workDir)

	// Change to the working directory so Claude can access the files
	originalDir, _ := os.Getwd()
	os.Chdir(workDir)
	defer os.Chdir(originalDir)

	fmt.Printf("📁 Working in directory: %s\n", workDir)
	fmt.Println("📄 Demo files created: package.json, src/main.go, docs/")

	// Create client with tool configuration
	fmt.Println("\n🔗 Creating client with Read/Write/Edit tools enabled...")
	client := claudecode.NewClient(
		claudecode.WithAllowedTools("Read", "Write", "Edit"),
		claudecode.WithSystemPrompt("You are a helpful software development assistant. You can read, write, and edit files to help with coding tasks. Be thorough and explain your actions clearly."),
	)

	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect client: %v", err)
	}

	defer func() {
		fmt.Println("\n🧹 Ending interactive session...")
		if err := client.Disconnect(); err != nil {
			log.Printf("Warning: Failed to disconnect cleanly: %v", err)
		}
		fmt.Println("👋 Session ended!")
	}()

	fmt.Println("✅ Connected! Starting interactive file manipulation session...")

	// Interactive conversation flow demonstrating tool usage with context
	conversation := []struct {
		turn        int
		description string
		query       string
	}{
		{
			turn:        1,
			description: "Project Analysis - Read and understand the codebase",
			query:       "Please read all the files in this project (package.json, src/main.go, and any docs) and give me an overview of what this project is about. What's its purpose and current state?",
		},
		{
			turn:        2,
			description: "Code Improvement - Enhance the main.go file based on analysis",
			query:       "Based on what you learned about this project, please improve the src/main.go file. Add better error handling, more features, and improve the code structure. Keep it consistent with the package.json dependencies.",
		},
		{
			turn:        3,
			description: "Documentation Creation - Generate comprehensive docs",
			query:       "Now create comprehensive documentation for this improved project. Write a detailed README.md file and add code comments. Update or create any other documentation files that would be helpful.",
		},
		{
			turn:        4,
			description: "Configuration Updates - Enhance project configuration",
			query:       "Review the package.json and suggest improvements. Add any missing scripts, dependencies, or configuration that would make this a more robust project. Update the file with your improvements.",
		},
		{
			turn:        5,
			description: "Final Review - Validate all changes",
			query:       "Please read all the files again and give me a summary of all the changes you made. Create a CHANGELOG.md file documenting the improvements.",
		},
	}

	// Execute the interactive conversation
	for _, turn := range conversation {
		fmt.Printf("\n%s\n", strings.Repeat("=", 70))
		fmt.Printf("🗣️  Turn %d: %s\n", turn.turn, turn.description)
		fmt.Printf("❓ Query: %s\n", turn.query)
		fmt.Println(strings.Repeat("-", 70))

		if err := executeInteractiveTurn(ctx, client, turn.turn, turn.query); err != nil {
			log.Printf("Turn %d failed: %v", turn.turn, err)
			continue
		}

		// Brief pause between turns for readability
		if turn.turn < len(conversation) {
			fmt.Printf("\n⏳ Preparing for next interaction...\n")
			time.Sleep(1 * time.Second)
		}
	}

	// Show final results
	fmt.Printf("\n%s\n", strings.Repeat("=", 70))
	fmt.Println("🎉 Interactive Client with Tools Session Completed!")
	fmt.Println("\n✨ What was demonstrated:")
	fmt.Println("   • Multi-turn conversation with persistent tool context")
	fmt.Println("   • Progressive file analysis and improvement")
	fmt.Println("   • Context preservation across tool operations")
	fmt.Println("   • Interactive development workflow")
	fmt.Println("   • Real-time streaming responses with tool results")

	fmt.Println("\n📂 Final project structure:")
	listProjectFiles(workDir)

	fmt.Println("\n💡 Key advantages of Client API for tool usage:")
	fmt.Println("   • Context maintained across multiple tool operations")
	fmt.Println("   • Can build complex workflows step by step")
	fmt.Println("   • Perfect for interactive development sessions")
	fmt.Println("   • Tool results inform subsequent interactions")
}

// executeInteractiveTurn handles a single turn in the conversation
func executeInteractiveTurn(ctx context.Context, client claudecode.Client, turnNum int, query string) error {
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
						fmt.Printf("\n💭 [Claude is analyzing: %s]\n", b.Thinking)
					}
				}
			case *claudecode.SystemMessage:
				// System messages (can be safely ignored)
			case *claudecode.ResultMessage:
				if msg.IsError {
					return fmt.Errorf("claude returned error: %s", msg.Result)
				}
				goto turnComplete
			default:
				fmt.Printf("\n📦 Received message type: %T\n", message)
			}
		case <-time.After(45 * time.Second):
			fmt.Printf("\n⏰ Turn %d timed out\n", turnNum)
			goto turnComplete
		}
	}

turnComplete:
	if !responseReceived {
		return fmt.Errorf("no response received for turn %d", turnNum)
	}

	fmt.Printf("\n✅ Turn %d completed successfully\n", turnNum)
	return nil
}

// setupDemoFiles creates a sample project for the demonstration
func setupDemoFiles(workDir string) error {
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return err
	}

	// Create src directory
	srcDir := filepath.Join(workDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		return err
	}

	// Create docs directory
	docsDir := filepath.Join(workDir, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		return err
	}

	// Create package.json
	packageJSON := `{
  "name": "demo-web-app",
  "version": "1.0.0",
  "description": "A sample web application for demonstrating Claude Code SDK",
  "main": "src/main.go",
  "scripts": {
    "build": "go build -o bin/app src/main.go"
  },
  "keywords": ["demo", "web", "go"],
  "author": "Claude Code SDK Demo",
  "license": "MIT",
  "dependencies": {
    "express": "^4.18.0"
  }
}`

	if err := os.WriteFile(filepath.Join(workDir, "package.json"), []byte(packageJSON), 0644); err != nil {
		return err
	}

	// Create basic main.go
	mainGo := `package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	})
	
	fmt.Println("Server starting on :8080")
	http.ListenAndServe(":8080", nil)
}
`

	if err := os.WriteFile(filepath.Join(srcDir, "main.go"), []byte(mainGo), 0644); err != nil {
		return err
	}

	// Create initial docs
	initialDocs := `# Project Documentation

This directory contains documentation for the demo web application.

Initial version - documentation to be expanded.
`

	if err := os.WriteFile(filepath.Join(docsDir, "initial.md"), []byte(initialDocs), 0644); err != nil {
		return err
	}

	return nil
}

// cleanupDemoFiles removes the demo directory
func cleanupDemoFiles(workDir string) {
	os.RemoveAll(workDir)
}

// listProjectFiles shows the final project structure
func listProjectFiles(dir string) {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Get relative path from workDir
		relPath, _ := filepath.Rel(dir, path)
		if relPath == "." {
			return nil
		}

		// Calculate indentation based on depth
		depth := strings.Count(relPath, string(os.PathSeparator))
		indent := strings.Repeat("  ", depth)

		if info.IsDir() {
			fmt.Printf("   %s📁 %s/\n", indent, filepath.Base(path))
		} else {
			fmt.Printf("   %s📄 %s (%d bytes)\n", indent, filepath.Base(path), info.Size())
		}
		return nil
	})
}