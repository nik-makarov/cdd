# Development journal

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
- `2026-11-13T17:14:54Z` — Prepare portable releases for Linux amd64/arm64 and
  macOS arm64 with deterministic archives, executable permissions, bundled
  licenses/instructions, SHA-256 checksums, and curated version-specific notes.
  Require a matching dated CHANGELOG section before versioned packaging and run
  reusable CI before draft upload and publication on stable tags. Move this
  engineering journal from CHANGELOG.md while preserving earlier entries; keep
  user-facing release notes in CHANGELOG.md and update dependency automation to
  write here. Make plain `make` print help, record the README demo in TODO.md,
  and enforce directory-selection scope with minimal dependencies. Vet,
  race-enabled application/tool tests, three cross-builds, preview packaging,
  checksums, extracted native-binary shell integration, Make help isolation,
  relative links, workflow Bash syntax, and actionlint without ShellCheck passed.
  Module checksums passed; vulnerability database DNS remained unavailable.
  No hosted CI, tag, draft/public release, signing, or notarization was performed.
  Paths: `Makefile`, `.github/`, `tools/release/`, `tools/notices/main.go`,
  `packaging/`, `CHANGELOG.md`, `docs/`, `TODO.md`, `README.md`, `AGENTS.md`,
  `CONTRIBUTING.md`, `THIRD_PARTY_NOTICES.md`, `tools/README.md`.
- `2026-11-13T17:36:01Z` — Refine the directory picker with a theme-aware current
  path, reverse-video full-row selection, muted secondary text, cyan directory
  suffixes, labeled red errors, a stable compact footer, and expandable keyboard
  help. Preserve literal filter input, stdout path output, and a visible text
  selection marker. Honor non-empty NO_COLOR values while retaining non-color
  emphasis; disable styling for TERM=dumb. Clip sanitized text before applying
  ANSI styles and keep the selected row visible in tiny windows. Add no runtime
  or build-tool dependencies. Formatting, vet, shuffled race-enabled tests, host
  packaging, and three cross-builds passed. Native PTY checks exercised help,
  filtering, selection, cancellation, color suppression, and a real zsh directory
  change through captured stdout. Theme screenshot review, real window resizing,
  Linux runtime behavior, hosted CI, and publication were not performed. Paths:
  `browser.go`, `browser_test.go`, `view.go`, `view_test.go`, `terminal.go`,
  `cli.go`, `README.md`, `CHANGELOG.md`, `docs/usage.md`, `docs/architecture.md`,
  `docs/testing.md`, `docs/development-journal.md`.
- `2026-11-13T17:39:06Z` — Add Intel macOS (darwin/amd64) to cross-builds,
  release archives, checksum tests, and the license inventory alongside Linux
  amd64/arm64 and Apple Silicon macOS. Align public instructions, project target
  policy, and dependency-update workflow summaries with the four-target matrix.
  Formatting, vet, race-enabled application/tool tests, host packaging, all four
  cross-builds, and preview archive generation passed. All four archive checksums
  verified; the Intel binary is Mach-O x86_64 with the expected Go platform
  metadata, and its archive has executable permissions, instructions, notes, and
  licenses. No dependencies changed. Intel runtime testing, hosted CI, and
  publication were not performed. Paths: `Makefile`, `tools/release/`,
  `tools/notices/main.go`, `.github/workflows/update-dependencies.yml`,
  `AGENTS.md`, `README.md`, `CHANGELOG.md`, `THIRD_PARTY_NOTICES.md`,
  `tools/README.md`, `docs/`.
- `2026-11-13T17:46:39Z` — Reduce executable size by replacing the application's
  direct fmt/flag usage with a small compatible argument parser, io.WriteString,
  strconv, and contextual errors implementing Unwrap. Preserve option syntax,
  error messages, sentinel/typed causes, stdout discipline, and shell behavior.
  At Go 1.27.4 with identical stripped cross-build flags and VCS metadata disabled,
  executables shrank by 192.00-248.16 KiB across the four targets; fmt remains a
  transitive import of the unchanged terminal modules. Formatting, vet,
  race-enabled application/tool tests, four cross-builds, packaging, and archive
  checksum checks passed. The parser matched standard flag in table tests and
  4,174,566 fuzz cases; a native PTY zsh integration check changed directory with
  the existing argument forms. No module versions changed. No hosted CI or
  publication ran. Paths: `cli.go`, `cli_test.go`, `errors.go`, `errors_test.go`,
  `browser.go`, `terminal.go`, `view.go`, `CHANGELOG.md`, `docs/architecture.md`,
  `docs/testing.md`, `docs/development-journal.md`.
