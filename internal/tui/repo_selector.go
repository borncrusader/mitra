package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"mitra/internal/proto"
)

type RepoSelectorModel struct {
	repos    []*proto.Repo
	cursor   int
	selected *proto.Repo
	quitting bool
}

func NewRepoSelector(repos []*proto.Repo) RepoSelectorModel {
	return RepoSelectorModel{
		repos: repos,
	}
}

func (m RepoSelectorModel) Init() tea.Cmd {
	return nil
}

func (m RepoSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.cursor < len(m.repos)-1 {
				m.cursor++
			}

		case "enter":
			m.selected = m.repos[m.cursor]
			m.quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m RepoSelectorModel) View() string {
	if m.quitting {
		return ""
	}

	s := "Select a repository:\n\n"

	for i, repo := range m.repos {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		s += fmt.Sprintf("%s %s/%s/%s\n", cursor, repo.Host, repo.Owner, repo.Repo)
	}

	s += "\nPress up/down to navigate, enter to select, q to quit.\n"

	return s
}

func (m RepoSelectorModel) Selected() *proto.Repo {
	return m.selected
}

func SelectRepo(repos []*proto.Repo) (*proto.Repo, error) {
	if len(repos) == 0 {
		return nil, fmt.Errorf("no repositories found")
	}

	p := tea.NewProgram(NewRepoSelector(repos))
	model, err := p.Run()
	if err != nil {
		return nil, err
	}

	m := model.(RepoSelectorModel)
	if m.Selected() == nil {
		return nil, fmt.Errorf("no repository selected")
	}

	return m.Selected(), nil
}
