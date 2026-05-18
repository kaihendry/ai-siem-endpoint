# Phase 1: CI/CD + Runtime - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-05-18
**Phase:** 1-ci-cd-runtime
**Areas discussed:** SAM deploy CI flags, Local dev story, Go + SAM versions in CI

---

## SAM Deploy CI Flags

| Option | Description | Selected |
|--------|-------------|----------|
| Yes — add `--no-fail-on-empty-changeset` | CI re-runs won't fail spuriously. Standard practice. | ✓ |
| No — keep strict | Empty changeset = failure. Forces awareness of no-change deploys. | |

**User's choice:** Add `--no-fail-on-empty-changeset`
**Notes:** Standard practice for SAM pipelines.

| Option | Description | Selected |
|--------|-------------|----------|
| Keep `--resolve-s3` | SAM auto-manages the bucket. No extra setup needed. | ✓ |
| Pre-existing named bucket | More control but requires maintaining a separate S3 bucket. | |

**User's choice:** Keep `--resolve-s3`

---

## Local Dev Story

| Option | Description | Selected |
|--------|-------------|----------|
| Keep `make deploy` — set AWS_PROFILE env var | Remove flag from Makefile; use `AWS_PROFILE=... make deploy` locally. | ✓ |
| Add `make deploy-local` target | Keep `deploy` clean for CI; add separate target with `--profile`. | |
| You decide | Claude's discretion. | |

**User's choice:** Remove flag from `deploy`/`destroy` targets; use `AWS_PROFILE` env var locally.

| Option | Description | Selected |
|--------|-------------|----------|
| Leave `make login` as-is | It's a local convenience target; hardcoded profile is fine. | |
| Parameterize it too | Use `$(PROFILE)` variable consistently throughout. | ✓ |

**User's choice:** Parameterize `make login` to use `$(PROFILE)`.

---

## Go + SAM Versions in CI

| Option | Description | Selected |
|--------|-------------|----------|
| Pin to '1.26' | Matches go.mod exactly. Deterministic builds. | |
| Float to 'stable' | Always uses latest stable Go. Gets security patches automatically. | ✓ |

**User's choice:** Float to `stable`.

| Option | Description | Selected |
|--------|-------------|----------|
| Latest (no version pin) | Always gets newest SAM CLI. Rarely breaks for simple pipelines. | ✓ |
| Pin to specific version | Maximum reproducibility but requires manual bumps. | |

**User's choice:** No SAM CLI version pin.

---

## Claude's Discretion

- Exact job/step names in the workflow YAML
- Go module caching strategy in CI
- Additional `sam deploy` flags beyond the decided ones
- Workflow filename

## Deferred Ideas

- PR-only test runs — user did not select for discussion
- Branch protection rules — out of scope
