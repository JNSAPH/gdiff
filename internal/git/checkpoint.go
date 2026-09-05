package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CheckpointRef is the ref gdiff advances each time a checkpoint is
// accepted. It's never reachable from any branch, so an accepted checkpoint
// never shows up in `git log` on a branch and a plain `git push` never sends
// it anywhere — that's the whole point.
const CheckpointRef = "refs/gdiff/checkpoint"

// Checkpoint returns the commit hash the checkpoint ref currently points at,
// and whether one has been set for this repo yet.
func (r *Repo) Checkpoint() (hash string, ok bool, err error) {
	root, err := r.root()
	if err != nil {
		return "", false, err
	}

	out, err := exec.Command("git", "-C", root, "rev-parse", "--verify", "-q", CheckpointRef).Output()
	if err != nil {
		return "", false, nil // not set yet — not an error
	}

	return strings.TrimSpace(string(out)), true, nil
}

// EnsureCheckpoint keeps the checkpoint ref in sync with HEAD: creating it
// pointed at HEAD the first time gdiff sees this repo, and fast-forwarding
// it whenever a real commit has moved HEAD somewhere the checkpoint chain
// doesn't already account for. Checkpoint-only usage (accept/reject) keeps
// HEAD an ancestor of the checkpoint the whole time — HEAD never moves — so
// this only ever does something because of an actual `git commit` made
// outside gdiff. That's the point: committing for real counts as accepting,
// the same as pressing y would. Safe to call any time.
func (r *Repo) EnsureCheckpoint() error {
	root, err := r.root()
	if err != nil {
		return err
	}

	headOut, err := exec.Command("git", "-C", root, "rev-parse", "HEAD").Output()
	if err != nil {
		return fmt.Errorf("resolving HEAD: %w", err)
	}
	head := strings.TrimSpace(string(headOut))

	checkpoint, ok, err := r.Checkpoint()
	if err != nil {
		return err
	}

	caughtUp := ok && exec.Command("git", "-C", root, "merge-base", "--is-ancestor", head, checkpoint).Run() == nil
	if caughtUp {
		return nil
	}

	return exec.Command("git", "-C", root, "update-ref", CheckpointRef, head).Run()
}

// AcceptCheckpoint snapshots the current working tree — including
// untracked files, respecting .gitignore — into a new commit parented on
// the current checkpoint, and moves the checkpoint ref to it. Doesn't touch
// HEAD or the current branch: the new commit is only ever reachable through
// CheckpointRef.
func (r *Repo) AcceptCheckpoint() error {
	if err := r.EnsureCheckpoint(); err != nil {
		return err
	}
	parent, _, err := r.Checkpoint()
	if err != nil {
		return err
	}

	root, err := r.root()
	if err != nil {
		return err
	}

	tmpIndex, err := tempIndexPath()
	if err != nil {
		return err
	}
	defer os.Remove(tmpIndex)
	env := append(os.Environ(), "GIT_INDEX_FILE="+tmpIndex)

	// A fresh index plus `add -A` stages exactly what's on disk right now —
	// a full snapshot, not a diff — so there's no need to seed it first.
	add := exec.Command("git", "-C", root, "add", "-A")
	add.Env = env
	if out, err := add.CombinedOutput(); err != nil {
		return fmt.Errorf("staging working tree: %w: %s", err, out)
	}

	writeTree := exec.Command("git", "-C", root, "write-tree")
	writeTree.Env = env
	treeOut, err := writeTree.Output()
	if err != nil {
		return fmt.Errorf("writing tree: %w", err)
	}
	tree := strings.TrimSpace(string(treeOut))

	commitOut, err := exec.Command("git", "-C", root, "commit-tree", tree, "-p", parent, "-m", "gdiff checkpoint").Output()
	if err != nil {
		return fmt.Errorf("creating checkpoint commit: %w", err)
	}
	commit := strings.TrimSpace(string(commitOut))

	return exec.Command("git", "-C", root, "update-ref", CheckpointRef, commit).Run()
}

