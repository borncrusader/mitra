package server

import (
	"fmt"
	"path/filepath"
	"strings"

	"mitra/internal/agents"
	"mitra/internal/git"
	"mitra/internal/proto"
)

type addWorktreeCmd struct {
	worktreeID   string
	branch       string
	parentBranch string
	responseChan chan<- *addWorktreeResult
}

type addWorktreeResult struct {
	worktree *proto.Worktree
	err      error
}

type deleteWorktreeCmd struct {
	worktreeID   string
	force        bool
	responseChan chan<- error
}

func (cmd *addWorktreeCmd) Execute(w *RepoWatcher) error {
	if !w.cloneReady {
		err := fmt.Errorf("repository clone not ready yet")
		cmd.responseChan <- &addWorktreeResult{err: err}
		return err
	}

	mainWorktree := w.service.state.FindMainWorktree(w.repoID)
	if mainWorktree == nil {
		err := fmt.Errorf("main worktree not found for repo: %s", w.repoID)
		cmd.responseChan <- &addWorktreeResult{err: err}
		return err
	}

	branchWithPrefix := w.cfg.Repo.BranchPrefix + cmd.branch

	if existing := w.service.state.CheckWorktreeExists(w.repoID, branchWithPrefix); existing != nil {
		w.logger.Info().
			Str("repo_id", w.repoID).
			Str("branch", cmd.branch).
			Msg("worktree already exists")
		cmd.responseChan <- &addWorktreeResult{worktree: existing}
		return nil
	}

	repoPath := filepath.Dir(mainWorktree.Path)
	worktreePath := filepath.Join(repoPath, cmd.branch)

	w.logger.Info().
		Str("repo_id", w.repoID).
		Str("branch", branchWithPrefix).
		Str("parent_branch", cmd.parentBranch).
		Str("path", worktreePath).
		Msg("creating worktree")

	if err := git.CreateWorktree(mainWorktree.Path, branchWithPrefix, worktreePath, cmd.parentBranch); err != nil {
		cmd.responseChan <- &addWorktreeResult{err: err}
		return err
	}

	if err := agents.SetupFiles(mainWorktree.Path, worktreePath); err != nil {
		w.logger.Warn().Err(err).Str("path", worktreePath).Msg("failed to setup claude files")
	}

	worktree := &proto.Worktree{
		Id:           cmd.worktreeID,
		RepoId:       w.repoID,
		Branch:       branchWithPrefix,
		Path:         worktreePath,
		IsMain:       false,
		ParentBranch: &cmd.parentBranch,
	}

	if err := w.service.state.AddWorktree(worktree); err != nil {
		cmd.responseChan <- &addWorktreeResult{err: err}
		return err
	}

	w.logger.Info().
		Str("repo_id", w.repoID).
		Str("branch", cmd.branch).
		Str("path", worktreePath).
		Msg("worktree created successfully")

	sessionCmd := NewCreateSessionCommand(cmd.worktreeID, worktreePath)
	w.service.sessionManager.SendCommand(sessionCmd)

	if err := <-sessionCmd.responseChan; err != nil {
		w.logger.Warn().
			Err(err).
			Str("session", cmd.worktreeID).
			Str("path", worktreePath).
			Msg("failed to create tmux session, continuing anyway")
	}

	cmd.responseChan <- &addWorktreeResult{worktree: worktree}
	return nil
}

func (cmd *deleteWorktreeCmd) Execute(w *RepoWatcher) error {
	if !w.cloneReady {
		err := fmt.Errorf("repository clone not ready yet")
		cmd.responseChan <- err
		return err
	}

	worktreeToDelete, index := w.service.state.FindWorktreeByID(cmd.worktreeID)
	if worktreeToDelete == nil {
		err := fmt.Errorf("worktree not found: %s", cmd.worktreeID)
		cmd.responseChan <- err
		return err
	}

	if worktreeToDelete.IsMain {
		err := fmt.Errorf("cannot delete main worktree")
		cmd.responseChan <- err
		return err
	}

	mainWorktree := w.service.state.FindMainWorktree(worktreeToDelete.RepoId)
	if mainWorktree == nil {
		err := fmt.Errorf("main worktree not found for repo")
		cmd.responseChan <- err
		return err
	}

	// Check if worktree is clean before deletion
	repo := w.service.state.FindRepoByID(worktreeToDelete.RepoId)
	if repo == nil {
		err := fmt.Errorf("repo not found for worktree")
		cmd.responseChan <- err
		return err
	}

	w.logger.Debug().
		Str("worktree_id", cmd.worktreeID).
		Str("path", worktreeToDelete.Path).
		Msg("checking if worktree is clean")

	if !cmd.force {
		isClean, reason, err := git.IsWorktreeClean(worktreeToDelete.Path, repo.MainBranch)
		if err != nil {
			err := fmt.Errorf("failed to check if worktree is clean: %w", err)
			cmd.responseChan <- err
			return err
		}

		w.logger.Debug().
			Str("worktree_id", cmd.worktreeID).
			Bool("is_clean", isClean).
			Str("reason", string(reason)).
			Msg("worktree clean check complete")

		if !isClean {
			err := fmt.Errorf("worktree is not clean: %s", reason)
			cmd.responseChan <- err
			return err
		}
	}

	w.logger.Info().
		Str("worktree_id", cmd.worktreeID).
		Str("branch", worktreeToDelete.Branch).
		Str("path", worktreeToDelete.Path).
		Msg("deleting worktree")

	w.logger.Debug().
		Str("worktree_id", cmd.worktreeID).
		Msg("removing git worktree")

	if err := git.RemoveWorktree(mainWorktree.Path, worktreeToDelete.Path); err != nil {
		if !strings.Contains(err.Error(), "is not a working tree") {
			cmd.responseChan <- err
			return err
		}
		w.logger.Warn().
			Str("worktree_id", cmd.worktreeID).
			Str("path", worktreeToDelete.Path).
			Msg("git worktree not found on disk, removing from config anyway")
	}

	w.logger.Debug().
		Str("worktree_id", cmd.worktreeID).
		Msg("git worktree removed")

	if err := w.service.state.DeleteWorktree(index); err != nil {
		cmd.responseChan <- err
		return err
	}

	w.logger.Info().
		Str("worktree_id", cmd.worktreeID).
		Msg("worktree deleted successfully")

	sessionCmd := NewKillSessionCommand(worktreeToDelete.Id)
	w.service.sessionManager.SendCommand(sessionCmd)

	if err := <-sessionCmd.responseChan; err != nil {
		w.logger.Warn().
			Err(err).
			Str("session", worktreeToDelete.Id).
			Msg("failed to kill tmux session, continuing anyway")
	}

	cmd.responseChan <- nil
	return nil
}
