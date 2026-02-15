package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"mitra/internal/proto"
)

type WorktreeSelectorModel struct {
	worktrees []*proto.Worktree
	cursor    int
	selected  *proto.Worktree
	quitting  bool
}

func NewWorktreeSelector(worktrees []*proto.Worktree) WorktreeSelectorModel {
	// Filter out main worktrees
	var filtered []*proto.Worktree
	for _, wt := range worktrees {
		if !wt.IsMain {
			filtered = append(filtered, wt)
		}
	}

	return WorktreeSelectorModel{
		worktrees: filtered,
	}
}

func (m WorktreeSelectorModel) Init() tea.Cmd {
	return nil
}

func (m WorktreeSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.worktrees)-1 {
				m.cursor++
			}

		case "enter":
			if len(m.worktrees) > 0 {
				m.selected = m.worktrees[m.cursor]
			}
			m.quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m WorktreeSelectorModel) View() string {
	if m.quitting {
		return ""
	}

	if len(m.worktrees) == 0 {
		return "No non-main worktrees found.\n"
	}

	s := "Select a worktree to delete:\n\n"

	for i, wt := range m.worktrees {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		parentInfo := ""
		if wt.ParentBranch != nil {
			parentInfo = fmt.Sprintf(" (from %s)", *wt.ParentBranch)
		}

		s += fmt.Sprintf("%s %s - %s%s\n", cursor, wt.Branch, wt.Path, parentInfo)
	}

	s += "\nPress up/down to navigate, enter to select, q to quit.\n"

	return s
}

func (m WorktreeSelectorModel) Selected() *proto.Worktree {
	return m.selected
}

func SelectWorktree(worktrees []*proto.Worktree) (*proto.Worktree, error) {
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

	p := tea.NewProgram(NewWorktreeSelector(worktrees))
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
