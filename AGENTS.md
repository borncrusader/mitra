# Mitra
Repos, Worktrees, Branches and Agents

# Important
1. Don't add any "Co-authored" text in commit messages
2. Don't add any unnecessary comments or logs unless I ask you to

# Project Structure
- `cmd/mitra/` - Main application entry point
- `internal/config/` - Configuration management
- `internal/server/` - HTTP server implementation
- `internal/util/` - Utility functions (random name generation, etc.)

# Make Targets
- `make build` - Build the binary to `bin/mitra`
- `make run-server` - Build and run the server
- `make dev` - Run in development mode
- `make test` - Run all tests
- `make clean` - Remove build artifacts
- `make help` - Show available targets

# Configuration
Config file: `~/.mitra/config.toml`
Default port: 9999

# Usage
```
Mitra - Repos, Worktrees, Branches and Agents

Usage:
  mitra [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  config      Manage configuration
  help        Help about any command
  server      Start the server

Flags:
  -h, --help   help for mitra

Use "mitra [command] --help" for more information about a command.
```
