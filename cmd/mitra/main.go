package main

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"mitra/internal/config"
	"mitra/internal/server"
	"mitra/internal/util"
)

var rootCmd = &cobra.Command{
	Use:   "mitra",
	Short: "Mitra - Repos, Worktrees, Branches and Agents",
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the server",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load config")
		}
		if err := server.Start(cfg); err != nil {
			log.Fatal().Err(err).Msg("failed to start server")
		}
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current config",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load config")
		}
		if err := config.Dump(cfg); err != nil {
			log.Fatal().Err(err).Msg("failed to dump config")
		}
	},
}

var configGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate default config file",
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.Generate(); err != nil {
			log.Fatal().Err(err).Msg("failed to generate config")
		}
		configPath, _ := config.Path()
		fmt.Printf("Config generated at %s\n", configPath)
	},
}

func init() {
	log.Logger = util.NewLogger(os.Stderr)

	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configGenerateCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(configCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
