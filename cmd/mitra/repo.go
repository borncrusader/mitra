package main

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/table"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"mitra/internal/proto"
	"mitra/internal/tui/selectors"
)

var repoCmd = &cobra.Command{
	Use:     "repo",
	Aliases: []string{"r"},
	Short:   "Manage repositories",
}

var repoAddCmd = &cobra.Command{
	Use:     "add <git-url>",
	Aliases: []string{"a"},
	Short:   "Add a repository",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		gitURL := args[0]

		cc, err := newClient()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create client")
		}
		defer cc.Close()

		resp, err := cc.Client.AddRepo(cc.Ctx, &proto.AddRepoRequest{
			Url: gitURL,
		})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to add repo")
		}

		log.Info().
			Str("owner", resp.Repo.Owner).
			Str("repo", resp.Repo.Repo).
			Str("id", resp.Repo.Id).
			Msg("added repo")
	},
}

var repoListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l", "ls"},
	Short:   "List repositories",
	Run: func(cmd *cobra.Command, args []string) {
		cc, err := newClient()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create client")
		}
		defer cc.Close()

		reposResp, err := cc.Client.ListRepos(cc.Ctx, &proto.ListReposRequest{})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to list repos")
		}

		if len(reposResp.Repos) == 0 {
			fmt.Println("No repositories found.")
			return
		}

		worktreesResp, err := cc.Client.ListWorktrees(cc.Ctx, &proto.ListWorktreesRequest{})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to list worktrees")
		}

		// Count worktrees per repo
		worktreeCounts := make(map[string]int)
		for _, wt := range worktreesResp.Worktrees {
			worktreeCounts[wt.RepoId]++
		}

		// Sort repos by worktree count (descending)
		sort.Slice(reposResp.Repos, func(i, j int) bool {
			return worktreeCounts[reposResp.Repos[i].Id] > worktreeCounts[reposResp.Repos[j].Id]
		})

		// Build table columns
		columns := []table.Column{
			{Title: "Repo ID", Width: 20},
			{Title: "Repository", Width: 50},
			{Title: "Worktrees", Width: 10},
		}

		// Build table rows
		rows := []table.Row{}
		for _, repo := range reposResp.Repos {
			repoPath := fmt.Sprintf("%s/%s/%s", repo.Host, repo.Owner, repo.Repo)
			count := worktreeCounts[repo.Id]
			rows = append(rows, table.Row{
				repo.Id,
				repoPath,
				fmt.Sprintf("%d", count),
			})
		}

		output := selectors.RenderTable(columns, rows)
		fmt.Print(output)
	},
}

var repoDeleteCmd = &cobra.Command{
	Use:     "delete [repo-id]",
	Aliases: []string{"d", "del", "rm"},
	Short:   "Delete a repository",
	Args:    cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		cc, err := newClient()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create client")
		}
		defer cc.Close()

		var repoID string

		if len(args) == 0 {
			listResp, err := cc.Client.ListRepos(cc.Ctx, &proto.ListReposRequest{})
			if err != nil {
				log.Fatal().Err(err).Msg("failed to list repos")
			}

			selectedRepo, err := selectors.SelectRepo(listResp.Repos, "Select a repository to delete:")
			if err != nil {
				log.Fatal().Err(err).Msg("failed to select repo")
			}

			repoID = selectedRepo.Id
		} else {
			repoID = args[0]
		}

		resp, err := cc.Client.DeleteRepo(cc.Ctx, &proto.DeleteRepoRequest{
			RepoId: repoID,
		})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to delete repo")
		}

		if resp.Success {
			log.Info().Str("repo_id", repoID).Msg(resp.Message)
		} else {
			log.Error().Str("repo_id", repoID).Msg(resp.Message)
		}
	},
}

func init() {
	repoCmd.AddCommand(repoAddCmd)
	repoCmd.AddCommand(repoListCmd)
	repoCmd.AddCommand(repoDeleteCmd)
	rootCmd.AddCommand(repoCmd)
}
