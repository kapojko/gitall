package git

import (
	"io"
	"strings"
	"testing"
)

func TestParsePorcelainV2_CleanRepository(t *testing.T) {
	input := `# branch.oid 9f78131cdb71b6b373a10180cdc1c5a315c457bf
# branch.head main
# branch.upstream origin/main
# branch.ab +0 -0
`
	status, err := ParsePorcelainV2(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status.Branch.Commit != "9f78131cdb71b6b373a10180cdc1c5a315c457bf" {
		t.Errorf("expected commit 9f78131cdb71b6b373a10180cdc1c5a315c457bf, got %s", status.Branch.Commit)
	}
	if status.Branch.Head != "main" {
		t.Errorf("expected head main, got %s", status.Branch.Head)
	}
	if status.Branch.Upstream != "origin/main" {
		t.Errorf("expected upstream origin/main, got %s", status.Branch.Upstream)
	}
	if status.HasChanges {
		t.Error("expected no changes")
	}
	if status.HasUnpushed {
		t.Error("expected no unpushed changes")
	}
	if len(status.Files) != 0 {
		t.Errorf("expected 0 files, got %d", len(status.Files))
	}
	if status.RepoStatus() != StatusOK {
		t.Errorf("expected StatusOK, got %v", status.RepoStatus())
	}
}

func TestParsePorcelainV2_InitialCommit(t *testing.T) {
	input := `# branch.oid (initial)
# branch.head main
`
	status, err := ParsePorcelainV2(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status.Branch.Commit != "" {
		t.Errorf("expected empty commit for initial, got %s", status.Branch.Commit)
	}
	if status.Branch.Head != "main" {
		t.Errorf("expected head main, got %s", status.Branch.Head)
	}
	if status.RepoStatus() != StatusOK {
		t.Errorf("expected StatusOK, got %v", status.RepoStatus())
	}
}

func TestParsePorcelainV2_AheadBehindStatus(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		unpushed bool
	}{
		{
			name: "ahead only",
			input: `# branch.oid 9f78131cdb71b6b373a10180cdc1c5a315c457bf
# branch.head main
# branch.upstream origin/main
# branch.ab +3 -0
`,
			unpushed: true,
		},
		{
			name: "behind only",
			input: `# branch.oid 9f78131cdb71b6b373a10180cdc1c5a315c457bf
# branch.head main
# branch.upstream origin/main
# branch.ab +0 -2
`,
			unpushed: false,
		},
		{
			name: "ahead and behind",
			input: `# branch.oid 9f78131cdb71b6b373a10180cdc1c5a315c457bf
# branch.head main
# branch.upstream origin/main
# branch.ab +5 -3
`,
			unpushed: true,
		},
		{
			name: "zero ahead and behind",
			input: `# branch.oid 9f78131cdb71b6b373a10180cdc1c5a315c457bf
# branch.head main
# branch.upstream origin/main
# branch.ab +0 -0
`,
			unpushed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := ParsePorcelainV2(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if status.HasUnpushed != tt.unpushed {
				t.Errorf("expected HasUnpushed=%v, got %v", tt.unpushed, status.HasUnpushed)
			}
		})
	}
}

func TestParsePorcelainV2_ModifiedFile(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		staged   bool
		unstaged bool
		change   FileChange
		path     string
	}{
		{
			name:     "staged modified",
			input:    "1 M. N... 100644 100644 100644 abc123 def456 modified.go\n",
			staged:   true,
			unstaged: false,
			change:   FileModified,
			path:     "modified.go",
		},
		{
			name:     "unstaged modified",
			input:    "1 .M N... 100644 100644 100644 abc123 def456 modified.go\n",
			staged:   false,
			unstaged: true,
			change:   FileModified,
			path:     "modified.go",
		},
		{
			name:     "staged and unstaged modified",
			input:    "1 MM N... 100644 100644 100644 abc123 def456 modified.go\n",
			staged:   true,
			unstaged: true,
			change:   FileModified,
			path:     "modified.go",
		},
		{
			name:     "modified with space separator",
			input:    "1 M. N... 100644 100644 100644 abc123 def456 path with spaces.go\n",
			staged:   true,
			unstaged: false,
			change:   FileModified,
			path:     "path with spaces.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := ParsePorcelainV2(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !status.HasChanges {
				t.Error("expected HasChanges=true")
			}
			if len(status.Files) != 1 {
				t.Fatalf("expected 1 file, got %d", len(status.Files))
			}
			f := status.Files[0]
			if f.Path != tt.path {
				t.Errorf("expected path %q, got %q", tt.path, f.Path)
			}
			if f.Change != tt.change {
				t.Errorf("expected change %v, got %v", tt.change, f.Change)
			}
			if f.Staged != tt.staged {
				t.Errorf("expected Staged=%v, got %v", tt.staged, f.Staged)
			}
			if f.Unstaged != tt.unstaged {
				t.Errorf("expected Unstaged=%v, got %v", tt.unstaged, f.Unstaged)
			}
		})
	}
}

