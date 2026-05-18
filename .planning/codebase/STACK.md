# Technology Stack

**Analysis Date:** 2026-05-18

## Languages

**Primary:**
- Go 1.26 - All application logic (`main.go`)

**Secondary:**
- YAML - SAM/CloudFormation infrastructure definition (`template.yml`)
- Makefile - Build and deployment automation (`Makefile`)

## Runtime

**Environment:**
- AWS Lambda custom runtime (`provided.al2`) on arm64
- Local development: Go standard HTTP server on port 8080 (default)

**Package Manager:**
- Go modules (`go.mod`, `go.sum`)
- Lockfile: `go.sum` present

## Frameworks

**Core:**
- `net/http` (stdlib) - HTTP routing and handler registration
- `github.com/apex/gateway/v2` v2.0.0 - Bridges `net/http` to AWS Lambda runtime

**Build/Dev:**
- AWS SAM CLI - Build and deploy tooling (`Makefile` targets: `build-MainFunction`, `deploy`, `destroy`)
- `go build` - Cross-compiled for `GOARCH=arm64 GOOS=linux`, output binary named `bootstrap`

**Testing:**
- Not detected — no test files (`*_test.go`) present

## Key Dependencies

**Critical:**
- `github.com/apex/gateway/v2` v2.0.0 - Enables dual-mode operation: same `net/http` handlers run locally or on Lambda
- `github.com/aws/aws-sdk-go-v2` v1.41.7 - AWS SDK core
- `github.com/aws/aws-sdk-go-v2/config` v1.32.17 - AWS credential/config loading (`config.LoadDefaultConfig`)
- `github.com/aws/aws-sdk-go-v2/service/dynamodb` v1.57.3 - DynamoDB client for event persistence

**Infrastructure (indirect):**
- `github.com/aws/aws-lambda-go` v1.54.0 - Lambda runtime interface (pulled in by apex/gateway)
- `github.com/aws/smithy-go` v1.25.1 - AWS SDK transport layer
- `github.com/pkg/errors` v0.9.1 - Error wrapping (transitive)

## Configuration

**Environment:**
- `DYNAMODB_TABLE` - DynamoDB table name (defaults to `mock-siem-events` if unset); injected via SAM `Environment.Variables` in `template.yml`
- `AWS_LAMBDA_FUNCTION_NAME` - Presence determines runtime mode (Lambda vs local HTTP server)
- `PORT` - Local HTTP listen port (defaults to `8080`)
- AWS credentials loaded via `config.LoadDefaultConfig` (supports env vars, shared credentials, EC2 IMDS, SSO)

**Build:**
- `template.yml` - SAM/CloudFormation template defining Lambda function, HTTP API, and DynamoDB table
- `Makefile` - Build method used by SAM (`BuildMethod: makefile`); cross-compiles to `arm64` Linux

## Platform Requirements

**Development:**
- Go 1.26+
- AWS SAM CLI (for `sam build` / `sam deploy`)
- AWS credentials configured (SSO profile: `AdministratorAccess-407461997746`)

**Production:**
- AWS Lambda (arm64, `provided.al2` runtime)
- AWS API Gateway HTTP API
- AWS DynamoDB (PAY_PER_REQUEST billing)
- Target region: `eu-west-2`
- CloudFormation stack name: `mock-siem-backend`

---

*Stack analysis: 2026-05-18*
