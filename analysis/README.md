# Claude Code SDK Python - Analysis Documentation

This directory contains a comprehensive analysis of the Python SDK for implementing a Go version with 100% feature parity.

## Navigation Guide

### 📋 Quick Reference
- **[Feature Matrix](feature-matrix.md)** - Complete feature coverage checklist
- **[Go Implementation Guide](12-go-implementation-guide.md)** - Go-specific architecture decisions

### 🏗️ Architecture & Core Systems
- **[01. Project Structure](01-project-structure.md)** - Build system, dependencies, development workflow
- **[02. Public API](02-public-api.md)** - API surface, exports, design principles
- **[03. Core Types](03-core-types.md)** - Type system, messages, content blocks, configuration
- **[04. Error System](04-error-system.md)** - Error hierarchy and handling patterns

### 🚀 Implementation Details
- **[05. Transport Layer](05-transport-layer.md)** - Abstract transport and subprocess overview
- **[06. Message Parsing](06-message-parsing.md)** - Message parser logic and content handling
- **[07. CLI Integration](07-cli-integration.md)** - CLI discovery, command building, process management
- **[08. Subprocess Details](08-subprocess-details.md)** - Complete subprocess implementation

### 📚 Usage & Patterns
- **[09. Usage Patterns](09-usage-patterns.md)** - Examples, IPython integration, advanced patterns
- **[10. Edge Cases](10-edge-cases.md)** - Critical edge cases and buffering strategies
- **[11. API Evolution](11-api-evolution.md)** - Changelog analysis and feature evolution

## Analysis Overview

**Coverage Level**: 100% - Every implementation detail analyzed
**Files Analyzed**: 15+ Python SDK files including source, examples, tests
**Lines Documented**: 800+ lines of detailed analysis across focused documents

### Analysis Methodology

1. **File-by-File Analysis**: Complete examination of all Python SDK source files
2. **Real-Time Documentation**: Findings documented as discovered
3. **Implementation Focus**: Every detail evaluated for Go implementation requirements
4. **Edge Case Discovery**: Test files analyzed for critical edge cases
5. **Usage Pattern Analysis**: Examples examined for practical integration patterns

### Key Discoveries

- **`close_stdin_after_prompt=True`** - Critical difference between query() and ClaudeSDKClient
- **Active consumption requirement** - Interrupts only work when consuming messages
- **Speculative JSON parsing** - Buffer accumulation strategy for partial JSON
- **5-second termination sequence** - SIGTERM → SIGKILL process management
- **1MB buffer protection** - Memory safety with graceful overflow handling

## For Go Implementation

Start with:
1. **[Feature Matrix](feature-matrix.md)** - Understand complete feature scope
2. **[Go Implementation Guide](12-go-implementation-guide.md)** - Architecture decisions
3. **[Core Types](03-core-types.md)** - Type system foundation
4. **[Transport Layer](05-transport-layer.md)** - Core architecture patterns

## Structure Benefits

- **📁 Focused Files**: Each file covers a specific domain
- **🔗 Cross-Referenced**: Related concepts linked between files  
- **📖 Reference-Friendly**: Quick lookup of specific implementation details
- **🔄 Maintainable**: Easy updates as Python SDK evolves
- **✅ Implementation Ready**: Complete roadmap for Go development