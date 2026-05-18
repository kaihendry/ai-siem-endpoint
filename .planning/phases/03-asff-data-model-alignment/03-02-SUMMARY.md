---
phase: 03-asff-data-model-alignment
plan: "02"
subsystem: storage
tags: [dynamodb, asff, putEvent, getEvent]
key-files:
  modified:
    - main.go
decisions:
  - "product_arn and aws_account_id stored as explicit DynamoDB String attributes alongside existing fields"
  - "Finding ASFF fields continue to be persisted via the existing JSON blob approach — no DynamoDB schema change required"
metrics:
  duration: "< 5 minutes"
  completed: "2026-05-18"
  tasks_completed: 2
  tasks_total: 2
---

# Phase 03 Plan 02: Persist product_arn and aws_account_id in DynamoDB Summary

**One-liner:** Extended putEvent to write `product_arn` and `aws_account_id` as explicit DynamoDB String attributes, and extended getEvent to read them back via the existing `strAttr` closure.

## What Was Done

- **Task 1 (putEvent):** Added two new entries to the item map in `putEvent` (main.go line 202-203), placed after `duration_ms` and before `findings`. Both use `AttributeValueMemberS` matching the existing String attribute pattern.

- **Task 2 (getEvent):** Added two `strAttr` assignments in `getEvent` (main.go lines 472-473) immediately after `run.Version = strAttr("version")`. The `strAttr` closure safely returns `""` for missing keys, preserving backwards compatibility with events stored before this change.

## Key Changes

| Attribute | DynamoDB Type | putEvent location | getEvent location |
|---|---|---|---|
| `product_arn` | String (S) | main.go:202 | main.go:472 |
| `aws_account_id` | String (S) | main.go:203 | main.go:473 |

The `Finding` struct ASFF fields (`ID`, `Title`, `GeneratorId`, `ASFFTypes`, `ASFFSeverity`, `Resources`, `CreatedAt`, `UpdatedAt`) are automatically included via the existing `json.Marshal(run.Findings)` blob — no DynamoDB schema changes were needed for those fields.

## Verification Results

```
go build ./...
BUILD OK

go test ./...
ok      github.com/kaihendry/ai-siem-endpoint   0.015s

grep "product_arn" main.go:
  78:    ProductArn   string `json:"product_arn,omitempty"`
  202:   "product_arn":    &types.AttributeValueMemberS{Value: run.ProductArn},
  472:   run.ProductArn   = strAttr("product_arn")

grep "aws_account_id" main.go:
  79:    AwsAccountId string `json:"aws_account_id,omitempty"`
  203:   "aws_account_id": &types.AttributeValueMemberS{Value: run.AwsAccountId},
  473:   run.AwsAccountId = strAttr("aws_account_id")
```

## Commit

`2dac480` — feat(03-02): persist product_arn and aws_account_id in DynamoDB putEvent/getEvent

## Deviations from Plan

None - plan executed exactly as written.

## Self-Check: PASSED

- main.go modified and verified
- go build: PASSED
- go test: PASSED
- grep checks: product_arn appears in putEvent (line 202) and getEvent (line 472); aws_account_id appears in putEvent (line 203) and getEvent (line 473)
- commit 2dac480 exists
