SOURCES = $(shell find . -name '*.go')

default: build

imports:
	@goimports -w $(SOURCES)

build: $(SOURCES)
	go build ./...

check: build
	go test ./... -test.short -run ^Test

fmt: $(SOURCES)
	@gofmt -w -s $(SOURCES)

precommit: fmt build check
