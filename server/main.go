package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/BurntSushi/toml"
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

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: mitra <command> [subcommand]")
		fmt.Println("Commands:")
		fmt.Println("  server         - Start the server")
		fmt.Println("  config show    - Show current config")
		fmt.Println("  config generate - Generate default config file")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "server":
		config, err := LoadConfig()
		if err != nil {
			log.Fatal(err)
		}
		startServer(config)
	case "config":
		if len(os.Args) < 3 {
			fmt.Println("Usage: mitra config <subcommand>")
			fmt.Println("Subcommands:")
			fmt.Println("  show     - Show current config")
			fmt.Println("  generate - Generate default config file")
			os.Exit(1)
		}

		subcommand := os.Args[2]
		switch subcommand {
		case "show":
			config, err := LoadConfig()
			if err != nil {
				log.Fatal(err)
			}
			dumpConfig(config)
		case "generate":
			if err := GenerateConfig(); err != nil {
				log.Fatal(err)
			}
			configPath, _ := ConfigPath()
			fmt.Printf("Config generated at %s\n", configPath)
		default:
			fmt.Printf("Unknown config subcommand: %s\n", subcommand)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}
