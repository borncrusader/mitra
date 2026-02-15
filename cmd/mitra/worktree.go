package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"mitra/internal/config"
	"mitra/internal/proto"
	"mitra/internal/storage"
)

var worktreeCmd = &cobra.Command{
	Use:   "worktree",
	Short: "Manage worktrees",
}

var worktreeAddCmd = &cobra.Command{
	Use:   "add <repo-id> <branch>",
	Short: "Add a new worktree",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		repoID := args[0]
		branch := args[1]

		cfg, err := config.Load()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load config")
		}

		conn, err := grpc.NewClient("localhost"+cfg.Server.GrpcPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to server")
		}
		defer conn.Close()

		client := proto.NewMitraServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resp, err := client.CreateWorktree(ctx, &proto.CreateWorktreeRequest{
			RepoId: repoID,
			Branch: branch,
		})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create worktree")
		}

		fmt.Printf("Created worktree: %s at %s\n", resp.Worktree.Branch, resp.Worktree.Path)
	},
}

var worktreeListCmd = &cobra.Command{
	Use:   "list [repo-id]",
	Short: "List worktrees",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load config")
		}

		conn, err := grpc.NewClient("localhost"+cfg.Server.GrpcPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to server")
		}
		defer conn.Close()

		client := proto.NewMitraServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		repoID := ""
		if len(args) > 0 {
			repoID = args[0]
		}

		resp, err := client.ListWorktrees(ctx, &proto.ListWorktreesRequest{
			RepoId: repoID,
		})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to list worktrees")
		}

		if len(resp.Worktrees) == 0 {
			return
		}

		storageWorktrees := make([]*storage.Worktree, len(resp.Worktrees))
		for i, wt := range resp.Worktrees {
			storageWorktrees[i] = &storage.Worktree{
				ID:           wt.Id,
				RepoID:       wt.RepoId,
				Branch:       wt.Branch,
				Path:         wt.Path,
				IsMain:       wt.IsMain,
				ParentBranch: wt.ParentBranch,
			}
		}

		worktreeStorage := struct {
			Worktrees []*storage.Worktree `toml:"worktrees"`
		}{
			Worktrees: storageWorktrees,
		}

		encoder := toml.NewEncoder(os.Stdout)
		if err := encoder.Encode(worktreeStorage); err != nil {
			log.Fatal().Err(err).Msg("failed to encode worktrees")
		}
	},
}

func init() {
	worktreeCmd.AddCommand(worktreeAddCmd)
	worktreeCmd.AddCommand(worktreeListCmd)
	rootCmd.AddCommand(worktreeCmd)
}
