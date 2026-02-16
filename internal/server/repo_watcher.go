package server

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"

	"mitra/internal/config"
	"mitra/internal/git"
)

type repoCommand interface {
	Execute(w *RepoWatcher) error
}

type RepoWatcher struct {
	logger      zerolog.Logger
	cfg         *config.Config
	repoURL     string
	repoID      string
	owner       string
	repoName    string
	repoDir     string
	mainBranch  string
	service     *MitraServiceServer
	commandChan chan repoCommand
	cloneReady  bool
}

func NewRepoWatcher(logger zerolog.Logger, cfg *config.Config, repoURL, repoID, owner, repoName, repoDir string, mainBranch string, service *MitraServiceServer) *RepoWatcher {
	return &RepoWatcher{
		logger: logger.With().
			Str("component", "repo-watcher").
			Str("owner", owner).
			Str("repo", repoName).
			Logger(),
		cfg:         cfg,
		repoURL:     repoURL,
		repoID:      repoID,
		owner:       owner,
		repoName:    repoName,
		repoDir:     repoDir,
		mainBranch:  mainBranch,
		service:     service,
		commandChan: make(chan repoCommand, 10),
		cloneReady:  false,
	}
}

func (w *RepoWatcher) Watch(ctx context.Context) {
	cloneDir := filepath.Join(w.repoDir, w.mainBranch)

	if _, err := os.Stat(cloneDir); !os.IsNotExist(err) {
		w.logger.Info().
			Str("path", cloneDir).
			Msg("clone directory already exists, skipping clone")
		w.cloneReady = true
	} else {
		w.logger.Info().
			Str("url", w.repoURL).
			Str("path", cloneDir).
			Str("branch", w.mainBranch).
			Msg("starting clone")

		if err := git.Clone(w.repoURL, cloneDir, w.mainBranch); err != nil {
			w.logger.Error().
				Err(err).
				Msg("failed to clone repository")
			return
		}

		w.logger.Info().
			Str("path", cloneDir).
			Msg("clone completed successfully")

		if err := w.createWorktreeEntry(w.mainBranch, cloneDir); err != nil {
			w.logger.Error().
				Err(err).
				Msg("failed to create worktree entry")
		}

		w.cloneReady = true
	}

	syncInterval := time.Duration(w.cfg.Repo.SyncIntervalSecs) * time.Second
	ticker := time.NewTicker(syncInterval)
	defer ticker.Stop()

	w.logger.Info().
		Int("interval_secs", w.cfg.Repo.SyncIntervalSecs).
		Msg("starting repo watcher")

	for {
		select {
		case <-ctx.Done():
			w.logger.Info().Msg("stopping watcher due to context cancellation")
			return
		case cmd := <-w.commandChan:
			w.logger.Debug().Msg("processing repo command")
			if err := cmd.Execute(w); err != nil {
				w.logger.Warn().
					Err(err).
					Msg("command execution failed")
			}
		case <-ticker.C:
			isClean, reason, err := git.IsClean(cloneDir)
			if err != nil {
				w.logger.Warn().
					Err(err).
					Msg("failed to check if worktree is clean")
				continue
			}

			if !isClean {
				w.logger.Debug().
					Str("reason", string(reason)).
					Msg("worktree is not clean, skipping pull")
				continue
			}

			w.logger.Info().
				Msg("pulling latest changes")

			if err := git.Pull(cloneDir); err != nil {
				w.logger.Warn().
					Err(err).
					Msg("failed to pull changes")
				continue
			}

			w.logger.Info().
				Msg("successfully pulled latest changes")
		}
	}
}

func (w *RepoWatcher) createWorktreeEntry(branch, path string) error {
	return w.service.AddWorktreeEntry(w.repoID, branch, path, true)
}

func (w *RepoWatcher) SendCommand(cmd repoCommand) {
	w.commandChan <- cmd
}
