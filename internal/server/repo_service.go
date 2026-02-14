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
		Repos: repos,
	}, nil
}

func (s *RepoServiceServer) AddRepo(ctx context.Context, req *proto.AddRepoRequest) (*proto.AddRepoResponse, error) {
	host, owner, repoName, err := parseGitURL(req.Url)
	if err != nil {
		return nil, err
	}

	repo := &proto.Repo{
		Id:    util.RandomName(),
		Url:   req.Url,
		Host:  host,
		Owner: owner,
		Repo:  repoName,
	}

	repos, err := storage.LoadRepos()
	if err != nil {
		return nil, err
	}

	repos = append(repos, repo)

	if err := storage.SaveRepos(repos); err != nil {
		return nil, err
	}

	return &proto.AddRepoResponse{
		Repo: repo,
	}, nil
}

func parseGitURL(url string) (host, owner, repo string, err error) {
	original := url

	// Remove protocol if present
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimSuffix(url, ".git")
	url = strings.Trim(url, "/")

	// Must have at least host/owner/repo
	parts := strings.Split(url, "/")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("invalid git URL: expected host/owner/repo, got: %s", original)
	}

	if parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("invalid git URL: host, owner, and repo cannot be empty, got: %s", original)
	}

	return parts[0], parts[1], parts[2], nil
}
