# TDD Implementation Tasks - Claude Code SDK for Go

## Overview

This file tracks the Test-Driven Development (TDD) implementation of the Claude Code SDK for Go, ensuring 100% behavioral parity with the Python SDK reference implementation located at `../claude-code-sdk-python/`.

## TDD Progress Legend

- 🔴 **RED**: Write failing test first
- 🟢 **GREEN**: Implement minimal code to pass test
- 🔵 **BLUE**: Refactor while keeping tests green
- ✅ **DONE**: Feature complete with full test coverage

## Test Parity Summary

**Total Python SDK Tests**: 83 tests across 8 test files  
**Total Go TDD Tasks**: 181 tasks (expanded for Go-specific requirements)  
**Current Progress**: 25/181 (14%)

### Python SDK Test Coverage Analysis
- `test_types.py`: 12 tests → Core types, content blocks, options
- `test_errors.py`: 5 tests → Complete error hierarchy
- `test_message_parser.py`: 17 tests → Message parsing and validation
- `test_transport.py`: 15 tests → CLI discovery and command building
- `test_subprocess_buffering.py`: 7 tests → Critical buffering edge cases
- `test_client.py`: 3 tests → Query function and options
- `test_streaming_client.py`: 20 tests → Client streaming and context management
- `test_integration.py`: 4 tests → End-to-end integration scenarios

### Why More Go Tasks Than Python Tests?

**Expansion Factor**: ~2.2x tasks (181 Go tasks from 83 Python tests)

**Reasons for expansion**:
1. **Go-Specific Requirements**: Interface compliance, type safety, concurrency patterns
2. **Language Differences**: Explicit error handling, memory management, goroutine safety  
3. **Implementation Breakdown**: Complex Python tests split into focused Go test cases
4. **Platform Coverage**: Cross-platform compatibility testing
5. **Performance Testing**: Go-specific benchmarking and optimization validation

## Current Phase: PHASE 1 - Foundation Types & Errors

**Progress**: 25/34 tasks (74%)

---

## PHASE 1: Foundation Types & Errors (34 tasks)

### Message Types (14 tasks) ✅ COMPLETE

#### T001: User Message Creation ✅ DONE
**Python Reference**: `test_types.py::TestMessageTypes::test_user_message_creation`  
**Go Target**: `types_test.go::TestUserMessageCreation`  
**Description**: Test creating UserMessage with string content  
**Acceptance**: Must create UserMessage with content field matching Python SDK structure

#### T002: Assistant Message with Text ✅ DONE
**Python Reference**: `test_types.py::TestMessageTypes::test_assistant_message_with_text`  
**Go Target**: `types_test.go::TestAssistantMessageWithText`  
**Description**: Test AssistantMessage with TextBlock content and model field  
**Acceptance**: Must support content blocks array and model specification

#### T003: Assistant Message with Thinking ✅ DONE
**Python Reference**: `test_types.py::TestMessageTypes::test_assistant_message_with_thinking`  
**Go Target**: `types_test.go::TestAssistantMessageWithThinking`  
**Description**: Test AssistantMessage with ThinkingBlock (thinking + signature)  
**Acceptance**: Must support thinking content and signature fields

#### T004: Tool Use Block Creation ✅ DONE
**Python Reference**: `test_types.py::TestMessageTypes::test_tool_use_block`  
**Go Target**: `types_test.go::TestToolUseBlockCreation`  
**Description**: Test ToolUseBlock with ID, name, and input parameters  
**Acceptance**: Must support tool_use_id, name, and input map

#### T005: Tool Result Block Creation ✅ DONE
**Python Reference**: `test_types.py::TestMessageTypes::test_tool_result_block`  
**Go Target**: `types_test.go::TestToolResultBlockCreation`  
**Description**: Test ToolResultBlock with content and error flag  
**Acceptance**: Must support tool_use_id, content, and is_error fields

#### T006: Result Message Creation ✅ DONE
**Python Reference**: `test_types.py::TestMessageTypes::test_result_message`  
**Go Target**: `types_test.go::TestResultMessageCreation`  
**Description**: Test ResultMessage with timing, cost, and session info  
**Acceptance**: Must include all fields: subtype, duration_ms, duration_api_ms, is_error, num_turns, session_id, total_cost_usd

#### T007: Text Block Implementation ✅ DONE
**Go Target**: `types_test.go::TestTextBlock`  
**Description**: Implement TextBlock content type  
**Acceptance**: Must implement ContentBlock interface with text field

#### T008: Thinking Block Implementation ✅ DONE
**Go Target**: `types_test.go::TestThinkingBlock`  
**Description**: Implement ThinkingBlock content type  
**Acceptance**: Must implement ContentBlock interface with thinking and signature fields

#### T009: System Message Implementation ✅ DONE
**Go Target**: `types_test.go::TestSystemMessage`  
**Description**: Implement SystemMessage with subtype and data preservation  
**Acceptance**: Must support arbitrary data fields and subtype discrimination

#### T010: User Message Mixed Content ✅ DONE
**Go Target**: `types_test.go::TestUserMessageMixedContent`  
**Description**: Test UserMessage with multiple content block types  
**Acceptance**: Must support array of different ContentBlock types

#### T011: Assistant Message Mixed Content ✅ DONE
**Go Target**: `types_test.go::TestAssistantMessageMixedContent`  
**Description**: Test AssistantMessage with text, thinking, and tool use blocks  
**Acceptance**: Must support heterogeneous content block arrays

#### T012: Message Interface Compliance ✅ DONE
**Go Target**: `types_test.go::TestMessageInterface`  
**Description**: Verify all message types implement Message interface  
**Acceptance**: All message types must return correct type strings

#### T013: Content Block Interface Compliance ✅ DONE
**Go Target**: `types_test.go::TestContentBlockInterface`  
**Description**: Verify all content blocks implement ContentBlock interface  
**Acceptance**: All content block types must return correct type strings

#### T014: Message Type Constants ✅ DONE
**Go Target**: `types_test.go::TestMessageTypeConstants`  
**Description**: Define and test message type string constants  
**Acceptance**: Must match Python SDK type strings exactly

### Configuration Options (11 tasks) ✅ COMPLETE

#### T015: Default Options Creation ✅ DONE
**Python Reference**: `test_types.py::TestOptions::test_default_options`  
**Go Target**: `options_test.go::TestDefaultOptions`  
**Description**: Test Options struct with default values  
**Acceptance**: Must match Python SDK defaults: allowed_tools=[], max_thinking_tokens=8000, etc.

