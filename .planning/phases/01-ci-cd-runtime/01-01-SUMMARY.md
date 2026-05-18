---
plan: 01-01
phase: 01-ci-cd-runtime
status: complete
completed_at: "2026-05-18"
requirements_covered:
  - CICD-01
  - CICD-02
  - CICD-03
  - RUNT-01
---

# Plan 01-01 Summary: Wire up CI/CD + Runtime Upgrade

## What was built

Three targeted changes to wire GitHub Actions OIDC deployment and fix the Lambda runtime.

### T01 — `.github/workflows/sam-pipeline.yml` (created)

New GitHub Actions workflow that:
- Triggers on push to `main` only
- Uses OIDC (`aws-actions/configure-aws-credentials@v4`) with role `arn:aws:iam::407461997746:role/github-actions-Role-56IHHM969DKJ` — no stored AWS secrets
- Runs `go test ./...` before deployment (test gate)
- Deploys via `make deploy`

### T02 — `Makefile` (updated)

- Removed `--profile $(PROFILE)` from `deploy` and `destroy` targets — CI uses OIDC-injected credentials
- Added `--no-fail-on-empty-changeset` to `deploy` — re-runs on unchanged stacks no longer fail
- Parameterized `login` target to use `$(PROFILE)` variable (was hardcoded `AdministratorAccess-407461997746`)

### T03 — `template.yml` (updated)

- `Globals.Function.Runtime` changed from `provided.al2` → `provided.al2023`
- No other changes; `Architectures: [arm64]` preserved

## Requirements covered

| Req | Description | Evidence |
|-----|-------------|----------|
| CICD-01 | Workflow triggers on push to main, uses OIDC | `sam-pipeline.yml` push/main trigger + `role-to-assume` |
| CICD-02 | No `--profile` flag in deploy/destroy | Makefile deploy and destroy targets have no `--profile` |
| CICD-03 | Tests run before deploy | `go test ./...` step precedes `make deploy` in workflow |
| RUNT-01 | Lambda runtime is provided.al2023 | `template.yml` Globals.Function.Runtime |

## Verification results

```
1. Runtime: provided.al2023  ✓
2. --profile absent from deploy/destroy  ✓
3. --no-fail-on-empty-changeset present  ✓
4. Workflow has role-to-assume, go test, make deploy in order  ✓
5. YAML lint passes  ✓
```

## Commits

- `feat(01-01/T01)`: add GitHub Actions OIDC deploy workflow
- `fix(01-01/T02)`: remove --profile from deploy/destroy, add --no-fail-on-empty-changeset
- `fix(01-01/T03)`: upgrade Lambda runtime from provided.al2 to provided.al2023
