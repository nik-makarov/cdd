VERSION ?= dev
PREFIX ?= $(HOME)/.local
BINDIR ?= $(PREFIX)/bin
LICENSEDIR ?= $(PREFIX)/share/licenses/0go

.PHONY: build notices cross-build fmt fmt-check test check audit install clean

build: notices
	go build -trimpath -ldflags "-s -w -X main.version=$(VERSION)" -o bin/0go .

notices:
	go -C tools run ./notices -root ..
	cp LICENSE bin/LICENSE

cross-build: notices
	@set -eu; for target in darwin/amd64 darwin/arm64 linux/amd64 linux/arm64; do \
		CGO_ENABLED=0 GOOS="$${target%/*}" GOARCH="$${target#*/}" go build -trimpath \
			-ldflags "-s -w -X main.version=$(VERSION)" -o "bin/0go-$${target%/*}-$${target#*/}" .; \
	done

test:
	go test ./...
	go -C tools test ./notices

fmt:
	gofmt -w .

fmt-check:
	@files="$$(gofmt -l .)"; \
	if [ -n "$$files" ]; then printf 'Run make fmt for:\n%s\n' "$$files"; exit 1; fi

check: fmt-check
	go vet ./...
	go test -race -shuffle=on ./...
	go -C tools vet ./notices
	go -C tools test -race -shuffle=on ./notices

audit:
	go mod verify
	go -C tools mod verify
	go tool -modfile=tools/go.mod govulncheck ./...
	go -C tools tool govulncheck ./notices

install: build
	install -d "$(BINDIR)"
	install -m 755 bin/0go "$(BINDIR)/0go"
	install -d "$(LICENSEDIR)"
	install -m 644 bin/LICENSE bin/THIRD_PARTY_LICENSES.txt "$(LICENSEDIR)/"

clean:
	rm -f bin/0go
	rm -f bin/0go-darwin-amd64 bin/0go-darwin-arm64 bin/0go-linux-amd64 bin/0go-linux-arm64
	rm -f bin/LICENSE bin/THIRD_PARTY_LICENSES.txt
