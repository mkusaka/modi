package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/pkg/errors"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("%+v", err))
		os.Exit(1)
	}
	os.Exit(0)
}

func run() error {
	remoteName := flag.String("remoteName", "origin", "target remote name")
	targetPath := "."
	if len(os.Args) == 2 {
		targetPath = os.Args[1]
	}

	r, err := git.PlainOpenWithOptions(targetPath, &git.PlainOpenOptions{
		DetectDotGit:          true,
		EnableDotGitCommonDir: false,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	remotes, err := r.Remotes()
	if err != nil {
		return errors.WithStack(err)
	}

	var remote *git.Remote
	if len(remotes) == 1 {
		remote = remotes[0]
	} else {
		for _, r2 := range remotes {
			if r2.Config().Name == *remoteName {
				remote = r2
			}
		}
	}

	if remote == nil {
		if len(remotes) == 0 {
			fmt.Printf("cannot find target remote name: %s, with no remote", *remoteName)
		} else {
			var remoteNames []string

			for _, re := range remotes {
				remoteNames = append(remoteNames, re.Config().Name)
			}
			fmt.Printf("cannot find target remote name: %s, current remotes: %s", *remoteName, strings.Join(remoteNames, ""))
		}
		os.Exit(1)
	}

	// remote has only one url...? right?
	u := remote.Config().URLs[0]

	parsed, err := url.Parse(u)
	if err != nil {
		return errors.WithStack(err)
	}

	p := parsed.Path
	domain := parsed.Host
	hd := path.Join(domain, p)
	normalized := strings.TrimSuffix(hd, ".git")
	wd, err := os.Getwd()

	if err != nil {
		return errors.WithStack(err)
	}

	gitPath, err := detectGitPath(wd)

	if err != nil {
		return errors.WithStack(err)
	}

	rel, err := filepath.Rel(strings.TrimSuffix(gitPath, ".git"), wd)

	if err != nil {
		return errors.WithStack(err)
	}

	currentPackagePath := path.Join(normalized, strings.TrimSuffix(rel, ".git"))

	fmt.Println(currentPackagePath)
	return nil
}

func detectGitPath(p string) (string, error) {
	// normalize the p
	p, err := filepath.Abs(p)
	if err != nil {
		return "", errors.WithStack(err)
	}

	for {
		fi, err := os.Stat(path.Join(p, ".git"))
		if err == nil {
			if !fi.IsDir() {
				return "", fmt.Errorf(".git exist but is not a directory")
			}
			return path.Join(p, ".git"), nil
		}
		if !os.IsNotExist(err) {
			// unknown error
			return "", err
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

func isGitDir(p string) (bool, error) {
	markers := []string{"HEAD", "objects", "refs"}

	for _, marker := range markers {
		_, err := os.Stat(path.Join(p, marker))
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
