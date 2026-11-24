# Development journal

Record actual project changes chronologically using UTC timestamps. Each entry
describes a logical change, affected paths, and verified scope. Preserve previous
entries, keep private data out, and distinguish local work from publication.

- `2026-11-13T16:00:56Z` — Prepare the initial public source repository under
  its original GitHub module path: interactive directory navigation, literal
  zsh/bash path handling, minimal Go runtime dependencies, MIT license, public English
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
- `2026-11-13T18:42:02Z` — Render files and non-directory links in the theme's
  dim foreground instead of yellow, keeping folders cyan and directory links
  magenta. Preserve reverse selection, link markers, and non-color emphasis
  under NO_COLOR. Update existing style assertions and usage/release notes.
  Formatting, vet, race-enabled application/tool tests, and host packaging
  passed. No dependencies changed or hosted CI/publication ran. Visual review
  across terminal themes was not performed. Paths: `view.go`, `view_test.go`,
  `docs/usage.md`, `CHANGELOG.md`, `docs/development-journal.md`.
- `2026-11-13T18:49:12Z` — Rename the application, source directory, commands,
  shell function/helper variables, executables, archives, notices, installation
  paths, release title, and documentation to cdd. Set both Go module paths to
  github.com/nik-makarov/cdd and its tools submodule; update the local SSH origin
  after the owner reported renaming the GitHub repository. Rephrase historical
  identity mentions neutrally without inventing past output under the new name.
  Remove obsolete generated binaries/archives and regenerate renamed previews.
  Module tidying, formatting, vet, race-enabled application/tool tests, host and
  four target builds, archive packaging/checksums, module verification, temporary
  installation, and both shell syntax checks passed. A native zsh PTY check
  changed directory through cdd with the renamed notification and zero status.
  Maintained text contains no previous command/module spelling. Dependency
  versions are unchanged; no Git push, hosted CI, or release publication ran.
  Paths: `cli.go`, `cli_test.go`, `view.go`, `view_test.go`, `shell.go`,
  `shell_test.go`, `Makefile`, `.gitignore`, `.github/workflows/release.yml`,
  `go.mod`, `tools/go.mod`, `tools/notices/main.go`, `tools/release/`,
  `README.md`, `CONTRIBUTING.md`, `SECURITY.md`, `THIRD_PARTY_NOTICES.md`,
  `CHANGELOG.md`, `packaging/README.md`, `docs/`.
- `2026-11-13T19:10:47Z` — Route shell-function output by the absolute-path
  contract instead of maintaining a second argument parser. Relay successful
  informational output verbatim and preserve selection, cancellation, errors,
  and newline-containing paths. Real zsh/bash regression tests exercise CLI
  aliases, explicit/repeated boolean values, initialization, help, and invalid
  arguments. Formatting, vet, shuffled race-enabled application/tool tests, and
  the host build passed. Host-binary informational checks and a native zsh PTY
  selection with --version=false passed. No dependencies changed; no hosted CI,
  Git push, or publication ran. Paths: `shell.go`, `shell_test.go`, `CHANGELOG.md`,
  `docs/usage.md`, `docs/testing.md`, `docs/development-journal.md`.
- `2026-11-13T19:35:01Z` — Publish automatically after checked main pushes,
  starting from VERSION=0.1.0 and incrementing stable patch tags. Queue releases,
  refuse out-of-order history, resume drafts, and reuse published assets on
  retry. Generate deterministic committed-source archives and source-built
  Homebrew formulae with matching SHA-256; update only Formula/cdd.rb in the
  separate tap using its scoped token. Use curated first-release notes and
  optional dated sections, otherwise incremental English commit subjects.
  Formatting, vet, shuffled race-enabled application/tool tests, and host build
  passed. An isolated committed snapshot passed automatic four-target/source
  packaging, archive checksums, extracted-source build/version checks, and both
  shell syntax checks. Workflow YAML/shell syntax and formula Ruby syntax
  passed. A disposable Homebrew copy loaded formula metadata before GitHub DNS
  blocked dependency lookup; complete brew installation/style/tests/audit were
  not run. Dependency versions are unchanged. No hosted CI ran; no secrets were
  set or remote branches/tags/releases published. Paths: `.github/workflows/ci.yml`,
  `.github/workflows/release.yml`, `VERSION`, `Makefile`, `tools/release/`,
  `tools/README.md`, `AGENTS.md`, `README.md`, `CHANGELOG.md`, `CONTRIBUTING.md`,
  `docs/releases.md`, `docs/maintenance.md`, `docs/testing.md`,
  `docs/development-journal.md`.
- `2026-11-13T20:04:58Z` — Infer the Homebrew version from the source archive
  URL instead of emitting a redundant version stanza. The owner-provided hosted
  log showed successful style, Go 1.27.4 source installation, and formula tests
  before strict audit rejected the duplicate version. Update the generator and
  user-facing notes. Formatting, vet, shuffled race-enabled application/tool
  tests, and host build passed locally; the corrected current tap formula also
  passed Ruby syntax checking. No dependency versions or published release
  assets changed, and no hosted rerun occurred during local verification. Paths:
  `tools/release/automatic.go`, `CHANGELOG.md`, `docs/testing.md`,
  `docs/development-journal.md`.
- `2026-11-13T20:27:59Z` — Skip the main-branch Release workflow when every
  changed path is Markdown or under docs/. Preserve releases for mixed changes,
  pull-request CI, and scheduled audits; deferred documentation is included in
  the next eligible release. Update release, contributor, and project guidance.
  Workflow YAML parsing, decoded branch/path filters, shell syntax, and diff
  checks passed locally. No Go code or dependencies changed. GitHub DNS blocked
  remote synchronization; no branch was pushed or hosted workflow run. Paths:
  `.github/workflows/release.yml`, `AGENTS.md`, `CHANGELOG.md`, `CONTRIBUTING.md`,
  `README.md`, `docs/releases.md`, `docs/development-journal.md`.
- `2026-11-13T20:40:39Z` — Make Homebrew the first and recommended README
  installation route, followed by persistent zsh setup using the requested
  literal echo command and source ~/.zshrc. Identify the Apple Silicon prefix,
  retain guidance for other prefixes and bash, and move archives/source builds
  below usage as alternatives. Move Go/Make requirements into the source-build
  section. All README shell snippets passed bash/zsh syntax checks; local link
  targets, heading anchors, and the diff were reviewed. No shell configuration
  was modified or installation performed. Only Markdown files changed, so this
  change does not trigger Release. No remote push or hosted check ran. Paths:
  `README.md`, `docs/development-journal.md`.
- `2026-11-13T20:52:31Z` — Display each column's directory path above its list,
  using existing header rows. Show parent/current/highlighted-directory paths
  in tree view and the same current path over both compact columns. Clear the
  right address for files and the current row, and the left address at root.
  Clip sanitized paths within each pane and reuse the anchored pane renderer
  for compact rows. Formatting, vet, shuffled race-enabled application/tool
  tests, and host build passed. Renderer cases verify path mapping, logical
  links, empty addresses, compact/narrow/short layouts, and Unicode/control-text
  clipping. A native 120-column PTY check verified preview paths, navigation,
  clearing, both compact headings, and a successful absolute-path selection.
  No dependencies changed; hosted CI and publication were not run. Paths:
  `view.go`, `view_test.go`, `CHANGELOG.md`, `docs/usage.md`, `docs/testing.md`,
  `docs/development-journal.md`.
