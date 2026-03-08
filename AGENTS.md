# Repository Guidelines

## Project Structure & Module Organization

This repository is a single-package Go CLI. Keep changes small and local.

- `main.go`: CLI entrypoint and module-path resolution logic.
- `main_test.go`: unit tests for argument parsing, remote normalization, and repo detection.
- `test.sh`: smoke test that verifies this repository resolves to `github.com/mkusaka/modi`.
- `.github/workflows/`: CI definitions for `go test` and workflow linting.
- `README.md`: user-facing usage and installation documentation.

Do not add extra packages or directories unless the CLI grows beyond a single file responsibility.

## Build, Test, and Development Commands

- `go run .`: run the CLI in the current repository.
- `go build ./...`: compile the binary and catch build regressions.
- `go test ./...`: run unit tests.
- `./test.sh`: run the repository-level smoke test.
- `gofmt -w main.go main_test.go`: apply standard Go formatting before commit.

Run `go test ./...` and `./test.sh` before opening a PR. If you touch workflows, also confirm the YAML remains valid.

## Coding Style & Naming Conventions

Follow standard Go style and let `gofmt` define formatting. Use tabs as emitted by `gofmt`, short focused functions, and explicit error returns. Prefer small helper functions such as `parseArgs` or `remoteModulePath` over embedding branching logic in `main()`.

Use descriptive test names with `TestXxx` and subtests for scenarios, for example `t.Run("linked worktree gitdir file", ...)`.

## Testing Guidelines

Tests use Go's built-in `testing` package. Favor table-driven tests for parsing and normalization logic, and temp directories for Git fixtures. Cover both success paths and user-facing failures, especially:

- remote URL forms (`git@...`, `https://...`, `ssh://...`)
- multi-remote selection
- nested directories, bare repos, and linked worktrees

## Commit & Pull Request Guidelines

Write commit messages in imperative English, e.g. `Expand README` or `Refactor CLI and add unit tests`.

PR titles and bodies must be in English. Create the PR body in a local Markdown file and use `gh pr create --body-file` or `gh pr edit --body-file`. Keep this section order:

`Summary`, `Changes`, `Why`, `Notes`, `Testing`

After creating or editing a PR, verify the rendered body with `gh pr view --json body --jq .body` and check it in the browser with `gh pr view --web`.
