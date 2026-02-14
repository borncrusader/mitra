# Mitra
Repos, Worktrees, Branches and Agents

# Important
1. Don't add any "Co-authored" text in commit messages
2. Don't add any unnecessary comments or logs unless I ask you to

# Project Structure
- `cmd/mitra/` - Main application entry point with CLI commands
- `internal/config/` - Configuration management (Go structs, not proto)
- `internal/proto/` - Protobuf definitions for gRPC services and data models
- `internal/server/` - HTTP + gRPC server implementation
- `internal/storage/` - Data persistence layer with TOML storage
- `internal/util/` - Utility functions (random name generation, etc.)

# Architecture
- **Config**: Regular Go structs with TOML serialization
- **Data Models**: Protobuf for gRPC services (Repo, etc.)
- **Server**: Dual HTTP (port 9999) + gRPC (port 9998)
- **Storage**: TOML files in `~/.mitra/` (config.toml, repo.toml)
- **CLI**: Cobra commands that communicate with gRPC server

# Make Targets
- `make build` - Build the binary to `bin/mitra`
- `make run-server` - Build and run the server
- `make dev` - Run in development mode
- `make test` - Run all tests
- `make protogen` - Regenerate proto and gRPC code
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
- `repo.toml` - Repository list (lowercase fields)

# Repositories
- Format: `host/owner/repo` (e.g., `github.com/owner/repo`)
- Supports: GitHub, GitLab, Bitbucket, custom git hosts
- URL formats: `https://host/owner/repo`, `host/owner/repo`
- CLI: `mitra repo add <url>`, `mitra repo list`

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
  server      Start the server

Flags:
  -h, --help   help for mitra

Use "mitra [command] --help" for more information about a command.
```

# Commands
- `mitra server` - Start HTTP + gRPC servers
- `mitra config show` - Show current config
- `mitra config generate` - Generate default config file
- `mitra repo add <url>` - Add a repository
- `mitra repo list` - List repos (TOML format)

# Development
- Proto regeneration: `make protogen` (includes gRPC code generation)
- Tests include validation for git URL parsing (GitHub, GitLab, etc.)
