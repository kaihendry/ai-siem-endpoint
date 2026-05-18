# Phase 1: CI/CD + Runtime - Context

**Gathered:** 2026-05-18
**Status:** Ready for planning

<domain>
## Phase Boundary

Establish a GitHub Actions pipeline that deploys to AWS on push to main using OIDC (no stored secrets), and upgrade the Lambda runtime from the deprecated `provided.al2` to `provided.al2023`. The phase ends when pushing to main automatically deploys a working service on the non-deprecated runtime.

</domain>

<decisions>
## Implementation Decisions

### SAM Deploy in CI
- **D-01:** `sam deploy` in CI MUST include `--no-fail-on-empty-changeset` so re-runs don't fail spuriously when nothing has changed.
- **D-02:** Keep `--resolve-s3` — SAM auto-manages the S3 artifacts bucket, no manual S3 setup required.

### Makefile Credentials
- **D-03:** Remove `--profile $(PROFILE)` from the `deploy` and `destroy` targets. Local deployments use `AWS_PROFILE=AdministratorAccess-407461997746 make deploy`. CI gets credentials from OIDC env vars.
- **D-04:** Parameterize `make login` to use `$(PROFILE)` variable (consistent with the existing `PROFILE ?= AdministratorAccess-407461997746` default) rather than hardcoding the profile name inline.

### CI Tool Versions
- **D-05:** Go version in CI: float to `stable` (use `go-version: stable` in `actions/setup-go`). Gets security patches automatically.
- **D-06:** SAM CLI version in CI: no pin — always use latest. Simple pipelines rarely break on SAM CLI updates.

### Claude's Discretion
- Exact job/step names in the workflow YAML
- Whether to cache Go modules in CI (`actions/cache` or `cache: true` in setup-go)
- Any additional `sam deploy` flags beyond those decided above
- Workflow filename (suggested: `.github/workflows/sam-pipeline.yml`)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Source Files to Modify
- `Makefile` — deploy/destroy/login targets; `--profile` must be removed from deploy/destroy, login parameterized
- `template.yml` — `Runtime: provided.al2` under `Globals.Function` must change to `provided.al2023`

### Requirements
- `.planning/REQUIREMENTS.md` — CICD-01, CICD-02, CICD-03, RUNT-01 (all Phase 1 requirements)

### Infrastructure Context
- `.planning/PROJECT.md` — IAM role ARN, AWS account ID, region, stack name, reference workflow

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `Makefile`: already has `STACK`, `PROFILE`, `REGION` variables at top — `PROFILE` default can stay as the documentation default for local use; just strip it from the `sam deploy`/`sam delete` commands
- `template.yml`: single-function SAM template, straightforward `Runtime` field change

### Established Patterns
- Makefile uses `?=` defaults — convention is to override via env var; `AWS_PROFILE` env var is the natural local-dev override after removing `--profile`
- SAM `BuildMethod: makefile` — CI must run `sam build` (which calls `make build-MainFunction`) then `sam deploy`

### Integration Points
- `.github/workflows/` directory does not yet exist — new file to create
- IAM OIDC role `arn:aws:iam::407461997746:role/github-actions-Role-56IHHM969DKJ` is already provisioned; `configure-aws-credentials` action uses it with `role-to-assume`

</code_context>

<specifics>
## Specific Ideas

- Reference workflow pattern: `kaihendry/helloworld` sam-pipeline.yml — uses OIDC + `actions/setup-go` + `aws-actions/setup-sam` + `aws-actions/configure-aws-credentials` + `make deploy`
- AWS account: `407461997746`, region: `eu-west-2`, stack: `mock-siem-backend`
- CICD-03 test gate: `go test ./...` step before deploy; currently passes trivially (no test files), Phase 2 adds real tests that will gate deployment

</specifics>

<deferred>
## Deferred Ideas

- PR-only test runs (test on PRs without deploying) — not selected for discussion; Claude's discretion whether to add a PR test job
- Branch protection rules — out of scope for this phase

</deferred>

---

*Phase: 1-ci-cd-runtime*
*Context gathered: 2026-05-18*
