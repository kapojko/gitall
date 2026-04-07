# gitall

gitall is a CLI utility that recursively walks directories to find git repositories and displays their combined status.

## Synopsis

```
gitall [path]
gitall --dir <path>
```

## Description

gitall searches the specified directory (or current directory if not specified) for git repositories. For each repository found, it displays the status using `git status --porcelain=v2 --branch` format and parses the output to determine overall repository state.

## Output Format

Each repository is displayed on a single line with tree-style indentation for submodules:

```
<path>                                    <status>
```

**Status values:**
- **OK** (green) - No changes, all commits pushed to remote
- **PUSH** (yellow) - No changes, but has non-pushed commits
- **CHANGES** (red) - Has active changes (modified, added, deleted files)

### Submodule Display

Submodules are shown with indentation under their parent repository:
```
/parent/repo                               OK
  └─ /parent/repo/submodule               OK
```

### Summary

After the list of repositories, a summary is displayed:
```
Summary:
  OK: 5
  PUSH: 2
  CHANGES: 1
```

## Options

- `-h, --help` - Display help information
- `--dir` - Starting directory (defaults to current directory)

## Exit Codes

- `0` - Command completed successfully
- `1` - Error occurred during execution

## Building

Build the binary:

```bash
go build -o bin/gitall ./cmd/gitall
```

On Windows:

```powershell
go build -o bin/gitall.exe ./cmd/gitall
```

## Testing

Run all tests:

```bash
go test ./...
```

Run tests with verbose output:

```bash
go test -v ./...
```

Run tests with coverage:

```bash
go test -cover ./...
```

## Linting

Run static analysis:

```bash
go vet ./...
```

Format code:

```bash
go fmt ./...
```

## Project Structure

```
gitall/
├── cmd/gitall/       # CLI entry point
├── internal/
│   ├── git/          # Git parsing and repo discovery
│   │   ├── parser.go  # Parses git status porcelain v2 output
│   │   └── walker.go # Recursively walks directories to find repos
│   ├── progress/     # Progress display
│   └── status/       # Status display formatting
├── docs/             # Documentation
│   ├── CODE.md       # Code structure
│   ├── CLI.md        # CLI usage
│   └── CONFIG.md     # Configuration
└── examples/         # Example files
```

## Dependencies

- Go 1.21+
- Cobra (CLI framework)
