GO ?= go
BINARY ?= bin/cms
CONFIG ?= config/config.yaml

.PHONY: build run dev test vet fmt fmt-check lint vuln migrate seed clean

build:
	$(GO) build -trimpath -ldflags "-s -w" -o $(BINARY) ./cmd/cms

run:
	$(GO) run ./cmd/cms -config $(CONFIG)

dev:
	$(GO) run ./cmd/cms -config $(CONFIG)

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

fmt:
	gofmt -w .

fmt-check:
	@out="$$(gofmt -l .)"; if [ -n "$$out" ]; then echo "$$out"; exit 1; fi

vuln:
	govulncheck ./...

lint: fmt-check vet vuln

migrate:
	$(GO) run ./cmd/cms -config $(CONFIG) -migrate

seed:
	$(GO) run ./cmd/cms -config $(CONFIG) -seed

clean:
	rm -rf bin
