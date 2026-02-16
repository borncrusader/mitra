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
	"mitra/internal/storage"
	"mitra/internal/util"
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

		resp, err := client.AddRepo(ctx, &proto.AddRepoRequest{
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