- `2026-11-13T21:12:05Z` — Emphasize the active directory address over pane
  labels. Render Current without brackets using the same dim style as Parent
  and Children; use the normal foreground for the active path prefix and retain
  its bold cyan basename. Dim side entries without losing type colors/markers,
  and remove bold current-folder emphasis from the parent pane. Formatting, vet,
  shuffled race-enabled application/tool tests, and host build passed. Updated
  renderer assertions cover hierarchy, side colors, NO_COLOR, and selection.
  A color-enabled native 120-column PTY check confirmed the emitted styling and
  successful directory selection. No dependencies changed; GitHub DNS prevented
  remote synchronization, and no push, hosted CI, or publication ran. Paths:
  `tree.go`, `tree_test.go`, `view.go`, `view_test.go`, `CHANGELOG.md`,
  `docs/usage.md`, `docs/testing.md`, `docs/development-journal.md`.
- `2026-11-13T21:18:36Z` — Focus the current-directory row before leaving the
  directory with Left, h, or Backspace; only navigate to the parent when . is
  already selected. Preserve completed filters on the first press, clear stale
  errors/preview state, and retain prior parent-focus behavior on the second.
  Update CLI help, dynamic footer hints, README controls, usage, and release
  notes. A regression test reproduced the old behavior before the fix.
  Formatting, vet, shuffled race-enabled application/tool tests, and host build
  passed. Cases include both views, files, links, compact second-column focus,
  filters, current selection, and logical symlink navigation. A native
  100-column PTY check verified both arrow presses, footer actions, parent
  focus, and successful directory selection. No dependencies changed; GitHub
  DNS blocked synchronization, and no push, hosted CI, or publication ran.
  Paths: `browser.go`, `browser_test.go`, `cli.go`, `view.go`, `README.md`,
  `CHANGELOG.md`, `docs/usage.md`, `docs/testing.md`, `docs/development-journal.md`.
- `2026-11-15T13:21:36Z` — Require loaded shell integration before browsing.
  Set an invocation-scoped marker in the generated zsh/bash function and report
  missing integration on stderr with initialization commands and instructions
  to reopen the terminal. Keep help, version, and initialization available before
  setup. Update CLI help, usage, and release notes. Formatting, vet, shuffled
  race-enabled application/tool tests, and host build passed with a temporary Go
  build cache after the default cache denied writes. The successful build also
  reported a nonfatal module-cache write warning. Tests cover missing/invalid
  markers, marker scope, informational commands, cancellation, and literal paths
  with quotes, metacharacters, Unicode, and trailing newlines. A host-binary
  check returned the setup message with status 1; a native zsh PTY check loaded
  integration, selected a directory, and changed the shell directory with status
  0. No dependencies changed or hosted CI/publication ran. Paths: `cli.go`,
  `cli_test.go`, `shell.go`, `shell_test.go`, `CHANGELOG.md`, `docs/usage.md`,
  `docs/development-journal.md`.
- `2026-11-15T13:26:22Z` — Align the missing-integration message with the README's
  Homebrew setup instructions. Show the same literal echo command that appends
  initialization to .zshrc and the source command that loads it, retain the
  terminal-reopen reminder, and include the README's prefix and bash guidance.
  Update CLI assertions and usage documentation. Formatting, vet, shuffled
  race-enabled application/tool tests, and host build passed with a temporary
  Go build cache; the build retained the nonfatal module-cache write warning.
  A host-binary check displayed the updated instructions and exited with status
  1. No shell configuration was modified. Paths: `shell.go`, `cli.go`,
  `cli_test.go`, `docs/usage.md`, `docs/development-journal.md`.
- `2026-11-15T13:29:49Z` — Highlight the missing-integration heading in bold red
  and the README setup commands in bold cyan using the existing terminal palette.
  Detect terminal support on the actual diagnostic writer; retain bold emphasis
  under NO_COLOR and plain text for TERM=dumb or redirected stderr. Preserve the
  instruction text and update usage/release notes. Formatting, vet, shuffled
  race-enabled application/tool tests, and host build passed with a temporary
  Go build cache; the build retained the nonfatal module-cache write warning.
  Native PTY checks verified color even with stdout redirected, bold-only output
  with NO_COLOR, and plain output with TERM=dumb. A captured-stderr host-binary
  check and CLI assertions confirmed no ANSI styling outside a terminal. No
  dependencies changed. Paths: `cli.go`, `cli_test.go`, `shell.go`,
  `CHANGELOG.md`, `docs/usage.md`, `docs/development-journal.md`.
- `2026-11-15T13:39:35Z` — Add directory creation with n, editable name input,
  Enter to create, and Esc to dismiss. Create one literal child of the current
  directory with os.Mkdir, preserve existing entries, and keep failed names
  editable. Refresh and focus the new directory, clear completed filters, and
  reveal newly created hidden directories. Add footer/help controls, clipped
  input rendering, usage instructions, release notes, and the requested narrow
  scope exception for creating navigation destinations. Formatting, vet,
  shuffled race-enabled application/tool tests, and host build passed with a
  temporary Go build cache; the build retained the nonfatal module-cache write
  warning. Tests cover both views, Unicode editing, literal metacharacters,
  cancellation, invalid names, existing files/directories/symlinks, filesystem
  failure and retry, logical symlink paths, and terminal-size constraints.
  A native zsh PTY check verified an existing-name error, creation and focus of
  a space-containing name, and a successful shell directory change. A bash PTY
  check verified cancellation in compact mode without creation or a directory
  change. No dependencies changed or hosted CI/publication ran for this change.
  Paths: `AGENTS.md`, `browser.go`, `browser_test.go`, `view.go`, `view_test.go`,
  `cli.go`, `README.md`, `CHANGELOG.md`, `docs/usage.md`, `docs/architecture.md`,
  `docs/testing.md`, `docs/development-journal.md`.
- `2026-11-15T14:14:01Z` — Preserve compatibility with shell functions loaded
  before the integration marker was introduced. Accept markerless stdout pipes
  used by those functions and explicit path-capture workflows while retaining
  setup guidance for direct terminal output and uncaptured file output. Update
  usage and release notes. A regression test first reproduced false rejection
  with both zsh and bash, then passed with the compatibility check. Formatting,
  vet, shuffled race-enabled application/tool tests, and host build passed with
  a temporary Go build cache; the build retained the nonfatal module-cache write
  warning. A native zsh PTY check selected a directory and changed the shell
  location through a markerless function; a direct terminal invocation still
  displayed the colored setup message and returned status 1. No dependencies
  or user shell configuration changed. Paths: `cli.go`, `shell.go`,
  `shell_test.go`, `CHANGELOG.md`, `docs/usage.md`, `docs/development-journal.md`.
