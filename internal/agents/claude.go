package agents

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"mitra/internal/util"
)

type ClaudeConfig struct {
	Projects map[string]ClaudeProjectConfig `json:"projects"`
}

type ClaudeProjectConfig struct {
	HasTrustDialogAccepted bool `json:"hasTrustDialogAccepted"`
}

func EnableTrustForDir(dirPath string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	claudeConfigPath := filepath.Join(homeDir, ".claude.json")

	var config ClaudeConfig

	data, err := os.ReadFile(claudeConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			config = ClaudeConfig{
				Projects: make(map[string]ClaudeProjectConfig),
			}
		} else {
			return fmt.Errorf("failed to read claude config: %w", err)
		}
	} else {
		if err := json.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("failed to parse claude config: %w", err)
		}
	}

	if config.Projects == nil {
		config.Projects = make(map[string]ClaudeProjectConfig)
	}

	absPath, err := filepath.Abs(dirPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	if _, exists := config.Projects[absPath]; !exists {
		config.Projects[absPath] = ClaudeProjectConfig{
			HasTrustDialogAccepted: true,
		}

		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal claude config: %w", err)
		}

		file, err := os.Create(claudeConfigPath)
		if err != nil {
			return fmt.Errorf("failed to create claude config: %w", err)
		}
		defer util.DeferCheck(file.Close)

		if _, err := file.Write(data); err != nil {
			return fmt.Errorf("failed to write claude config: %w", err)
		}
	}

	return nil
}

// SetupFiles symlinks Claude-local files from the main worktree into a newly
// created worktree. Silently skips any file that does not exist in the source.
func SetupFiles(mainWorktreePath, worktreePath string) error {
	files := []string{
		".claude/settings.local.json",
		"CLAUDE.local.md",
	}

	for _, file := range files {
		if err := symlinkFile(mainWorktreePath, worktreePath, file); err != nil {
			return err
		}
	}

	return nil
}

func symlinkFile(srcDir, dstDir, relPath string) error {
	src := filepath.Join(srcDir, relPath)
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil
	}

	dst := filepath.Join(dstDir, relPath)
	if _, err := os.Lstat(dst); err == nil {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	return os.Symlink(src, dst)
}
