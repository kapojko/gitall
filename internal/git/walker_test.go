package git

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestWalker_DiscoverSingleRepo(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}

	walker := NewWalker(tmpDir)
	repos, err := walker.Walk(context.Background())
	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	if len(repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(repos))
	}
	if repos[0].Path != tmpDir {
		t.Errorf("expected path %s, got %s", tmpDir, repos[0].Path)
	}
}

func TestWalker_DiscoverSiblingRepos(t *testing.T) {
	tmpDir := t.TempDir()

	repo1 := filepath.Join(tmpDir, "repo1")
	gitDir1 := filepath.Join(repo1, ".git")
	if err := os.MkdirAll(gitDir1, 0755); err != nil {
		t.Fatalf("failed to create repo1 .git: %v", err)
	}

	repo2 := filepath.Join(tmpDir, "repo2")
	gitDir2 := filepath.Join(repo2, ".git")
	if err := os.MkdirAll(gitDir2, 0755); err != nil {
		t.Fatalf("failed to create repo2 .git: %v", err)
	}

	walker := NewWalker(tmpDir)
	repos, err := walker.Walk(context.Background())
	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	if len(repos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(repos))
	}

	repoPaths := make(map[string]bool)
	for _, r := range repos {
		repoPaths[r.Path] = true
	}

	if !repoPaths[repo1] {
		t.Error("expected repo1 to be discovered")
	}
	if !repoPaths[repo2] {
		t.Error("expected repo2 to be discovered")
	}
}

func TestWalker_DiscoverMultipleTopLevelRepos(t *testing.T) {
	tmpDir := t.TempDir()

	repo1 := filepath.Join(tmpDir, "repo1")
	gitDir1 := filepath.Join(repo1, ".git")
	if err := os.MkdirAll(gitDir1, 0755); err != nil {
		t.Fatalf("failed to create repo1 .git: %v", err)
	}

	repo2 := filepath.Join(tmpDir, "repo2")
	gitDir2 := filepath.Join(repo2, ".git")
	if err := os.MkdirAll(gitDir2, 0755); err != nil {
		t.Fatalf("failed to create repo2 .git: %v", err)
	}

	walker := NewWalker(tmpDir)
	repos, err := walker.Walk(context.Background())
	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	if len(repos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(repos))
	}

	repoPaths := make(map[string]bool)
	for _, r := range repos {
		repoPaths[r.Path] = true
	}

	if !repoPaths[repo1] {
		t.Error("expected repo1 to be discovered")
	}
	if !repoPaths[repo2] {
		t.Error("expected repo2 to be discovered")
	}
}

func TestWalker_IgnoresNonGitDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	normalDir := filepath.Join(tmpDir, "normal")
	if err := os.MkdirAll(normalDir, 0755); err != nil {
		t.Fatalf("failed to create normal dir: %v", err)
	}

	nestedDir := filepath.Join(normalDir, "nested")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("failed to create nested dir: %v", err)
	}

	walker := NewWalker(tmpDir)
	repos, err := walker.Walk(context.Background())
	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	if len(repos) != 0 {
		t.Fatalf("expected 0 repos, got %d", len(repos))
	}
}

func TestWalker_ProgressCallback(t *testing.T) {
	tmpDir := t.TempDir()

	repo1 := filepath.Join(tmpDir, "repo1")
	gitDir1 := filepath.Join(repo1, ".git")
	if err := os.MkdirAll(gitDir1, 0755); err != nil {
		t.Fatalf("failed to create repo1 .git: %v", err)
	}

	repo2 := filepath.Join(tmpDir, "repo2")
	gitDir2 := filepath.Join(repo2, ".git")
	if err := os.MkdirAll(gitDir2, 0755); err != nil {
		t.Fatalf("failed to create repo2 .git: %v", err)
	}

	called := false
	walker := NewWalker(tmpDir)
	walker.SetProgressCallback(func(current, total int) {
		called = true
		if total == 0 {
			t.Error("total should not be 0")
		}
	})

	_, err := walker.Walk(context.Background())
	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	if !called {
		t.Error("expected progress callback to be called")
	}
}

func TestWalker_ContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()

	for i := 0; i < 5; i++ {
		subdir := filepath.Join(tmpDir, "subdir")
		if err := os.MkdirAll(subdir, 0755); err != nil {
			t.Fatalf("failed to create subdir: %v", err)
		}
		tmpDir = subdir
	}

	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	walker := NewWalker(t.TempDir())
	_, err := walker.Walk(ctx)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestWalker_SkipDuplicateGitDirs(t *testing.T) {
	tmpDir := t.TempDir()

	repo1 := filepath.Join(tmpDir, "repo1")
	gitDir1 := filepath.Join(repo1, ".git")
	if err := os.MkdirAll(gitDir1, 0755); err != nil {
		t.Fatalf("failed to create repo1 .git: %v", err)
	}

	repo2 := filepath.Join(tmpDir, "repo2")
	gitDir2 := filepath.Join(repo2, ".git")
	if err := os.MkdirAll(gitDir2, 0755); err != nil {
		t.Fatalf("failed to create repo2 .git: %v", err)
	}

	repo3 := filepath.Join(tmpDir, "repo3")
	gitDir3 := filepath.Join(repo3, ".git")
	if err := os.MkdirAll(gitDir3, 0755); err != nil {
		t.Fatalf("failed to create repo3 .git: %v", err)
	}

	walker := NewWalker(tmpDir)
	repos, err := walker.Walk(context.Background())
	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	if len(repos) != 3 {
		t.Errorf("expected 3 repos, got %d", len(repos))
	}
}
