package status

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"gitall/internal/git"
	"gitall/internal/progress"
)

type RepoResult struct {
	Path        string
	Status      *git.Status
	Depth       int
	IsSubmodule bool
}

type Command struct {
	Writer io.Writer
}

func NewCommand(w io.Writer) *Command {
	return &Command{Writer: w}
}

func (c *Command) Execute(ctx context.Context, root string) error {
	walker := git.NewWalker(root)

	prog := progress.New(c.Writer)
	prog.SetPrefix("Walking directories")

	walker.SetProgressCallback(func(current, total int) {
		prog.SetCurrent(current, total)
	})

	repos, err := walker.Walk(ctx)
	prog.Finish()

	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("failed to walk directory tree: %w", err)
	}

	results := c.buildResults(repos)

	c.printResults(results)

	ok, push, changes := c.countStatuses(results)

	c.printSummary(ok, push, changes)

	return nil
}

func (c *Command) buildResults(repos []*git.Repository) []RepoResult {
	results := make([]RepoResult, 0, len(repos))

	for _, repo := range repos {
		depth := 0
		isSub := false

		p := repo.Parent
		for p != nil {
			depth++
			isSub = true
			p = p.Parent
		}

		results = append(results, RepoResult{
			Path:        repo.Path,
			Status:      repo.Status,
			Depth:       depth,
			IsSubmodule: isSub,
		})
	}

	return results
}

func (c *Command) printResults(results []RepoResult) {
	if len(results) == 0 {
		fmt.Fprintln(c.Writer, "No git repositories found.")
		return
	}

	maxLineLen := 0
	for i, r := range results {
		var sb strings.Builder
		if r.Depth > 0 {
			for d := 0; d < r.Depth-1; d++ {
				sb.WriteString("│   ")
			}
			isLast := true
			for j := i + 1; j < len(results); j++ {
				if results[j].Depth == r.Depth {
					isLast = false
					break
				}
				if results[j].Depth < r.Depth {
					break
				}
			}
			if isLast {
				sb.WriteString("└── ")
			} else {
				sb.WriteString("├── ")
			}
		}
		sb.WriteString(r.Path)
		lineLen := utf8.RuneCountInString(sb.String())
		if lineLen > maxLineLen {
			maxLineLen = lineLen
		}
	}

	alignColumn := maxLineLen + 2

	fmt.Fprintln(c.Writer)

	for i, r := range results {
		var line strings.Builder

		if r.Depth > 0 {
			isLast := true
			for j := i + 1; j < len(results); j++ {
				if results[j].Depth == r.Depth {
					isLast = false
					break
				}
				if results[j].Depth < r.Depth {
					break
				}
			}

			for d := 0; d < r.Depth-1; d++ {
				line.WriteString("│   ")
			}

			if isLast {
				line.WriteString("└── ")
			} else {
				line.WriteString("├── ")
			}
		}

		line.WriteString(r.Path)

		statusStr := c.formatStatus(r.Status)

		padding := alignColumn - utf8.RuneCountInString(line.String())
		if padding < 1 {
			padding = 1
		}

		fmt.Fprintf(c.Writer, "%s%*s%s\n", line.String(), padding, "", statusStr)
	}
}

func (c *Command) formatStatus(status *git.Status) string {
	repoStatus := status.RepoStatus()

	switch repoStatus {
	case git.StatusChanges:
		return colorize("CHANGES", Red)
	case git.StatusPush:
		return colorize("PUSH", Yellow)
	default:
		return colorize("OK", Green)
	}
}

func (c *Command) countStatuses(results []RepoResult) (ok, push, changes int) {
	for _, r := range results {
		switch r.Status.RepoStatus() {
		case git.StatusOK:
			ok++
		case git.StatusPush:
			push++
		case git.StatusChanges:
			changes++
		}
	}
	return
}

func (c *Command) printSummary(ok, push, changes int) {
	fmt.Fprintln(c.Writer)
	fmt.Fprintln(c.Writer, "Summary:")

	if ok > 0 {
		fmt.Fprintf(c.Writer, "  %s: %d\n", colorize("OK", Green), ok)
	}
	if push > 0 {
		fmt.Fprintf(c.Writer, "  %s: %d\n", colorize("PUSH", Yellow), push)
	}
	if changes > 0 {
		fmt.Fprintf(c.Writer, "  %s: %d\n", colorize("CHANGES", Red), changes)
	}

	if ok == 0 && push == 0 && changes == 0 {
		fmt.Fprintln(c.Writer, "  No repositories found.")
	}
}

const (
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Red    = "\033[31m"
	Reset  = "\033[0m"
)

func colorize(text, color string) string {
	return color + text + Reset
}

func RunStatus(root string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cmd := NewCommand(os.Stdout)
	return cmd.Execute(ctx, root)
}
