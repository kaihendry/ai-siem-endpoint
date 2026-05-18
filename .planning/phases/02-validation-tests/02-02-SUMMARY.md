---
phase: 2
plan: "02-02"
subsystem: testing
tags: [go, testing, unit-tests, httptest, dynamodb-mock]
dependency_graph:
  requires: ["02-01"]
  provides: [unit-tests, handler-tests, test-suite]
  affects: [main.go, main_test.go]
tech_stack:
  added: [testing, net/http/httptest]
  patterns: [interface-injection, mock-putter, TestMain-fixture]
key_files:
  created: [main_test.go]
  modified: [main.go]
decisions:
  - "Extracted dynamoPutter interface from main.go to enable DynamoDB mock injection in tests"
  - "Used package-level eventPutter var (defaulting to dynamoClient) as the injection point for mocks"
  - "TestMain installs mock putter before any test runs — avoids real AWS calls in all test cases"
  - "413 test uses large valid-JSON-prefix body instead of random bytes — ensures MaxBytesReader triggers before JSON parse error"
metrics:
  duration: "~5 minutes"
  completed: "2026-05-18"
  tasks_completed: 2
  tasks_total: 2
---

# Phase 2 Plan 02: Unit and Handler Tests Summary

**One-liner:** mockPutter interface injection enables DynamoDB-free unit tests for putEvent attribute construction and handlePost HTTP validation paths.

## What Was Built

- `main_test.go` (new file, package main) with 6 passing tests
- Minimal refactor of `main.go`: extracted `dynamoPutter` interface and `eventPutter` package-level var

### Tests Added

| Test | What It Verifies | Status |
|------|-----------------|--------|
| `TestPutEventAttributes` | DynamoDB item map keys, types, and values from putEvent | PASS |
| `TestHandlePost/happy_path_returns_201_with_run_id` | Valid POST returns 201 with run_id in JSON body | PASS |
| `TestHandlePost/missing_run_id_returns_400` | Empty run_id returns 400 with error field | PASS |
| `TestHandlePost/missing_timestamp_returns_400` | Omitted timestamp returns 400 with error field | PASS |
| `TestHandlePost/missing_host_returns_400` | Empty host returns 400 with error field | PASS |
| `TestHandlePost/body_too_large_returns_413` | Body > 1 MiB returns 413 | PASS |

### Changes to main.go

- Added `dynamoPutter` interface (single method: `PutItem`)
- Added `eventPutter dynamoPutter` package-level var
- `init()` sets `eventPutter = dynamoClient` after constructing the real client
- `putEvent()` now calls `eventPutter.PutItem(...)` instead of `dynamoClient.PutItem(...)`
- `dynamoClient` remains `*dynamodb.Client` for Query and GetItem in listEvents/getEvent

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] 413 test used random bytes that triggered JSON error before size limit**

- **Found during:** T02 first test run
- **Issue:** `bytes.Repeat([]byte("x"), (1<<20)+1)` returned 400 (invalid JSON) instead of 413 because the JSON decoder hit the invalid character before reading enough bytes to trigger MaxBytesReader
- **Fix:** Changed large body to `{"run_id":"` + 1 MiB + 1 bytes of `"a"` + `"}` — a syntactically valid JSON prefix that forces the decoder to read past the limit before returning an error
- **Files modified:** main_test.go
- **Commit:** be72f82

## Self-Check: PASSED

- main_test.go: FOUND at /home/hendry/ai-siem-endpoint/main_test.go
- main.go (modified): FOUND at /home/hendry/ai-siem-endpoint/main.go
- Commit be72f82: FOUND in git log
- go test ./...: exits 0, 6 tests pass

## Known Stubs

None — all tests are fully wired with real handler and mock DynamoDB.

## Threat Flags

None — test-only changes, no new network endpoints or auth paths.
