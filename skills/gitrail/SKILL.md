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

Always use `--json` for agent use to get structured, machine-readable output.

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

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "gitrail NDJSON output",
  "description": "Schema for each line of gitrail --json output. One JSON object per changed file.",
  "type": "object",
  "required": ["status", "path"],
  "properties": {
    "from": {
      "type": "string",
      "description": "Start commit hash. Present for Modified, Renamed, and Deleted entries."
    },
    "to": {
      "type": "string",
      "description": "End commit hash. Present for Added, Modified, and Renamed entries."
    },
    "status": {
      "type": "string",
      "enum": ["Added", "Modified", "Renamed", "Deleted"],
      "description": "Type of file change."
    },
    "path": {
      "type": "string",
      "description": "File path at the end commit. For Deleted entries, the path at the start commit."
    },
    "old_path": {
      "type": "string",
      "description": "Original file path before rename. Only present when the file was renamed."
    }
  },
  "allOf": [
    {
      "if": { "properties": { "status": { "const": "Added" } } },
      "then": { "required": ["to"], "properties": { "from": false, "old_path": false } }
    },
    {
      "if": { "properties": { "status": { "const": "Modified" } } },
      "then": { "required": ["from", "to"], "properties": { "old_path": false } }
    },
    {
      "if": { "properties": { "status": { "const": "Renamed" } } },
      "then": { "required": ["from", "to", "old_path"] }
    },
    {
      "if": { "properties": { "status": { "const": "Deleted" } } },
      "then": { "required": ["from"], "properties": { "to": false, "old_path": false } }
    }
  ],
  "additionalProperties": false
}
```

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