func TestParsePorcelainV2_AddedFile(t *testing.T) {
	input := `# branch.oid 9f78131cdb71b6b373a10180cdc1c5a315c457bf
# branch.head main
# branch.upstream origin/main
# branch.ab +0 -0
1 A. N... 100644 100644 100644 0000000 abc123 newfile.go
`
	status, err := ParsePorcelainV2(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !status.HasChanges {
		t.Error("expected HasChanges=true")
	}
	if len(status.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(status.Files))
	}
	f := status.Files[0]
	if f.Path != "newfile.go" {
		t.Errorf("expected path newfile.go, got %s", f.Path)
	}
	if f.Change != FileAdded {
		t.Errorf("expected FileAdded, got %v", f.Change)
	}
	if !f.Staged {
		t.Error("expected Staged=true")
	}
	if f.Unstaged {
		t.Error("expected Unstaged=false")
	}
}

func TestParsePorcelainV2_DeletedFile(t *testing.T) {
	tests := []struct {
		name     string
		xy       string
		staged   bool
		unstaged bool
	}{
		{name: "staged deleted", xy: "D.", staged: true, unstaged: false},
		{name: "unstaged deleted", xy: ".D", staged: false, unstaged: true},
		{name: "both deleted", xy: "DD", staged: true, unstaged: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := "1 " + tt.xy + " N... 100644 100644 100644 abc123 0000000 deleted.go\n"
			status, err := ParsePorcelainV2(strings.NewReader(input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !status.HasChanges {
				t.Error("expected HasChanges=true")
			}
			if len(status.Files) != 1 {
				t.Fatalf("expected 1 file, got %d", len(status.Files))
			}
			f := status.Files[0]
			if f.Change != FileDeleted {
				t.Errorf("expected FileDeleted, got %v", f.Change)
			}
			if f.Staged != tt.staged {
				t.Errorf("expected Staged=%v, got %v", tt.staged, f.Staged)
			}
			if f.Unstaged != tt.unstaged {
				t.Errorf("expected Unstaged=%v, got %v", tt.unstaged, f.Unstaged)
			}
		})
	}
}

func TestParsePorcelainV2_RenamedFile(t *testing.T) {
	input := "2 R. N... 100644 100644 100644 abc123 def456 R100 old_name.go\t new_name.go\n"
	status, err := ParsePorcelainV2(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !status.HasChanges {
		t.Error("expected HasChanges=true")
	}
	if len(status.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(status.Files))
	}
	f := status.Files[0]
	if f.Path != "new_name.go" {
		t.Errorf("expected path new_name.go, got %s", f.Path)
	}
	if f.Change != FileRenamed {
		t.Errorf("expected FileRenamed, got %v", f.Change)
	}
	if !f.Staged {
		t.Error("expected Staged=true")
	}
	if f.Unstaged {
		t.Error("expected Unstaged=false")
	}
}

func TestParsePorcelainV2_CopiedFile(t *testing.T) {
	input := "2 C. N... 100644 100644 100644 abc123 def456 C75 original.go\t copy.go\n"
	status, err := ParsePorcelainV2(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !status.HasChanges {
		t.Error("expected HasChanges=true")
	}
	if len(status.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(status.Files))
	}
	f := status.Files[0]
	if f.Path != "copy.go" {
		t.Errorf("expected path copy.go, got %s", f.Path)
	}
	if f.Change != FileCopied {
		t.Errorf("expected FileCopied, got %v", f.Change)
	}
	if !f.Staged {
		t.Error("expected Staged=true")
	}
}

func TestParsePorcelainV2_TypeChanged(t *testing.T) {
	tests := []struct {
		name     string
		xy       string
		staged   bool
		unstaged bool
	}{
		{name: "staged type changed", xy: "T.", staged: true, unstaged: false},
		{name: "both type changed", xy: "TT", staged: true, unstaged: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := "1 " + tt.xy + " N... 100644 100644 100644 abc123 def456 type_changed.go\n"
			status, err := ParsePorcelainV2(strings.NewReader(input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !status.HasChanges {
				t.Error("expected HasChanges=true")
			}
			if len(status.Files) != 1 {
				t.Fatalf("expected 1 file, got %d", len(status.Files))
			}
			f := status.Files[0]
			if f.Change != FileTypeChanged {
				t.Errorf("expected FileTypeChanged, got %v", f.Change)
			}
			if f.Staged != tt.staged {
				t.Errorf("expected Staged=%v, got %v", tt.staged, f.Staged)
			}
			if f.Unstaged != tt.unstaged {
				t.Errorf("expected Unstaged=%v, got %v", tt.unstaged, f.Unstaged)
			}
		})
	}
}

func TestParsePorcelainV2_UnmergedEntries(t *testing.T) {
	tests := []struct {
		name string
		xy   string
	}{
		{name: "both deleted", xy: "DD"},
		{name: "added by us", xy: "AU"},
		{name: "deleted by them", xy: "UD"},
		{name: "added by them", xy: "UA"},
		{name: "deleted by us", xy: "DU"},
		{name: "both added", xy: "AA"},
		{name: "both modified", xy: "UU"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := "u " + tt.xy + " N... 100644 100644 100644 100644 abc123 abc123 abc123 unmerged.go\n"
			status, err := ParsePorcelainV2(strings.NewReader(input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !status.HasChanges {
				t.Error("expected HasChanges=true")
			}
			if len(status.Files) != 1 {
				t.Fatalf("expected 1 file, got %d", len(status.Files))
			}
			f := status.Files[0]
			if f.Change != FileUnmerged {
				t.Errorf("expected FileUnmerged, got %v", f.Change)
			}
			if f.Path != "unmerged.go" {
				t.Errorf("expected path unmerged.go, got %s", f.Path)
			}
			if !f.Staged {
				t.Error("expected Staged=true for unmerged")
			}
			if !f.Unstaged {
				t.Error("expected Unstaged=true for unmerged")
			}
		})
	}
}

func TestParsePorcelainV2_UntrackedFile(t *testing.T) {
	input := `# branch.oid 9f78131cdb71b6b373a10180cdc1c5a315c457bf
# branch.head main
# branch.upstream origin/main
# branch.ab +0 -0
? untracked.go
? path with spaces.go
`
	status, err := ParsePorcelainV2(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !status.HasChanges {
		t.Error("expected HasChanges=true")
	}
	if len(status.Files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(status.Files))
	}
	if status.Files[0].Change != FileUntracked {
		t.Errorf("expected FileUntracked, got %v", status.Files[0].Change)
	}
	if status.Files[0].Path != "untracked.go" {
		t.Errorf("expected path untracked.go, got %s", status.Files[0].Path)
	}
	if status.Files[1].Path != "path with spaces.go" {
		t.Errorf("expected path 'path with spaces.go', got %s", status.Files[1].Path)
	}
}

func TestParsePorcelainV2_IgnoredFile(t *testing.T) {
	input := `# branch.oid 9f78131cdb71b6b373a10180cdc1c5a315c457bf
# branch.head main
# branch.upstream origin/main
# branch.ab +0 -0
! ignored.go
`
	status, err := ParsePorcelainV2(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status.HasChanges {
		t.Error("expected HasChanges=false for ignored files only")
	}
	if len(status.Files) != 0 {
		t.Errorf("expected 0 tracked files, got %d", len(status.Files))
	}
}

func TestParsePorcelainV2_SubmoduleStates(t *testing.T) {
	tests := []struct {
		name        string
		subState    string
		isSubmodule bool
	}{
		{name: "not submodule", subState: "N...", isSubmodule: false},
		{name: "submodule with commit change", subState: "SC..", isSubmodule: true},
		{name: "submodule with tracked changes", subState: "S.M.", isSubmodule: true},
		{name: "submodule with untracked files", subState: "S..U", isSubmodule: true},
		{name: "submodule with all states", subState: "SCM.", isSubmodule: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := "1 M. " + tt.subState + " 100644 100644 100644 abc123 def456 submodule.go\n"
			status, err := ParsePorcelainV2(strings.NewReader(input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(status.Files) != 1 {
				t.Fatalf("expected 1 file, got %d", len(status.Files))
			}
			f := status.Files[0]
			if f.IsSubmodule != tt.isSubmodule {
				t.Errorf("expected IsSubmodule=%v, got %v", tt.isSubmodule, f.IsSubmodule)
			}
		})
	}
}

func TestParsePorcelainV2_UnmergedSubmodule(t *testing.T) {
	input := "u DD S... 100644 100644 100644 100644 abc123 abc123 abc123 abc123 unmerged_submodule\n"
	status, err := ParsePorcelainV2(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(status.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(status.Files))
	}
	f := status.Files[0]
	if !f.IsSubmodule {
		t.Error("expected IsSubmodule=true")
	}
	if f.Change != FileUnmerged {
		t.Errorf("expected FileUnmerged, got %v", f.Change)
	}
}

func TestParsePorcelainV2_MultipleChanges(t *testing.T) {
	input := `# branch.oid 9f78131cdb71b6b373a10180cdc1c5a315c457bf
# branch.head main
# branch.upstream origin/main
# branch.ab +0 -0
1 A. N... 100644 100644 100644 0000000 abc123 newfile.go
1 .M N... 100644 100644 100644 abc123 def456 modified.go
1 D. N... 100644 100644 100644 abc123 0000000 deleted.go
? untracked_file.go
`
	status, err := ParsePorcelainV2(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !status.HasChanges {
		t.Error("expected HasChanges=true")
	}
	if len(status.Files) != 4 {
		t.Errorf("expected 4 files, got %d", len(status.Files))
	}
	if status.RepoStatus() != StatusChanges {
		t.Errorf("expected StatusChanges, got %v", status.RepoStatus())
	}
}

func TestParsePorcelainV2_DetachedHead(t *testing.T) {
	input := `# branch.oid 9f78131cdb71b6b373a10180cdc1c5a315c457bf
# branch.head (detached)
`
	status, err := ParsePorcelainV2(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status.Branch.Head != "(detached)" {
		t.Errorf("expected head (detached), got %s", status.Branch.Head)
	}
}

func TestParsePorcelainV2_NoUpstream(t *testing.T) {
	input := `# branch.oid 9f78131cdb71b6b373a10180cdc1c5a315c457bf
# branch.head main
`
	status, err := ParsePorcelainV2(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status.Branch.Upstream != "" {
		t.Errorf("expected empty upstream, got %s", status.Branch.Upstream)
	}
}

func TestParsePorcelainV2_UnmodifiedEntry(t *testing.T) {
	input := "1 .. N... 100644 100644 100644 abc123 abc123 unchanged.go\n"
	status, err := ParsePorcelainV2(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status.HasChanges {
		t.Error("expected HasChanges=false for unmodified entry")
	}
	if len(status.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(status.Files))
	}
	f := status.Files[0]
	if f.Change != FileUnmodified {
		t.Errorf("expected FileUnmodified, got %v", f.Change)
	}
}

func TestParsePorcelainV2_EmptyInput(t *testing.T) {
	status, err := ParsePorcelainV2(strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status.Branch.Commit != "" {
		t.Errorf("expected empty commit, got %s", status.Branch.Commit)
	}
	if status.Branch.Head != "" {
		t.Errorf("expected empty head, got %s", status.Branch.Head)
	}
	if status.HasChanges {
		t.Error("expected HasChanges=false")
	}
	if len(status.Files) != 0 {
		t.Errorf("expected 0 files, got %d", len(status.Files))
	}
}

func TestParsePorcelainV2_UnknownHeader(t *testing.T) {
	input := `# branch.oid 9f78131cdb71b6b373a10180cdc1c5a315c457bf
# branch.head main
# some.other.header some value
1 M. N... 100644 100644 100644 abc123 def456 file.go
`
	status, err := ParsePorcelainV2(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(status.Files) != 1 {
		t.Errorf("expected 1 file, got %d", len(status.Files))
	}
}

func TestParsePorcelainV2_CommentLine(t *testing.T) {
	input := `# branch.oid 9f78131cdb71b6b373a10180cdc1c5a315c457bf
# branch.head main
# This is a comment
1 M. N... 100644 100644 100644 abc123 def456 file.go
`
	status, err := ParsePorcelainV2(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(status.Files) != 1 {
		t.Errorf("expected 1 file, got %d", len(status.Files))
	}
}

func TestParsePorcelainV2_InvalidLineFormats(t *testing.T) {
	tests := []struct {
		name  string
		input string
		files int
	}{
		{name: "empty line", input: "\n", files: 0},
		{name: "single char line", input: "1\n", files: 0},
		{name: "incomplete ordinary entry", input: "1 M. N... 100644\n", files: 0},
		{name: "incomplete unmerged entry", input: "u DD N... 100644 100644\n", files: 0},
		{name: "untracked without path", input: "?\n", files: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := ParsePorcelainV2(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(status.Files) != tt.files {
				t.Errorf("expected %d files, got %d", tt.files, len(status.Files))
			}
		})
	}
}

func TestParsePorcelainV2_RepoStatus(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedStatus RepoStatus
	}{
		{
			name: "StatusOK - no changes, no unpushed",
			input: `# branch.oid 9f78131cdb71b6b373a10180cdc1c5a315c457bf
# branch.head main
# branch.upstream origin/main
# branch.ab +0 -0
`,
			expectedStatus: StatusOK,
		},
		{
			name: "StatusPush - no changes, ahead",
			input: `# branch.oid 9f78131cdb71b6b373a10180cdc1c5a315c457bf
# branch.head main
# branch.upstream origin/main
# branch.ab +3 -0
`,
			expectedStatus: StatusPush,
		},
		{
			name: "StatusChanges - has changes",
			input: `# branch.oid 9f78131cdb71b6b373a10180cdc1c5a315c457bf
# branch.head main
1 M. N... 100644 100644 100644 abc123 def456 file.go
`,
			expectedStatus: StatusChanges,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := ParsePorcelainV2(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if status.RepoStatus() != tt.expectedStatus {
				t.Errorf("expected %v, got %v", tt.expectedStatus, status.RepoStatus())
			}
		})
	}
}

func TestParsePorcelainV2_XYStates(t *testing.T) {
	xyTests := []struct {
		xy               string
		expectedStaged   bool
		expectedUnstaged bool
		expectedChange   FileChange
	}{
		{"..", false, false, FileUnmodified},
		{"M.", true, false, FileModified},
		{".M", false, true, FileModified},
		{"MM", true, true, FileModified},
		{"A.", true, false, FileAdded},
		{"D.", true, false, FileDeleted},
		{".D", false, true, FileDeleted},
		{"R.", true, false, FileRenamed},
		{"C.", true, false, FileCopied},
		{"T.", true, false, FileTypeChanged},
	}

	for _, tt := range xyTests {
		input := "1 " + tt.xy + " N... 100644 100644 100644 abc123 def456 file.go\n"
		status, err := ParsePorcelainV2(strings.NewReader(input))
		if err != nil {
			t.Fatalf("unexpected error for XY %q: %v", tt.xy, err)
		}
		if len(status.Files) != 1 {
			t.Fatalf("expected 1 file for XY %q, got %d", tt.xy, len(status.Files))
		}
		f := status.Files[0]
		if f.Staged != tt.expectedStaged {
			t.Errorf("XY %q: expected Staged=%v, got %v", tt.xy, tt.expectedStaged, f.Staged)
		}
		if f.Unstaged != tt.expectedUnstaged {
			t.Errorf("XY %q: expected Unstaged=%v, got %v", tt.xy, tt.expectedUnstaged, f.Unstaged)
		}
		if f.Change != tt.expectedChange {
			t.Errorf("XY %q: expected Change=%v, got %v", tt.xy, tt.expectedChange, f.Change)
		}
	}
}

func TestParsePorcelainV2_ComplexScenarios(t *testing.T) {
	input := `# branch.oid 9f78131cdb71b6b373a10180cdc1c5a315c457bf
# branch.head main
# branch.upstream origin/main
# branch.ab +2 -1
1 A. N... 100644 100644 100644 0000000 abc123 added.go
1 .M N... 100644 100644 100644 abc123 def456 modified_works.go
1 D. N... 100644 100644 100644 abc123 0000000 deleted.go
1 M. N... 100644 100644 100644 abc123 def456 modified_staged.go
2 R. N... 100644 100644 100644 abc123 def456 R100 old.go	 new.go
u DD N... 100644 100644 100644 100644 abc123 abc123 abc123 both_deleted.go
u AU N... 100644 100644 100644 100644 0000000 abc123 abc123 added_by_us.go
? brand_new.go
`
	status, err := ParsePorcelainV2(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !status.HasChanges {
		t.Error("expected HasChanges=true")
	}
	if !status.HasUnpushed {
		t.Error("expected HasUnpushed=true")
	}
	if len(status.Files) != 8 {
		t.Errorf("expected 8 files, got %d", len(status.Files))
	}

	fileMap := make(map[string]FileEntry)
	for _, f := range status.Files {
		fileMap[f.Path] = f
	}

	if _, ok := fileMap["added.go"]; !ok {
		t.Error("expected added.go in files")
	}
	if fileMap["added.go"].Change != FileAdded {
		t.Errorf("expected FileAdded for added.go, got %v", fileMap["added.go"].Change)
	}

	if _, ok := fileMap["modified_works.go"]; !ok {
		t.Error("expected modified_works.go in files")
	}
	if !fileMap["modified_works.go"].Unstaged {
		t.Error("expected modified_works.go to be unstaged")
	}

	if _, ok := fileMap["deleted.go"]; !ok {
		t.Error("expected deleted.go in files")
	}
	if fileMap["deleted.go"].Change != FileDeleted {
		t.Errorf("expected FileDeleted for deleted.go, got %v", fileMap["deleted.go"].Change)
	}

	if _, ok := fileMap["modified_staged.go"]; !ok {
		t.Error("expected modified_staged.go in files")
	}
	if !fileMap["modified_staged.go"].Staged {
		t.Error("expected modified_staged.go to be staged")
	}

	if _, ok := fileMap["new.go"]; !ok {
		t.Error("expected new.go in files")
	}
	if fileMap["new.go"].Change != FileRenamed {
		t.Errorf("expected FileRenamed for new.go, got %v", fileMap["new.go"].Change)
	}

	if _, ok := fileMap["both_deleted.go"]; !ok {
		t.Error("expected both_deleted.go in files")
	}
	if fileMap["both_deleted.go"].Change != FileUnmerged {
		t.Errorf("expected FileUnmerged for both_deleted.go, got %v", fileMap["both_deleted.go"].Change)
	}

	if _, ok := fileMap["added_by_us.go"]; !ok {
		t.Error("expected added_by_us.go in files")
	}

	if _, ok := fileMap["brand_new.go"]; !ok {
		t.Error("expected brand_new.go in files")
	}
	if fileMap["brand_new.go"].Change != FileUntracked {
		t.Errorf("expected FileUntracked for brand_new.go, got %v", fileMap["brand_new.go"].Change)
	}
}

func TestParsePorcelainV2_RealWorldExample(t *testing.T) {
	input := `# branch.oid 543da8a1091eeb1cb205f75245b417e846a191c0
# branch.head master
# branch.upstream origin/master
# branch.ab +1 -0
1 .M N... 100644 100644 100644 6b8195f5ac65893ce6f16c5aaacab8a421f649b2 6b8195f5ac65893ce6f16c5aaacab8a421f649b2 impl/gwsynthesis/Nano1k_Test.vg
1 .M N... 100644 100644 100644 754272aa391ea58350e18bb4a39c33dafb27b9b0 754272aa391ea58350e18bb4a39c33dafb27b9b0 impl/gwsynthesis/Nano1k_Test_syn.rpt.html
`
	status, err := ParsePorcelainV2(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status.Branch.Commit != "543da8a1091eeb1cb205f75245b417e846a191c0" {
		t.Errorf("unexpected commit: %s", status.Branch.Commit)
	}
	if status.Branch.Head != "master" {
		t.Errorf("unexpected head: %s", status.Branch.Head)
	}
	if status.Branch.Upstream != "origin/master" {
		t.Errorf("unexpected upstream: %s", status.Branch.Upstream)
	}
	if !status.HasUnpushed {
		t.Error("expected HasUnpushed=true")
	}
	if !status.HasChanges {
		t.Error("expected HasChanges=true")
	}
	if len(status.Files) != 2 {
		t.Errorf("expected 2 files, got %d", len(status.Files))
	}
	for _, f := range status.Files {
		if !f.Unstaged {
			t.Errorf("expected %s to be unstaged", f.Path)
		}
		if f.Change != FileModified {
			t.Errorf("expected FileModified for %s, got %v", f.Path, f.Change)
		}
	}
}

type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func TestParsePorcelainV2_ReadError(t *testing.T) {
	_, err := ParsePorcelainV2(&errorReader{})
	if err == nil {
		t.Error("expected error for read failure")
	}
}

func TestParsePorcelainV2_PathWithSpaces(t *testing.T) {
	input := "1 M. N... 100644 100644 100644 abc123 def456 path with spaces in it.go\n"
	status, err := ParsePorcelainV2(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(status.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(status.Files))
	}
	if status.Files[0].Path != "path with spaces in it.go" {
		t.Errorf("expected 'path with spaces in it.go', got %q", status.Files[0].Path)
	}
}

func TestParsePorcelainV2_PathWithTabInRenamed(t *testing.T) {
	input := "2 R. N... 100644 100644 100644 abc123 def456 R100 old_path.go\tnew_path.go\n"
	status, err := ParsePorcelainV2(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(status.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(status.Files))
	}
	if status.Files[0].Path != "new_path.go" {
		t.Errorf("expected 'new_path.go', got %q", status.Files[0].Path)
	}
}
