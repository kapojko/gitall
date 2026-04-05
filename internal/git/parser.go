package git

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

type RepoStatus int

const (
	StatusOK RepoStatus = iota
	StatusPush
	StatusChanges
)

type FileChange int

const (
	FileUnmodified FileChange = iota
	FileModified
	FileAdded
	FileDeleted
	FileRenamed
	FileCopied
	FileUnmerged
	FileTypeChanged
	FileUntracked
)

type FileEntry struct {
	Path        string
	Change      FileChange
	Staged      bool
	Unstaged    bool
	IsSubmodule bool
}

type BranchInfo struct {
	Commit   string
	Head     string
	Upstream string
	Ahead    int
	Behind   int
}

type Status struct {
	Branch      BranchInfo
	Files       []FileEntry
	HasChanges  bool
	HasUnpushed bool
	IsSubmodule bool
}

func (s *Status) RepoStatus() RepoStatus {
	if s.HasChanges {
		return StatusChanges
	}
	if s.HasUnpushed {
		return StatusPush
	}
	return StatusOK
}

func ParsePorcelainV2(reader io.Reader) (*Status, error) {
	scanner := bufio.NewScanner(reader)
	status := &Status{
		Files: make([]FileEntry, 0),
	}

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "# branch.oid ") {
			commit := strings.TrimPrefix(line, "# branch.oid ")
			if commit != "(initial)" {
				status.Branch.Commit = commit
			}
		} else if strings.HasPrefix(line, "# branch.head ") {
			status.Branch.Head = strings.TrimPrefix(line, "# branch.head ")
		} else if strings.HasPrefix(line, "# branch.upstream ") {
			status.Branch.Upstream = strings.TrimPrefix(line, "# branch.upstream ")
		} else if strings.HasPrefix(line, "# branch.ab ") {
			ab := strings.TrimPrefix(line, "# branch.ab ")
			parts := strings.Fields(ab)
			for _, p := range parts {
				if strings.HasPrefix(p, "+") {
					aheadStr := strings.TrimPrefix(p, "+")
					if ahead, err := strconv.Atoi(aheadStr); err == nil && ahead > 0 {
						status.HasUnpushed = true
					}
				}
				if strings.HasPrefix(p, "-") {
					behindStr := strings.TrimPrefix(p, "-")
					if behind, err := strconv.Atoi(behindStr); err == nil && behind > 0 {
					}
				}
			}
		} else if len(line) > 0 && line[0] != '#' {
			changed := parseFileLineIntoStatus(line, status)
			if changed {
				status.HasChanges = true
			}
		}
	}

	return status, scanner.Err()
}

func parseFileLineIntoStatus(line string, status *Status) bool {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return false
	}

	first := parts[0]
	if len(first) < 1 {
		return false
	}

	switch first[0] {
	case '1', '2':
		if len(parts) < 7 {
			return false
		}
		entry := parseOrdinaryEntry(parts)
		if entry != nil {
			status.Files = append(status.Files, *entry)
			return entry.Change != FileUnmodified
		}
	case 'u':
		if len(parts) < 10 {
			return false
		}
		entry := parseUnmergedEntry(parts)
		if entry != nil {
			status.Files = append(status.Files, *entry)
			return true
		}
	case '?':
		status.Files = append(status.Files, FileEntry{
			Path:   strings.Join(parts[1:], " "),
			Change: FileUntracked,
		})
		return true
	case '!':
		return false
	}

	return false
}

func parseOrdinaryEntry(parts []string) *FileEntry {
	xy := parts[1]
	submoduleState := parts[2]

	entry := &FileEntry{
		IsSubmodule: strings.HasPrefix(submoduleState, "S"),
		Change:      FileUnmodified,
	}

	if len(xy) >= 2 {
		entry.Staged = xy[0] != '.' && xy[0] != ' '
		entry.Unstaged = xy[1] != '.' && xy[1] != ' '

		if xy[0] != '.' && xy[0] != ' ' {
			switch xy[0] {
			case 'M':
				entry.Change = FileModified
			case 'A':
				entry.Change = FileAdded
			case 'D':
				entry.Change = FileDeleted
			case 'R':
				entry.Change = FileRenamed
			case 'C':
				entry.Change = FileCopied
			case 'T':
				entry.Change = FileTypeChanged
			case 'U':
				entry.Change = FileUnmerged
			}
		} else if xy[1] != '.' && xy[1] != ' ' {
			switch xy[1] {
			case 'M':
				entry.Change = FileModified
			case 'D':
				entry.Change = FileDeleted
			}
		}
	}

	if len(parts) >= 9 {
		idx := 8
		if parts[0] == "2" {
			idx = 10
		}
		if idx < len(parts) {
			entry.Path = strings.Join(parts[idx:], " ")
		}
	}

	return entry
}

func parseUnmergedEntry(parts []string) *FileEntry {
	return &FileEntry{
		Path:        strings.Join(parts[10:], " "),
		Change:      FileUnmerged,
		Staged:      true,
		Unstaged:    true,
		IsSubmodule: strings.HasPrefix(parts[2], "S"),
	}
}
