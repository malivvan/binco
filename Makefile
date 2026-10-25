# Developer entry points.

GO        ?= go
PKGS      := $(shell $(GO) list ./...)
COVEROUT  ?= coverage.out

.PHONY: build test race vet fmt lint cover cover-check doc-check portable portable-quick

build:
	$(GO) build ./...

test:
	$(GO) test ./...

race:
	$(GO) test -race ./...

vet:
	$(GO) vet ./...

fmt:
	@gofmt -l .

lint:
	golangci-lint run --build-tags=fuzz

# Coverage of every package in the module.
cover:
	$(GO) test -coverprofile=$(COVEROUT) $(PKGS)
	$(GO) tool cover -func=$(COVEROUT) | tail -1

# Fail unless every library statement is covered. Statements that cannot run
# are listed in NOCOVER.md and treated as covered; everything else must count.
cover-check: cover
	@sh scripts/covercheck.sh $(COVEROUT) NOCOVER.md

# Prove that every target Go supports builds, and that the test files compile on
# a representative cross-section (both byte orders, 32-bit, wasm). Slow: CI runs
# it on every push, locally run it before a release.
portable:
	sh scripts/portable.sh

portable-quick:
	sh scripts/portable.sh --quick

# README.md and AGENTS.md are part of the contract: this checks their links,
# documented symbols, numbering and table of contents.
doc-check:
	python3 scripts/doccheck.py
