# Releases

## Supported archives

| Operating system | Architecture | Filename suffix |
| --- | --- | --- |
| Linux | x64 / amd64 | `linux_amd64.tar.gz` |
| Linux | arm64 | `linux_arm64.tar.gz` |
| macOS | Apple Silicon / arm64 | `darwin_arm64.tar.gz` |

Each archive contains `0go`, `README.md` with standalone instructions, the MIT
`LICENSE`, `THIRD_PARTY_LICENSES.txt`, and version-specific `RELEASE_NOTES.md`.
`SHA256SUMS` accompanies the archives on GitHub. No Go installation is needed
to use a prebuilt executable. The workflow does not sign or notarize binaries.

## Use an archive

Download the archive for your machine and the SHA256SUMS file from the same
GitHub Release. For example, after substituting an available version:

```sh
shasum -a 256 0go_vX.Y.Z_linux_amd64.tar.gz
```

Compare the result with the entry for that exact filename in SHA256SUMS.
Extract into an empty directory, then initialize the shell function:

```sh
tar -xzf 0go_vX.Y.Z_linux_amd64.tar.gz
eval "$(./0go --init zsh)"
0go
```

For bash, use `--init bash`. The function lasts for the current shell and records
the absolute binary path. Place it at a permanent location before adding its
`--init` line to shell startup files. See [usage](usage.md) for details.

## Local packaging

```sh
make dist
```

This builds the three targets and writes preview archives under `dist/dev/`,
using the Unreleased section from CHANGELOG.md. It does not publish anything.
The archive format and checksums are deterministic for identical inputs.

For a versioned local preview, prepare a matching section and run:

```sh
make dist VERSION=vX.Y.Z
```

Replace `vX.Y.Z` with the chosen stable version. Output is under `dist/vX.Y.Z/`.
`--version` in each binary receives the same version. Generated archives and
license files are excluded from Git. `make clean` removes them.

## Write release notes

Use CHANGELOG.md for changes users need to know: additions, behavior changes,
fixes, and compatibility or installation requirements. The engineering journal
with UTC timestamps and verification details remains in
[development-journal.md](development-journal.md).

Before a release, move completed Unreleased changes under a heading like:

```markdown
## v0.1.0 - 2026-11-13

### Added

- Browse and select directories with keyboard navigation.

### Fixed

- Describe the user-visible problem and resulting behavior.
```

Use the chosen version and actual release date. Keep an Unreleased section for
future work. A versioned package fails if its section is absent, duplicated,
undated, or empty; it never substitutes Unreleased notes for a tagged release.

## Publish through GitHub Actions

Commit the dated notes and the related source changes first. When the maintainer
pushes an existing stable `vX.Y.Z` Git tag to GitHub, the Release workflow:

1. Calls the ordinary CI, including Linux/macOS tests and vulnerability auditing.
2. Builds and packages Linux amd64/arm64 and macOS arm64 binaries.
3. Verifies the archive checksums.
4. Creates a draft GitHub Release with the archives, checksums, and curated notes.
5. Publishes the draft only after uploads succeed.

The workflow uses the built-in repository token, with write permission only in
the publication job. It verifies that the tag already exists and does not create
one. Stable tags only are supported; preview versions and `dev` cannot publish.
An existing release is not overwritten on a rerun. If uploading fails after draft
creation, inspect that draft before retrying; the workflow does not delete it.

Local packaging does not establish that hosted CI or publication works. No tag,
release, signing identity, or GitHub repository setting is created locally.
