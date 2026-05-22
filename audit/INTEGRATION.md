# ai-siem-endpoint audit package — integration guide

This package contains the shared types for POSTing audit results to the
ai-siem-endpoint service. Import it so your types stay in sync with the
endpoint automatically.

## Import

```bash
go get github.com/kaihendry/ai-siem-endpoint/audit
```

```go
import "github.com/kaihendry/ai-siem-endpoint/audit"
```

## POST /

Send a JSON-encoded `audit.AuditRun` to the endpoint root. Required fields:
`run_id`, `timestamp` (RFC3339), `host`. All other fields are optional.

```go
run := audit.AuditRun{
    SchemaVersion: "1.0",
    RunID:         "unique-run-id",
    Timestamp:     time.Now().UTC(),
    Host:          "my-host",
    User:          "ci",
    Mode:          "ci",
    Version:       "1.2.3",
    Score:         85,
    ExitCode:      0,
    DurationMs:    1234,

    // ASFF identity — populate if forwarding to AWS Security Hub
    ProductArn:   "arn:aws:securityhub:eu-west-2:ACCOUNT_ID:product/ACCOUNT_ID/default",
    AwsAccountId: "123456789012",

    Findings: []audit.Finding{
        {
            // Required by the endpoint schema
            Type:        "policy",
            Severity:    "HIGH",
            Module:      "guardrails",
            Description: "Guardrail policy violation detected",

            // ASFF fields — omit if not targeting Security Hub
            ID:          "arn:aws:securityhub:eu-west-2:ACCOUNT_ID:finding/UUID",
            Title:       "Guardrail Policy Violation",
            GeneratorId: "ai-check-guardrails/policy",
            ASFFTypes:   []string{"Software and Configuration Checks/AWS Security Best Practices"},
            ASFFSeverity: &audit.SeverityASFF{
                Label:    "HIGH",   // CRITICAL | HIGH | MEDIUM | LOW | INFORMATIONAL
                Original: "HIGH",
            },
            Resources: []audit.ResourceASFF{
                {
                    ID:        "arn:aws:lambda:eu-west-2:ACCOUNT_ID:function:my-fn",
                    Type:      "AwsLambdaFunction",
                    Region:    "eu-west-2",    // optional
                    Partition: "aws",          // optional
                },
            },
            CreatedAt: time.Now().UTC().Format(time.RFC3339),
            UpdatedAt: time.Now().UTC().Format(time.RFC3339),
        },
    },
}

body, err := json.Marshal(run)
// POST body to https://<endpoint>/
```

## Field notes

| Field | Notes |
|-------|-------|
| `Finding.ID` | JSON tag is `"Id"` (capital I, lowercase d) — matches ASFF wire format |
| `Finding.ASFFSeverity` | Pointer — omit the field entirely if not using ASFF |
| `ResourceASFF.ID` / `.Type` | Required within the struct if `Resources` slice is non-empty |
| `AuditRun.ProductArn` | ARN format: `arn:aws:securityhub:REGION:ACCOUNT:product/ACCOUNT/default` |
| `AuditRun.Timestamp` | Encoded as RFC3339 in JSON; pass a `time.Time` value |

## Backwards compatibility

All ASFF fields use `omitempty`. Existing payloads without ASFF fields
continue to be accepted — no changes required to clients that don't need
Security Hub forwarding.
