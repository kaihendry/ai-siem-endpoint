# Codebase Concerns

**Analysis Date:** 2026-05-18

## Security Considerations

**No authentication or authorization on any endpoint:**
- Risk: Any internet client can POST arbitrary audit data to the endpoint, overwriting or flooding the DynamoDB table. Any client can also read all event data via `GET /` and `GET /event/{sk}`.
- Files: `main.go` (`handlePost`, `handleGet`, `handleDetail`), `template.yml` (`HttpApi` resource)
- Current mitigation: None. The AWS API Gateway `HttpApi` resource has no `Auth` property configured, meaning the endpoint is fully public.
- Recommendations: Add API key or IAM authorization to the `HttpApi` resource in `template.yml`, or add a shared-secret middleware check in the handlers (e.g., validate a bearer token from an `Authorization` header).

**No request body size limit:**
- Risk: An attacker can send a multi-megabyte POST body with thousands of findings, causing excessive DynamoDB write costs and inflated Lambda memory usage.
- Files: `main.go` line 114 (`json.NewDecoder(r.Body).Decode`)
- Current mitigation: None. `http.MaxBytesReader` is not used.
- Recommendations: Wrap `r.Body` with `http.MaxBytesReader(w, r.Body, 1<<20)` before decoding (1 MB cap is reasonable for audit payloads).

**Hardcoded AWS account ID and IAM profile in Makefile:**
- Risk: The AWS account ID `407461997746` and the profile name `AdministratorAccess-407461997746` are committed to the repository. This leaks account enumeration information and couples the build to a single individual's AWS SSO profile.
- Files: `Makefile` lines 2, 29
- Current mitigation: None.
- Recommendations: Replace with environment variable defaults, e.g., `PROFILE ?= $${AWS_PROFILE}`, and document the setup in README.md.

**No CORS or security response headers:**
- Risk: Browser-based clients making cross-origin requests receive no CORS policy. The HTML pages lack `Content-Security-Policy`, `X-Frame-Options`, and `X-Content-Type-Options` headers, leaving the UI open to clickjacking and MIME-sniffing attacks.
- Files: `main.go` (`handleGet`, `handleDetail`)
- Current mitigation: None.
- Recommendations: Add a middleware or per-handler header-setting step for `X-Frame-Options: DENY`, `X-Content-Type-Options: nosniff`, and a restrictive `Content-Security-Policy`.

## Tech Debt

**Deprecated Lambda runtime (`provided.al2`):**
- Issue: `template.yml` specifies `Runtime: provided.al2`. AWS deprecated Amazon Linux 2 custom runtimes; end-of-life for `provided.al2` was February 2026 — before this analysis date.
- Files: `template.yml` line 10, `.aws-sam/build/template.yaml`
- Impact: Lambda functions using this runtime will eventually stop receiving security patches and may cease to deploy or execute.
- Fix approach: Change `Runtime: provided.al2` to `Runtime: provided.al2023` in `template.yml` and rebuild with `make deploy`.

**Single-partition DynamoDB design (hotspot):**
- Issue: Every item is written and queried with the same partition key `"all"` (`main.go` lines 142, 240, 388). This concentrates all reads and writes on a single DynamoDB partition.
- Files: `main.go` (`putEvent`, `listEvents`, `getEvent`)
- Impact: At high ingest rates (hundreds of requests/sec) DynamoDB throttles on that partition regardless of `PAY_PER_REQUEST` billing mode. The query also returns at most 50 items and cannot be paginated further.
- Fix approach: For multi-host deployments, change the partition key to `host` (or a date-shard like `YYYY-MM-DD`) and use a GSI for the current "all events" view. For the current scale this is low risk but documents an architectural ceiling.

**No DynamoDB TTL or data retention policy:**
- Issue: The `EventsTable` in `template.yml` has no `TimeToLiveSpecification`. Data accumulates indefinitely.
- Files: `template.yml` (`EventsTable` resource)
- Impact: Growing storage costs over time; no automatic pruning of old audit records.
- Fix approach: Add `TimeToLiveSpecification: { AttributeName: ttl, Enabled: true }` to the table definition and set a `ttl` attribute (e.g., `time.Now().Add(90*24*time.Hour).Unix()`) when writing items in `putEvent`.

**No DynamoDB deletion protection:**
- Issue: The `EventsTable` has no `DeletionProtectionEnabled: true`. A `sam delete` or stack update can permanently destroy all audit history.
- Files: `template.yml` (`EventsTable` resource)
- Impact: Accidental or intentional stack deletion loses all SIEM data.
- Fix approach: Add `DeletionProtectionEnabled: true` to the `EventsTable` properties.

**`context.TODO()` used in production init path:**
- Issue: `newDynamoClient()` calls `config.LoadDefaultConfig(context.TODO())` at startup, bypassing any proper request-scoped context or shutdown signal.
- Files: `main.go` line 76
- Impact: Minor; blocks clean shutdown hooks from propagating to AWS config loading.
- Fix approach: Use `context.Background()` (semantically more accurate for startup) or pass a real context with a timeout.

