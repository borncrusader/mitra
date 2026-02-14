package main

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Server ServerConfig `toml:"server"`
}

type ServerConfig struct {
	Port string `toml:"port"`
}

func LoadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(homeDir, ".mitra", "config.toml")

	config := &Config{
		Server: ServerConfig{
			Port: ":9999",
		},
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return config, nil
	}

	if _, err := toml.DecodeFile(configPath, config); err != nil {
		return nil, err
	}

	return config, nil
}
