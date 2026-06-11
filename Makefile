BINARY  := asd
GOBIN   := $(shell go env GOBIN)
ifeq ($(GOBIN),)
GOBIN   := $(shell go env GOPATH)/bin
endif
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: build install uninstall clean

## build: compile a local ./asd binary
build:
	go build $(LDFLAGS) -o $(BINARY) .

## install: install asd into your Go bin (run `fish_add_path $(GOBIN)` once)
install:
	go install $(LDFLAGS) .
	@echo "installed $(VERSION) -> $(GOBIN)/$(BINARY)"

## uninstall: remove the installed binary
uninstall:
	rm -f $(GOBIN)/$(BINARY)

## clean: remove the local build artifact
clean:
	rm -f $(BINARY)
