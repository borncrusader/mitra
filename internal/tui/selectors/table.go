package selectors

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

// TableSelectorModel is a generic table selector for any data
type TableSelectorModel struct {
	table       table.Model
	selectedIdx int
	quitting    bool
	cancelled   bool
	promptText  string
}

// NewTableSelector creates a new generic table selector
func NewTableSelector(columns []table.Column, rows []table.Row, promptText string) TableSelectorModel {
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(len(rows)+1),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return TableSelectorModel{
		table:       t,
		selectedIdx: -1,
		promptText:  promptText,
	}
}

func (m TableSelectorModel) Init() tea.Cmd {
	return nil
}

func (m TableSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.cancelled = true
			m.quitting = true
			return m, tea.Quit

		case "enter":
			m.selectedIdx = m.table.Cursor()
			m.quitting = true
			return m, tea.Quit
		}
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m TableSelectorModel) View() string {
	if m.quitting {
		return ""
	}

	prompt := ""
	if m.promptText != "" {
		prompt = m.promptText + "\n\n"
	}

	return prompt + baseStyle.Render(m.table.View()) + "\n\nPress ↑/↓ to navigate • enter to select • q to quit\n"
}

// SelectedIndex returns the selected row index, or -1 if cancelled
func (m TableSelectorModel) SelectedIndex() int {
	if m.cancelled {
		return -1
	}
	return m.selectedIdx
}

// RunTableSelector runs a table selector and returns the selected index
func RunTableSelector(columns []table.Column, rows []table.Row, promptText string) (int, error) {
	if len(rows) == 0 {
		return -1, nil
	}

	p := tea.NewProgram(NewTableSelector(columns, rows, promptText))
	model, err := p.Run()
	if err != nil {
		return -1, err
	}

	m := model.(TableSelectorModel)
	return m.SelectedIndex(), nil
}

// RenderTable renders a non-interactive table for display
func RenderTable(columns []table.Column, rows []table.Row) string {
	if len(rows) == 0 {
		return "No items found.\n"
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(false),
		// +1 to account for the header
		table.WithHeight(len(rows)+1),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	t.SetStyles(s)

	return baseStyle.Render(t.View()) + "\n"
}
