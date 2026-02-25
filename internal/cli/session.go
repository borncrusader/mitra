package cli

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"mitra/internal/proto"
	"mitra/internal/tmux"
	"mitra/internal/tui/selectors"
)

func (c *Command) sessionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "session",
		Aliases: []string{"s", "sess"},
		Short:   "Manage sessions",
	}
	cmd.AddCommand(c.sessionListCmd())
	cmd.AddCommand(c.sessionAttachCmd())
	return cmd
}

func (c *Command) sessionListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"l", "ls"},
		Short:   "List sessions",
		Run: func(cmd *cobra.Command, args []string) {
			cc, err := c.newClient()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to create client")
			}
			defer cc.Close()

			sessionsResp, err := cc.Client.ListSessions(cc.Ctx, &proto.ListSessionsRequest{})
			if err != nil {
				log.Fatal().Err(err).Msg("failed to list sessions")
			}

			if len(sessionsResp.Sessions) == 0 {
				fmt.Println("No sessions found.")
				return
			}

			worktreesResp, err := cc.Client.ListWorktrees(cc.Ctx, &proto.ListWorktreesRequest{})
			if err != nil {
				log.Fatal().Err(err).Msg("failed to list worktrees")
			}

			reposResp, err := cc.Client.ListRepos(cc.Ctx, &proto.ListReposRequest{})
			if err != nil {
				log.Fatal().Err(err).Msg("failed to list repos")
			}

			worktreeMap := make(map[string]*proto.Worktree)
			for _, wt := range worktreesResp.Worktrees {
				worktreeMap[wt.Id] = wt
			}

			repoMap := make(map[string]*proto.Repo)
			for _, r := range reposResp.Repos {
				repoMap[r.Id] = r
			}

			columns := []table.Column{
				{Title: "Repository", Width: selectors.ColumnWidths["Repository"]},
				{Title: "Branch", Width: selectors.ColumnWidths["Branch"]},
				{Title: "Worktree ID", Width: selectors.ColumnWidths["Worktree ID"]},
			}

			rows := []table.Row{}
			for _, session := range sessionsResp.Sessions {
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

			fmt.Print(selectors.RenderTable(columns, rows))
		},
	}
}

func (c *Command) sessionAttachCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "attach [session-id]",
		Aliases: []string{"a", "at"},
		Short:   "Attach to a session",
		Args:    cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {
			cc, err := c.newClient()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to create client")
			}
			defer cc.Close()

			var sessionID string

			if len(args) == 0 {
				sessionsResp, err := cc.Client.ListSessions(cc.Ctx, &proto.ListSessionsRequest{})
				if err != nil {
					log.Fatal().Err(err).Msg("failed to list sessions")
				}

				worktreesResp, err := cc.Client.ListWorktrees(cc.Ctx, &proto.ListWorktreesRequest{})
				if err != nil {
					log.Fatal().Err(err).Msg("failed to list worktrees")
				}

				reposResp, err := cc.Client.ListRepos(cc.Ctx, &proto.ListReposRequest{})
				if err != nil {
					log.Fatal().Err(err).Msg("failed to list repos")
				}

				selectedSession, err := selectors.SelectSession(sessionsResp.Sessions, worktreesResp.Worktrees, reposResp.Repos)
				if err != nil {
					log.Fatal().Err(err).Msg("failed to select session")
				}

				sessionID = selectedSession.Id
			} else {
				sessionID = args[0]

				_, err = cc.Client.GetSession(cc.Ctx, &proto.GetSessionRequest{
					SessionId: sessionID,
				})
				if err != nil {
					log.Fatal().Err(err).Msg("failed to get session")
				}
			}

			if err := tmux.AttachSessionExec(sessionID); err != nil {
				log.Fatal().Err(err).Msg("failed to attach to session")
			}
		},
	}
}
