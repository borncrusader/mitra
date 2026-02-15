package server

import (
	"mitra/internal/tmux"
)

type createSessionCmd struct {
	worktreeID   string
	path         string
	responseChan chan error
}

type killSessionCmd struct {
	worktreeID   string
	responseChan chan error
}

func (cmd *createSessionCmd) Execute(sm *SessionManager) error {
	sm.mu.RLock()
	exists := sm.sessions[cmd.worktreeID]
	sm.mu.RUnlock()

	if exists {
		sm.logger.Info().
			Str("session", cmd.worktreeID).
			Msg("tmux session already tracked")
		cmd.responseChan <- nil
		return nil
	}

	sessionExists, err := tmux.SessionExists(cmd.worktreeID)
	if err != nil {
		cmd.responseChan <- err
		return err
	}

	if sessionExists {
		sm.mu.Lock()
		sm.sessions[cmd.worktreeID] = true
		sm.mu.Unlock()

		sm.logger.Info().
			Str("session", cmd.worktreeID).
			Msg("tmux session already exists")
		cmd.responseChan <- nil
		return nil
	}

	if err := tmux.CreateSession(cmd.worktreeID, cmd.path, sm.cfg.Session.Panes); err != nil {
		cmd.responseChan <- err
		return err
	}

	sm.mu.Lock()
	sm.sessions[cmd.worktreeID] = true
	sm.mu.Unlock()

	sm.logger.Info().
		Str("session", cmd.worktreeID).
		Str("path", cmd.path).
		Msg("tmux session created")

	cmd.responseChan <- nil
	return nil
}

func (cmd *killSessionCmd) Execute(sm *SessionManager) error {
	sessionExists, err := tmux.SessionExists(cmd.worktreeID)
	if err != nil {
		cmd.responseChan <- err
		return err
	}

	if !sessionExists {
		sm.mu.Lock()
		delete(sm.sessions, cmd.worktreeID)
		sm.mu.Unlock()

		cmd.responseChan <- nil
		return nil
	}

	if err := tmux.KillSession(cmd.worktreeID); err != nil {
		cmd.responseChan <- err
		return err
	}

	sm.mu.Lock()
	delete(sm.sessions, cmd.worktreeID)
	sm.mu.Unlock()

	sm.logger.Info().
		Str("session", cmd.worktreeID).
		Msg("tmux session killed")

	cmd.responseChan <- nil
	return nil
}

func NewCreateSessionCommand(worktreeID, path string) *createSessionCmd {
	return &createSessionCmd{
		worktreeID:   worktreeID,
		path:         path,
		responseChan: make(chan error, 1),
	}
}

func NewKillSessionCommand(worktreeID string) *killSessionCmd {
	return &killSessionCmd{
		worktreeID:   worktreeID,
		responseChan: make(chan error, 1),
	}
}

type addSessionsCmd struct {
	responseChan chan error
}

func (cmd *addSessionsCmd) Execute(sm *SessionManager) error {
	worktrees := sm.state.GetWorktrees("")

	sm.logger.Info().
		Int("worktrees", len(worktrees)).
		Msg("adding tmux sessions")

	for _, wt := range worktrees {
		sessionExists, err := tmux.SessionExists(wt.Id)
		if err != nil {
			sm.logger.Warn().
				Err(err).
				Str("session", wt.Id).
				Msg("failed to check tmux session")
			continue
		}

		if sessionExists {
			sm.mu.Lock()
			sm.sessions[wt.Id] = true
			sm.mu.Unlock()

			sm.logger.Debug().
				Str("session", wt.Id).
				Msg("tmux session already exists")
			continue
		}

		if err := tmux.CreateSession(wt.Id, wt.Path, sm.cfg.Session.Panes); err != nil {
			sm.logger.Warn().
				Err(err).
				Str("session", wt.Id).
				Str("path", wt.Path).
				Msg("failed to create tmux session")
		} else {
			sm.mu.Lock()
			sm.sessions[wt.Id] = true
			sm.mu.Unlock()

			sm.logger.Info().
				Str("session", wt.Id).
				Str("path", wt.Path).
				Msg("tmux session created")
		}
	}

	sm.logger.Info().
		Int("sessions", len(sm.sessions)).
		Msg("tmux sessions added")

	cmd.responseChan <- nil
	return nil
}

func NewAddSessionsCommand() *addSessionsCmd {
	return &addSessionsCmd{
		responseChan: make(chan error, 1),
	}
}
