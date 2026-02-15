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
	"mitra/internal/git"
	"mitra/internal/proto"
	"mitra/internal/storage"
	"mitra/internal/util"
)

type MitraServiceServer struct {
	proto.UnimplementedMitraServiceServer
	logger    zerolog.Logger
	cfg       *config.Config
	ctx       context.Context
	wg        *sync.WaitGroup
	repos     []*proto.Repo
	worktrees []*proto.Worktree
	mu        sync.RWMutex
}

func NewMitraServiceServer(logger zerolog.Logger, cfg *config.Config, ctx context.Context, wg *sync.WaitGroup) (*MitraServiceServer, error) {
	repos, err := storage.LoadRepos()
	if err != nil {
		return nil, fmt.Errorf("failed to load repos: %w", err)
	}

	worktrees, err := storage.LoadWorktrees()
	if err != nil {
		return nil, fmt.Errorf("failed to load worktrees: %w", err)
	}

	return &MitraServiceServer{
		logger:    logger.With().Str("service", "mitra").Logger(),
		cfg:       cfg,
		ctx:       ctx,
		wg:        wg,
		repos:     repos,
		worktrees: worktrees,
	}, nil
}

func (s *MitraServiceServer) StartExistingWatchers() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	s.logger.Info().
		Int("count", len(s.repos)).
		Msg("starting watchers for existing repos")

	for _, repo := range s.repos {
		repoDir := filepath.Join(s.cfg.Repo.Dir, repo.Owner, repo.Repo)

		watcher := NewRepoWatcher(s.logger, s.cfg, repo.Url, repo.Id, repo.Owner, repo.Repo, repoDir, s)
		s.wg.Add(1)
		go func(w *RepoWatcher) {
			defer s.wg.Done()
			w.Watch(s.ctx)
		}(watcher)

		s.logger.Info().
			Str("owner", repo.Owner).
			Str("repo", repo.Repo).
			Msg("started watcher for existing repo")
	}

	return nil
}

func (s *MitraServiceServer) ListRepos(ctx context.Context, req *proto.ListReposRequest) (*proto.ListReposResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return &proto.ListReposResponse{
		Repos: s.repos,
	}, nil
}

func (s *MitraServiceServer) AddRepo(ctx context.Context, req *proto.AddRepoRequest) (*proto.AddRepoResponse, error) {
	host, owner, repoName, err := parseGitURL(req.Url)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, existingRepo := range s.repos {
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

	repo := &proto.Repo{
		Id:    util.RandomName(),
		Url:   req.Url,
		Host:  host,
		Owner: owner,
		Repo:  repoName,
	}

	s.logger.Info().
		Str("host", host).
		Str("owner", owner).
		Str("repo", repoName).
		Str("url", req.Url).
		Msg("adding repository")

	s.repos = append(s.repos, repo)

	if err := storage.SaveRepos(s.repos); err != nil {
		s.repos = s.repos[:len(s.repos)-1]
		return nil, err
	}

	repoDir := filepath.Join(s.cfg.Repo.Dir, owner, repoName)
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create repo directory: %w", err)
	}

	watcher := NewRepoWatcher(s.logger, s.cfg, req.Url, repo.Id, owner, repoName, repoDir, s)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		watcher.Watch(s.ctx)
	}()

	return &proto.AddRepoResponse{
		Repo: repo,
	}, nil
}

func (s *MitraServiceServer) ListWorktrees(ctx context.Context, req *proto.ListWorktreesRequest) (*proto.ListWorktreesResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []*proto.Worktree
	for _, wt := range s.worktrees {
		if req.RepoId == "" || wt.RepoId == req.RepoId {
			filtered = append(filtered, wt)
		}
	}

	return &proto.ListWorktreesResponse{
		Worktrees: filtered,
	}, nil
}

