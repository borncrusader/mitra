package selectors

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"

	"mitra/internal/proto"
)

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

	// Create repo map for quick lookup
	repoMap := make(map[string]*proto.Repo)
	for _, r := range repos {
		repoMap[r.Id] = r
	}

	// Build table columns
	columns := []table.Column{
		{Title: "Worktree ID", Width: ColumnWidths["Worktree ID"]},
		{Title: "Repository", Width: ColumnWidths["Repository"]},
		{Title: "Branch", Width: ColumnWidths["Branch"]},
		{Title: "Parent Branch", Width: ColumnWidths["Parent Branch"]},
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

	selectedIdx, err := RunTableSelector(columns, rows, "Select a worktree to delete:")
	if err != nil {
		return nil, err
	}

	if selectedIdx == -1 {
		return nil, fmt.Errorf("no worktree selected")
	}

	return filtered[selectedIdx], nil
}
