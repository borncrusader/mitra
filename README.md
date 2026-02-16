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
# Launch TUI dashboard (default)
./bin/mitra

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

# Create a worktree (interactive TUI if no args)
./bin/mitra worktree add [repo-id] [branch[:parent-branch]]
# Examples:
#   mitra worktree add                       # interactive repo selection
#   mitra worktree add repo-123              # uses generated ID, main as parent
#   mitra worktree add repo-123 feature      # 'feature' from main
#   mitra worktree add repo-123 feature:dev  # 'feature' from dev
#   mitra worktree add repo-123 :dev         # generated ID from dev

# List worktrees
./bin/mitra worktree list [repo-id]

# Delete a worktree (interactive TUI if no args)
./bin/mitra worktree delete [worktree-id]
# Examples:
#   mitra worktree delete                    # interactive worktree selection
#   mitra worktree delete worktree-id-123    # delete specific worktree

# List sessions
./bin/mitra session list

# Attach to a session (interactive TUI if no args)
./bin/mitra session attach [session-id]
# Examples:
#   mitra session attach                     # interactive session selection
#   mitra session attach session-id-123      # attach to specific session
```

## Configuration

Config file: `~/.mitra/config.toml`

```toml
[server]
port = ":9999"
grpc_port = ":9998"

[repo]
dir = "~/code/work"
sync_interval_secs = 600
branch_prefix = "username/"  # Default: $USER/

[session]
type = "tmux"  # Default: tmux
panes = [
  "0.0:nvim",
  "0.1:git status"
]

[agents]
codex = false   # Default: false

[agents.claude]
enabled = true            # Default: true
trust_by_default = false  # Default: false
```

### Session Configuration
- **type**: Type of session manager to use (default: `tmux`)
- **panes**: Configure initial panes when creating tmux sessions
  - Format: `<window>.<pane>:<command with args>`
  - Pane can be 0 or 1 (left and right panes)
  - Commands are executed automatically when the session is created

### Agents Configuration
- **claude**: Claude AI coding agent configuration
  - **enabled**: Enable Claude agent (default: `true`)
  - **trust_by_default**: Trust Claude operations by default (default: `false`)
- **codex**: Enable Codex AI coding agent (default: `false`)

## Development

```bash
# Run in dev mode
make dev

# Run tests
make test

# Run linter
make lint

# Run linter with auto-fix
make lint-fix

# Generate zsh completion
make completion

# Clean build artifacts
make clean
```

### Installing golangci-lint

```bash
# macOS
brew install golangci-lint

# Or using go install
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```
