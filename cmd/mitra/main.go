package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"server/internal/config"
	"server/internal/server"
)

var rootCmd = &cobra.Command{
	Use:   "mitra",
	Short: "Mitra - Repos, Worktrees, Branches and Agents",
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the server",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			log.Fatal(err)
		}
		if err := server.Start(cfg); err != nil {
			log.Fatal(err)
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
			log.Fatal(err)
		}
		if err := config.Dump(cfg); err != nil {
			log.Fatal(err)
		}
	},
}

var configGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate default config file",
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.Generate(); err != nil {
			log.Fatal(err)
		}
		configPath, _ := config.Path()
		fmt.Printf("Config generated at %s\n", configPath)
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configGenerateCmd)
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(configCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
