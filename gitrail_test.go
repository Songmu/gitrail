package gitrail

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
)

func TestRunSkillsList(t *testing.T) {
	var out, errOut bytes.Buffer
	err := Run(context.Background(), []string{"skills", "list"}, &out, &errOut)
	if err != nil {
		t.Fatalf("Run skills list: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "gitrail") {
		t.Errorf("skills list output %q should contain skill name 'gitrail'", got)
	}
	if !strings.Contains(got, "git") {
		t.Errorf("skills list output %q should contain skill description mentioning 'git'", got)
	}
}

func TestRunVersion(t *testing.T) {
	var out bytes.Buffer
	err := Run(context.Background(), []string{"--version"}, &out, os.Stderr)
	if err != nil {
		t.Fatalf("Run --version: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "gitrail") {
		t.Errorf("version output %q missing 'gitrail'", got)
	}
}

func TestRunMissingSince(t *testing.T) {
	var out, errOut bytes.Buffer
	err := Run(context.Background(), []string{"--until=2026-03-01"}, &out, &errOut)
	if err == nil {
		t.Fatal("expected error for missing --since")
	}
}

func TestRunMissingUntil(t *testing.T) {
	var out, errOut bytes.Buffer
	err := Run(context.Background(), []string{"--since=2026-01-01"}, &out, &errOut)
	if err == nil {
		t.Fatal("expected error for missing --until")
	}
}

func TestRunTextOutput(t *testing.T) {
	gm := newTestRepo(t)
	ctx := context.Background()

	testCommit(t, gm, "2026-01-10T00:00:00Z", "initial", map[string]string{
		"bar.go": "package main\n\nfunc bar() {}\n",
		"foo.go": "package main\n\nfunc foo() {}\n",
	})
	testCommit(t, gm, "2026-02-10T00:00:00Z", "changes", map[string]string{
		"baz.go": "package main\n\nfunc baz() {} // new file\n",
		"foo.go": "package main\n\nfunc foo() { /* modified */ }\n",
	})
	testDeleteCommit(t, gm, "2026-02-10T00:00:01Z", "bar.go", "delete bar")

	var out bytes.Buffer
	err := Run(ctx, []string{
		"-C", gm.RepoPath(),
		"--since=2026-01-15T00:00:00Z",
		"--until=2026-03-01T00:00:00Z",
	}, &out, os.Stderr)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	got := out.String()
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")

	// First line is commit range header
	if !strings.Contains(lines[0], "..") {
		t.Errorf("first line should be commit range, got %q", lines[0])
	}
	// Second line is blank
	if len(lines) < 2 || lines[1] != "" {
		t.Errorf("second line should be blank, got %q", lines[1])
	}
	// File lines
	fileLines := lines[2:]
	if len(fileLines) != 3 {
		t.Errorf("expected 3 file lines, got %d: %v", len(fileLines), fileLines)
	}
	// Files sorted alphabetically: bar.go (D), baz.go (A), foo.go (M)
	if !strings.HasPrefix(fileLines[0], "D\tbar.go") {
		t.Errorf("line 0: want D\\tbar.go, got %q", fileLines[0])
	}
	if !strings.HasPrefix(fileLines[1], "A\tbaz.go") {
		t.Errorf("line 1: want A\\tbaz.go, got %q", fileLines[1])
	}
	if !strings.HasPrefix(fileLines[2], "M\tfoo.go") {
		t.Errorf("line 2: want M\\tfoo.go, got %q", fileLines[2])
	}
}

func TestRunTextOutputNoChanges(t *testing.T) {
	gm := newTestRepo(t)
	ctx := context.Background()

	testCommit(t, gm, "2026-02-01T00:00:00Z", "only commit", map[string]string{
		"foo.go": "package main\n",
	})

	var out bytes.Buffer
	err := Run(ctx, []string{
		"-C", gm.RepoPath(),
		"--since=2026-01-15T00:00:00Z",
		"--until=2026-02-15T00:00:00Z",
	}, &out, os.Stderr)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	got := out.String()
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	// Should only have the commit range header, no blank line or file lines
	if len(lines) != 1 {
		t.Errorf("expected 1 line (no changes), got %d lines: %v", len(lines), lines)
	}
	if !strings.Contains(lines[0], "..") {
		t.Errorf("first line should be commit range, got %q", lines[0])
	}
}

func TestRunJSONOutput(t *testing.T) {
	gm := newTestRepo(t)
	ctx := context.Background()

	testCommit(t, gm, "2026-01-10T00:00:00Z", "initial", map[string]string{
		"bar.go": "package main\n\nfunc bar() {}\n",
		"foo.go": "package main\n\nfunc foo() {}\n",
	})
	testCommit(t, gm, "2026-02-10T00:00:00Z", "changes", map[string]string{
		"baz.go": "package main\n\nfunc baz() {} // new file\n",
		"foo.go": "package main\n\nfunc foo() { /* modified */ }\n",
	})
	testDeleteCommit(t, gm, "2026-02-10T00:00:01Z", "bar.go", "delete bar")

	var out bytes.Buffer
	err := Run(ctx, []string{
		"-C", gm.RepoPath(),
		"--since=2026-01-15T00:00:00Z",
		"--until=2026-03-01T00:00:00Z",
		"--json",
	}, &out, os.Stderr)
	if err != nil {
		t.Fatalf("Run --json: %v", err)
	}

	got := out.String()
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 JSON lines, got %d: %v", len(lines), lines)
	}
	// bar.go: Deleted → has from, no to
	if !strings.Contains(lines[0], `"status":"Deleted"`) {
		t.Errorf("line 0 should be Deleted, got %q", lines[0])
	}
	if !strings.Contains(lines[0], `"from"`) {
		t.Errorf("line 0 (Deleted) should have from, got %q", lines[0])
	}
	if strings.Contains(lines[0], `"to"`) {
		t.Errorf("line 0 (Deleted) should not have to, got %q", lines[0])
	}
	// baz.go: Added → no from, has to
	if !strings.Contains(lines[1], `"status":"Added"`) {
		t.Errorf("line 1 should be Added, got %q", lines[1])
	}
	if strings.Contains(lines[1], `"from"`) {
		t.Errorf("line 1 (Added) should not have from, got %q", lines[1])
	}
	if !strings.Contains(lines[1], `"to"`) {
		t.Errorf("line 1 (Added) should have to, got %q", lines[1])
	}
	// foo.go: Modified → has both from and to
	if !strings.Contains(lines[2], `"status":"Modified"`) {
		t.Errorf("line 2 should be Modified, got %q", lines[2])
	}
	if !strings.Contains(lines[2], `"from"`) {
		t.Errorf("line 2 (Modified) should have from, got %q", lines[2])
	}
	if !strings.Contains(lines[2], `"to"`) {
		t.Errorf("line 2 (Modified) should have to, got %q", lines[2])
	}
}

func TestRunJSONOutputWithRename(t *testing.T) {
	gm := newTestRepo(t)
	ctx := context.Background()

	testCommit(t, gm, "2026-01-10T00:00:00Z", "initial", map[string]string{
		"old.go": "package main\n",
	})
	testRenameCommit(t, gm, "2026-02-01T00:00:00Z", "old.go", "new.go", "rename")

	var out bytes.Buffer
	err := Run(ctx, []string{
		"-C", gm.RepoPath(),
		"--since=2026-01-05T00:00:00Z",
		"--until=2026-03-01T00:00:00Z",
		"--json",
	}, &out, os.Stderr)
	if err != nil {
		t.Fatalf("Run --json: %v", err)
	}

	got := out.String()
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 1 {
		t.Errorf("expected 1 JSON line, got %d: %v", len(lines), lines)
	}
	if !strings.Contains(lines[0], `"status":"Modified"`) {
		t.Errorf("expected Modified, got %q", lines[0])
	}
	if !strings.Contains(lines[0], `"old_path":"old.go"`) {
		t.Errorf("expected old_path, got %q", lines[0])
	}
	if !strings.Contains(lines[0], `"path":"new.go"`) {
		t.Errorf("expected path new.go, got %q", lines[0])
	}
}

func TestRunJSONOutputNoChanges(t *testing.T) {
	gm := newTestRepo(t)
	ctx := context.Background()

	testCommit(t, gm, "2026-02-01T00:00:00Z", "only commit", map[string]string{
		"foo.go": "package main\n",
	})

	var out bytes.Buffer
	err := Run(ctx, []string{
		"-C", gm.RepoPath(),
		"--since=2026-01-15T00:00:00Z",
		"--until=2026-02-15T00:00:00Z",
		"--json",
	}, &out, os.Stderr)
	if err != nil {
		t.Fatalf("Run --json: %v", err)
	}

	// No changes → no NDJSON output
	if out.Len() != 0 {
		t.Errorf("expected empty output for no changes, got %q", out.String())
	}
}

func TestRunExitCode2(t *testing.T) {
	gm := newTestRepo(t)
	ctx := context.Background()

	testCommit(t, gm, "2026-02-01T00:00:00Z", "first commit", map[string]string{
		"foo.go": "package main\n",
	})

	var out, errOut bytes.Buffer
	err := Run(ctx, []string{
		"-C", gm.RepoPath(),
		"--since=2025-01-01T00:00:00Z",
		"--until=2025-12-31T00:00:00Z",
	}, &out, &errOut)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if ec, ok := err.(interface{ ExitCode() int }); !ok || ec.ExitCode() != 2 {
		t.Errorf("expected exit code 2, got %v", err)
	}
}

func TestRunPathspec(t *testing.T) {
	gm := newTestRepo(t)
	ctx := context.Background()

	testCommit(t, gm, "2026-01-10T00:00:00Z", "initial", map[string]string{
		"src/main.go":  "package main\n",
		"docs/doc.txt": "hello\n",
	})
	testCommit(t, gm, "2026-02-10T00:00:00Z", "modify both", map[string]string{
		"src/main.go":  "package main\n\n// modified\n",
		"docs/doc.txt": "hello world\n",
	})

	var out bytes.Buffer
	err := Run(ctx, []string{
		"-C", gm.RepoPath(),
		"--since=2026-01-15T00:00:00Z",
		"--until=2026-03-01T00:00:00Z",
		"--",
		"src/",
	}, &out, os.Stderr)
	if err != nil {
		t.Fatalf("Run with pathspec: %v", err)
	}

	got := out.String()
	if strings.Contains(got, "docs/") {
		t.Errorf("docs/ should be filtered out, got %q", got)
	}
	if !strings.Contains(got, "src/main.go") {
		t.Errorf("src/main.go should be in output, got %q", got)
	}
}
