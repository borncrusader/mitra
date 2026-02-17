package selectors

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"

	"mitra/internal/proto"
)

func SelectRepo(repos []*proto.Repo, promptText string) (*proto.Repo, error) {
	if len(repos) == 0 {
		return nil, fmt.Errorf("no repositories found")
	}

	// Build table columns
	columns := []table.Column{
		{Title: "Repo ID", Width: 20},
		{Title: "Host", Width: 20},
		{Title: "Owner", Width: 20},
		{Title: "Repository", Width: 25},
		{Title: "Main Branch", Width: 15},
	}

	// Build table rows
	rows := []table.Row{}
	for _, repo := range repos {
		rows = append(rows, table.Row{
			repo.Id,
			repo.Host,
			repo.Owner,
			repo.Repo,
			repo.MainBranch,
		})
	}

	if promptText == "" {
		promptText = "Select a repository:"
	}

	selectedIdx, err := RunTableSelector(columns, rows, promptText)
	if err != nil {
		return nil, err
	}

	if selectedIdx == -1 {
		return nil, fmt.Errorf("no repository selected")
	}

	return repos[selectedIdx], nil
}
