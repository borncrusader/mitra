package selectors

import (
	"sort"

	"mitra/internal/proto"
	"mitra/internal/util"
)

// ColumnWidths defines the standard widths for table columns across the application
var ColumnWidths = map[string]int{
	"Repo ID":       25,
	"Repository":    50,
	"Branch":        35,
	"Worktree ID":   25,
	"Parent Branch": 35,
	"Status":        20,
	"Worktrees":     10,
}

// SortReposByWorktreeCount sorts repos by worktree count and timestamp
// Primary sort: worktree count (direction based on ascending parameter)
// Secondary sort: timestamp (always descending - newest first)
func SortReposByWorktreeCount(repos []*proto.Repo, worktrees []*proto.Worktree, ascending bool) {
	// Count worktrees per repo
	worktreeCounts := make(map[string]int)
	for _, wt := range worktrees {
		worktreeCounts[wt.RepoId]++
	}

	sort.Slice(repos, func(i, j int) bool {
		countI := worktreeCounts[repos[i].Id]
		countJ := worktreeCounts[repos[j].Id]

		// Primary sort: worktree count
		if countI != countJ {
			if ascending {
				return countI < countJ
			}
			return countI > countJ
		}

		// Secondary sort: timestamp (descending - newest first)
		timestampI := util.ExtractTimestamp(repos[i].Id)
		timestampJ := util.ExtractTimestamp(repos[j].Id)
		return timestampI > timestampJ
	})
}

// GroupWorktreesByRepo groups and sorts worktrees by repository
// Returns a slice of repo IDs sorted by worktree count (descending), then timestamp
type RepoWorktrees struct {
	Repo      *proto.Repo
	Worktrees []*proto.Worktree
}

func GroupWorktreesByRepo(worktrees []*proto.Worktree, repos []*proto.Repo) []RepoWorktrees {
	// Create repo map
	repoMap := make(map[string]*proto.Repo)
	for _, r := range repos {
		repoMap[r.Id] = r
	}

	// Group worktrees by repo
	worktreesByRepo := make(map[string][]*proto.Worktree)
	for _, wt := range worktrees {
		worktreesByRepo[wt.RepoId] = append(worktreesByRepo[wt.RepoId], wt)
	}

	// Create sorted list
	var result []RepoWorktrees
	for repoID, wts := range worktreesByRepo {
		if repo := repoMap[repoID]; repo != nil {
			result = append(result, RepoWorktrees{
				Repo:      repo,
				Worktrees: wts,
			})
		}
	}

	// Sort by worktree count (descending), then timestamp
	sort.Slice(result, func(i, j int) bool {
		countI := len(result[i].Worktrees)
		countJ := len(result[j].Worktrees)

		if countI != countJ {
			return countI > countJ
		}

		timestampI := util.ExtractTimestamp(result[i].Repo.Id)
		timestampJ := util.ExtractTimestamp(result[j].Repo.Id)
		return timestampI > timestampJ
	})

	return result
}

type RepoSessions struct {
	Repo     *proto.Repo
	Sessions []*proto.Session
}

func GroupSessionsByRepo(sessions []*proto.Session, worktrees []*proto.Worktree, repos []*proto.Repo) []RepoSessions {
	worktreeMap := make(map[string]*proto.Worktree)
	for _, wt := range worktrees {
		worktreeMap[wt.Id] = wt
	}

	repoMap := make(map[string]*proto.Repo)
	for _, r := range repos {
		repoMap[r.Id] = r
	}

	sessionsByRepo := make(map[string][]*proto.Session)
	for _, s := range sessions {
		wt := worktreeMap[s.WorktreeId]
		if wt == nil {
			continue
		}
		sessionsByRepo[wt.RepoId] = append(sessionsByRepo[wt.RepoId], s)
	}

	var result []RepoSessions
	for repoID, ss := range sessionsByRepo {
		if repo := repoMap[repoID]; repo != nil {
			result = append(result, RepoSessions{
				Repo:     repo,
				Sessions: ss,
			})
		}
	}

	sort.Slice(result, func(i, j int) bool {
		countI := len(result[i].Sessions)
		countJ := len(result[j].Sessions)

		if countI != countJ {
			return countI > countJ
		}

		timestampI := util.ExtractTimestamp(result[i].Repo.Id)
		timestampJ := util.ExtractTimestamp(result[j].Repo.Id)
		return timestampI > timestampJ
	})

	return result
}