- `2026-11-15T18:43:16Z` — Add mouse navigation in the active tree pane and both
  compact columns: click to highlight, double-click to select, right-click to
  browse, and wheel to move selection. Decode fragmented SGR and legacy mouse
  packets, ignore non-entry regions, preserve input-mode behavior, and reuse
  directory-selection guards. Keep scrolled rows stable under the pointer and
  disable mouse reporting during terminal cleanup. Add controls help and usage,
  architecture, testing, and release documentation. Formatting, vet, shuffled
  race-enabled application/tool tests, and host build passed with a temporary
  Go build cache; the build retained the nonfatal module-cache write warning.
  gopls reported no diagnostics. Decoder fuzzing passed 970,426 executions.
  Tests cover rendered-row hit testing across sizes and modes, both compact
  columns, empty regions, current-directory selection, stale directories, files,
  click-state resets, filtering, and directory-name input. Native PTY checks
  injected mouse reports into the host binary and verified highlighting,
  right-click browsing, double-click selection with a zsh directory change,
  second-column selection with a bash directory change, wheel movement, and
  cancellation without a directory change. Selection and cancellation emitted
  mouse-disable sequences. Physical pointing-device behavior was not tested.
  No dependencies changed or hosted CI/publication ran for this change. Paths:
  `browser.go`, `terminal.go`, `terminal_test.go`, `mouse.go`, `mouse_test.go`,
  `tree.go`, `view.go`, `cli.go`, `README.md`, `CHANGELOG.md`, `docs/usage.md`,
  `docs/architecture.md`, `docs/testing.md`, `docs/development-journal.md`.
- `2026-11-15T19:17:26Z` — Add a persistent parent-directory row after the current
  row, except at the filesystem root. Keep both rows outside filtering and entry
  counts, leave their Children preview blank, and restore focus to the previous
  directory when opening the logical parent. Change double-clicks to browse
  directories without completing selection and remove right-click navigation.
  Add a prominent footer button labeled Select directory with its Enter shortcut;
  clicking it reuses Enter's directory validation and completes selection.
  Shorten the label for narrow windows, dim it for files, and hide it during
  filter/name editing. Update controls, usage, architecture, and release notes.
  Formatting, vet, shuffled race-enabled application/tool tests, and host build
  passed with a temporary Go build cache; the build retained the nonfatal
  module-cache write warning. gopls reported no diagnostics. Tests cover root
  behavior, filtering, hidden entries, logical symlink parents, focus restoration,
  counts/previews, double-click browsing, Enter selection, and button geometry
  across terminal sizes, modes, help, and clipping. Button tests also cover
  NO_COLOR emphasis, inactive file styling, and file/removed-directory guards.
  Native PTY checks injected mouse reports to open a child, return through ..,
  and confirm with the button through zsh; a bash check opened a directory from
  the second compact column and confirmed with Enter. Both changed the shell
  directory as expected and disabled mouse reporting on exit. Terminal output
  showed the button with color and with bold-only emphasis. No dependencies
  changed or hosted CI/publication ran. Paths: `browser.go`, `browser_test.go`,
  `entry.go`, `mouse.go`, `mouse_test.go`, `view.go`, `view_test.go`, `tree.go`,
  `tree_test.go`, `cli.go`, `README.md`, `CHANGELOG.md`, `docs/usage.md`,
  `docs/architecture.md`, `docs/testing.md`, `docs/development-journal.md`.
- `2026-11-15T19:33:01Z` — Place the parent row above the current-directory row
  and prefix it with a left arrow in both views. Preserve initial focus on the
  current directory, including empty filters and the filesystem root. Make Left,
  h, Backspace, and their contextual hint identify the current row by name rather
  than its previous position. Update ordering/navigation assertions and usage
  text. Formatting, vet, shuffled race-enabled application/tool tests, and host
  build passed with a temporary Go build cache; the build retained the nonfatal
  module-cache write warning. gopls reported no diagnostics. A native zsh PTY
  check displayed the parent arrow above the highlighted current row in tree
  and compact views, injected a double-click on the first row to open the parent
  and restore child focus, then confirmed with Enter and verified the resulting
  shell directory. No dependencies changed or hosted CI/publication ran. Paths:
  `browser.go`, `browser_test.go`, `view.go`, `view_test.go`, `mouse_test.go`,
  `cli.go`, `CHANGELOG.md`, `docs/usage.md`, `docs/architecture.md`,
  `docs/development-journal.md`.
- `2026-11-15T19:46:10Z` — Replace the single-line selection button with a centered
  three-line Change Directory block at the bottom, showing the platform shortcut
  and highlighted destination. Sanitize and clip long paths from the left, keep
  every button row clickable, retain NO_COLOR emphasis, and omit the complete
  block when input mode or available space prevents displaying it. Make plain
  Enter browse like Right; confirm only through the button or Cmd+Enter on macOS
  and Ctrl+Enter on Linux. Push/pop the kitty keyboard disambiguation mode and
  decode extended Enter, cancellation, refresh, Tab, and navigation keys; ignore
  confirmation repeats and key releases. Document terminal compatibility and
  update help, usage, architecture, testing, and release notes. Formatting, vet,
  shuffled race-enabled application/tool tests, and host build passed with a
  temporary Go build cache; the build retained the nonfatal module-cache write
  warning. gopls reported no diagnostics. Decoder fuzzing passed 4,130,011
  executions. Tests cover both platform modifiers, browsing/confirmation
  separation, input modes, centered button geometry, all three clickable rows,
  destination updates/sanitization/clipping, and file/stale-directory guards.
  An isolated background agterm session showed the new layout, updated the button
  path after Enter browsing, and exited after an injected Cmd+Enter report; the
  temporary session closed and the user's keymap was only read. Native PTY checks
  verified a zsh directory change after Enter browsing and an injected Cmd+Enter,
  a bash directory change from clicking the destination row, and extended Ctrl+C
  cancellation without changing directory. Cleanup emitted keyboard-mode pop and
  mouse-disable sequences. Physical modifier-key presses and a Linux runtime were
  not tested. No dependencies changed or hosted CI/publication ran. Paths:
  `browser.go`, `browser_test.go`, `terminal.go`, `terminal_test.go`, `keyboard.go`,
  `keyboard_test.go`, `mouse.go`, `mouse_test.go`, `view.go`, `view_test.go`, `cli.go`,
  `README.md`, `CHANGELOG.md`, `docs/usage.md`, `docs/architecture.md`,
  `docs/testing.md`, `docs/development-journal.md`.
- `2026-11-15T19:51:08Z` — Confirm the current directory and exit when plain Enter
  is pressed on the `.` row, reusing destination validation. Keep Enter browsing
  other directories and finishing filter or directory-name input. Make the
  compact controls hint reflect the highlighted row and document the exception
  in CLI help, usage, architecture, testing, and release notes. Formatting, vet,
  shuffled race-enabled application/tool tests, and host build passed with a
  temporary Go build cache; the build retained the nonfatal module-cache write
  warning. gopls reported no diagnostics. Regression cases cover current-row
  confirmation in both views, after Left, and after finishing an unmatched
  filter, while retaining child/parent browsing and modified confirmation. A
  native zsh PTY check confirmed that Enter on `.` exits and changes the shell's
  working directory to the displayed location. No dependencies changed or hosted
  CI/publication ran. Paths: `browser.go`, `browser_test.go`, `keyboard_test.go`,
  `view.go`, `view_test.go`, `cli.go`, `README.md`, `CHANGELOG.md`, `docs/usage.md`,
  `docs/architecture.md`, `docs/testing.md`, `docs/development-journal.md`.