#### T016: Options with Tools ✅ DONE
**Python Reference**: `test_types.py::TestOptions::test_claude_code_options_with_tools`  
**Go Target**: `options_test.go::TestOptionsWithTools`  
**Description**: Test Options with allowed_tools and disallowed_tools  
**Acceptance**: Must support tool filtering arrays

#### T017: Permission Mode Options ✅ DONE
**Python Reference**: `test_types.py::TestOptions::test_claude_code_options_with_permission_mode`  
**Go Target**: `options_test.go::TestPermissionModeOptions`  
**Description**: Test all permission modes: default, acceptEdits, plan, bypassPermissions  
**Acceptance**: Must support all four permission mode values

#### T018: System Prompt Options ✅ DONE
**Python Reference**: `test_types.py::TestOptions::test_claude_code_options_with_system_prompt`  
**Go Target**: `options_test.go::TestSystemPromptOptions`  
**Description**: Test system_prompt and append_system_prompt  
**Acceptance**: Must support both primary and append system prompts

#### T019: Session Continuation Options ✅ DONE
**Python Reference**: `test_types.py::TestOptions::test_claude_code_options_with_session_continuation`  
**Go Target**: `options_test.go::TestSessionContinuationOptions`  
**Description**: Test continue_conversation and resume options  
**Acceptance**: Must support conversation state management

#### T020: Model Specification Options ✅ DONE
**Python Reference**: `test_types.py::TestOptions::test_claude_code_options_with_model_specification`  
**Go Target**: `options_test.go::TestModelSpecificationOptions`  
**Description**: Test model and permission_prompt_tool_name  
**Acceptance**: Must support model selection and custom permission tools

#### T021: Functional Options Pattern ✅ DONE
**Go Target**: `options_test.go::TestFunctionalOptionsPattern`  
**Description**: Test WithSystemPrompt, WithAllowedTools, etc. functional options  
**Acceptance**: Must provide fluent configuration API

#### T022: MCP Server Configuration ✅ DONE
**Go Target**: `options_test.go::TestMcpServerConfiguration`  
**Description**: Test MCP server config types (stdio, SSE, HTTP)  
**Acceptance**: Must support all three MCP server configuration types

#### T023: Extra Args Support ✅ DONE
**Go Target**: `options_test.go::TestExtraArgsSupport`  
**Description**: Test arbitrary CLI flag support via ExtraArgs  
**Acceptance**: Must support map[string]*string for custom flags

#### T024: Options Validation ✅ DONE
**Go Target**: `options_test.go::TestOptionsValidation`  
**Description**: Test options field validation and constraints  
**Acceptance**: Must validate option combinations and constraints

#### T025: NewOptions Constructor ✅ DONE
**Go Target**: `options_test.go::TestNewOptionsConstructor`  
**Description**: Test Options creation with functional options  
**Acceptance**: Must apply functional options correctly with defaults

### Error System (9 tasks)

#### T026: Base SDK Error 🔴 RED
**Python Reference**: `test_errors.py::TestErrorTypes::test_base_error`  
**Go Target**: `errors_test.go::TestBaseSDKError`  
**Description**: Test base ClaudeSDKError interface and implementation  
**Acceptance**: Must implement error interface with Type() method

#### T027: CLI Not Found Error 🔴 RED
**Python Reference**: `test_errors.py::TestErrorTypes::test_cli_not_found_error`  
**Go Target**: `errors_test.go::TestCLINotFoundError`  
**Description**: Test CLINotFoundError with helpful installation message  
**Acceptance**: Must inherit from base error and include installation guidance

#### T028: Connection Error 🔴 RED
**Python Reference**: `test_errors.py::TestErrorTypes::test_connection_error`  
**Go Target**: `errors_test.go::TestConnectionError`  
**Description**: Test CLIConnectionError for connection failures  
**Acceptance**: Must represent connection-related failures

#### T029: Process Error with Details 🔴 RED
**Python Reference**: `test_errors.py::TestErrorTypes::test_process_error`  
**Go Target**: `errors_test.go::TestProcessErrorWithDetails`  
**Description**: Test ProcessError with exit_code and stderr  
**Acceptance**: Must include exit_code, stderr fields and formatted error message

#### T030: JSON Decode Error 🔴 RED
**Python Reference**: `test_errors.py::TestErrorTypes::test_json_decode_error`  
**Go Target**: `errors_test.go::TestJSONDecodeError`  
**Description**: Test CLIJSONDecodeError with line and position info  
**Acceptance**: Must include original JSON line and parsing error details

#### T031: Message Parse Error 🔴 RED
**Go Target**: `errors_test.go::TestMessageParseError`  
**Description**: Test MessageParseError with raw data context  
**Acceptance**: Must preserve original data that failed to parse

#### T032: Error Hierarchy 🔴 RED
**Go Target**: `errors_test.go::TestErrorHierarchy`  
**Description**: Verify all errors implement SDKError interface  
**Acceptance**: Must support type checking and error wrapping

#### T033: Error Context Preservation 🔴 RED
**Go Target**: `errors_test.go::TestErrorContextPreservation`  
**Description**: Test error wrapping with fmt.Errorf %w verb  
**Acceptance**: Must support errors.Is() and errors.As() checking

#### T034: Error Message Formatting 🔴 RED
**Go Target**: `errors_test.go::TestErrorMessageFormatting`  
**Description**: Test error message formatting with contextual info  
**Acceptance**: Must provide helpful error messages with suggestions

---

## PHASE 2: Message Parsing & Validation (48 tasks)

### JSON Message Parsing (25 tasks)

#### T035: Parse Valid User Message 🔴 RED
**Python Reference**: `test_message_parser.py::TestMessageParser::test_parse_valid_user_message`  
**Go Target**: `internal/parser/json_test.go::TestParseValidUserMessage`  
**Description**: Parse user message with text content  
**Acceptance**: Must create UserMessage with TextBlock content

#### T036: Parse User Message with Tool Use 🔴 RED
**Python Reference**: `test_message_parser.py::TestMessageParser::test_parse_user_message_with_tool_use`  
**Go Target**: `internal/parser/json_test.go::TestParseUserMessageWithToolUse`  
**Description**: Parse user message with tool_use content block  
**Acceptance**: Must create ToolUseBlock with ID, name, input

#### T037: Parse User Message with Tool Result 🔴 RED
**Python Reference**: `test_message_parser.py::TestMessageParser::test_parse_user_message_with_tool_result`  
**Go Target**: `internal/parser/json_test.go::TestParseUserMessageWithToolResult`  
**Description**: Parse user message with tool_result content block  
**Acceptance**: Must create ToolResultBlock with tool_use_id, content

