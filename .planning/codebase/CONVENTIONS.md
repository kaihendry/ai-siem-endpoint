# Coding Conventions

**Analysis Date:** 2026-05-18

## Naming Patterns

**Files:**
- Single `main.go` entry point — everything in `package main`
- No multi-file package splitting; all logic is co-located in one file

**Types (structs):**
- PascalCase: `AuditRun`, `Finding`, `SummaryRow`, `summaryRowView`, `summaryTemplateData`, `detailTemplateData`
- Exported types use PascalCase; unexported template data types use camelCase prefix: `summaryTemplateData`, `detailTemplateData`

**Functions:**
- camelCase for unexported: `handlePost`, `handleGet`, `handleDetail`, `putEvent`, `listEvents`, `getEvent`, `writeJSON`, `newDynamoClient`
- Handler functions follow the `handle{Verb}` convention: `handlePost`, `handleGet`, `handleDetail`
- Helper/data-access functions are named after the operation: `putEvent`, `listEvents`, `getEvent`

**Variables:**
- camelCase: `tableName`, `dynamoClient`, `summaryTemplate`, `detailTemplate`
- Short locals inside closures: `strAttr`, `numAttr` (inline helper funcs)
- Loop variables: `row`, `item`, `i`

**Constants:**
- camelCase for unexported template string consts: `summaryTmpl`, `detailTmpl`

**JSON Tags:**
- All struct fields use snake_case JSON tags matching the DynamoDB attribute names: `run_id`, `schema_version`, `duration_ms`
- Optional fields use `omitempty`: `json:"confidence,omitempty"`

## Code Style

**Formatting:**
- Standard `gofmt` formatting (enforced by Go toolchain)
- No `.prettierrc` or custom formatter config present

**Linting:**
- No `.golangci.yml` or explicit linter config; project relies on `go vet` defaults

**Import Organization:**
- Standard Go convention: stdlib first, then third-party, separated by a blank line
- Example from `main.go`:
  ```go
  import (
      "context"
      "encoding/base64"
      // ... stdlib
  
      "github.com/apex/gateway/v2"
      "github.com/aws/aws-sdk-go-v2/aws"
      // ... third-party
  )
  ```

## Import Organization

**Order:**
1. Standard library packages
2. Third-party packages (blank line separator)

**Path Aliases:**
- None used; all imports use full module paths

## Error Handling

**Patterns:**
- Errors are returned up the call stack using `fmt.Errorf("context: %w", err)` for wrapping
- At handler boundaries, errors are logged with `slog.Error(...)` and an HTTP error response is written — never panic
- On fatal startup errors (config load failure), `os.Exit(1)` is called after logging
- Ignored errors are explicit: `run.Timestamp, _ = time.Parse(time.RFC3339, ts)` (parse errors silently default to zero time)
- Template execution errors after `w.Header()` is written are logged but not recoverable (standard Go http pattern)

**Error Response Format:**
```go
writeJSON(w, http.StatusBadRequest, map[string]string{"error": "message"})
```
- JSON errors always use the key `"error"` with a string value
- 400 for client validation failures, 500 for storage/internal failures

## Logging

**Framework:** `log/slog` (stdlib, Go 1.21+)

**Patterns:**
- Lambda environment: JSON handler to stderr — `slog.NewJSONHandler(os.Stderr, nil)`
- Local development: default text handler
- All log calls use structured key-value pairs: `slog.Error("putEvent failed", "err", err)`
- Log keys: `"err"`, `"addr"` — lowercase, short, consistent

## Comments

**When to Comment:**
- Task-tracking comments reference spec ticket numbers: `// T006: domain types matching internal/audit/AuditRun`
- Comments describe the purpose of a block, not individual lines
- No JSDoc-style godoc on functions; functions are short enough to be self-documenting

**Pattern:**
```go
// T010–T012: POST / handler
func handlePost(w http.ResponseWriter, r *http.Request) {
```

## Function Design

**Size:** Functions are small and single-purpose; longest function (`listEvents`) is ~35 lines
**Parameters:** HTTP handlers use the standard `(w http.ResponseWriter, r *http.Request)` signature; data-access functions take `ctx context.Context` as first parameter
**Return Values:** Data-access functions return `(value, error)`; detail handler returns `(*AuditRun, string, error)` where the string is `userAgent`

## Module Design

**Package:** Single `package main` — no sub-packages
**Exports:** All domain types (`AuditRun`, `Finding`, `SummaryRow`) are exported; helper types scoped to template rendering are unexported (`summaryTemplateData`, `summaryRowView`, `detailTemplateData`)
**Global State:** Two package-level vars initialized in `init()`: `dynamoClient *dynamodb.Client` and `tableName string`. Template vars (`summaryTemplate`, `detailTemplate`) are package-level `var` initialized with `template.Must(...)`
**Barrel Files:** Not applicable (single file)

## Environment Configuration

**Pattern:**
- Environment variable lookup with defaults:
  ```go
  tableName = os.Getenv("DYNAMODB_TABLE")
  if tableName == "" {
      tableName = "mock-siem-events"
  }
  ```
- Runtime mode detection via `os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != ""`

## DynamoDB Attribute Access

**Pattern:** Type-assertion with ok-check; no helper library:
```go
if v, ok := item["sk"].(*types.AttributeValueMemberS); ok {
    row.SK = v.Value
}
```
Inline helper closures (`strAttr`, `numAttr`) are used in `getEvent` to reduce repetition.

---

*Convention analysis: 2026-05-18*