- `2026-11-15T19:55:33Z` — Move the three-row Change Directory button above the
  keyboard shortcuts in the footer, including expanded help. Update the existing
  rendered-position assertion, usage, architecture, and release notes. Formatting,
  vet, shuffled race-enabled application/tool tests, and host build passed with
  a temporary Go build cache; the build retained the nonfatal module-cache write
  warning. Existing mouse tests verified all three clickable rows against their
  rendered positions across views, terminal sizes, and input/help modes. gopls
  reported no diagnostics. No dependencies changed or hosted CI/publication ran.
  Paths: `view.go`, `view_test.go`, `CHANGELOG.md`, `docs/usage.md`,
  `docs/architecture.md`, `docs/development-journal.md`.
- `2026-11-15T20:20:53Z` — Save the previous picker state in commit `8f0c772`
  before revising navigation and confirmation styling. Make Left, h, and
  Backspace focus the parent row before opening it, preserving completed filters
  on the first press and leaving the filesystem root unchanged. Emphasize the
  parent as a Back to parent action. Replace the filled confirmation block with
  a narrower rounded outline, a shared action/shortcut line, and a destination
  line with an accented basename. Keep all four rows clickable above the controls.
  Budget UI arrows and box drawing as single cells and anchor the right border
  independently of filename glyph widths. Allow two cells of horizontal drift
  between double-clicks on the same entry and row. Document the interactions and
  design references from NN/g, Textual, and W3C. Formatting, vet, shuffled
  race-enabled application/tool tests, and host build passed with a temporary Go
  build cache; the build retained the nonfatal module-cache write warning. gopls
  reported no diagnostics. Tests cover parent focus from current/child/file/link
  rows, filter preservation, root handling, logical symlinks, pointer drift and
  its limit, outline geometry, Unicode clipping, and rendered click boundaries
  across views, sizes, and help/input modes. An isolated background agterm session
  displayed the new block, verified both Left steps, and exited after an injected
  click on Change Directory; the temporary session closed. Physical trackpad
  gestures and a Linux runtime were not tested. No dependencies changed or hosted
  CI/publication ran. Paths: `browser.go`, `browser_test.go`, `view.go`,
  `view_test.go`, `mouse.go`, `mouse_test.go`, `terminal_test.go`, `cli.go`,
  `README.md`, `CHANGELOG.md`, `docs/usage.md`, `docs/architecture.md`,
  `docs/testing.md`, `docs/development-journal.md`.
- `2026-11-15T20:25:04Z` — Remove the literal dots from the displayed parent-row
  label, leaving the arrow and Back to parent action. Retain the internal `..`
  entry and its navigation behavior. Update the rendered-label and mouse-target
  assertions, CLI help, and documentation. Focused view and parent mouse-navigation
  tests passed; the host binary rebuilt with the existing nonfatal module-cache
  write warning. gopls reported no diagnostics. No dependencies changed or hosted
  CI/publication ran. Paths: `view.go`, `view_test.go`, `mouse_test.go`, `cli.go`,
  `README.md`, `CHANGELOG.md`, `docs/usage.md`, `docs/architecture.md`,
  `docs/development-journal.md`.
- `2026-11-15T20:27:33Z` — Show Enter on the Change Directory button when the
  current-directory row is highlighted, and retain the platform modifier shortcut
  for other entries. Keep the button dimensions and visibility threshold stable
  when focus changes. Update existing button assertions and usage, architecture,
  and release notes. Focused button appearance, destination, hit-testing, and
  directory-guard tests passed; the host binary rebuilt with the existing nonfatal
  module-cache write warning. gopls reported no diagnostics. No dependencies changed
  or hosted CI/publication ran. Paths: `view.go`, `view_test.go`, `CHANGELOG.md`,
  `docs/usage.md`, `docs/architecture.md`, `docs/development-journal.md`.
- `2026-11-15T20:30:50Z` — Move the parent action's arrow into the left gutter
  so Back to parent aligns with directory names. Keep the arrow in place when
  highlighted; retain the full-row highlight and use a trailing selection marker
  when styling is unavailable. Update usage documentation. Existing view, tree,
  entry, and mouse tests passed, and the host binary rebuilt with the existing
  nonfatal module-cache write warning. gopls reported no diagnostics. A native
  PTY check showed the arrow before and after Left focus, then cancelled with
  terminal cleanup and exit status 130. No dependencies changed or hosted
  CI/publication ran. Paths: `view.go`, `docs/usage.md`,
  `docs/development-journal.md`.
- `2026-11-15T20:34:04Z` — Hide the confirmation shortcut on the inactive button
  when a file is highlighted, and omit opening/Enter actions from the short
  controls. Keep directory shortcuts, button geometry, and text-input controls
  intact. Update the existing file-button assertion, usage, architecture, and
  release notes. Focused button, view, creation-prompt, and input-mode tests passed;
  the host binary rebuilt with the existing nonfatal module-cache write warning.
  gopls reported no diagnostics. No dependencies changed or hosted CI/publication
  ran. Paths: `view.go`, `view_test.go`, `CHANGELOG.md`, `docs/usage.md`,
  `docs/architecture.md`, `docs/development-journal.md`.
- `2026-11-15T20:42:35Z` — Save the preceding UI revision in commit `d232119`
  after `make check build` passed, then shorten application-owned interface text
  from directory/directories to dir/dirs. Label the current row `(current dir)`
  and the confirmation button Change Dir. Update controls, counts, CLI help,
  diagnostic operation text, rendered-label assertions, error expectations, and
  documentation references. Preserve paths, internal identifiers, and underlying
  operating-system errors. Application/tool tests and the host build passed after
  the wording changes; the build retained the nonfatal module-cache write warning.
  gopls reported no diagnostics, and the rebuilt help output was inspected.
  No dependencies changed or hosted CI/publication ran. Paths: `view.go`,
  `browser.go`, `cli.go`, `entry.go`, `view_test.go`, `mouse_test.go`,
  `keyboard_test.go`, `errors_test.go`, `README.md`, `CHANGELOG.md`,
  `docs/usage.md`, `docs/testing.md`, `docs/development-journal.md`.
- `2026-11-15T20:48:35Z` — Remove trailing slashes from folder labels in tree
  and compact views, side panes, and the entry-type help legend. Retain `@` for
  directory symlinks and preserve filesystem paths. Update existing entry, view,
  tree, and mouse assertions; restrict mouse row matching to the list body so
  destination paths are not mistaken for entries. Update usage, architecture,
  and release notes. Application/tool tests and the host build passed; the build
  retained the nonfatal module-cache write warning. gopls reported no diagnostics,
  and `git diff --check` passed. No dependencies changed or hosted CI/publication
  ran. Paths: `entry.go`, `view.go`, `entry_test.go`, `view_test.go`, `tree_test.go`,
  `mouse_test.go`, `CHANGELOG.md`, `docs/usage.md`, `docs/architecture.md`,
  `docs/development-journal.md`.
