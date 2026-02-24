package migration

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/rs/zerolog"

	"mitra/internal/config"
	"mitra/internal/git"
	"mitra/internal/proto"
	"mitra/internal/storage"
	"mitra/internal/util"
)

// Run performs migrations and validations on startup
func Run(logger zerolog.Logger) error {
	logger.Info().Msg("running migrations")

	if err := ensureConfigFiles(logger); err != nil {
		return fmt.Errorf("failed to ensure config files: %w", err)
	}

	if err := ensureRepoMainBranches(logger); err != nil {
		return fmt.Errorf("failed to ensure repo main branches: %w", err)
	}

	if err := ensureMainWorktrees(logger); err != nil {
		return fmt.Errorf("failed to ensure main worktrees: %w", err)
	}

	if err := syncUntrackedWorktrees(logger); err != nil {
		return fmt.Errorf("failed to sync untracked worktrees: %w", err)
	}

	logger.Info().Msg("migrations completed")
	return nil
}

func ensureConfigFiles(logger zerolog.Logger) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	mitraDir := filepath.Join(homeDir, ".mitra")
	if err := os.MkdirAll(mitraDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(mitraDir, "config.toml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		logger.Info().Str("path", configPath).Msg("creating config.toml")
		if err := createConfigFile(configPath); err != nil {
			return err
		}
	} else {
		// Config exists, ensure all default fields are present
		if err := ensureConfigDefaults(logger, configPath); err != nil {
			return err
		}
	}

	repoPath := filepath.Join(mitraDir, "repo.toml")
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		logger.Info().Str("path", repoPath).Msg("creating repo.toml")
		if err := createEmptyRepoFile(repoPath); err != nil {
			return err
		}
	}

	worktreePath := filepath.Join(mitraDir, "worktree.toml")
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		logger.Info().Str("path", worktreePath).Msg("creating worktree.toml")
		if err := createEmptyWorktreeFile(worktreePath); err != nil {
			return err
		}
	}

	return nil
}

func createConfigFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer util.DeferCheck(file.Close)

	cfg := config.Default()
	encoder := toml.NewEncoder(file)
	return encoder.Encode(cfg)
}

func ensureConfigDefaults(logger zerolog.Logger, path string) error {
	// Load existing config
	existingCfg, err := config.Load()
	if err != nil {
		return err
	}

	defaultCfg := config.Default()
	updated := false

	// Check Server config
	if existingCfg.Server.Port == "" {
		logger.Info().Msg("adding missing server.port to config")
		existingCfg.Server.Port = defaultCfg.Server.Port
		updated = true
	}
	if existingCfg.Server.GrpcPort == "" {
		logger.Info().Msg("adding missing server.grpc_port to config")
		existingCfg.Server.GrpcPort = defaultCfg.Server.GrpcPort
		updated = true
	}

	// Check Repo config
	if existingCfg.Repo.Dir == "" {
		logger.Info().Msg("adding missing repo.dir to config")
		existingCfg.Repo.Dir = defaultCfg.Repo.Dir
		updated = true
	}
	if existingCfg.Repo.SyncIntervalSecs == 0 {
		logger.Info().Msg("adding missing repo.sync_interval_secs to config")
		existingCfg.Repo.SyncIntervalSecs = defaultCfg.Repo.SyncIntervalSecs
		updated = true
	}
	if existingCfg.Repo.BranchPrefix == "" {
		logger.Info().Msg("adding missing repo.branch_prefix to config")
		existingCfg.Repo.BranchPrefix = defaultCfg.Repo.BranchPrefix
		updated = true
	}

	// Check Session config
	if existingCfg.Session.Type == "" {
		logger.Info().Msg("adding missing session.type to config")
		existingCfg.Session.Type = defaultCfg.Session.Type
		updated = true
	}
	if len(existingCfg.Session.Panes) == 0 {
		logger.Info().Msg("adding missing session.panes to config")
		existingCfg.Session.Panes = defaultCfg.Session.Panes
		updated = true
	}

	// Check Agents config
	if !existingCfg.Agents.Claude.Enabled && !existingCfg.Agents.Codex {
		logger.Info().Msg("adding missing agents config")
		existingCfg.Agents = defaultCfg.Agents
		updated = true
	}

	// Save if updated
	if updated {
		logger.Info().Str("path", path).Msg("updating config with missing defaults")
		file, err := os.Create(path)
		if err != nil {
			return err
		}
		defer util.DeferCheck(file.Close)

		encoder := toml.NewEncoder(file)
		return encoder.Encode(existingCfg)
	}

	return nil
}

func createEmptyRepoFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer util.DeferCheck(file.Close)

	emptyStorage := struct {
		Repos []*storage.Repo `toml:"repos"`
	}{
		Repos: []*storage.Repo{},
	}

	encoder := toml.NewEncoder(file)
	return encoder.Encode(emptyStorage)
}

func createEmptyWorktreeFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer util.DeferCheck(file.Close)

	emptyStorage := struct {
		Worktrees []*storage.Worktree `toml:"worktrees"`
	}{
		Worktrees: []*storage.Worktree{},
	}

	encoder := toml.NewEncoder(file)
	return encoder.Encode(emptyStorage)
}

