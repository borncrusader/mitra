package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type testModel struct {
	width  int
	height int
}

func (m testModel) Init() tea.Cmd {
	return nil
}

func (m testModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m testModel) View() string {
	if m.width == 0 {
		return ""
	}

	style := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	content := "Hello, World!\n\n" + lipgloss.NewStyle().Faint(true).Render("press q to quit")

	lines := strings.Split(style.Render(content), "\n")
	if len(lines) > m.height {
		lines = lines[:m.height]
	}
	return strings.Join(lines, "\n")
}

func RunTest() error {
	_, err := tea.NewProgram(testModel{}, tea.WithAltScreen()).Run()
	return err
}

// ensure testModel satisfies tea.Model at compile time
var _ tea.Model = testModel{}
