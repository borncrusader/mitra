package selectors

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/table"

	"mitra/internal/proto"
	"mitra/internal/util"
)

func SelectRepo(repos []*proto.Repo, worktrees []*proto.Worktree, promptText string, sortAscending bool) (*proto.Repo, error) {
	if len(repos) == 0 {
		return nil, fmt.Errorf("no repositories found")
	}

	// Count worktrees per repo
	worktreeCounts := make(map[string]int)
	for _, wt := range worktrees {
		worktreeCounts[wt.RepoId]++
	}

	// Sort repos by worktree count, then by timestamp
	sort.Slice(repos, func(i, j int) bool {
		countI := worktreeCounts[repos[i].Id]
		countJ := worktreeCounts[repos[j].Id]

		// Primary sort: worktree count
		if countI != countJ {
			if sortAscending {
				return countI < countJ
			}
			return countI > countJ
		}

		// Secondary sort: timestamp (descending - newest first)
		timestampI := util.ExtractTimestamp(repos[i].Id)
		timestampJ := util.ExtractTimestamp(repos[j].Id)
		return timestampI > timestampJ
	})

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
