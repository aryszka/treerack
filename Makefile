SOURCES = $(shell find . -name '*.go')
PARSERS = $(shell find . -name '*.treerack')

default: build

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
	CODECOV_TOKEN=a2b0290b-62b6-419f-bf61-4182e479aec4 bash codecov -f .coverprofile

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
