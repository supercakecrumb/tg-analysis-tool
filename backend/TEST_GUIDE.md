# Testing Guide for tg-analysis-tool Backend

## Overview

This guide covers the testing strategy and practices for the tg-analysis-tool backend. We use Go's built-in testing framework with table-driven tests and benchmarks.

## Test Structure

```
backend/
├── internal/
│   ├── api/
│   │   ├── server_test.go          # Server and middleware tests
│   │   └── handlers/
│   │       ├── handlers_test.go    # General handler tests
│   │       └── auth_test.go        # Authentication handler tests
│   ├── config/
│   │   └── config_test.go          # Configuration tests
│   └── infrastructure/
│       └── postgres/
│           └── postgres_test.go    # Database connection tests
```

## Running Tests

### All Tests
```bash
cd backend
go test ./...
```

### Specific Package
```bash
go test ./internal/config
```

### With Coverage
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Verbose Output
```bash
go test -v ./...
```

### Run Benchmarks
```bash
go test -bench=. ./...
```

## Test Categories

### 1. Unit Tests

Unit tests focus on individual functions and methods in isolation.

**Example: Configuration Tests**
- Test environment variable parsing
- Test default values
- Test validation logic
- Test error cases

### 2. Handler Tests

HTTP handler tests use `httptest` package to test API endpoints.

**Example: Authentication Handler Tests**
- Test QR code generation endpoint
- Test phone login flow
- Test verification code submission
- Test 2FA password flow
- Test error responses

### 3. Integration Tests

Integration tests verify component interactions (marked with build tag).

```bash
go test -tags=integration ./...
```

**Note**: Integration tests require:
- Running PostgreSQL instance
- Running Redis instance
- Test database setup

### 4. Benchmark Tests

Performance benchmarks for critical paths.

**Examples**:
- JSON response encoding
- Database query performance
- Middleware overhead
- Authentication flows

## Test Patterns

### Table-Driven Tests

```go
func TestSomething(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    "test",
            expected: "TEST",
            wantErr:  false,
        },
        {
            name:     "empty input",
            input:    "",
            expected: "",
            wantErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := SomeFunction(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("SomeFunction() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if result != tt.expected {
                t.Errorf("SomeFunction() = %v, want %v", result, tt.expected)
            }
        })
    }
}
```

### HTTP Handler Tests

```go
func TestHandler(t *testing.T) {
    h := &Handler{}
    
    req, err := http.NewRequest("GET", "/endpoint", nil)
    if err != nil {
        t.Fatal(err)
    }
    
    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(h.HandleEndpoint)
    handler.ServeHTTP(rr, req)
    
    if status := rr.Code; status != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v want %v",
            status, http.StatusOK)
    }
}
```

## Mock Strategies

### Interface Mocking

For database and external services, we use interface-based mocking:

```go
type MockDB struct {
    HealthFunc func(context.Context) error
}

func (m *MockDB) Health(ctx context.Context) error {
    if m.HealthFunc != nil {
        return m.HealthFunc(ctx)
    }
    return nil
}
```

### Test Fixtures

Common test data is defined in fixtures:

```go
var testUser = &User{
    ID:        "test-user-id",
    Phone:     "+1234567890",
    FirstName: "Test",
    LastName:  "User",
}
```

## Coverage Goals

- **Target**: 80% code coverage
- **Critical paths**: 90%+ coverage
  - Authentication flows
  - Data validation
  - Error handling
  - Security functions

## CI/CD Integration

Tests run automatically on:
- Pull requests
- Commits to main branch
- Pre-deployment

GitHub Actions workflow runs:
1. Unit tests
2. Coverage report
3. Linting
4. Build verification

## Best Practices

1. **Test Naming**: Use descriptive names that explain what is being tested
   - Good: `TestLoginQR_ValidRequest_ReturnsQRCode`
   - Bad: `TestLogin1`

2. **Test Independence**: Each test should be independent and not rely on others

3. **Clean Test Data**: Always clean up test data in integration tests

4. **Error Messages**: Provide clear error messages that help debugging

5. **Parallel Tests**: Use `t.Parallel()` for tests that can run concurrently

6. **Skip Long Tests**: Use `testing.Short()` to skip long-running tests:
   ```go
   if testing.Short() {
       t.Skip("Skipping integration test")
   }
   ```

## Debugging Tests

### Verbose Logging
```bash
go test -v -run TestSpecificFunction ./internal/config
```

### Debug Output
Use `t.Logf()` for debug output that only shows on failure:
```go
t.Logf("Debug: value = %+v", value)
```

### Test Specific Scenarios
```bash
# Run tests matching pattern
go test -run TestLogin ./...

# Run specific subtest
go test -run TestHandler/valid_request ./...
```

## Future Improvements

1. **Add E2E Tests**: Full workflow tests with real Telegram API (test account)
2. **Load Testing**: Use vegeta or similar for API load testing
3. **Mutation Testing**: Verify test quality with mutation testing
4. **Contract Tests**: API contract validation
5. **Security Tests**: Automated security scanning

## Troubleshooting

### Common Issues

1. **Import Errors**: Run `go mod tidy` to resolve dependencies
2. **Database Tests Fail**: Ensure test database is running
3. **Flaky Tests**: Look for time-dependent or order-dependent logic
4. **Slow Tests**: Profile with `go test -cpuprofile`

### Getting Help

- Check test output carefully
- Run with `-v` flag for verbose output
- Use debugger with tests
- Review similar test patterns in codebase