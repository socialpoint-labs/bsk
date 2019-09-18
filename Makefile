SOURCES=$(shell find . -name "*.go" | grep -v vendor/)
PACKAGES=$(shell go list ./...)
LINTER_VERSION=1.18.0

deps:
	go get -t -u ./...

test:
	go test ./...

test-ci:
	echo "mode: count" > coverage-all.out
	$(foreach pkg,$(PACKAGES),\
		GORACE="halt_on_error=1" go test -v -race -cover -coverprofile=coverage.out $(pkg) || exit 1;\
		tail -n +2 coverage.out >> coverage-all.out;)

install-tools:
	go get github.com/mattn/goveralls
	go get golang.org/x/tools/cmd/cover
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(GOPATH)/bin v$(LINTER_VERSION)

lint:
	golangci-lint run
	go mod verify

ci-check: test-ci
