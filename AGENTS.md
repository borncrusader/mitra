# Mitra
Repos, Worktrees, Branches and Agents

# Important
1. Don't add any "Co-authored" text in commit messages
2. Don't add any unnecessary comments or logs unless I ask you to
3. When you update any CLI commands, update the CLAUDE.md, AGENTS.md and README.md files

# Project Structure
- `cmd/mitra/` - Main application entry point with CLI commands
- `internal/config/` - Configuration management (Go structs, not proto)
- `internal/proto/` - Protobuf definitions for gRPC services and data models
- `internal/server/` - HTTP + gRPC server implementation
- `internal/storage/` - Data persistence layer with TOML storage
- `internal/util/` - Utility functions (random name generation, etc.)

# Architecture
- **Config**: Regular Go structs with TOML serialization
- **Data Models**: Protobuf for gRPC services (MitraService with Repo and Worktree)
- **Server**: Dual HTTP (port 9999) + gRPC (port 9998)
- **Storage**: TOML files in `~/.mitra/` (config.toml, repo.toml, worktree.toml)
- **CLI**: Cobra commands that communicate with gRPC server
- **In-Memory Cache**: Repos and worktrees cached for fast access

# Make Targets
- `make build` - Build the binary to `bin/mitra`
- `make run-server` - Build and run the server
- `make dev` - Run in development mode
- `make test` - Run all tests
- `make lint` - Run golangci-lint
- `make protogen` - Regenerate proto and gRPC code
- `make completion` - Generate zsh completion file
- `make clean` - Remove build artifacts
- `make help` - Show available targets

# Configuration
Config file: `~/.mitra/config.toml`
- HTTP port: 9999 (default)
- gRPC port: 9998 (default)
- Repo directory: `~/code/work` (default)

# Data Storage
Location: `~/.mitra/`
- `config.toml` - Application configuration
- `repo.toml` - Repository list
- `worktree.toml` - Worktree list (linked to repos)

# Repositories
- Format: `host/owner/repo` (e.g., `github.com/owner/repo`)
- Supports: GitHub, GitLab, Bitbucket, custom git hosts
- URL formats: `https://host/owner/repo`, `host/owner/repo`, `git@host:owner/repo`
- CLI: `mitra repo add <url>`, `mitra repo list`

# Worktrees
- Automatically creates main/master worktree when repo is added
- Create new worktrees with `mitra worktree add <repo-id> [branch[:parent-branch]]`
- Branch syntax:
  - Omit branch: uses generated ID as branch name, main as parent
  - `feature`: uses 'feature' as branch, main as parent
  - `feature:develop`: uses 'feature' as branch, 'develop' as parent
  - `:develop`: uses generated ID as branch, 'develop' as parent
- All worktrees stored in repo directory: `~/code/work/owner/repo/branch`
- Only main worktree is monitored and synced automatically
- CLI: `mitra worktree add <repo-id> [branch[:parent-branch]]`, `mitra worktree list [repo-id]`

# Usage
```
Mitra - Repos, Worktrees, Branches and Agents

Usage:
  mitra [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  config      Manage configuration
  help        Help about any command
  repo        Manage repositories
  serve       Start the server
  worktree    Manage worktrees

Flags:
  -h, --help   help for mitra

Use "mitra [command] --help" for more information about a command.
```

# Commands
- `mitra serve` - Start HTTP + gRPC servers
- `mitra config show` - Show current config
- `mitra config generate` - Generate default config file
- `mitra repo add <url>` - Add a repository
- `mitra repo list` - List repos (TOML format)
- `mitra worktree add <repo-id> [branch[:parent-branch]]` - Create a new worktree
- `mitra worktree list [repo-id]` - List worktrees (TOML format)

# Development
- Proto regeneration: `make protogen` (includes gRPC code generation)
- Tests include validation for git URL parsing (GitHub, GitLab, etc.)
