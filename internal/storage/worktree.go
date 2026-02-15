package storage

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"

	"mitra/internal/proto"
)

type Worktree struct {
	ID           string  `toml:"id"`
	RepoID       string  `toml:"repo_id"`
	Branch       string  `toml:"branch"`
	Path         string  `toml:"path"`
	IsMain       bool    `toml:"is_main"`
	ParentBranch *string `toml:"parent_branch,omitempty"`
}

type WorktreeStorage struct {
	Worktrees []*Worktree `toml:"worktrees"`
}

func toStorageWorktree(pw *proto.Worktree) *Worktree {
	return &Worktree{
		ID:           pw.Id,
		RepoID:       pw.RepoId,
		Branch:       pw.Branch,
		Path:         pw.Path,
		IsMain:       pw.IsMain,
		ParentBranch: pw.ParentBranch,
	}
}

func toProtoWorktree(sw *Worktree) *proto.Worktree {
	return &proto.Worktree{
		Id:           sw.ID,
		RepoId:       sw.RepoID,
		Branch:       sw.Branch,
		Path:         sw.Path,
		IsMain:       sw.IsMain,
		ParentBranch: sw.ParentBranch,
	}
}

func WorktreePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".mitra", "worktree.toml"), nil
}

func LoadWorktrees() ([]*proto.Worktree, error) {
	worktreePath, err := WorktreePath()
	if err != nil {
		return nil, err
	}

	storage := &WorktreeStorage{
		Worktrees: []*Worktree{},
	}

	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		return []*proto.Worktree{}, nil
	}

	if _, err := toml.DecodeFile(worktreePath, storage); err != nil {
		return nil, err
	}

	protoWorktrees := make([]*proto.Worktree, len(storage.Worktrees))
	for i, w := range storage.Worktrees {
		protoWorktrees[i] = toProtoWorktree(w)
	}

	return protoWorktrees, nil
}

func SaveWorktrees(worktrees []*proto.Worktree) error {
	worktreePath, err := WorktreePath()
	if err != nil {
		return err
	}

	worktreeDir := filepath.Dir(worktreePath)
	if err := os.MkdirAll(worktreeDir, 0755); err != nil {
		return err
	}

	storageWorktrees := make([]*Worktree, len(worktrees))
	for i, w := range worktrees {
		storageWorktrees[i] = toStorageWorktree(w)
	}

	storage := &WorktreeStorage{
		Worktrees: storageWorktrees,
	}

	file, err := os.Create(worktreePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	return encoder.Encode(storage)
}
