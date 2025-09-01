# Client Test Conversion Plan

## Overview
Convert `client_test.go` (3,583 lines, 22 test functions) to table-driven tests in a new file `client_table_test.go` while maintaining 100% test parity and following Go idioms.

## Current Analysis
- **22 distinct test functions** with 68+ sub-tests using `t.Run()`
- **Repetitive patterns**: Context setup (24x), transport creation, client lifecycle
- **Mock transport**: Single comprehensive `clientMockTransport` with error injection
- **Test categories**: Connection, queries, errors, config validation, concurrency, resources

## Test Function Inventory

### Connection Management (4 functions)
- `TestClientAutoConnectContextManager` - Defer-based resource management
- `TestClientManualConnection` - Manual connect/disconnect lifecycle
- `TestClientConnectionState` - Connection state validation
- `TestClientSessionManagement` - Session handling

### Query Operations (3 functions)
- `TestClientQueryExecution` - Single query execution
- `TestClientStreamQuery` - Stream query handling
- `TestClientMessageReception` - Message reception/iteration

### Error Handling (3 functions)
- `TestClientErrorPropagation` - Transport error propagation (5 scenarios)
- `TestClientInterruptFunctionality` - Interrupt handling
- `TestClientResponseIterator` - Response iteration errors

### Configuration & Validation (4 functions)
- `TestClientConfigurationValidation` - Config validation (8 sub-tests)
- `TestClientInterfaceCompliance` - Interface compliance (9 sub-tests)
- `TestClientConfigurationApplication` - Option application (10 sub-tests)
- `TestClientDefaultConfiguration` - Default values (6 sub-tests)

### Concurrency & Performance (4 functions)
- `TestClientConcurrentAccess` - Concurrent operations
- `TestClientContextPropagation` - Context handling (10 sub-tests)
- `TestClientBackpressure` - Backpressure handling (3 sub-tests)
- `TestClientMessageOrdering` - Message ordering (6 sub-tests)

### Resource Management (2 functions)
- `TestClientResourceCleanup` - Resource cleanup (5 sub-tests)
- `TestClientFactoryFunction` - Factory functions (6 sub-tests)

### Advanced Features (2 functions)
- `TestClientTransportSelection` - Transport selection
- `TestClientMultipleSessions` - Multiple session handling (4 sub-tests)

## Implementation Phases

### Phase 1: Extract Common Helpers ✅
**Goal**: Create reusable helpers to eliminate repetitive code

**Tasks**:
- [ ] Create `setupTestContext(timeout time.Duration) (context.Context, context.CancelFunc)` helper
- [ ] Create `newMockTransport(options ...MockTransportOption) *clientMockTransport` factory
- [ ] Create `setupClient(transport Transport, options ...ClientOption) *Client` helper
- [ ] Create `connectClient(ctx context.Context, client *Client) error` helper
- [ ] Create assertion helpers: `assertConnected()`, `assertDisconnected()`, `assertError()`

**Expected Impact**: Eliminates 24+ repetitive setup patterns

### Phase 2: Convert High-Impact Functions (Table-Driven) ⏳
**Goal**: Convert functions with clear table-driven potential

**Priority 1**:
- [ ] `TestClientConfigurationValidation` → `TestClientConfigurationValidation_TableDriven`
  - 8 sub-tests with clear input/output patterns
  - Various config combinations and expected outcomes
- [ ] `TestClientErrorPropagation` → `TestClientErrorPropagation_TableDriven`  
  - 5 distinct error injection scenarios
  - Clear error type/message expectations
- [ ] `TestClientContextPropagation` → `TestClientContextPropagation_TableDriven`
  - 10 sub-tests with context scenarios
  - Timeout, cancellation, and deadline patterns

**Expected Impact**: ~30% reduction in these test functions

### Phase 3: Systematic Conversion 📋
**Goal**: Convert remaining suitable functions

**Configuration Tests**:
- [ ] `TestClientConfigurationApplication` (10 sub-tests)
- [ ] `TestClientDefaultConfiguration` (6 sub-tests)
- [ ] `TestClientInterfaceCompliance` (9 sub-tests)

**Resource & Lifecycle Tests**:
- [ ] `TestClientResourceCleanup` (5 sub-tests)
- [ ] `TestClientFactoryFunction` (6 sub-tests)

**Concurrency Tests** (Keep existing structure - complex async patterns):
- ✅ `TestClientConcurrentAccess` - Complex concurrency, keep as-is
- ✅ `TestClientBackpressure` - Performance testing, keep as-is  
- ✅ `TestClientMessageOrdering` - Timing-sensitive, keep as-is

