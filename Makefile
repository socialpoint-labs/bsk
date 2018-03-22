SOURCES=$(shell find . -name "*.go" | grep -v vendor/)
PACKAGES=$(shell go list ./...)

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
	go get github.com/GeertJohan/fgt
	go get github.com/mattn/goveralls
	go get golang.org/x/tools/cmd/cover
	go get golang.org/x/tools/cmd/goimports
	go get github.com/golang/lint/golint
	go get github.com/kisielk/errcheck
	go get honnef.co/go/tools/cmd/gosimple
	go get mvdan.cc/interfacer
	go get honnef.co/go/tools/cmd/staticcheck

lint:
	fgt go fmt $(PACKAGES)
	fgt goimports -w $(SOURCES)
	fgt golint $(PACKAGES)
	fgt go vet $(PACKAGES)
	fgt gosimple $(PACKAGES)
	fgt interfacer $(PACKAGES)
	# ignore deferred calls to io.Closer
	fgt errcheck -ignore Close $(PACKAGES)
	staticcheck $(PACKAGES)

ci-check: lint test-ci