#### T038: Parse User Message with Tool Result Error 🔴 RED
**Python Reference**: `test_message_parser.py::TestMessageParser::test_parse_user_message_with_tool_result_error`  
**Go Target**: `internal/parser/json_test.go::TestParseUserMessageWithToolResultError`  
**Description**: Parse tool_result with is_error flag  
**Acceptance**: Must handle error tool results correctly

#### T039: Parse User Message Mixed Content 🔴 RED
**Python Reference**: `test_message_parser.py::TestMessageParser::test_parse_user_message_with_mixed_content`  
**Go Target**: `internal/parser/json_test.go::TestParseUserMessageMixedContent`  
**Description**: Parse user message with multiple content block types  
**Acceptance**: Must handle heterogeneous content block arrays

#### T040: Parse Valid Assistant Message 🔴 RED
**Python Reference**: `test_message_parser.py::TestMessageParser::test_parse_valid_assistant_message`  
**Go Target**: `internal/parser/json_test.go::TestParseValidAssistantMessage`  
**Description**: Parse assistant message with text and tool_use  
**Acceptance**: Must create AssistantMessage with mixed content

#### T041: Parse Assistant Message with Thinking 🔴 RED
**Python Reference**: `test_message_parser.py::TestMessageParser::test_parse_assistant_message_with_thinking`  
**Go Target**: `internal/parser/json_test.go::TestParseAssistantMessageWithThinking`  
**Description**: Parse assistant message with thinking block  
**Acceptance**: Must create ThinkingBlock with thinking and signature

#### T042: Parse Valid System Message 🔴 RED
**Python Reference**: `test_message_parser.py::TestMessageParser::test_parse_valid_system_message`  
**Go Target**: `internal/parser/json_test.go::TestParseValidSystemMessage`  
**Description**: Parse system message with subtype  
**Acceptance**: Must create SystemMessage with subtype field

#### T043: Parse Valid Result Message 🔴 RED
**Python Reference**: `test_message_parser.py::TestMessageParser::test_parse_valid_result_message`  
**Go Target**: `internal/parser/json_test.go::TestParseValidResultMessage`  
**Description**: Parse result message with timing and session info  
**Acceptance**: Must create ResultMessage with all required fields

#### T044: Parse Invalid Data Type Error 🔴 RED
**Python Reference**: `test_message_parser.py::TestMessageParser::test_parse_invalid_data_type`  
**Go Target**: `internal/parser/json_test.go::TestParseInvalidDataTypeError`  
**Description**: Handle non-dict input with MessageParseError  
**Acceptance**: Must raise appropriate error for invalid input types

#### T045: Parse Missing Type Field Error 🔴 RED
**Python Reference**: `test_message_parser.py::TestMessageParser::test_parse_missing_type_field`  
**Go Target**: `internal/parser/json_test.go::TestParseMissingTypeFieldError`  
**Description**: Handle missing 'type' field in message  
**Acceptance**: Must detect and report missing type field

#### T046: Parse Unknown Message Type Error 🔴 RED
**Python Reference**: `test_message_parser.py::TestMessageParser::test_parse_unknown_message_type`  
**Go Target**: `internal/parser/json_test.go::TestParseUnknownMessageTypeError`  
**Description**: Handle unknown message types  
**Acceptance**: Must reject unknown message types with clear error

#### T047: Parse User Message Missing Fields 🔴 RED
**Python Reference**: `test_message_parser.py::TestMessageParser::test_parse_user_message_missing_fields`  
**Go Target**: `internal/parser/json_test.go::TestParseUserMessageMissingFields`  
**Description**: Validate required fields in user messages  
**Acceptance**: Must detect missing required fields

#### T048: Parse Assistant Message Missing Fields 🔴 RED
**Python Reference**: `test_message_parser.py::TestMessageParser::test_parse_assistant_message_missing_fields`  
**Go Target**: `internal/parser/json_test.go::TestParseAssistantMessageMissingFields`  
**Description**: Validate required fields in assistant messages  
**Acceptance**: Must detect missing required fields

#### T049: Parse System Message Missing Fields 🔴 RED
**Python Reference**: `test_message_parser.py::TestMessageParser::test_parse_system_message_missing_fields`  
**Go Target**: `internal/parser/json_test.go::TestParseSystemMessageMissingFields`  
**Description**: Validate required fields in system messages  
**Acceptance**: Must detect missing required fields

#### T050: Parse Result Message Missing Fields 🔴 RED
**Python Reference**: `test_message_parser.py::TestMessageParser::test_parse_result_message_missing_fields`  
**Go Target**: `internal/parser/json_test.go::TestParseResultMessageMissingFields`  
**Description**: Validate required fields in result messages  
**Acceptance**: Must detect missing required fields

#### T051: Message Parse Error Contains Data 🔴 RED
**Python Reference**: `test_message_parser.py::TestMessageParser::test_message_parse_error_contains_data`  
**Go Target**: `internal/parser/json_test.go::TestMessageParseErrorContainsData`  
**Description**: Verify MessageParseError preserves original data  
**Acceptance**: Must include original data in parse error

#### T052: Content Block Type Discrimination 🔴 RED
**Go Target**: `internal/parser/json_test.go::TestContentBlockTypeDiscrimination`  
**Description**: Parse content blocks based on type field  
**Acceptance**: Must create correct ContentBlock types

#### T053: JSON Union Type Handling 🔴 RED
**Go Target**: `internal/parser/json_test.go::TestJSONUnionTypeHandling`  
**Description**: Handle JSON union types with custom UnmarshalJSON  
**Acceptance**: Must discriminate types based on "type" field

#### T054: Optional Field Handling 🔴 RED
**Go Target**: `internal/parser/json_test.go::TestOptionalFieldHandling`  
**Description**: Handle optional fields with pointer types  
**Acceptance**: Must distinguish between nil and zero values

#### T055: Raw Data Preservation 🔴 RED
**Go Target**: `internal/parser/json_test.go::TestRawDataPreservation`  
**Description**: Preserve unknown fields for extensibility  
**Acceptance**: Must not lose unrecognized JSON fields

#### T056: Nested Content Parsing 🔴 RED
**Go Target**: `internal/parser/json_test.go::TestNestedContentParsing`  
**Description**: Parse nested content structures  
**Acceptance**: Must handle complex nested JSON correctly

