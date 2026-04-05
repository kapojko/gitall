package status

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"gitall/internal/git"
)

func TestBuildResults_DepthCalculation(t *testing.T) {
	cmd := NewCommand(&bytes.Buffer{})

	childRepo := &git.Repository{
		Path:   "/child",
		Status: &git.Status{},
	}
	parentRepo := &git.Repository{
		Path:   "/parent",
		Status: &git.Status{},
	}
	grandparentRepo := &git.Repository{
		Path:   "/grandparent",
		Status: &git.Status{},
	}

	childRepo.Parent = parentRepo
	parentRepo.Parent = grandparentRepo

	results := cmd.buildResults([]*git.Repository{grandparentRepo, parentRepo, childRepo})

	if results[0].Depth != 0 {
		t.Errorf("grandparent expected depth 0, got %d", results[0].Depth)
	}
	if results[1].Depth != 1 {
		t.Errorf("parent expected depth 1, got %d", results[1].Depth)
	}
	if results[2].Depth != 2 {
		t.Errorf("child expected depth 2, got %d", results[2].Depth)
	}
}

func TestBuildResults_IsSubmodule(t *testing.T) {
	cmd := NewCommand(&bytes.Buffer{})

	rootRepo := &git.Repository{
		Path:   "/root",
		Status: &git.Status{},
	}
	childRepo := &git.Repository{
		Path:   "/root/child",
		Status: &git.Status{},
	}
	childRepo.Parent = rootRepo

	results := cmd.buildResults([]*git.Repository{rootRepo, childRepo})

	if results[0].IsSubmodule {
		t.Error("root should not be submodule")
	}
	if !results[1].IsSubmodule {
		t.Error("child should be submodule")
	}
}

func TestBuildResults_DeepNesting(t *testing.T) {
	cmd := NewCommand(&bytes.Buffer{})

	var repos []*git.Repository
	var prev *git.Repository
	for i := 0; i < 5; i++ {
		repo := &git.Repository{
			Path:   "/level" + string(rune('0'+i)),
			Status: &git.Status{},
		}
		if prev != nil {
			repo.Parent = prev
		}
		repos = append(repos, repo)
		prev = repo
	}

	results := cmd.buildResults(repos)

	for i, r := range results {
		if r.Depth != i {
			t.Errorf("level %d expected depth %d, got %d", i, i, r.Depth)
		}
		expectedSubmodule := i > 0
		if r.IsSubmodule != expectedSubmodule {
			t.Errorf("level %d expected IsSubmodule=%v, got %v", i, expectedSubmodule, r.IsSubmodule)
		}
	}
}

func TestFormatStatus_OK(t *testing.T) {
	cmd := NewCommand(&bytes.Buffer{})

	status := &git.Status{HasChanges: false, HasUnpushed: false}
	result := cmd.formatStatus(status)

	if !strings.Contains(result, "OK") {
		t.Errorf("expected OK in result, got %s", result)
	}
	if !strings.Contains(result, Green) {
		t.Errorf("expected green color in result, got %s", result)
	}
}

func TestFormatStatus_Push(t *testing.T) {
	cmd := NewCommand(&bytes.Buffer{})

	status := &git.Status{HasChanges: false, HasUnpushed: true}
	result := cmd.formatStatus(status)

	if !strings.Contains(result, "PUSH") {
		t.Errorf("expected PUSH in result, got %s", result)
	}
	if !strings.Contains(result, Yellow) {
		t.Errorf("expected yellow color in result, got %s", result)
	}
}

func TestFormatStatus_Changes(t *testing.T) {
	cmd := NewCommand(&bytes.Buffer{})

	status := &git.Status{HasChanges: true}
	result := cmd.formatStatus(status)

	if !strings.Contains(result, "CHANGES") {
		t.Errorf("expected CHANGES in result, got %s", result)
	}
	if !strings.Contains(result, Red) {
		t.Errorf("expected red color in result, got %s", result)
	}
}

func TestCountStatuses(t *testing.T) {
	cmd := NewCommand(&bytes.Buffer{})

	results := []RepoResult{
		{Status: &git.Status{HasChanges: false, HasUnpushed: false}},
		{Status: &git.Status{HasChanges: false, HasUnpushed: false}},
		{Status: &git.Status{HasChanges: false, HasUnpushed: true}},
		{Status: &git.Status{HasChanges: true}},
	}

	ok, push, changes := cmd.countStatuses(results)

	if ok != 2 {
		t.Errorf("expected ok=2, got %d", ok)
	}
	if push != 1 {
		t.Errorf("expected push=1, got %d", push)
	}
	if changes != 1 {
		t.Errorf("expected changes=1, got %d", changes)
	}
}

