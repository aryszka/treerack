SOURCES = $(shell find . -name '*.go')
PARSERS = $(shell find . -name '*.treerack')

default: build

deps:
	go get golang.org/x/tools/cmd/goimports
	go get -t ./...

imports: $(SOURCES)
	@goimports -w $(SOURCES)

build: $(SOURCES)
	go build ./...

check: imports build $(PARSERS)
	go test -test.short -run ^Test

check-full: imports build $(PARSERS)
	go test

.coverprofile: $(SOURCES) imports
	go test -coverprofile .coverprofile

cover: .coverprofile
	go tool cover -func .coverprofile

show-cover: .coverprofile
	go tool cover -html .coverprofile

publish-coverage: .coverprofile
	curl -s https://codecov.io/bash -o codecov
	bash codecov -Zf .coverprofile

cpu.out: $(SOURCES) $(PARSERS)
	go test -v -run TestMMLFile -cpuprofile cpu.out

cpu: cpu.out
	go tool pprof -top cpu.out

fmt: $(SOURCES)
	@gofmt -w -s $(SOURCES)

check-fmt: $(SOURCES)
	@if [ "$$(gofmt -s -d $(SOURCES))" != "" ]; then false; else true; fi

vet:
	@go vet

precommit: fmt build check-full

clean:
	@rm -f *.test
	@rm -f cpu.out
	@rm -f .coverprofile
	@go clean -i ./...

ci-trigger: deps check-fmt build check-full
ifeq ($(TRAVIS_BRANCH)_$(TRAVIS_PULL_REQUEST), master_false)
	make publish-coverage
endif
