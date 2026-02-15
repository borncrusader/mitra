# Mitra

Repos, Worktrees, Branches and Agents

## Quick Start

```bash
# Build
make build

# Run server
make run-server

# Run tests
make test
```

## Commands

```bash
# Start server
./bin/mitra serve

# Show config
./bin/mitra config show

# Generate default config
./bin/mitra config generate

# Add a repository
./bin/mitra repo add https://github.com/owner/repo

# List repositories
./bin/mitra repo list

# Create a worktree
./bin/mitra worktree add <repo-id> [branch[:parent-branch]]

# List worktrees
./bin/mitra worktree list [repo-id]
```

## Configuration

Config file: `~/.mitra/config.toml`

```toml
[server]
port = ":9999"
```

## Development

```bash
# Run in dev mode
make dev

# Run tests
make test

# Generate zsh completion
make completion

# Clean build artifacts
make clean
```