#### T057: Type Safety Validation 🔴 RED
**Go Target**: `internal/parser/json_test.go::TestTypeSafetyValidation`  
**Description**: Ensure type-safe message parsing  
**Acceptance**: Must prevent type confusion attacks

#### T058: Parser Performance 🔴 RED
**Go Target**: `internal/parser/json_test.go::TestParserPerformance`  
**Description**: Benchmark parser performance  
**Acceptance**: Must parse messages efficiently

#### T059: Parser Memory Usage 🔴 RED
**Go Target**: `internal/parser/json_test.go::TestParserMemoryUsage`  
**Description**: Test parser memory efficiency  
**Acceptance**: Must not leak memory during parsing

### Buffering Edge Cases (23 tasks)

#### T060: Multiple JSON Objects Single Line 🔴 RED
**Python Reference**: `test_subprocess_buffering.py::TestSubprocessBuffering::test_multiple_json_objects_on_single_line`  
**Go Target**: `internal/parser/json_test.go::TestMultipleJSONObjectsSingleLine`  
**Description**: Parse multiple JSON objects concatenated on single line  
**Acceptance**: Must handle stdout buffering edge cases correctly

#### T061: Embedded Newlines in JSON Strings 🔴 RED
**Go Target**: `internal/parser/json_test.go::TestEmbeddedNewlinesInJSONStrings`  
**Description**: Handle newlines within JSON string values  
**Acceptance**: Must not break on embedded newlines

#### T062: Buffer Overflow Protection 🔴 RED
**Go Target**: `internal/parser/json_test.go::TestBufferOverflowProtection`  
**Description**: Test 1MB buffer size limit protection  
**Acceptance**: Must reset buffer and return error when limit exceeded

#### T063: Speculative JSON Parsing 🔴 RED
**Go Target**: `internal/parser/json_test.go::TestSpeculativeJSONParsing`  
**Description**: Accumulate partial JSON until complete  
**Acceptance**: Must continue accumulation on parse errors, not fail

#### T064: Partial Message Accumulation 🔴 RED
**Go Target**: `internal/parser/json_test.go::TestPartialMessageAccumulation`  
**Description**: Handle incomplete JSON messages  
**Acceptance**: Must buffer incomplete messages correctly

#### T065: Buffer Reset on Success 🔴 RED
**Go Target**: `internal/parser/json_test.go::TestBufferResetOnSuccess`  
**Description**: Reset buffer after successful JSON parse  
**Acceptance**: Must clear buffer to prevent memory growth

#### T066: Concurrent Buffer Access 🔴 RED
**Go Target**: `internal/parser/json_test.go::TestConcurrentBufferAccess`  
**Description**: Test thread-safe buffer operations  
**Acceptance**: Must handle concurrent access safely

#### T067: Buffer State Management 🔴 RED
**Go Target**: `internal/parser/json_test.go::TestBufferStateManagement`  
**Description**: Manage buffer state across parse attempts  
**Acceptance**: Must maintain consistent buffer state

#### T068: Large Message Handling 🔴 RED
**Go Target**: `internal/parser/json_test.go::TestLargeMessageHandling`  
**Description**: Handle messages approaching buffer limit  
**Acceptance**: Must process large messages efficiently

#### T069: Malformed JSON Recovery 🔴 RED
**Go Target**: `internal/parser/json_test.go::TestMalformedJSONRecovery`  
**Description**: Recover from malformed JSON input  
**Acceptance**: Must continue parsing after encountering bad JSON

#### T070: Line Boundary Edge Cases 🔴 RED
**Go Target**: `internal/parser/json_test.go::TestLineBoundaryEdgeCases`  
**Description**: Handle JSON spanning multiple line boundaries  
**Acceptance**: Must not break on line boundaries within JSON

#### T071: Buffer Size Tracking 🔴 RED
**Go Target**: `internal/parser/json_test.go::TestBufferSizeTracking`  
**Description**: Accurately track buffer size  
**Acceptance**: Must report correct buffer sizes

#### T072: Memory Pressure Handling 🔴 RED
**Go Target**: `internal/parser/json_test.go::TestMemoryPressureHandling`  
**Description**: Handle low memory conditions gracefully  
**Acceptance**: Must degrade gracefully under memory pressure

#### T073: Stream Interruption Recovery 🔴 RED
**Go Target**: `internal/parser/json_test.go::TestStreamInterruptionRecovery`  
**Description**: Recover from stream interruptions  
**Acceptance**: Must handle broken streams

#### T074: JSON Escape Sequence Handling 🔴 RED
**Go Target**: `internal/parser/json_test.go::TestJSONEscapeSequenceHandling`  
**Description**: Correctly handle JSON escape sequences  
**Acceptance**: Must parse escaped characters correctly

#### T075: Unicode String Handling 🔴 RED
**Go Target**: `internal/parser/json_test.go::TestUnicodeStringHandling`  
**Description**: Handle Unicode in JSON strings  
**Acceptance**: Must support full Unicode range

#### T076: Empty Message Handling 🔴 RED
**Go Target**: `internal/parser/json_test.go::TestEmptyMessageHandling`  
**Description**: Handle empty or whitespace-only lines  
**Acceptance**: Must skip empty content appropriately

#### T077: Rapid Message Burst 🔴 RED
**Go Target**: `internal/parser/json_test.go::TestRapidMessageBurst`  
**Description**: Handle rapid succession of messages  
**Acceptance**: Must maintain parsing accuracy under load

#### T078: Parser State Consistency 🔴 RED
**Go Target**: `internal/parser/json_test.go::TestParserStateConsistency`  
**Description**: Maintain consistent parser state  
**Acceptance**: Must not leave parser in inconsistent state

#### T079: Buffer Lifecycle Management 🔴 RED
**Go Target**: `internal/parser/json_test.go::TestBufferLifecycleManagement`  
**Description**: Properly manage buffer lifecycle  
**Acceptance**: Must clean up buffers appropriately

#### T080: Error State Recovery 🔴 RED
**Go Target**: `internal/parser/json_test.go::TestErrorStateRecovery`  
**Description**: Recover from error states  
**Acceptance**: Must continue parsing after errors

#### T081: Performance Under Load 🔴 RED
**Go Target**: `internal/parser/json_test.go::TestPerformanceUnderLoad`  
**Description**: Maintain performance under high message load  
**Acceptance**: Must not degrade significantly under load

#### T082: Memory Leak Prevention 🔴 RED
**Go Target**: `internal/parser/json_test.go::TestMemoryLeakPrevention`  
**Description**: Prevent memory leaks in parser  
**Acceptance**: Must not leak memory over time

