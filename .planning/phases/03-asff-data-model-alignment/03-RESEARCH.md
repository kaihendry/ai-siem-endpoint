# Phase 3: ASFF Data Model Alignment - Research

**Researched:** 2026-05-18
**Domain:** AWS Security Finding Format (ASFF) struct alignment in Go / DynamoDB
**Confidence:** HIGH

---

## Summary

This phase adds ASFF-required fields to the existing `Finding` struct without replacing existing fields, enabling findings to round-trip through Security Hub or any ASFF-aware tooling without a translation layer. The ASFF spec has exactly 10 required top-level fields on `AwsSecurityFinding` (the finding container): `AwsAccountId`, `CreatedAt`, `Description`, `GeneratorId`, `Id`, `ProductArn`, `Resources`, `SchemaVersion`, `Title`, `UpdatedAt`. Seven of these have no analogue in the current `Finding` struct. Two (`Description`, `SchemaVersion`) exist but live at the wrong level or have name mismatches. `Severity` is optional at the top level per the API but is listed as a required field in the ASFF user guide requirements (ASFF-01 explicitly names `Severity.Label`).

The ASFF model describes **individual findings**, not batch runs. `AuditRun` is a proprietary batch container with no ASFF equivalent — it should be kept as-is. The right strategy per ASFF-02 is additive: new optional fields on `Finding`, not a rewrite of `AuditRun`. `ProductArn` and `AwsAccountId` are run-level identity fields best added to `AuditRun` so they can be stamped onto all findings at ingest time. `Resources` (min 1 item) must live inside each `Finding` since ASFF requires at least one resource per finding.

DynamoDB storage currently serialises the `[]Finding` slice as a JSON blob string. New fields on `Finding` are automatically included in that blob — no DynamoDB schema migration is needed. The only DynamoDB change needed is persisting two new `AuditRun` run-level fields (`product_arn`, `aws_account_id`) as top-level attributes alongside the existing ones.

**Primary recommendation:** Add ASFF fields as `omitempty` optional struct fields on `Finding` and two identity fields on `AuditRun`; persist the two new `AuditRun` fields as DynamoDB attributes; update `putEvent` and tests.

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| ASFF-01 | `Finding` struct maps to required ASFF fields: Id, ProductArn, GeneratorId, AwsAccountId, Types, CreatedAt, UpdatedAt, Severity.Label, Title, Description, Resources | Verified ASFF required fields from AWS API ref; mapping table below shows exact field disposition |
| ASFF-02 | `AuditRun` retains its existing fields for backwards compatibility; ASFF fields are added, not replacing | All new fields use `omitempty`; existing JSON tags unchanged |
| ASFF-03 | DynamoDB storage persists new ASFF fields | Findings blob automatically carries new fields; two new `AuditRun` fields need explicit DynamoDB attributes |
| ASFF-04 | POST / accepts and stores ASFF-extended payloads | No handler changes beyond putEvent DynamoDB attribute additions |
| ASFF-05 | Tests cover new ASFF fields in putEvent and handler | Extend TestPutEventAttributes + add ASFF round-trip test in TestHandlePost |
</phase_requirements>

---

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| ASFF field storage (Finding blob) | Database / Storage (`putEvent`) | — | Findings are JSON-marshalled before PutItem; new fields go along automatically |
| AuditRun identity fields (ProductArn, AwsAccountId) | API / Backend (`putEvent`) | — | Run-level ASFF identity must be persisted as explicit DynamoDB attributes |
| Request decoding of new fields | API / Backend (`handlePost` → `json.Decode`) | — | Go `json.Decoder` picks up new struct fields transparently |
| Backwards compatibility | API / Backend (struct tags) | — | `omitempty` ensures old clients sending neither field still decode cleanly |
| HTML display of new fields | Frontend Server (HTML templates) | — | Optional: detail template can render ASFF fields if desired; not required for ASFF-01–05 |

---

## ASFF Required Fields — Complete Reference