// AcceptCheckpointFile is AcceptCheckpoint narrowed to a single file: only
// change's current content moves into the new checkpoint commit — every
// other pending file keeps exactly the content the checkpoint already has.
func (r *Repo) AcceptCheckpointFile(change FileChange) error {
	if err := r.EnsureCheckpoint(); err != nil {
		return err
	}
	parent, _, err := r.Checkpoint()
	if err != nil {
		return err
	}

	root, err := r.root()
	if err != nil {
		return err
	}

	tmpIndex, err := tempIndexPath()
	if err != nil {
		return err
	}
	defer os.Remove(tmpIndex)
	env := append(os.Environ(), "GIT_INDEX_FILE="+tmpIndex)

	// Seed from the checkpoint's own tree, not an empty index — every path
	// besides this one needs to keep the content the checkpoint already
	// recorded, not disappear.
	readTree := exec.Command("git", "-C", root, "read-tree", parent)
	readTree.Env = env
	if out, err := readTree.CombinedOutput(); err != nil {
		return fmt.Errorf("seeding index from checkpoint: %w: %s", err, out)
	}

	// A rename or deletion means the old path shouldn't survive into the
	// new checkpoint tree.
	if change.From != nil && (change.To == nil || *change.From != *change.To) {
		rm := exec.Command("git", "-C", root, "rm", "--cached", "--ignore-unmatch", "--", *change.From)
		rm.Env = env
		if out, err := rm.CombinedOutput(); err != nil {
			return fmt.Errorf("removing %s: %w: %s", *change.From, err, out)
		}
	}

	// A deletion has no current path to stage; anything else does.
	if change.To != nil {
		add := exec.Command("git", "-C", root, "add", "--", *change.To)
		add.Env = env
		if out, err := add.CombinedOutput(); err != nil {
			return fmt.Errorf("staging %s: %w: %s", *change.To, err, out)
		}
	}

	writeTree := exec.Command("git", "-C", root, "write-tree")
	writeTree.Env = env
	treeOut, err := writeTree.Output()
	if err != nil {
		return fmt.Errorf("writing tree: %w", err)
	}
	tree := strings.TrimSpace(string(treeOut))

	commitOut, err := exec.Command("git", "-C", root, "commit-tree", tree, "-p", parent, "-m", "gdiff checkpoint").Output()
	if err != nil {
		return fmt.Errorf("creating checkpoint commit: %w", err)
	}
	commit := strings.TrimSpace(string(commitOut))

	return exec.Command("git", "-C", root, "update-ref", CheckpointRef, commit).Run()
}

// RejectCheckpointFile restores change's path(s) to what the checkpoint has
// — recreating a deleted-since-checkpoint file, reverting a modified one, or
// removing one the checkpoint never had (a new file, or the new side of a
// rename). Destructive, same as RejectCheckpoint, just narrower.
func (r *Repo) RejectCheckpointFile(change FileChange) error {
	checkpoint, ok, err := r.Checkpoint()
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("no checkpoint set for this repo")
	}

	root, err := r.root()
	if err != nil {
		return err
	}

	paths := map[string]bool{}
	if change.From != nil {
		paths[*change.From] = true
	}
	if change.To != nil {
		paths[*change.To] = true
	}

	var toUnstage []string
	for path := range paths {
		inCheckpoint := exec.Command("git", "-C", root, "cat-file", "-e", checkpoint+":"+path).Run() == nil

		if inCheckpoint {
			checkout := exec.Command("git", "-C", root, "checkout", checkpoint, "--", path)
			if out, err := checkout.CombinedOutput(); err != nil {
				return fmt.Errorf("restoring %s: %w: %s", path, err, out)
			}
			toUnstage = append(toUnstage, path)
			continue
		}

		if err := os.Remove(filepath.Join(root, path)); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("removing %s: %w", path, err)
		}
	}

	if len(toUnstage) == 0 {
		return nil
	}

	// `checkout <ref> -- <path>` stages what it restores; unstage so the
	// result reads as ordinary uncommitted changes, matching
	// RejectCheckpoint's whole-tree behavior.
	args := append([]string{"-C", root, "reset", "--"}, toUnstage...)
	if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
		return fmt.Errorf("unstaging: %w: %s", err, out)
	}

	return nil
}

// RejectCheckpoint resets the working tree back to the checkpoint,
// discarding everything since. Destructive: anything not committed
// somewhere else is gone once this runs — that's exactly what an accepted
// checkpoint is, a snapshot that was never committed anywhere real.
func (r *Repo) RejectCheckpoint() error {
	checkpoint, ok, err := r.Checkpoint()
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("no checkpoint set for this repo")
	}

	root, err := r.root()
	if err != nil {
		return err
	}

	// Resets the index and working tree to the checkpoint's tree without
	// moving HEAD or the branch. `git reset --hard` would move the branch
	// pointer too, which is exactly the leak into real history this
	// mechanism exists to avoid.
	readTree := exec.Command("git", "-C", root, "read-tree", "--reset", "-u", checkpoint)
	if out, err := readTree.CombinedOutput(); err != nil {
		return fmt.Errorf("restoring checkpoint: %w: %s", err, out)
	}

	if out, err := exec.Command("git", "-C", root, "clean", "-fd").CombinedOutput(); err != nil {
		return fmt.Errorf("removing files added since checkpoint: %w: %s", err, out)
	}

	// read-tree --reset leaves everything staged relative to HEAD; unstage
	// it so the result reads as ordinary uncommitted changes, not a
	// pre-staged commit waiting to happen.
	if out, err := exec.Command("git", "-C", root, "reset").CombinedOutput(); err != nil {
		return fmt.Errorf("unstaging: %w: %s", err, out)
	}

	return nil
}

// tempIndexPath returns a path for a scratch git index that doesn't exist
// yet. Git creates it fresh the first time it's used with GIT_INDEX_FILE,
// but errors on an existing zero-length file, so this can't just use
// os.CreateTemp's file directly.
func tempIndexPath() (string, error) {
	f, err := os.CreateTemp("", "gdiff-index-*")
	if err != nil {
		return "", err
	}
	path := f.Name()
	f.Close()
	os.Remove(path)
	return path, nil
}
