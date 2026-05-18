# External Integrations

**Analysis Date:** 2026-05-18

## APIs & External Services

**AWS Lambda Runtime:**
- Service: AWS Lambda (custom runtime `provided.al2`, arm64)
  - SDK/Client: `github.com/apex/gateway/v2` (bridges `net/http` to Lambda invoke events)
  - Auth: IAM execution role (managed by SAM/CloudFormation)

**AWS API Gateway:**
- Service: AWS Serverless HttpApi (HTTP API v2)
  - Defined in: `template.yml` resource `HttpApi`
  - Routes all `GET /`, `GET /event/`, `POST /` to `MainFunction`
  - Output URL: `https://${HttpApi}.execute-api.${AWS::Region}.amazonaws.com/`

## Data Storage

**Databases:**
- Type: AWS DynamoDB
  - Table name: injected via `DYNAMODB_TABLE` env var (CloudFormation ref `EventsTable`); defaults to `mock-siem-events` locally (`main.go:23-26`)
  - Billing: PAY_PER_REQUEST
  - Schema: composite key — `pk` (String, always `"all"`) + `sk` (String, format `RFC3339#run_id`)
  - Access pattern: single-partition scan with `ScanIndexForward: false, Limit: 50` for listing; key lookup for detail
  - Client: `github.com/aws/aws-sdk-go-v2/service/dynamodb` v1.57.3
  - Connection env var: `DYNAMODB_TABLE`
  - IAM policy: `DynamoDBCrudPolicy` (PutItem, GetItem, Query)

**File Storage:**
- Local filesystem only (no S3 or other object storage)

**Caching:**
- None

## Authentication & Identity

**Auth Provider:**
- None — the HTTP API has no authentication layer; any caller can POST audit runs
- AWS SDK auth uses `config.LoadDefaultConfig` supporting: environment variables, shared credentials file, EC2 IMDS, AWS SSO (`main.go:75-82`)
- Deployment uses SSO profile `AdministratorAccess-407461997746` (`Makefile:2`)

## Monitoring & Observability

**Error Tracking:**
- None — no third-party error tracking

**Logs:**
- `log/slog` (stdlib structured logging)
- JSON format when running on Lambda (`slog.NewJSONHandler`); plain text locally (`main.go:87-88`)
- Log destinations: stderr (Lambda → CloudWatch Logs automatically)

## CI/CD & Deployment

**Hosting:**
- AWS Lambda (arm64) fronted by AWS API Gateway HTTP API
- Region: `eu-west-2`
- CloudFormation stack: `mock-siem-backend`

**CI Pipeline:**
- None detected — deployment is manual via `make deploy` (SAM CLI)
- Build: `sam build` invokes `make build-MainFunction` which cross-compiles Go binary to `bootstrap`

## Environment Configuration

**Required env vars (runtime):**
- `DYNAMODB_TABLE` - DynamoDB table name (provided automatically by SAM via CloudFormation; must be set manually for local runs if not using default)

**Optional env vars:**
- `AWS_LAMBDA_FUNCTION_NAME` - Presence switches to Lambda gateway mode
- `PORT` - Local server port (default `8080`)
- Standard AWS SDK env vars: `AWS_REGION`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_PROFILE`, etc.

**Secrets location:**
- `.env` file: not detected
- Credentials managed through AWS SSO (`aws sso login`) or environment variables for local dev

## Webhooks & Callbacks

**Incoming:**
- `POST /` — Receives `AuditRun` JSON payloads from `ai-check-guardrails` clients (identified via `AI_GUARDRAILS_SIEM_ENDPOINT` env var on the client side)

**Outgoing:**
- None — the service is purely a receiver and viewer; it does not call external webhooks

---

*Integration audit: 2026-05-18*