[CITED: docs.aws.amazon.com/securityhub/1.0/APIReference/API_AwsSecurityFinding.html]
[CITED: docs.aws.amazon.com/securityhub/latest/userguide/asff-required-attributes.md]

The 10 required fields on `AwsSecurityFinding`:

| ASFF Field | Type | Constraint | Required |
|------------|------|------------|----------|
| `AwsAccountId` | String | Length 12, pattern `.*\S.*` | Yes |
| `CreatedAt` | String (ISO 8601) | RFC 3339 format | Yes |
| `Description` | String | 1–1024 chars | Yes |
| `GeneratorId` | String | 1–512 chars | Yes |
| `Id` | String | 1–512 chars | Yes |
| `ProductArn` | String | 12–2048 chars, ARN format | Yes |
| `Resources` | Array of Resource | Min 1, max 32 items | Yes |
| `SchemaVersion` | String | Must be `"2018-10-08"` | Yes |
| `Title` | String | 1–256 chars | Yes |
| `UpdatedAt` | String (ISO 8601) | RFC 3339 format | Yes |

Additional field required by ASFF-01 (listed in ROADMAP):
| `Severity` | Object | Optional in API spec; `Label` or `Normalized` must be present if object is provided | Optional* |
| `Types` | []String | namespace/category/classifier format; max 50 items | Optional |

*Note: `Severity` is listed as optional at the API level but ASFF-01 specifically calls out `Severity.Label` as a required mapping. The finding will be valid without it, but ASFF-01 requires the struct to support it.

### Resource Object (within Resources array)

[CITED: docs.aws.amazon.com/securityhub/1.0/APIReference/API_Resource.html]

| Field | Type | Required |
|-------|------|----------|
| `Id` | String | Yes |
| `Type` | String (1–256 chars) | Yes |
| `Partition` | String (`aws`, `aws-cn`, `aws-us-gov`) | No |
| `Region` | String (1–16 chars) | No |
| `ResourceRole` | String | No |
| `Tags` | map[string]string | No |
| `Details` | ResourceDetails object | No |

### Severity Object

[CITED: docs.aws.amazon.com/securityhub/1.0/APIReference/API_Severity.html]

| Field | Type | Valid Values | Required |
|-------|------|-------------|----------|
| `Label` | String | `INFORMATIONAL`, `LOW`, `MEDIUM`, `HIGH`, `CRITICAL` | No (but preferred) |
| `Normalized` | Integer | 0–100 (deprecated) | No |
| `Original` | String | Any native severity string | No |

At least `Label` OR `Normalized` must be present if the `Severity` object is included.

### Remediation Object (optional)

[CITED: docs.aws.amazon.com/securityhub/1.0/APIReference/API_Remediation.html]
[CITED: docs.aws.amazon.com/securityhub/1.0/APIReference/API_Recommendation.html]

```
Remediation:
  Recommendation:
    Text: string (1–512 chars, optional)
    Url:  string (optional)
```

All fields optional. The existing `Finding.Remediation string` maps to `Remediation.Recommendation.Text`.

---

## Current Model Gap Analysis

### Finding struct — field-by-field mapping

| Current Field | JSON Tag | ASFF Equivalent | Action |
|---------------|----------|-----------------|--------|
| `Type` | `"type"` | `Types[0]` (partial) | Keep as-is; add new `ASFFTypes []string` field |
| `Severity` | `"severity"` | `Severity.Original` (partial) | Keep as-is; add new `ASFFSeverity *SeverityASFF` field |
| `Module` | `"module"` | `GeneratorId` (partial) | Keep; add `GeneratorId string` |
| `Resource` | `"resource"` | `Resources[0].Id` (partial) | Keep; add `Resources []ResourceASFF` |
| `Description` | `"description"` | `Description` | Direct match — no change |
| `Remediation` | `"remediation"` | `Remediation.Recommendation.Text` (partial) | Keep as plain string for compat; add `RemediationASFF *RemediationASFF` if full struct needed |
| `Confidence` | `"confidence,omitempty"` | `Confidence` (0–100 int at finding level) | Keep current float64 for compat (existing UA sends 0.0–1.0); add `ASFFConfidence *int` if needed |
| — | — | `Id` | Add `ID string` |
| — | — | `Title` | Add `Title string` |
| — | — | `CreatedAt` | Add `CreatedAt string` |
| — | — | `UpdatedAt` | Add `UpdatedAt string` |