func ensureRepoMainBranches(logger zerolog.Logger) error {
	repos, err := storage.LoadRepos()
	if err != nil {
		return err
	}

	updated := false
	for _, repo := range repos {
		if repo.MainBranch == "" {
			logger.Info().
				Str("repo_id", repo.Id).
				Str("url", repo.Url).
				Msg("detecting main branch for repo")

			mainBranch, err := git.GetMainBranch(repo.Url)
			if err != nil {
				logger.Warn().
					Err(err).
					Str("repo_id", repo.Id).
					Msg("failed to detect main branch, using 'main'")
				mainBranch = "main"
			}

			repo.MainBranch = mainBranch
			updated = true

			logger.Info().
				Str("repo_id", repo.Id).
				Str("main_branch", mainBranch).
				Msg("set main branch for repo")
		}
	}

	if updated {
		if err := storage.SaveRepos(repos); err != nil {
			return err
		}
		logger.Info().Msg("updated repos with main branches")
	}

	return nil
}

func ensureMainWorktrees(logger zerolog.Logger) error {
	repos, err := storage.LoadRepos()
	if err != nil {
		return err
	}

	worktrees, err := storage.LoadWorktrees()
	if err != nil {
		return err
	}

	worktreeMap := make(map[string]bool)
	for _, wt := range worktrees {
		if wt.IsMain {
			worktreeMap[wt.RepoId] = true
		}
	}

	var newWorktrees []*proto.Worktree
	for _, repo := range repos {
		if !worktreeMap[repo.Id] {
			logger.Info().
				Str("repo_id", repo.Id).
				Str("owner", repo.Owner).
				Str("repo", repo.Repo).
				Msg("creating missing main worktree")

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			mainBranch := repo.MainBranch
			if mainBranch == "" {
				mainBranch = "main"
			}

			worktreePath := filepath.Join(cfg.Repo.Dir, repo.Owner, repo.Repo, mainBranch)

			wt := &proto.Worktree{
				Id:           util.RandomName(),
				RepoId:       repo.Id,
				Branch:       mainBranch,
				Path:         worktreePath,
				IsMain:       true,
				ParentBranch: nil,
			}
			newWorktrees = append(newWorktrees, wt)
		}
	}

	if len(newWorktrees) > 0 {
		allWorktrees := append(worktrees, newWorktrees...)
		if err := storage.SaveWorktrees(allWorktrees); err != nil {
			return err
		}
		logger.Info().Int("count", len(newWorktrees)).Msg("created missing main worktrees")
	}

	return nil
}

func syncUntrackedWorktrees(logger zerolog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if !cfg.Repo.SyncUntrackedWorktrees {
		return nil
	}

	repos, err := storage.LoadRepos()
	if err != nil {
		return err
	}

	worktrees, err := storage.LoadWorktrees()
	if err != nil {
		return err
	}

	tracked := make(map[string]map[string]bool)
	mainWorktreeByRepo := make(map[string]*proto.Worktree)
	for _, wt := range worktrees {
		if tracked[wt.RepoId] == nil {
			tracked[wt.RepoId] = make(map[string]bool)
		}
		tracked[wt.RepoId][wt.Branch] = true
		if wt.IsMain {
			mainWorktreeByRepo[wt.RepoId] = wt
		}
	}

	var newWorktrees []*proto.Worktree
	for _, repo := range repos {
		mainWt, ok := mainWorktreeByRepo[repo.Id]
		if !ok {
			continue
		}

		if _, err := os.Stat(mainWt.Path); os.IsNotExist(err) {
			continue
		}

		gitWorktrees, err := git.ListWorktrees(mainWt.Path)
		if err != nil {
			logger.Warn().Err(err).Str("repo_id", repo.Id).Msg("failed to list git worktrees")
			continue
		}

		for _, gwt := range gitWorktrees {
			if gwt.IsMain || gwt.Branch == "" {
				continue
			}

			if tracked[repo.Id] != nil && tracked[repo.Id][gwt.Branch] {
				continue
			}

			logger.Info().
				Str("repo_id", repo.Id).
				Str("branch", gwt.Branch).
				Str("path", gwt.Path).
				Msg("adding untracked worktree to config")

			wt := &proto.Worktree{
				Id:           util.RandomName(),
				RepoId:       repo.Id,
				Branch:       gwt.Branch,
				Path:         gwt.Path,
				IsMain:       false,
				ParentBranch: nil,
			}
			newWorktrees = append(newWorktrees, wt)

			if tracked[repo.Id] == nil {
				tracked[repo.Id] = make(map[string]bool)
			}
			tracked[repo.Id][gwt.Branch] = true
		}
	}

	if len(newWorktrees) > 0 {
		allWorktrees := append(worktrees, newWorktrees...)
		if err := storage.SaveWorktrees(allWorktrees); err != nil {
			return err
		}
		logger.Info().Int("count", len(newWorktrees)).Msg("synced untracked worktrees")
	}

	return nil
}
