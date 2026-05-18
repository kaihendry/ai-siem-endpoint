# ai-siem-endpoint

## What This Is

A Go HTTP service deployed on AWS Lambda that acts as a shared SIEM (Security Information and Event Management) backend for AI guardrail audit tools. Multiple reporting user agents (from `ai-check-guardrails`) POST their `AuditRun` results here; the service stores them in DynamoDB and provides a web dashboard to review findings. The endpoint needs to be reliably deployed via CI/CD so any UA can always report to a live, versioned URL.

## Core Value

A stable, publicly reachable endpoint that any `ai-check-guardrails` UA can POST audit results to and trust will store and display them correctly.

## Requirements

### Validated

- ✓ Receives AuditRun events via POST / and stores them in DynamoDB — existing
- ✓ Lists last 50 events in an HTML summary dashboard (GET /) — existing
- ✓ Shows full event detail view (GET /event/{sk}) — existing
- ✓ Runs on AWS Lambda (arm64, provided.al2) + API Gateway — existing
- ✓ Dual-mode: same binary runs locally (HTTP server) and on Lambda — existing

### Active

- [ ] GitHub Actions CI/CD pipeline deploys on push to main using OIDC role assumption
- [ ] Makefile deploy target works in CI (no personal --profile flag)
- [ ] Lambda runtime upgraded from deprecated provided.al2 to provided.al2023
- [ ] Incoming AuditRun payloads validated against required schema (reject malformed requests)
- [ ] Request body size capped to prevent abuse (http.MaxBytesReader)
- [ ] Basic Go unit tests covering handler logic and DynamoDB serialization
- [ ] CI pipeline runs tests before deploying

### Out of Scope

- Authentication/authorization on endpoints — explicitly deferred; internal/dev tool only for now
- DynamoDB partition key redesign — current single-partition design acceptable at this scale
- Pagination for GET / — deferred to a future milestone
- Mobile or API-first client SDK — web dashboard is sufficient for v1

## Context

- Codebase is a single `main.go` file (~473 lines) with all logic: handlers, DynamoDB storage, HTML templates
- Infrastructure defined in `template.yml` (AWS SAM / CloudFormation)
- Known issue: `template.yml` uses `Runtime: provided.al2` which reached end-of-life February 2026
- Makefile hard-codes `--profile AdministratorAccess-407461997746` which prevents CI usage
- Zero test files currently — any refactor or new validation logic has no safety net
- AWS region: eu-west-2, CloudFormation stack: `mock-siem-backend`
- IAM role for GitHub Actions already provisioned: `arn:aws:iam::407461997746:role/github-actions-Role-56IHHM969DKJ`
- Reference workflow: kaihendry/helloworld sam-pipeline.yml (OIDC + setup-go + setup-sam + configure-aws-credentials + make deploy)

## Constraints

- **Tech stack**: Go + AWS SAM + DynamoDB — stay with this stack, no new services
- **CI platform**: GitHub Actions only
- **AWS region**: eu-west-2 (hard requirement from existing infrastructure)
- **Lambda arch**: arm64 / provided.al2023 (upgrade from al2, keep arm64)

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Makefile deploy target stripped of --profile | CI doesn't use SSO profiles; credentials come from OIDC env vars | — Pending |
| Schema validation inline in handlePost | Simple validation, no external library needed | — Pending |
| provided.al2023 runtime | al2 reached EOL Feb 2026 | — Pending |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd:complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-05-18 after initialization*