### AuditRun struct — ASFF run-level identity

ASFF fields that describe the **product/account** rather than a specific finding belong on `AuditRun`:

| ASFF Field | Proposed AuditRun Field | Notes |
|------------|------------------------|-------|
| `ProductArn` | `ProductArn string` | ARN format; empty = not set by client |
| `AwsAccountId` | `AwsAccountId string` | 12-digit AWS account; empty = not set |

`SchemaVersion` already exists on `AuditRun` (field `SchemaVersion`, tag `"schema_version"`). For ASFF it must equal `"2018-10-08"` — clients that forward to Security Hub must set this value; the service stores whatever is sent.

---

## Standard Stack

No new packages required. All implementation uses:

| Library | Already Present | Purpose |
|---------|----------------|---------|
| `encoding/json` (stdlib) | Yes | Marshal/unmarshal new struct fields |
| `github.com/aws/aws-sdk-go-v2/service/dynamodb` | Yes | PutItem with new attributes |
| `net/http` (stdlib) | Yes | No handler changes |

**No new dependencies.** This phase is pure Go struct changes + DynamoDB attribute additions.

---

## Package Legitimacy Audit

No new packages are installed in this phase. Section not applicable.

---

## Architecture Patterns

### System Architecture Diagram

```
POST /  (existing flow, unchanged)
   |
   v
handlePost
   |
   +-- json.Decode(r.Body) --> AuditRun{
   |       existing fields...
   |       ProductArn string   (NEW - ASFF run identity)
   |       AwsAccountId string (NEW - ASFF run identity)
   |       Findings: []Finding{
   |           existing fields...
   |           ID string           (NEW)
   |           Title string        (NEW)
   |           GeneratorId string  (NEW)
   |           ASFFSeverity        (NEW - nested)
   |           Resources []ResourceASFF (NEW - nested)
   |           CreatedAt string    (NEW)
   |           UpdatedAt string    (NEW)
   |           ASFFTypes []string  (NEW)
   |       }
   |   }
   |
   v
putEvent
   |
   +-- json.Marshal(run.Findings) --> findings blob (includes new fields automatically)
   |
   +-- DynamoDB PutItem: add product_arn, aws_account_id attributes
   |
   v
DynamoDB item (existing + 2 new attributes)
```

### Recommended Struct Changes

```go
// New nested types — add to main.go after existing Finding struct

type SeverityASFF struct {
    Label      string  `json:"Label,omitempty"`
    Normalized *int    `json:"Normalized,omitempty"`
    Original   string  `json:"Original,omitempty"`
}

type ResourceASFF struct {
    ID   string `json:"Id"`
    Type string `json:"Type"`
    // optional: Partition, Region omitted for minimal compliance
}

// Updated Finding struct — all new fields are omitempty
type Finding struct {
    // Existing fields (unchanged for backwards compatibility)
    Type        string   `json:"type"`
    Severity    string   `json:"severity"`
    Module      string   `json:"module"`
    Resource    string   `json:"resource"`
    Description string   `json:"description"`
    Remediation string   `json:"remediation"`
    Confidence  *float64 `json:"confidence,omitempty"`

    // New ASFF fields (all omitempty — old payloads still decode)
    ID          string        `json:"Id,omitempty"`
    Title       string        `json:"title,omitempty"`
    GeneratorId string        `json:"generator_id,omitempty"`
    ASFFTypes   []string      `json:"asff_types,omitempty"`
    ASFFSeverity *SeverityASFF `json:"asff_severity,omitempty"`
    Resources   []ResourceASFF `json:"resources,omitempty"`
    CreatedAt   string        `json:"created_at,omitempty"`
    UpdatedAt   string        `json:"updated_at,omitempty"`
}

// Updated AuditRun struct — two new ASFF identity fields
type AuditRun struct {
    SchemaVersion string    `json:"schema_version"`
    RunID         string    `json:"run_id"`
    Timestamp     time.Time `json:"timestamp"`
    Host          string    `json:"host"`
    User          string    `json:"user"`
    Mode          string    `json:"mode"`
    Version       string    `json:"version"`
    Findings      []Finding `json:"findings"`
    Score         int       `json:"score"`
    ExitCode      int       `json:"exit_code"`
    DurationMs    int64     `json:"duration_ms"`

    // New ASFF run-level identity (omitempty — backwards compatible)
    ProductArn    string    `json:"product_arn,omitempty"`
    AwsAccountId  string    `json:"aws_account_id,omitempty"`
}
```

