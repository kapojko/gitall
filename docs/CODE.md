# Code Structure

## Overview

gitall is a Go CLI utility that recursively walks directories to find git repositories and display their combined status.

## Package Structure

### `cmd/gitall`
Entry point for the CLI application. Contains:
- `main.go` - Application entry point, initializes Cobra root command

### `internal/git`
Git parsing and repository discovery logic. Contains:
- `parser.go` - Parses `git status --porcelain=v2 --branch` output
- `walker.go` - Recursively walks directory tree to find git repositories

### `internal/status`
Status display command implementation. Contains:
- `status.go` - Formats and displays git repository status results

### `internal/progress`
Progress display functionality. Contains:
- `progress.go` - Thread-safe progress bar display during directory walking

## Key Interfaces

### `git.Status`
Represents the status of a single git repository:
- `Branch` - Branch information (commit, head, upstream, ahead/behind)
- `Files` - List of changed files
- `HasChanges` - Whether repository has active changes
- `HasUnpushed` - Whether repository has commits not pushed to remote
- `RepoStatus()` - Returns overall status: OK, PUSH, or CHANGES

### `git.FileEntry`
Represents a single changed file:
- `Path` - File path
- `Change` - Type of change (Modified, Added, Deleted, etc.)
- `Staged/Unstaged` - Whether change is staged
- `IsSubmodule` - Whether entry is a submodule

### `git.BranchInfo`
Branch information:
- `Commit` - Current commit hash
- `Head` - Current branch name
- `Upstream` - Upstream branch name
- `Ahead/Behind` - Number of commits ahead/behind upstream

### `git.Repository`
Represents a discovered repository:
- `Path` - Repository root path
- `Status` - Parsed git status
- `Parent` - Parent repository (for submodules)

### `git.Walker`
Walks directory tree to discover repositories:
- `Walk(ctx)` - Returns list of discovered repositories

### `status.RepoResult`
Formatted result for display:
- `Path` - Repository path
- `Status` - Git status
- `Depth` - Nesting depth (for submodules)
- `IsSubmodule` - Whether this is a submodule
