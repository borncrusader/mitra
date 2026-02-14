package config

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

func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Port: ":9999",
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
