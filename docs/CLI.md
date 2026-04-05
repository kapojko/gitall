# CLI Documentation

## gitall

`gitall` is a CLI utility to recursively walk directories with git projects and perform combined actions.

## Synopsis

```
gitall [path]
gitall --dir <path>
```

## Commands

gitall is a single command utility. It recursively walks all subdirectories and displays combined git status for all repositories found.

**Usage:**
```
gitall [path]
```

**Flags:**
- `-h, --help` - Display help
- `--dir` - Starting directory (defaults to current directory)

## Examples

### Display status of current directory
```bash
gitall
```

### Display status of specific directory
```bash
gitall /path/to/directory
gitall --dir /path/to/directory
```

## Output Format

The status command displays each repository on a single line with the following format:

```
<path>                                    <status>
```

- **OK** (green) - No changes, all commits pushed to remote
- **PUSH** (yellow) - No changes, but has non-pushed commits
- **CHANGES** (red) - Has active changes (modified, added, deleted files)

### Submodules

Submodules are displayed with indentation under their parent repository:
```
/parent/repo                               OK
  └─ /parent/repo/submodule               OK
```

## Summary

After the list of repositories, a summary is displayed:
```
Summary:
  OK: 5
  PUSH: 2
  CHANGES: 1
```

## Exit Codes

- `0` - Command completed successfully
- `1` - Error occurred during execution
