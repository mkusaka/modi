# modi

`modi` prints the Go module path that matches the current Git repository remote and the directory you point it at.

It is useful when you are creating a new `go.mod` and want to avoid manually translating a Git remote such as `git@github.com:owner/repo.git` into `github.com/owner/repo`.

## What it does

- Reads the repository remote URL from Git metadata.
- Supports common SSH and HTTPS remotes such as `git@github.com:owner/repo.git` and `https://github.com/owner/repo.git`.
- Appends the path relative to the repository root, so it also works from nested directories inside a repository.
- Works in regular repositories, bare repositories, and linked worktrees.

## Installation

```sh
go install github.com/mkusaka/modi@latest
```

## Usage

Run it in the repository root:

```sh
modi
# github.com/owner/repo

go mod init "$(modi)"
```

Run it from a nested directory:

```sh
cd internal/cli
modi
# github.com/owner/repo/internal/cli

go mod init "$(modi)"
```

You can also pass a target path explicitly:

```sh
modi internal/cli
# github.com/owner/repo/internal/cli
```

When a repository has more than one remote, `origin` is used by default. You can select another one with `-remoteName`:

```sh
modi -remoteName upstream
modi -remoteName upstream internal/cli
```

If the repository only has one remote, `modi` uses it regardless of its name.

## Requirements

- Run `modi` inside a Git repository, a linked worktree, or a bare repository.
- The repository must have at least one remote configured.
- The remote host and path should match the module import path you want to use.

`modi` does not rewrite vanity import paths. If your module path intentionally differs from the Git remote, edit `go.mod` manually.

## Testing

```sh
go test ./...
./test.sh
```

`go test` covers argument parsing, remote URL normalization, repository root detection, and module path resolution. `./test.sh` is a smoke test against this repository.
