package cli

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"mitra/internal/proto"
	"mitra/internal/tui/selectors"
)

func (c *Command) repoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "repo",
		Aliases: []string{"r"},
		Short:   "Manage repositories",
	}
	cmd.AddCommand(c.repoAddCmd())
	cmd.AddCommand(c.repoListCmd())
	cmd.AddCommand(c.repoDeleteCmd())
	return cmd
}

func (c *Command) repoAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "add <git-url>",
		Aliases: []string{"a"},
		Short:   "Add a repository",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			gitURL := args[0]

			cc, err := c.newClient()
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
}

func (c *Command) repoListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"l", "ls"},
		Short:   "List repositories",
		Run: func(cmd *cobra.Command, args []string) {
			cc, err := c.newClient()
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

			selectors.SortReposByWorktreeCount(reposResp.Repos, worktreesResp.Worktrees, false)

			worktreeCounts := make(map[string]int)
			for _, wt := range worktreesResp.Worktrees {
				worktreeCounts[wt.RepoId]++
			}

			columns := []table.Column{
				{Title: "Repo ID", Width: selectors.ColumnWidths["Repo ID"]},
				{Title: "Repository", Width: selectors.ColumnWidths["Repository"]},
				{Title: "Worktrees", Width: selectors.ColumnWidths["Worktrees"]},
			}

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

			fmt.Print(selectors.RenderTable(columns, rows))
		},
	}
}

func (c *Command) repoDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "delete [repo-id]",
		Aliases: []string{"d", "del", "rm"},
		Short:   "Delete a repository",
		Args:    cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {
			cc, err := c.newClient()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to create client")
			}
			defer cc.Close()

			var repoID string

			if len(args) == 0 {
				reposResp, err := cc.Client.ListRepos(cc.Ctx, &proto.ListReposRequest{})
				if err != nil {
					log.Fatal().Err(err).Msg("failed to list repos")
				}

				worktreesResp, err := cc.Client.ListWorktrees(cc.Ctx, &proto.ListWorktreesRequest{})
				if err != nil {
					log.Fatal().Err(err).Msg("failed to list worktrees")
				}

				selectedRepo, err := selectors.SelectRepo(reposResp.Repos, worktreesResp.Worktrees, "Select a repository to delete:", true)
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
}