- `2026-11-15T20:52:45Z` — Confirm the current-directory row with a double-click
  through the existing Change Dir handler, including while filtering. Keep a
  single click as highlighting and other directory double-clicks as browsing.
  Update mouse tests for both views, active filtering, and rejection of a removed
  current directory. Update CLI help, README, usage, architecture, testing guidance,
  and release notes. Application/tool tests and the host build passed; the build
  retained the nonfatal module-cache write warning. gopls reported no diagnostics,
  and `git diff --check` passed. No dependencies changed or hosted CI/publication
  ran. Paths: `mouse.go`, `mouse_test.go`, `cli.go`, `README.md`, `CHANGELOG.md`,
  `docs/usage.md`, `docs/architecture.md`, `docs/testing.md`,
  `docs/development-journal.md`.
- `2026-11-15T21:06:13Z` — Add an expandable tree as the second view in the
  panes/tree/list cycle. Apply the WAI-ARIA tree keyboard model and filesystem
  guide/disclosure conventions documented by Textual, retaining cdd's explicit
  destination confirmation. Load immediate children on expansion, keep branch
  state, filter expanded rows with ancestor context, preserve logical symlink
  paths, and refresh expanded branches with focus fallback after deletion. Add
  disclosure hit targets, nested directory creation, contextual controls, empty
  states, and deep-indentation clipping. Preserve nested destination focus when
  switching to a flat view. Update existing mode fixtures, add tree interaction
  and rendering coverage, and document controls, filter scope, and design sources.
  `make check build` passed, including application/tool vet and shuffled race
  tests; the host build retained the nonfatal module-cache write warning. gopls
  reported no diagnostics, and `git diff --check` passed. Native PTY checks at
  100x24 and 80x20 verified nested expansion, mouse disclosure, view cycling,
  retained destination paths, visible arrow controls, modified-Enter selection,
  and double-click confirmation of the current row with terminal cleanup and
  exit status 0. No dependencies changed or hosted CI/publication ran.
  Paths: `outline.go`, `outline_test.go`, `browser.go`, `browser_test.go`,
  `view.go`, `view_test.go`, `mouse.go`, `mouse_test.go`, `keyboard_test.go`,
  `tree_test.go`, `cli.go`, `README.md`, `CHANGELOG.md`, `docs/usage.md`,
  `docs/architecture.md`, `docs/testing.md`, `docs/development-journal.md`.
- `2026-11-15T21:22:48Z` — Make Enter confirm any highlighted tree directory,
  including filtered matches, while preserving directory-name input and file
  guards. Anchor the tree at home when it contains the browsing location,
  otherwise at `/`; reveal ancestor chains and retain hidden ancestors without
  exposing unrelated hidden entries. Replace synthetic tree navigation rows with
  real parent nodes. Track the shell's immutable launch directory independently
  from CLI starting paths and browsing, and mark it with yellow/bold styling,
  `(starting dir)`, and a From header. Keep text/bold emphasis with NO_COLOR.
  Update tree controls, button shortcuts, root/source/confirmation coverage, and
  documentation. `make check build` passed, including vet and shuffled race tests
  for application and tools; the build retained the nonfatal module-cache write
  warning. gopls reported no diagnostics, and `git diff --check` passed. A native
  100x24 PTY check verified the home-rooted ancestor chain, the source marker both
  focused and unfocused, an Enter button shortcut on a child, and plain-Enter
  confirmation returning that child's path with terminal cleanup and exit 0.
  No dependencies changed or hosted CI/publication ran. Paths: `browser.go`,
  `outline.go`, `outline_test.go`, `view.go`, `cli.go`, `README.md`, `CHANGELOG.md`,
  `docs/usage.md`, `docs/architecture.md`, `docs/testing.md`,
  `docs/development-journal.md`.
- `2026-11-15T21:27:44Z` — Replace the starting directory's yellow/bold accent
  with the terminal theme's dim foreground, including the From header. Retain
  dimming under the selection bar and preserve the textual starting-dir marker
  with NO_COLOR. Update the existing styling assertion and documentation. Restore
  chronological placement of the preceding journal entry. Application/tool tests
  and the host build passed; the build retained the nonfatal module-cache write
  warning. gopls reported no diagnostics, and `git diff --check` passed. No
  dependencies changed or hosted CI/publication ran. Paths: `view.go`, `outline.go`,
  `outline_test.go`, `README.md`, `CHANGELOG.md`, `docs/usage.md`,
  `docs/architecture.md`, `docs/testing.md`, `docs/development-journal.md`.
- `2026-11-15T21:36:48Z` — Remember the last successfully selected view using a
  stable mode name in the user's configuration directory. Save mode changes
  immediately, including before cancellation, with atomic file replacement.
  Restore the preference on interactive startup and fall back to panes for
  missing, unreadable, unknown, or unavailable views. Keep navigation usable when
  saving fails and leave preferences untouched after failed mode transitions.
  Add isolated temporary-file coverage for persistence, restart, cancellation,
  fallback, and write failures; update help and documentation. `make check build`
  passed, including vet and shuffled race tests for application and tools; the
  build retained the nonfatal module-cache write warning. gopls reported no
  diagnostics, and `git diff --check` passed. No dependencies changed or hosted
  CI/publication ran. Paths: `preferences.go`, `preferences_test.go`, `browser.go`,
  `outline.go`, `cli.go`, `README.md`, `CHANGELOG.md`, `docs/usage.md`,
  `docs/architecture.md`, `docs/testing.md`, `docs/development-journal.md`.
- `2026-11-15T22:09:12Z` — Add a visible filesystem-root button and backslash
  shortcut in tree view. Preserve nested selection and expanded branches while
  revealing ancestors above home, and retain root scope during session navigation.
  Cover rendering, mouse/keyboard activation, focus, and hidden-button hit testing.
  `make check build`, gopls diagnostics, and `git diff --check` passed; the host
  build retained the nonfatal module-cache write warning. No dependencies changed
  or publication ran. Paths: `browser.go`, `outline.go`, `outline_test.go`,
  `view.go`, `mouse.go`, `cli.go`, `CHANGELOG.md`, `docs/usage.md`,
  `docs/development-journal.md`.
- `2026-11-15T22:11:05Z` — Make visible header path components clickable for
  browsing, with tree paths following the highlighted directory or file parent.
  Share pane locations between rendering and hit testing; ignore ellipses,
  separators, margins, and hidden headers. Cover clipped, Unicode, control-bearing,
  and literal paths plus navigation without shell confirmation. `make check build`
  and `git diff --check` passed; the build retained the nonfatal module-cache write
  warning. No dependencies changed or publication ran. Paths: `breadcrumb.go`,
  `breadcrumb_test.go`, `view.go`, `mouse.go`, `CHANGELOG.md`, `docs/usage.md`,
  `docs/development-journal.md`.
- `2026-11-15T22:13:57Z` — Default interactive startup to the expandable tree
  when no valid preference exists and order the view cycle as tree, panes, list.
  Preserve saved pane/list choices and retain panes if ancestor loading fails.
  Update cycle, focus, restoration, and fallback checks plus user documentation.
  `make check build`, gopls diagnostics, and `git diff --check` passed; the build
  retained the nonfatal module-cache write warning. No dependencies changed or
  publication ran. Paths: `outline.go`, `preferences.go`, `preferences_test.go`,
  `outline_test.go`, `browser_test.go`, `view.go`, `cli.go`, `README.md`,
  `CHANGELOG.md`, `docs/usage.md`, `docs/architecture.md`,
  `docs/development-journal.md`.
