<!-- GSD:project-start source:PROJECT.md -->
## Project

**ai-siem-endpoint**

A Go HTTP service deployed on AWS Lambda that acts as a shared SIEM (Security Information and Event Management) backend for AI guardrail audit tools. Multiple reporting user agents (from `ai-check-guardrails`) POST their `AuditRun` results here; the service stores them in DynamoDB and provides a web dashboard to review findings. The endpoint needs to be reliably deployed via CI/CD so any UA can always report to a live, versioned URL.

**Core Value:** A stable, publicly reachable endpoint that any `ai-check-guardrails` UA can POST audit results to and trust will store and display them correctly.

### Constraints

- **Tech stack**: Go + AWS SAM + DynamoDB — stay with this stack, no new services
- **CI platform**: GitHub Actions only
- **AWS region**: eu-west-2 (hard requirement from existing infrastructure)
- **Lambda arch**: arm64 / provided.al2023 (upgrade from al2, keep arm64)
<!-- GSD:project-end -->

<!-- GSD:stack-start source:codebase/STACK.md -->
## Technology Stack

## Languages
- Go 1.26 - All application logic (`main.go`)
- YAML - SAM/CloudFormation infrastructure definition (`template.yml`)
- Makefile - Build and deployment automation (`Makefile`)
## Runtime
- AWS Lambda custom runtime (`provided.al2`) on arm64
- Local development: Go standard HTTP server on port 8080 (default)
- Go modules (`go.mod`, `go.sum`)
- Lockfile: `go.sum` present
## Frameworks
- `net/http` (stdlib) - HTTP routing and handler registration
- `github.com/apex/gateway/v2` v2.0.0 - Bridges `net/http` to AWS Lambda runtime
- AWS SAM CLI - Build and deploy tooling (`Makefile` targets: `build-MainFunction`, `deploy`, `destroy`)
- `go build` - Cross-compiled for `GOARCH=arm64 GOOS=linux`, output binary named `bootstrap`
- Not detected — no test files (`*_test.go`) present
## Key Dependencies
- `github.com/apex/gateway/v2` v2.0.0 - Enables dual-mode operation: same `net/http` handlers run locally or on Lambda
- `github.com/aws/aws-sdk-go-v2` v1.41.7 - AWS SDK core
- `github.com/aws/aws-sdk-go-v2/config` v1.32.17 - AWS credential/config loading (`config.LoadDefaultConfig`)
- `github.com/aws/aws-sdk-go-v2/service/dynamodb` v1.57.3 - DynamoDB client for event persistence
- `github.com/aws/aws-lambda-go` v1.54.0 - Lambda runtime interface (pulled in by apex/gateway)
- `github.com/aws/smithy-go` v1.25.1 - AWS SDK transport layer
- `github.com/pkg/errors` v0.9.1 - Error wrapping (transitive)
## Configuration
- `DYNAMODB_TABLE` - DynamoDB table name (defaults to `mock-siem-events` if unset); injected via SAM `Environment.Variables` in `template.yml`
- `AWS_LAMBDA_FUNCTION_NAME` - Presence determines runtime mode (Lambda vs local HTTP server)
- `PORT` - Local HTTP listen port (defaults to `8080`)
- AWS credentials loaded via `config.LoadDefaultConfig` (supports env vars, shared credentials, EC2 IMDS, SSO)
- `template.yml` - SAM/CloudFormation template defining Lambda function, HTTP API, and DynamoDB table
- `Makefile` - Build method used by SAM (`BuildMethod: makefile`); cross-compiles to `arm64` Linux
## Platform Requirements
- Go 1.26+
- AWS SAM CLI (for `sam build` / `sam deploy`)
- AWS credentials configured (SSO profile: `AdministratorAccess-407461997746`)
- AWS Lambda (arm64, `provided.al2` runtime)
- AWS API Gateway HTTP API
- AWS DynamoDB (PAY_PER_REQUEST billing)
- Target region: `eu-west-2`
- CloudFormation stack name: `mock-siem-backend`
<!-- GSD:stack-end -->

<!-- GSD:conventions-start source:CONVENTIONS.md -->
## Conventions

