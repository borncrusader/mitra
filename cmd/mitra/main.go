package main

import (
	"os"

	"github.com/rs/zerolog/log"

	"mitra/internal/cli"
	"mitra/internal/util"
)

func main() {
	log.Logger = util.NewLogger(os.Stderr)

	c, err := cli.New()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize mitra")
	}

	if err := c.Root().Execute(); err != nil {
		os.Exit(1)
	}
}