- `2026-11-15T22:18:16Z` — Remove compact directory-only mode, its two-column
  layout, column-switching controls, and obsolete mode-specific tests. Retain
  pane navigation coverage and tree/pane focus preservation. Treat saved `list`
  values as unknown and restore tree. Update current documentation while keeping
  historical verification records. `make check build`, gopls diagnostics, and
  `git diff --check` passed; the build retained the nonfatal module-cache write
  warning. No dependencies changed or publication ran. Paths: `outline.go`,
  `preferences.go`, `browser.go`, `view.go`, `mouse.go`, `breadcrumb.go`, `cli.go`,
  `browser_test.go`, `outline_test.go`, `preferences_test.go`, `view_test.go`,
  `mouse_test.go`, `keyboard_test.go`, `tree_test.go`, `README.md`, `CHANGELOG.md`,
  `docs/usage.md`, `docs/architecture.md`, `docs/testing.md`,
  `docs/development-journal.md`.
- `2026-11-15T22:22:45Z` — Add explicit Space preview: Quick Look through
  `qlmanage -p` on macOS and a bounded, scrollable text viewer on Linux. Preserve
  browsing focus, sanitize rendered contents, reject special files, and handle
  binary data and UTF-8 at the read limit. Add a cancellable terminal-owned child
  runner with isolated stdout and bounded diagnostics. Update the project scope
  for explicitly requested preview. Tests use a fake CLI rather than launching
  desktop apps. `make check build`, gopls diagnostics, and `git diff --check`
  passed; the build retained the nonfatal module-cache write warning. No
  dependencies changed or publication ran. Paths: `actions.go`, `actions_test.go`,
  `preview.go`, `preview_test.go`, `browser.go`, `view.go`, `terminal.go`,
  `shell_test.go`, `cli.go`, `AGENTS.md`, `README.md`, `CHANGELOG.md`,
  `docs/usage.md`, `docs/architecture.md`, `docs/testing.md`,
  `docs/development-journal.md`.
- `2026-11-15T22:25:46Z` — Open regular files with Enter through macOS `open`
  or Linux `xdg-open`, retaining browser focus and directory-only confirmation.
  Support Enter from the built-in preview and filtered tree results; preserve
  pane filter input behavior. Cancel and join an active launcher before replacing
  it. Update file hints and project scope. Fake opener tests verify both platform
  commands, literal paths, stale-file rejection, and separate cd guards.
  `make check build`, gopls diagnostics, and `git diff --check` passed; the build
  retained the nonfatal module-cache write warning. No dependencies changed or
  publication ran. Paths: `actions.go`, `open_test.go`, `browser.go`,
  `browser_test.go`, `outline_test.go`, `view.go`, `view_test.go`, `preview.go`,
  `terminal.go`, `cli.go`, `AGENTS.md`, `README.md`, `CHANGELOG.md`,
  `docs/usage.md`, `docs/architecture.md`, `docs/testing.md`,
  `docs/development-journal.md`.
- `2026-11-15T22:31:36Z` — Add Zed launching through its installed CLI, passing
  a folder as a project or only the selected file with `--new --`. Ignore missing
  CLI installations and support Linux's `zeditor` alias. Embed default key settings
  from JSON and merge validated user overrides without rewriting them; use the
  configured key in input handling and help, including built-in preview. Update
  project scope and configuration documentation. Tests cover custom/disabled keys,
  malformed/conflicting settings, text-input precedence, missing CLI, new-window
  arguments, and retained browsing state. `make check build cross-build` passed
  for the host and Linux/macOS amd64/arm64 targets; builds retained the nonfatal
  module-cache write warning. gopls and `git diff --check` passed. Native PTY
  acceptance with temporary configuration and fake desktop CLIs verified default
  tree startup, root-button and breadcrumb clicks, Space/Enter/configured editor
  dispatch, folder versus file arguments, the two-view cycle, saved mode, empty
  cancellation stdout, directory-only selection output, and terminal cleanup.
  Desktop applications were not opened by these acceptance checks. No dependencies
  changed or publication ran. Paths: `config.go`, `config/default.json`,
  `config_test.go`, `zed_test.go`, `actions.go`, `browser.go`, `preferences.go`,
  `cli.go`, `view.go`, `AGENTS.md`, `README.md`, `CHANGELOG.md`,
  `docs/configuration.md`, `docs/usage.md`, `docs/architecture.md`,
  `docs/testing.md`, `docs/development-journal.md`.
- `2026-11-15T22:38:15Z` — Add Shift+N to create an empty named file inside the
  highlighted directory or beside the highlighted file in either view. Use
  exclusive creation, preserve literal names, reject existing entries including
  symlinks, keep failed input editable, and focus the new file without opening it
  or selecting it for cd. Show the creation destination, support shifted CSI-u
  letters, reserve the shortcut from editor bindings, and document the requested
  scope expansion. Tests cover both views, hidden files, directory symlinks,
  Unicode and shell metacharacters, cancellation, missing-parent retry, input
  precedence, and non-overwrite behavior. `make check build`, gopls diagnostics,
  a ten-second keyboard fuzz run, and `git diff --check` passed; the build retained
  the nonfatal module-cache write warning. No dependencies changed or publication
  ran. Paths: `browser.go`, `outline.go`, `view.go`, `keyboard.go`, `config.go`,
  `create_file_test.go`, `config_test.go`, `terminal_test.go`, `cli.go`, `AGENTS.md`,
  `README.md`, `CHANGELOG.md`, `docs/configuration.md`, `docs/usage.md`,
  `docs/architecture.md`, `docs/testing.md`, `docs/development-journal.md`.
- `2026-11-15T22:44:55Z` — Generalize the editor action with one configurable
  command string, preferring Zed and its `zeditor` alias before VS Code. Parse
  quoted arguments and `||` alternatives without invoking a shell; append the
  selected path literally and skip unavailable executables. Preserve new-window
  defaults and report installed-editor failures without launching a fallback.
  Introduce `keys.open_in_editor`, retain the old key alias, and update hints and
  scope documentation. Verify priority, missing/disabled editors, custom quoted
  commands, legacy settings, folder/file arguments on both platforms, literal
  names, preview dispatch, stale paths, bounded errors, and cancellation using
  fake CLIs. `make check build`, gopls diagnostics, and `git diff --check` passed;
  the build retained the nonfatal module-cache write warning. Checked the official
  Zed and VS Code CLI references for new-window flags. Tests did not open desktop
  applications. No dependencies changed or publication ran. Paths: `editor.go`,
  `editor_test.go` (replacing `zed_test.go`), `actions.go`, `browser.go`, `view.go`,
  `config.go`, `config/default.json`, `config_test.go`, `cli.go`, `AGENTS.md`,
  `README.md`, `CHANGELOG.md`, `docs/configuration.md`, `docs/usage.md`,
  `docs/architecture.md`, `docs/testing.md`, `docs/development-journal.md`.
