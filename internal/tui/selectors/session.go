package selectors

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"

	"mitra/internal/proto"
)

func SelectSession(sessions []*proto.Session, worktrees []*proto.Worktree, repos []*proto.Repo) (*proto.Session, error) {
	if len(sessions) == 0 {
		return nil, fmt.Errorf("no sessions found")
	}

	// Create lookup maps
	worktreeMap := make(map[string]*proto.Worktree)
	for _, wt := range worktrees {
		worktreeMap[wt.Id] = wt
	}

	repoMap := make(map[string]*proto.Repo)
	for _, r := range repos {
		repoMap[r.Id] = r
	}

	// Build table columns
	columns := []table.Column{
		{Title: "Repository", Width: 40},
		{Title: "Branch", Width: 30},
		{Title: "Worktree ID", Width: 20},
	}

	// Build table rows
	rows := []table.Row{}
	for _, session := range sessions {
		worktree := worktreeMap[session.WorktreeId]

		repoPath := "unknown"
		branch := "unknown"

		if worktree != nil {
			branch = worktree.Branch
			repo := repoMap[worktree.RepoId]
			if repo != nil {
				repoPath = fmt.Sprintf("%s/%s/%s", repo.Host, repo.Owner, repo.Repo)
			}
		}

		rows = append(rows, table.Row{
			repoPath,
			branch,
			session.WorktreeId,
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
