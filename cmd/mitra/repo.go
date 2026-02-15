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

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Manage repositories",
}

var repoAddCmd = &cobra.Command{
	Use:   "add <git-url>",
	Short: "Add a repository",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		gitURL := args[0]

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

		resp, err := client.AddRepo(ctx, &proto.AddRepoRequest{
			Url: gitURL,
		})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to add repo")
		}

		fmt.Printf("Added repo: %s/%s (id: %s)\n", resp.Repo.Owner, resp.Repo.Repo, resp.Repo.Id)
	},
}

var repoListCmd = &cobra.Command{
	Use:   "list",
	Short: "List repositories",
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

		resp, err := client.ListRepos(ctx, &proto.ListReposRequest{})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to list repos")
		}

		if len(resp.Repos) == 0 {
			return
		}

		storageRepos := make([]*storage.Repo, len(resp.Repos))
		for i, r := range resp.Repos {
			storageRepos[i] = &storage.Repo{
				ID:    r.Id,
				URL:   r.Url,
				Host:  r.Host,
				Owner: r.Owner,
				Repo:  r.Repo,
			}
		}

		repoStorage := struct {
			Repos []*storage.Repo `toml:"repos"`
		}{
			Repos: storageRepos,
		}

		encoder := toml.NewEncoder(os.Stdout)
		if err := encoder.Encode(repoStorage); err != nil {
			log.Fatal().Err(err).Msg("failed to encode repos")
		}
	},
}

func init() {
	repoCmd.AddCommand(repoAddCmd)
	repoCmd.AddCommand(repoListCmd)
	rootCmd.AddCommand(repoCmd)
}
