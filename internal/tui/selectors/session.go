package selectors

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"mitra/internal/proto"
)

type SessionSelectorModel struct {
	sessions []*proto.Session
	cursor   int
	selected *proto.Session
	quitting bool
}

func NewSessionSelector(sessions []*proto.Session) SessionSelectorModel {
	return SessionSelectorModel{
		sessions: sessions,
	}
}

func (m SessionSelectorModel) Init() tea.Cmd {
	return nil
}

func (m SessionSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.cursor < len(m.sessions)-1 {
				m.cursor++
			}

		case "enter":
			if len(m.sessions) > 0 {
				m.selected = m.sessions[m.cursor]
			}
			m.quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m SessionSelectorModel) View() string {
	if m.quitting {
		return ""
	}

	if len(m.sessions) == 0 {
		return "No sessions found.\n"
	}

	s := "Select a session to attach:\n\n"

	for i, session := range m.sessions {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		s += fmt.Sprintf("%s %s - %s\n", cursor, session.Id, session.Name)
	}

	s += "\nPress up/down to navigate, enter to select, q to quit.\n"

	return s
}

func (m SessionSelectorModel) Selected() *proto.Session {
	return m.selected
}

func SelectSession(sessions []*proto.Session) (*proto.Session, error) {
	if len(sessions) == 0 {
		return nil, fmt.Errorf("no sessions found")
	}

	p := tea.NewProgram(NewSessionSelector(sessions))
	model, err := p.Run()
	if err != nil {
		return nil, err
	}

	m := model.(SessionSelectorModel)
	if m.Selected() == nil {
		return nil, fmt.Errorf("no session selected")
	}

	return m.Selected(), nil
}
