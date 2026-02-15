package tmux

import (
	"testing"
)

func TestParsePaneConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      string
		expected    *PaneConfig
		expectError bool
	}{
		{
			name:   "valid config with window 0 pane 0",
			config: "0.0:nvim",
			expected: &PaneConfig{
				Window:  0,
				Pane:    0,
				Command: "nvim",
			},
			expectError: false,
		},
		{
			name:   "valid config with window 0 pane 1",
			config: "0.1:git status",
			expected: &PaneConfig{
				Window:  0,
				Pane:    1,
				Command: "git status",
			},
			expectError: false,
		},
		{
			name:   "valid config with multi-word command",
			config: "0.0:make build && make test",
			expected: &PaneConfig{
				Window:  0,
				Pane:    0,
				Command: "make build && make test",
			},
			expectError: false,
		},
		{
			name:        "invalid config - no colon",
			config:      "0.0nvim",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "invalid config - no dot",
			config:      "00:nvim",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "invalid config - pane number > 1",
			config:      "0.2:nvim",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "invalid config - negative pane number",
			config:      "0.-1:nvim",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "invalid config - non-numeric window",
			config:      "a.0:nvim",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "invalid config - non-numeric pane",
			config:      "0.a:nvim",
			expected:    nil,
			expectError: true,
		},
		{
			name:   "valid config with empty command",
			config: "0.0:",
			expected: &PaneConfig{
				Window:  0,
				Pane:    0,
				Command: "",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParsePaneConfig(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result.Window != tt.expected.Window {
				t.Errorf("expected window %d, got %d", tt.expected.Window, result.Window)
			}
			if result.Pane != tt.expected.Pane {
				t.Errorf("expected pane %d, got %d", tt.expected.Pane, result.Pane)
			}
			if result.Command != tt.expected.Command {
				t.Errorf("expected command %q, got %q", tt.expected.Command, result.Command)
			}
		})
	}
}
