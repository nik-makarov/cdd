# Project instructions

The repository language is English only: documentation, comments, user-facing
messages, issue and pull request text, and commit messages. Use Go Unicode
escapes for non-ASCII test fixtures so Unicode behavior remains covered.
Use concise imperative English commit subjects without Conventional Commits
prefixes. Describe the concrete change and keep unrelated work separate.

Use idiomatic Go. If the `golang-how-to` skill is available, load it for Go
coding, review, debugging, and setup work. The project must remain usable
without any personal agent skills or local tooling configuration.

Keep this small CLI simple. UI and diagnostics go to the terminal/stderr;
stdout is reserved for the selected path. Preserve shell integration tests,
especially cancellation and filenames containing shell metacharacters.
The application's scope is directory navigation and selection, creating a
directory as a navigation destination, explicit file preview/opening, Zed project
or single-file launching, and shell integration for cd. Do not add other file
management, background services, networking, or unrelated capabilities.
Features must directly support browsing and choosing a directory or the explicit
file actions listed above.
File contents may be read only for an explicitly requested preview. Opening a
file must use the operating system's default application or the configured Zed
binding on explicit input. Launch Zed only through its installed CLI.
Never execute a selected file directly, modify it inside cdd, or return it as a
selected directory. File actions preserve browsing state and the stdout contract.

Minimize dependencies. Prefer the standard library; remaining external modules
must have a clear purpose and verifiable maintainers. Currently only the Go
project's golang.org/x/term and golang.org/x/sys are allowed. Review any proposed
addition and its entire transitive dependency graph before introducing it.
Pin development tools in the separate tools module. Keep its dependencies out
of the application module and runtime license inventory. The notice generator
must use only the standard library and fail on missing license inputs.

Use the latest stable Go release, with the minimum version in go.mod.
Run `make check build` before committing code changes. Run `make audit` when
dependencies change; report network failures instead of claiming a clean audit.
Keep documentation and examples free of machine-specific paths.

Append a chronological English docs/development-journal.md entry with actual UTC
time, affected paths, and verified scope for every lasting logical change. Include
it in the same commit. Do not record plans or imply publication from local
checks. Keep the README concise and detailed documentation under docs/.
Keep CHANGELOG.md user-facing, with Unreleased changes and dated version sections.
Automatic releases start from VERSION and increment patch on successful main
pushes, except changes limited to Markdown files or docs/. Mixed changes release
normally. CI owns release tags; never create or publish them locally. Prefer a
curated CHANGELOG section for a known release, otherwise use clear English
commit subjects for incremental automatic notes. Release archives target only
Linux amd64/arm64 and macOS amd64/arm64; publishing belongs to GitHub Actions.

Docker images must only be published by repository CI/CD, never locally.
