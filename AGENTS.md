# gitall AGENTS file

_gitall_ is a command-line utility written in Go to recursively walk directories with git projects and perform combined action.

## Technologies

- go language
- Cobra for CLI structure
- Viper for configuration

## Repository Structure

- cmd/ - Main applications (entry points for the CLI and subcommands)
- internal/ - Private application code (business logic, services, helpers)
- api/ - API client definitions or interface contracts
- configs/ - Configuration templates and default files
- scripts/ - Build, install, and analysis scripts
- tests/ - Extra integration tests or test fixtures (unit tests reside with source files \_test.go)
- docs/ - Documentation

## Documentation

- docs/CODE.md Code structure. ALWAYS update this file when new packages are added, renamed, or removed; with short description of each package's function.
- docs/CLI.md Command usage and flags. ALWAYS update this file when commands or flags change.
- docs/CONFIG.md Configuration file rules. ALWAYS update this file when config keys or environment variable names change.

## Coding Guidelines

- Follow standard Go style (gofmt, goimports) and Effective Go guidelines.
- Handle errors explicitly; do not ignore errors (\_ = ... only if strictly necessary).
- Use internal/ for logic that should not be imported by other projects.
- Prefer returning errors over panic/log.Fatal in library code (within internal/ or pkg/).
- Use structured logging (e.g., slog or zerolog) rather than fmt.Printf for output.

## Development Commands

Install deps: go mod download
Tidy deps: go mod tidy
Run app: go run ./cmd/[project-name] [args]
Build binary: go build -o bin/[project-name] ./cmd/[project-name]
Run tests: go test ./...
Run tests (verbose): go test -v ./...
Run tests (coverage): go test -cover ./...
Lint: golangci-lint run (requires golangci-lint installed)
Format: go fmt ./...
Static Analysis: go vet ./...

## Tool usage

When using tools, consider:

- Don't cd to project directory before running a command, you are already there
- Host system is Windows, with Powershell

## Commit rules

- Analyze all staged and non-staged git changes: git diff, git diff --staged. Review the complete patch context.
- Check for:
    - Error handling: ensuring returned errors are handled.
    - Context propagation: passing context.Context correctly down the call stack.
    - Concurrency issues: race conditions or leaked goroutines.
    - Missing tests for new logic (keep unit tests in the same package).
    - go.mod / go.sum consistency (did you add a dependency? run go mod tidy).
- Report detailed information on all issues found (if any)
- Verify: - Tests OK (go test ./...) - Build OK (go build ./...) - Linter OK (golangci-lint run) - Vet OK (go vet ./...)
  Never commit: binary files in root, vendor/ directory (unless vendoring is used), IDE configuration files.
