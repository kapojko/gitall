# Configuration

gitall currently does not require any configuration files. All settings are passed via command-line arguments.

## Environment Variables

No environment variables are currently used.

## Future Configuration

Future versions may support configuration via:
- `~/.gitall.yaml` - User-level configuration
- `./.gitall.yaml` - Project-level configuration
- Environment variables with prefix `GITALL_`

## Git Configuration

gitall relies on the git configuration of each repository it visits. It does not read or modify git config.

## Terminal Colors

gitall uses ANSI color codes for status display:
- Green (`\033[32m`) - OK status
- Yellow (`\033[33m`) - PUSH status
- Red (`\033[31m`) - CHANGES status
- Reset (`\033[0m`) - Reset color

Colors are displayed when the terminal supports them.
