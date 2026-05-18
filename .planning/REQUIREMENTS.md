# Requirements: ai-siem-endpoint

**Defined:** 2026-05-18
**Core Value:** A stable, publicly reachable endpoint that any ai-check-guardrails UA can POST audit results to and trust will store and display them correctly.

## v1 Requirements

### CI/CD

- [ ] **CICD-01**: GitHub Actions workflow deploys to AWS on push to main using OIDC role assumption (no stored secrets)
- [ ] **CICD-02**: Makefile deploy target works without --profile flag (CI uses OIDC-injected credentials)
- [ ] **CICD-03**: CI pipeline runs tests and only deploys if they pass

### Runtime

- [ ] **RUNT-01**: Lambda runtime upgraded from deprecated provided.al2 to provided.al2023 in template.yml

### Validation

- [x] **VALD-01**: POST / rejects payloads missing required fields (run_id, timestamp, host) with HTTP 400 and descriptive error
- [x] **VALD-02**: POST / rejects payloads exceeding 1 MB with HTTP 413

### Tests

- [x] **TEST-01**: Unit tests cover putEvent DynamoDB attribute construction (round-trip serialization)
- [x] **TEST-02**: Handler tests cover POST / happy path and validation error paths using httptest

## v2 Requirements

### Security

- **SEC-01**: API key or shared-secret bearer token authentication on all endpoints
- **SEC-02**: Security response headers (X-Frame-Options, X-Content-Type-Options, CSP) on HTML responses

### Observability

- **OBS-01**: Pagination for GET / (cursor-based, beyond 50 events)
- **OBS-02**: DynamoDB TTL for automatic data expiry (configurable retention period)

### Resilience

- **RES-01**: DynamoDB deletion protection enabled on EventsTable
- **RES-02**: Conditional PUT to prevent silent overwrites on duplicate run_id + timestamp

## Out of Scope

| Feature | Reason |
|---------|--------|
| Authentication | Internal/dev tool; auth deferred to a later milestone |
| DynamoDB partition key redesign | Current single-partition design acceptable at this scale |
| Mobile/native client | Web dashboard sufficient |
| CORS headers | No cross-origin browser clients in v1 |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| CICD-01 | Phase 1 | Pending |
| CICD-02 | Phase 1 | Pending |
| CICD-03 | Phase 1 | Pending |
| RUNT-01 | Phase 1 | Pending |
| VALD-01 | Phase 2 | Complete |
| VALD-02 | Phase 2 | Complete |
| TEST-01 | Phase 2 | Complete |
| TEST-02 | Phase 2 | Complete |

**Coverage:**
- v1 requirements: 8 total
- Mapped to phases: 8
- Unmapped: 0 ✓

---
*Requirements defined: 2026-05-18*
*Last updated: 2026-05-18 after initial definition*
