# Development tools

This separate module pins `golang.org/x/vuln/cmd/govulncheck` with a Go `tool`
directive. `go.mod` and `go.sum` record the scanner's version, complete module
requirements, and checksums. Dependabot proposes updates for this module.

Use `make audit` from the repository root. It verifies both modules, then runs
the scanner against the application and the standard-library-only notice
and release tools. No global executable or `PATH` change is needed. Go downloads
tools into its cache as necessary; vulnerability data is queried over the network.

The application module retains only its two runtime dependencies. Tool
dependencies are not application imports and are excluded from redistribution
notices unless application code actually imports them.

## Notice generator

`notices/` contains the project's own license collection command and tests.
It uses only the Go standard library. `make notices`, `make build`, and
`make cross-build` run it to create one `bin/THIRD_PARTY_LICENSES.txt`.
See [third-party notices](../THIRD_PARTY_NOTICES.md) for the packaging contract.

## Release packager

`release/` packages the four supported distribution targets, using only the
standard library. It validates versions and curated notes, includes licenses
and executable permissions, and writes SHA-256 checksums. `make dist` previews
Unreleased notes with VERSION=dev. `-next-version` computes the next automatic
version from VERSION and Git tags; `-automatic` also packages committed source,
generates the source-built Homebrew formula, and selects curated or incremental
release notes. Automatic packaging requires a clean committed checkout and
never publishes anything itself. See [releases](../docs/releases.md).

## Dependency updates

The GitHub **Update dependencies** workflow updates the Go release and both
modules, validates the candidate, and opens a PR. See
[maintenance](../docs/maintenance.md) for the button and required repository
settings. It keeps the application and tool module graphs separate.

To update the scanner deliberately:

```sh
cd tools
go get -tool golang.org/x/vuln/cmd/govulncheck@VERSION
go mod tidy
```

Replace `VERSION` with a reviewed release. Review both tool lock files and the
transitive dependency graph, then run the root verification commands. Do not
introduce additional linters or frameworks without a concrete need.
