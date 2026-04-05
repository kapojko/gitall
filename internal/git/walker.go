package git

import (
	"bufio"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ProgressCallback func(current, total int)

type Repository struct {
	Path   string
	Status *Status
	Parent *Repository
}

type Walker struct {
	root           string
	repos          []*Repository
	progressCb     ProgressCallback
	estimatedTotal int
}

func NewWalker(root string) *Walker {
	return &Walker{root: root}
}

func (w *Walker) SetProgressCallback(cb ProgressCallback) {
	w.progressCb = cb
}

func (w *Walker) Walk(ctx context.Context) ([]*Repository, error) {
	w.repos = make([]*Repository, 0)

	absRoot, err := filepath.Abs(w.root)
	if err != nil {
		return nil, err
	}

	w.estimateTopLevelDirs(absRoot)

	visited := make(map[string]bool)

	current := 0
	err = w.walkDir(ctx, absRoot, nil, visited, 0, &current)
	if err != nil {
		return nil, err
	}

	return w.repos, nil
}

func (w *Walker) estimateTopLevelDirs(root string) {
	entries, err := os.ReadDir(root)
	if err != nil {
		w.estimatedTotal = 10
		return
	}

	count := 0
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != ".git" {
			count++
		}
	}

	if count == 0 {
		count = 10
	}

	w.estimatedTotal = count
}

func (w *Walker) walkDir(ctx context.Context, dir string, parent *Repository, visited map[string]bool, depth int, current *int) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if depth == 1 {
		*current++
		if w.progressCb != nil && w.estimatedTotal > 0 {
			w.progressCb(*current, w.estimatedTotal)
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	hasGit := false
	var gitDir string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()

		if name == ".git" {
			hasGit = true
			gitDir = filepath.Join(dir, name)
			break
		}
	}

	if hasGit {
		realGitDir, err := filepath.EvalSymlinks(gitDir)
		if err == nil {
			gitDir = realGitDir
		}

		var repo *Repository
		if !visited[gitDir] {
			visited[gitDir] = true

			status, err := w.getRepoStatus(ctx, gitDir)
			if err != nil {
				return err
			}

			repo = &Repository{
				Path:   dir,
				Status: status,
				Parent: parent,
			}
			w.repos = append(w.repos, repo)
		}

		currentParent := parent
		if repo != nil {
			currentParent = repo
		}

		submodules, err := w.getSubmodules(ctx, dir)
		if err == nil && len(submodules) > 0 {
			for _, subPath := range submodules {
				subGitDir := filepath.Join(subPath, ".git")
				if visited[subGitDir] {
					continue
				}
				visited[subGitDir] = true

				subStatus, err := w.getRepoStatus(ctx, subGitDir)
				if err != nil {
					continue
				}
				subStatus.IsSubmodule = true

				w.repos = append(w.repos, &Repository{
					Path:   subPath,
					Status: subStatus,
					Parent: currentParent,
				})
			}
		}

		return nil
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		if name == ".git" {
			continue
		}

		subdir := filepath.Join(dir, name)

		err := w.walkDir(ctx, subdir, parent, visited, depth+1, current)
		if err != nil {
			if err == context.Canceled || err == context.DeadlineExceeded {
				return err
			}
		}
	}

	return nil
}

func (w *Walker) getSubmodules(ctx context.Context, repoPath string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", "submodule", "status")
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	submodules := make([]string, 0)
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		subPath := filepath.Join(repoPath, parts[1])
		submodules = append(submodules, subPath)
	}

	return submodules, nil
}

func (w *Walker) getRepoStatus(ctx context.Context, gitDir string) (*Status, error) {
	repoPath := filepath.Dir(gitDir)

	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain=v2", "--branch")
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return &Status{}, nil
	}

	status, err := ParsePorcelainV2(strings.NewReader(string(output)))
	if err != nil {
		return nil, err
	}

	return status, nil
}
