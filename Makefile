GO ?= go
GO_MODULE_FLAGS ?= -mod=readonly

.PHONY: list test race vet verify

# The repository's top-level vendor/ directory is an upstream research corpus,
# not a Go module vendor tree. Explicit module mode prevents Go from treating
# that evidence directory as generated dependency state or overwriting it.
list:
	$(GO) list $(GO_MODULE_FLAGS) -m all
	$(GO) list $(GO_MODULE_FLAGS) ./...

test:
	$(GO) test $(GO_MODULE_FLAGS) ./...

race:
	$(GO) test $(GO_MODULE_FLAGS) -race ./...

vet:
	$(GO) vet $(GO_MODULE_FLAGS) ./...

verify: list vet test race
