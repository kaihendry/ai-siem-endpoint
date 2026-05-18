# Codebase Structure

**Analysis Date:** 2026-05-18

## Directory Layout

```
ai-siem-endpoint/
├── main.go          # Entire application: types, handlers, storage, templates, entry point
├── go.mod           # Module declaration and direct dependencies
├── go.sum           # Dependency checksums
├── Makefile         # Build, deploy, destroy, local-run targets
├── template.yml     # AWS SAM infrastructure definition (Lambda + API GW + DynamoDB)
├── mock-backend     # Pre-built ARM64 Linux binary (gitignored, local artifact)
├── .aws-sam/        # SAM build output directory (gitignored)
│   ├── build.toml
│   └── build/
│       ├── template.yaml   # SAM-processed CloudFormation template
│       └── MainFunction/
│           └── bootstrap   # Compiled Lambda bootstrap binary
├── .claude/         # Claude agent definitions and GSD tooling (not application code)
└── .planning/       # GSD planning documents
    └── codebase/    # Codebase analysis documents (this directory)
```

## Directory Purposes

**Root (`/`):**
- Purpose: Entire Go application lives here as a single package (`package main`)
- Contains: One `.go` source file, Go module files, Makefile, SAM template
- Key files: `main.go`, `template.yml`, `Makefile`, `go.mod`

**`.aws-sam/` (gitignored):**
- Purpose: SAM CLI build artifacts — processed template and compiled binary
- Contains: `build/MainFunction/bootstrap` (the Lambda deployment binary), `build/template.yaml`
- Generated: Yes
- Committed: No

**`.claude/` (not application code):**
- Purpose: Claude agent definitions and GSD (Get Shit Done) workflow tooling
- Contains: Agent markdown files, slash command definitions, workflow templates
- Generated: No (managed separately)
- Committed: Yes

**`.planning/codebase/` (this directory):**
- Purpose: Architecture and convention documents for AI-assisted development
- Contains: ARCHITECTURE.md, STRUCTURE.md, and other codebase analysis docs
- Generated: Yes (by GSD map-codebase)
- Committed: Yes

## Key File Locations

**Entry Points:**
- `main.go:85`: `main()` — dual-mode entry (Lambda vs local HTTP server)
- `main.go:22`: `init()` — DynamoDB client bootstrap

**Infrastructure Definition:**
- `template.yml`: SAM template — Lambda function, HTTP API, DynamoDB table, IAM policies
- `Makefile`: `build-MainFunction` target used by SAM; `deploy`/`destroy` for lifecycle management

**Core Logic:**
- `main.go:112`: `handlePost` — POST / ingest handler
- `main.go:286`: `handleGet` — GET / summary UI handler
- `main.go:439`: `handleDetail` — GET /event/{sk} detail UI handler
- `main.go:133`: `putEvent` — DynamoDB write
- `main.go:235`: `listEvents` — DynamoDB query
- `main.go:384`: `getEvent` — DynamoDB point read

**Domain Types:**
- `main.go:32`: `Finding` struct
- `main.go:42`: `AuditRun` struct
- `main.go:57`: `SummaryRow` struct

**Templates (inline strings):**
- `main.go:168`: `summaryTmpl` — HTML summary page
- `main.go:310`: `detailTmpl` — HTML event detail page

**Testing:**
- No test files present (`*_test.go` not found)

## Naming Conventions

**Files:**
- Single file: `main.go` — no multi-file convention to observe
- Build output: `bootstrap` (required name for Lambda custom runtime)

**Functions:**
- HTTP handlers: `handle<Noun>` (e.g., `handlePost`, `handleGet`, `handleDetail`)
- Storage functions: verb+noun, no prefix (e.g., `putEvent`, `listEvents`, `getEvent`)
- Helpers: descriptive camelCase (e.g., `newDynamoClient`, `writeJSON`)

**Types:**
- Domain structs: PascalCase matching upstream schema names (e.g., `AuditRun`, `Finding`)
- View types: PascalCase with `View` suffix for template-specific shapes (e.g., `summaryRowView`, `detailTemplateData`)
- Template data: `<name>TemplateData` (e.g., `summaryTemplateData`, `detailTemplateData`)

**Variables:**
- Module-level singletons: camelCase (e.g., `dynamoClient`, `tableName`)
- Pre-compiled templates: camelCase with `Template` suffix (e.g., `summaryTemplate`, `detailTemplate`)

**DynamoDB Keys:**
- Partition key: `pk` (always literal `"all"`)
- Sort key: `sk` — composite `<RFC3339-timestamp>#<run_id>`
- Attribute names: snake_case matching JSON field names

## Where to Add New Code

**New HTTP endpoint:**
- Register handler in `main()` at `main.go:90` using Go 1.22 method+path syntax
- Implement handler function in `main.go` following `handle<Noun>` naming
- Add DynamoDB access function if needed, following `verb+noun` pattern

**New domain field on AuditRun:**
- Add JSON-tagged field to `AuditRun` struct (`main.go:42`)
- Add corresponding DynamoDB attribute in `putEvent` (`main.go:141`)
- Add extraction in `getEvent` (`main.go:399`) and `listEvents` (`main.go:249`) as needed
- Update HTML templates if the field should be displayed

**New HTML template:**
- Define template string as a `const` following `summaryTmpl`/`detailTmpl` pattern
- Compile with `template.Must(template.New(...).Funcs(...).Parse(...))` at package level
- Create a `<name>TemplateData` struct for the template's data

**Utilities:**
- Add to `main.go` — no separate utility files; the codebase is intentionally single-file

## Special Directories

**`.aws-sam/`:**
- Purpose: SAM CLI build cache and processed CloudFormation artifacts
- Generated: Yes, by `sam build`
- Committed: No (in `.gitignore`)

**`.planning/`:**
- Purpose: GSD workflow planning and codebase analysis documents
- Generated: Partially (map-codebase writes here; human planners also write here)
- Committed: Yes

---

*Structure analysis: 2026-05-18*