func (s *MitraServiceServer) AddWorktree(ctx context.Context, req *proto.AddWorktreeRequest) (*proto.AddWorktreeResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var mainWorktree *proto.Worktree
	for _, wt := range s.worktrees {
		if wt.RepoId == req.RepoId && wt.IsMain {
			mainWorktree = wt
			break
		}
	}

	if mainWorktree == nil {
		return nil, fmt.Errorf("main worktree not found for repo: %s", req.RepoId)
	}

	worktreeID := util.RandomName()
	branch := worktreeID
	if req.Branch != nil && *req.Branch != "" {
		branch = *req.Branch
	}

	branchWithPrefix := s.cfg.Repo.BranchPrefix + branch

	for _, existing := range s.worktrees {
		if existing.RepoId == req.RepoId && existing.Branch == branchWithPrefix {
			s.logger.Info().
				Str("repo_id", req.RepoId).
				Str("branch", branch).
				Msg("worktree already exists")
			return &proto.AddWorktreeResponse{
				Worktree: existing,
			}, nil
		}
	}

	parentBranch := mainWorktree.Branch
	if req.ParentBranch != nil && *req.ParentBranch != "" {
		parentBranch = *req.ParentBranch
	}

	repoPath := filepath.Dir(mainWorktree.Path)
	worktreePath := filepath.Join(repoPath, branch)

	s.logger.Info().
		Str("repo_id", req.RepoId).
		Str("branch", branchWithPrefix).
		Str("parent_branch", parentBranch).
		Str("path", worktreePath).
		Msg("creating worktree")

	if err := git.CreateWorktree(mainWorktree.Path, branchWithPrefix, worktreePath, parentBranch); err != nil {
		return nil, err
	}

	worktree := &proto.Worktree{
		Id:           worktreeID,
		RepoId:       req.RepoId,
		Branch:       branchWithPrefix,
		Path:         worktreePath,
		IsMain:       false,
		ParentBranch: &parentBranch,
	}

	s.worktrees = append(s.worktrees, worktree)

	if err := storage.SaveWorktrees(s.worktrees); err != nil {
		s.worktrees = s.worktrees[:len(s.worktrees)-1]
		return nil, err
	}

	s.logger.Info().
		Str("repo_id", req.RepoId).
		Str("branch", branch).
		Str("path", worktreePath).
		Msg("worktree created successfully")

	return &proto.AddWorktreeResponse{
		Worktree: worktree,
	}, nil
}

func (s *MitraServiceServer) AddWorktreeEntry(repoID, branch, path string, isMain bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, wt := range s.worktrees {
		if wt.RepoId == repoID && wt.Branch == branch {
			s.logger.Info().
				Str("repo_id", repoID).
				Str("branch", branch).
				Msg("worktree entry already exists")
			return nil
		}
	}

	worktree := &proto.Worktree{
		Id:           util.RandomName(),
		RepoId:       repoID,
		Branch:       branch,
		Path:         path,
		IsMain:       isMain,
		ParentBranch: nil,
	}

	s.worktrees = append(s.worktrees, worktree)

	if err := storage.SaveWorktrees(s.worktrees); err != nil {
		s.worktrees = s.worktrees[:len(s.worktrees)-1]
		return err
	}

	s.logger.Info().
		Str("repo_id", repoID).
		Str("branch", branch).
		Str("path", path).
		Bool("is_main", isMain).
		Msg("worktree entry added")

	return nil
}

func (s *MitraServiceServer) DeleteWorktree(ctx context.Context, req *proto.DeleteWorktreeRequest) (*proto.DeleteWorktreeResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var worktreeToDelete *proto.Worktree
	var index int

	for i, wt := range s.worktrees {
		if wt.Id == req.WorktreeId {
			worktreeToDelete = wt
			index = i
			break
		}
	}

	if worktreeToDelete == nil {
		return nil, fmt.Errorf("worktree not found: %s", req.WorktreeId)
	}

	if worktreeToDelete.IsMain {
		return nil, fmt.Errorf("cannot delete main worktree")
	}

	var mainWorktree *proto.Worktree
	for _, wt := range s.worktrees {
		if wt.RepoId == worktreeToDelete.RepoId && wt.IsMain {
			mainWorktree = wt
			break
		}
	}

	if mainWorktree == nil {
		return nil, fmt.Errorf("main worktree not found for repo")
	}

	s.logger.Info().
		Str("worktree_id", req.WorktreeId).
		Str("branch", worktreeToDelete.Branch).
		Str("path", worktreeToDelete.Path).
		Msg("deleting worktree")

	if err := git.RemoveWorktree(mainWorktree.Path, worktreeToDelete.Path); err != nil {
		return nil, err
	}

	s.worktrees = append(s.worktrees[:index], s.worktrees[index+1:]...)

	if err := storage.SaveWorktrees(s.worktrees); err != nil {
		return nil, err
	}

	s.logger.Info().
		Str("worktree_id", req.WorktreeId).
		Msg("worktree deleted successfully")

	return &proto.DeleteWorktreeResponse{
		Success: true,
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
