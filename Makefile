# A Self-Documenting Makefile: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html

VERSION=$(shell git describe --always --tags | cut -d "v" -f 2)
LINKER_FLAGS=-s -w -X github.com/oasdiff/oasdiff/build.Version=${VERSION}
GOLANGCILINT_VERSION=v1.52.2

.PHONY: test
test: doc-breaking-changes localize ## Run tests
	@echo "==> Running tests..."
	go test ./...

.PHONY: coverage
coverage: ## Run tests with coverage
	@echo "==> Running tests with coverage..."
	go test -coverprofile=coverage.out ./...

.PHONY: coverage-html
coverage-html: coverage ## Generate HTML coverage report
	@echo "==> Generating HTML coverage report..."
	go tool cover -html=coverage.out -o coverage.html

.PHONY: build
build: ## Build oasdiff binary
	@echo "==> Building oasdiff binary"
	go build -ldflags "$(LINKER_FLAGS)" -o ./bin/oasdiff .

.PHONY: install
install: deps ## Install oasdiff binary
	@echo "==> Installing oasdiff binary..."
	go install -ldflags "$(LINKER_FLAGS)" .

.PHONY: doc-breaking-changes
doc-breaking-changes: ## Generate documentation for breaking changes
	@echo "==> Updating breaking changes documentation..."
	./scripts/doc_breaking_changes.sh > docs/BREAKING-CHANGES-EXAMPLES.md

.PHONY: deps
deps:  ## Download go module dependencies
	@echo "==> Installing go.mod dependencies..."
	go mod download
	go mod tidy

.PHONY: lint
lint: ## Run linter
	go fmt ./...
	go vet ./...
	golangci-lint run --enable=unused
	
.PHONY: localize
localize: ## Compile localized changelog messages
	@echo "==> Compiling localized changelog messages..."
	go install github.com/m1/go-localize@latest
	go-localize -input checker/localizations_src -output checker/localizations 
	go fmt ./checker/localizations

.PHONY: devtools
devtools:  ## Install dev tools
	@echo "==> Installing dev tools..."
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin $(GOLANGCILINT_VERSION)

.PHONY: help
help: ## Show this help message
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: link-git-hooks
link-git-hooks: ## Install git hooks
	@echo "==> Installing all git hooks..."
	find .git/hooks -type l -exec rm {} \;
	find .githooks -type f -exec ln -sf ../../{} .git/hooks/ \;