### putEvent DynamoDB changes

Add two attributes to the `item` map in `putEvent`:

```go
// In putEvent, add to the item map:
"product_arn":     &types.AttributeValueMemberS{Value: run.ProductArn},
"aws_account_id":  &types.AttributeValueMemberS{Value: run.AwsAccountId},
```

The `findings` blob already captures all `Finding` struct fields via `json.Marshal`, so new `Finding` fields require zero DynamoDB schema changes.

### getEvent deserialization

`getEvent` already unmarshals the `findings` blob into `[]Finding` — new fields are picked up automatically by `json.Unmarshal`. Two new `AuditRun` fields need explicit reads:

```go
run.ProductArn   = strAttr("product_arn")
run.AwsAccountId = strAttr("aws_account_id")
```

### Anti-Patterns to Avoid

- **Replacing existing Finding fields with ASFF names:** Breaks every existing UA that POSTs with the old schema. Use `omitempty` additive fields.
- **Storing ASFF fields as separate DynamoDB top-level attributes:** The findings blob pattern is already established; stay consistent. Only run-level fields (ProductArn, AwsAccountId) warrant top-level DynamoDB attributes.
- **Requiring ASFF fields in handlePost validation:** ASFF-02 explicitly requires old payloads without ASFF fields to still be accepted. Do not add `if run.Findings[i].ID == ""` validation.
- **Changing SchemaVersion semantics:** The existing `schema_version` field on `AuditRun` is the tool's own version, not ASFF's fixed `"2018-10-08"`. Do not rename or repurpose it.
- **Using ASFF Severity.Normalized:** It is deprecated per the API spec. Use `Severity.Label` only.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| ASFF JSON serialization | Custom marshal logic | Standard `encoding/json` with struct tags | Go's json package handles nested structs, omitempty, and pointer fields natively |
| Severity label validation | Custom enum check | Accept any string, no validation | ASFF-02 requires backwards compat; ASFF UAs are trusted clients |
| DynamoDB attribute type conversion | Custom type system | Existing `AttributeValueMemberS` / `AttributeValueMemberN` pattern | Already established in putEvent; stay consistent |

---

## Common Pitfalls

### Pitfall 1: ASFF Severity vs. existing Severity field naming conflict
**What goes wrong:** Both the existing `Severity string` field and the new ASFF `Severity` object want the same JSON key `"severity"`. A direct rename breaks existing clients.
**Why it happens:** ASFF uses PascalCase (`Severity`) while the existing schema uses lowercase (`severity`).
**How to avoid:** Name the new field `ASFFSeverity` with JSON tag `"asff_severity"`. The existing `Severity string` field keeps `"severity"`. Clients that want ASFF compliance use `"asff_severity"`.
**Warning signs:** If you see `json:"Severity"` on the new field, stop — that conflicts with existing UAs that send `"Severity"` as a plain string in some payloads.

