package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"mitra/internal/config"
	"mitra/internal/proto"
)

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Manage repositories",
}

var repoAddCmd = &cobra.Command{
	Use:   "add <github-url>",
	Short: "Add a repository",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		githubURL := args[0]

		cfg, err := config.Load()
		if err != nil {
			log.Fatal(err)
		}

		conn, err := grpc.NewClient("localhost"+cfg.Server.GrpcPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		client := proto.NewRepoServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := client.AddRepo(ctx, &proto.AddRepoRequest{
			GithubUrl: githubURL,
		})
		if err != nil {
			log.Fatalf("Failed to add repo: %v", err)
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
			log.Fatal(err)
		}

		conn, err := grpc.NewClient("localhost"+cfg.Server.GrpcPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		client := proto.NewRepoServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := client.ListRepos(ctx, &proto.ListReposRequest{})
		if err != nil {
			log.Fatalf("Failed to list repos: %v", err)
		}

		if len(resp.Repos) == 0 {
			fmt.Println("No repos configured")
			return
		}

		for _, repo := range resp.Repos {
			fmt.Printf("%s/%s\n", repo.Owner, repo.Repo)
			fmt.Printf("  ID: %s\n", repo.Id)
			fmt.Printf("  URL: %s\n", repo.GithubUrl)
			fmt.Println()
		}
	},
}

func init() {
	repoCmd.AddCommand(repoAddCmd)
	repoCmd.AddCommand(repoListCmd)
	rootCmd.AddCommand(repoCmd)
}
