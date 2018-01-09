SOURCES = $(shell find . -name '*.go')
PARSERS = $(shell find . -name '*.treerack')

default: build

deps:
	go get golang.org/x/tools/cmd/goimports
	go get -t ./...

imports: $(SOURCES)
	@echo imports
	@goimports -w $(SOURCES)

build: $(SOURCES)
	go build
	go build -o cmd/treerack/treerack ./cmd/treerack

install: $(SOURCES)
	go install ./cmd/treerack

head: $(SOURCES) fmt
	go run scripts/createhead.go -- \
		char.go \
		sequence.go \
		choice.go \
		idset.go \
		results.go \
		context.go \
		nodehead.go \
		syntaxhead.go \
	> head.go
	@gofmt -s -w head.go

generate: $(SOURCES) $(PARSERS) fmt head install
	treerack generate -export -package-name self < syntax.treerack > self/self.go.next
	@mv self/self.go{.next,}
	@gofmt -s -w self/self.go

regenerate: $(SOURCES) $(PARSERS) fmt head install
	treerack generate -export -package-name self < syntax.treerack > self/self.go.next
	@mv self/self.go{.next,}
	treerack generate -export -package-name self < syntax.treerack > self/self.go.next
	@mv self/self.go{.next,}
	@gofmt -s -w self/self.go

check-generate: $(SOURCES) $(PARSERS)
	@echo checking head
	@mv head.go head.go.backup
	@go run scripts/createhead.go -- \
		char.go \
		sequence.go \
		choice.go \
		idset.go \
		results.go \
		context.go \
		nodehead.go \
		syntaxhead.go \
	> head.go
	@gofmt -s -w head.go
	@if ! diff head.go head.go.backup > /dev/null; then \
		mv head.go.backup head.go; \
		echo head does not match; \
		false; \
	fi
	@echo checking self
	@mv self/self.go self/self.go.backup
	@treerack generate -export -package-name self < syntax.treerack > self/self.go.next
	@mv self/self.go{.next,}
	@gofmt -s -w self/self.go
	@if ! diff self/self.go self/self.go.backup > /dev/null; then \
		mv self/self.go.backup self/self.go; \
		echo self does not match; \
		false; \
	fi

	@echo ok
	@mv head.go.backup head.go
	@mv self/self.go.backup self/self.go

check: imports build $(PARSERS)
	go test -test.short -run ^Test
	go test ./cmd/treerack -test.short -run ^Test

checkall: imports build $(PARSERS)
	go test
	go test ./cmd/treerack

.coverprofile: $(SOURCES) imports
	go test -coverprofile .coverprofile

cover: .coverprofile
	go tool cover -func .coverprofile

showcover: .coverprofile
	go tool cover -html .coverprofile

.coverprofile-cmd: $(SOURCES) imports
	go test ./cmd/treerack -coverprofile .coverprofile-cmd

cover-cmd: .coverprofile-cmd
	go tool cover -func .coverprofile-cmd

showcover-cmd: .coverprofile-cmd
	go tool cover -html .coverprofile-cmd

# command line interface not included
publishcoverage: .coverprofile
	curl -s https://codecov.io/bash -o codecov
	bash codecov -Zf .coverprofile

cpu.out: $(SOURCES) $(PARSERS)
	go test -v -run TestMMLFile -cpuprofile cpu.out

cpu: cpu.out
	go tool pprof -top cpu.out

fmt: $(SOURCES)
	@echo fmt
	@gofmt -w -s $(SOURCES)

checkfmt: $(SOURCES)
	@echo check fmt
	@if [ "$$(gofmt -s -d $(SOURCES))" != "" ]; then false; else true; fi

vet:
	go vet ./...

precommit: fmt check-generate vet build checkall

clean:
	rm -f *.test
	rm -f cpu.out
	rm -f .coverprofile
	go clean -i ./...

ci-trigger: deps checkfmt build checkall
ifeq ($(TRAVIS_BRANCH)_$(TRAVIS_PULL_REQUEST), master_false)
	make publishcoverage
endif