---

## PHASE 3: Transport & CLI Integration (38 tasks)

### CLI Discovery & Command Building (15 tasks)

#### T083: CLI Not Found Error 🔴 RED
**Python Reference**: `test_transport.py::TestSubprocessCLITransport::test_find_cli_not_found`  
**Go Target**: `internal/cli/discovery_test.go::TestCLINotFoundError`  
**Description**: Handle CLI binary not found with helpful error  
**Acceptance**: Must include Node.js dependency check and installation guidance

#### T084: Build Basic Command 🔴 RED
**Python Reference**: `test_transport.py::TestSubprocessCLITransport::test_build_command_basic`  
**Go Target**: `internal/cli/discovery_test.go::TestBuildBasicCommand`  
**Description**: Build basic CLI command with required flags  
**Acceptance**: Must include --output-format stream-json --verbose

#### T085: CLI Path Accepts PathLib Path 🔴 RED
**Python Reference**: `test_transport.py::TestSubprocessCLITransport::test_cli_path_accepts_pathlib_path`  
**Go Target**: `internal/cli/discovery_test.go::TestCLIPathAcceptsPath`  
**Description**: Accept both string and path types for CLI path  
**Acceptance**: Must work with filepath.Path types

#### T086: CLI Discovery PATH Lookup 🔴 RED
**Go Target**: `internal/cli/discovery_test.go::TestCLIDiscoveryPATHLookup`  
**Description**: Search for claude in system PATH  
**Acceptance**: Must use exec.LookPath("claude") first

#### T087: CLI Discovery NPM Global 🔴 RED
**Go Target**: `internal/cli/discovery_test.go::TestCLIDiscoveryNPMGlobal`  
**Description**: Search ~/.npm-global/bin/claude  
**Acceptance**: Must check global npm installation location

#### T088: CLI Discovery System Wide 🔴 RED
**Go Target**: `internal/cli/discovery_test.go::TestCLIDiscoverySystemWide`  
**Description**: Search /usr/local/bin/claude  
**Acceptance**: Must check system-wide installation

#### T089: CLI Discovery User Local 🔴 RED
**Go Target**: `internal/cli/discovery_test.go::TestCLIDiscoveryUserLocal`  
**Description**: Search ~/.local/bin/claude  
**Acceptance**: Must check user local installation

#### T090: CLI Discovery Project Local 🔴 RED
**Go Target**: `internal/cli/discovery_test.go::TestCLIDiscoveryProjectLocal`  
**Description**: Search ~/node_modules/.bin/claude  
**Acceptance**: Must check project-local installation

#### T091: CLI Discovery Yarn Global 🔴 RED
**Go Target**: `internal/cli/discovery_test.go::TestCLIDiscoveryYarnGlobal`  
**Description**: Search ~/.yarn/bin/claude  
**Acceptance**: Must check Yarn global installation

#### T092: Node.js Dependency Validation 🔴 RED
**Go Target**: `internal/cli/discovery_test.go::TestNodeJSDependencyValidation`  
**Description**: Validate Node.js is available  
**Acceptance**: Must check for node binary and provide helpful error

#### T093: Command Building All Options 🔴 RED
**Go Target**: `internal/cli/discovery_test.go::TestCommandBuildingAllOptions`  
**Description**: Build command with all configuration options  
**Acceptance**: Must support all CLI flags from Options struct

#### T094: ExtraArgs Support 🔴 RED
**Go Target**: `internal/cli/discovery_test.go::TestExtraArgsSupport`  
**Description**: Support arbitrary CLI flags via ExtraArgs  
**Acceptance**: Must handle map[string]*string for custom flags

#### T095: Close Stdin Flag Handling 🔴 RED
**Go Target**: `internal/cli/discovery_test.go::TestCloseStdinFlagHandling`  
**Description**: Handle --print vs --input-format based on closeStdin  
**Acceptance**: Must differentiate between one-shot and streaming modes

#### T096: Working Directory Validation 🔴 RED
**Go Target**: `internal/cli/discovery_test.go::TestWorkingDirectoryValidation`  
**Description**: Validate working directory exists  
**Acceptance**: Must check cwd exists before starting process

#### T097: CLI Version Detection 🔴 RED
**Go Target**: `internal/cli/discovery_test.go::TestCLIVersionDetection`  
**Description**: Detect Claude CLI version for compatibility  
**Acceptance**: Must support version checking and feature detection

### Subprocess Transport (23 tasks)

#### T098: Subprocess Connection 🔴 RED
**Go Target**: `internal/subprocess/transport_test.go::TestSubprocessConnection`  
**Description**: Establish subprocess connection to Claude CLI  
**Acceptance**: Must start process with proper stdin/stdout/stderr handling

#### T099: Subprocess Disconnection 🔴 RED
**Go Target**: `internal/subprocess/transport_test.go::TestSubprocessDisconnection`  
**Description**: Cleanly disconnect from subprocess  
**Acceptance**: Must terminate process with proper cleanup

#### T100: 5-Second Termination Sequence 🔴 RED
**Go Target**: `internal/subprocess/transport_test.go::TestFiveSecondTerminationSequence`  
**Description**: Implement SIGTERM → wait 5s → SIGKILL sequence  
**Acceptance**: Must follow exact termination timing from Python SDK

#### T101: Process Lifecycle Management 🔴 RED
**Go Target**: `internal/subprocess/transport_test.go::TestProcessLifecycleManagement`  
**Description**: Manage complete process lifecycle  
**Acceptance**: Must handle startup, running, and shutdown states

#### T102: Stdin Message Sending 🔴 RED
**Go Target**: `internal/subprocess/transport_test.go::TestStdinMessageSending`  
**Description**: Send JSON messages to subprocess stdin  
**Acceptance**: Must serialize and send StreamMessage objects

#### T103: Stdout Message Receiving 🔴 RED
**Go Target**: `internal/subprocess/transport_test.go::TestStdoutMessageReceiving`  
**Description**: Receive JSON messages from subprocess stdout  
**Acceptance**: Must parse streaming JSON from stdout

#### T104: Stderr Isolation 🔴 RED
**Go Target**: `internal/subprocess/transport_test.go::TestStderrIsolation`  
**Description**: Isolate stderr using temporary files  
**Acceptance**: Must prevent stderr from causing deadlocks

