package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"mitra/internal/proto"
)

func Default() *proto.Config {
	homeDir, _ := os.UserHomeDir()
	return &proto.Config{
		Server: &proto.ServerConfig{
			Port:     ":9999",
			GrpcPort: ":9998",
		},
		Repo: &proto.RepoConfig{
			Dir: filepath.Join(homeDir, "code", "work"),
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

func Load() (*proto.Config, error) {
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

func Dump(cfg *proto.Config) error {
	encoder := toml.NewEncoder(os.Stdout)
	return encoder.Encode(cfg)
}
