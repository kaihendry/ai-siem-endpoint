---
phase: 03-asff-data-model-alignment
plan: "03"
subsystem: testing
tags: [go, dynamodb, asff, security-hub, unit-tests]

# Dependency graph
requires:
  - phase: 03-asff-data-model-alignment
    provides: SeverityASFF, ResourceASFF types, ASFF Finding fields, ProductArn/AwsAccountId on AuditRun, DynamoDB product_arn/aws_account_id writes
provides:
  - Unit test coverage for all ASFF fields in TestPutEventAttributes DynamoDB attribute assertions
  - End-to-end handler round-trip test for ASFF-extended POST payload in TestHandlePost
affects: [future ASFF field expansions, security-hub integration testing]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Extend existing table-driven tests with struct field additions and appended assertions rather than separate test functions"
    - "Use strings.Contains on serialized JSON blob to assert nested ASFF sub-fields indirectly"

key-files:
  created: []
  modified:
    - main_test.go

key-decisions:
  - "Reuse local postJSON helper signature (body []byte) in new sub-test rather than changing to (t, body) to avoid touching existing sub-tests"
  - "Assert ASFF severity label via strings.Contains on the findings JSON blob rather than deserializing — consistent with existing test pattern"

patterns-established:
  - "Append ASFF assertions to existing TestPutEventAttributes block, keeping related assertions grouped by feature area"

requirements-completed: []

# Metrics
duration: 5min
completed: 2026-05-18
---

# Phase 03 Plan 03: ASFF Test Coverage Summary

**Extended test suite with DynamoDB attribute assertions for product_arn/aws_account_id and an ASFF full-payload round-trip handler sub-test**

## Performance

- **Duration:** ~5 min
- **Started:** 2026-05-18T00:00:00Z
- **Completed:** 2026-05-18T00:05:00Z
- **Tasks:** 2
- **Files modified:** 1

## Accomplishments
- Updated TestPutEventAttributes AuditRun literal with ProductArn, AwsAccountId, and a fully-populated ASFF Finding (ID, Title, GeneratorId, ASFFSeverity, Resources, CreatedAt, UpdatedAt)
- Added three new DynamoDB attribute assertions: product_arn string value, aws_account_id string value, finding-001 in findings JSON blob, "Label":"HIGH" ASFF severity in findings JSON blob
- Added asff_extended_payload_returns_201 sub-test to TestHandlePost that posts a full ASFF payload including nested asff_severity and resources objects and verifies 201 + run_id in response

## Task Commits

Each task was committed atomically:

1. **Task 1 + Task 2: Extend test suite with ASFF field assertions and round-trip sub-test** - `cf9f37f` (test)

## Test Output

```
=== RUN   TestPutEventAttributes
--- PASS: TestPutEventAttributes (0.00s)
=== RUN   TestHandlePost
=== RUN   TestHandlePost/happy_path_returns_201_with_run_id
=== RUN   TestHandlePost/missing_run_id_returns_400
=== RUN   TestHandlePost/missing_timestamp_returns_400
=== RUN   TestHandlePost/missing_host_returns_400
=== RUN   TestHandlePost/body_too_large_returns_413
=== RUN   TestHandlePost/asff_extended_payload_returns_201
--- PASS: TestHandlePost (0.01s)
    --- PASS: TestHandlePost/happy_path_returns_201_with_run_id (0.00s)
    --- PASS: TestHandlePost/missing_run_id_returns_400 (0.00s)
    --- PASS: TestHandlePost/missing_timestamp_returns_400 (0.00s)
    --- PASS: TestHandlePost/missing_host_returns_400 (0.00s)
    --- PASS: TestHandlePost/body_too_large_returns_413 (0.01s)
    --- PASS: TestHandlePost/asff_extended_payload_returns_201 (0.00s)
PASS
ok  	github.com/kaihendry/ai-siem-endpoint	0.015s
```

## Verification Grep Counts

```
grep -c "product_arn" main_test.go    # 3 (>= 2 required)
grep -c "aws_account_id" main_test.go # 3 (>= 2 required)
grep "finding-001" main_test.go       # shows assertion string (present)
```

## Assertions Added

Total new assertions in this plan: **4**
1. product_arn DynamoDB attribute value check
2. aws_account_id DynamoDB attribute value check
3. findings JSON blob contains "finding-001"
4. findings JSON blob contains `"Label":"HIGH"` (ASFF severity)

Plus 1 new handler sub-test (asff_extended_payload_returns_201) validating status 201 and run_id in response body.

## Files Created/Modified
- `/home/hendry/ai-siem-endpoint/main_test.go` - Extended TestPutEventAttributes AuditRun literal with ASFF fields; added 4 DynamoDB attribute assertions; added asff_extended_payload_returns_201 sub-test to TestHandlePost

## Decisions Made
- Reused the local `postJSON(body []byte)` helper signature in the new sub-test rather than changing to `postJSON(t, body)` to avoid touching existing sub-tests and breaking their expectations
- Asserted nested ASFF severity field via `strings.Contains(findingsAttr.Value, `"Label":"HIGH"`)` on the serialized blob, consistent with the existing "guardrails" assertion pattern

## Deviations from Plan

The plan's new sub-test snippet called `postJSON(t, body)` but the existing local helper only accepts `(body []byte)`. Applied Rule 3 (auto-fix blocking) inline: called `postJSON(body)` to match the established signature. No behavioral difference; the `t` parameter would have been unused.

---

**Total deviations:** 1 auto-fixed (Rule 3 - signature mismatch in plan snippet)
**Impact on plan:** Trivial signature alignment; test behavior identical to plan intent.

## Issues Encountered
None beyond the postJSON signature adaptation.

## Self-Check

- [x] main_test.go modified and committed at cf9f37f
- [x] All 6 sub-tests pass (go test ./... -v)
- [x] product_arn count: 3 (>= 2)
- [x] aws_account_id count: 3 (>= 2)
- [x] finding-001 assertion present

## Self-Check: PASSED

## Next Phase Readiness
Phase 03 ASFF data model alignment is complete across all 3 plans:
- Plan 01: SeverityASFF/ResourceASFF types added
- Plan 02: ASFF fields on Finding/AuditRun, DynamoDB read/write
- Plan 03: Test coverage for all new ASFF fields

The service is ready to accept and store ASFF-compatible audit run payloads.

---
*Phase: 03-asff-data-model-alignment*
*Completed: 2026-05-18*
