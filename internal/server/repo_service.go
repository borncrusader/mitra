package server

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/rs/zerolog"
	"mitra/internal/config"
	"mitra/internal/proto"
	"mitra/internal/storage"
	"mitra/internal/util"
)

type RepoServiceServer struct {
	proto.UnimplementedRepoServiceServer
	logger zerolog.Logger
	cfg    *config.Config
	ctx    context.Context
	wg     *sync.WaitGroup
}

func NewRepoServiceServer(logger zerolog.Logger, cfg *config.Config, ctx context.Context, wg *sync.WaitGroup) *RepoServiceServer {
	return &RepoServiceServer{
		logger: logger.With().Str("service", "repo").Logger(),
		cfg:    cfg,
		ctx:    ctx,
		wg:     wg,
	}
}

func (s *RepoServiceServer) StartExistingWatchers() error {
	repos, err := storage.LoadRepos()
	if err != nil {
		return fmt.Errorf("failed to load existing repos: %w", err)
	}

	s.logger.Info().
		Int("count", len(repos)).
		Msg("starting watchers for existing repos")

	for _, repo := range repos {
		repoDir := filepath.Join(s.cfg.Repo.Dir, repo.Owner, repo.Repo)

		watcher := NewRepoWatcher(s.logger, s.cfg)
		s.wg.Add(1)
		go func(r *proto.Repo, dir string) {
			defer s.wg.Done()
			watcher.Watch(s.ctx, r.Url, r.Owner, r.Repo, dir)
		}(repo, repoDir)

		s.logger.Info().
			Str("owner", repo.Owner).
			Str("repo", repo.Repo).
			Msg("started watcher for existing repo")
	}

	return nil
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

	for _, existingRepo := range repos {
		if existingRepo.Host == host && existingRepo.Owner == owner && existingRepo.Repo == repoName {
			s.logger.Info().
				Str("host", host).
				Str("owner", owner).
				Str("repo", repoName).
				Msg("repository already exists, skipping")
			return &proto.AddRepoResponse{
				Repo: existingRepo,
			}, nil
		}
	}

	s.logger.Info().
		Str("host", host).
		Str("owner", owner).
		Str("repo", repoName).
		Str("url", req.Url).
		Msg("adding repository")

	repos = append(repos, repo)

	if err := storage.SaveRepos(repos); err != nil {
		return nil, err
	}

	repoDir := filepath.Join(s.cfg.Repo.Dir, owner, repoName)
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create repo directory: %w", err)
	}

	watcher := NewRepoWatcher(s.logger, s.cfg)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		watcher.Watch(s.ctx, req.Url, owner, repoName, repoDir)
	}()

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