#### T105: Environment Variable Setting 🔴 RED
**Go Target**: `internal/subprocess/transport_test.go::TestEnvironmentVariableSetting`  
**Description**: Set CLAUDE_CODE_ENTRYPOINT environment variable  
**Acceptance**: Must set to "sdk-go" or "sdk-go-client"

#### T106: Concurrent I/O Handling 🔴 RED
**Go Target**: `internal/subprocess/transport_test.go::TestConcurrentIOHandling`  
**Description**: Handle stdin/stdout concurrently with goroutines  
**Acceptance**: Must not block on I/O operations

#### T107: Process Error Handling 🔴 RED
**Go Target**: `internal/subprocess/transport_test.go::TestProcessErrorHandling`  
**Description**: Handle subprocess errors and exit codes  
**Acceptance**: Must capture exit codes and stderr for errors

#### T108: Message Channel Management 🔴 RED
**Go Target**: `internal/subprocess/transport_test.go::TestMessageChannelManagement`  
**Description**: Manage message and error channels  
**Acceptance**: Must provide separate channels for messages and errors

#### T109: Backpressure Handling 🔴 RED
**Go Target**: `internal/subprocess/transport_test.go::TestBackpressureHandling`  
**Description**: Handle backpressure in message channels  
**Acceptance**: Must prevent blocking when channels are full

#### T110: Context Cancellation 🔴 RED
**Go Target**: `internal/subprocess/transport_test.go::TestContextCancellation`  
**Description**: Support context cancellation throughout transport  
**Acceptance**: Must respect context cancellation and timeouts

#### T111: Resource Cleanup 🔴 RED
**Go Target**: `internal/subprocess/transport_test.go::TestResourceCleanup`  
**Description**: Clean up all resources on shutdown  
**Acceptance**: Must not leak file descriptors or goroutines

#### T112: Process State Tracking 🔴 RED
**Go Target**: `internal/subprocess/transport_test.go::TestProcessStateTracking`  
**Description**: Track subprocess connection state  
**Acceptance**: Must accurately report connection status

#### T113: Interrupt Signal Handling 🔴 RED
**Go Target**: `internal/subprocess/transport_test.go::TestInterruptSignalHandling`  
**Description**: Handle interrupt signals to subprocess  
**Acceptance**: Must send proper interrupt signals

#### T114: Message Ordering Guarantees 🔴 RED
**Go Target**: `internal/subprocess/transport_test.go::TestMessageOrderingGuarantees`  
**Description**: Maintain message ordering through transport  
**Acceptance**: Must preserve message order

#### T115: Transport Reconnection 🔴 RED
**Go Target**: `internal/subprocess/transport_test.go::TestTransportReconnection`  
**Description**: Handle transport reconnection scenarios  
**Acceptance**: Must support reconnecting after disconnection

#### T116: Performance Under Load 🔴 RED
**Go Target**: `internal/subprocess/transport_test.go::TestPerformanceUnderLoad`  
**Description**: Maintain performance under high message throughput  
**Acceptance**: Must handle high-frequency message exchange

#### T117: Memory Usage Optimization 🔴 RED
**Go Target**: `internal/subprocess/transport_test.go::TestMemoryUsageOptimization`  
**Description**: Optimize memory usage in transport layer  
**Acceptance**: Must not accumulate memory over time

#### T118: Error Recovery 🔴 RED
**Go Target**: `internal/subprocess/transport_test.go::TestErrorRecovery`  
**Description**: Recover from transport errors gracefully  
**Acceptance**: Must continue operation after recoverable errors

#### T119: Subprocess Security 🔴 RED
**Go Target**: `internal/subprocess/transport_test.go::TestSubprocessSecurity`  
**Description**: Ensure subprocess runs with minimal permissions  
**Acceptance**: Must prevent privilege escalation

#### T120: Platform Compatibility 🔴 RED
**Go Target**: `internal/subprocess/transport_test.go::TestPlatformCompatibility`  
**Description**: Work across Windows, macOS, and Linux  
**Acceptance**: Must handle platform-specific process management

---

## PHASE 4: Core APIs (43 tasks)

### Query Function (12 tasks)

#### T121: Simple Query Execution 🔴 RED
**Python Reference**: `test_client.py::TestQueryFunction::test_query_single_prompt`  
**Go Target**: `query_test.go::TestSimpleQueryExecution`  
**Description**: Execute simple text query with Query function  
**Acceptance**: Must return MessageIterator for one-shot queries

#### T122: Query with Options 🔴 RED
**Python Reference**: `test_client.py::TestQueryFunction::test_query_with_options`  
**Go Target**: `query_test.go::TestQueryWithOptions`  
**Description**: Execute query with various configuration options  
**Acceptance**: Must apply all configuration options correctly

#### T123: Query Response Processing 🔴 RED
**Go Target**: `query_test.go::TestQueryResponseProcessing`  
**Description**: Process query response messages  
**Acceptance**: Must iterate through response messages correctly

#### T124: Query Error Handling 🔴 RED
**Go Target**: `query_test.go::TestQueryErrorHandling`  
**Description**: Handle errors during query execution  
**Acceptance**: Must propagate errors appropriately

#### T125: Query Context Cancellation 🔴 RED
**Go Target**: `query_test.go::TestQueryContextCancellation`  
**Description**: Support context cancellation in queries  
**Acceptance**: Must respect context deadlines and cancellation

#### T126: Query Stream Input 🔴 RED
**Go Target**: `query_test.go::TestQueryStreamInput`  
**Description**: Execute QueryStream with message channel  
**Acceptance**: Must consume from message channel correctly

#### T127: Query Automatic Cleanup 🔴 RED
**Go Target**: `query_test.go::TestQueryAutomaticCleanup`  
**Description**: Automatically clean up resources after query  
**Acceptance**: Must close transport and clean up automatically

#### T128: Query Transport Selection 🔴 RED
**Go Target**: `query_test.go::TestQueryTransportSelection`  
**Description**: Use appropriate transport for queries  
**Acceptance**: Must use subprocess transport with close_stdin=true

#### T129: Query Message Iterator 🔴 RED
**Go Target**: `query_test.go::TestQueryMessageIterator`  
**Description**: Implement MessageIterator for query results  
**Acceptance**: Must provide iterator interface for streaming results

#### T130: Query Timeout Handling 🔴 RED
**Go Target**: `query_test.go::TestQueryTimeoutHandling`  
**Description**: Handle query timeouts gracefully  
**Acceptance**: Must timeout appropriately with context

#### T131: Query Resource Management 🔴 RED
**Go Target**: `query_test.go::TestQueryResourceManagement`  
**Description**: Properly manage resources during queries  
**Acceptance**: Must not leak resources

