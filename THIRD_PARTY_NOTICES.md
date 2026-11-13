# Third-party notices

cdd uses the [MIT license](LICENSE). Its dependencies and the Go standard library
retain their upstream licenses.

`make notices`, `make build`, and `make cross-build` generate one
`bin/THIRD_PARTY_LICENSES.txt` from the application packages used on Linux
amd64/arm64, macOS amd64/arm64, and the local build host when different. It includes
module names and versions, complete
upstream LICENSE, NOTICE, and patent files when present, and the Go standard
library license. Missing or empty primary license files fail the build.

The inventory excludes test-only and development-tool dependencies unless the
application actually imports them. The generator does not vendor source code or
maintain a second copy of upstream licenses in the repository. Generated notices
and binaries are excluded from Git; versions and checksums remain in the module
files. Module replacements require licensing review before packaging.

Builds also copy the project's own MIT license to `bin/LICENSE`. Keep it and
`bin/THIRD_PARTY_LICENSES.txt` with any redistributed standalone executable.
`make dist` includes both files in every release archive alongside its executable,
instructions, and user-facing release notes.
`make install` places license files under
`$(PREFIX)/share/licenses/cdd` by default; `LICENSEDIR` can override that location.

After changing runtime dependencies, Go toolchains, or build targets, regenerate
and review the notices and run the relevant checks. These files do not relicense
third-party code under MIT.