### Pitfall 2: Resources array minimum of 1
**What goes wrong:** A finding submitted to Security Hub with `Resources: []` or `Resources: null` is rejected — ASFF requires min 1 resource.
**Why it happens:** The ASFF API contract mandates at least one resource item.
**How to avoid:** Document this in the struct comment. Since ASFF-02 forbids making Resources required in our service, the validation constraint is the responsibility of any downstream Security Hub forwarding code, not this service.
**Warning signs:** Any downstream BatchImportFindings call that sends findings with empty Resources will be rejected by Security Hub.

### Pitfall 3: Timestamp format for ASFF fields
**What goes wrong:** ASFF `CreatedAt`/`UpdatedAt` must be ISO 8601 / RFC 3339, but `time.Time.Format(time.RFC3339)` may omit milliseconds that Security Hub expects.
**Why it happens:** Security Hub accepts `2017-03-22T13:22:13.933Z` (with milliseconds) and also `2017-03-22T13:22:13Z`. Go's `time.RFC3339` excludes milliseconds; `time.RFC3339Nano` includes them with nanosecond precision.
**How to avoid:** Store `CreatedAt`/`UpdatedAt` as strings (as received from client); the client is responsible for correct formatting. The service stores and returns verbatim.
**Warning signs:** Fields like `"CreatedAt": "2024-01-15 10:00:00"` (space separator) will be rejected by Security Hub; UAs must format correctly.

### Pitfall 4: ProductArn ARN format validation
**What goes wrong:** If the service validates ProductArn as a proper ARN, many existing integrations break before they can be updated.
**Why it happens:** ASFF specifies a strict ARN format, but ASFF-02 requires accepting old payloads without these fields.
**How to avoid:** Store ProductArn verbatim with no validation. Security Hub will reject invalid ARNs at submission time, not our service.

### Pitfall 5: Findings blob size growth
**What goes wrong:** Adding 7+ new optional fields to each `Finding` and then marshalling potentially hundreds of findings into a single DynamoDB String attribute can approach DynamoDB's 400 KB item limit.
**Why it happens:** DynamoDB has a hard 400 KB item size limit. Each `ResourceASFF` element adds ~50 bytes; for runs with many findings, this could be significant.
**How to avoid:** This is an existing architectural constraint noted in CLAUDE.md as an anti-pattern. For Phase 3, it is acceptable to add the fields without addressing the limit. Note for future phases.

---

## Code Examples

### Minimal ASFF-compliant Finding (what a UA would POST)
```json
{
  "type": "policy",
  "severity": "HIGH",
  "module": "guardrails",
  "resource": "arn:aws:lambda:eu-west-2:123456789012:function:my-fn",
  "description": "Guardrail policy violation detected",
  "remediation": "Review the guardrail configuration",
  "Id": "arn:aws:securityhub:eu-west-2:123456789012:finding/abc-123",
  "title": "Guardrail Policy Violation",
  "generator_id": "ai-check-guardrails/policy",
  "asff_types": ["Software and Configuration Checks/AWS Security Best Practices"],
  "asff_severity": { "Label": "HIGH", "Original": "HIGH" },
  "resources": [{ "Id": "arn:aws:lambda:eu-west-2:123456789012:function:my-fn", "Type": "AwsLambdaFunction" }],
  "created_at": "2024-01-15T10:00:00Z",
  "updated_at": "2024-01-15T10:00:00Z"
}
```

### AuditRun with ASFF identity fields
```json
{
  "run_id": "abc-123",
  "timestamp": "2024-01-15T10:00:00Z",
  "host": "ci-runner-01",
  "product_arn": "arn:aws:securityhub:eu-west-2:123456789012:product/123456789012/default",
  "aws_account_id": "123456789012",
  "findings": [...]
}
```

### Go struct pattern for nested ASFF objects
```go
// Source: docs.aws.amazon.com/securityhub/1.0/APIReference/API_Severity.html
type SeverityASFF struct {
    Label    string `json:"Label,omitempty"`
    Original string `json:"Original,omitempty"`
}

// Pointer receiver makes the entire object omittable when nil
type Finding struct {
    // ...existing fields...
    ASFFSeverity *SeverityASFF  `json:"asff_severity,omitempty"`
    Resources    []ResourceASFF `json:"resources,omitempty"`
}
```

