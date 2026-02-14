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

type RepoWatcher struct {
	logger zerolog.Logger
	cfg    *config.Config
}

func NewRepoWatcher(logger zerolog.Logger, cfg *config.Config) *RepoWatcher {
	return &RepoWatcher{
		logger: logger.With().Str("component", "repo-watcher").Logger(),
		cfg:    cfg,
	}
}

func (w *RepoWatcher) Watch(ctx context.Context, repoURL, owner, repoName, repoDir string) {
	logger := w.logger.With().
		Str("owner", owner).
		Str("repo", repoName).
		Logger()

	defaultBranch, err := git.GetDefaultBranch(repoURL)
	if err != nil {
		logger.Warn().
			Err(err).
			Str("url", repoURL).
			Msg("failed to detect default branch, using 'main'")
		defaultBranch = "main"
	}

	logger.Info().
		Str("branch", defaultBranch).
		Msg("default branch detected")

	cloneDir := filepath.Join(repoDir, defaultBranch)

	if _, err := os.Stat(cloneDir); !os.IsNotExist(err) {
		logger.Info().
			Str("path", cloneDir).
			Msg("clone directory already exists, skipping clone")
	} else {
		logger.Info().
			Str("url", repoURL).
			Str("path", cloneDir).
			Str("branch", defaultBranch).
			Msg("starting clone")

		if err := git.Clone(repoURL, cloneDir, defaultBranch); err != nil {
			logger.Error().
				Err(err).
				Msg("failed to clone repository")
			return
		}

		logger.Info().
			Str("path", cloneDir).
			Msg("clone completed successfully")
	}

	syncInterval := time.Duration(w.cfg.Repo.SyncIntervalSecs) * time.Second
	ticker := time.NewTicker(syncInterval)
	defer ticker.Stop()

	logger.Info().
		Dur("interval", syncInterval).
		Msg("starting periodic sync")

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("stopping watcher due to context cancellation")
			return
		case <-ticker.C:
			isClean, err := git.IsClean(cloneDir)
			if err != nil {
				logger.Warn().
					Err(err).
					Msg("failed to check if worktree is clean")
				continue
			}

			if !isClean {
				logger.Debug().
					Msg("worktree is not clean, skipping pull")
				continue
			}

			logger.Info().
				Msg("pulling latest changes")

			if err := git.Pull(cloneDir); err != nil {
				logger.Warn().
					Err(err).
					Msg("failed to pull changes")
				continue
			}

			logger.Info().
				Msg("successfully pulled latest changes")
		}
	}
}
