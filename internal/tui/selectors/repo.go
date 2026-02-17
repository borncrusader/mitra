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
		{Title: "Repository", Width: 50},
	}

	// Build table rows
	rows := []table.Row{}
	for _, repo := range repos {
		repoPath := fmt.Sprintf("%s/%s/%s", repo.Host, repo.Owner, repo.Repo)
		rows = append(rows, table.Row{
			repo.Id,
			repoPath,
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
