gitrail
=======

[![Test Status](https://github.com/Songmu/gitrail/actions/workflows/test.yaml/badge.svg?branch=main)][actions]
[![Coverage Status](https://codecov.io/gh/Songmu/gitrail/branch/main/graph/badge.svg)][codecov]
[![MIT License](https://img.shields.io/github/license/Songmu/gitrail)][license]
[![PkgGoDev](https://pkg.go.dev/badge/github.com/Songmu/gitrail)][PkgGoDev]

[actions]: https://github.com/Songmu/gitrail/actions?workflow=test
[codecov]: https://codecov.io/gh/Songmu/gitrail
[license]: https://github.com/Songmu/gitrail/blob/main/LICENSE
[PkgGoDev]: https://pkg.go.dev/github.com/Songmu/gitrail

gitrail tracks file changes over a given time period in a Git repository. It wraps git commands to show which files were added, modified, deleted, or renamed between two points in time, with rename chain detection.

## CLI Usage

```console
% gitrail --since="2026-01-01" --until="2026-03-01"
abc123..def456

A	src/new.go
D	src/removed.go
M	src/bar.go	src/old_bar.go
M	src/foo.go
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--since` | **required** | Start time |
| `--until` | **required** | End time |
| `-C` | current directory | Path to git repository |
| `--branch` | HEAD | Target branch or revision |
| `--json` | false | NDJSON output |
| `-- pathspec...` | none | Git pathspec filters (e.g. `'*.go'`, `':!vendor/'`) |

Time formats are passed directly to git — ISO 8601 (`2026-01-01`), relative dates (`"1 month ago"`), etc. are all supported.

### Pathspec Filtering

Use `--` to pass pathspec patterns and restrict the output to specific paths:

```console
# Only show changes under src/
% gitrail --since="2026-01-01" --until="2026-03-01" -- 'src/'

# Only .go files, excluding vendor/
% gitrail --since="2026-01-01" --until="2026-03-01" -- '*.go' ':!vendor/'
```

### NDJSON Output

With `--json`, each line is a self-contained JSON object:

```console
% gitrail --since="2026-01-01" --until="2026-03-01" --json
{"to":"def456","status":"Added","path":"src/new.go"}
{"from":"abc123","to":"def456","status":"Modified","path":"src/foo.go"}
{"from":"abc123","to":"def456","status":"Modified","path":"src/bar.go","old_path":"src/old_bar.go"}
{"from":"abc123","status":"Deleted","path":"src/removed.go"}
```

The JSON schema is available at [`schema/output.schema.json`](schema/output.schema.json).

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success (with or without changes) |
| 1 | Error (not a repo, reversed commits, start commit not found, etc.) |
| 2 | End commit not found (out of history range) |

## Agent Skills

gitrail ships an [agentskills](https://agentskills.io)-compatible skill definition so coding agents can discover and use it directly.

```console
# List embedded skills
% gitrail skills list

# Install skill(s) for the current user
% gitrail skills install

# Install into the current repository
% gitrail skills install --scope=repo
```

Run `gitrail skills --help` for all available subcommands.

## Library Usage

```go
g := gitrail.New("/path/to/repo")
result, err := g.Trail(ctx, "main",
    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
    time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
)
for _, c := range result.Changes {
    fmt.Printf("%s\t%s\n", c.Status, c.Path)
}
```

See [pkg.go.dev](https://pkg.go.dev/github.com/Songmu/gitrail) for full API documentation.

## Installation

```console
# Install with Homebrew (macOS/Linux)
% brew install Songmu/tap/gitrail

# Install the latest version. (Install it into ./bin/ by default).
% curl -sfL https://raw.githubusercontent.com/Songmu/gitrail/main/install.sh | sh -s

# Specify installation directory ($(go env GOPATH)/bin/) and version.
% curl -sfL https://raw.githubusercontent.com/Songmu/gitrail/main/install.sh | sh -s -- -b $(go env GOPATH)/bin [vX.Y.Z]

# In alpine linux (as it does not come with curl by default)
% wget -O - -q https://raw.githubusercontent.com/Songmu/gitrail/main/install.sh | sh -s [vX.Y.Z]

# go install
% go install github.com/Songmu/gitrail/cmd/gitrail@latest
```

## Author

[Songmu](https://github.com/Songmu)
