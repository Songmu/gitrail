package gitrail

import (
	"os"
	"strings"
	"testing"

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
	out, _, err := gm.Do("rev-parse", "HEAD")
	if err != nil {
		t.Fatalf("git rev-parse HEAD: %v", err)
	}
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
	out, _, err := gm.Do("rev-parse", "HEAD")
	if err != nil {
		t.Fatalf("git rev-parse HEAD: %v", err)
	}
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
	out, _, err := gm.Do("rev-parse", "HEAD")
	if err != nil {
		t.Fatalf("git rev-parse HEAD: %v", err)
	}
	return strings.TrimSpace(out)
}
