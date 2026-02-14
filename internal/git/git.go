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
