PACKAGES=$(shell go list ./...)
LINTER_VERSION=1.42.1

lint:
	goimports -w .
	go mod verify
	golangci-lint run
.PHONY: lint

test:
	GORACE="halt_on_error=1" go test -race ./...
.PHONY: test

test-integration:
	GORACE="halt_on_error=1" go test -tags=integration -race ./...
.PHONY: test-integration

ci-lint:
	fgt goimports -l .
	go mod verify
	golangci-lint run
.PHONY: ci-lint

ci-test:
	echo "mode: count" > coverage-all.out
	$(foreach pkg,$(PACKAGES),\
		GORACE="halt_on_error=1" go test -race -cover -coverprofile=coverage.out $(pkg) || exit 1;\
		tail -n +2 coverage.out >> coverage-all.out;)
.PHONY: ci-test

ci-check: ci-lint ci-test
.PHONY: ci-check

install-tools:
	go install github.com/GeertJohan/fgt@latest
	go install golang.org/x/tools/cmd/cover@latest
	go install golang.org/x/tools/cmd/goimports@latest
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(GOPATH)/bin v$(LINTER_VERSION)
.PHONY: install-tools
