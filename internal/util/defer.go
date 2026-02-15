package util

import "github.com/rs/zerolog/log"

// DeferCheck handles deferred function errors by logging them
func DeferCheck(fn func() error) {
	if err := fn(); err != nil {
		log.Error().Err(err).Msg("deferred function error")
	}
}
