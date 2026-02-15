package server

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog"

	"mitra/internal/config"
	"mitra/internal/proto"
	"mitra/internal/storage"
	"mitra/internal/tmux"
)

// State manages the application state for repos, worktrees, and tmux sessions
type State struct {
	mu        sync.RWMutex
	repos     []*proto.Repo
	worktrees []*proto.Worktree
	sessions  map[string]bool // session name -> exists
	logger    zerolog.Logger
	cfg       *config.Config
}

// NewState creates a new state manager and loads data from storage
func NewState(logger zerolog.Logger, cfg *config.Config) (*State, error) {
	repos, err := storage.LoadRepos()
	if err != nil {
		return nil, fmt.Errorf("failed to load repos: %w", err)
	}

	worktrees, err := storage.LoadWorktrees()
	if err != nil {
		return nil, fmt.Errorf("failed to load worktrees: %w", err)
	}

	return &State{
		repos:     repos,
		worktrees: worktrees,
		sessions:  make(map[string]bool),
		logger:    logger.With().Str("component", "state").Logger(),
		cfg:       cfg,
	}, nil
}

// HydrateTmuxSessions creates tmux sessions for all existing worktrees
func (s *State) HydrateTmuxSessions() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Info().
		Int("worktrees", len(s.worktrees)).
		Msg("hydrating tmux sessions")

	for _, wt := range s.worktrees {
		sessionName := wt.Id

		exists, err := tmux.SessionExists(sessionName)
		if err != nil {
			s.logger.Warn().
				Err(err).
				Str("session", sessionName).
				Msg("failed to check tmux session")
			continue
		}

		if exists {
			s.sessions[sessionName] = true
			s.logger.Debug().
				Str("session", sessionName).
				Msg("tmux session already exists")
			continue
		}

		if err := tmux.CreateSession(sessionName, wt.Path); err != nil {
			s.logger.Warn().
				Err(err).
				Str("session", sessionName).
				Str("path", wt.Path).
				Msg("failed to create tmux session")
		} else {
			s.sessions[sessionName] = true
			s.logger.Info().
				Str("session", sessionName).
				Str("path", wt.Path).
				Msg("tmux session created")
		}
	}

	s.logger.Info().
		Int("sessions", len(s.sessions)).
		Msg("tmux session hydration complete")

	return nil
}

// AddRepo adds a new repository
func (s *State) AddRepo(repo *proto.Repo) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.repos = append(s.repos, repo)

	if err := storage.SaveRepos(s.repos); err != nil {
		s.repos = s.repos[:len(s.repos)-1]
		return err
	}

	return nil
}

// GetRepos returns all repositories
func (s *State) GetRepos() []*proto.Repo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.repos
}

// FindRepoByID finds a repository by ID
func (s *State) FindRepoByID(id string) *proto.Repo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, repo := range s.repos {
		if repo.Id == id {
			return repo
		}
	}
	return nil
}

// CheckRepoExists checks if a repository already exists
func (s *State) CheckRepoExists(host, owner, repo string) *proto.Repo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, existingRepo := range s.repos {
		if existingRepo.Host == host && existingRepo.Owner == owner && existingRepo.Repo == repo {
			return existingRepo
		}
	}
	return nil
}

// AddWorktree adds a new worktree
func (s *State) AddWorktree(worktree *proto.Worktree) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.worktrees = append(s.worktrees, worktree)

	if err := storage.SaveWorktrees(s.worktrees); err != nil {
		s.worktrees = s.worktrees[:len(s.worktrees)-1]
		return err
	}

	return nil
}

// GetWorktrees returns all worktrees, optionally filtered by repo ID
func (s *State) GetWorktrees(repoID string) []*proto.Worktree {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if repoID == "" {
		return s.worktrees
	}

	var filtered []*proto.Worktree
	for _, wt := range s.worktrees {
		if wt.RepoId == repoID {
			filtered = append(filtered, wt)
		}
	}
	return filtered
}

// FindMainWorktree finds the main worktree for a repository
func (s *State) FindMainWorktree(repoID string) *proto.Worktree {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, wt := range s.worktrees {
		if wt.RepoId == repoID && wt.IsMain {
			return wt
		}
	}
	return nil
}

// CheckWorktreeExists checks if a worktree with the given branch already exists
func (s *State) CheckWorktreeExists(repoID, branch string) *proto.Worktree {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, wt := range s.worktrees {
		if wt.RepoId == repoID && wt.Branch == branch {
			return wt
		}
	}
	return nil
}

// FindWorktreeByID finds a worktree by ID
func (s *State) FindWorktreeByID(id string) (*proto.Worktree, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for i, wt := range s.worktrees {
		if wt.Id == id {
			return wt, i
		}
	}
	return nil, -1
}

// DeleteWorktree removes a worktree from state
func (s *State) DeleteWorktree(index int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.worktrees = append(s.worktrees[:index], s.worktrees[index+1:]...)

	if err := storage.SaveWorktrees(s.worktrees); err != nil {
		return err
	}

	return nil
}

// CreateTmuxSession creates a tmux session and tracks it
func (s *State) CreateTmuxSession(sessionName, path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	exists, err := tmux.SessionExists(sessionName)
	if err != nil {
		return err
	}

	if exists {
		s.sessions[sessionName] = true
		s.logger.Info().
			Str("session", sessionName).
			Msg("tmux session already exists")
		return nil
	}

	if err := tmux.CreateSession(sessionName, path); err != nil {
		return err
	}

	s.sessions[sessionName] = true
	s.logger.Info().
		Str("session", sessionName).
		Str("path", path).
		Msg("tmux session created")

	return nil
}

// KillTmuxSession kills a tmux session and removes it from tracking
func (s *State) KillTmuxSession(sessionName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	exists, err := tmux.SessionExists(sessionName)
	if err != nil {
		return err
	}

	if !exists {
		delete(s.sessions, sessionName)
		return nil
	}

	if err := tmux.KillSession(sessionName); err != nil {
		return err
	}

	delete(s.sessions, sessionName)
	s.logger.Info().
		Str("session", sessionName).
		Msg("tmux session killed")

	return nil
}

// GetSessions returns all tracked session names
func (s *State) GetSessions() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make([]string, 0, len(s.sessions))
	for name := range s.sessions {
		sessions = append(sessions, name)
	}
	return sessions
}
