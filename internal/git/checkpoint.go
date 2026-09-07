package git

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// checkpointRefPrefix namespaces the refs gdiff advances on accept. Nothing
// under it is reachable from a branch, so git log and git push never see it.
const checkpointRefPrefix = "refs/gdiff/checkpoint/"

// legacyCheckpointRef held the one shared checkpoint before per-branch refs.
const legacyCheckpointRef = "refs/gdiff/checkpoint"

// CheckpointRef is the current branch's checkpoint ref. One per branch: a
// shared one makes the branch you left the baseline for the one you're on.
func (r *Repo) CheckpointRef() (string, error) {
	branch := r.Branch()
	if branch == "" {
		return "", errors.New("resolving the current branch")
	}
	return checkpointRefPrefix + branch, nil
}

// Checkpoint returns the commit the checkpoint ref points at, if it's set.
func (r *Repo) Checkpoint() (hash string, ok bool, err error) {
	root, err := r.root()
	if err != nil {
		return "", false, err
	}

	ref, err := r.CheckpointRef()
	if err != nil {
		return "", false, err
	}

	out, err := exec.Command("git", "-C", root, "rev-parse", "--verify", "-q", ref).Output()
	if err != nil {
		return "", false, nil // not set yet — not an error
	}

	return strings.TrimSpace(string(out)), true, nil
}

// EnsureCheckpoint keeps the checkpoint ref in sync with HEAD: created at
// HEAD the first time, fast-forwarded when a real commit moved HEAD. Only a
// commit made outside gdiff trips it — accept/reject never move HEAD — so
// committing for real counts as accepting. Safe to call any time.
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

	ref, err := r.CheckpointRef()
	if err != nil {
		return err
	}

	// A ref at legacyCheckpointRef blocks every ref beneath it. Dropping it
	// costs that repo one baseline, re-made at HEAD on the next line.
	_ = exec.Command("git", "-C", root, "update-ref", "-d", legacyCheckpointRef).Run()

	return exec.Command("git", "-C", root, "update-ref", ref, head).Run()
}

// AcceptCheckpoint snapshots the working tree — untracked files included,
// .gitignore respected — into a commit parented on the current checkpoint.
// HEAD and the branch don't move: only the ref reaches the new commit.
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

	// A fresh index plus `add -A` stages a full snapshot, not a diff.
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

	ref, err := r.CheckpointRef()
	if err != nil {
		return err
	}

	return exec.Command("git", "-C", root, "update-ref", ref, commit).Run()
}

// AcceptCheckpointFile is AcceptCheckpoint for one file: every other pending
// file keeps the content the checkpoint already has.
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

	// Seed from the checkpoint's tree, not an empty index — every other path
	// has to keep its checkpointed content rather than disappear.
	readTree := exec.Command("git", "-C", root, "read-tree", parent)
	readTree.Env = env
	if out, err := readTree.CombinedOutput(); err != nil {
		return fmt.Errorf("seeding index from checkpoint: %w: %s", err, out)
	}

	// A rename or deletion shouldn't leave the old path in the new tree.
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

	ref, err := r.CheckpointRef()
	if err != nil {
		return err
	}

	return exec.Command("git", "-C", root, "update-ref", ref, commit).Run()
}

// RejectCheckpointFile restores one path to what the checkpoint has, or drops
// it if the checkpoint never had it. Destructive, like RejectCheckpoint.
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

	// `checkout <ref> -- <path>` stages what it restores; unstage so the result
	// reads as ordinary uncommitted changes.
	args := append([]string{"-C", root, "reset", "--"}, toUnstage...)
	if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
		return fmt.Errorf("unstaging: %w: %s", err, out)
	}

	return nil
}

// RejectCheckpoint resets the working tree to the checkpoint. Destructive: an
// accepted checkpoint was never committed anywhere real, so this is final.
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

	// Resets index and tree without moving HEAD or the branch — `reset --hard`
	// would move the branch pointer, the leak into real history this avoids.
	readTree := exec.Command("git", "-C", root, "read-tree", "--reset", "-u", checkpoint)
	if out, err := readTree.CombinedOutput(); err != nil {
		return fmt.Errorf("restoring checkpoint: %w: %s", err, out)
	}

	if out, err := exec.Command("git", "-C", root, "clean", "-fd").CombinedOutput(); err != nil {
		return fmt.Errorf("removing files added since checkpoint: %w: %s", err, out)
	}

	// read-tree --reset leaves everything staged; unstage so the result reads
	// as ordinary uncommitted changes.
	if out, err := exec.Command("git", "-C", root, "reset").CombinedOutput(); err != nil {
		return fmt.Errorf("unstaging: %w: %s", err, out)
	}

	return nil
}

// tempIndexPath returns a path for a scratch index that doesn't exist yet:
// git errors on an existing zero-length file, so os.CreateTemp won't do.
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
