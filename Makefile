COVERAGE_FILE ?= coverage.out
LOCAL_BIN := $(CURDIR)/bin
EASYP_BIN := $(LOCAL_BIN)/easyp
GOLANGCI_BIN := $(LOCAL_BIN)/golangci-lint

bin/mockgen:
	@echo "Installing mockgen..."
	@mkdir -p $(LOCAL_BIN)
	GOBIN=$(LOCAL_BIN) go install go.uber.org/mock/mockgen@v0.6.0

bin/easyp:
	@echo "Installing easyp..."
	@mkdir -p $(LOCAL_BIN)
	GOBIN=$(LOCAL_BIN) go install github.com/easyp-tech/easyp/cmd/easyp@v0.14.0

bin/protoc-gen-go:
	@echo "Installing protoc-gen-go..."
	@mkdir -p $(LOCAL_BIN)
	GOBIN=$(LOCAL_BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1

bin/protoc-gen-go-grpc:
	@echo "Installing protoc-gen-go-grpc..."
	@mkdir -p $(LOCAL_BIN)
	GOBIN=$(LOCAL_BIN) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0

bin/protoc-gen-validate:
	@echo "Installing protoc-gen-validate..."
	@mkdir -p $(LOCAL_BIN)
	GOBIN=$(LOCAL_BIN) go install github.com/envoyproxy/protoc-gen-validate@v1.2.1

bin/protoc-gen-grpc-gateway:
	@echo "Installing protoc-gen-grpc-gateway..."
	@mkdir -p $(LOCAL_BIN)
	GOBIN=$(LOCAL_BIN) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.18.1

bin/protoc-gen-openapiv2:
	@echo "Installing protoc-gen-openapiv2..."
	@mkdir -p $(LOCAL_BIN)
	GOBIN=$(LOCAL_BIN) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.18.1

generate-proto: bin/easyp bin/protoc-gen-go bin/protoc-gen-go-grpc bin/protoc-gen-validate bin/protoc-gen-grpc-gateway bin/protoc-gen-openapiv2
	PATH="$(PATH):$(LOCAL_BIN)" && $(EASYP_BIN) mod download && $(EASYP_BIN) generate

generate-mocks: bin/mockgen
	PATH="$(PATH):$(LOCAL_BIN)" go generate ./...
	$(LOCAL_BIN)/mockgen -destination=internal/bot/client/mocks/mock_generated_scrapper_client.go -package=mocks gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/generated/api/scrapper ScrapperClient
	$(LOCAL_BIN)/mockgen -destination=internal/scrapper/client/mocks/mock_generated_bot_client.go -package=mocks gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/generated/api/bot BotClient
	$(LOCAL_BIN)/mockgen -destination=internal/scrapper/controller/mocks/mock_generated_scrapper_server.go -package=mocks gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/generated/api/scrapper ScrapperServer

.PHONY: build
build: generate-proto
	@mkdir -p $(LOCAL_BIN)
	go build -o ./bin/app .urlshort/cmd/urlshort

# test: run all tests
.PHONY: test
test: generate-mocks
	@go test -coverpkg='./urlshort/...' --race -count=1 -coverprofile='$(COVERAGE_FILE)' ./...
	@go tool cover -func='$(COVERAGE_FILE)' | grep ^total | tr -s '\t'

bin/golangci-lint:
	@echo "Installing golangci-lint..."
	@mkdir -p $(LOCAL_BIN)
	curl -sSfL https://golangci-lint.run/install.sh | sh -s v2.10.1

.PHONY: lint
lint: bin/golangci-lint
	@echo 'Running linter on files...'
	$(GOLANGCI_BIN) run \
	--config=.golangci.yaml \
	--max-issues-per-linter=0 \
	--max-same-issues=0