package main

import (
	"context"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"mitra/internal/config"
	"mitra/internal/proto"
	"mitra/internal/tmux"
	"mitra/internal/tui"
	"mitra/internal/util"
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
		cfg, err := config.Load()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load config")
		}

		conn, err := grpc.NewClient("localhost"+cfg.Server.GrpcPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to server")
		}
		defer util.DeferCheck(conn.Close)

		client := proto.NewMitraServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := client.ListSessions(ctx, &proto.ListSessionsRequest{})
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
		cfg, err := config.Load()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load config")
		}

		conn, err := grpc.NewClient("localhost"+cfg.Server.GrpcPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to server")
		}
		defer util.DeferCheck(conn.Close)

		client := proto.NewMitraServiceClient(conn)

		var sessionID string

		if len(args) == 0 {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			listResp, err := client.ListSessions(ctx, &proto.ListSessionsRequest{})
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

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			_, err = client.GetSession(ctx, &proto.GetSessionRequest{
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
