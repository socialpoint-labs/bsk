PACKAGES=$(shell go list ./...)
LINTER_VERSION=1.53.3

lint:
	goimports -w .
	go mod verify
	golangci-lint run --fix
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

ci-test: export AWS_ACCESS_KEY_ID := x
ci-test: export AWS_SECRET_ACCESS_KEY := x
ci-test: export SP_BSK_AWS_ENDPOINT := http://localhost:4566
ci-test:
	echo "mode: count" > coverage-all.out
	$(foreach pkg,$(PACKAGES),\
		GORACE="halt_on_error=1" go test -race -cover -coverprofile=coverage.out $(pkg) || exit 1;\
		tail -n +2 coverage.out >> coverage-all.out;)
.PHONY: ci-test

install-tools:
	go install github.com/GeertJohan/fgt@latest
	go install golang.org/x/tools/cmd/cover@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v$(LINTER_VERSION)
.PHONY: install-tools

### Docker ###
up:
	docker-compose up --remove-orphans
.PHONY: up

up-build:
	docker-compose up --remove-orphans --build
.PHONY: up-build

up-daemon:
	docker-compose run --rm starter
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
