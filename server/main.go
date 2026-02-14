package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
}

func startServer(config *Config) {
	http.HandleFunc("/", helloHandler)

	log.Printf("Server starting on http://localhost%s", config.Server.Port)

	if err := http.ListenAndServe(config.Server.Port, nil); err != nil {
		log.Fatal(err)
	}
}

func dumpConfig(config *Config) {
	encoder := toml.NewEncoder(os.Stdout)
	if err := encoder.Encode(config); err != nil {
		log.Fatal(err)
	}
}

var rootCmd = &cobra.Command{
	Use:   "mitra",
	Short: "Mitra - Repos, Worktrees, Branches and Agents",
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the server",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := LoadConfig()
		if err != nil {
			log.Fatal(err)
		}
		startServer(config)
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
		config, err := LoadConfig()
		if err != nil {
			log.Fatal(err)
		}
		dumpConfig(config)
	},
}

var configGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate default config file",
	Run: func(cmd *cobra.Command, args []string) {
		if err := GenerateConfig(); err != nil {
			log.Fatal(err)
		}
		configPath, _ := ConfigPath()
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
