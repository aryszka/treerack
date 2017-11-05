SOURCES = $(shell find . -name '*.go')
PARSERS = $(shell find . -name '*.treerack')

default: build

deps:
	go get golang.org/x/tools/cmd/goimports
	go get github.com/zalando/skipper/eskip

imports: $(SOURCES)
	@goimports -w $(SOURCES)

build: $(SOURCES)
	go build ./...

check: imports build $(PARSERS)
	go test -test.short -run ^Test

check-all: imports build $(PARSERS)
	go test

.coverprofile: $(SOURCES) imports
	go test -coverprofile .coverprofile

cover: .coverprofile
	go tool cover -func .coverprofile

show-cover: .coverprofile
	go tool cover -html .coverprofile

publish-coverage: .coverprofile
	curl -s https://codecov.io/bash -o codecov
	bash codecov -f .coverprofile

cpu.out: $(SOURCES) $(PARSERS)
	go test -v -run TestMMLFile -cpuprofile cpu.out

cpu: cpu.out
	go tool pprof -top cpu.out

fmt: $(SOURCES)
	@gofmt -w -s $(SOURCES)

precommit: fmt build check-all

clean:
	@rm -f *.test
	@rm -f cpu.out
	@go clean -i ./...

ci-trigger: deps build check-all
ifeq ($(TRAVIS_BRANCH)_$(TRAVIS_PULL_REQUEST), master_false)
	make publish-coverage
endif
