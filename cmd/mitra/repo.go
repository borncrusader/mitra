package main

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"mitra/internal/proto"
	"mitra/internal/storage"
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

		resp, err := cc.Client.ListRepos(cc.Ctx, &proto.ListReposRequest{})
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

			selectedRepo, err := selectors.SelectRepo(listResp.Repos)
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
