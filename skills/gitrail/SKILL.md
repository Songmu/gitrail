---
name: gitrail
description: >
  CLI that tracks file changes (added, modified, deleted, renamed)
  over a time period in a Git repository, with rename chain detection.
  Keywords: git, diff, file changes, rename tracking.
license: MIT
compatibility:
  - claude
  - codex
  - agents
allowed_tools:
  - Bash
  - Read
---

## Overview

`gitrail` shows which files were added, modified, deleted, or renamed between two points in time in a Git repository. It detects rename chains so a file renamed multiple times appears under its final name with the original path in `old_path`.

## CLI Usage

```
gitrail --since=<time> --until=<time> [options] [-- pathspec...]
```

### Flags

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--since` | yes | — | Start of the time range |
| `--until` | yes | — | End of the time range |
| `-C` | no | current dir | Path to the git repository |
| `--branch` | no | HEAD | Branch or revision to inspect |
| `--json` | no | false | Emit NDJSON output instead of text |

Time values are passed directly to git: ISO 8601 (`2026-01-01`), relative strings (`"1 month ago"`), RFC 2822, etc.

Use `--` to add git pathspec filters:

```bash
gitrail --since="2026-01-01" --until="2026-03-01" -- 'src/' ':!vendor/'
```

### Text Output Format

The output begins with `<from-commit>..<to-commit>` on the first line, followed by a blank line, then tab-separated records:

```
abc123..def456

A	src/new.go
D	src/removed.go
M	src/bar.go	src/old_bar.go
M	src/foo.go
```

- `A` — Added (file appeared in the period)
- `D` — Deleted (file removed in the period)
- `M <path>` — Modified in place
- `M <new-path> <old-path>` — Renamed (and optionally modified); `old_path` is the original name

### NDJSON Output (`--json`)

One JSON object per line:

```jsonl
{"to":"def456","status":"Added","path":"src/new.go"}
{"from":"abc123","to":"def456","status":"Modified","path":"src/foo.go"}
{"from":"abc123","to":"def456","status":"Modified","path":"src/bar.go","old_path":"src/old_bar.go"}
{"from":"abc123","status":"Deleted","path":"src/removed.go"}
```

Schema: `schema/output.schema.json` in the repository root.

| Field | Type | When present |
|-------|------|--------------|
| `status` | `"Added"` \| `"Modified"` \| `"Deleted"` | always |
| `path` | string | always |
| `old_path` | string | Modified with rename only |
| `from` | string (commit SHA) | Modified, Deleted |
| `to` | string (commit SHA) | Added, Modified |

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error (not a git repo, reversed commits, start commit not found, etc.) |
| 2 | End commit not found (time range is out of repository history) |

## Common Agent Use Cases

**Find all files that changed in the last month:**
```bash
gitrail --since="1 month ago" --until="now"
```

**Inspect changes on a feature branch:**
```bash
gitrail --since="2026-01-01" --until="2026-03-01" --branch=feature/my-branch
```

**Limit to Go sources, excluding generated code:**
```bash
gitrail --since="2026-01-01" --until="2026-03-01" -- '*.go' ':!*_gen.go'
```

**Machine-readable output for scripting:**
```bash
gitrail --since="2026-01-01" --until="2026-03-01" --json | jq 'select(.status=="Added")'
```

