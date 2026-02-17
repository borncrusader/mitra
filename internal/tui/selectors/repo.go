package selectors

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/table"

	"mitra/internal/proto"
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

	// Sort repos by worktree count
	sort.Slice(repos, func(i, j int) bool {
		if sortAscending {
			return worktreeCounts[repos[i].Id] < worktreeCounts[repos[j].Id]
		}
		return worktreeCounts[repos[i].Id] > worktreeCounts[repos[j].Id]
	})

	// Build table columns
	columns := []table.Column{
		{Title: "Repo ID", Width: 20},
		{Title: "Repository", Width: 50},
		{Title: "Worktrees", Width: 10},
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
