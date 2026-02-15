package main

import (
	"context"
	"fmt"
	"os"
	"strings"
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
	Use:   "add <repo-id> [branch[:parent-branch]]",
	Short: "Add a new worktree",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		repoID := args[0]

		var branch string
		var parentBranch *string

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

		req := &proto.AddWorktreeRequest{
			RepoId: repoID,
		}
		if branch != "" {
			req.Branch = &branch
		}
		if parentBranch != nil {
			req.ParentBranch = parentBranch
		}

		resp, err := client.AddWorktree(ctx, req)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to add worktree")
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
