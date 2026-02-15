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
	logger   zerolog.Logger
	cfg      *config.Config
	repoURL  string
	owner    string
	repoName string
	repoDir  string
}

func NewRepoWatcher(logger zerolog.Logger, cfg *config.Config, repoURL, owner, repoName, repoDir string) *RepoWatcher {
	return &RepoWatcher{
		logger: logger.With().
			Str("component", "repo-watcher").
			Str("owner", owner).
			Str("repo", repoName).
			Logger(),
		cfg:      cfg,
		repoURL:  repoURL,
		owner:    owner,
		repoName: repoName,
		repoDir:  repoDir,
	}
}

func (w *RepoWatcher) Watch(ctx context.Context) {
	defaultBranch, err := git.GetDefaultBranch(w.repoURL)
	if err != nil {
		w.logger.Warn().
			Err(err).
			Str("url", w.repoURL).
			Msg("failed to detect default branch, using 'main'")
		defaultBranch = "main"
	}

	w.logger.Info().
		Str("branch", defaultBranch).
		Msg("default branch detected")

	cloneDir := filepath.Join(w.repoDir, defaultBranch)

	if _, err := os.Stat(cloneDir); !os.IsNotExist(err) {
		w.logger.Info().
			Str("path", cloneDir).
			Msg("clone directory already exists, skipping clone")
	} else {
		w.logger.Info().
			Str("url", w.repoURL).
			Str("path", cloneDir).
			Str("branch", defaultBranch).
			Msg("starting clone")

		if err := git.Clone(w.repoURL, cloneDir, defaultBranch); err != nil {
			w.logger.Error().
				Err(err).
				Msg("failed to clone repository")
			return
		}

		w.logger.Info().
			Str("path", cloneDir).
			Msg("clone completed successfully")
	}

	syncInterval := time.Duration(w.cfg.Repo.SyncIntervalSecs) * time.Second
	ticker := time.NewTicker(syncInterval)
	defer ticker.Stop()

	w.logger.Info().
		Int("interval_secs", w.cfg.Repo.SyncIntervalSecs).
		Msg("starting periodic sync")

	for {
		select {
		case <-ctx.Done():
			w.logger.Info().Msg("stopping watcher due to context cancellation")
			return
		case <-ticker.C:
			isClean, err := git.IsClean(cloneDir)
			if err != nil {
				w.logger.Warn().
					Err(err).
					Msg("failed to check if worktree is clean")
				continue
			}

			if !isClean {
				w.logger.Debug().
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
