# Verification

Use the latest stable Go release and run commands from the source root. Install
both zsh and bash to exercise the complete shell integration suite. Tests for an
unavailable shell are skipped locally; CI supplies both shells.

## Automated checks

| Command | Scope |
| --- | --- |
| `make` / `make help` | Print available commands without building or installing |
| `make check` | Formatting, vet, and shuffled race-enabled application and build-tool tests |
| `make test` | Application and build-tool tests without race instrumentation |
| `make build` | Host binary and complete redistribution license files |
| `make cross-build` | Linux and macOS amd64/arm64 binaries with license files |
| `make dist` | Three preview archives, SHA-256 checksums, licenses, and curated notes |
| `make audit` | Checksums for both modules and vulnerability scans of the application and build tools |

Application tests cover browsing, parent selection, case-insensitive filtering,
hidden directories, symlinks, removed directories, cancellation, narrow views,
and display sanitization. Styled-view tests cover terminal color policy,
non-empty `NO_COLOR` (including `0`), tiny windows, scrolling with expanded help,
literal `?` filter input, and sanitization of dynamic text before styling.
Real zsh and bash subprocesses verify literal path
handling, trailing newlines, quotes, metacharacters, cancellation, partial output
on failure, and help passthrough.

Input-decoder tests cover complete and fragmented terminal sequences, Unicode,
malformed input, and Esc timeout behavior. For decoder changes, also run:

```sh
go test -run '^$' -fuzz FuzzDecodeKey -fuzztime=10s
```

Notice-generator tests check complete LICENSE/NOTICE/patent inputs, rejection of
missing or empty licenses, the Homebrew Go license layout, and exclusion of
development tools from application notices. `make build` inventories the four
distribution targets plus the host when different; a successful cross-compilation
is not a runtime test on that operating system or architecture.

Release-tool tests verify version-specific notes, required dates, duplicate and
missing sections, archive contents and executable permissions, SHA-256 sums,
deterministic packaging, and failure before output for missing license inputs.
Historical evidence below records the build matrix used at each stage. The
current distribution matrix contains Linux and macOS amd64/arm64.

`make audit` uses the pinned scanner in [tools/](../tools/README.md). First-time
tool downloads need access to the Go module proxy; vulnerability queries need
`vuln.go.dev`. A network error is not a clean audit result. A clean result reports
known reachable vulnerabilities only and does not establish overall safety.

## Manual terminal acceptance

Use a disposable directory tree and a real interactive terminal:

1. Run the quick start in both shells and select a child directory. Check `pwd`
   after the picker exits, then run `0go` from the new directory.
2. Open several levels, return to a parent, and confirm the previous child remains
   selected. Select the `.` row in an empty directory.
3. Filter a Unicode name, edit the filter with Backspace, and finish editing with
   Enter or Esc before selecting or cancelling.
4. Toggle hidden directories, open a directory symlink, and remove a selected
   directory from another shell to check recovery after an error.
5. Resize the terminal while browsing. Check narrow windows and long names.
   Open help with `?`, navigate with it expanded, and close it with Esc. Check
   that the selected row stays visible and the help does not cancel the picker.
   Compare dark and light terminal themes, then try `NO_COLOR=1 0go` and confirm
   that the selection remains visible without added colors.
6. Cancel with Esc, `q`, and Ctrl+C. Confirm the working directory is unchanged,
   normal input and echo return, and the cursor is visible.
7. Check handled termination signals and restore behavior separately from an
   uncatchable kill. Do not use a production shell for destructive process tests.

Automated model tests do not verify visual layout or real terminal restoration.
Terminal emulators can differ in key sequences and Unicode display width.
Use [usage](usage.md) for expected controls and the shell contract.

## Local evidence, 2026-11-13

- `make check build` passed on macOS arm64 with Go 1.27.4, including vet,
  race-enabled application tests, notice-generator tests, and host packaging.
- `make cross-build` produced macOS/Linux amd64/arm64 binaries. Build metadata
  showed only `golang.org/x/term` and `golang.org/x/sys` as external application
  modules. Generated notices include both modules' license and patent text and
  the effective Go standard library license.
- Installation into a temporary prefix passed. Installed license files matched
  the generated files, the binary reported its version, and both generated shell
  functions passed shell syntax checks. No user installation was performed.
- Relative documentation links passed a local check. Actionlint 1.7.12 accepted
  the workflow; ShellCheck was unavailable and was explicitly disabled for that
  local workflow check. Hosted Actions execution was not performed.
