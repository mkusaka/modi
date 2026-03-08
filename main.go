package main

import (
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/pkg/errors"
)

func main() {
	modulePath, err := run(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("%+v", err))
		os.Exit(1)
	}

	fmt.Println(modulePath)
}

type options struct {
	remoteName string
	targetPath string
}

func run(args []string) (string, error) {
	opts, err := parseArgs(args)
	if err != nil {
		return "", err
	}

	return resolveModulePath(opts.targetPath, opts.remoteName)
}

func parseArgs(args []string) (options, error) {
	opts := options{
		remoteName: "origin",
		targetPath: ".",
	}

	fs := flag.NewFlagSet("modi", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&opts.remoteName, "remoteName", opts.remoteName, "target remote name")

	if err := fs.Parse(args); err != nil {
		return options{}, errors.WithStack(err)
	}

	switch len(fs.Args()) {
	case 0:
		return opts, nil
	case 1:
		opts.targetPath = fs.Arg(0)
		return opts, nil
	default:
		return options{}, fmt.Errorf("expected at most one path argument, got %d", len(fs.Args()))
	}
}

func resolveModulePath(targetPath, remoteName string) (string, error) {
	absTargetPath, err := filepath.Abs(targetPath)
	if err != nil {
		return "", errors.WithStack(err)
	}

	r, err := git.PlainOpenWithOptions(absTargetPath, &git.PlainOpenOptions{
		DetectDotGit:          true,
		EnableDotGitCommonDir: true,
	})
	if err != nil {
		return "", errors.WithStack(err)
	}

	remotes, err := r.Remotes()
	if err != nil {
		return "", errors.WithStack(err)
	}

	remote, err := selectRemote(remotes, remoteName)
	if err != nil {
		return "", err
	}

	if len(remote.Config().URLs) == 0 {
		return "", fmt.Errorf("remote %q does not have any URLs", remote.Config().Name)
	}

	prefix, err := remoteModulePath(remote.Config().URLs[0])
	if err != nil {
		return "", err
	}

	repoRoot, err := detectRepoRoot(absTargetPath)
	if err != nil {
		return "", err
	}

	rel, err := filepath.Rel(repoRoot, absTargetPath)
	if err != nil {
		return "", errors.WithStack(err)
	}

	currentPackagePath := prefix
	if rel != "." {
		currentPackagePath = path.Join(prefix, filepath.ToSlash(rel))
	}

	return currentPackagePath, nil
}

func selectRemote(remotes []*git.Remote, remoteName string) (*git.Remote, error) {
	if len(remotes) == 0 {
		return nil, fmt.Errorf("cannot find target remote name: %s, with no remote", remoteName)
	}

	if len(remotes) == 1 {
		return remotes[0], nil
	}

	for _, remote := range remotes {
		if remote.Config().Name == remoteName {
			return remote, nil
		}
	}

	remoteNames := make([]string, 0, len(remotes))
	for _, remote := range remotes {
		remoteNames = append(remoteNames, remote.Config().Name)
	}

	return nil, fmt.Errorf("cannot find target remote name: %s, current remotes: %s", remoteName, strings.Join(remoteNames, ", "))
}

func remoteModulePath(rawURL string) (string, error) {
	if strings.Contains(rawURL, "://") {
		parsed, err := url.Parse(rawURL)
		if err != nil {
			return "", errors.WithStack(err)
		}

		host := parsed.Hostname()
		if host == "" {
			return "", fmt.Errorf("remote URL missing host: %s", rawURL)
		}

		return normalizeModulePath(host, parsed.Path), nil
	}

	scpLike := rawURL
	if idx := strings.LastIndex(scpLike, "@"); idx >= 0 {
		scpLike = scpLike[idx+1:]
	}

	host, repoPath, ok := strings.Cut(scpLike, ":")
	if !ok || host == "" || repoPath == "" {
		return "", fmt.Errorf("unsupported remote URL: %s", rawURL)
	}

	return normalizeModulePath(host, repoPath), nil
}

func normalizeModulePath(host, repoPath string) string {
	return strings.TrimSuffix(path.Join(host, strings.TrimPrefix(filepath.ToSlash(repoPath), "/")), ".git")
}

func detectRepoRoot(p string) (string, error) {
	p, err := filepath.Abs(p)
	if err != nil {
		return "", errors.WithStack(err)
	}

	for {
		fi, err := os.Stat(filepath.Join(p, ".git"))
		if err == nil {
			if fi.IsDir() {
				return p, nil
			}

			if fi.Mode().IsRegular() {
				ok, err := isGitDirFile(filepath.Join(p, ".git"))
				if err != nil {
					return "", err
				}
				if ok {
					return p, nil
				}
			}

			return "", fmt.Errorf(".git exists but is neither a directory nor a valid gitdir file")
		}
		if !os.IsNotExist(err) {
			// unknown error
			return "", errors.WithStack(err)
		}

		// detect bare repo
		ok, err := isGitDir(p)
		if err != nil {
			return "", err
		}
		if ok {
			return p, nil
		}

		if parent := filepath.Dir(p); parent == p {
			return "", errors.New(".git not found")
		} else {
			p = parent
		}
	}
}

func isGitDirFile(p string) (bool, error) {
	content, err := os.ReadFile(p)
	if err != nil {
		return false, errors.WithStack(err)
	}

	line := strings.TrimSpace(string(content))
	if !strings.HasPrefix(line, "gitdir:") {
		return false, nil
	}

	gitDir := strings.TrimSpace(strings.TrimPrefix(line, "gitdir:"))
	if gitDir == "" {
		return false, fmt.Errorf("gitdir file does not specify a path")
	}

	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(filepath.Dir(p), gitDir)
	}

	return isGitDir(gitDir)
}

func isGitDir(p string) (bool, error) {
	markers := []string{"HEAD", "objects", "refs"}

	for _, marker := range markers {
		_, err := os.Stat(filepath.Join(p, marker))
		if err == nil {
			continue
		}
		if !os.IsNotExist(err) {
			// unknown error
			return false, errors.WithStack(err)
		} else {
			return false, nil
		}
	}

	return true, nil
}
