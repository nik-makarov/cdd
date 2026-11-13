VERSION ?= dev
BINDIR ?= $(HOME)/.local/bin
GOVULNCHECK_VERSION := v1.9.0

.PHONY: build fmt fmt-check test check audit install clean

build:
	go build -trimpath -ldflags "-s -w -X main.version=$(VERSION)" -o bin/0go .

test:
	go test ./...

fmt:
	gofmt -w *.go

fmt-check:
	@files="$$(gofmt -l *.go)"; \
	if [ -n "$$files" ]; then printf 'Run make fmt for:\n%s\n' "$$files"; exit 1; fi

check: fmt-check
	go vet ./...
	go test -race -shuffle=on ./...

audit:
	go mod verify
	go run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) ./...

install: build
	install -d "$(BINDIR)"
	install -m 755 bin/0go "$(BINDIR)/0go"

clean:
	rm -f bin/0go
