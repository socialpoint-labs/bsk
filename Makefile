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

install-tools-ci:
	go install github.com/GeertJohan/fgt@latest
	go install golang.org/x/tools/cmd/cover@latest
	go install golang.org/x/tools/cmd/goimports@latest
.PHONY: install-tools-ci

### Docker ###
up:
	docker-compose up --remove-orphans
.PHONY: up

up-build:
	docker-compose up --remove-orphans --build
.PHONY: up-build

up-daemon:
	docker-compose up --remove-orphans -d
.PHONY: up-daemon

down:
	docker-compose down -v
.PHONY: down

ps:
	docker-compose ps
.PHONY: ps

logs:
	docker-compose logs -f bsk
.PHONY: logs

restart:
	docker-compose restart bsk
.PHONY: restart

bash:
	docker-compose exec bsk bash
.PHONY: bash

docker-check:
	docker-compose exec bsk bash -c "make check"
.PHONY: docker-check

docker-ci-test:
	docker-compose exec -T bsk bash -c "make ci-test"
.PHONY: docker-ci-test
