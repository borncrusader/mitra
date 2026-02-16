package main

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"mitra/internal/proto"
	"mitra/internal/tmux"
	"mitra/internal/tui"
)

var sessionCmd = &cobra.Command{
	Use:     "session",
	Aliases: []string{"s", "sess"},
	Short:   "Manage sessions",
}

var sessionListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l", "ls"},
	Short:   "List sessions",
	Run: func(cmd *cobra.Command, args []string) {
		cc, err := newClient()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create client")
		}
		defer cc.Close()

		resp, err := cc.Client.ListSessions(cc.Ctx, &proto.ListSessionsRequest{})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to list sessions")
		}

		if len(resp.Sessions) == 0 {
			return
		}

		type sessionOutput struct {
			ID         string `toml:"id"`
			WorktreeID string `toml:"worktree_id"`
			Name       string `toml:"name"`
		}

		sessions := make([]sessionOutput, len(resp.Sessions))
		for i, s := range resp.Sessions {
			sessions[i] = sessionOutput{
				ID:         s.Id,
				WorktreeID: s.WorktreeId,
				Name:       s.Name,
			}
		}

		sessionStorage := struct {
			Sessions []sessionOutput `toml:"sessions"`
		}{
			Sessions: sessions,
		}

		encoder := toml.NewEncoder(os.Stdout)
		if err := encoder.Encode(sessionStorage); err != nil {
			log.Fatal().Err(err).Msg("failed to encode sessions")
		}
	},
}

var sessionAttachCmd = &cobra.Command{
	Use:     "attach [session-id]",
	Aliases: []string{"a", "at"},
	Short:   "Attach to a session",
	Args:    cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		cc, err := newClient()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create client")
		}
		defer cc.Close()

		var sessionID string

		if len(args) == 0 {
			listResp, err := cc.Client.ListSessions(cc.Ctx, &proto.ListSessionsRequest{})
			if err != nil {
				log.Fatal().Err(err).Msg("failed to list sessions")
			}

			selectedSession, err := tui.SelectSession(listResp.Sessions)
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

func init() {
	sessionCmd.AddCommand(sessionListCmd)
	sessionCmd.AddCommand(sessionAttachCmd)
	rootCmd.AddCommand(sessionCmd)
}
