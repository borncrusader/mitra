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
	"mitra/internal/tmux"
	"mitra/internal/util"
)

type MitraServiceServer struct {
	proto.UnimplementedMitraServiceServer
	logger         zerolog.Logger
	cfg            *config.Config
	ctx            context.Context
	wg             *sync.WaitGroup
	state          *State
	watchers       map[string]*RepoWatcher
	sessionManager *SessionManager
	mu             sync.RWMutex
}

func NewMitraServiceServer(logger zerolog.Logger, cfg *config.Config, ctx context.Context, wg *sync.WaitGroup) (*MitraServiceServer, error) {
	state, err := NewState(logger, cfg)
	if err != nil {
		return nil, err
	}

	sessionManager := NewSessionManager(logger, cfg, state)

	wg.Add(1)
	go func() {
		defer wg.Done()
		sessionManager.Start(ctx)
	}()

	// Add tmux sessions for existing worktrees
	addCmd := NewAddSessionsCommand()
	sessionManager.SendCommand(addCmd)
	if err := <-addCmd.responseChan; err != nil {
		logger.Warn().Err(err).Msg("failed to add tmux sessions, continuing anyway")
	}

	return &MitraServiceServer{
		logger:         logger,
		cfg:            cfg,
		ctx:            ctx,
		wg:             wg,
		state:          state,
		watchers:       make(map[string]*RepoWatcher),
		sessionManager: sessionManager,
	}, nil
}

func (s *MitraServiceServer) StartExistingWatchers() error {
	repos := s.state.GetRepos()

	s.logger.Info().
		Int("count", len(repos)).
		Msg("starting watchers for existing repos")

	for _, repo := range repos {
		repoDir := filepath.Join(s.cfg.Repo.Dir, repo.Owner, repo.Repo)

		watcher := NewRepoWatcher(s.logger, s.cfg, repo.Url, repo.Id, repo.Owner, repo.Repo, repo.MainBranch, repoDir, s)

		s.mu.Lock()
		s.watchers[repo.Id] = watcher
		s.mu.Unlock()

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

	mainBranch, err := git.GetMainBranch(req.Url)
	if err != nil {
		s.logger.Warn().
			Err(err).
			Str("url", req.Url).
			Msg("failed to detect main branch, using 'main'")
		mainBranch = "main"
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
		Str("mainBranch", mainBranch).
		Msg("adding repository")

	if err := s.state.AddRepo(repo); err != nil {
		return nil, err
	}

	repoDir := filepath.Join(s.cfg.Repo.Dir, owner, repoName)
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create repo directory: %w", err)
	}

	watcher := NewRepoWatcher(s.logger, s.cfg, req.Url, repo.Id, owner, repoName, mainBranch, repoDir, s)

	s.mu.Lock()
	s.watchers[repo.Id] = watcher
	s.mu.Unlock()

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
	s.mu.RLock()
	watcher, exists := s.watchers[req.RepoId]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("repo watcher not found for repo: %s", req.RepoId)
	}

	mainWorktree := s.state.FindMainWorktree(req.RepoId)
	if mainWorktree == nil {
		return nil, fmt.Errorf("main worktree not found for repo: %s", req.RepoId)
	}

	worktreeID := util.RandomName()
	branch := worktreeID
	if req.Branch != nil && *req.Branch != "" {
		branch = *req.Branch
	}

	parentBranch := mainWorktree.Branch
	if req.ParentBranch != nil && *req.ParentBranch != "" {
		parentBranch = *req.ParentBranch
	}

	responseChan := make(chan *addWorktreeResult, 1)
	cmd := &addWorktreeCmd{
		worktreeID:   worktreeID,
		branch:       branch,
		parentBranch: parentBranch,
		responseChan: responseChan,
	}

	watcher.SendCommand(cmd)

	result := <-responseChan
	if result.err != nil {
		return nil, result.err
	}

	return &proto.AddWorktreeResponse{
		Worktree: result.worktree,
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
	worktreeToDelete, _ := s.state.FindWorktreeByID(req.WorktreeId)
	if worktreeToDelete == nil {
		return nil, fmt.Errorf("worktree not found: %s", req.WorktreeId)
	}

	s.mu.RLock()
	watcher, exists := s.watchers[worktreeToDelete.RepoId]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("repo watcher not found for repo: %s", worktreeToDelete.RepoId)
	}

	responseChan := make(chan error, 1)
	cmd := &deleteWorktreeCmd{
		worktreeID:   req.WorktreeId,
		responseChan: responseChan,
	}

	watcher.SendCommand(cmd)

	err := <-responseChan
	if err != nil {
		return nil, err
	}

	return &proto.DeleteWorktreeResponse{
		Success: true,
	}, nil
}

func (s *MitraServiceServer) ListSessions(ctx context.Context, req *proto.ListSessionsRequest) (*proto.ListSessionsResponse, error) {
	sessionNames, err := tmux.ListSessions()
	if err != nil {
		return nil, fmt.Errorf("failed to list tmux sessions: %w", err)
	}

	var sessions []*proto.Session
	for _, name := range sessionNames {
		worktree, _ := s.state.FindWorktreeByID(name)

		if worktree == nil {
			continue
		}

		session := &proto.Session{
			Id:         name,
			WorktreeId: name,
			Name:       fmt.Sprintf("%s (%s)", worktree.Branch, worktree.Path),
		}

		sessions = append(sessions, session)
	}

	return &proto.ListSessionsResponse{
		Sessions: sessions,
	}, nil
}

func (s *MitraServiceServer) GetSession(ctx context.Context, req *proto.GetSessionRequest) (*proto.GetSessionResponse, error) {
	exists, err := tmux.SessionExists(req.SessionId)
	if err != nil {
		return nil, fmt.Errorf("failed to check session: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("session not found: %s", req.SessionId)
	}

	worktree, _ := s.state.FindWorktreeByID(req.SessionId)

	session := &proto.Session{
		Id:         req.SessionId,
		WorktreeId: req.SessionId,
		Name:       req.SessionId,
	}

	if worktree != nil {
		session.Name = fmt.Sprintf("%s (%s)", worktree.Branch, worktree.Path)
	}

	return &proto.GetSessionResponse{
		Session: session,
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