#### T132: Query Performance 🔴 RED
**Go Target**: `query_test.go::TestQueryPerformance`  
**Description**: Ensure query performance meets requirements  
**Acceptance**: Must execute queries efficiently

### Client Interface (31 tasks)

#### T133: Client Auto Connect Context Manager 🔴 RED
**Python Reference**: `test_streaming_client.py::TestClaudeSDKClientStreaming::test_auto_connect_with_context_manager`  
**Go Target**: `client_test.go::TestClientAutoConnectContextManager`  
**Description**: Test automatic connection with Go defer pattern  
**Acceptance**: Must auto-connect and clean up with defer

#### T134: Client Manual Connection 🔴 RED
**Go Target**: `client_test.go::TestClientManualConnection`  
**Description**: Test manual Connect/Disconnect lifecycle  
**Acceptance**: Must support explicit connection management

#### T135: Client Query Execution 🔴 RED
**Go Target**: `client_test.go::TestClientQueryExecution`  
**Description**: Execute queries through Client interface  
**Acceptance**: Must send queries via connected client

#### T136: Client Stream Query 🔴 RED
**Go Target**: `client_test.go::TestClientStreamQuery`  
**Description**: Execute QueryStream with message channel  
**Acceptance**: Must handle streaming message input

#### T137: Client Message Reception 🔴 RED
**Go Target**: `client_test.go::TestClientMessageReception`  
**Description**: Receive messages through client channel  
**Acceptance**: Must provide message channel for receiving

#### T138: Client Response Iterator 🔴 RED
**Go Target**: `client_test.go::TestClientResponseIterator`  
**Description**: Get response iterator from client  
**Acceptance**: Must provide MessageIterator for responses

#### T139: Client Interrupt Functionality 🔴 RED
**Go Target**: `client_test.go::TestClientInterruptFunctionality`  
**Description**: Send interrupt through client  
**Acceptance**: Must interrupt ongoing operations

#### T140: Client Session Management 🔴 RED
**Go Target**: `client_test.go::TestClientSessionManagement`  
**Description**: Manage session IDs through client  
**Acceptance**: Must support custom session IDs

#### T141: Client Connection State 🔴 RED
**Go Target**: `client_test.go::TestClientConnectionState`  
**Description**: Track and report connection state  
**Acceptance**: Must accurately report connection status

#### T142: Client Error Propagation 🔴 RED
**Go Target**: `client_test.go::TestClientErrorPropagation`  
**Description**: Propagate errors through client interface  
**Acceptance**: Must surface transport errors appropriately

#### T143: Client Concurrent Access 🔴 RED
**Go Target**: `client_test.go::TestClientConcurrentAccess`  
**Description**: Handle concurrent access to client safely  
**Acceptance**: Must be thread-safe for concurrent operations

#### T144: Client Resource Cleanup 🔴 RED
**Go Target**: `client_test.go::TestClientResourceCleanup`  
**Description**: Clean up client resources properly  
**Acceptance**: Must not leak resources on disconnect

#### T145: Client Configuration Application 🔴 RED
**Go Target**: `client_test.go::TestClientConfigurationApplication`  
**Description**: Apply functional options to client  
**Acceptance**: Must respect all configuration options

#### T146: Client Transport Selection 🔴 RED
**Go Target**: `client_test.go::TestClientTransportSelection`  
**Description**: Use appropriate transport for client mode  
**Acceptance**: Must use subprocess transport with close_stdin=false

#### T147: Client Message Ordering 🔴 RED
**Go Target**: `client_test.go::TestClientMessageOrdering`  
**Description**: Maintain message ordering in client  
**Acceptance**: Must preserve message order

#### T148: Client Backpressure 🔴 RED
**Go Target**: `client_test.go::TestClientBackpressure`  
**Description**: Handle backpressure in client channels  
**Acceptance**: Must handle slow consumers gracefully

#### T149: Client Context Propagation 🔴 RED
**Go Target**: `client_test.go::TestClientContextPropagation`  
**Description**: Propagate context through client operations  
**Acceptance**: Must respect context in all operations

#### T150: Client Reconnection 🔴 RED
**Go Target**: `client_test.go::TestClientReconnection`  
**Description**: Support client reconnection  
**Acceptance**: Must handle reconnection scenarios

#### T151: Client Multiple Sessions 🔴 RED
**Go Target**: `client_test.go::TestClientMultipleSessions`  
**Description**: Handle multiple sessions in single client  
**Acceptance**: Must support session multiplexing

#### T152: Client Performance 🔴 RED
**Go Target**: `client_test.go::TestClientPerformance`  
**Description**: Ensure client performance under load  
**Acceptance**: Must handle high-frequency operations

#### T153: Client Memory Management 🔴 RED
**Go Target**: `client_test.go::TestClientMemoryManagement`  
**Description**: Manage memory efficiently in client  
**Acceptance**: Must not accumulate memory over time

#### T154: Client Graceful Shutdown 🔴 RED
**Go Target**: `client_test.go::TestClientGracefulShutdown`  
**Description**: Shutdown client gracefully  
**Acceptance**: Must complete pending operations before shutdown

#### T155: Client Error Recovery 🔴 RED
**Go Target**: `client_test.go::TestClientErrorRecovery`  
**Description**: Recover from client errors  
**Acceptance**: Must continue operation after recoverable errors

#### T156: Client State Consistency 🔴 RED
**Go Target**: `client_test.go::TestClientStateConsistency`  
**Description**: Maintain consistent client state  
**Acceptance**: Must not enter inconsistent states

#### T157: Client Configuration Validation 🔴 RED
**Go Target**: `client_test.go::TestClientConfigurationValidation`  
**Description**: Validate client configuration  
**Acceptance**: Must reject invalid configurations

#### T158: Client Interface Compliance 🔴 RED
**Go Target**: `client_test.go::TestClientInterfaceCompliance`  
**Description**: Ensure Client interface compliance  
**Acceptance**: Must implement all Client interface methods

#### T159: Client Factory Function 🔴 RED
**Go Target**: `client_test.go::TestClientFactoryFunction`  
**Description**: Test NewClient factory function  
**Acceptance**: Must create properly configured clients

#### T160: Client Option Application Order 🔴 RED
**Go Target**: `client_test.go::TestClientOptionApplicationOrder`  
**Description**: Apply functional options in correct order  
**Acceptance**: Must handle option ordering correctly

