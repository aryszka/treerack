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

cpu.out: $(SOURCES)
	go test -v -run TestMMLFile -cpuprofile cpu.out

cpu: cpu.out
	go tool pprof -top cpu.out

precommit: fmt build check

clean:
	@rm -f *.test
	@rm -f cpu.out
	@go clean -i ./...
