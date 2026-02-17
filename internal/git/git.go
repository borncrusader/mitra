package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
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

func GetDefaultBranch(repoURL string) (string, error) {
	repoURL = normalizeURL(repoURL)

	cmd := exec.Command("git", "ls-remote", "--symref", repoURL, "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get default branch: %w", err)
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

func IsClean(repoPath string) (bool, error) {
	// Check for uncommitted and untracked changes
	cmd := exec.Command("git", "-C", repoPath, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check git status: %w", err)
	}
	if len(strings.TrimSpace(string(output))) > 0 {
		return false, nil
	}

	// Check for stashed changes
	cmd = exec.Command("git", "-C", repoPath, "stash", "list")
	output, err = cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check git stash: %w", err)
	}
	if len(strings.TrimSpace(string(output))) > 0 {
		return false, nil
	}

	// Get current branch name
	cmd = exec.Command("git", "-C", repoPath, "branch", "--show-current")
	output, err = cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to get current branch: %w", err)
	}
	currentBranch := strings.TrimSpace(string(output))

	// If on main/master, no need to check merge status
	if currentBranch == "main" || currentBranch == "master" {
		return true, nil
	}

	// Check if current branch has been merged to main/master
	// Try main first
	cmd = exec.Command("git", "-C", repoPath, "branch", "--merged", "main")
	output, err = cmd.Output()
	if err != nil {
		// Try master if main doesn't exist
		cmd = exec.Command("git", "-C", repoPath, "branch", "--merged", "master")
		output, err = cmd.Output()
		if err != nil {
			return false, fmt.Errorf("failed to check merge status: %w", err)
		}
	}

	// Check if current branch is in the merged branches list
	mergedBranches := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, branch := range mergedBranches {
		branch = strings.TrimSpace(strings.TrimPrefix(branch, "*"))
		if branch == currentBranch {
			return true, nil
		}
	}

	return false, nil
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
