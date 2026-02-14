package server

import (
	"context"
	"fmt"
	"strings"

	"mitra/internal/proto"
	"mitra/internal/storage"
	"mitra/internal/util"
)

type RepoServiceServer struct {
	proto.UnimplementedRepoServiceServer
}

func NewRepoServiceServer() *RepoServiceServer {
	return &RepoServiceServer{}
}

func (s *RepoServiceServer) ListRepos(ctx context.Context, req *proto.ListReposRequest) (*proto.ListReposResponse, error) {
	repos, err := storage.LoadRepos()
	if err != nil {
		return nil, err
	}

	return &proto.ListReposResponse{
		Repos: repos.Repos,
	}, nil
}

func (s *RepoServiceServer) AddRepo(ctx context.Context, req *proto.AddRepoRequest) (*proto.AddRepoResponse, error) {
	owner, repoName, err := parseGitHubURL(req.GithubUrl)
	if err != nil {
		return nil, err
	}

	repo := &proto.Repo{
		Id:        util.RandomName(),
		GithubUrl: req.GithubUrl,
		Owner:     owner,
		Repo:      repoName,
	}

	repos, err := storage.LoadRepos()
	if err != nil {
		return nil, err
	}

	repos.Repos = append(repos.Repos, repo)

	if err := storage.SaveRepos(repos); err != nil {
		return nil, err
	}

	return &proto.AddRepoResponse{
		Repo: repo,
	}, nil
}

func parseGitHubURL(url string) (owner, repo string, err error) {
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "github.com/")
	url = strings.TrimSuffix(url, ".git")
	url = strings.Trim(url, "/")

	parts := strings.Split(url, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid GitHub URL format, expected owner/repo")
	}

	return parts[0], parts[1], nil
}