func TestCountStatuses_AllOK(t *testing.T) {
	cmd := NewCommand(&bytes.Buffer{})

	results := []RepoResult{
		{Status: &git.Status{HasChanges: false, HasUnpushed: false}},
		{Status: &git.Status{HasChanges: false, HasUnpushed: false}},
	}

	ok, push, changes := cmd.countStatuses(results)

	if ok != 2 {
		t.Errorf("expected ok=2, got %d", ok)
	}
	if push != 0 {
		t.Errorf("expected push=0, got %d", push)
	}
	if changes != 0 {
		t.Errorf("expected changes=0, got %d", changes)
	}
}

func TestPrintResults_SingleRepo(t *testing.T) {
	var buf bytes.Buffer
	cmd := NewCommand(&buf)

	results := []RepoResult{
		{Path: "/test/repo", Depth: 0, Status: &git.Status{HasChanges: false}},
	}

	cmd.printResults(results)

	output := buf.String()
	if !strings.Contains(output, "/test/repo") {
		t.Errorf("expected path in output, got %s", output)
	}
}

func TestPrintResults_TreeFormatting(t *testing.T) {
	var buf bytes.Buffer
	cmd := NewCommand(&buf)

	results := []RepoResult{
		{Path: "/root", Depth: 0, Status: &git.Status{HasChanges: false}},
		{Path: "/root/child1", Depth: 1, Status: &git.Status{HasChanges: false}},
		{Path: "/root/child2", Depth: 1, Status: &git.Status{HasChanges: false}},
		{Path: "/root/child1/grandchild", Depth: 2, Status: &git.Status{HasChanges: false}},
	}

	cmd.printResults(results)

	output := buf.String()

	if !strings.Contains(output, "├── /root/child1") {
		t.Errorf("expected tree branch for child1, got %s", output)
	}
	if !strings.Contains(output, "└── /root/child2") {
		t.Errorf("expected tree last branch for child2, got %s", output)
	}
	if !strings.Contains(output, "│   └── /root/child1/grandchild") {
		t.Errorf("expected tree continuation for grandchild, got %s", output)
	}
}

func TestPrintResults_EmptyResults(t *testing.T) {
	var buf bytes.Buffer
	cmd := NewCommand(&buf)

	results := []RepoResult{}

	cmd.printResults(results)

	output := buf.String()
	if !strings.Contains(output, "No git repositories found") {
		t.Errorf("expected 'No git repositories found' message, got %s", output)
	}
}

func TestPrintSummary_WithStatuses(t *testing.T) {
	var buf bytes.Buffer
	cmd := NewCommand(&buf)

	cmd.printSummary(2, 1, 3)

	output := buf.String()
	if !strings.Contains(output, "Summary:") {
		t.Errorf("expected 'Summary:' in output, got %s", output)
	}
	if !strings.Contains(output, ": 2") || !strings.Contains(output, "OK") {
		t.Errorf("expected OK count in output, got %s", output)
	}
	if !strings.Contains(output, ": 1") || !strings.Contains(output, "PUSH") {
		t.Errorf("expected PUSH count in output, got %s", output)
	}
	if !strings.Contains(output, ": 3") || !strings.Contains(output, "CHANGES") {
		t.Errorf("expected CHANGES count in output, got %s", output)
	}
}

func TestPrintSummary_OnlyOK(t *testing.T) {
	var buf bytes.Buffer
	cmd := NewCommand(&buf)

	cmd.printSummary(5, 0, 0)

	output := buf.String()
	if !strings.Contains(output, ": 5") || !strings.Contains(output, "OK") {
		t.Errorf("expected 'OK: 5' in output, got %s", output)
	}
	if strings.Contains(output, "PUSH") {
		t.Errorf("expected no PUSH in output, got %s", output)
	}
	if strings.Contains(output, "CHANGES") {
		t.Errorf("expected no CHANGES in output, got %s", output)
	}
}

func TestPrintSummary_NoRepos(t *testing.T) {
	var buf bytes.Buffer
	cmd := NewCommand(&buf)

	cmd.printSummary(0, 0, 0)

	output := buf.String()
	if !strings.Contains(output, "No repositories found") {
		t.Errorf("expected 'No repositories found' in output, got %s", output)
	}
}

func TestColorize(t *testing.T) {
	result := colorize("test", Red)
	expected := Red + "test" + Reset

	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestExecute_Integration(t *testing.T) {
	tmpDir := t.TempDir()

	gitDir := tmpDir + "/.git"
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}

	var buf bytes.Buffer
	cmd := NewCommand(&buf)

	err := cmd.Execute(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "OK") {
		t.Errorf("expected OK status in output, got %s", output)
	}
	if !strings.Contains(output, "Summary:") {
		t.Errorf("expected Summary in output, got %s", output)
	}
}
