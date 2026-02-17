package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"mitra/internal/proto"
	"mitra/internal/tui/selectors"
)

var worktreeCmd = &cobra.Command{
	Use:     "worktree",
	Aliases: []string{"w", "wt"},
	Short:   "Manage worktrees",
}

var worktreeAddCmd = &cobra.Command{
	Use:     "add [repo-id] [branch[:parent-branch]]",
	Aliases: []string{"a"},
	Short:   "Add a new worktree",
	Args:    cobra.RangeArgs(0, 2),
	Run: func(cmd *cobra.Command, args []string) {
		cc, err := newClient()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create client")
		}
		defer cc.Close()

		var repoID string
		var branch string
		var parentBranch *string

		if len(args) == 0 {
			reposResp, err := cc.Client.ListRepos(cc.Ctx, &proto.ListReposRequest{})
			if err != nil {
				log.Fatal().Err(err).Msg("failed to list repos")
			}

			worktreesResp, err := cc.Client.ListWorktrees(cc.Ctx, &proto.ListWorktreesRequest{})
			if err != nil {
				log.Fatal().Err(err).Msg("failed to list worktrees")
			}

			selectedRepo, err := selectors.SelectRepo(reposResp.Repos, worktreesResp.Worktrees, "Select a repository for the new worktree:", false)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to select repo")
			}

			repoID = selectedRepo.Id
		} else {
			repoID = args[0]

			if len(args) == 2 {
				branchArg := args[1]
				parts := strings.SplitN(branchArg, ":", 2)

				branch = parts[0]
				if len(parts) == 2 {
					if parts[1] != "" {
						parentBranch = &parts[1]
					}
				}
			}
		}

		req := &proto.AddWorktreeRequest{
			RepoId: repoID,
		}
		if branch != "" {
			req.Branch = &branch
		}
		if parentBranch != nil {
			req.ParentBranch = parentBranch
		}

		resp, err := cc.Client.AddWorktree(cc.Ctx, req)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to add worktree")
		}

		log.Info().
			Str("branch", resp.Worktree.Branch).
			Str("path", resp.Worktree.Path).
			Msg("created worktree")
	},
}

var worktreeListCmd = &cobra.Command{
	Use:     "list [repo-id]",
	Aliases: []string{"l", "ls"},
	Short:   "List worktrees",
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cc, err := newClient()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create client")
		}
		defer cc.Close()

		repoID := ""
		if len(args) > 0 {
			repoID = args[0]
		}

		worktreesResp, err := cc.Client.ListWorktrees(cc.Ctx, &proto.ListWorktreesRequest{
			RepoId: repoID,
		})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to list worktrees")
		}

		if len(worktreesResp.Worktrees) == 0 {
			fmt.Println("No worktrees found.")
			return
		}

		reposResp, err := cc.Client.ListRepos(cc.Ctx, &proto.ListReposRequest{})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to list repos")
		}

		// Create repo map for quick lookup
		repoMap := make(map[string]*proto.Repo)
		for _, r := range reposResp.Repos {
			repoMap[r.Id] = r
		}

		// Group worktrees by repo
		worktreesByRepo := make(map[string][]*proto.Worktree)
		for _, wt := range worktreesResp.Worktrees {
			worktreesByRepo[wt.RepoId] = append(worktreesByRepo[wt.RepoId], wt)
		}

		// Build table columns (without Repository column since it's in section headers)
		columns := []table.Column{
			{Title: "Worktree ID", Width: selectors.ColumnWidths["Worktree ID"]},
			{Title: "Branch", Width: selectors.ColumnWidths["Branch"]},
			{Title: "Parent Branch", Width: selectors.ColumnWidths["Parent Branch"]},
			{Title: "Status", Width: selectors.ColumnWidths["Status"]},
		}

		// Get repos sorted by worktree count (descending)
		type repoWithCount struct {
			repo  *proto.Repo
			count int
		}
		var sortedRepos []repoWithCount
		for repoID, worktrees := range worktreesByRepo {
			if repo := repoMap[repoID]; repo != nil {
				sortedRepos = append(sortedRepos, repoWithCount{repo, len(worktrees)})
			}
		}
		sort.Slice(sortedRepos, func(i, j int) bool {
			return sortedRepos[i].count > sortedRepos[j].count
		})

		// Style for section headers
		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")).
			PaddingLeft(1)

		// Print each repo section
		for i, rc := range sortedRepos {
			// Print section header
			repoPath := fmt.Sprintf("%s/%s/%s", rc.repo.Host, rc.repo.Owner, rc.repo.Repo)
			header := fmt.Sprintf("Repository: %s (%d worktrees)", repoPath, rc.count)
			fmt.Println(headerStyle.Render(header))

			// Build rows for this repo
			rows := []table.Row{}
			for _, wt := range worktreesByRepo[rc.repo.Id] {
				parentBranch := "-"
				if wt.ParentBranch != nil {
					parentBranch = *wt.ParentBranch
				}

				rows = append(rows, table.Row{
					wt.Id,
					wt.Branch,
					parentBranch,
					wt.Status,
				})
			}

			// Print table for this repo
			output := selectors.RenderTable(columns, rows)
			fmt.Print(output)

			// Add spacing between sections (except after last section)
			if i < len(sortedRepos)-1 {
				fmt.Println()
			}
		}
	},
}

var worktreeDeleteCmd = &cobra.Command{
	Use:     "delete [worktree-id]",
	Aliases: []string{"d", "del", "rm"},
	Short:   "Delete a worktree",
	Args:    cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		cc, err := newClient()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create client")
		}
		defer cc.Close()

		var worktreeID string

		if len(args) == 0 {
			worktreesResp, err := cc.Client.ListWorktrees(cc.Ctx, &proto.ListWorktreesRequest{})
			if err != nil {
				log.Fatal().Err(err).Msg("failed to list worktrees")
			}

			reposResp, err := cc.Client.ListRepos(cc.Ctx, &proto.ListReposRequest{})
			if err != nil {
				log.Fatal().Err(err).Msg("failed to list repos")
			}

			selectedWorktree, err := selectors.SelectWorktree(worktreesResp.Worktrees, reposResp.Repos)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to select worktree")
			}

			worktreeID = selectedWorktree.Id
		} else {
			worktreeID = args[0]
		}

		resp, err := cc.Client.DeleteWorktree(cc.Ctx, &proto.DeleteWorktreeRequest{
			WorktreeId: worktreeID,
		})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to delete worktree")
		}

		if resp.Success {
			log.Info().Str("worktree_id", worktreeID).Msg(resp.Message)
		} else {
			log.Error().Str("worktree_id", worktreeID).Msg(resp.Message)
		}
	},
}

func init() {
	worktreeCmd.AddCommand(worktreeAddCmd)
	worktreeCmd.AddCommand(worktreeListCmd)
	worktreeCmd.AddCommand(worktreeDeleteCmd)
	rootCmd.AddCommand(worktreeCmd)
}
