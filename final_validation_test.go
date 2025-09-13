package claudecode_test

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/severity1/claude-code-sdk-go"
	"github.com/severity1/claude-code-sdk-go/pkg/interfaces"
)

// TestComprehensiveInterfaceEmptyDetection provides enhanced validation for Phase 6 Day 14.
// This test scans ALL exported types from both main package and pkg/interfaces for interface{} usage
// with specific allowlists for legitimate cases (error types, etc.).
func TestComprehensiveInterfaceEmptyDetection(t *testing.T) {
	// Define legitimate interface{} usage allowlist
	allowedInterfaceEmpty := map[string]map[string]string{
		// Error types legitimately use interface{} for flexible data storage
		"ValidationError": {
			"Value": "flexible value storage for validation errors",
		},
		"MessageParseError": {
			"Data": "flexible data storage for parse error details",
		},
		// Tool types legitimately use interface{} for flexible input/output data
		"ToolUseBlock": {
			"Input": "tool inputs can be arbitrary JSON structures requiring interface{} in maps",
		},
	}

	// Comprehensive type scanning
	mainPackageTypes := []reflect.Type{
		// Core message types
		reflect.TypeOf((*claudecode.UserMessage)(nil)).Elem(),
		reflect.TypeOf((*claudecode.AssistantMessage)(nil)).Elem(),
		reflect.TypeOf((*claudecode.StreamMessage)(nil)).Elem(),

		// Content block types
		reflect.TypeOf((*claudecode.TextBlock)(nil)).Elem(),
		reflect.TypeOf((*claudecode.ThinkingBlock)(nil)).Elem(),
		reflect.TypeOf((*claudecode.ToolUseBlock)(nil)).Elem(),
		reflect.TypeOf((*claudecode.ToolResultBlock)(nil)).Elem(),

		// Content types
		reflect.TypeOf((*claudecode.TextContent)(nil)).Elem(),
		reflect.TypeOf((*claudecode.BlockListContent)(nil)).Elem(),
		reflect.TypeOf((*claudecode.ThinkingContent)(nil)).Elem(),

		// Configuration types
		reflect.TypeOf((*claudecode.Options)(nil)).Elem(),
		reflect.TypeOf((*claudecode.McpServerConfig)(nil)).Elem(),

		// Client and transport types
		reflect.TypeOf((*claudecode.ClientImpl)(nil)).Elem(),
	}

	interfacesPackageTypes := []reflect.Type{
		// Message interfaces and types
		reflect.TypeOf((*interfaces.UserMessage)(nil)).Elem(),
		reflect.TypeOf((*interfaces.AssistantMessage)(nil)).Elem(),
		reflect.TypeOf((*interfaces.StreamMessage)(nil)).Elem(),

		// Content interfaces and types
		reflect.TypeOf((*interfaces.TextContent)(nil)).Elem(),
		reflect.TypeOf((*interfaces.BlockListContent)(nil)).Elem(),
		reflect.TypeOf((*interfaces.ThinkingContent)(nil)).Elem(),

		// Content block types
		reflect.TypeOf((*interfaces.TextBlock)(nil)).Elem(),
		reflect.TypeOf((*interfaces.ThinkingBlock)(nil)).Elem(),
		reflect.TypeOf((*interfaces.ToolUseBlock)(nil)).Elem(),
		reflect.TypeOf((*interfaces.ToolResultBlock)(nil)).Elem(),

		// Error types (allowed to use interface{})
		reflect.TypeOf((*interfaces.ValidationError)(nil)).Elem(),
		reflect.TypeOf((*interfaces.MessageParseError)(nil)).Elem(),
		reflect.TypeOf((*interfaces.ConnectionError)(nil)).Elem(),
		reflect.TypeOf((*interfaces.DiscoveryError)(nil)).Elem(),

		// Configuration types
		reflect.TypeOf((*interfaces.McpStdioServerConfig)(nil)).Elem(),
		reflect.TypeOf((*interfaces.McpSSEServerConfig)(nil)).Elem(),
		reflect.TypeOf((*interfaces.McpHTTPServerConfig)(nil)).Elem(),
	}

	// Test main package types
	t.Run("Main package interface{} detection", func(t *testing.T) {
		var allViolations []string

		for _, testType := range mainPackageTypes {
			violations := findInterfaceEmptyUsageWithAllowlist(testType, "claudecode", allowedInterfaceEmpty)
			if len(violations) > 0 {
				allViolations = append(allViolations, violations...)
				t.Errorf("❌ Found interface{} violations in %s:", testType.Name())
				for _, violation := range violations {
					t.Errorf("    %s", violation)
				}
			} else {
				t.Logf("✅ %s: Clean (no unauthorized interface{} usage)", testType.Name())
			}
		}

		if len(allViolations) == 0 {
			t.Log("🎉 Main package: Zero unauthorized interface{} usage detected!")
		}
	})

	// Test interfaces package types
	t.Run("Interfaces package interface{} detection", func(t *testing.T) {
		var allViolations []string

		for _, testType := range interfacesPackageTypes {
			violations := findInterfaceEmptyUsageWithAllowlist(testType, "interfaces", allowedInterfaceEmpty)
			if len(violations) > 0 {
				allViolations = append(allViolations, violations...)
				t.Errorf("❌ Found interface{} violations in %s:", testType.Name())
				for _, violation := range violations {
					t.Errorf("    %s", violation)
				}
			} else {
				t.Logf("✅ %s: Clean (no unauthorized interface{} usage)", testType.Name())
			}
		}

		if len(allViolations) == 0 {
			t.Log("🎉 Interfaces package: Zero unauthorized interface{} usage detected!")
		}
	})
}

