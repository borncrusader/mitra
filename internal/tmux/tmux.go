package tmux

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"
)

// PaneConfig represents a pane configuration
type PaneConfig struct {
	Window  int
	Pane    int
	Command string
}

// ParsePaneConfig parses a pane configuration string in format "<window>.<pane>:<command>"
func ParsePaneConfig(config string) (*PaneConfig, error) {
	parts := strings.SplitN(config, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid pane config format: %s (expected <window>.<pane>:<command>)", config)
	}

	locationParts := strings.Split(parts[0], ".")
	if len(locationParts) != 2 {
		return nil, fmt.Errorf("invalid pane location format: %s (expected <window>.<pane>)", parts[0])
	}

	var window, pane int
	if _, err := fmt.Sscanf(locationParts[0], "%d", &window); err != nil {
		return nil, fmt.Errorf("invalid window number: %s", locationParts[0])
	}
	if _, err := fmt.Sscanf(locationParts[1], "%d", &pane); err != nil {
		return nil, fmt.Errorf("invalid pane number: %s", locationParts[1])
	}

	if pane < 0 || pane > 1 {
		return nil, fmt.Errorf("pane number must be 0 or 1, got %d", pane)
	}

	if window < 0 {
		return nil, fmt.Errorf("window number must be non-negative, got %d", window)
	}

	return &PaneConfig{
		Window:  window,
		Pane:    pane,
		Command: strings.TrimSpace(parts[1]),
	}, nil
}

// SessionExists checks if a tmux session with the given name exists
func SessionExists(name string) (bool, error) {
	cmd := exec.Command("tmux", "has-session", "-t", name)
	err := cmd.Run()
	if err != nil {
		// Exit code 1 means session doesn't exist
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, fmt.Errorf("failed to check tmux session: %w", err)
	}
	return true, nil
}

// CreateSession creates a new tmux session with the given name, working directory, and pane configurations
func CreateSession(name, workingDir string, paneConfigs []string) error {
	cmd := exec.Command("tmux", "new-session", "-d", "-s", name, "-c", workingDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create tmux session: %w", err)
	}

	if len(paneConfigs) == 0 {
		return nil
	}

	panes := make(map[int]map[int]*PaneConfig)
	for _, configStr := range paneConfigs {
		paneConfig, err := ParsePaneConfig(configStr)
		if err != nil {
			return fmt.Errorf("failed to parse pane config: %w", err)
		}

		if panes[paneConfig.Window] == nil {
			panes[paneConfig.Window] = make(map[int]*PaneConfig)
		}
		panes[paneConfig.Window][paneConfig.Pane] = paneConfig
	}

	for window := 0; window <= 0; window++ {
		windowPanes := panes[window]
		if windowPanes == nil {
			continue
		}

		target := fmt.Sprintf("%s:%d", name, window)

		if pane0, ok := windowPanes[0]; ok && pane0.Command != "" {
			if err := SendKeys(fmt.Sprintf("%s.0", target), pane0.Command); err != nil {
				return err
			}
		}

		if pane1, ok := windowPanes[1]; ok {
			if err := SplitWindow(target, workingDir); err != nil {
				return err
			}

			if pane1.Command != "" {
				if err := SendKeys(fmt.Sprintf("%s.1", target), pane1.Command); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// ListSessions returns a list of all tmux session names
func ListSessions() ([]string, error) {
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}")
	output, err := cmd.Output()
	if err != nil {
		// No sessions is not an error
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to list tmux sessions: %w", err)
	}

	sessions := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(sessions) == 1 && sessions[0] == "" {
		return []string{}, nil
	}

	return sessions, nil
}

// KillSession kills a tmux session with the given name
func KillSession(name string) error {
	cmd := exec.Command("tmux", "kill-session", "-t", name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to kill tmux session: %w", err)
	}
	return nil
}

// AttachSession attaches to a tmux session with the given name
func AttachSession(name string) error {
	cmd := exec.Command("tmux", "attach-session", "-t", name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to attach to tmux session: %w", err)
	}
	return nil
}

// AttachSessionExec attaches to a tmux session by replacing the current process
// This is useful for CLI commands where you want to "become" the tmux attach process
func AttachSessionExec(name string) error {
	return syscall.Exec("/usr/bin/tmux", []string{"tmux", "attach-session", "-t", name}, syscall.Environ())
}

// SplitWindow splits a tmux window horizontally
func SplitWindow(target, workingDir string) error {
	cmd := exec.Command("tmux", "split-window", "-h", "-t", target, "-c", workingDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to split window: %w", err)
	}
	return nil
}

// SendKeys sends keys/commands to a specific pane
func SendKeys(target, command string) error {
	cmd := exec.Command("tmux", "send-keys", "-t", target, command, "Enter")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send keys: %w", err)
	}
	return nil
}
