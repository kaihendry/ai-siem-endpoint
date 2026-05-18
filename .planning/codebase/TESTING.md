# Testing Patterns

**Analysis Date:** 2026-05-18

## Test Framework

**Runner:**
- None currently configured. No `*_test.go` files exist in the repository.
- Go's built-in `testing` package would be the natural choice when tests are added.
- Config: Not applicable

**Assertion Library:**
- Not applicable (no tests present)

**Run Commands:**
```bash
go test ./...          # Run all tests (none exist yet)
go test -v ./...       # Verbose output
go test -cover ./...   # With coverage
```

## Test File Organization

**Location:**
- No test files exist. Go convention (and the expected pattern for this codebase) is co-located `_test.go` files next to source:
  - `main_test.go` alongside `main.go`

**Naming:**
- Go convention: `<source_file>_test.go`, test functions `Test<FunctionName>(t *testing.T)`

**Structure (expected):**
```
/home/hendry/ai-siem-endpoint/
├── main.go
└── main_test.go   # Does not exist yet
```

## Test Structure

**Suite Organization:**
- No existing tests. Go uses `testing.T` and table-driven tests as idiomatic pattern.
- Expected pattern for this codebase:
  ```go
  func TestHandlePost(t *testing.T) {
      tests := []struct {
          name   string
          body   string
          status int
      }{
          {"valid run", `{"run_id":"abc",...}`, 201},
          {"missing run_id", `{}`, 400},
          {"bad JSON", `not-json`, 400},
      }
      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) {
              // ...
          })
      }
  }
  ```

## Mocking

**Framework:**
- No mocking framework is present or configured.
- The `dynamoClient` global variable in `main.go` is a concrete `*dynamodb.Client` — not an interface — which makes it difficult to mock without refactoring.

**What to Mock (when tests are added):**
- `dynamoClient`: Extract a DynamoDB interface covering `PutItem`, `Query`, `GetItem`; inject via constructor or test setup
- AWS config loading (`newDynamoClient`): Use `aws.Config` with a custom endpoint or mock transport

**What NOT to Mock:**
- HTTP routing — use `net/http/httptest.NewRecorder()` and real `http.Request` values
- JSON marshal/unmarshal — test with real Go values

## Fixtures and Factories

**Test Data:**
- No fixtures defined. Expected pattern when added:
  ```go
  func newTestRun() AuditRun {
      return AuditRun{
          RunID:     "test-run-1",
          Timestamp: time.Now().UTC(),
          Host:      "test-host",
          User:      "test-user",
          Mode:      "audit",
          Findings:  []Finding{},
          Score:     100,
      }
  }
  ```

**Location:**
- Co-locate in `main_test.go` or a `testdata/` subdirectory for JSON payloads

## Coverage

**Requirements:** None enforced (no CI pipeline or coverage threshold configured)

**View Coverage:**
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Test Types

**Unit Tests:**
- None exist. High-value targets: `putEvent`, `listEvents`, `getEvent`, `handlePost` (validation path), `handleDetail` (base64 decoding path)

**Integration Tests:**
- None exist. Would require a local DynamoDB (e.g., `amazon/dynamodb-local` Docker image) or AWS `sam local` with DynamoDB Local

**E2E Tests:**
- Not used. The `mock-backend` binary (committed to repo) is a pre-built ARM64 binary used for local end-to-end testing, not an automated test suite

## Testability Constraints

**Global state barrier:**
- `dynamoClient` and `tableName` are set in `init()` and are package-level globals (`main.go:70-73`). This prevents dependency injection without refactoring.
- To make handlers testable, extract a handler struct:
  ```go
  type App struct {
      db        DynamoQuerier
      tableName string
  }
  ```

**Handler registration:**
- Routes are registered directly on the default `http.DefaultServeMux` in `main()`. Extracting to a function returning an `http.Handler` would enable `httptest`-based tests.

## Common Patterns (Expected)

**Async Testing:**
```go
// Not applicable — all handlers are synchronous
```

**HTTP Handler Testing:**
```go
func TestHandlePost(t *testing.T) {
    rr := httptest.NewRecorder()
    body := strings.NewReader(`{"run_id":"abc","timestamp":"2026-01-01T00:00:00Z"}`)
    req := httptest.NewRequest(http.MethodPost, "/", body)
    handlePost(rr, req)
    if rr.Code != http.StatusCreated {
        t.Errorf("expected 201, got %d", rr.Code)
    }
}
```

**Error Path Testing:**
```go
func TestHandlePost_MissingRunID(t *testing.T) {
    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
    handlePost(rr, req)
    if rr.Code != http.StatusBadRequest {
        t.Errorf("expected 400, got %d", rr.Code)
    }
}
```

---

*Testing analysis: 2026-05-18*
