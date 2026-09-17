---
name: gitrail
description: >
  Use this skill whenever the user asks what files changed during a historical
  time window in a Git repository. Choose gitrail for date-based or
  relative-time investigations, release/compliance audits, and
  repository-structure change reports that need a reproducible file inventory
  rather than a commit list or patch. It identifies Added, Modified, Deleted,
  and Renamed files, including edited renames and multi-step rename histories
  with original and final paths. Use it when results must be scoped to a branch
  or revision, filtered by directories, path patterns, extensions, or
  exclusions such as vendor/generated files, or emitted as counts or
  machine-readable JSON/NDJSON. Also recognize equivalent non-English requests
  for files added, updated, deleted, or renamed over a period. Do not use for
  author/commit-message summaries, changelogs, line-level diffs, comparing
  current branch contents, schema validation, or filesystem watching.
license: MIT
allowed-tools:
  - Bash
  - Read
---

## Overview

`gitrail` shows which files were added, modified, deleted, or renamed between two points in time in a Git repository. It detects rename chains so a file renamed multiple times appears under its final name with the original path in `old_path`.

Always use `--json` for agent use to get structured, machine-readable output.

## Availability and Installation

Before invoking `gitrail`, check whether it is available with
`command -v gitrail`.

If the command is unavailable, do not install it automatically. Explain that
`gitrail` is required and ask the user to install it or approve installation.
Prefer the fully qualified Homebrew formula on macOS or Linux:

```console
brew install Songmu/tap/gitrail
```

This command trusts only the requested formula from the third-party tap; a
separate `brew trust` step is not needed for installation or routine updates.

If Homebrew is unavailable but Go is installed, use:

```console
go install github.com/Songmu/gitrail/cmd/gitrail@latest
```

## CLI Usage

```
gitrail --json --since=<time> --until=<time> [options] [-- pathspec...]
```

### Flags

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--since` | yes | — | Start of the time range |
| `--until` | yes | — | End of the time range |
| `--json` | no | false | Emit NDJSON output (one JSON object per changed file; recommended for agents) |
| `-C` | no | current dir | Path to the git repository |
| `--branch` | no | HEAD | Branch or revision to inspect |

Time values are passed directly to git: ISO 8601 (`2026-01-01`), relative strings (`"1 month ago"`), RFC 2822, etc.

Use `--` to add git pathspec filters:

```bash
gitrail --json --since="2026-01-01" --until="2026-03-01" -- 'src/' ':!vendor/'
```

## Output Format (`--json`)

Each line of output is a JSON object describing one changed file:

```jsonl
{"to":"def456","status":"Added","path":"src/new.go"}
{"from":"abc123","to":"def456","status":"Modified","path":"src/foo.go"}
{"from":"abc123","to":"def456","status":"Modified","path":"src/bar.go","old_path":"src/old_bar.go"}
{"from":"abc123","to":"def456","status":"Renamed","path":"src/baz.go","old_path":"src/old_baz.go"}
{"from":"abc123","status":"Deleted","path":"src/removed.go"}
```

### JSON Schema

The schema for each NDJSON line is bundled at
[`assets/output.schema.json`](assets/output.schema.json). Read that file when
validating or interpreting `--json` output.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error (not a git repo, reversed commits, start commit not found, etc.) |
| 2 | End commit not found (time range is out of repository history) |

## Common Agent Use Cases

**Find all files that changed in the last month:**
```bash
gitrail --json --since="1 month ago" --until="now"
```

**Inspect changes on a feature branch:**
```bash
gitrail --json --since="2026-01-01" --until="2026-03-01" --branch=feature/my-branch
```

**Limit to Go sources, excluding generated code:**
```bash
gitrail --json --since="2026-01-01" --until="2026-03-01" -- '*.go' ':!*_gen.go'
```

**Filter only added files:**
```bash
gitrail --json --since="2026-01-01" --until="2026-03-01" | jq 'select(.status=="Added")'
```
