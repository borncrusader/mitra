package server

import (
	"fmt"
	"path/filepath"

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

	w.logger.Info().
		Str("worktree_id", cmd.worktreeID).
		Str("branch", worktreeToDelete.Branch).
		Str("path", worktreeToDelete.Path).
		Msg("deleting worktree")

	if err := git.RemoveWorktree(mainWorktree.Path, worktreeToDelete.Path); err != nil {
		cmd.responseChan <- err
		return err
	}

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
