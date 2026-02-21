package selectors

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"

	"mitra/internal/proto"
)

func SelectRepo(repos []*proto.Repo, worktrees []*proto.Worktree, promptText string, sortAscending bool) (*proto.Repo, error) {
	if len(repos) == 0 {
		return nil, fmt.Errorf("no repositories found")
	}

	// Sort repos using shared function
	SortReposByWorktreeCount(repos, worktrees, sortAscending)

	// Count worktrees per repo for display
	worktreeCounts := make(map[string]int)
	for _, wt := range worktrees {
		worktreeCounts[wt.RepoId]++
	}

	// Build table columns
	columns := []table.Column{
		{Title: "Repo ID", Width: ColumnWidths["Repo ID"]},
		{Title: "Repository", Width: ColumnWidths["Repository"]},
		{Title: "Worktrees", Width: ColumnWidths["Worktrees"]},
	}

	// Build table rows
	rows := []table.Row{}
	for _, repo := range repos {
		repoPath := fmt.Sprintf("%s/%s/%s", repo.Host, repo.Owner, repo.Repo)
		count := worktreeCounts[repo.Id]
		rows = append(rows, table.Row{
			repo.Id,
			repoPath,
			fmt.Sprintf("%d", count),
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
