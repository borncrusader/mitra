package selectors

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"

	"mitra/internal/proto"
)

func SelectSession(sessions []*proto.Session) (*proto.Session, error) {
	if len(sessions) == 0 {
		return nil, fmt.Errorf("no sessions found")
	}

	// Build table columns
	columns := []table.Column{
		{Title: "Session ID", Width: 20},
		{Title: "Worktree ID", Width: 20},
		{Title: "Name", Width: 60},
	}

	// Build table rows
	rows := []table.Row{}
	for _, session := range sessions {
		rows = append(rows, table.Row{
			session.Id,
			session.WorktreeId,
			session.Name,
		})
	}

	selectedIdx, err := RunTableSelector(columns, rows, "Select a session to attach:")
	if err != nil {
		return nil, err
	}

	if selectedIdx == -1 {
		return nil, fmt.Errorf("no session selected")
	}

	return sessions[selectedIdx], nil
}
