// Package git wraps go-git with the few operations gdiff needs.
package git

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	gogit "github.com/go-git/go-git/v5"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// Repo is an opened repository. Open it once and reuse it — every query
// below goes through it, so nothing has to hit the filesystem again.
type Repo struct {
	repo *gogit.Repository
}

// Open opens the repository containing path, searching parent directories
// for the .git directory if necessary.
func Open(path string) (*Repo, error) {
	repo, err := gogit.PlainOpenWithOptions(path, &gogit.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return nil, fmt.Errorf("opening repository: %w", err)
	}
	return &Repo{repo: repo}, nil
}

// ChangeType classifies what happened to a file.
type ChangeType int

const (
	ChangeTypeNew ChangeType = iota
	ChangeTypeModified
	ChangeTypeDeleted
	ChangeTypeRenamed
)

// Symbol returns a one-character glyph for compact display, e.g. a file
// list or a count like "+3 ~12".
func (t ChangeType) Symbol() string {
	switch t {
	case ChangeTypeNew:
		return "+"
	case ChangeTypeDeleted:
		return "-"
	case ChangeTypeRenamed:
		return "»"
	default:
		return "~"
	}
}

// FileChange is one changed file. From is nil for a new file, To is nil for
// a deleted one, and they differ for a rename.
type FileChange struct {
	From *string
	To   *string
	Type ChangeType
}

// Name returns the path to display: where the file is now, or where it was
// for a deletion.
func (c FileChange) Name() string {
	switch {
	case c.To != nil:
		return *c.To
	case c.From != nil:
		return *c.From
	default:
		return ""
	}
}

// root returns the repository's working tree root — the directory every
// shelled-out git command below needs to run in.
func (r *Repo) root() (string, error) {
	worktree, err := r.repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("getting worktree: %w", err)
	}
	return worktree.Filesystem.Root(), nil
}

// Name returns the repository's name, i.e. the directory its working tree
// sits in.
func (r *Repo) Name() string {
	root, err := r.root()
	if err != nil {
		return ""
	}
	return filepath.Base(root)
}

// Branch returns the checked-out branch's short name, or a short commit
// hash when HEAD is detached. Shells out for the same reason headContents
// does: go-git's Head() can't resolve a linked worktree's HEAD.
func (r *Repo) Branch() string {
	root, err := r.root()
	if err != nil {
		return ""
	}

	if out, err := exec.Command("git", "-C", root, "branch", "--show-current").Output(); err == nil {
		if name := strings.TrimSpace(string(out)); name != "" {
			return name
		}
	}

	out, err := exec.Command("git", "-C", root, "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// ChangedFiles lists the working tree's changes, sorted by path. Goes
// through ChangedFilesAt rather than go-git's own Worktree().Status() —
// go-git misreads a linked worktree's index, so this needs to work the same
// whether r is the primary worktree or a linked one.
func (r *Repo) ChangedFiles() ([]FileChange, error) {
	root, err := r.root()
	if err != nil {
		return nil, err
	}
	return ChangedFilesAt(root)
}

// LineType classifies a single line of a diff.
type LineType int

const (
	LineEqual LineType = iota
	LineAdded
	LineDeleted
)

// DiffLine is one line of a diff, pre-classified so callers can style it.
// Old and New are its 1-based number per side, 0 where it doesn't appear.
type DiffLine struct {
	Type LineType
	Text string
	Old  int
	New  int
}

// FileDiff diffs filePath (relative to the repo root) at HEAD against its
// current content. A side it's missing from counts as empty.
func (r *Repo) FileDiff(filePath string) ([]DiffLine, error) {
	return r.FileDiffAgainst("HEAD", filePath)
}

// FileDiffAgainst is FileDiff against an arbitrary ref instead of HEAD — the
// checkpoint ref, in particular.
func (r *Repo) FileDiffAgainst(ref, filePath string) ([]DiffLine, error) {
	root, err := r.root()
	if err != nil {
		return nil, err
	}

	oldContent, err := contentAt(root, ref, filePath)
	if err != nil {
		return nil, fmt.Errorf("reading %s content: %w", ref, err)
	}

	newContent, err := r.worktreeContents(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading working tree content: %w", err)
	}

	return lineDiff(oldContent, newContent), nil
}

// contentAt returns filePath's content at ref, or "" if it wasn't there
// (including when ref doesn't resolve at all — a repo with no commits yet).
// Shells out rather than resolving the ref through go-git: a linked
// worktree's branch ref lives in the main repo's refs, reached only via the
// "commondir" indirection go-git doesn't follow, so it errors "reference not
// found" for one.
func contentAt(root, ref, filePath string) (string, error) {
	out, err := exec.Command("git", "-C", root, "show", ref+":"+filePath).Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", nil
		}
		return "", err
	}

	return string(out), nil
}

// worktreeContents returns filePath's current content, or "" if it's gone.
func (r *Repo) worktreeContents(filePath string) (string, error) {
	worktree, err := r.repo.Worktree()
	if err != nil {
		return "", err
	}

	file, err := worktree.Filesystem.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// lineDiff computes a line-level diff, classifying each line as added,
// deleted, or unchanged.
func lineDiff(oldContent, newContent string) []DiffLine {
	dmp := diffmatchpatch.New()
	a, b, lines := dmp.DiffLinesToChars(oldContent, newContent)
	diffs := dmp.DiffCharsToLines(dmp.DiffMain(a, b, false), lines)

	// Number each side independently: an added line only advances the new
	// side's counter, a deleted line only the old side's.
	var out []DiffLine
	oldNum, newNum := 0, 0

	for _, d := range diffs {
		lineType := LineEqual
		switch d.Type {
		case diffmatchpatch.DiffInsert:
			lineType = LineAdded
		case diffmatchpatch.DiffDelete:
			lineType = LineDeleted
		}

		text := strings.TrimSuffix(d.Text, "\n")
		for _, line := range strings.Split(text, "\n") {
			diffLine := DiffLine{Type: lineType, Text: line}

			if lineType != LineAdded {
				oldNum++
				diffLine.Old = oldNum
			}
			if lineType != LineDeleted {
				newNum++
				diffLine.New = newNum
			}

			out = append(out, diffLine)
		}
	}

	return out
}
