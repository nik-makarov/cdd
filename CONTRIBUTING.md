# Contributing

Bug reports and focused pull requests are welcome. Include your operating
system, shell, terminal emulator, Go version, and steps to reproduce the issue.
Replace private directory names and paths before posting terminal output.

## Development

Use the latest stable Go release. The minimum required version is in `go.mod`.
The application supports macOS and Linux with zsh and bash. Install both shells
to run the complete shell integration test suite.

```sh
make fmt
make check build
make audit
```

`make audit` downloads a pinned Go vulnerability scanner into Go's cache and
uses the public Go vulnerability database. It requires network access and does
not add dependencies to the application's module files.

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

Contributions are licensed under the project's [MIT license](LICENSE).
