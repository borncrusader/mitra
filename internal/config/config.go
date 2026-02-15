package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"

	"mitra/internal/util"
)

type Config struct {
	Server  ServerConfig  `toml:"server"`
	Repo    RepoConfig    `toml:"repo"`
	Session SessionConfig `toml:"session"`
	Agents  AgentsConfig  `toml:"agents"`
}

type ServerConfig struct {
	Port     string `toml:"port"`
	GrpcPort string `toml:"grpc_port"`
}

type RepoConfig struct {
	Dir              string `toml:"dir"`
	SyncIntervalSecs int    `toml:"sync_interval_secs"`
	BranchPrefix     string `toml:"branch_prefix"`
}

type SessionConfig struct {
	Type  string   `toml:"type"`
	Panes []string `toml:"panes"`
}

type AgentsConfig struct {
	Claude ClaudeAgentConfig `toml:"claude"`
	Codex  bool              `toml:"codex"`
}

type ClaudeAgentConfig struct {
	Enabled         bool `toml:"enabled"`
	TrustByDefault  bool `toml:"trust_by_default"`
}

func Default() *Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	username := os.Getenv("USER")
	if username == "" {
		username = "user"
	}

	return &Config{
		Server: ServerConfig{
			Port:     ":9999",
			GrpcPort: ":9998",
		},
		Repo: RepoConfig{
			Dir:              filepath.Join(homeDir, "code", "work"),
			SyncIntervalSecs: 600,
			BranchPrefix:     username + "/",
		},
		Session: SessionConfig{
			Type: "tmux",
			Panes: []string{
				"0.0:claude",
				"0.1:nvim",
			},
		},
		Agents: AgentsConfig{
			Claude: ClaudeAgentConfig{
				Enabled:        true,
				TrustByDefault: false,
			},
			Codex: false,
		},
	}
}

func Path() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".mitra", "config.toml"), nil
}

func Load() (*Config, error) {
	configPath, err := Path()
	if err != nil {
		return nil, err
	}

	var cfg Config

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &cfg, nil
	}

	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func Generate() error {
	configPath, err := Path()
	if err != nil {
		return err
	}

	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config file already exists at %s", configPath)
	}

	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer util.DeferCheck(file.Close)

	cfg := Default()
	encoder := toml.NewEncoder(file)
	return encoder.Encode(cfg)
}

func Dump(cfg *Config) error {
	encoder := toml.NewEncoder(os.Stdout)
	return encoder.Encode(cfg)
}