## Naming Patterns
- Single `main.go` entry point — everything in `package main`
- No multi-file package splitting; all logic is co-located in one file
- PascalCase: `AuditRun`, `Finding`, `SummaryRow`, `summaryRowView`, `summaryTemplateData`, `detailTemplateData`
- Exported types use PascalCase; unexported template data types use camelCase prefix: `summaryTemplateData`, `detailTemplateData`
- camelCase for unexported: `handlePost`, `handleGet`, `handleDetail`, `putEvent`, `listEvents`, `getEvent`, `writeJSON`, `newDynamoClient`
- Handler functions follow the `handle{Verb}` convention: `handlePost`, `handleGet`, `handleDetail`
- Helper/data-access functions are named after the operation: `putEvent`, `listEvents`, `getEvent`
- camelCase: `tableName`, `dynamoClient`, `summaryTemplate`, `detailTemplate`
- Short locals inside closures: `strAttr`, `numAttr` (inline helper funcs)
- Loop variables: `row`, `item`, `i`
- camelCase for unexported template string consts: `summaryTmpl`, `detailTmpl`
- All struct fields use snake_case JSON tags matching the DynamoDB attribute names: `run_id`, `schema_version`, `duration_ms`
- Optional fields use `omitempty`: `json:"confidence,omitempty"`
## Code Style
- Standard `gofmt` formatting (enforced by Go toolchain)
- No `.prettierrc` or custom formatter config present
- No `.golangci.yml` or explicit linter config; project relies on `go vet` defaults
- Standard Go convention: stdlib first, then third-party, separated by a blank line
- Example from `main.go`:
## Import Organization
- None used; all imports use full module paths
## Error Handling
- Errors are returned up the call stack using `fmt.Errorf("context: %w", err)` for wrapping
- At handler boundaries, errors are logged with `slog.Error(...)` and an HTTP error response is written — never panic
- On fatal startup errors (config load failure), `os.Exit(1)` is called after logging
- Ignored errors are explicit: `run.Timestamp, _ = time.Parse(time.RFC3339, ts)` (parse errors silently default to zero time)
- Template execution errors after `w.Header()` is written are logged but not recoverable (standard Go http pattern)
- JSON errors always use the key `"error"` with a string value
- 400 for client validation failures, 500 for storage/internal failures
## Logging
- Lambda environment: JSON handler to stderr — `slog.NewJSONHandler(os.Stderr, nil)`
- Local development: default text handler
- All log calls use structured key-value pairs: `slog.Error("putEvent failed", "err", err)`
- Log keys: `"err"`, `"addr"` — lowercase, short, consistent
## Comments
- Task-tracking comments reference spec ticket numbers: `// T006: domain types matching internal/audit/AuditRun`
- Comments describe the purpose of a block, not individual lines
- No JSDoc-style godoc on functions; functions are short enough to be self-documenting
## Function Design
## Module Design
## Environment Configuration
- Environment variable lookup with defaults:
- Runtime mode detection via `os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != ""`
## DynamoDB Attribute Access
<!-- GSD:conventions-end -->

<!-- GSD:architecture-start source:ARCHITECTURE.md -->
## Architecture

