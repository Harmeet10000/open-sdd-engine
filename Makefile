GO ?= go
GO_MODULE_FLAGS ?= -mod=readonly
MIN_COVERAGE ?= 70
GOLANGCI_LINT_VERSION ?= v2.14.0
GOVULNCHECK_VERSION ?= v1.9.0
GOSEC_VERSION ?= v2.29.0
ACTIONLINT_VERSION ?= v1.7.12

GO_FILES := $(shell git ls-files '*.go'; git ls-files --others --exclude-standard '*.go')

.DEFAULT_GOAL := help

.PHONY: actionlint ci commitlint coverage fmt fmt-check help install-tools \
	lint list mod-check race security test test-current vet vuln

help: ## Show the supported development commands.
	@awk 'BEGIN {FS = ":.*## "; printf "Usage: make <target>\n\nTargets:\n"} /^[a-zA-Z0-9_-]+:.*## / {printf "  %-18s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# The repository's top-level vendor/ directory is an upstream research corpus,
# not a Go module vendor tree. Explicit module mode prevents Go from treating
# that evidence directory as generated dependency state or overwriting it.
list: ## Verify all module and package imports resolve without mutation.
	$(GO) list $(GO_MODULE_FLAGS) -m all
	$(GO) list $(GO_MODULE_FLAGS) ./...

fmt: ## Format every tracked or untracked Go source file.
	gofmt -w $(GO_FILES)

fmt-check: ## Fail when committed Go source is not gofmt-clean.
	@test -z "$$(gofmt -l $(GO_FILES))" || { \
		echo "Go files require formatting:"; \
		gofmt -l $(GO_FILES); \
		exit 1; \
	}

mod-check: ## Prove go.mod and go.sum are already tidy.
	GOFLAGS=-mod=mod $(GO) mod tidy -diff
	$(GO) mod verify

test: ## Run deterministic unit and integration tests in shuffled order.
	$(GO) test $(GO_MODULE_FLAGS) -shuffle=on ./...

test-current: ## Run tests with additional runtime diagnostics.
	$(GO) test $(GO_MODULE_FLAGS) -count=1 -shuffle=on ./...

race: ## Run the full Go suite with the race detector.
	$(GO) test $(GO_MODULE_FLAGS) -race -count=1 -shuffle=on ./...

coverage: ## Enforce the repository's minimum statement coverage.
	$(GO) test $(GO_MODULE_FLAGS) -covermode=atomic -coverprofile=coverage.out ./...
	@coverage="$$(GOFLAGS='$(GO_MODULE_FLAGS)' $(GO) tool cover -func=coverage.out | awk '/^total:/ {gsub("%", "", $$3); print $$3}')"; \
	awk -v coverage="$$coverage" -v minimum="$(MIN_COVERAGE)" 'BEGIN { \
		printf "Total coverage: %.1f%% (minimum %.1f%%)\n", coverage, minimum; \
		if (coverage + 0 < minimum + 0) exit 1; \
	}'

vet: ## Run the Go toolchain's correctness analyzer.
	$(GO) vet $(GO_MODULE_FLAGS) ./...

lint: ## Run the pinned golangci-lint policy.
	GOFLAGS=$(GO_MODULE_FLAGS) golangci-lint run ./...

vuln: ## Report reachable vulnerabilities in Go call paths.
	GOFLAGS=$(GO_MODULE_FLAGS) govulncheck ./...

security: ## Run source-level security checks.
	GOFLAGS=$(GO_MODULE_FLAGS) gosec -quiet $$(GOFLAGS=$(GO_MODULE_FLAGS) $(GO) list -f '{{.Dir}}' ./...)

actionlint: ## Validate all GitHub Actions workflows.
	actionlint

commitlint: ## Validate commits since COMMIT_BASE (default: origin/main).
	@base="$${COMMIT_BASE:-origin/main}"; \
	merge_base="$$(git merge-base "$$base" HEAD)"; \
	$(GO) run $(GO_MODULE_FLAGS) ./internal/tools/commitlint -base "$$merge_base" -head HEAD

install-tools: ## Install exactly the analysis tool versions used by CI.
	GOBIN="$$(go env GOPATH)/bin" $(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	GOBIN="$$(go env GOPATH)/bin" $(GO) install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)
	GOBIN="$$(go env GOPATH)/bin" $(GO) install github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION)
	GOBIN="$$(go env GOPATH)/bin" $(GO) install github.com/rhysd/actionlint/cmd/actionlint@$(ACTIONLINT_VERSION)

ci: fmt-check mod-check list vet test race coverage lint security actionlint ## Reproduce required CI checks locally.
