package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
)

func TestParseArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		want    options
		wantErr string
	}{
		{
			name: "defaults",
			args: nil,
			want: options{
				remoteName: "origin",
				targetPath: ".",
			},
		},
		{
			name: "remoteName and path",
			args: []string{"-remoteName", "upstream", "internal/cli"},
			want: options{
				remoteName: "upstream",
				targetPath: "internal/cli",
			},
		},
		{
			name:    "too many paths",
			args:    []string{"a", "b"},
			wantErr: "expected at most one path argument",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseArgs(tt.args)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("parseArgs returned error: %v", err)
			}

			if got != tt.want {
				t.Fatalf("parseArgs returned %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestRemoteModulePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		rawURL  string
		want    string
		wantErr string
	}{
		{
			name:   "scp style",
			rawURL: "git@github.com:mkusaka/modi.git",
			want:   "github.com/mkusaka/modi",
		},
		{
			name:   "https",
			rawURL: "https://github.com/mkusaka/modi.git",
			want:   "github.com/mkusaka/modi",
		},
		{
			name:   "ssh with scheme",
			rawURL: "ssh://git@github.com/mkusaka/modi.git",
			want:   "github.com/mkusaka/modi",
		},
		{
			name:    "unsupported local path",
			rawURL:  "/tmp/modi.git",
			wantErr: "unsupported remote URL",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := remoteModulePath(tt.rawURL)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("remoteModulePath returned error: %v", err)
			}

			if got != tt.want {
				t.Fatalf("remoteModulePath returned %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectRepoRoot(t *testing.T) {
	t.Parallel()

	t.Run("nested repository directory", func(t *testing.T) {
		t.Parallel()

		root := initRepo(t, false, nil)
		nested := filepath.Join(root, "internal", "cli")
		if err := os.MkdirAll(nested, 0o755); err != nil {
			t.Fatalf("mkdir nested: %v", err)
		}

		got, err := detectRepoRoot(nested)
		if err != nil {
			t.Fatalf("detectRepoRoot returned error: %v", err)
		}

		if got != root {
			t.Fatalf("detectRepoRoot returned %q, want %q", got, root)
		}
	})

	t.Run("linked worktree gitdir file", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		nested := filepath.Join(root, "internal", "cli")
		if err := os.MkdirAll(nested, 0o755); err != nil {
			t.Fatalf("mkdir nested: %v", err)
		}

		actualGitDir := filepath.Join(root, ".worktree-git")
		if err := os.MkdirAll(filepath.Join(actualGitDir, "objects"), 0o755); err != nil {
			t.Fatalf("mkdir objects: %v", err)
		}
		if err := os.MkdirAll(filepath.Join(actualGitDir, "refs"), 0o755); err != nil {
			t.Fatalf("mkdir refs: %v", err)
		}
		if err := os.WriteFile(filepath.Join(actualGitDir, "HEAD"), []byte("ref: refs/heads/main\n"), 0o644); err != nil {
			t.Fatalf("write HEAD: %v", err)
		}
		if err := os.WriteFile(filepath.Join(root, ".git"), []byte("gitdir: .worktree-git\n"), 0o644); err != nil {
			t.Fatalf("write .git file: %v", err)
		}

		got, err := detectRepoRoot(nested)
		if err != nil {
			t.Fatalf("detectRepoRoot returned error: %v", err)
		}

		if got != root {
			t.Fatalf("detectRepoRoot returned %q, want %q", got, root)
		}
	})

	t.Run("bare repository", func(t *testing.T) {
		t.Parallel()

		root := initRepo(t, true, nil)

		got, err := detectRepoRoot(root)
		if err != nil {
			t.Fatalf("detectRepoRoot returned error: %v", err)
		}

		if got != root {
			t.Fatalf("detectRepoRoot returned %q, want %q", got, root)
		}
	})
}

func TestResolveModulePath(t *testing.T) {
	t.Parallel()

	t.Run("uses path relative to repository root", func(t *testing.T) {
		t.Parallel()

		root := initRepo(t, false, map[string]string{
			"origin": "git@github.com:example/modi.git",
		})
		nested := filepath.Join(root, "internal", "cli")
		if err := os.MkdirAll(nested, 0o755); err != nil {
			t.Fatalf("mkdir nested: %v", err)
		}

		got, err := resolveModulePath(nested, "origin")
		if err != nil {
			t.Fatalf("resolveModulePath returned error: %v", err)
		}

		if got != "github.com/example/modi/internal/cli" {
			t.Fatalf("resolveModulePath returned %q", got)
		}
	})

	t.Run("selects requested remote when multiple remotes exist", func(t *testing.T) {
		t.Parallel()

		root := initRepo(t, false, map[string]string{
			"origin":   "git@github.com:example/origin.git",
			"upstream": "https://github.com/example/upstream.git",
		})

		got, err := resolveModulePath(root, "upstream")
		if err != nil {
			t.Fatalf("resolveModulePath returned error: %v", err)
		}

		if got != "github.com/example/upstream" {
			t.Fatalf("resolveModulePath returned %q", got)
		}
	})

	t.Run("returns helpful error for missing remote", func(t *testing.T) {
		t.Parallel()

		root := initRepo(t, false, map[string]string{
			"origin":   "git@github.com:example/modi.git",
			"upstream": "https://github.com/example/upstream.git",
		})

		_, err := resolveModulePath(root, "fork")
		if err == nil || !strings.Contains(err.Error(), "cannot find target remote name") {
			t.Fatalf("expected missing remote error, got %v", err)
		}
	})
}

func initRepo(t *testing.T, bare bool, remotes map[string]string) string {
	t.Helper()

	root := t.TempDir()
	repo, err := git.PlainInit(root, bare)
	if err != nil {
		t.Fatalf("init repo: %v", err)
	}

	for name, remoteURL := range remotes {
		if _, err := repo.CreateRemote(&gitconfig.RemoteConfig{
			Name: name,
			URLs: []string{remoteURL},
		}); err != nil {
			t.Fatalf("create remote %q: %v", name, err)
		}
	}

	return root
}
