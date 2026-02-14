package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Server ServerConfig `toml:"server"`
	Repo   RepoConfig   `toml:"repo"`
}

type ServerConfig struct {
	Port     string `toml:"port"`
	GrpcPort string `toml:"grpc_port"`
}

type RepoConfig struct {
	Dir                  string `toml:"dir"`
	SyncIntervalSecs     int    `toml:"sync_interval_secs"`
}

func Default() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		Server: ServerConfig{
			Port:     ":9999",
			GrpcPort: ":9998",
		},
		Repo: RepoConfig{
			Dir:              filepath.Join(homeDir, "code", "work"),
			SyncIntervalSecs: 600,
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

	cfg := Default()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return cfg, nil
	}

	if _, err := toml.DecodeFile(configPath, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func Generate() error {
	configPath, err := Path()
	if err != nil {
		return err
	}

	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	cfg := Default()
	encoder := toml.NewEncoder(file)
	return encoder.Encode(cfg)
}

func Dump(cfg *Config) error {
	encoder := toml.NewEncoder(os.Stdout)
	return encoder.Encode(cfg)
}
