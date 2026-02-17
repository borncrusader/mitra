package selectors

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"mitra/internal/proto"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type WorktreeSelectorModel struct {
	table     table.Model
	worktrees []*proto.Worktree
	repos     map[string]*proto.Repo
	selected  *proto.Worktree
	quitting  bool
}

func NewWorktreeSelector(worktrees []*proto.Worktree, repos []*proto.Repo) WorktreeSelectorModel {
	// Filter out main worktrees
	var filtered []*proto.Worktree
	for _, wt := range worktrees {
		if !wt.IsMain {
			filtered = append(filtered, wt)
		}
	}

	// Create repo map for quick lookup
	repoMap := make(map[string]*proto.Repo)
	for _, r := range repos {
		repoMap[r.Id] = r
	}

	// Build table columns
	columns := []table.Column{
		{Title: "Worktree ID", Width: 20},
		{Title: "Repository", Width: 35},
		{Title: "Branch", Width: 30},
		{Title: "Parent Branch", Width: 20},
	}

	// Build table rows
	rows := []table.Row{}
	for _, wt := range filtered {
		repo := repoMap[wt.RepoId]
		repoStr := "unknown"
		if repo != nil {
			repoStr = fmt.Sprintf("%s/%s/%s", repo.Host, repo.Owner, repo.Repo)
		}

		parentBranch := "-"
		if wt.ParentBranch != nil {
			parentBranch = *wt.ParentBranch
		}

		rows = append(rows, table.Row{
			wt.Id,
			repoStr,
			wt.Branch,
			parentBranch,
		})
	}

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

	return WorktreeSelectorModel{
		table:     t,
		worktrees: filtered,
		repos:     repoMap,
	}
}

func (m WorktreeSelectorModel) Init() tea.Cmd {
	return nil
}

func (m WorktreeSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if len(m.worktrees) > 0 {
				selectedIdx := m.table.Cursor()
				if selectedIdx < len(m.worktrees) {
					m.selected = m.worktrees[selectedIdx]
				}
			}
			m.quitting = true
			return m, tea.Quit
		}
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m WorktreeSelectorModel) View() string {
	if m.quitting {
		return ""
	}

	if len(m.worktrees) == 0 {
		return "No non-main worktrees found.\n"
	}

	return baseStyle.Render(m.table.View()) + "\n\nPress ↑/↓ to navigate • enter to select • q to quit\n"
}

func (m WorktreeSelectorModel) Selected() *proto.Worktree {
	return m.selected
}

func SelectWorktree(worktrees []*proto.Worktree, repos []*proto.Repo) (*proto.Worktree, error) {
	// Filter out main worktrees
	var filtered []*proto.Worktree
	for _, wt := range worktrees {
		if !wt.IsMain {
			filtered = append(filtered, wt)
		}
	}

	if len(filtered) == 0 {
		return nil, fmt.Errorf("no non-main worktrees found")
	}

	p := tea.NewProgram(NewWorktreeSelector(worktrees, repos))
	model, err := p.Run()
	if err != nil {
		return nil, err
	}

	m := model.(WorktreeSelectorModel)
	if m.Selected() == nil {
		return nil, fmt.Errorf("no worktree selected")
	}

	return m.Selected(), nil
}