**`json.NewEncoder(w).Encode` error not checked in `writeJSON`:**
- Issue: `writeJSON` at `main.go` line 472 discards the error from `json.NewEncoder(w).Encode(v)`. If the write fails mid-stream (e.g., client disconnects), the error is silently swallowed.
- Files: `main.go` line 472
- Impact: Low; the response status is already set so recovery is not possible, but logging the error would aid diagnostics.
- Fix approach: `if err := json.NewEncoder(w).Encode(v); err != nil { slog.Error("writeJSON encode", "err", err) }`.

**Template execute errors after response header write:**
- Issue: In both `handleGet` (line 303) and `handleDetail` (line 464), `Content-Type` and `200 OK` are written before `summaryTemplate.Execute` / `detailTemplate.Execute` is called. If the template errors mid-stream, the error is only logged; the client receives a partial HTML document with no error indication.
- Files: `main.go` lines 302–305, 463–465
- Impact: Clients silently receive truncated HTML pages when a template execution error occurs.
- Fix approach: Render the template to a `bytes.Buffer` first, then write the header and buffer to the response only on success.

## Known Bugs

**Duplicate `run_id` submissions silently overwrite existing records:**
- Symptoms: Posting the same `run_id` with a different `timestamp` creates a new item (different `sk`); posting the same `run_id` with the same `timestamp` silently replaces the existing DynamoDB item because `PutItem` has no `ConditionExpression`.
- Files: `main.go` lines 134, 159 (`putEvent`)
- Trigger: Client retries or replayed requests with identical `run_id` + `timestamp`.
- Workaround: None currently.

## Performance Bottlenecks

**`listEvents` hard-coded limit of 50 with no pagination:**
- Problem: `listEvents` uses `Limit: aws.Int32(50)` with no `ExclusiveStartKey` loop. Only the 50 most recent items are returned regardless of actual data volume; there is no UI or API pagination.
- Files: `main.go` lines 236–244
- Cause: The query returns up to 50 items and discards `out.LastEvaluatedKey`.
- Improvement path: Implement cursor-based pagination using `out.LastEvaluatedKey` as an `ExclusiveStartKey` for subsequent queries, exposed via a `?cursor=` query parameter on `GET /`.

## Fragile Areas

**Monolithic `main.go` (473 lines, all logic in one file):**
- Files: `main.go`
- Why fragile: All HTTP handlers, DynamoDB access, HTML templates, and type definitions live in a single file. Adding new endpoints or storage backends requires editing this single file, increasing merge conflict risk.
- Safe modification: When adding new routes or storage logic, extract DynamoDB operations into a separate `store.go` file and templates into a `templates/` directory.
- Test coverage: Zero — no `*_test.go` files exist in the repository.

**Global `dynamoClient` initialized in `init()`:**
- Files: `main.go` lines 70–82
- Why fragile: The global `dynamoClient` and `tableName` are set in `init()`. If `config.LoadDefaultConfig` fails (e.g., no AWS credentials in local dev), the process calls `os.Exit(1)` before any handler is registered, making local development without AWS credentials impossible without additional configuration.
- Safe modification: Guard with a local-only mock or `os.Getenv("AWS_LAMBDA_FUNCTION_NAME") == ""` check to skip DynamoDB initialization when running locally without credentials.

## Test Coverage Gaps

**No tests whatsoever:**
- What's not tested: All HTTP handler logic, DynamoDB serialization/deserialization, base64 URL encoding for the event detail route, score class template function, `mul` template function, `putEvent`, `listEvents`, `getEvent`.
- Files: `main.go` (entire file)
- Risk: Any refactor silently breaks request handling, DynamoDB attribute mapping, or HTML rendering with no automated detection.
- Priority: High — at minimum, table-driven unit tests for `putEvent` attribute construction and handler tests using `httptest.NewRecorder` would cover the most critical paths.

## Dependencies at Risk

**`github.com/apex/gateway/v2 v2.0.0` — pinned to exact version, low activity:**
- Risk: The `apex/gateway` library bridges Lambda API Gateway Proxy v2 events to `net/http`. It is pinned at `v2.0.0` (2021) with no updates. If AWS changes the Lambda invocation contract or the library has an unpatched bug, there is no upstream fix path.
- Impact: Lambda HTTP handling may silently mis-route or drop requests.
- Migration plan: Evaluate replacing with the official `aws/aws-lambda-go` `httpadapter` package (`github.com/awslabs/aws-lambda-go-api-proxy`), which is actively maintained by AWS.

**`github.com/aws/aws-sdk-go-v2/service/signin v1.0.11` — unexpected indirect dependency:**
- Risk: This package appears as an indirect dependency but is not a well-known public AWS service SDK. It may be a transitive pull from the SSO/SSOOIDC auth chain. Its presence is unexplained in the module graph.
- Impact: Unknown; could introduce unreferenced code or be removed in a future aws-sdk-go-v2 release, causing build failures.
- Migration plan: Run `go mod tidy` to verify it is still required; if not, remove it.

---

*Concerns audit: 2026-05-18*