#### T161: Client Default Configuration 🔴 RED
**Go Target**: `client_test.go::TestClientDefaultConfiguration`  
**Description**: Use appropriate default configuration  
**Acceptance**: Must have sensible defaults

#### T162: Client Custom Transport 🔴 RED
**Go Target**: `client_test.go::TestClientCustomTransport`  
**Description**: Support custom transport implementations  
**Acceptance**: Must accept custom Transport interfaces

#### T163: Client Protocol Compliance 🔴 RED
**Go Target**: `client_test.go::TestClientProtocolCompliance`  
**Description**: Ensure protocol compliance in client  
**Acceptance**: Must follow Claude Code CLI protocol exactly

---

## PHASE 5: Integration & Advanced Features (18 tasks)

### End-to-End Integration (18 tasks)

#### T164: Simple Query Response Integration 🔴 RED
**Python Reference**: `test_integration.py::TestIntegration::test_simple_query_response`  
**Go Target**: `integration_test.go::TestSimpleQueryResponseIntegration`  
**Description**: End-to-end simple query with text response  
**Acceptance**: Must work with real CLI subprocess

#### T165: Query with Tools Integration 🔴 RED
**Go Target**: `integration_test.go::TestQueryWithToolsIntegration`  
**Description**: End-to-end query with tool usage  
**Acceptance**: Must handle tool use and tool result blocks

#### T166: Streaming Client Integration 🔴 RED
**Go Target**: `integration_test.go::TestStreamingClientIntegration`  
**Description**: End-to-end streaming client interaction  
**Acceptance**: Must maintain persistent connection

#### T167: Interrupt During Streaming 🔴 RED
**Go Target**: `integration_test.go::TestInterruptDuringStreaming`  
**Description**: Test interrupt while actively consuming messages  
**Acceptance**: Must interrupt only when consuming messages

#### T168: Session Continuation Integration 🔴 RED
**Go Target**: `integration_test.go::TestSessionContinuationIntegration`  
**Description**: Continue conversation across client instances  
**Acceptance**: Must preserve conversation state

#### T169: MCP Integration Test 🔴 RED
**Go Target**: `integration_test.go::TestMCPIntegrationTest`  
**Description**: Test MCP server integration  
**Acceptance**: Must work with MCP-enabled Claude CLI

#### T170: Permission Mode Integration 🔴 RED
**Go Target**: `integration_test.go::TestPermissionModeIntegration`  
**Description**: Test different permission modes  
**Acceptance**: Must respect permission mode settings

#### T171: Working Directory Integration 🔴 RED
**Go Target**: `integration_test.go::TestWorkingDirectoryIntegration`  
**Description**: Test working directory specification  
**Acceptance**: Must operate in specified directory

#### T172: Error Handling Integration 🔴 RED
**Go Target**: `integration_test.go::TestErrorHandlingIntegration`  
**Description**: Test error propagation end-to-end  
**Acceptance**: Must surface CLI errors properly

#### T173: Large Response Integration 🔴 RED
**Go Target**: `integration_test.go::TestLargeResponseIntegration`  
**Description**: Handle large responses without issues  
**Acceptance**: Must process large messages efficiently

#### T174: Concurrent Client Integration 🔴 RED
**Go Target**: `integration_test.go::TestConcurrentClientIntegration`  
**Description**: Test multiple concurrent clients  
**Acceptance**: Must handle multiple CLI processes

#### T175: Resource Cleanup Integration 🔴 RED
**Go Target**: `integration_test.go::TestResourceCleanupIntegration`  
**Description**: Verify no resource leaks in integration  
**Acceptance**: Must clean up all processes and resources

#### T176: Performance Integration Test 🔴 RED
**Go Target**: `integration_test.go::TestPerformanceIntegrationTest`  
**Description**: Validate performance in integration scenario  
**Acceptance**: Must meet performance requirements

#### T177: Stress Test Integration 🔴 RED
**Go Target**: `integration_test.go::TestStressTestIntegration`  
**Description**: Stress test with high load  
**Acceptance**: Must remain stable under stress

#### T178: CLI Version Compatibility 🔴 RED
**Go Target**: `integration_test.go::TestCLIVersionCompatibility`  
**Description**: Test compatibility across CLI versions  
**Acceptance**: Must work with supported CLI versions

#### T179: Cross-Platform Integration 🔴 RED
**Go Target**: `integration_test.go::TestCrossPlatformIntegration`  
**Description**: Verify cross-platform compatibility  
**Acceptance**: Must work on Windows, macOS, Linux

#### T180: Network Isolation Integration 🔴 RED
**Go Target**: `integration_test.go::TestNetworkIsolationIntegration`  
**Description**: Test behavior without network access  
**Acceptance**: Must handle offline scenarios gracefully

#### T181: Full Feature Integration 🔴 RED
**Go Target**: `integration_test.go::TestFullFeatureIntegration`  
**Description**: Exercise all features in single test  
**Acceptance**: Must demonstrate complete functionality

---

## Progress Tracking

### Overall Progress
- **Total Tasks**: 181 tasks
- **Completed**: 25 ✅ (14%)
- **In Progress**: 0 🔵 (0%)
- **Ready for Implementation**: 9 🔴 (Phase 1 remaining)

### Phase Progress
- **Phase 1**: 25/34 (74%) - Foundation Types & Errors
- **Phase 2**: 0/48 (0%) - Message Parsing & Validation  
- **Phase 3**: 0/38 (0%) - Transport & CLI Integration
- **Phase 4**: 0/43 (0%) - Core APIs
- **Phase 5**: 0/18 (0%) - Integration & Advanced Features

### Next Recommended Tasks
1. **T026**: Base SDK Error (Error handling foundation)
2. **T027**: CLI Not Found Error (Complete error system)
3. **T028**: Connection Error (Transport error handling)

### Implementation Guidelines

**TDD Cycle Process**:
1. 🔴 **RED**: Write the failing test first, ensuring it fails for the right reason
2. 🟢 **GREEN**: Write minimal code to make the test pass (no more, no less)
3. 🔵 **BLUE**: Refactor to improve code quality while keeping all tests green
4. ✅ **DONE**: Mark complete when feature passes all tests and meets acceptance criteria

**Acceptance Criteria**: Each task must exactly match the behavior of the corresponding Python SDK test to ensure 100% behavioral parity.

**Test File Organization**: Tests are organized alongside implementation files following Go conventions, with integration tests in dedicated files.

---

*This document tracks progress toward 100% feature parity with the Python SDK through systematic Test-Driven Development.*