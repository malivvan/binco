# Developer entry points.
#
# Every target except `demos` works on the library packages only. The demos are
# example programs, not library code, and are excluded from coverage on purpose
# (see NOCOVER.md).

GO        ?= go
DEMO      := /demos
PKGS      := $(shell $(GO) list ./... | grep -v '$(DEMO)')
COVEROUT  ?= coverage.out

.PHONY: build test race vet fmt generate lint cover cover-check demos

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

generate:
	$(GO) generate ./...

lint:
	golangci-lint run --build-tags=fuzz

# Coverage of every library package. If the package list ever contains a demo
# path the target fails instead of silently measuring example code.
cover:
	@if echo '$(PKGS)' | grep -q '$(DEMO)'; then \
		echo "demos must be excluded from coverage"; exit 1; \
	fi
	$(GO) test -coverprofile=$(COVEROUT) $(PKGS)
	$(GO) tool cover -func=$(COVEROUT) | tail -1

# Fail unless every library statement is covered. Statements that cannot run
# are listed in NOCOVER.md and treated as covered; everything else must count.
cover-check: cover
	@sh scripts/covercheck.sh $(COVEROUT) NOCOVER.md

demos:
	$(GO) build ./demos/...
