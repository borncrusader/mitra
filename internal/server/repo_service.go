package server

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"mitra/internal/config"
	"mitra/internal/git"
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

	log.Printf("Adding repository: %s/%s/%s", host, owner, repoName)

	repos, err := storage.LoadRepos()
	if err != nil {
		return nil, err
	}

	repos = append(repos, repo)

	if err := storage.SaveRepos(repos); err != nil {
		return nil, err
	}

	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	repoDir := filepath.Join(cfg.Repo.Dir, owner, repoName)
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create repo directory: %w", err)
	}

	defaultBranch, err := git.GetDefaultBranch(req.Url)
	if err != nil {
		log.Printf("Failed to detect default branch, using 'main': %v", err)
		defaultBranch = "main"
	}

	log.Printf("Default branch detected: %s", defaultBranch)

	cloneDir := filepath.Join(repoDir, defaultBranch)

	if _, err := os.Stat(cloneDir); !os.IsNotExist(err) {
		log.Printf("Clone directory already exists: %s, skipping clone", cloneDir)
		return &proto.AddRepoResponse{
			Repo: repo,
		}, nil
	}

	log.Printf("Starting clone of %s into %s", req.Url, cloneDir)

	if err := git.Clone(req.Url, cloneDir, defaultBranch); err != nil {
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	log.Printf("Successfully cloned repository to %s", cloneDir)

	return &proto.AddRepoResponse{
		Repo: repo,
	}, nil
}

func parseGitURL(url string) (host, owner, repo string, err error) {
	original := url

	if strings.HasPrefix(url, "git@") {
		url = strings.TrimPrefix(url, "git@")
		url = strings.TrimSuffix(url, ".git")

		colonIdx := strings.Index(url, ":")
		if colonIdx == -1 {
			return "", "", "", fmt.Errorf("invalid SSH git URL: expected git@host:owner/repo, got: %s", original)
		}

		host = url[:colonIdx]
		pathPart := url[colonIdx+1:]

		parts := strings.Split(pathPart, "/")
		if len(parts) != 2 {
			return "", "", "", fmt.Errorf("invalid SSH git URL: expected git@host:owner/repo, got: %s", original)
		}

		if host == "" || parts[0] == "" || parts[1] == "" {
			return "", "", "", fmt.Errorf("invalid SSH git URL: host, owner, and repo cannot be empty, got: %s", original)
		}

		return host, parts[0], parts[1], nil
	}

	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimSuffix(url, ".git")
	url = strings.Trim(url, "/")

	parts := strings.Split(url, "/")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("invalid git URL: expected host/owner/repo, got: %s", original)
	}

	if parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("invalid git URL: host, owner, and repo cannot be empty, got: %s", original)
	}

	return parts[0], parts[1], parts[2], nil
}
