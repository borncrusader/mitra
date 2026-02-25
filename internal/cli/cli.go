package cli

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"mitra/internal"
	"mitra/internal/config"
	"mitra/internal/server"
	"mitra/internal/tui"
)

type Command struct {
	*mitra.Mitra
}

func New() (*Command, error) {
	m, err := mitra.New()
	if err != nil {
		return nil, err
	}
	return &Command{m}, nil
}

func (c *Command) Root() *cobra.Command {
	root := &cobra.Command{
		Use:   "mitra",
		Short: "Mitra - Repos, Worktrees, Branches and Agents",
		Run: func(cmd *cobra.Command, args []string) {
			if err := tui.RunDashboard(); err != nil {
				log.Fatal().Err(err).Msg("failed to run dashboard")
			}
		},
	}

	root.AddCommand(c.serveCmd())
	root.AddCommand(c.configCmd())
	root.AddCommand(c.repoCmd())
	root.AddCommand(c.worktreeCmd())
	root.AddCommand(c.sessionCmd())

	return root
}

func (c *Command) serveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the server",
		Run: func(cmd *cobra.Command, args []string) {
			if err := server.Start(c.Cfg); err != nil {
				log.Fatal().Err(err).Msg("failed to start server")
			}
		},
	}
}

func (c *Command) configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "config",
		Aliases: []string{"c", "cfg"},
		Short:   "Manage configuration",
	}
	cmd.AddCommand(c.configShowCmd())
	cmd.AddCommand(c.configGenerateCmd())
	return cmd
}

func (c *Command) configShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "show",
		Aliases: []string{"s"},
		Short:   "Show current config",
		Run: func(cmd *cobra.Command, args []string) {
			if err := config.Dump(c.Cfg); err != nil {
				log.Fatal().Err(err).Msg("failed to dump config")
			}
		},
	}
}

func (c *Command) configGenerateCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "generate",
		Aliases: []string{"g", "gen"},
		Short:   "Generate default config file",
		Run: func(cmd *cobra.Command, args []string) {
			if err := config.Generate(); err != nil {
				log.Fatal().Err(err).Msg("failed to generate config")
			}
			configPath, err := config.Path()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to get config path")
			}
			log.Info().Str("path", configPath).Msg("config generated")
		},
	}
}
