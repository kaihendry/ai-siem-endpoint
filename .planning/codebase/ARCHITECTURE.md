<!-- refreshed: 2026-05-18 -->
# Architecture

**Analysis Date:** 2026-05-18

## System Overview

```text
┌──────────────────────────────────────────────────────────────┐
│                    HTTP Entry Point                          │
│   POST /          GET /          GET /event/{sk}             │
│   `main.go:90`    `main.go:92`   `main.go:91`                │
└──────────┬────────────┬─────────────────┬────────────────────┘
           │            │                 │
           ▼            ▼                 ▼
┌──────────────┐ ┌──────────────┐ ┌────────────────────────┐
│ handlePost   │ │ handleGet    │ │ handleDetail            │
│ `main.go:112`│ │ `main.go:286`│ │ `main.go:439`           │
└──────┬───────┘ └──────┬───────┘ └──────────┬─────────────┘
       │                │                     │
       ▼                ▼                     ▼
┌──────────────┐ ┌──────────────┐ ┌────────────────────────┐
│  putEvent    │ │  listEvents  │ │  getEvent               │
│ `main.go:133`│ │ `main.go:235`│ │ `main.go:384`           │
└──────┬───────┘ └──────┬───────┘ └──────────┬─────────────┘
       │                │                     │
       ▼                ▼                     ▼
┌──────────────────────────────────────────────────────────────┐
│           AWS DynamoDB (EventsTable)                         │
│   pk="all" (fixed)    sk=<RFC3339-timestamp>#<run_id>        │
└──────────────────────────────────────────────────────────────┘
```

## Component Responsibilities

| Component | Responsibility | File |
|-----------|----------------|------|
| `main()` | Dual-mode entry: AWS Lambda via apex/gateway or local HTTP server | `main.go:85` |
| `init()` | Bootstrap DynamoDB client and table name from env | `main.go:22` |
| `handlePost` | Decode and validate `AuditRun` JSON, delegate to putEvent, return JSON | `main.go:112` |
| `handleGet` | Query last 50 events, render HTML summary table | `main.go:286` |
| `handleDetail` | Decode base64 SK, fetch single event, render HTML detail view | `main.go:439` |
| `putEvent` | Serialize `AuditRun` into DynamoDB PutItem | `main.go:133` |
| `listEvents` | DynamoDB Query on pk="all", reverse-sorted, limit 50 | `main.go:235` |
| `getEvent` | DynamoDB GetItem by pk/sk, deserialize into `AuditRun` | `main.go:384` |
| `writeJSON` | Set Content-Type and encode JSON response | `main.go:469` |

## Pattern Overview

**Overall:** Single-file Go HTTP server, dual-mode (Lambda/local), thin handler-to-storage pattern

**Key Characteristics:**
- No routing framework — uses Go 1.22+ method+path pattern syntax directly on `http.DefaultServeMux`
- No service layer — handlers call DynamoDB functions directly
- All state is in DynamoDB; no in-process mutable state beyond the DynamoDB client singleton
- HTML rendered server-side via `html/template` with inline template strings
- Lambda adapter via `github.com/apex/gateway/v2` wraps `http.DefaultServeMux` transparently

## Layers

**HTTP Handlers (presentation):**
- Purpose: Parse requests, call storage functions, render responses (JSON or HTML)
- Location: `main.go:112`, `main.go:286`, `main.go:439`
- Contains: Request parsing, base64 encoding/decoding of SK, template execution
- Depends on: Storage functions, template variables
- Used by: HTTP router (DefaultServeMux)

**Storage Functions:**
- Purpose: Translate domain types to/from DynamoDB attribute maps
- Location: `main.go:133`, `main.go:235`, `main.go:384`
- Contains: DynamoDB SDK calls, attribute map construction, field extraction helpers
- Depends on: `dynamoClient` singleton, `tableName` env var
- Used by: HTTP handlers

**Domain Types:**
- Purpose: Typed representation of audit run payloads matching the upstream `ai-check-guardrails` schema
- Location: `main.go:32` (`Finding`), `main.go:42` (`AuditRun`), `main.go:57` (`SummaryRow`)
- Contains: JSON-tagged structs for wire format
- Depends on: Nothing (pure types)
- Used by: All handlers and storage functions

**Templates:**
- Purpose: Server-side HTML rendering for the web UI
- Location: `main.go:168` (`summaryTmpl`), `main.go:310` (`detailTmpl`)
- Contains: Inline Go template strings with `html/template` func maps
- Depends on: `summaryTemplateData`, `detailTemplateData` view structs
- Used by: `handleGet`, `handleDetail`

**Infrastructure / Bootstrap:**
- Purpose: AWS config load, DynamoDB client construction, environment wiring
- Location: `main.go:22` (`init`), `main.go:75` (`newDynamoClient`), `main.go:85` (`main`)
- Contains: Mode detection (`AWS_LAMBDA_FUNCTION_NAME`), structured logging setup, port binding
- Depends on: `aws-sdk-go-v2/config`, `apex/gateway/v2`
- Used by: Go runtime

## Data Flow

### POST / — Ingest Audit Run