### Phase 4: Code Organization & Optimization 🔄
**Goal**: Final organization and verification

**Tasks**:
- [ ] Create `client_table_test.go` with converted tests
- [ ] Group related table tests logically
- [ ] Add comprehensive test documentation
- [ ] Verify 100% test parity with original
- [ ] Benchmark performance comparison
- [ ] Update test running instructions in CLAUDE.md

## Table-Driven Test Structure Template

```go
func TestClientFeature_TableDriven(t *testing.T) {
    tests := []struct {
        name           string
        transportSetup func() *clientMockTransport
        clientOptions  []ClientOption
        operation      func(ctx context.Context, client *Client) error
        wantErr        bool
        wantErrType    error
        validate       func(t *testing.T, transport *clientMockTransport, client *Client)
    }{
        // Test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx, cancel := setupTestContext(5 * time.Second)
            defer cancel()
            
            transport := tt.transportSetup()
            client := setupClient(transport, tt.clientOptions...)
            defer client.Disconnect()
            
            err := tt.operation(ctx, client)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("operation() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if tt.validate != nil {
                tt.validate(t, transport, client)
            }
        })
    }
}
```

## Helper Function Specifications

### Context Helpers
```go
// setupTestContext creates a context with timeout and cancel function
func setupTestContext(timeout time.Duration) (context.Context, context.CancelFunc)

// setupTestContextWithDeadline creates context with specific deadline
func setupTestContextWithDeadline(deadline time.Time) (context.Context, context.CancelFunc)
```

### Transport Helpers
```go
// MockTransportOption configures mock transport
type MockTransportOption func(*clientMockTransport)

func WithConnectError(err error) MockTransportOption
func WithSendError(err error) MockTransportOption
func WithSlowSend() MockTransportOption
func WithAsyncError(err error) MockTransportOption

// newMockTransport creates configured mock transport
func newMockTransport(options ...MockTransportOption) *clientMockTransport
```

### Client Helpers
```go
// setupClient creates client with transport and options
func setupClient(transport Transport, options ...ClientOption) *Client

// connectClient connects client with error handling
func connectClient(ctx context.Context, client *Client) error

// disconnectClient safely disconnects client
func disconnectClient(client *Client) error
```

### Assertion Helpers
```go
// assertConnected verifies transport connection state
func assertConnected(t *testing.T, transport *clientMockTransport)

// assertDisconnected verifies transport disconnection
func assertDisconnected(t *testing.T, transport *clientMockTransport)

// assertError verifies error type and message
func assertError(t *testing.T, err error, wantType error, wantMsg string)

// assertMessageCount verifies sent message count
func assertMessageCount(t *testing.T, transport *clientMockTransport, expected int)
```

## Quality Assurance

### Test Parity Verification
- [ ] Compare test coverage reports before/after conversion
- [ ] Verify all error conditions are still tested
- [ ] Ensure all edge cases remain covered
- [ ] Validate timing-sensitive tests still work

### Performance Impact
- [ ] Benchmark test execution time before/after
- [ ] Verify no degradation in test performance
- [ ] Measure memory usage during test runs

### Code Quality Checks
- [ ] Run `go vet ./...` on new test file
- [ ] Run `golangci-lint run` for linting
- [ ] Verify adherence to Go testing conventions
- [ ] Check for proper test naming and organization

## Expected Benefits

### Quantitative Improvements
- **~40% code reduction** through helper extraction
- **24+ eliminated repetitive patterns** (context setup, transport creation)
- **Consistent error handling** across all tests
- **Improved test maintainability** with standardized patterns

### Qualitative Improvements
- **Better Go idioms** with table-driven approach
- **Easier test case addition** for new scenarios
- **Cleaner separation** of test logic and test data
- **Enhanced readability** with consistent structure
- **Simplified debugging** with standardized helpers

## Migration Strategy

1. **Parallel Development**: Create new file alongside existing tests
2. **Gradual Migration**: Convert functions in phases
3. **Continuous Verification**: Maintain test parity at each step
4. **Final Cutover**: Replace old tests once verification complete
5. **Documentation Update**: Update CLAUDE.md with new patterns

## Risk Mitigation

- **Backup Strategy**: Keep original `client_test.go` until full verification
- **Incremental Approach**: Convert and verify each function individually  
- **Regression Testing**: Run full test suite after each phase
- **Code Review**: Thorough review of table-driven test logic
- **Performance Monitoring**: Track test execution metrics throughout conversion