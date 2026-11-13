# Dependency updates

## Update with a button

After publishing this workflow on the default branch:

1. Open **Actions** on GitHub.
2. Select **Update dependencies**.
3. Choose **Run workflow**, keep the default branch selected, and run it.

The workflow installs the latest stable Go release, updates the `go` directives
in both modules to that exact release, and refreshes application dependencies
and the pinned Go tools with their checksums. Go's normal minor/patch upgrade
rules apply; a new major module path requires a separate reviewed migration.
Preview and development Go releases are not selected. An incompatible dependency
that requires a newer toolchain fails the run instead of silently selecting one.

The candidate must pass formatting, vet, race-enabled tests, host packaging,
Linux amd64/arm64 and macOS amd64/arm64 cross-builds, module verification, and
vulnerability scans on Ubuntu. An added or removed runtime module path stops the workflow for manual
review. Generated notices are rebuilt; tool dependency changes remain visible in
the pull request. These checks do not replace review of maintainers or source.

If versions changed, the workflow appends a UTC development-journal entry and creates
a pull request titled **Update Go and dependencies**. Review it and merge it using
the normal repository process. The update workflow itself does not merge or
release anything; merging its PR into main triggers the ordinary automatic
release and Homebrew tap update.
If an update PR is already open, the run links to it and stops. If nothing changed,
it reports that the project is up to date without creating a commit or PR.
Each new PR uses a new branch; remote history is never force-pushed or rewritten.

GitHub Actions versions are maintained by the existing weekly Dependabot
configuration. They are outside this button's Go/module update scope.

## GitHub setup

The workflow must be on the default branch before the button is available.
Running it requires repository write access. It runs only from the default
branch and uses the repository's built-in `GITHUB_TOKEN`; no personal token or
extra third-party Action is required.

Enable **Settings > Actions > General > Workflow permissions > Allow GitHub
Actions to create and approve pull requests**. Organization policy may prevent
that setting; in that case an administrator must enable the capability. The
workflow requests `contents: write` and `pull-requests: write` and supplies the
token only to the GitHub CLI steps that find or publish the PR.

GitHub may require **Approve workflows to run** on a PR created by its built-in
token. Approve the ordinary CI when prompted; it provides the separate Linux
and macOS checks. Successful Ubuntu validation in the updater does not stand in
for those platform checks. Configure branch protection separately if they must
be required before merging.

Repository settings, hosted execution, PR creation, and permissions cannot be
validated by editing local workflow files. No GitHub setting is changed locally.

GitHub documents the [manual workflow button](https://docs.github.com/en/actions/how-tos/manage-workflow-runs/manually-run-a-workflow),
[repository Actions permissions](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/enabling-features-for-your-repository/managing-github-actions-settings-for-a-repository),
and [token-triggered workflow behavior](https://docs.github.com/en/actions/how-tos/write-workflows/choose-when-workflows-run/trigger-a-workflow).

## Review an update

- Check the selected Go release and the module version diff in the PR body.
- Review both modules' requirements and checksums, including tool transitives.
- Inspect regenerated licenses when dependencies change their license inputs.
- Check the updater run and separate PR CI, including macOS runtime tests.
- Treat network failures or scanner findings as failed checks, never clean audits.

For a manual update, use the latest stable Go and run the same Go commands shown
in `.github/workflows/update-dependencies.yml`, then `make check build cross-build`
and `make audit`. Keep the change journal with the same commit.
