package storage

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"

	"mitra/internal/proto"
	"mitra/internal/util"
)

type Repo struct {
	ID    string `toml:"id"`
	URL   string `toml:"url"`
	Host  string `toml:"host"`
	Owner string `toml:"owner"`
	Repo  string `toml:"repo"`
}

type RepoStorage struct {
	Repos []*Repo `toml:"repos"`
}

func toStorageRepo(pr *proto.Repo) *Repo {
	return &Repo{
		ID:    pr.Id,
		URL:   pr.Url,
		Host:  pr.Host,
		Owner: pr.Owner,
		Repo:  pr.Repo,
	}
}

func toProtoRepo(sr *Repo) *proto.Repo {
	return &proto.Repo{
		Id:    sr.ID,
		Url:   sr.URL,
		Host:  sr.Host,
		Owner: sr.Owner,
		Repo:  sr.Repo,
	}
}

func RepoPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".mitra", "repo.toml"), nil
}

func LoadRepos() ([]*proto.Repo, error) {
	repoPath, err := RepoPath()
	if err != nil {
		return nil, err
	}

	storage := &RepoStorage{
		Repos: []*Repo{},
	}

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return []*proto.Repo{}, nil
	}

	if _, err := toml.DecodeFile(repoPath, storage); err != nil {
		return nil, err
	}

	protoRepos := make([]*proto.Repo, len(storage.Repos))
	for i, r := range storage.Repos {
		protoRepos[i] = toProtoRepo(r)
	}

	return protoRepos, nil
}

func SaveRepos(repos []*proto.Repo) error {
	repoPath, err := RepoPath()
	if err != nil {
		return err
	}

	repoDir := filepath.Dir(repoPath)
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		return err
	}

	storageRepos := make([]*Repo, len(repos))
	for i, r := range repos {
		storageRepos[i] = toStorageRepo(r)
	}

	storage := &RepoStorage{
		Repos: storageRepos,
	}

	file, err := os.Create(repoPath)
	if err != nil {
		return err
	}
	defer util.DeferCheck(file.Close)

	encoder := toml.NewEncoder(file)
	return encoder.Encode(storage)
}
