package gitrail

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/Songmu/gitmock"
)

func TestParseNameStatus(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []FileChange
	}{
		{
			name:  "empty",
			input: "",
			want:  nil,
		},
		{
			name:  "added",
			input: "A\tsrc/new.go",
			want:  []FileChange{{Status: Added, Path: "src/new.go"}},
		},
		{
			name:  "modified",
			input: "M\tsrc/foo.go",
			want:  []FileChange{{Status: Modified, Path: "src/foo.go"}},
		},
		{
			name:  "deleted",
			input: "D\tsrc/old.go",
			want:  []FileChange{{Status: Deleted, Path: "src/old.go"}},
		},
		{
			name:  "renamed",
			input: "R95\tsrc/old.go\tsrc/new.go",
			want:  []FileChange{{Status: Modified, Path: "src/new.go", OldPath: "src/old.go"}},
		},
		{
			name:  "copied",
			input: "C100\tsrc/orig.go\tsrc/copy.go",
			want:  []FileChange{{Status: Added, Path: "src/copy.go"}},
		},
		{
			name:  "type_changed",
			input: "T\tsrc/link.go",
			want:  []FileChange{{Status: Modified, Path: "src/link.go"}},
		},
		{
			name:  "multiple",
			input: "A\tnew.go\nM\texisting.go\nD\told.go",
			want: []FileChange{
				{Status: Added, Path: "new.go"},
				{Status: Modified, Path: "existing.go"},
				{Status: Deleted, Path: "old.go"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseNameStatus(tc.input)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("parseNameStatus(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestParseRenameEvents(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  [][]renameEvent
	}{
		{
			name:  "empty",
			input: "",
			want:  nil,
		},
		{
			name:  "single_commit_single_rename",
			input: "abcdef1234567890abcdef1234567890abcdef12\nR100\told.go\tnew.go",
			want: [][]renameEvent{
				{{old: "old.go", new: "new.go"}},
			},
		},
		{
			name:  "two_commits",
			input: "abcdef1234567890abcdef1234567890abcdef12\nR100\tb.go\tc.go\nfedcba9876543210fedcba9876543210fedcba98\nR95\ta.go\tb.go",
			want: [][]renameEvent{
				{{old: "b.go", new: "c.go"}},
				{{old: "a.go", new: "b.go"}},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseRenameEvents(tc.input)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("parseRenameEvents(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestTrailBasic(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, gm *gitmock.GitMock)
		opts  trailOpts
		want  []FileChange
	}{
		{
			name: "add_modify_delete",
			setup: func(t *testing.T, gm *gitmock.GitMock) {
				testCommit(t, gm, "2026-01-10T00:00:00Z", "initial", map[string]string{
					"foo.go": "package main\n\nfunc foo() {}\n",
					"bar.go": "package main\n\nfunc bar() {}\n",
				})
				testCommit(t, gm, "2026-02-10T00:00:00Z", "changes", map[string]string{
					"foo.go": "package main\n\nfunc foo() { /* modified */ }\n",
					"baz.go": "package main\n\nfunc baz() {} // completely new file\n",
				})
				testDeleteCommit(t, gm, "2026-02-10T00:00:01Z", "bar.go", "delete bar")
			},
			opts: trailOpts{
				Since: "2026-01-15T00:00:00Z",
				Until: "2026-03-01T00:00:00Z",
			},
			want: []FileChange{
				{Status: Deleted, Path: "bar.go"},
				{Status: Added, Path: "baz.go"},
				{Status: Modified, Path: "foo.go"},
			},
		},
		{
			name: "same_commit_no_changes",
			setup: func(t *testing.T, gm *gitmock.GitMock) {
				testCommit(t, gm, "2026-02-01T00:00:00Z", "only commit", map[string]string{
					"foo.go": "package main\n",
				})
			},
			opts: trailOpts{
				Since: "2026-01-15T00:00:00Z",
				Until: "2026-02-15T00:00:00Z",
			},
			want: nil,
		},
		{
			name: "start_commit_fallback",
			setup: func(t *testing.T, gm *gitmock.GitMock) {
				testCommit(t, gm, "2026-02-01T00:00:00Z", "first commit", map[string]string{
					"foo.go": "package main\n",
				})
			},
			opts: trailOpts{
				Since: "2026-01-01T00:00:00Z",
				Until: "2026-03-01T00:00:00Z",
			},
			want: nil,
		},
		{
			name: "rename",
			setup: func(t *testing.T, gm *gitmock.GitMock) {
				testCommit(t, gm, "2026-01-10T00:00:00Z", "initial", map[string]string{
					"old.go": "package main\n",
				})
				testRenameCommit(t, gm, "2026-03-01T00:00:00Z", "old.go", "new.go", "rename")
			},
			opts: trailOpts{
				Since: "2026-01-05T00:00:00Z",
				Until: "2026-04-01T00:00:00Z",
			},
			want: []FileChange{
				{Status: Renamed, Path: "new.go", OldPath: "old.go"},
			},
		},
		{
			name: "rename_chain",
			setup: func(t *testing.T, gm *gitmock.GitMock) {
				testCommit(t, gm, "2026-01-10T00:00:00Z", "initial", map[string]string{
					"a.go": "package main\n",
				})
				testRenameCommit(t, gm, "2026-02-01T00:00:00Z", "a.go", "b.go", "rename a→b")
				testRenameCommit(t, gm, "2026-03-01T00:00:00Z", "b.go", "c.go", "rename b→c")
			},
			opts: trailOpts{
				Since: "2026-01-05T00:00:00Z",
				Until: "2026-04-01T00:00:00Z",
			},
			want: []FileChange{
				{Status: Renamed, Path: "c.go", OldPath: "a.go"},
			},
		},
		{
			name: "rename_with_modification",
			setup: func(t *testing.T, gm *gitmock.GitMock) {
				testCommit(t, gm, "2026-01-10T00:00:00Z", "initial", map[string]string{
					"old.go": "package main\n",
				})
				testCommit(t, gm, "2026-02-01T00:00:00Z", "modify", map[string]string{
					"old.go": "package main\n\n// updated\n",
				})
				testRenameCommit(t, gm, "2026-03-01T00:00:00Z", "old.go", "new.go", "rename")
			},
			opts: trailOpts{
				Since: "2026-01-05T00:00:00Z",
				Until: "2026-04-01T00:00:00Z",
			},
			want: []FileChange{
				{Status: Modified, Path: "new.go", OldPath: "old.go"},
			},
		},
		{
			name: "rename_new_file_ignored",
			setup: func(t *testing.T, gm *gitmock.GitMock) {
				testCommit(t, gm, "2026-01-10T00:00:00Z", "initial", map[string]string{
					"existing.go": "package main\n",
				})
				testCommit(t, gm, "2026-02-01T00:00:00Z", "add new.go", map[string]string{
					"new.go": "package main\n",
				})
				testRenameCommit(t, gm, "2026-03-01T00:00:00Z", "new.go", "renamed.go", "rename new→renamed")
			},
			opts: trailOpts{
				Since: "2026-01-05T00:00:00Z",
				Until: "2026-04-01T00:00:00Z",
			},
			want: []FileChange{
				{Status: Added, Path: "renamed.go"},
			},
		},
		{
			name: "pathspec",
			setup: func(t *testing.T, gm *gitmock.GitMock) {
				testCommit(t, gm, "2026-01-10T00:00:00Z", "initial", map[string]string{
					"src/foo.go":   "package main\n",
					"tests/bar.go": "package main\n",
				})
				testCommit(t, gm, "2026-02-10T00:00:00Z", "modify both", map[string]string{
					"src/foo.go":   "package main\n\n// modified\n",
					"tests/bar.go": "package main\n\n// modified\n",
				})
			},
			opts: trailOpts{
				Since:     "2026-01-15T00:00:00Z",
				Until:     "2026-03-01T00:00:00Z",
				Pathspecs: []string{"src/"},
			},
			want: []FileChange{
				{Status: Modified, Path: "src/foo.go"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gm := newTestRepo(t)
			tt.setup(t, gm)
			opts := tt.opts
			opts.Dir = gm.RepoPath()
			result, err := trail(context.Background(), &opts, os.Stderr)
			if err != nil {
				t.Fatalf("trail: %v", err)
			}
			if !reflect.DeepEqual(result.Changes, tt.want) {
				t.Errorf("Changes = %v, want %v", result.Changes, tt.want)
			}
		})
	}
}

func TestTrailError(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T, gm *gitmock.GitMock)
		opts         trailOpts
		wantExitCode int // 0 means just check error is non-nil
	}{
		{
			name: "end_commit_not_found",
			setup: func(t *testing.T, gm *gitmock.GitMock) {
				testCommit(t, gm, "2026-02-01T00:00:00Z", "first commit", map[string]string{
					"foo.go": "package main\n",
				})
			},
			opts: trailOpts{
				Since: "2025-01-01T00:00:00Z",
				Until: "2025-12-31T00:00:00Z",
			},
			wantExitCode: 2,
		},
		{
			name: "start_commit_not_found_empty_repo",
			setup: func(t *testing.T, gm *gitmock.GitMock) {
				// empty repo: no commits
			},
			opts: trailOpts{
				Since: "2026-01-01T00:00:00Z",
				Until: "2026-12-31T00:00:00Z",
			},
		},
		{
			name: "reversed_commits",
			setup: func(t *testing.T, gm *gitmock.GitMock) {
				testCommit(t, gm, "2026-01-10T00:00:00Z", "initial", map[string]string{
					"foo.go": "package main\n",
				})
				testCommit(t, gm, "2026-02-10T00:00:00Z", "update", map[string]string{
					"foo.go": "package main\n\n// v2\n",
				})
			},
			opts: trailOpts{
				Since: "2026-02-15T00:00:00Z",
				Until: "2026-01-15T00:00:00Z",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gm := newTestRepo(t)
			tt.setup(t, gm)
			opts := tt.opts
			opts.Dir = gm.RepoPath()
			_, err := trail(context.Background(), &opts, os.Stderr)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if tt.wantExitCode != 0 {
				ec, ok := err.(interface{ ExitCode() int })
				if !ok || ec.ExitCode() != tt.wantExitCode {
					t.Errorf("expected exit code %d, got err=%v", tt.wantExitCode, err)
				}
			}
		})
	}
}

func TestGitrailTrailMethod(t *testing.T) {
	gm := newTestRepo(t)
	ctx := context.Background()

	testCommit(t, gm, "2026-01-10T00:00:00Z", "initial", map[string]string{
		"foo.go": "package main\n\nfunc foo() {}\n",
	})
	testCommit(t, gm, "2026-02-10T00:00:00Z", "modify", map[string]string{
		"foo.go": "package main\n\nfunc foo() { /* modified */ }\n",
	})

	g := &Gitrail{Dir: gm.RepoPath(), ErrStream: os.Stderr}
	since := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	until := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

	result, err := g.Trail(ctx, "", since, until)
	if err != nil {
		t.Fatalf("Trail: %v", err)
	}

	want := []FileChange{
		{Status: Modified, Path: "foo.go"},
	}
	if !reflect.DeepEqual(result.Changes, want) {
		t.Errorf("Changes = %v, want %v", result.Changes, want)
	}
}

func TestGitrailTrailMethodNilReceiver(t *testing.T) {
	// A nil *Gitrail should not panic; it will fail because there's no valid Dir.
	// We check it returns an error gracefully rather than panicking.
	ctx := context.Background()
	var g *Gitrail
	since := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)

	// Run in an empty temporary directory so we don't depend on or slow down the
	// real current working directory (which might be a large git repository).
	origWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir to temp dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origWD)
	})

	// Should not panic; expect an error because there is no valid git repo / Dir.
	if _, err := g.Trail(ctx, "", since, until); err == nil {
		t.Fatalf("Trail with nil receiver and empty temp dir: expected error, got nil")
	}
}