- Both module checksum checks passed. The pinned scanner started for the
  application and the notice generator, but both scans failed on DNS access to
  `vuln.go.dev`. Neither scan is recorded as clean. The owner's earlier audit of
  the unchanged application dependencies reported no vulnerabilities.
- Earlier manual native-terminal checks in this development session verified
  real shell directory changes, navigation/filtering, cancellation, and terminal
  restoration. They were not repeated for this documentation and packaging
  change. Cross-built binaries were not runtime-tested on other platforms.

GitHub security settings, publication, and hosted CI remain unverified. These
results describe local preparation, not a release or deployment.

## Manual update workflow validation, 2026-11-13

Actionlint 1.7.12 accepted both workflow files with ShellCheck explicitly disabled;
all updater `run` blocks also passed Bash syntax checks. The actual update block
was executed in a disposable Git clone against a cached module proxy. Starting
from Go 1.26.4 directives and older runtime modules, it updated both Go directives,
runtime packages, and existing tool dependencies. A second run reported no changes.
A synthetic runtime module-path change was rejected before PR publication.

The cached-proxy fixture disabled the unavailable checksum service only for that
isolated exercise; the production workflow keeps Go's default checksum settings.
The resulting candidate passed `make check build cross-build`, including
race-enabled tests and all four platform builds. Both module checksum checks
passed under the normal environment. The vulnerability scan could not reach
`vuln.go.dev`, so a complete successful updater run was not established.

No hosted workflow, repository permission change, authenticated Git push, or PR
creation was executed. The working repository's module files were not updated
by the fixture. GitHub authentication, permission checks, and PR CI approval
behavior require the published workflow and repository settings described in
[maintenance](maintenance.md).

## Release packaging validation, 2026-11-13

`make check build dist` passed on macOS arm64 with Go 1.27.4, including vet,
race-enabled application and build-tool tests, and the three cross-builds. Release
tests covered versioned notes, invalid dates and versions, required archive files,
executable permissions, checksums, reproducibility, and missing-license failures.

All three real preview archives passed SHA-256 verification. The macOS arm64
archive was extracted and its executable reported `0go dev`; its generated shell
function performed a real `cd` in a terminal. Linux archives were cross-built,
not runtime-tested here. Plain `make` in a directory containing only the Makefile
printed help and created no build output.

All workflow Bash blocks passed syntax checks, and actionlint accepted all three
workflows with ShellCheck disabled. Relative documentation links passed. Both
module checksum checks passed, while the vulnerability scan failed on unavailable
database DNS. No tag, hosted CI run, draft release, public release, signature, or
notarization was created. At this stage the release targets were Linux
amd64/arm64 and macOS arm64; Intel macOS packaging was added afterward.

## Interface validation, 2026-11-13

`make check build cross-build` passed on macOS arm64 with Go 1.27.4. Renderer
tests covered colored, non-colored, and unstyled output; valid UTF-8 and complete
ANSI styles when clipping; dynamic control-character sanitization; visible
selection with filters, errors, and help; and terminal sizes from 1x1 to 80x24.
Help-state tests verified that Esc closes help without cancelling, navigation
remains available, and `?` stays literal while editing a filter.

Native PTY sessions exercised the styled picker at 80x24 and 60x18, expanded help,
filtering, selection, cancellation, `NO_COLOR`, and `TERM=dumb`. A generated zsh
function selected a filtered child and changed the current shell's directory
with exit status zero, confirming that UI styles did not enter the captured path.
Cancellation returned 130 and restored the cursor and screen. This checked the
terminal protocol and interaction, not a screenshot-based review across themes.

No dependencies changed, so a new vulnerability scan was not run for this UI
change. Light/dark theme appearance, real window resizing, and Linux runtime
behavior remain manual acceptance checks. No hosted CI or publication ran.

## Intel macOS packaging validation, 2026-11-13

`make check build dist` passed on macOS arm64 with Go 1.27.4, including application
and build-tool vet/race tests, all four cross-builds, and preview packaging. All
four archives passed SHA-256 verification. The Intel macOS executable was
identified as Mach-O x86_64; Go build metadata confirmed GOOS=darwin,
GOARCH=amd64, and CGO_ENABLED=0. Its archive contains the executable with mode
0755, installation instructions, release notes, and both required license files.

The notice inventory now includes Intel macOS explicitly. No dependencies
changed. This establishes cross-compilation and packaging, not runtime testing
on an Intel Mac. Hosted CI and release publication were not performed.
