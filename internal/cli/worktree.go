package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"mitra/internal/proto"
	"mitra/internal/tui/selectors"
)

func (c *Command) worktreeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "worktree",
		Aliases: []string{"w", "wt"},
		Short:   "Manage worktrees",
	}
	cmd.AddCommand(c.worktreeAddCmd())
	cmd.AddCommand(c.worktreeListCmd())
	cmd.AddCommand(c.worktreeDeleteCmd())
	return cmd
}

func (c *Command) worktreeAddCmd() *cobra.Command {
	cobraCmd := &cobra.Command{
		Use:     "add [repo-id] [branch[:parent-branch]]",
		Aliases: []string{"a"},
		Short:   "Add a new worktree",
		Args:    cobra.RangeArgs(0, 2),
		Run: func(cmd *cobra.Command, args []string) {
			cc, err := c.newClient()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to create client")
			}
			defer cc.Close()

			nextName, _ := cmd.Flags().GetBool("next")

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
			if nextName {
				req.NextName = &nextName
			} else if branch != "" {
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
	cobraCmd.Flags().Bool("next", false, "Use next available Greek alphabet name")
	return cobraCmd
}

func (c *Command) worktreeListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list [repo-id]",
		Aliases: []string{"l", "ls"},
		Short:   "List worktrees",
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cc, err := c.newClient()
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

			groupedWorktrees := selectors.GroupWorktreesByRepo(worktreesResp.Worktrees, reposResp.Repos)

			columns := []table.Column{
				{Title: "Worktree ID", Width: selectors.ColumnWidths["Worktree ID"]},
				{Title: "Branch", Width: selectors.ColumnWidths["Branch"]},
				{Title: "Parent Branch", Width: selectors.ColumnWidths["Parent Branch"]},
				{Title: "Status", Width: selectors.ColumnWidths["Status"]},
			}

			headerStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("12")).
				PaddingLeft(1)

			for i, rw := range groupedWorktrees {
				repoPath := fmt.Sprintf("%s/%s/%s", rw.Repo.Host, rw.Repo.Owner, rw.Repo.Repo)
				header := fmt.Sprintf("Repository: %s (%d worktrees)", repoPath, len(rw.Worktrees))
				fmt.Println(headerStyle.Render(header))

				rows := []table.Row{}
				for _, wt := range rw.Worktrees {
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

				fmt.Print(selectors.RenderTable(columns, rows))

				if i < len(groupedWorktrees)-1 {
					fmt.Println()
				}
			}
		},
	}
}

func (c *Command) worktreeDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "delete [worktree-id]",
		Aliases: []string{"d", "del", "rm"},
		Short:   "Delete a worktree",
		Args:    cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {
			cc, err := c.newClient()
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
				return
			}

			log.Warn().Str("worktree_id", worktreeID).Msg(resp.Message)

			fmt.Printf("Force delete anyway? [y/N] ")
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			if strings.ToLower(strings.TrimSpace(scanner.Text())) != "y" {
				return
			}

			resp, err = cc.Client.DeleteWorktree(cc.Ctx, &proto.DeleteWorktreeRequest{
				WorktreeId: worktreeID,
				Force:      true,
			})
			if err != nil {
				log.Fatal().Err(err).Msg("failed to force delete worktree")
			}

			if resp.Success {
				log.Info().Str("worktree_id", worktreeID).Msg(resp.Message)
			} else {
				log.Error().Str("worktree_id", worktreeID).Msg(resp.Message)
			}
		},
	}
}