### Test pattern for new ASFF fields in putEvent
```go
// Extend TestPutEventAttributes:
run := AuditRun{
    RunID:        "test-run-asff",
    Timestamp:    time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
    Host:         "testhost",
    ProductArn:   "arn:aws:securityhub:eu-west-2:123456789012:product/123456789012/default",
    AwsAccountId: "123456789012",
    Findings: []Finding{
        {
            Type:        "policy",
            Severity:    "HIGH",
            Module:      "guardrails",
            Description: "test",
            ID:          "finding-001",
            Title:       "Test Finding",
            GeneratorId: "ai-check-guardrails/policy",
            ASFFSeverity: &SeverityASFF{Label: "HIGH", Original: "HIGH"},
            Resources:   []ResourceASFF{{ID: "arn:aws:ec2::123456789012:instance/i-abc", Type: "AwsEc2Instance"}},
            CreatedAt:   "2024-01-15T10:00:00Z",
            UpdatedAt:   "2024-01-15T10:00:00Z",
        },
    },
}
// Assert product_arn and aws_account_id in DynamoDB item
// Assert findings JSON blob contains "finding-001" and "HIGH"
```

---

## ASFF Types Taxonomy for AI Guardrails

[CITED: docs.aws.amazon.com/securityhub/latest/userguide/asff-required-attributes.md]

Recommended `Types` values for AI guardrail findings:

| Finding Category | Recommended ASFF Type |
|-----------------|----------------------|
| Policy violation | `Software and Configuration Checks/AWS Security Best Practices` |
| Data leakage | `Sensitive Data Identifications` |
| Unusual AI output | `Unusual Behaviors` |
| Access control issue | `Software and Configuration Checks/Industry and Regulatory Standards` |

Partial paths are valid: `"Software and Configuration Checks"` alone is acceptable.

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `Severity.Normalized` (0–100 int) | `Severity.Label` (string enum) | ASFF initial design; `Normalized` is now deprecated | Use `Label`; `Normalized` is auto-derived by Security Hub |
| `WorkflowState` | `Workflow.Status` | 2020 ASFF update | `WorkflowState` is deprecated; `Workflow` object is the current pattern |
| Separate `ProductFields` for custom data | `UserDefinedFields` for custom key-value pairs | Ongoing ASFF evolution | Both exist; `UserDefinedFields` preferred for generic custom data |

---

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `Severity` is effectively required by ASFF-01 even though it is Optional at the API level | ASFF Required Fields | If truly optional, no struct change needed for Severity — safe to add as optional `omitempty` |
| A2 | Existing UAs send `Finding.Severity` as a plain string like `"HIGH"` or `"WARN"` rather than the ASFF Label values (`INFORMATIONAL`/`LOW`/`MEDIUM`/`HIGH`/`CRITICAL`) | Gap Analysis | If UAs already use ASFF Label values, the existing `severity` field and new `asff_severity.Label` are redundant — but keeping both is still safe |
| A3 | `product_arn` and `aws_account_id` are supplied by client UAs rather than generated by the service | Architecture | If the service should auto-derive these from Lambda environment, the implementation changes — but ASFF-02 says add fields, not auto-populate |

---

## Open Questions

1. **Should the service auto-populate `CreatedAt`/`UpdatedAt` if the client omits them?**
   - What we know: ASFF requires these for Security Hub import. The service currently stores verbatim data.
   - What's unclear: ASFF-04 says "accepts and stores" — implies store what is sent, not generate.
   - Recommendation: Do not auto-populate in Phase 3. Document that downstream Security Hub forwarding requires the UA to supply these.

