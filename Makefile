STACK ?= mock-siem-backend
PROFILE ?= AdministratorAccess-407461997746
REGION ?= eu-west-2

build-MainFunction:
	CGO_ENABLED=0 GOARCH=arm64 GOOS=linux go build -o $(ARTIFACTS_DIR)/bootstrap .

deploy:
	sam build
	sam deploy \
		--stack-name $(STACK) \
		--region $(REGION) \
		--profile $(PROFILE) \
		--capabilities CAPABILITY_IAM \
		--resolve-s3 \
		--no-confirm-changeset

destroy:
	sam delete \
		--stack-name $(STACK) \
		--region $(REGION) \
		--profile $(PROFILE) \
		--no-prompts

local:
	go run .

login:
	aws sso login --profile AdministratorAccess-407461997746 --use-device-code

.PHONY: build-MainFunction deploy destroy local
