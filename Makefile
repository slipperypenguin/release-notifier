DIST := dist
BIN := bin
EXECUTABLE := release-notifier

PWD := $(shell pwd)
VERSION := $(shell cat VERSION)
SHA := $(shell cat COMMIT 2>/dev/null || git rev-parse --short=8 HEAD)
DATE := $(shell date -u '+%FT%T%z')

GOLDFLAGS += -X "main.version=$(VERSION)"
GOLDFLAGS += -X "main.date=$(DATE)"
GOLDFLAGS += -X "main.commit=$(SHA)"
GOLDFLAGS += -extldflags '-static'

GO := CGO_ENABLED=0 go
GOARM := CGO_ENABLED=0 GOARCH=arm GOARM=7 go

GOOS ?= $(shell go version | cut -d' ' -f4 | cut -d'/' -f1)
GOARCH ?= $(shell go version | cut -d' ' -f4 | cut -d'/' -f2)
PACKAGES ?= $(shell go list ./... | grep -v /vendor/ | grep -v /tests)
TAGS ?= netgo

.PHONY: all
all: clean test build

.PHONY: clean
clean:
	$(GO) clean -i ./...
	find . -type f -name "coverage.out" -delete

.PHONY: lint
lint:
	@which golint > /dev/null; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/golang/lint/golint; \
	fi
	STATUS=0; for PKG in $(PACKAGES); do golint -set_exit_status $$PKG || STATUS=1; done; exit $$STATUS

.PHONY: test
test:
	STATUS=0; for PKG in $(PACKAGES); do go test -cover -coverprofile $$GOPATH/src/$$PKG/coverage.out $$PKG || STATUS=1; done; exit $$STATUS

.PHONY: tests
tests: test lint

.PHONY: build
build: buildlinux buildarm

.PHONY: buildlinux
buildlinux: $(EXECUTABLE)-$(GOOS)-$(GOARCH)
$(EXECUTABLE)-$(GOOS)-$(GOARCH): $(wildcard *.go)
	$(GO) build -tags '$(TAGS)' -ldflags '-s -w $(GOLDFLAGS)' -o $(EXECUTABLE)

.PHONY: buildarm
buildarm: $(EXECUTABLE)-$(GOOS)
$(EXECUTABLE)-$(GOOS): $(wildcard *.go)
	$(GOARM) build -tags '$(TAGS)' -ldflags '-s -w $(GOLDFLAGS)' -o $(EXECUTABLE)