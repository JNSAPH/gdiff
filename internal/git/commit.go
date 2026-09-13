package git

import (
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Commit is one entry from the log.
type Commit struct {
	Hash    string // full hash; every command below is given this, not Short
	Short   string
	Subject string
	Author  string
	When    time.Time
}

// commitFormat is one tab-separated line per commit, the subject last so it
// can hold a tab of its own.
const commitFormat = "%H\t%h\t%an\t%ct\t%s"

// Commits lists the newest limit commits reachable from HEAD, newest first.
// Shells out for the same reason Branch() does: go-git can't resolve a linked
// worktree's HEAD.
func (r *Repo) Commits(limit int) ([]Commit, error) {
	root, err := r.root()
	if err != nil {
		return nil, err
	}

	out, err := exec.Command("git", "-C", root, "log",
		"--format="+commitFormat, "-n", strconv.Itoa(limit)).Output()
	if err != nil {
		return nil, fmt.Errorf("listing commits: %w", err)
	}

	return parseCommits(string(out)), nil
}

// CommitFiles lists what one commit changed against its first parent. `git
// show` is what handles the two shapes `git diff <hash>^ <hash>` can't: a
// merge, which reads as its first-parent diff, and the root commit, which has
// no parent and reads as all-new.
func (r *Repo) CommitFiles(hash string) ([]FileChange, error) {
	root, err := r.root()
	if err != nil {
		return nil, err
	}

	out, err := exec.Command("git", "-C", root, "show",
		"--format=", "--name-status", "--first-parent", hash).Output()
	if err != nil {
		return nil, fmt.Errorf("listing files in %s: %w", hash, err)
	}

	return parseNameStatus(string(out)), nil
}

// FileDiffIn diffs one file the way the commit changed it: its content at the
// first parent against its content at the commit. Neither side is the working
// tree, and a rename reads its two paths separately.
func (r *Repo) FileDiffIn(hash string, change FileChange) ([]DiffLine, error) {
	root, err := r.root()
	if err != nil {
		return nil, err
	}

	// A nil side is a file the commit added or deleted; the root commit's
	// missing parent lands here too, since contentAt reads an unresolvable
	// ref as empty.
	var oldContent, newContent string
	if change.From != nil {
		if oldContent, err = contentAt(root, hash+"^", *change.From); err != nil {
			return nil, fmt.Errorf("reading parent content: %w", err)
		}
	}
	if change.To != nil {
		if newContent, err = contentAt(root, hash, *change.To); err != nil {
			return nil, fmt.Errorf("reading %s content: %w", hash, err)
		}
	}

	return lineDiff(oldContent, newContent), nil
}

// parseCommits reads commitFormat's lines by field.
func parseCommits(out string) []Commit {
	var commits []Commit

	for line := range strings.SplitSeq(out, "\n") {
		fields := strings.SplitN(line, "\t", 5)
		if len(fields) < 5 {
			continue
		}

		commits = append(commits, Commit{
			Hash:    fields[0],
			Short:   fields[1],
			Author:  fields[2],
			When:    parseUnix(fields[3]),
			Subject: fields[4],
		})
	}

	return commits
}

// parseNameStatus reads `git show --name-status`: a status letter, a tab, then
// the path — a rename carrying its old and new path in two separate fields.
func parseNameStatus(out string) []FileChange {
	var changes []FileChange

	for line := range strings.SplitSeq(out, "\n") {
		fields := strings.Split(line, "\t")
		if len(fields) < 2 || fields[0] == "" {
			continue
		}

		path := fields[1]
		from, to := path, path
		change := FileChange{From: &from, To: &to}

		switch fields[0][0] { // "R100" and "M" both key off the first letter
		case 'A':
			change.From = nil
			change.Type = ChangeTypeNew
		case 'D':
			change.To = nil
			change.Type = ChangeTypeDeleted
		case 'R':
			if len(fields) > 2 {
				newPath := fields[2]
				change.To = &newPath
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
