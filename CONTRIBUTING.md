# Contributing

Bug reports and focused pull requests are welcome. Include your operating
system, shell, terminal emulator, Go version, and steps to reproduce the issue.
Replace private directory names and paths before posting terminal output.

Use English for documentation, comments, user-facing messages, issues, pull
requests, and commits. Write non-ASCII test fixture data using Go Unicode
escapes to preserve Unicode coverage while keeping source text in English.
Use concise imperative commit subjects describing the concrete change, without
Conventional Commits prefixes. Keep unrelated changes in separate commits.

## Development

Use the latest stable Go release. The minimum required version is in `go.mod`.
The application supports macOS and Linux with zsh and bash. Install both shells
to run the complete shell integration test suite.

```sh
make fmt
make check build
make audit
```

`make audit` uses the scanner pinned in [tools/](tools/README.md) and the public
Go vulnerability database. It requires network access and does not add tool
dependencies to the application's module files. Run `make cross-build` for
packaging and platform changes. See [verification](docs/testing.md) for manual
terminal acceptance and the limits of automated checks.

For changes to input decoding, also run the fuzz test:

```sh
go test -run '^$' -fuzz FuzzDecodeKey -fuzztime=10s
```

Keep changes small and add tests for meaningful behavior. Preserve cancellation,
terminal restoration, stdout isolation, and literal handling of directory names
containing whitespace, quotes, shell metacharacters, or newlines.

## Dependencies

Prefer the standard library. New dependencies need a concrete justification,
verified maintainers, a compatible license, and a review of their full transitive
dependency graph. Keep versions and checksums pinned in `go.mod` and `go.sum`.
Do not add test frameworks or CLI frameworks for features covered by Go itself.

Keep development tools in their separate module with reviewed versions and
checksums. Runtime dependency changes also require `make notices` and review of
the generated redistribution file; missing licenses must fail the build.

## Change journal

Append an English entry to [CHANGELOG.md](CHANGELOG.md) for each logical lasting
change, including documentation, tests, dependencies, agent instructions, and
tooling. Use the actual UTC timestamp in `YYYY-MM-DDTHH:MM:SSZ`, affected paths,
and verified scope. Preserve earlier entries and include the journal update in
the same commit. Distinguish local validation from publication or hosted CI.
Do not record plans, read-only findings, or routine checks without a lasting
change. Documentation changes need a link and command review.

Contributions are licensed under the project's [MIT license](LICENSE).
