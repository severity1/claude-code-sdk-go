// Package main demonstrates using Claude Code SDK Client API with Read/Write tools.
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

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	workDir := "./client_tools_demo"
	if err := setupDemoFiles(workDir); err != nil {
		log.Fatalf("Failed to setup demo files: %v", err)
	}
	defer cleanupDemoFiles(workDir)

	originalDir, _ := os.Getwd()
	os.Chdir(workDir)
	defer os.Chdir(originalDir)

	fmt.Printf("📁 Working in directory: %s\n", workDir)
	fmt.Println("📄 Demo files created: package.json, src/main.go, docs/")

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

	conversation := []struct {
		turn        int
		description string
		query       string
	}{
		{
			turn:        1,
			description: "Project Analysis - Read and understand the codebase",
			query:       "Please read all the files in this project (package.json, src/main.go, and any docs) and give me an overview of what this project is about.",
		},
		{
			turn:        2,
			description: "Code Improvement - Enhance the main.go file and create README",
			query:       "Based on what you learned, please improve the src/main.go file with better error handling and create a simple README.md file documenting the project.",
		},
	}

	for _, turn := range conversation {
		fmt.Printf("\n%s\n", strings.Repeat("=", 70))
		fmt.Printf("🗣️  Turn %d: %s\n", turn.turn, turn.description)
		fmt.Printf("❓ Query: %s\n", turn.query)
		fmt.Println(strings.Repeat("-", 70))

		if err := executeInteractiveTurn(ctx, client, turn.turn, turn.query); err != nil {
			log.Printf("Turn %d failed: %v", turn.turn, err)
			continue
		}

		if turn.turn < len(conversation) {
			fmt.Printf("\n⏳ Preparing for next interaction...\n")
			time.Sleep(1 * time.Second)
		}
	}

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
}

func executeInteractiveTurn(ctx context.Context, client claudecode.Client, turnNum int, query string) error {
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
			case *claudecode.UserMessage:
				if blocks, ok := msg.Content.([]claudecode.ContentBlock); ok {
					for _, block := range blocks {
						switch b := block.(type) {
						case *claudecode.TextBlock:
							fmt.Printf("📤 Tool: %s\n", b.Text)
						case *claudecode.ToolResultBlock:
							if contentStr, ok := b.Content.(string); ok {
								displayContent := contentStr
								if strings.Contains(contentStr, "<tool_use_error>") {
									displayContent = strings.ReplaceAll(contentStr, "<tool_use_error>", "⚠️ ")
									displayContent = strings.ReplaceAll(displayContent, "</tool_use_error>", "")
									fmt.Printf("🔧 Tool Issue (id: %s): %s\n", b.ToolUseID[:8]+"...", strings.TrimSpace(displayContent))
								} else if len(displayContent) > 150 {
									fmt.Printf("🔧 Tool Result (id: %s): %s...\n", b.ToolUseID[:8]+"...", displayContent[:150])
								} else {
									fmt.Printf("🔧 Tool Result (id: %s): %s\n", b.ToolUseID[:8]+"...", displayContent)
								}
							} else {
								fmt.Printf("🔧 Tool Result (id: %s): <structured data>\n", b.ToolUseID[:8]+"...")
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

func setupDemoFiles(workDir string) error {
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return err
	}

	srcDir := filepath.Join(workDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		return err
	}

	docsDir := filepath.Join(workDir, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		return err
	}

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

	initialDocs := `# Project Documentation

This directory contains documentation for the demo web application.

Initial version - documentation to be expanded.
`

	if err := os.WriteFile(filepath.Join(docsDir, "initial.md"), []byte(initialDocs), 0644); err != nil {
		return err
	}

	return nil
}

func cleanupDemoFiles(workDir string) {
	os.RemoveAll(workDir)
}

func listProjectFiles(dir string) {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(dir, path)
		if relPath == "." {
			return nil
		}

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