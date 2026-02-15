package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
)

type DashboardModel struct {
	width    int
	height   int
	quitting bool
}

func NewDashboard() DashboardModel {
	return DashboardModel{}
}

func (m DashboardModel) Init() tea.Cmd {
	return nil
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m DashboardModel) View() string {
	if m.quitting {
		return ""
	}

	lines := []string{
		"███╗   ███╗██╗████████╗██████╗  █████╗ ",
		"████╗ ████║██║╚══██╔══╝██╔══██╗██╔══██╗",
		"██╔████╔██║██║   ██║   ██████╔╝███████║",
		"██║╚██╔╝██║██║   ██║   ██╔══██╗██╔══██║",
		"██║ ╚═╝ ██║██║   ██║   ██║  ██║██║  ██║",
		"╚═╝     ╚═╝╚═╝   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝",
		"",
		"Repos, Worktrees, Branches and Agents",
		"",
		"",
		"Press q to quit",
	}

	contentHeight := len(lines)

	var centered []string
	for _, line := range lines {
		lineWidth := runewidth.StringWidth(line)
		leftPadding := (m.width - lineWidth) / 2
		if leftPadding < 0 {
			leftPadding = 0
		}
		centered = append(centered, strings.Repeat(" ", leftPadding)+line)
	}

	topPadding := (m.height - contentHeight) / 2
	if topPadding < 0 {
		topPadding = 0
	}

	result := strings.Repeat("\n", topPadding) + strings.Join(centered, "\n")
	return result
}

func RunDashboard() error {
	p := tea.NewProgram(NewDashboard(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
