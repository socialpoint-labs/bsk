PACKAGES=$(shell go list ./...)
LINTER_VERSION=1.18.0

lint:
	goimports -w .
	go mod verify
	golangci-lint run
.PHONY: lint

test:
	go test ./...
.PHONY: test

ci-lint:
	fgt goimports -l .
	go mod verify
	golangci-lint run
.PHONY: ci-lint

ci-test:
	echo "mode: count" > coverage-all.out
	$(foreach pkg,$(PACKAGES),\
		GORACE="halt_on_error=1" go test -v -race -cover -coverprofile=coverage.out $(pkg) || exit 1;\
		tail -n +2 coverage.out >> coverage-all.out;)
.PHONY: ci-test

ci-check: ci-lint ci-test
.PHONY: ci-check

install-tools:
	go get -u github.com/GeertJohan/fgt
	go get -u golang.org/x/tools/cmd/cover
	go get -u golang.org/x/tools/cmd/goimports
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(GOPATH)/bin v$(LINTER_VERSION)
.PHONY: install-tools
