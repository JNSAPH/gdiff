package git

import (
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Branch is one branch by name, merging the local ref with any remote one.
type Branch struct {
	Name     string // short name, never carrying a remote prefix
	Current  bool
	Local    bool
	Remote   string // a remote that has it, empty when none does
	Upstream string
	Ahead    int
	Behind   int
	Updated  time.Time // the branch tip's commit date
}

// branchFormat is one tab-separated line per ref, "*" marking the current.
const branchFormat = "%(refname)\t%(upstream:short)\t%(upstream:track)\t%(committerdate:unix)\t%(HEAD)"

// Branches lists local and remote branches, current first then newest. go-git
// knows nothing about upstreams or ahead/behind, so this shells out.
func (r *Repo) Branches() ([]Branch, error) {
	root, err := r.root()
	if err != nil {
		return nil, err
	}

	out, err := exec.Command("git", "-C", root, "for-each-ref",
		"--format="+branchFormat, "refs/heads", "refs/remotes").Output()
	if err != nil {
		return nil, fmt.Errorf("listing branches: %w", err)
	}

	return parseBranches(string(out)), nil
}

// Checkout switches to name, which git resolves to a local or tracking branch.
func (r *Repo) Checkout(name string) error {
	root, err := r.root()
	if err != nil {
		return err
	}

	if out, err := exec.Command("git", "-C", root, "switch", "--", name).CombinedOutput(); err != nil {
		return fmt.Errorf("switching to %s: %w: %s", name, err, strings.TrimSpace(string(out)))
	}

	return nil
}

// parseBranches folds refs/heads/x and refs/remotes/*/x into one entry.
func parseBranches(out string) []Branch {
	byName := map[string]*Branch{}
	var order []string

	entry := func(name string) *Branch {
		if b, ok := byName[name]; ok {
			return b
		}
		b := &Branch{Name: name}
		byName[name] = b
		order = append(order, name)
		return b
	}

	for line := range strings.SplitSeq(out, "\n") {
		fields := strings.Split(line, "\t")
		if len(fields) < 5 {
			continue
		}
		ref, upstream, track, updated, head := fields[0], fields[1], fields[2], fields[3], fields[4]

		switch {
		case strings.HasPrefix(ref, "refs/heads/"):
			b := entry(strings.TrimPrefix(ref, "refs/heads/"))
			b.Local = true
			b.Current = head == "*"
			b.Upstream = upstream
			b.Ahead, b.Behind = parseTrack(track)
			b.Updated = parseUnix(updated)

		case strings.HasPrefix(ref, "refs/remotes/"):
			remote, name, ok := strings.Cut(strings.TrimPrefix(ref, "refs/remotes/"), "/")
			if !ok || name == "HEAD" {
				continue // origin/HEAD is a symbolic ref, not a branch
			}
			b := entry(name)
			b.Remote = remote
			if !b.Local {
				b.Updated = parseUnix(updated)
			}
		}
	}

	branches := make([]Branch, 0, len(order))
	for _, name := range order {
		branches = append(branches, *byName[name])
	}

	sort.SliceStable(branches, func(i, j int) bool {
		if branches[i].Current != branches[j].Current {
			return branches[i].Current
		}
		return branches[i].Updated.After(branches[j].Updated)
	})

	return branches
}

// parseTrack reads for-each-ref's "[ahead 3, behind 2]" upstream summary.
func parseTrack(track string) (ahead, behind int) {
	for part := range strings.SplitSeq(strings.Trim(track, "[]"), ", ") {
		kind, count, ok := strings.Cut(part, " ")
		if !ok {
			continue
		}
		n, err := strconv.Atoi(count)
		if err != nil {
			continue
		}
		switch kind {
		case "ahead":
			ahead = n
		case "behind":
			behind = n
		}
	}
	return ahead, behind
}

// parseUnix reads a unix timestamp, returning the zero time if it can't.
func parseUnix(s string) time.Time {
	secs, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return time.Time{}
	}
	return time.Unix(secs, 0)
}