## System Overview
```text
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
- No routing framework — uses Go 1.22+ method+path pattern syntax directly on `http.DefaultServeMux`
- No service layer — handlers call DynamoDB functions directly
- All state is in DynamoDB; no in-process mutable state beyond the DynamoDB client singleton
- HTML rendered server-side via `html/template` with inline template strings
- Lambda adapter via `github.com/apex/gateway/v2` wraps `http.DefaultServeMux` transparently
## Layers
- Purpose: Parse requests, call storage functions, render responses (JSON or HTML)
- Location: `main.go:112`, `main.go:286`, `main.go:439`
- Contains: Request parsing, base64 encoding/decoding of SK, template execution
- Depends on: Storage functions, template variables
- Used by: HTTP router (DefaultServeMux)
- Purpose: Translate domain types to/from DynamoDB attribute maps
- Location: `main.go:133`, `main.go:235`, `main.go:384`
- Contains: DynamoDB SDK calls, attribute map construction, field extraction helpers
- Depends on: `dynamoClient` singleton, `tableName` env var
- Used by: HTTP handlers
- Purpose: Typed representation of audit run payloads matching the upstream `ai-check-guardrails` schema
- Location: `main.go:32` (`Finding`), `main.go:42` (`AuditRun`), `main.go:57` (`SummaryRow`)
- Contains: JSON-tagged structs for wire format
- Depends on: Nothing (pure types)
- Used by: All handlers and storage functions
- Purpose: Server-side HTML rendering for the web UI
- Location: `main.go:168` (`summaryTmpl`), `main.go:310` (`detailTmpl`)
- Contains: Inline Go template strings with `html/template` func maps
- Depends on: `summaryTemplateData`, `detailTemplateData` view structs
- Used by: `handleGet`, `handleDetail`
- Purpose: AWS config load, DynamoDB client construction, environment wiring
- Location: `main.go:22` (`init`), `main.go:75` (`newDynamoClient`), `main.go:85` (`main`)
- Contains: Mode detection (`AWS_LAMBDA_FUNCTION_NAME`), structured logging setup, port binding
- Depends on: `aws-sdk-go-v2/config`, `apex/gateway/v2`
- Used by: Go runtime
## Data Flow
### POST / — Ingest Audit Run
### GET / — Summary Dashboard
### GET /event/{sk} — Event Detail
- No in-process state beyond the `dynamoClient` singleton (module-level var, initialized in `init`)
- All persistent state lives in DynamoDB table `EventsTable`
- DynamoDB key design: single-partition (`pk="all"`) with composite range key `timestamp#run_id` for reverse-chron ordering
## Key Abstractions
- Purpose: Wire format for a single endpoint audit execution result, matching the upstream `ai-check-guardrails` domain schema
- Examples: `main.go:42`
- Pattern: JSON-tagged struct, decoded directly from request body
- Purpose: Individual security finding within an `AuditRun`
- Examples: `main.go:32`
- Pattern: JSON-tagged struct, stored as JSON blob within the DynamoDB `findings` attribute
- Purpose: Flattened view type for the HTML summary table; `summaryRowView` adds `SKEncoded` for safe URL embedding
- Examples: `main.go:57`, `main.go:217`
- Pattern: Read-model separation from the write/storage model
## Entry Points
- Location: `main.go:94` (`gateway.ListenAndServe`)
- Triggers: AWS API Gateway HTTP API event via Lambda invocation
- Responsibilities: Adapts Lambda payload to `net/http` request/response, then delegates to registered handlers
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
### Single partition key for all records
## Error Handling
- Storage errors logged with `slog.Error` then mapped to 500 (`main.go:126`, `main.go:289`, `main.go:455`)
- Validation errors return 400 with JSON body via `writeJSON` (`main.go:115`, `main.go:120`)
- Template execution errors logged but response headers already sent — silent client failure (`main.go:304`, `main.go:465`)
- `init()` calls `os.Exit(1)` on AWS config load failure (`main.go:79`)
## Cross-Cutting Concerns
<!-- GSD:architecture-end -->

<!-- GSD:skills-start source:skills/ -->
## Project Skills

No project skills found. Add skills to any of: `.claude/skills/`, `.agents/skills/`, `.cursor/skills/`, `.github/skills/`, or `.codex/skills/` with a `SKILL.md` index file.
<!-- GSD:skills-end -->

<!-- GSD:workflow-start source:GSD defaults -->
## GSD Workflow Enforcement

Before using Edit, Write, or other file-changing tools, start work through a GSD command so planning artifacts and execution context stay in sync.

Use these entry points:
- `/gsd-quick` for small fixes, doc updates, and ad-hoc tasks
- `/gsd-debug` for investigation and bug fixing
- `/gsd-execute-phase` for planned phase work

Do not make direct repo edits outside a GSD workflow unless the user explicitly asks to bypass it.
<!-- GSD:workflow-end -->



<!-- GSD:profile-start -->
## Developer Profile

> Profile not yet configured. Run `/gsd-profile-user` to generate your developer profile.
> This section is managed by `generate-claude-profile` -- do not edit manually.
<!-- GSD:profile-end -->