2. **Should `ProductArn` and `AwsAccountId` be stamped from Lambda env vars?**
   - What we know: `AWS_LAMBDA_FUNCTION_NAME` is available; AWS account ID would require an STS call.
   - What's unclear: ASFF-03 says "persist new ASFF fields" — if client doesn't send them, do we derive them?
   - Recommendation: Phase 3 stores what the client sends. Auto-derivation is a follow-up.

3. **Does the detail template (`detailTmpl`) need updating to display new ASFF fields?**
   - What we know: ASFF-01 through ASFF-05 focus on storage and testing, not display.
   - What's unclear: Whether ASFF-05 test coverage includes a round-trip that checks the GET /event/{sk} response.
   - Recommendation: Update `getEvent` to read new `product_arn`/`aws_account_id` attributes, but leave the HTML template unchanged unless ASFF-05 tests explicitly exercise the detail view.

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go 1.26+ | Struct changes, go test | Yes | go1.26.2 (arm64) | — |
| `go test ./...` | ASFF-05 | Yes | Passes in 0.016s | — |
| DynamoDB (live) | ASFF-03 acceptance | Not needed for unit tests | — | mockPutter covers all putEvent tests |

No missing dependencies. All changes are pure Go struct + test additions using the existing mock infrastructure.

---

## Security Domain

> `security_enforcement` not set in config — treating as enabled.

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | No | Not relevant — ASFF field addition |
| V3 Session Management | No | Not relevant |
| V4 Access Control | No | Not relevant |
| V5 Input Validation | Yes | New fields are `omitempty` — no new required field validation added; no injection risk from string fields stored verbatim |
| V6 Cryptography | No | Not relevant |

### Known Threat Patterns

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Oversized `resources` array in Finding | Denial of Service | Existing 1 MiB `MaxBytesReader` in `handlePost` limits total payload size |
| Malicious ARN injection in ProductArn | Tampering | Stored verbatim as DynamoDB String; no execution or routing based on value; DynamoDB attribute injection not possible via string values |
| ASFF field name collision with future struct fields | Tampering | All new fields use `omitempty` and distinct JSON tags; reviewed for conflicts |

---

## Sources

### Primary (HIGH confidence)
- [CITED: docs.aws.amazon.com/securityhub/1.0/APIReference/API_AwsSecurityFinding.html] — Complete required/optional field list with types and constraints
- [CITED: docs.aws.amazon.com/securityhub/latest/userguide/asff-required-attributes.md] — Required ASFF fields documentation with examples
- [CITED: docs.aws.amazon.com/securityhub/1.0/APIReference/API_Severity.html] — Severity.Label valid values, Normalized deprecation
- [CITED: docs.aws.amazon.com/securityhub/1.0/APIReference/API_Resource.html] — Resource required fields (Id, Type)
- [CITED: docs.aws.amazon.com/securityhub/1.0/APIReference/API_Remediation.html] — Remediation.Recommendation structure
- [CITED: docs.aws.amazon.com/securityhub/1.0/APIReference/API_Recommendation.html] — Recommendation.Text, Recommendation.Url

### Secondary (MEDIUM confidence)
- Codebase: `/home/hendry/ai-siem-endpoint/main.go` — Current Finding and AuditRun structs, putEvent implementation, getEvent deserialization
- Codebase: `/home/hendry/ai-siem-endpoint/main_test.go` — Existing test patterns (mockPutter, TestMain, TestPutEventAttributes)

---

## Metadata

**Confidence breakdown:**
- ASFF required fields: HIGH — fetched directly from AWS API reference documentation
- Field mapping / gap analysis: HIGH — derived from verified struct comparison against spec
- Architecture (additive approach): HIGH — constrained by ASFF-02 (no replacement)
- DynamoDB impact: HIGH — findings blob pattern is well-understood from existing code
- Go struct patterns: HIGH — standard `encoding/json` with `omitempty` and pointer fields

**Research date:** 2026-05-18
**Valid until:** 2026-08-18 (ASFF spec is stable; AWS rarely changes required fields without major versioning)