- `2026-11-15T23:00:13Z` — Split editor configuration into independent dir and
  file commands. Use Zed's `--classic` for dirs to reuse an open project or open
  a new project window; retain `--new` for individual files and VS Code fallback
  in both chains. Select the command from the highlighted entry type, including
  symlinks, and hide hints for independently disabled actions. Preserve the legacy
  shared command as an override for both types, with explicit per-type values
  taking precedence. Update help and documentation. Fake CLI checks verified
  default flags, custom commands in both views, symlink dispatch, independent
  disabling, preview behavior, and configuration precedence. Zed 1.28.1 accepted
  `--classic --help`; project matching was checked against its tagged source and
  VS Code's source, without opening desktop applications. `make check build`,
  gopls diagnostics, and `git diff --check` passed; the build retained the nonfatal
  module-cache write warning. No dependencies changed or publication ran. Paths:
  `config/default.json`, `config.go`, `config_test.go`, `editor.go`,
  `editor_test.go`, `actions.go`, `view.go`, `cli.go`, `README.md`, `CHANGELOG.md`,
  `docs/configuration.md`, `docs/usage.md`, `docs/architecture.md`,
  `docs/testing.md`, `docs/development-journal.md`.
- `2026-11-15T23:09:49Z` — Add a Rename footer button and `r` shortcut with
  prefilled, replace-on-type modal input. Rename a literal child name using native
  exclusive operations on macOS/Linux, preserving contents and refusing existing
  destinations. Capture and recheck entry identity without following the final
  symlink; protect navigation rows and current/starting locations and ancestors.
  Refresh tree caches and focus after success. Update scope, help, and docs.
  Tests cover both modes, nonempty dirs, links, hidden/Unicode/metacharacter names,
  invalid names, overwrite refusal, stale sources, cancellation, mouse activation,
  sanitization, input precedence, and footer ordering. `make check build
  cross-build`, gopls diagnostics, and `git diff --check` passed for the host and
  Linux/macOS amd64/arm64; builds retained the nonfatal module-cache write warning.
  No dependencies changed or publication ran. Paths: `mutation.go`,
  `mutation_test.go`, `rename_darwin.go`, `rename_linux.go`, `entry_buttons.go`,
  `browser.go`, `mouse.go`, `view.go`, `view_test.go`, `config.go`, `config_test.go`,
  `cli.go`, `AGENTS.md`, `README.md`, `CHANGELOG.md`, `docs/configuration.md`,
  `docs/usage.md`, `docs/architecture.md`, `docs/testing.md`,
  `docs/development-journal.md`.
- `2026-11-15T23:23:32Z` — Add a Delete footer button and Cmd+Backspace on macOS
  or Ctrl+Backspace on Linux. Require an exact basename in an initially empty
  confirmation modal; reject mismatches, stale targets, protected locations, and
  names that cannot be safely confirmed. Freeze an immutable target while the
  cancellable worker moves it through Foundation's system Trash API on macOS or
  `gio trash` on Linux. Never fall back to permanent deletion. Refresh listings
  after completion and clear confirmation after failure. Keep plain Backspace,
  other text inputs, and selected-path stdout unchanged. Update scope and docs.
  Fake-CLI tests cover both views, files, nonempty dirs, symlinks, exact matching,
  cancellation, unavailable/failed/no-op backends, replaced targets, shortcuts,
  repeat/release suppression, mouse activation, and small modal sizes. Native
  macOS tests moved disposable files, nonempty dirs, and valid/broken symlinks to
  Trash, checked identity and untouched link targets, and restored every fixture.
  The first full run exposed a test fixture the host filesystem refused to create
  because its name was invalid UTF-8; remove that fixture while retaining the
  application guard. The corrected `make check build cross-build`, gopls checks,
  ten-second keyboard fuzz run, and `git diff --check` passed. Builds retained the
  nonfatal module-cache write warning. Cross-builds cover Linux/macOS amd64/arm64;
  the Linux backend was exercised with fake CLIs, not a native desktop. No
  dependencies changed or publication ran. Paths: `trash.go`, `trash_test.go`,
  `trash_darwin_test.go`, `mutation.go`, `mutation_view.go`, `entry_buttons.go`,
  `actions.go`, `actions_test.go`, `terminal.go`, `keyboard.go`, `browser.go`,
  `view.go`, `cli.go`, `AGENTS.md`, `README.md`, `CHANGELOG.md`, `docs/usage.md`,
  `docs/architecture.md`, `docs/testing.md`, `docs/development-journal.md`.