- `2026-11-13T17:57:59Z` — Report successful shell directory changes as an
  original-to-resulting path notification on stderr, using builtin printf with
  shell escaping to keep control characters and newlines out of the terminal.
  Preserve stdout, cancellation/failure status, and silent unchanged or empty
  selections; retain the shell's own error when cd fails. Document that existing
  shell functions must be refreshed after a binary update. Formatting, vet,
  race-enabled application/tool tests, and host packaging passed. Actual zsh/bash
  subprocesses checked notifications, hostile-looking names, terminal controls,
  stdout isolation, and failure/silence cases. A native PTY check confirmed the
  message, directory change, and zero exit status. No dependencies changed;
  hosted CI and publication did not run. Paths: `shell.go`, `shell_test.go`,
  `CHANGELOG.md`, `docs/usage.md`, `docs/testing.md`, `docs/development-journal.md`.
- `2026-11-13T18:15:08Z` — Display long directory lists in two columns when each
  has at least 24 cells, retaining a single column for short lists and narrow
  windows. Keep one ordered list/cursor, use Tab or Shift+Tab to switch columns,
  and page through their combined capacity. Preserve Right/Left directory
  navigation, filtering, parent focus, and selection through resize. Anchor the
  divider independently of Unicode widths and keep clipping/style boundaries
  safe. Formatting, vet, race-enabled application/tool tests, host packaging,
  four cross-builds, and 2,859,871 decoder fuzz cases passed. A native PTY check
  browsed 40 synthetic directories, switched columns, paged, resized 60/40/60
  cells without losing selection, and selected a right-column directory through
  zsh with exit status zero. No dependencies changed. Theme screenshot review,
  Linux runtime testing, hosted CI, and publication were not performed. Paths:
  `browser.go`, `browser_test.go`, `view.go`, `view_test.go`, `terminal.go`,
  `terminal_test.go`, `cli.go`, `README.md`, `CHANGELOG.md`, `docs/usage.md`,
  `docs/architecture.md`, `docs/testing.md`, `docs/development-journal.md`.
- `2026-11-13T18:33:04Z` — Add default parent/current/child tree browsing and
  switch to the existing compact directory-only view with v. Keep selection in
  the middle pane, cache contextual listings without reading file contents, and
  preserve folder focus across modes and resizing. Show files in yellow and
  directory symlinks in magenta with distinct suffixes; files can be highlighted
  but never opened or chosen as a destination. Preserve logical directory-link
  navigation and isolate preview read failures. Extend mode, entry, layout,
  color, sanitization, and file-action tests, keeping prior compact navigation
  coverage. Formatting, vet, race-enabled application/tool tests, host packaging,
  and four cross-builds passed. A native PTY check verified file rejection,
  child previews, link navigation, mode switching, 100/40/100-cell resizing, and
  a successful zsh directory change through a symlink. No dependencies changed.
  Screenshot review across themes, Linux runtime testing, hosted CI, and
  publication were not performed. Paths: `entry.go`, `entry_test.go`, `tree.go`,
  `tree_test.go`, `browser.go`, `browser_test.go`, `view.go`, `view_test.go`,
  `cli.go`, `README.md`, `CHANGELOG.md`, `AGENTS.md`, `docs/usage.md`,
  `docs/architecture.md`, `docs/testing.md`, `docs/development-journal.md`.
- `2026-11-13T18:40:49Z` — Clear Children for the current-directory row instead
  of copying the Current listing into its preview. Keep this pane blank at
  startup, after entering a folder, returning to `.`, refreshing, and switching
  views; retain previews for actual highlighted child directories. Regression
  tests reproduced all five failure paths before the fix. Formatting, vet,
  race-enabled application/tool tests, host packaging, and targeted rendering
  checks passed. A native PTY check confirmed preview population before opening
  a child and clearing immediately afterward. No dependencies changed; hosted
  CI and publication were not performed. Paths: `browser.go`, `browser_test.go`,
  `tree.go`, `tree_test.go`, `CHANGELOG.md`, `docs/usage.md`, `docs/testing.md`,
  `docs/development-journal.md`.