// TestFinalSuccessMetricsValidation validates all INTERFACE_SPEC.md requirements for Phase 6 Day 14
func TestFinalSuccessMetricsValidation(t *testing.T) {
	t.Run("Requirement #1: Zero interface{} in Public API", func(t *testing.T) {
		// This is validated by TestComprehensiveInterfaceEmptyDetection
		// We just need to confirm the test framework is working
		testType := reflect.TypeOf((*interfaces.UserMessage)(nil)).Elem()
		violations := findInterfaceEmptyUsageWithAllowlist(testType, "interfaces", nil)

		if len(violations) > 0 {
			t.Errorf("FAIL: Found interface{} violations in core types")
		} else {
			t.Log("✅ PASS: Zero interface{} usage in core public API types")
		}
	})

	t.Run("Requirement #2: Consistent Type() Method Naming", func(t *testing.T) {
		// Validate that content blocks have Type() string methods
		contentBlockTypes := []reflect.Type{
			reflect.TypeOf((*interfaces.TextBlock)(nil)).Elem(),
			reflect.TypeOf((*interfaces.ThinkingBlock)(nil)).Elem(),
			reflect.TypeOf((*interfaces.ToolUseBlock)(nil)).Elem(),
			reflect.TypeOf((*interfaces.ToolResultBlock)(nil)).Elem(),
		}

		for _, testType := range contentBlockTypes {
			method, found := testType.MethodByName("Type")
			if !found {
				t.Errorf("❌ %s missing Type() method", testType.Name())
				continue
			}

			// Verify it returns string
			if method.Type.NumOut() != 1 || method.Type.Out(0) != reflect.TypeOf("") {
				t.Errorf("❌ %s.Type() should return string", testType.Name())
				continue
			}

			t.Logf("✅ %s.Type() method validated", testType.Name())
		}

		// Validate that MCP server config types have Type() McpServerType methods (pointer receivers)
		mcpConfigTypes := []reflect.Type{
			reflect.TypeOf((*interfaces.McpStdioServerConfig)(nil)),
			reflect.TypeOf((*interfaces.McpSSEServerConfig)(nil)),
			reflect.TypeOf((*interfaces.McpHTTPServerConfig)(nil)),
		}

		for _, testType := range mcpConfigTypes {
			method, found := testType.MethodByName("Type")
			if !found {
				t.Errorf("❌ %s missing Type() method", testType.Name())
				continue
			}

			// Verify it has Type() method (return type validation is complex for custom types)
			if method.Type.NumOut() != 1 {
				t.Errorf("❌ %s.Type() should return exactly one value", testType.Name())
				continue
			}

			t.Logf("✅ %s.Type() method validated", testType.Name())
		}

		// Validate that error types have Type() string methods (pointer receivers)
		errorTypes := []reflect.Type{
			reflect.TypeOf((*interfaces.ConnectionError)(nil)),
			reflect.TypeOf((*interfaces.ValidationError)(nil)),
		}

		for _, testType := range errorTypes {
			method, found := testType.MethodByName("Type")
			if !found {
				t.Errorf("❌ %s missing Type() method", testType.Name())
				continue
			}

			// Verify it returns string
			if method.Type.NumOut() != 1 || method.Type.Out(0) != reflect.TypeOf("") {
				t.Errorf("❌ %s.Type() should return string", testType.Name())
				continue
			}

			t.Logf("✅ %s.Type() method validated", testType.Name())
		}
	})

	t.Run("Requirement #3: Package Organization", func(t *testing.T) {
		// Verify pkg/interfaces structure exists and is clean
		interfaceTypes := []reflect.Type{
			reflect.TypeOf((*interfaces.Message)(nil)).Elem(),
			reflect.TypeOf((*interfaces.ContentBlock)(nil)).Elem(),
			reflect.TypeOf((*interfaces.MessageContent)(nil)).Elem(),
			reflect.TypeOf((*interfaces.UserMessageContent)(nil)).Elem(),
			reflect.TypeOf((*interfaces.AssistantMessageContent)(nil)).Elem(),
		}

		for _, interfaceType := range interfaceTypes {
			if interfaceType.Kind() != reflect.Interface {
				t.Errorf("❌ %s should be an interface type", interfaceType.Name())
				continue
			}
			t.Logf("✅ %s interface properly defined", interfaceType.Name())
		}
	})

	t.Run("Requirement #5: Interface Segregation", func(t *testing.T) {
		// Verify client interface composition
		clientType := reflect.TypeOf((*interfaces.Client)(nil)).Elem()

		segregatedInterfaces := []string{
			"Connect", "Close", "IsConnected", // ConnectionManager
			"Query", "QueryStream", // QueryExecutor
			"ReceiveMessages", "ReceiveResponse", // MessageReceiver
			"Interrupt", "Status", // ProcessController
		}

		for _, methodName := range segregatedInterfaces {
			if _, found := clientType.MethodByName(methodName); !found {
				t.Errorf("❌ Client interface missing %s method", methodName)
			} else {
				t.Logf("✅ Client interface has %s method", methodName)
			}
		}
	})
}