- `2026-11-17T11:33:04Z` — Fix Cmd+Backspace (Finder's Cmd+Delete) in macOS
  terminals that send the legacy Ctrl+U byte. Inspect agterm's active keymap and
  config, and confirm Ghostty's default binding emits `0x15` even in kitty mode.
  Reproduce the ignored input with a failing decoder-to-browser regression test
  before accepting that byte on macOS. Preserve exact-name confirmation, modal
  input isolation, cancellation, and Linux's ignored legacy Ctrl+U. Document the
  unavoidable legacy Ctrl+U alias and retain distinct kitty Ctrl+U handling.
  Exercise decoded shortcuts through both views and fake Trash commands. Native
  macOS tests moved disposable files, nonempty dirs, and valid/broken symlinks
  through the corrected input path into system Trash, then restored all fixtures.
  `make check build cross-build`, gopls diagnostics, the ten-second keyboard fuzz
  run, and `git diff --check` passed. Cross-builds cover Linux/macOS amd64/arm64;
  builds retained the nonfatal module-cache write warning. No dependencies or
  terminal settings changed. The Homebrew installation remains separate from the
  local `bin/cdd` build; no publication ran. Paths: `terminal.go`,
  `terminal_test.go`, `trash_test.go`, `trash_darwin_test.go`, `cli.go`,
  `CHANGELOG.md`, `docs/usage.md`, `docs/architecture.md`, `docs/testing.md`,
  `docs/development-journal.md`.
- `2026-11-17T11:45:00Z` — Add a Finder footer button and `f` binding on macOS
  for the highlighted dir, including current/parent locations and dir links.
  Invoke `open -b com.apple.finder` through the existing external-action runner
  with one literal absolute path and dir validation before queueing and launch.
  Keep browsing position and shell selection intact. Hide the action for files
  and during text input; leave `f` literal while filtering or naming. Reserve
  the binding against editor overrides. Wrap footer actions into centered rows,
  retain mutation protection separately, and use rendered rows for hit testing.
  Update help, scope, configuration, and usage docs. Tests cover both views,
  current/parent dirs, links, stale/replaced dirs, file and platform guards,
  retained state, keyboard/mouse equivalence, narrow/clipped layouts, input
  precedence, literal CLI arguments, and bounded errors through a fake `open`.
  Checked the installed macOS `open` manual for explicit application selection;
  no desktop windows were opened by tests. `make check build cross-build`, gopls
  diagnostics, and `git diff --check` passed. Cross-builds cover Linux/macOS
  amd64/arm64; builds retained the nonfatal module-cache write warning. No
  dependencies changed or publication ran. Paths: `actions.go`, `browser.go`,
  `entry_buttons.go`, `view.go`, `config.go`, `config_test.go`, `mutation_test.go`,
  `finder_test.go`, `cli.go`, `AGENTS.md`, `README.md`, `CHANGELOG.md`,
  `docs/configuration.md`, `docs/usage.md`, `docs/architecture.md`,
  `docs/testing.md`, `docs/development-journal.md`.
- `2026-11-17T11:47:58Z` — Make the Delete footer button compact by displaying
  only its action label and using the same color as Finder and Rename. Keep the
  platform keyboard shortcut in controls help and retain exact-name confirmation.
  Update the changelog. Existing button, navigation, and Trash tests passed with
  `make check build`; gopls diagnostics and `git diff --check` also passed. The
  build retained the nonfatal module-cache write warning. No dependencies changed
  or publication ran. Paths: `entry_buttons.go`, `CHANGELOG.md`,
  `docs/development-journal.md`.
- `2026-11-17T21:17:56Z` — Add a compact Copy Path button and `y` binding in
  both views, including files, current/parent dirs, and valid or broken symlinks.
  Send the logical absolute path as exact stdin without shell quoting, added
  newlines, or file-content reads. Use macOS `pbcopy` with a UTF-8 locale and
  Linux `wl-copy` or `xclip` according to the active display and installed tools.
  Keep browsing and shell selection intact; report successful copying separately
  from errors. Accept the Linux clipboard owner's inherited stderr only after
  successful parent exit. Preserve confirmation through mouse release and retain
  rendered geometry for hit testing. Reserve the binding and update help/docs.
  Fake CLI tests verify literal Unicode/metacharacter/newline input, session
  selection, missing tools, bounded errors, cancellation, and forked ownership.
  Model/view tests cover both modes, links, navigation rows, wrapped/clipped
  buttons, input isolation, and safe feedback. Reproduce premature feedback
  dismissal with a failing regression test before fixing it. The user's real
  clipboard was not accessed; native desktop clipboard integration was not run.
  `make check build cross-build`, gopls diagnostics, and `git diff --check` passed.
  Cross-builds cover Linux/macOS amd64/arm64; builds retained the nonfatal
  module-cache write warning. No dependencies changed or publication ran.
  Paths: `clipboard.go`, `clipboard_test.go`, `actions.go`, `actions_test.go`,
  `browser.go`, `terminal.go`, `entry_buttons.go`, `view.go`, `config.go`,
  `config_test.go`, `cli.go`, `AGENTS.md`, `README.md`, `CHANGELOG.md`,
  `docs/configuration.md`, `docs/usage.md`, `docs/architecture.md`,
  `docs/testing.md`, `docs/development-journal.md`.
- `2026-11-18T19:42:34Z` — Add macOS Cmd+C and Cmd+V and Copy/Paste buttons
  for one file, folder or symlink in either view. Exchange file URLs with
  AppKit. Show the full destination for Enter confirmation, including the
  proposed copy suffix when a name is occupied. Stage the copy, recheck
  source and parent identities and install it with an exclusive rename.
  Tests exercised recursive folders, symlinks, concurrent pastes, cancellation,
  read-only staging cleanup, keyboard/button dispatch, bracketed paste and
  scrolling through a long destination. A private native pasteboard round trip
  passed for files, folders and symlinks, including broken links. It rejected
  text and multiple entries. The general pasteboard and Finder UI were not
  used during verification. `make check build cross-build` and a ten-second
  `FuzzDecodeKey` run passed. The builds reported a nonfatal module-cache
  write warning. Paths: `file_clipboard.go`, `file_clipboard_darwin_test.go`,
  `entry_copy.go`, `entry_copy_test.go`, `paste.go`, `paste_test.go`,
  `bracketed_paste.go`, `bracketed_paste_test.go`, `actions.go`, `browser.go`,
  `cli.go`, `entry_buttons.go`, `keyboard.go`, `mutation_view.go`, `outline.go`,
  `rename_darwin.go`, `rename_linux.go`, `terminal.go`, `view.go`, `AGENTS.md`,
  `README.md`, `CHANGELOG.md`, `docs/architecture.md`, `docs/testing.md`,
  `docs/usage.md`, `docs/development-journal.md`.
- `2026-11-18T23:26:26Z` — Remove the four-row Change Dir button and its mouse hit
  target from both views. Check that the browsing footer uses only the rows
  needed for entry actions and shortcut hints. Update CLI help and usage
  documentation. `make check build` passed with a nonfatal module-cache write
  warning. In a terminal, the built picker showed tree entries and footer
  actions, and Enter returned the highlighted directory.
  Paths: `view.go`, `view_test.go`, `mouse.go`, `mouse_test.go`, `keyboard_test.go`,
  `outline_test.go`, `cli.go`, `README.md`, `CHANGELOG.md`, `docs/usage.md`,
  `docs/architecture.md`, `docs/testing.md`, `docs/development-journal.md`.
- `2026-11-19T16:46:57Z` — Show the 50 most recently confirmed unique directory
  paths beside the tree. Save selections atomically under a lock shared by
  concurrent sessions. Anchor displayed paths at home or the filesystem root.
  In a terminal, verified that Tab switched full-width panels at 79 columns,
  both panels fit at 80, selecting an older entry moved it to the top on the
  next run, and cancelling produced empty stdout. `make check build cross-build`
  passed for Linux and macOS on amd64 and arm64 with a module-cache write warning.
  Paths: `recent.go`, `recent_view.go`, `recent_test.go`, `browser.go`, `cli.go`,
  `view.go`, `outline.go`, `mouse.go`, `entry_buttons.go`, `preferences.go`,
  `breadcrumb.go`, `breadcrumb_test.go`, `README.md`, `CHANGELOG.md`,
  `docs/usage.md`, `docs/architecture.md`, `docs/testing.md`, `docs/development-journal.md`.
- `2026-11-19T16:52:13Z` — Size the recent places panel from the widest
  expanded, filtered tree row and the starting path. Tests covered the exact
  split boundary, off-screen filenames, filtering, and mouse targets. In an
  82-column terminal, expanding a branch with a long filename hid history,
  Tab switched panels and back, and collapsing the branch restored the split.
  `make check build` passed with a module-cache write warning.
  Paths: `browser.go`, `outline.go`, `recent_view.go`, `view.go`, `cli.go`,
  `outline_test.go`, `recent_test.go`, `CHANGELOG.md`, `README.md`,
  `docs/architecture.md`, `docs/testing.md`, `docs/usage.md`, `docs/development-journal.md`.
- `2026-11-24T18:38:11Z` — Reveal a recent directory in the tree on Enter,
  modified Enter, or double-click. Tests covered follow-up actions, hidden
  symlink paths, stale tree listings, and unavailable destinations. In a
  terminal, Rename targeted the revealed directory, Esc cancelled with empty
  stdout, and a second Enter returned the exact path. `make check build` passed
  with a module-cache write warning.
  Updated `recent.go`, `outline.go`, `browser.go`, `mouse.go`, `view.go`,
  `recent_test.go`, `cli.go`, `README.md`, `CHANGELOG.md`, `docs/usage.md`,
  and `docs/development-journal.md`.
