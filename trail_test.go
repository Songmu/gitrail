package gitrail

import (
	"context"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Songmu/gitmock"
)

// newTestRepo creates a new gitmock repository with "main" as the default branch.
func newTestRepo(t *testing.T) *gitmock.GitMock {
	t.Helper()
	gm, err := gitmock.New("")
	if err != nil {
		t.Fatalf("gitmock.New: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(gm.RepoPath()) })
	if _, _, err := gm.Init("-b", "main"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	if _, _, err := gm.Do("config", "user.email", "test@example.com"); err != nil {
		t.Fatalf("git config user.email: %v", err)
	}
	if _, _, err := gm.Do("config", "user.name", "Test User"); err != nil {
		t.Fatalf("git config user.name: %v", err)
	}
	return gm
}

// testCommit creates a commit at the given ISO date, adding or modifying the provided files.
func testCommit(t *testing.T, gm *gitmock.GitMock, date, msg string, files map[string]string) string {
	t.Helper()
	t.Setenv("GIT_COMMITTER_DATE", date)
	for path, content := range files {
		if err := gm.PutFile(path, content); err != nil {
			t.Fatalf("PutFile %s: %v", path, err)
		}
		if _, _, err := gm.Add(path); err != nil {
			t.Fatalf("git add %s: %v", path, err)
		}
	}
	if _, _, err := gm.Commit("--date="+date, "-m", msg); err != nil {
		t.Fatalf("git commit: %v", err)
	}
	out, _, _ := gm.Do("rev-parse", "HEAD")
	return strings.TrimSpace(out)
}

// testRenameCommit creates a commit that renames a file at the given date.
func testRenameCommit(t *testing.T, gm *gitmock.GitMock, date, oldPath, newPath, msg string) string {
	t.Helper()
	t.Setenv("GIT_COMMITTER_DATE", date)
	if _, _, err := gm.Mv(oldPath, newPath); err != nil {
		t.Fatalf("git mv %s %s: %v", oldPath, newPath, err)
	}
	if _, _, err := gm.Commit("--date="+date, "-m", msg); err != nil {
		t.Fatalf("git commit: %v", err)
	}
	out, _, _ := gm.Do("rev-parse", "HEAD")
	return strings.TrimSpace(out)
}

// testDeleteCommit creates a commit that deletes a file at the given date.
func testDeleteCommit(t *testing.T, gm *gitmock.GitMock, date, path, msg string) string {
	t.Helper()
	t.Setenv("GIT_COMMITTER_DATE", date)
	if _, _, err := gm.Rm(path); err != nil {
		t.Fatalf("git rm %s: %v", path, err)
	}
	if _, _, err := gm.Commit("--date="+date, "-m", msg); err != nil {
		t.Fatalf("git commit: %v", err)
	}
	out, _, _ := gm.Do("rev-parse", "HEAD")
	return strings.TrimSpace(out)
}

func TestParseNameStatus(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   []FileChange
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
			name: "multiple",
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
		name   string
		input  string
		want   [][]renameEvent
	}{
		{
			name:  "empty",
			input: "",
			want:  nil,
		},
		{
			name: "single_commit_single_rename",
			input: "abcdef1234567890abcdef1234567890abcdef12\nR100\told.go\tnew.go",
			want: [][]renameEvent{
				{{old: "old.go", new: "new.go"}},
			},
		},
		{
			name: "two_commits",
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
	gm := newTestRepo(t)
	ctx := context.Background()

	// Initial commit: create foo.go and bar.go
	testCommit(t, gm, "2026-01-10T00:00:00Z", "initial", map[string]string{
		"foo.go": "package main\n\nfunc foo() {}\n",
		"bar.go": "package main\n\nfunc bar() {}\n",
	})

	// Commit 1: modify foo.go, delete bar.go, add baz.go
	testCommit(t, gm, "2026-02-10T00:00:00Z", "changes", map[string]string{
		"foo.go": "package main\n\nfunc foo() { /* modified */ }\n",
		"baz.go": "package main\n\nfunc baz() {} // completely new file\n",
	})
	testDeleteCommit(t, gm, "2026-02-10T00:00:01Z", "bar.go", "delete bar")

	result, err := trail(ctx, &trailOpts{
		Dir:   gm.RepoPath(),
		Since: "2026-01-15T00:00:00Z",
		Until: "2026-03-01T00:00:00Z",
	}, os.Stderr)
	if err != nil {
		t.Fatalf("trail: %v", err)
	}

	want := []FileChange{
		{Status: Deleted, Path: "bar.go"},
		{Status: Added, Path: "baz.go"},
		{Status: Modified, Path: "foo.go"},
	}
	if !reflect.DeepEqual(result.Changes, want) {
		t.Errorf("Changes = %v, want %v", result.Changes, want)
	}
}

func TestTrailSameCommit(t *testing.T) {
	gm := newTestRepo(t)
	ctx := context.Background()

	testCommit(t, gm, "2026-02-01T00:00:00Z", "only commit", map[string]string{
		"foo.go": "package main\n",
	})

	// both since and until resolve to the same commit
	result, err := trail(ctx, &trailOpts{
		Dir:   gm.RepoPath(),
		Since: "2026-01-15T00:00:00Z",
		Until: "2026-02-15T00:00:00Z",
	}, os.Stderr)
	if err != nil {
		t.Fatalf("trail: %v", err)
	}

	if result.StartCommit != result.EndCommit {
		t.Errorf("expected StartCommit == EndCommit, got %s vs %s", result.StartCommit, result.EndCommit)
	}
	if len(result.Changes) != 0 {
		t.Errorf("expected no changes, got %v", result.Changes)
	}
}

func TestTrailStartCommitFallback(t *testing.T) {
	gm := newTestRepo(t)
	ctx := context.Background()

	// Only one commit, created after since
	h := testCommit(t, gm, "2026-02-01T00:00:00Z", "first commit", map[string]string{
		"foo.go": "package main\n",
	})

	// since is before the only commit → fallback triggers
	result, err := trail(ctx, &trailOpts{
		Dir:   gm.RepoPath(),
		Since: "2026-01-01T00:00:00Z",
		Until: "2026-03-01T00:00:00Z",
	}, os.Stderr)
	if err != nil {
		t.Fatalf("trail: %v", err)
	}

	// fallback start commit = first commit after since = h
	// end commit also = h (--before=2026-03-01 finds h)
	if result.StartCommit != h {
		t.Errorf("StartCommit = %s, want %s", result.StartCommit, h)
	}
	if result.StartCommit != result.EndCommit {
		t.Errorf("expected same commit, got %s vs %s", result.StartCommit, result.EndCommit)
	}
	if len(result.Changes) != 0 {
		t.Errorf("expected no changes for same commit, got %v", result.Changes)
	}
}

func TestTrailEndCommitNotFound(t *testing.T) {
	gm := newTestRepo(t)
	ctx := context.Background()

	testCommit(t, gm, "2026-02-01T00:00:00Z", "first commit", map[string]string{
		"foo.go": "package main\n",
	})

	// until is before any commits
	_, err := trail(ctx, &trailOpts{
		Dir:   gm.RepoPath(),
		Since: "2025-01-01T00:00:00Z",
		Until: "2025-12-31T00:00:00Z",
	}, os.Stderr)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if ec, ok := err.(interface{ ExitCode() int }); !ok || ec.ExitCode() != 2 {
		t.Errorf("expected exit code 2, got err=%v", err)
	}
}

func TestTrailStartCommitNotFound(t *testing.T) {
	gm := newTestRepo(t)
	ctx := context.Background()

	// empty repo: no commits at all
	_, _, _ = gm.Do("config", "user.email", "test@example.com")
	_, _, _ = gm.Do("config", "user.name", "Test User")

	_, err := trail(ctx, &trailOpts{
		Dir:   gm.RepoPath(),
		Since: "2026-01-01T00:00:00Z",
		Until: "2026-12-31T00:00:00Z",
	}, os.Stderr)
	if err == nil {
		t.Fatal("expected error for empty repo, got nil")
	}
	if ec, ok := err.(interface{ ExitCode() int }); !ok || ec.ExitCode() != 1 {
		t.Errorf("expected exit code 1, got err=%v", err)
	}
}

func TestTrailReversedCommits(t *testing.T) {
	gm := newTestRepo(t)
	ctx := context.Background()

	testCommit(t, gm, "2026-01-10T00:00:00Z", "initial", map[string]string{
		"foo.go": "package main\n",
	})
	testCommit(t, gm, "2026-02-10T00:00:00Z", "update", map[string]string{
		"foo.go": "package main\n\n// v2\n",
	})

	// since is later than until → reversed commits
	_, err := trail(ctx, &trailOpts{
		Dir:   gm.RepoPath(),
		Since: "2026-02-15T00:00:00Z", // → endCommit of since = commit at 2026-02-10
		Until: "2026-01-15T00:00:00Z", // → endCommit of until = commit at 2026-01-10
	}, os.Stderr)
	if err == nil {
		t.Fatal("expected error for reversed commits, got nil")
	}
	if ec, ok := err.(interface{ ExitCode() int }); !ok || ec.ExitCode() != 1 {
		t.Errorf("expected exit code 1 for reversed commits, got err=%v", err)
	}
}

func TestTrailRename(t *testing.T) {
	gm := newTestRepo(t)
	ctx := context.Background()

	testCommit(t, gm, "2026-01-10T00:00:00Z", "initial", map[string]string{
		"old.go": "package main\n",
	})
	// modify old.go
	testCommit(t, gm, "2026-02-01T00:00:00Z", "modify", map[string]string{
		"old.go": "package main\n\n// updated\n",
	})
	// rename old.go → new.go
	testRenameCommit(t, gm, "2026-03-01T00:00:00Z", "old.go", "new.go", "rename")

	result, err := trail(ctx, &trailOpts{
		Dir:   gm.RepoPath(),
		Since: "2026-01-05T00:00:00Z",
		Until: "2026-04-01T00:00:00Z",
	}, os.Stderr)
	if err != nil {
		t.Fatalf("trail: %v", err)
	}

	want := []FileChange{
		{Status: Modified, Path: "new.go", OldPath: "old.go"},
	}
	if !reflect.DeepEqual(result.Changes, want) {
		t.Errorf("Changes = %v, want %v", result.Changes, want)
	}
}

func TestTrailRenameChain(t *testing.T) {
	gm := newTestRepo(t)
	ctx := context.Background()

	testCommit(t, gm, "2026-01-10T00:00:00Z", "initial", map[string]string{
		"a.go": "package main\n",
	})
	testRenameCommit(t, gm, "2026-02-01T00:00:00Z", "a.go", "b.go", "rename a→b")
	testRenameCommit(t, gm, "2026-03-01T00:00:00Z", "b.go", "c.go", "rename b→c")

	result, err := trail(ctx, &trailOpts{
		Dir:   gm.RepoPath(),
		Since: "2026-01-05T00:00:00Z",
		Until: "2026-04-01T00:00:00Z",
	}, os.Stderr)
	if err != nil {
		t.Fatalf("trail: %v", err)
	}

	want := []FileChange{
		{Status: Modified, Path: "c.go", OldPath: "a.go"},
	}
	if !reflect.DeepEqual(result.Changes, want) {
		t.Errorf("Changes = %v, want %v", result.Changes, want)
	}
}

func TestTrailRenameNewFileIgnored(t *testing.T) {
	gm := newTestRepo(t)
	ctx := context.Background()

	testCommit(t, gm, "2026-01-10T00:00:00Z", "initial", map[string]string{
		"existing.go": "package main\n",
	})
	// Add new.go (didn't exist at start)
	testCommit(t, gm, "2026-02-01T00:00:00Z", "add new.go", map[string]string{
		"new.go": "package main\n",
	})
	// Rename new.go → renamed.go
	testRenameCommit(t, gm, "2026-03-01T00:00:00Z", "new.go", "renamed.go", "rename new→renamed")

	result, err := trail(ctx, &trailOpts{
		Dir:   gm.RepoPath(),
		Since: "2026-01-05T00:00:00Z",
		Until: "2026-04-01T00:00:00Z",
	}, os.Stderr)
	if err != nil {
		t.Fatalf("trail: %v", err)
	}

	// renamed.go should be Added (not Modified), because new.go didn't exist at start
	want := []FileChange{
		{Status: Added, Path: "renamed.go"},
	}
	if !reflect.DeepEqual(result.Changes, want) {
		t.Errorf("Changes = %v, want %v", result.Changes, want)
	}
}

func TestTrailPathspec(t *testing.T) {
	gm := newTestRepo(t)
	ctx := context.Background()

	testCommit(t, gm, "2026-01-10T00:00:00Z", "initial", map[string]string{
		"src/foo.go":   "package main\n",
		"tests/bar.go": "package main\n",
	})
	testCommit(t, gm, "2026-02-10T00:00:00Z", "modify both", map[string]string{
		"src/foo.go":   "package main\n\n// modified\n",
		"tests/bar.go": "package main\n\n// modified\n",
	})

	result, err := trail(ctx, &trailOpts{
		Dir:       gm.RepoPath(),
		Since:     "2026-01-15T00:00:00Z",
		Until:     "2026-03-01T00:00:00Z",
		Pathspecs: []string{"src/"},
	}, os.Stderr)
	if err != nil {
		t.Fatalf("trail: %v", err)
	}

	want := []FileChange{
		{Status: Modified, Path: "src/foo.go"},
	}
	if !reflect.DeepEqual(result.Changes, want) {
		t.Errorf("Changes = %v, want %v", result.Changes, want)
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
	// We just check it returns an error gracefully rather than panicking.
	ctx := context.Background()
	var g *Gitrail
	since := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	// Should not panic; will error (no git repo in current directory or empty)
	// We just want to confirm no nil pointer dereference.
	_, _ = g.Trail(ctx, "", since, until)
}