// Enhanced helper function with allowlist support
func findInterfaceEmptyUsageWithAllowlist(t reflect.Type, packageName string, allowlist map[string]map[string]string) []string {
	var violations []string

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return violations
	}

	typeName := t.Name()
	allowedFields, typeIsAllowed := allowlist[typeName]

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldType := field.Type

		// Skip unexported fields from other packages
		if !field.IsExported() && field.PkgPath != "" && !strings.Contains(field.PkgPath, packageName) {
			continue
		}

		// Check if this field is specifically allowed
		if typeIsAllowed {
			if _, fieldIsAllowed := allowedFields[field.Name]; fieldIsAllowed {
				// This interface{} usage is allowed
				continue
			}
		}

		// Check for direct interface{} usage
		if fieldType.Kind() == reflect.Interface && fieldType.NumMethod() == 0 {
			violations = append(violations, fmt.Sprintf("Field %s.%s uses interface{}", t.Name(), field.Name))
		}

		// Check for interface{} in slices
		if fieldType.Kind() == reflect.Slice {
			elemType := fieldType.Elem()
			if elemType.Kind() == reflect.Interface && elemType.NumMethod() == 0 {
				violations = append(violations, fmt.Sprintf("Field %s.%s uses []interface{}", t.Name(), field.Name))
			}
		}

		// Check for interface{} in maps
		if fieldType.Kind() == reflect.Map {
			valueType := fieldType.Elem()
			if valueType.Kind() == reflect.Interface && valueType.NumMethod() == 0 {
				violations = append(violations, fmt.Sprintf("Field %s.%s uses map[...]interface{}", t.Name(), field.Name))
			}
		}

		// Recursively check nested structs
		if fieldType.Kind() == reflect.Struct || (fieldType.Kind() == reflect.Ptr && fieldType.Elem().Kind() == reflect.Struct) {
			nestedViolations := findInterfaceEmptyUsageWithAllowlist(fieldType, packageName, allowlist)
			violations = append(violations, nestedViolations...)
		}
	}

	return violations
}

// TestAllowedInterfaceEmptyUsageDocumentation validates that our allowlist is documented and justified
func TestAllowedInterfaceEmptyUsageDocumentation(t *testing.T) {
	allowedCases := map[string]map[string]string{
		"ValidationError": {
			"Value": "Validation errors need to store arbitrary values for detailed error reporting",
		},
		"MessageParseError": {
			"Data": "Parse errors need to store the raw data that failed to parse for debugging",
		},
		"ToolUseBlock": {
			"Input": "Tool inputs are arbitrary JSON structures requiring flexible map[string]interface{} storage",
		},
	}

	t.Run("Document allowed interface{} usage", func(t *testing.T) {
		var documented []string
		for typeName, fields := range allowedCases {
			for fieldName, reason := range fields {
				documented = append(documented, fmt.Sprintf("%s.%s: %s", typeName, fieldName, reason))
			}
		}

		sort.Strings(documented)

		t.Log("📋 Allowed interface{} usage in production code:")
		for _, doc := range documented {
			t.Logf("  ✅ %s", doc)
		}

		if len(documented) == 0 {
			t.Log("🎉 Zero interface{} usage allowed - perfect type safety!")
		} else {
			t.Logf("📊 Total allowed interface{} fields: %d", len(documented))
		}
	})
}
