# Project instructions

The repository language is English only: documentation, comments, user-facing
messages, issue and pull request text, and commit messages. Use Go Unicode
escapes for non-ASCII test fixtures so Unicode behavior remains covered.

Use idiomatic Go. If the `golang-how-to` skill is available, load it for Go
coding, review, debugging, and setup work. The project must remain usable
without any personal agent skills or local tooling configuration.

Keep this small CLI simple. UI and diagnostics go to the terminal/stderr;
stdout is reserved for the selected path. Preserve shell integration tests,
especially cancellation and filenames containing shell metacharacters.

Minimize dependencies. Prefer the standard library; remaining external modules
must have a clear purpose and verifiable maintainers. Currently only the Go
project's golang.org/x/term and golang.org/x/sys are allowed. Review any proposed
addition and its entire transitive dependency graph before introducing it.

Use the latest stable Go release, with the minimum version in go.mod.
Run `make check build` before committing code changes. Run `make audit` when
dependencies change; report network failures instead of claiming a clean audit.
Keep documentation and examples free of machine-specific paths.

Docker images must only be published by repository CI/CD, never locally.
