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
	"mitra/internal/util"
)

type MitraServiceServer struct {
	proto.UnimplementedMitraServiceServer
	logger zerolog.Logger
	cfg    *config.Config
	ctx    context.Context
	wg     *sync.WaitGroup
	state  *State
}

func NewMitraServiceServer(logger zerolog.Logger, cfg *config.Config, ctx context.Context, wg *sync.WaitGroup) (*MitraServiceServer, error) {
	state, err := NewState(logger, cfg)
	if err != nil {
		return nil, err
	}

	// Hydrate tmux sessions for existing worktrees
	if err := state.HydrateTmuxSessions(); err != nil {
		logger.Warn().Err(err).Msg("failed to hydrate tmux sessions, continuing anyway")
	}

	return &MitraServiceServer{
		logger: logger.With().Str("service", "mitra").Logger(),
		cfg:    cfg,
		ctx:    ctx,
		wg:     wg,
		state:  state,
	}, nil
}

func (s *MitraServiceServer) StartExistingWatchers() error {
	repos := s.state.GetRepos()

	s.logger.Info().
		Int("count", len(repos)).
		Msg("starting watchers for existing repos")

	for _, repo := range repos {
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
	return &proto.ListReposResponse{
		Repos: s.state.GetRepos(),
	}, nil
}

func (s *MitraServiceServer) AddRepo(ctx context.Context, req *proto.AddRepoRequest) (*proto.AddRepoResponse, error) {
	host, owner, repoName, err := parseGitURL(req.Url)
	if err != nil {
		return nil, err
	}

	// Check if repo already exists
	if existingRepo := s.state.CheckRepoExists(host, owner, repoName); existingRepo != nil {
		s.logger.Info().
			Str("host", host).
			Str("owner", owner).
			Str("repo", repoName).
			Msg("repository already exists, skipping")
		return &proto.AddRepoResponse{
			Repo: existingRepo,
		}, nil
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

	if err := s.state.AddRepo(repo); err != nil {
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
	return &proto.ListWorktreesResponse{
		Worktrees: s.state.GetWorktrees(req.RepoId),
	}, nil
}

func (s *MitraServiceServer) AddWorktree(ctx context.Context, req *proto.AddWorktreeRequest) (*proto.AddWorktreeResponse, error) {
	mainWorktree := s.state.FindMainWorktree(req.RepoId)
	if mainWorktree == nil {
		return nil, fmt.Errorf("main worktree not found for repo: %s", req.RepoId)
	}

	worktreeID := util.RandomName()
	branch := worktreeID
	if req.Branch != nil && *req.Branch != "" {
		branch = *req.Branch
	}

	branchWithPrefix := s.cfg.Repo.BranchPrefix + branch

	if existing := s.state.CheckWorktreeExists(req.RepoId, branchWithPrefix); existing != nil {
		s.logger.Info().
			Str("repo_id", req.RepoId).
			Str("branch", branch).
			Msg("worktree already exists")
		return &proto.AddWorktreeResponse{
			Worktree: existing,
		}, nil
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

	if err := s.state.AddWorktree(worktree); err != nil {
		return nil, err
	}

	s.logger.Info().
		Str("repo_id", req.RepoId).
		Str("branch", branch).
		Str("path", worktreePath).
		Msg("worktree created successfully")

	// Create tmux session for the worktree
	sessionName := branch
	if err := s.state.CreateTmuxSession(sessionName, worktreePath); err != nil {
		s.logger.Warn().
			Err(err).
			Str("session", sessionName).
			Str("path", worktreePath).
			Msg("failed to create tmux session, continuing anyway")
	}

	return &proto.AddWorktreeResponse{
		Worktree: worktree,
	}, nil
}

func (s *MitraServiceServer) AddWorktreeEntry(repoID, branch, path string, isMain bool) error {
	if existing := s.state.CheckWorktreeExists(repoID, branch); existing != nil {
		s.logger.Info().
			Str("repo_id", repoID).
			Str("branch", branch).
			Msg("worktree entry already exists")
		return nil
	}

	worktree := &proto.Worktree{
		Id:           util.RandomName(),
		RepoId:       repoID,
		Branch:       branch,
		Path:         path,
		IsMain:       isMain,
		ParentBranch: nil,
	}

	if err := s.state.AddWorktree(worktree); err != nil {
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
	worktreeToDelete, index := s.state.FindWorktreeByID(req.WorktreeId)
	if worktreeToDelete == nil {
		return nil, fmt.Errorf("worktree not found: %s", req.WorktreeId)
	}

	if worktreeToDelete.IsMain {
		return nil, fmt.Errorf("cannot delete main worktree")
	}

	mainWorktree := s.state.FindMainWorktree(worktreeToDelete.RepoId)
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

	if err := s.state.DeleteWorktree(index); err != nil {
		return nil, err
	}

	s.logger.Info().
		Str("worktree_id", req.WorktreeId).
		Msg("worktree deleted successfully")

	// Kill tmux session if it exists
	sessionName := strings.TrimPrefix(worktreeToDelete.Branch, s.cfg.Repo.BranchPrefix)
	if err := s.state.KillTmuxSession(sessionName); err != nil {
		s.logger.Warn().
			Err(err).
			Str("session", sessionName).
			Msg("failed to kill tmux session, continuing anyway")
	}

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
