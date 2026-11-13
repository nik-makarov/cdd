# Change journal

Record actual project changes chronologically using UTC timestamps. Each entry
describes a logical change, affected paths, and verified scope. Preserve previous
entries, keep private data out, and distinguish local work from publication.

- `2026-11-13T16:00:56Z` — Prepare the initial public source repository under
  `github.com/nik-makarov/0go`: interactive directory navigation, literal zsh/bash
  path handling, minimal Go runtime dependencies, MIT license, public English
  documentation, and SHA-pinned GitHub checks. Local formatting, vet,
  race-enabled tests, host and Linux builds, module verification, and workflow
  validation passed. The owner's separate audit reported no vulnerabilities;
  the agent environment could not reach the vulnerability database. Hosted CI,
  GitHub settings, and publication were not executed. Paths: root Go source and
  tests, `go.mod`, `go.sum`, `Makefile`, `LICENSE`, `.github/`, and public docs.
- `2026-11-13T16:06:01Z` — Require English repository text and represent
  non-ASCII test data with Unicode escapes while preserving Unicode coverage.
  Formatting, vet, race-enabled tests, and the host build passed. Local change;
  no remote publication. Paths: `AGENTS.md`, `CONTRIBUTING.md`,
  `browser_test.go`, `shell_test.go`, `terminal_test.go`.
- `2026-11-13T16:18:44Z` — Align public documentation and packaging with the
  owner's repository format: shorten the README, add usage/architecture/testing
  guides and a chronological journal, and document GitHub private reporting as a
  separate owner setting. Pin govulncheck and its checksums in a tools module,
  preserving the application's two runtime dependencies. Add a standard-library
  notice generator covering macOS/Linux amd64/arm64 application modules and the
  Go standard library, copying complete license/NOTICE/patent inputs and rejecting
  missing or empty licenses. Builds and installation include the MIT license and
  one generated redistribution file. Update CI, dependency updates, and contribution
  rules. Formatting, vet, race-enabled application/generator tests, all four
  cross-builds, isolated installation, shell syntax checks, relative links, module
  checksums, and actionlint without ShellCheck passed. Both vulnerability scans
  stopped on unavailable database DNS; hosted CI, GitHub settings, publication,
  and cross-platform runtime testing were not performed. Paths: `README.md`,
  `docs/`, `CHANGELOG.md`, `tools/`, `Makefile`, `.github/`, `AGENTS.md`,
  `CONTRIBUTING.md`, `SECURITY.md`, `THIRD_PARTY_NOTICES.md`.
- `2026-11-13T16:37:32Z` — Add a manually dispatched GitHub workflow that selects
  stable Go, updates both Go directives and existing application/tool modules,
  validates the candidate, and opens a reviewable PR with a UTC journal entry.
  Stop when an update PR is already open, skip unchanged candidates, and reject
  runtime module-path changes. Use the built-in repository token without new
  Actions or automatic merging; each PR gets a new branch without force-pushing.
  Document the button, required GitHub settings, PR CI approval, and review scope;
  preserve Dependabot for Actions updates and formalize imperative English commit
  subjects. Actionlint without ShellCheck and Bash syntax checks passed. An
  isolated cached-proxy test verified upgrades, the no-change case, and the graph
  guard; the upgraded candidate passed vet, race-enabled tests, host packaging,
  and four cross-builds. Checksum verification passed, but database DNS prevented
  a clean vulnerability scan. Hosted execution, permission changes, Git push, and
  PR creation were not performed; working module versions remain unchanged.
  Paths: `.github/workflows/update-dependencies.yml`, `docs/maintenance.md`,
  `docs/testing.md`, `README.md`, `tools/README.md`, `AGENTS.md`, `CONTRIBUTING.md`.
