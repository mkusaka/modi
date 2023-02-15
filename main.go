package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"path"
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

	fmt.Println(normalized)
	return nil
}
