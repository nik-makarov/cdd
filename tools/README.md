# Development tools

This separate module pins `golang.org/x/vuln/cmd/govulncheck` with a Go `tool`
directive. `go.mod` and `go.sum` record the scanner's version, complete module
requirements, and checksums. Dependabot proposes updates for this module.

Use `make audit` from the repository root. It verifies both modules, then runs
the scanner against the application and the standard-library-only notice
generator. No global executable or `PATH` change is needed. Go downloads tools
into its cache as necessary; vulnerability data is queried over the network.

The application module retains only its two runtime dependencies. Tool
dependencies are not application imports and are excluded from redistribution
notices unless application code actually imports them.

## Notice generator

`notices/` contains the project's own license collection command and tests.
It uses only the Go standard library. `make notices`, `make build`, and
`make cross-build` run it to create one `bin/THIRD_PARTY_LICENSES.txt`.
See [third-party notices](../THIRD_PARTY_NOTICES.md) for the packaging contract.

To update the scanner deliberately:

```sh
cd tools
go get -tool golang.org/x/vuln/cmd/govulncheck@VERSION
go mod tidy
```

Replace `VERSION` with a reviewed release. Review both tool lock files and the
transitive dependency graph, then run the root verification commands. Do not
introduce additional linters or frameworks without a concrete need.
