package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type DirtyReason string

const (
	DirtyReasonClean              DirtyReason = ""
	DirtyReasonUncommittedChanges DirtyReason = "uncommitted_changes"
	DirtyReasonUntrackedFiles     DirtyReason = "untracked_files"
	DirtyReasonStashedChanges     DirtyReason = "stashed_changes"
	DirtyReasonMergeInProgress    DirtyReason = "merge_in_progress"
	DirtyReasonRebaseInProgress   DirtyReason = "rebase_in_progress"
	DirtyReasonCherryPickInProgress DirtyReason = "cherry_pick_in_progress"
	DirtyReasonRevertInProgress   DirtyReason = "revert_in_progress"
	DirtyReasonBisectInProgress   DirtyReason = "bisect_in_progress"
	DirtyReasonDetachedHead       DirtyReason = "detached_head"
	DirtyReasonUnpushedCommits    DirtyReason = "unpushed_commits"
	DirtyReasonUnmergedBranches   DirtyReason = "unmerged_branches"
	DirtyReasonCurrentBranchNotMerged DirtyReason = "current_branch_not_merged"
)

func normalizeURL(repoURL string) string {
	if strings.HasPrefix(repoURL, "git@") {
		return repoURL
	}

	if !strings.HasPrefix(repoURL, "http://") && !strings.HasPrefix(repoURL, "https://") {
		repoURL = "https://" + repoURL
	}

	if !strings.HasSuffix(repoURL, ".git") {
		repoURL = repoURL + ".git"
	}

	return repoURL
}

func GetMainBranch(repoURL string) (string, error) {
	repoURL = normalizeURL(repoURL)

	cmd := exec.Command("git", "ls-remote", "--symref", repoURL, "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get main branch: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "ref:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				refParts := strings.Split(parts[1], "/")
				if len(refParts) >= 3 {
					return refParts[len(refParts)-1], nil
				}
			}
		}
	}

	return "main", nil
}

func Clone(repoURL, targetDir, branch string) error {
	repoURL = normalizeURL(repoURL)

	cmd := exec.Command("git", "clone", "--branch", branch, repoURL, targetDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func IsClean(repoPath string, mainBranch string) (bool, DirtyReason, error) {
	// Check for uncommitted and untracked changes
	cmd := exec.Command("git", "-C", repoPath, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false, DirtyReasonClean, fmt.Errorf("failed to check git status: %w", err)
	}
	if len(strings.TrimSpace(string(output))) > 0 {
		return false, DirtyReasonUncommittedChanges, nil
	}

	// Check for stashed changes
	cmd = exec.Command("git", "-C", repoPath, "stash", "list")
	output, err = cmd.Output()
	if err != nil {
		return false, DirtyReasonClean, fmt.Errorf("failed to check git stash: %w", err)
	}
	if len(strings.TrimSpace(string(output))) > 0 {
		return false, DirtyReasonStashedChanges, nil
	}

	// Check for merge in progress
	gitDir := fmt.Sprintf("%s/.git", repoPath)
	if _, err := os.Stat(fmt.Sprintf("%s/MERGE_HEAD", gitDir)); err == nil {
		return false, DirtyReasonMergeInProgress, nil
	}

	// Check for rebase in progress
	if _, err := os.Stat(fmt.Sprintf("%s/rebase-merge", gitDir)); err == nil {
		return false, DirtyReasonRebaseInProgress, nil
	}
	if _, err := os.Stat(fmt.Sprintf("%s/rebase-apply", gitDir)); err == nil {
		return false, DirtyReasonRebaseInProgress, nil
	}

	// Check for cherry-pick in progress
	if _, err := os.Stat(fmt.Sprintf("%s/CHERRY_PICK_HEAD", gitDir)); err == nil {
		return false, DirtyReasonCherryPickInProgress, nil
	}

	// Check for revert in progress
	if _, err := os.Stat(fmt.Sprintf("%s/REVERT_HEAD", gitDir)); err == nil {
		return false, DirtyReasonRevertInProgress, nil
	}

	// Check for bisect in progress
	if _, err := os.Stat(fmt.Sprintf("%s/BISECT_LOG", gitDir)); err == nil {
		return false, DirtyReasonBisectInProgress, nil
	}

	// Get current branch name
	cmd = exec.Command("git", "-C", repoPath, "branch", "--show-current")
	output, err = cmd.Output()
	if err != nil {
		return false, DirtyReasonClean, fmt.Errorf("failed to get current branch: %w", err)
	}
	currentBranch := strings.TrimSpace(string(output))

	// Check for detached HEAD
	if currentBranch == "" {
		return false, DirtyReasonDetachedHead, nil
	}

	// Check for unpushed commits on current branch
	cmd = exec.Command("git", "-C", repoPath, "rev-list", "--count", "@{u}..HEAD")
	output, err = cmd.Output()
	if err == nil {
		unpushedCount := strings.TrimSpace(string(output))
		if unpushedCount != "0" && unpushedCount != "" {
			return false, DirtyReasonUnpushedCommits, nil
		}
	}

	// Check for unmerged local branches
	cmd = exec.Command("git", "-C", repoPath, "branch", "--no-merged", mainBranch)
	output, err = cmd.Output()
	if err != nil {
		return false, DirtyReasonClean, fmt.Errorf("failed to check unmerged branches: %w", err)
	}
	unmergedBranches := strings.TrimSpace(string(output))
	if len(unmergedBranches) > 0 {
		return false, DirtyReasonUnmergedBranches, nil
	}

	// If not on main branch, check if current branch has been merged
	if currentBranch != mainBranch {
		cmd = exec.Command("git", "-C", repoPath, "branch", "--merged", mainBranch)
		output, err = cmd.Output()
		if err != nil {
			return false, DirtyReasonClean, fmt.Errorf("failed to check merge status: %w", err)
		}

		// Check if current branch is in the merged branches list
		mergedBranches := strings.Split(strings.TrimSpace(string(output)), "\n")
		isMerged := false
		for _, branch := range mergedBranches {
			branch = strings.TrimSpace(strings.TrimPrefix(branch, "*"))
			if branch == currentBranch {
				isMerged = true
				break
			}
		}

		if !isMerged {
			return false, DirtyReasonCurrentBranchNotMerged, nil
		}
	}

	return true, DirtyReasonClean, nil
}

func Pull(repoPath string) error {
	cmd := exec.Command("git", "-C", repoPath, "pull")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to pull: %w, output: %s", err, string(output))
	}

	return nil
}

func CreateWorktree(mainWorktreePath, branch, worktreePath, parentBranch string) error {
	cmd := exec.Command("git", "-C", mainWorktreePath, "worktree", "add", "-b", branch, worktreePath, parentBranch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create git worktree: %w, output: %s", err, string(output))
	}

	return nil
}

func RemoveWorktree(mainWorktreePath, worktreePath string) error {
	cmd := exec.Command("git", "-C", mainWorktreePath, "worktree", "remove", worktreePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove git worktree: %w, output: %s", err, string(output))
	}

	return nil
}
