package gitrail

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"
)

// ChangeStatus represents the type of file change.
type ChangeStatus string

const (
	// Added means the file was added.
	Added ChangeStatus = "Added"
	// Modified means the file was modified (or renamed).
	Modified ChangeStatus = "Modified"
	// Deleted means the file was deleted.
	Deleted ChangeStatus = "Deleted"
)

// FileChange represents a single file change.
type FileChange struct {
	Status  ChangeStatus
	Path    string // path at end commit (or start commit for Deleted)
	OldPath string // rename source path (only set when renamed)
}

// Result is the output from the Trail method.
type Result struct {
	StartCommit string
	EndCommit   string
	Changes     []FileChange
}

// Gitrail tracks file changes over a given time period in a Git repository.
type Gitrail struct {
	Dir       string    // repo path, defaults to current directory
	ErrStream io.Writer // where git error output is written; defaults to os.Stderr
}

// New creates a new Gitrail with the given repository directory.
func New(dir string) *Gitrail {
	return &Gitrail{Dir: dir}
}

func (g *Gitrail) errStream() io.Writer {
	if g != nil && g.ErrStream != nil {
		return g.ErrStream
	}
	return os.Stderr
}

// Trail returns file changes between since and until on the given branch.
// pathspecs optionally restricts which paths are considered.
// An empty branch defaults to HEAD.
func (g *Gitrail) Trail(ctx context.Context, branch string, since, until time.Time, pathspecs ...string) (*Result, error) {
	dir := ""
	if g != nil {
		dir = g.Dir
	}
	return trail(ctx, &trailOpts{
		Dir:       dir,
		Since:     since.Format(time.RFC3339),
		Until:     until.Format(time.RFC3339),
		Branch:    branch,
		Pathspecs: pathspecs,
	}, g.errStream())
}

type trailOpts struct {
	Dir       string
	Since     string
	Until     string
	Branch    string
	Pathspecs []string
}

func trail(ctx context.Context, opts *trailOpts, errStream io.Writer) (*Result, error) {
	startCommit, err := findStartCommit(ctx, opts.Dir, opts.Since, opts.Branch, errStream)
	if err != nil {
		return nil, err
	}

	endCommit, err := findEndCommit(ctx, opts.Dir, opts.Until, opts.Branch, errStream)
	if err != nil {
		return nil, err
	}

	if startCommit == endCommit {
		return &Result{
			StartCommit: startCommit,
			EndCommit:   endCommit,
		}, nil
	}

	// Validate that start is an ancestor of end.
	if err := validateAncestor(ctx, opts.Dir, startCommit, endCommit, errStream); err != nil {
		return nil, err
	}

	changes, err := getDiff(ctx, opts.Dir, startCommit, endCommit, opts.Pathspecs, errStream)
	if err != nil {
		return nil, err
	}

	if len(changes) > 0 {
		changes, err = applyRenameDetection(ctx, opts.Dir, startCommit, endCommit, opts.Pathspecs, changes, errStream)
		if err != nil {
			return nil, err
		}
	}

	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Path < changes[j].Path
	})

	return &Result{
		StartCommit: startCommit,
		EndCommit:   endCommit,
		Changes:     changes,
	}, nil
}

func findStartCommit(ctx context.Context, dir, since, branch string, errStream io.Writer) (string, error) {
	args := []string{"log", "--first-parent", fmt.Sprintf("--before=%s", since), "-1", "--format=%H"}
	if branch != "" {
		args = append(args, branch)
	}
	commit, err := gitCmd(ctx, dir, errStream, args...)
	if err != nil {
		return "", err
	}
	if commit != "" {
		return commit, nil
	}

	// Fallback: find the first commit after since (use --reverse to get oldest-first, take first).
	args = []string{"log", "--first-parent", fmt.Sprintf("--after=%s", since), "--reverse", "--format=%H"}
	if branch != "" {
		args = append(args, branch)
	}
	output, err := gitCmd(ctx, dir, errStream, args...)
	if err != nil {
		return "", err
	}
	if output == "" {
		return "", fmt.Errorf("start commit not found")
	}
	// Take the first line (oldest commit after since).
	commit = strings.SplitN(output, "\n", 2)[0]
	return commit, nil
}

func findEndCommit(ctx context.Context, dir, until, branch string, errStream io.Writer) (string, error) {
	args := []string{"log", "--first-parent", fmt.Sprintf("--before=%s", until), "-1", "--format=%H"}
	if branch != "" {
		args = append(args, branch)
	}
	commit, err := gitCmd(ctx, dir, errStream, args...)
	if err != nil {
		return "", err
	}
	if commit == "" {
		return "", newExitError(2, "end commit not found")
	}
	return commit, nil
}

func validateAncestor(ctx context.Context, dir, startCommit, endCommit string, errStream io.Writer) error {
	_, exitCode, err := gitCmdAllowFail(ctx, dir, errStream, "merge-base", "--is-ancestor", startCommit, endCommit)
	if err != nil {
		return fmt.Errorf("failed to check ancestry: %w", err)
	}
	if exitCode == 1 {
		return fmt.Errorf("start commit is not an ancestor of end commit (commits may be in reversed order)")
	}
	if exitCode != 0 {
		return newExitError(exitCode, fmt.Sprintf("git merge-base --is-ancestor failed with exit code %d", exitCode))
	}
	return nil
}