1. HTTP POST body decoded to `AuditRun` struct (`main.go:113`)
2. `run_id` presence validated (`main.go:118`)
3. `putEvent` constructs SK as `<RFC3339>#+<run_id>` and builds DynamoDB attribute map (`main.go:133`)
4. Findings serialized to JSON string for storage in a single DynamoDB string attribute (`main.go:136`)
5. DynamoDB `PutItem` called (`main.go:159`)
6. `{"run_id": ..., "sk": ...}` JSON returned with 201 (`main.go:130`)

### GET / — Summary Dashboard

1. `listEvents` issues DynamoDB `Query` on `pk="all"`, reverse sorted, limit 50 (`main.go:235`)
2. Items unpacked into `[]SummaryRow` via attribute type-assertion helpers (`main.go:249`)
3. Each row gets a base64-URL-encoded SK for the detail link (`main.go:297`)
4. `summaryTemplate` executed to HTML response (`main.go:303`)

### GET /event/{sk} — Event Detail

1. URL path suffix extracted and base64-URL-decoded to recover raw SK (`main.go:440`)
2. `getEvent` calls DynamoDB `GetItem` by `pk="all"` + decoded SK (`main.go:385`)
3. Findings JSON string unmarshalled back into `[]Finding` (`main.go:430`)
4. `detailTemplate` executed with `detailTemplateData` (`main.go:464`)

**State Management:**
- No in-process state beyond the `dynamoClient` singleton (module-level var, initialized in `init`)
- All persistent state lives in DynamoDB table `EventsTable`
- DynamoDB key design: single-partition (`pk="all"`) with composite range key `timestamp#run_id` for reverse-chron ordering

## Key Abstractions

**AuditRun:**
- Purpose: Wire format for a single endpoint audit execution result, matching the upstream `ai-check-guardrails` domain schema
- Examples: `main.go:42`
- Pattern: JSON-tagged struct, decoded directly from request body

**Finding:**
- Purpose: Individual security finding within an `AuditRun`
- Examples: `main.go:32`
- Pattern: JSON-tagged struct, stored as JSON blob within the DynamoDB `findings` attribute

**SummaryRow / summaryRowView:**
- Purpose: Flattened view type for the HTML summary table; `summaryRowView` adds `SKEncoded` for safe URL embedding
- Examples: `main.go:57`, `main.go:217`
- Pattern: Read-model separation from the write/storage model

## Entry Points

**Lambda Entry:**
- Location: `main.go:94` (`gateway.ListenAndServe`)
- Triggers: AWS API Gateway HTTP API event via Lambda invocation
- Responsibilities: Adapts Lambda payload to `net/http` request/response, then delegates to registered handlers

**Local HTTP Server Entry:**
- Location: `main.go:104` (`http.ListenAndServe`)
- Triggers: Direct process execution without `AWS_LAMBDA_FUNCTION_NAME` set
- Responsibilities: Standard Go HTTP server on `PORT` env var (default 8080)

## Architectural Constraints

- **Threading:** Go standard library HTTP server; each request handled in its own goroutine. DynamoDB client is safe for concurrent use.
- **Global state:** `dynamoClient` and `tableName` are module-level vars initialized once in `init()` (`main.go:70`)
- **Circular imports:** Not applicable — single-package, single-file codebase
- **Partition key design:** All events share `pk="all"`. This is intentional for a mock/dev tool but limits horizontal DynamoDB scalability
- **Findings storage:** Findings are stored as a JSON-encoded string blob inside a DynamoDB String attribute, not as a native List attribute. Queries cannot filter on individual findings.
- **Lambda runtime:** `provided.al2` custom runtime; binary compiled as `bootstrap` for arm64

## Anti-Patterns

### Findings stored as JSON blob

**What happens:** `run.Findings` is marshalled to a JSON string and stored in a single DynamoDB String attribute (`main.go:136`).
**Why it's wrong:** Makes server-side filtering or querying by finding severity/type impossible without fetching and deserializing the full item.
**Do this instead:** Store findings as a DynamoDB List of Maps, or use a separate `findings` table with pk=run_id for queryability.

### Single partition key for all records

**What happens:** Every item is written with `pk="all"` (`main.go:142`).
**Why it's wrong:** All reads and writes hit a single DynamoDB partition, which is a hot-partition bottleneck at scale.
**Do this instead:** For a production system, partition by host, date bucket, or account. For this mock tool, the current design is acceptable.

## Error Handling

**Strategy:** Log and return HTTP error. No panics, no custom error types.

**Patterns:**
- Storage errors logged with `slog.Error` then mapped to 500 (`main.go:126`, `main.go:289`, `main.go:455`)
- Validation errors return 400 with JSON body via `writeJSON` (`main.go:115`, `main.go:120`)
- Template execution errors logged but response headers already sent — silent client failure (`main.go:304`, `main.go:465`)
- `init()` calls `os.Exit(1)` on AWS config load failure (`main.go:79`)

## Cross-Cutting Concerns

**Logging:** `log/slog` structured logger; JSON format in Lambda (`slog.NewJSONHandler`), default text format locally (`main.go:87`)
**Validation:** Minimal — only `run_id` presence is validated; no schema enforcement beyond JSON decode
**Authentication:** None — the endpoint is unauthenticated; intended for internal/dev use only

---

*Architecture analysis: 2026-05-18*
