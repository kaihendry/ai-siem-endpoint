# Roadmap: ai-siem-endpoint

**Created:** 2026-05-18
**Phases:** 2
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

**Coverage:** 8/8 v1 requirements mapped ✓

---
*Roadmap created: 2026-05-18*