func getDiff(ctx context.Context, dir, startCommit, endCommit string, pathspecs []string, errStream io.Writer) ([]FileChange, error) {
	args := []string{"diff", "--name-status", startCommit, endCommit}
	if len(pathspecs) > 0 {
		args = append(args, "--")
		args = append(args, pathspecs...)
	}
	output, err := gitCmd(ctx, dir, errStream, args...)
	if err != nil {
		return nil, err
	}
	return parseNameStatus(output), nil
}

func parseNameStatus(output string) []FileChange {
	if output == "" {
		return nil
	}
	var changes []FileChange
	for _, line := range strings.Split(output, "\n") {
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) == 0 {
			continue
		}
		status := fields[0]
		switch {
		case status == "A" && len(fields) >= 2:
			changes = append(changes, FileChange{Status: Added, Path: fields[1]})
		case status == "M" && len(fields) >= 2:
			changes = append(changes, FileChange{Status: Modified, Path: fields[1]})
		case status == "D" && len(fields) >= 2:
			changes = append(changes, FileChange{Status: Deleted, Path: fields[1]})
		case strings.HasPrefix(status, "R") && len(fields) >= 3:
			changes = append(changes, FileChange{Status: Modified, Path: fields[2], OldPath: fields[1]})
		case strings.HasPrefix(status, "C") && len(fields) >= 3:
			changes = append(changes, FileChange{Status: Added, Path: fields[2]})
		case status == "T" && len(fields) >= 2:
			changes = append(changes, FileChange{Status: Modified, Path: fields[1]})
		}
	}
	return changes
}

// renameEvent is an old→new rename pair from a single commit.
type renameEvent struct {
	old string
	new string
}

func parseRenameEvents(output string) [][]renameEvent {
	if output == "" {
		return nil
	}
	var commits [][]renameEvent
	var current []renameEvent
	inCommit := false

	for _, line := range strings.Split(output, "\n") {
		if line == "" {
			continue
		}
		// Check if line looks like a commit hash (40- or 64-char hex).
		if isCommitHash(line) {
			if inCommit && len(current) > 0 {
				commits = append(commits, current)
			}
			current = nil
			inCommit = true
			continue
		}
		if inCommit && strings.HasPrefix(line, "R") {
			fields := strings.Split(line, "\t")
			if len(fields) >= 3 {
				current = append(current, renameEvent{old: fields[1], new: fields[2]})
			}
		}
	}
	if inCommit && len(current) > 0 {
		commits = append(commits, current)
	}
	return commits
}

func isCommitHash(s string) bool {
	// Accept SHA-1 (40 chars) or SHA-256 (64 chars) commit hashes.
	if len(s) != 40 && len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func applyRenameDetection(ctx context.Context, dir, startCommit, endCommit string, pathspecs []string, changes []FileChange, errStream io.Writer) ([]FileChange, error) {
	args := []string{"log", "--first-parent", "-M", "--diff-filter=R", "--name-status", "--format=%H", startCommit + ".." + endCommit}
	if len(pathspecs) > 0 {
		args = append(args, "--")
		args = append(args, pathspecs...)
	}
	output, err := gitCmd(ctx, dir, errStream, args...)
	if err != nil {
		return nil, err
	}
	if output == "" {
		return changes, nil
	}

	// Parse rename events (commits in newest-first order from git log).
	renamesByCommit := parseRenameEvents(output)
	if len(renamesByCommit) == 0 {
		return changes, nil
	}

	// Build originMap: current_path -> origin_path
	// Process commits from oldest to newest (reverse of git log output).
	originMap := make(map[string]string)
	for i := len(renamesByCommit) - 1; i >= 0; i-- {
		for _, r := range renamesByCommit[i] {
			origin := r.old
			if o, ok := originMap[r.old]; ok {
				origin = o
				delete(originMap, r.old)
			}
			originMap[r.new] = origin
		}
	}

	// Build sets for quick lookup.
	deletedSet := make(map[string]bool)
	addedSet := make(map[string]bool)
	for _, c := range changes {
		switch c.Status {
		case Deleted:
			deletedSet[c.Path] = true
		case Added:
			addedSet[c.Path] = true
		}
	}

	// Find rename mappings: added_path -> original_deleted_path.
	// Only apply if the origin existed at start (i.e., is in deletedSet).
	renameMapping := make(map[string]string)
	for path, origin := range originMap {
		if addedSet[path] && deletedSet[origin] {
			renameMapping[path] = origin
		}
	}

	if len(renameMapping) == 0 {
		return changes, nil
	}

	// Build set of deleted paths that are being renamed.
	renamedAwayPaths := make(map[string]bool)
	for _, origin := range renameMapping {
		renamedAwayPaths[origin] = true
	}

	// Rebuild changes with renames applied.
	result := make([]FileChange, 0, len(changes))
	for _, c := range changes {
		if c.Status == Deleted && renamedAwayPaths[c.Path] {
			// This deleted file was renamed to something else; skip.
			continue
		}
		if c.Status == Added {
			if origin, ok := renameMapping[c.Path]; ok {
				result = append(result, FileChange{
					Status:  Modified,
					Path:    c.Path,
					OldPath: origin,
				})
				continue
			}
		}
		result = append(result, c)
	}
	return result, nil
}
