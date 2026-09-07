package git

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

// Worktree describes one of the repository's linked working trees.
type Worktree struct {
	Path   string
	Branch string // short branch name, or a short commit hash when detached
}

// Worktrees lists the linked working trees. go-git has no API for it, and
// git's porcelain format is the intended way to read it, so this shells out.
func (r *Repo) Worktrees() ([]Worktree, error) {
	worktree, err := r.repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("getting worktree: %w", err)
	}
	root := worktree.Filesystem.Root()

	out, err := exec.Command("git", "-C", root, "worktree", "list", "--porcelain").Output()
	if err != nil {
		return nil, fmt.Errorf("listing worktrees: %w", err)
	}

	return parseWorktreeList(string(out)), nil
}

// parseWorktreeList reads `git worktree list --porcelain`: blank-line separated
// entries, each a "worktree <path>" line plus a branch or detached line.
func parseWorktreeList(out string) []Worktree {
	var worktrees []Worktree
	var current *Worktree
	var head string

	flush := func() {
		if current == nil {
			return
		}
		if current.Branch == "" && len(head) >= 7 {
			current.Branch = head[:7]
		}
		worktrees = append(worktrees, *current)
		current, head = nil, ""
	}

	for line := range strings.SplitSeq(out, "\n") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			flush()
			current = &Worktree{Path: strings.TrimPrefix(line, "worktree ")}
		case strings.HasPrefix(line, "HEAD "):
			head = strings.TrimPrefix(line, "HEAD ")
		case strings.HasPrefix(line, "branch refs/heads/"):
			if current != nil {
				current.Branch = strings.TrimPrefix(line, "branch refs/heads/")
			}
		}
	}
	flush()

	return worktrees
}

// ChangedFilesAt lists uncommitted changes in the tree rooted at path. Safe on
// any worktree, unlike go-git's Status(), which misreads a linked one's index.
func ChangedFilesAt(path string) ([]FileChange, error) {
	out, err := exec.Command("git", "-C", path, "status", "--porcelain").Output()
	if err != nil {
		return nil, fmt.Errorf("git status: %w", err)
	}

	return parsePorcelainStatus(string(out), false), nil
}

// ChangedFilesAgainst lists path's changes relative to ref, untracked files
// included. status and diff only compare against HEAD, so this seeds a scratch
// index from ref's tree and runs status against that.
func ChangedFilesAgainst(path, ref string) ([]FileChange, error) {
	tmpIndex, err := tempIndexPath()
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpIndex)
	env := append(os.Environ(), "GIT_INDEX_FILE="+tmpIndex)

	readTree := exec.Command("git", "-C", path, "read-tree", ref)
	readTree.Env = env
	if out, err := readTree.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("seeding index from %s: %w: %s", ref, err, out)
	}

	status := exec.Command("git", "-C", path, "status", "--porcelain")
	status.Env = env
	out, err := status.Output()
	if err != nil {
		return nil, fmt.Errorf("git status: %w", err)
	}

	return parsePorcelainStatus(string(out), true), nil
}

// parsePorcelainStatus reads `git status --porcelain`: two status characters,
// a space, then the path. worktreeColumnOnly takes the second character always,
// right when the index was seeded from something other than HEAD.
func parsePorcelainStatus(out string, worktreeColumnOnly bool) []FileChange {
	var changes []FileChange

	for line := range strings.SplitSeq(out, "\n") {
		if len(line) < 4 {
			continue
		}

		code := line[0]
		switch {
		case worktreeColumnOnly:
			code = line[1]
			if code == ' ' {
				continue // no change relative to the seeded index
			}
		case code == ' ' || code == '?':
			// A file can carry a staged and an unstaged status — take either.
			code = line[1]
		}

		path := line[3:]
		from, to := path, path
		change := FileChange{From: &from, To: &to}

		switch code {
		case 'A', '?':
			change.From = nil
			change.Type = ChangeTypeNew
		case 'D':
			change.To = nil
			change.Type = ChangeTypeDeleted
		case 'R':
			oldPath, newPath, ok := strings.Cut(path, " -> ")
			if ok {
				change.From, change.To = &oldPath, &newPath
			}
			change.Type = ChangeTypeRenamed
		default:
			change.Type = ChangeTypeModified
		}

		changes = append(changes, change)
	}

	sort.Slice(changes, func(i, j int) bool { return changes[i].Name() < changes[j].Name() })

	return changes
}
