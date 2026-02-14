package storage

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"mitra/internal/proto"
)

type RepoStorage struct {
	Repos []*proto.Repo `toml:"repos"`
}

func RepoPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".mitra", "repo.toml"), nil
}

func LoadRepos() (*RepoStorage, error) {
	repoPath, err := RepoPath()
	if err != nil {
		return nil, err
	}

	storage := &RepoStorage{
		Repos: []*proto.Repo{},
	}

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return storage, nil
	}

	if _, err := toml.DecodeFile(repoPath, storage); err != nil {
		return nil, err
	}

	return storage, nil
}

func SaveRepos(storage *RepoStorage) error {
	repoPath, err := RepoPath()
	if err != nil {
		return err
	}

	repoDir := filepath.Dir(repoPath)
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		return err
	}

	file, err := os.Create(repoPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	return encoder.Encode(storage)
}
