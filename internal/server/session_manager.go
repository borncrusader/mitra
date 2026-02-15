package server

import (
	"context"
	"sync"

	"github.com/rs/zerolog"

	"mitra/internal/config"
)

type sessionCommand interface {
	Execute(sm *SessionManager) error
}

type SessionManager struct {
	logger      zerolog.Logger
	cfg         *config.Config
	state       *State
	commandChan chan sessionCommand
	sessions    map[string]bool
	mu          sync.RWMutex
}

func NewSessionManager(logger zerolog.Logger, cfg *config.Config, state *State) *SessionManager {
	return &SessionManager{
		logger: logger.With().
			Str("component", "session-manager").
			Logger(),
		cfg:         cfg,
		state:       state,
		commandChan: make(chan sessionCommand, 10),
		sessions:    make(map[string]bool),
	}
}

func (sm *SessionManager) Start(ctx context.Context) {
	sm.logger.Info().Msg("starting session manager")

	for {
		select {
		case <-ctx.Done():
			sm.logger.Info().Msg("stopping session manager due to context cancellation")
			return
		case cmd := <-sm.commandChan:
			sm.logger.Debug().Msg("processing session command")
			if err := cmd.Execute(sm); err != nil {
				sm.logger.Warn().
					Err(err).
					Msg("session command execution failed")
			}
		}
	}
}

func (sm *SessionManager) SendCommand(cmd sessionCommand) {
	sm.commandChan <- cmd
}
