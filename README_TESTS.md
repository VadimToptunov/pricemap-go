# Testing Guide

This document describes the testing strategy and how to run tests for the PriceMap Go project.

## Test Structure

### Unit Tests

Unit tests are located alongside the source code in `*_test.go` files:

- **`utils/`** - Tests for utility functions (validation, currency conversion, cities)
- **`services/`** - Tests for business logic (factors, transport, cache, metrics)
- **`parsers/`** - Tests for parser base functionality
- **`api/`** - Tests for API handlers and endpoints

### Integration Tests

Integration tests are in `api/integration_test.go` and require a running database. Use the `-short` flag to skip them:

```bash
go test ./api -short
```

## Running Tests

### Run All Tests

```bash
go test ./...
```

### Run Tests with Coverage

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Run Specific Package Tests

```bash
# Unit tests only
go test ./utils -v
go test ./services -v
go test ./parsers -v

# API tests (short mode, no DB required)
go test ./api -v -short
```

### Run Tests in Short Mode

Short mode skips integration tests that require database:

```bash
go test ./... -short
```

### Use Test Script

A convenience script is provided:

```bash
./test.sh
```

## Test Coverage

Current test coverage includes:

- ✅ Property validation
- ✅ Currency conversion
- ✅ City data utilities
- ✅ Factor calculation
- ✅ Transport score calculation
- ✅ Cache service
- ✅ Metrics service
- ✅ Parser base functionality
- ✅ API handlers (basic)
- ✅ Heatmap aggregation

## Writing New Tests

### Unit Test Example

```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name string
        input string
        want string
    }{
        {
            name: "normal case",
            input: "test",
            want: "TEST",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := MyFunction(tt.input)
            if got != tt.want {
                t.Errorf("MyFunction() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### API Test Example

```go
func TestHandler_Endpoint(t *testing.T) {
    router := setupTestRouter()
    
    req, _ := http.NewRequest("GET", "/api/v1/endpoint", nil)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusOK, w.Code)
}
```

## Test Dependencies

Tests use the following external packages:

- `github.com/stretchr/testify/assert` - Assertions
- Standard Go `testing` package

## Continuous Integration

Tests are automatically run in CI/CD pipeline (see `.github/workflows/ci.yml`).

## Notes

- Some API tests may fail if database is not connected (expected behavior)
- Integration tests require a running PostgreSQL instance
- Use `-short` flag to skip long-running or integration tests

