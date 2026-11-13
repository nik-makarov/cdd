.DEFAULT_GOAL := help

VERSION ?= dev
RELEASE_FLAGS ?=
PREFIX ?= $(HOME)/.local
BINDIR ?= $(PREFIX)/bin
LICENSEDIR ?= $(PREFIX)/share/licenses/cdd

.PHONY: help build notices cross-build dist fmt fmt-check test check audit install clean

help:
	@printf '%s\n' \
		'cdd development commands (run from the source root):' \
		'  make build        Build the host binary and license files' \
		'  make cross-build  Build Linux and macOS amd64/arm64 binaries' \
		'  make dist         Package local preview archives under dist/dev/' \
		'  make test         Run application and build-tool tests' \
		'  make check        Check formatting, vet, and race-enabled tests' \
		'  make audit        Verify modules and scan known vulnerabilities' \
		'  make fmt          Format Go source' \
		'  make notices      Generate redistribution license files' \
		'  make install      Install the binary and licenses under PREFIX' \
		'  make clean        Remove generated binaries, archives, and notices' \
		'' \
		'Use VERSION=vX.Y.Z for a versioned build or archive with matching CHANGELOG notes.'

build: notices
	go build -trimpath -ldflags "-s -w -X main.version=$(VERSION)" -o bin/cdd .

notices:
	go -C tools run ./notices -root ..
	cp LICENSE bin/LICENSE

cross-build: notices
	@set -eu; for target in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64; do \
		CGO_ENABLED=0 GOOS="$${target%/*}" GOARCH="$${target#*/}" go build -trimpath \
			-ldflags "-s -w -X main.version=$(VERSION)" -o "bin/cdd-$${target%/*}-$${target#*/}" .; \
	done

dist: cross-build
	go -C tools run ./release -root .. -version "$(VERSION)" $(RELEASE_FLAGS)

test:
	go test ./...
	go -C tools test ./...

fmt:
	gofmt -w .

fmt-check:
	@files="$$(gofmt -l .)"; \
	if [ -n "$$files" ]; then printf 'Run make fmt for:\n%s\n' "$$files"; exit 1; fi

check: fmt-check
	go vet ./...
	go test -race -shuffle=on ./...
	go -C tools vet ./...
	go -C tools test -race -shuffle=on ./...

audit:
	go mod verify
	go -C tools mod verify
	go tool -modfile=tools/go.mod govulncheck ./...
	go -C tools tool govulncheck ./...

install: build
	install -d "$(BINDIR)"
	install -m 755 bin/cdd "$(BINDIR)/cdd"
	install -d "$(LICENSEDIR)"
	install -m 644 bin/LICENSE bin/THIRD_PARTY_LICENSES.txt "$(LICENSEDIR)/"

clean:
	rm -f bin/cdd
	rm -f bin/cdd-darwin-amd64 bin/cdd-darwin-arm64 bin/cdd-linux-amd64 bin/cdd-linux-arm64
	rm -f bin/LICENSE bin/THIRD_PARTY_LICENSES.txt
	rm -rf dist
