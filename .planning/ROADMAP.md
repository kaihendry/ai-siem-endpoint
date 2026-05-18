# Roadmap: ai-siem-endpoint

**Created:** 2026-05-18
**Phases:** 3
**Requirements covered:** 8/8 ✓

---

### Phase 1: CI/CD + Runtime

**Goal:** Pushing to main automatically deploys the service to AWS using OIDC (no stored secrets), on the non-deprecated Lambda runtime.
**Mode:** mvp

**Requirements:**
- CICD-01: GitHub Actions workflow deploys on push to main using OIDC
- CICD-02: Makefile works without --profile flag
- CICD-03: CI pipeline runs tests before deploying
- RUNT-01: Lambda runtime upgraded to provided.al2023

**Plans:** 1 plan

Plans:
- [x] 01-01-PLAN.md — Create GitHub Actions workflow, fix Makefile credentials, upgrade Lambda runtime

**Success Criteria:**
1. `.github/workflows/sam-pipeline.yml` exists and triggers on push to main
2. `make deploy` succeeds in CI without `--profile` flag (credentials from OIDC env)
3. Lambda function uses `provided.al2023` runtime after deployment

---

### Phase 2: Validation + Tests ✓ Complete

**Goal:** The endpoint validates incoming payloads, caps request size, and has a test suite that CI runs before deploying.
**Mode:** mvp

**Requirements:**
- VALD-01: POST / rejects payloads missing required fields with HTTP 400 ✓
- VALD-02: POST / rejects payloads exceeding 1 MB with HTTP 413 ✓
- TEST-01: Unit tests for putEvent DynamoDB attribute construction ✓
- TEST-02: Handler tests for POST / happy path and validation errors ✓

**Plans:** 2 plans

Plans:
- [x] 02-01-PLAN.md — Add body size limit (1 MB → 413) and required field validation (timestamp, host → 400)
- [x] 02-02-PLAN.md — Unit tests (putEvent attrs) + httptest handler tests (6 tests passing)

**Success Criteria:**
1. POST / with missing required field returns HTTP 400 with descriptive error ✓
2. POST / with body > 1 MB returns HTTP 413 ✓
3. `go test ./...` passes and runs in the CI workflow before deployment ✓

---

### Phase 3: ASFF Data Model Alignment

**Goal:** Align the `AuditRun` and `Finding` data model with the AWS Security Finding Format (ASFF) so findings stored in this service can be forwarded to AWS Security Hub or consumed by ASFF-aware tooling without translation.

**Requirements:**
- ASFF-01: `Finding` struct maps to required ASFF fields (Id, ProductArn, GeneratorId, AwsAccountId, Types, CreatedAt, UpdatedAt, Severity.Label, Title, Description, Resources)
- ASFF-02: `AuditRun` retains its existing fields for backwards compatibility; ASFF fields are added, not replacing
- ASFF-03: DynamoDB storage persists new ASFF fields
- ASFF-04: POST / accepts and stores ASFF-extended payloads
- ASFF-05: Tests cover new ASFF fields in putEvent and handler

**Plans:** 3/3 plans complete

Plans:
- [x] 03-01-PLAN.md — Add SeverityASFF/ResourceASFF types; extend Finding (8 new fields) and AuditRun (2 new fields) with omitempty
- [x] 03-02-PLAN.md — Extend putEvent to persist product_arn/aws_account_id; extend getEvent to read them back
- [x] 03-03-PLAN.md — Extend TestPutEventAttributes with ASFF assertions; add ASFF round-trip sub-test in TestHandlePost

**Success Criteria:**
1. A finding stored via POST / round-trips all ASFF required fields
2. `go test ./...` passes with coverage of new ASFF fields
3. Existing non-ASFF payloads still accepted (backwards compatible)

---

## Requirement Coverage

| Requirement | Phase | Status |
|-------------|-------|--------|
| CICD-01 | Phase 1 | ✓ Complete |
| CICD-02 | Phase 1 | ✓ Complete |
| CICD-03 | Phase 1 | ✓ Complete |
| RUNT-01 | Phase 1 | ✓ Complete |
| VALD-01 | Phase 2 | ✓ Complete |
| VALD-02 | Phase 2 | ✓ Complete |
| TEST-01 | Phase 2 | ✓ Complete |
| TEST-02 | Phase 2 | ✓ Complete |
| ASFF-01 | Phase 3 | Planned |
| ASFF-02 | Phase 3 | Planned |
| ASFF-03 | Phase 3 | Planned |
| ASFF-04 | Phase 3 | Planned |
| ASFF-05 | Phase 3 | Planned |

**Coverage:** 8/8 v1 requirements mapped ✓ + 5/5 Phase 3 requirements planned

---
*Roadmap created: 2026-05-18*
