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

All commands support short aliases for faster typing.

```bash
# Launch TUI dashboard (default)
./bin/mitra

# Start server
./bin/mitra serve

# Show config (aliases: c s)
./bin/mitra config show
./bin/mitra c s

# Generate default config (aliases: c g, c gen)
./bin/mitra config generate
./bin/mitra c g

# Add a repository (aliases: r a)
./bin/mitra repo add https://github.com/owner/repo
./bin/mitra r a https://github.com/owner/repo

# List repositories (aliases: r l, r ls)
./bin/mitra repo list
./bin/mitra r l

# Create a worktree (interactive TUI if no args) (aliases: w a, wt a)
./bin/mitra worktree add [repo-id] [branch[:parent-branch]]
./bin/mitra w a [repo-id] [branch[:parent-branch]]
# Examples:
#   mitra w a                                # interactive repo selection
#   mitra w a repo-123                       # uses generated ID, main as parent
#   mitra w a repo-123 feature               # 'feature' from main
#   mitra w a repo-123 feature:dev           # 'feature' from dev
#   mitra w a repo-123 :dev                  # generated ID from dev

# List worktrees (aliases: w l, w ls)
./bin/mitra worktree list [repo-id]
./bin/mitra w l [repo-id]

# Delete a worktree (interactive TUI if no args) (aliases: w d, w del, w rm)
./bin/mitra worktree delete [worktree-id]
./bin/mitra w d [worktree-id]
# Examples:
#   mitra w d                                # interactive worktree selection
#   mitra w d worktree-id-123                # delete specific worktree

# List sessions (aliases: s l, s ls, sess l)
./bin/mitra session list
./bin/mitra s l

# Attach to a session (interactive TUI if no args) (aliases: s a, s at, sess a)
./bin/mitra session attach [session-id]
./bin/mitra s a [session-id]
# Examples:
#   mitra s a                                # interactive session selection
#   mitra s a session-id-123                 # attach to specific session
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
