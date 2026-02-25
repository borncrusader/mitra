# Mitra
Repos, Worktrees, Branches and Agents

# Important
1. Don't add any "Co-authored" text in commit messages
2. Don't add any unnecessary comments or logs unless I ask you to
3. When you update any CLI commands, update the CLAUDE.md, AGENTS.md and README.md files

# Project Structure
- `cmd/mitra/` - Main application entry point (minimal launcher)
- `internal/mitra.go` - Umbrella `Mitra` struct (holds config, created in main)
- `internal/cli/` - All CLI commands (cobra) structured as methods on `Command`
- `internal/agents/` - AI agent integrations (Claude settings, trust setup)
- `internal/config/` - Configuration management (Go structs, not proto)
- `internal/migration/` - Startup migrations and validations
- `internal/proto/` - Protobuf definitions for gRPC services and data models
- `internal/server/` - HTTP + gRPC server implementation
- `internal/storage/` - Data persistence layer with TOML storage
- `internal/tui/` - TUI dashboard and test harness
- `internal/util/` - Utility functions (random name generation, etc.)

# Architecture
- **Config**: Regular Go structs with TOML serialization
- **Data Models**: Protobuf for gRPC services (MitraService with Repo and Worktree)
- **Server**: Dual HTTP (port 9999) + gRPC (port 9998)
- **Storage**: TOML files in `~/.mitra/` (config.toml, repo.toml, worktree.toml)
- **CLI**: Cobra commands that communicate with gRPC server
- **In-Memory Cache**: Repos and worktrees cached for fast access
- **Migrations**: Automatic on startup - creates missing files, adds missing config fields, ensures main worktrees exist, optionally syncs untracked git worktrees

# Make Targets
- `make build` - Build the binary to `bin/mitra`
- `make run-server` - Build and run the server
- `make dev` - Run in development mode
- `make test` - Run all tests
- `make lint` - Run golangci-lint
- `make lint-fix` - Run golangci-lint with auto-fix
- `make protogen` - Regenerate proto and gRPC code
- `make completion` - Generate zsh completion file
- `make clean` - Remove build artifacts
- `make help` - Show available targets

# Configuration
Config file: `~/.mitra/config.toml`
- HTTP port: 9999 (default)
- gRPC port: 9998 (default)
- Repo directory: `~/code/work` (default)
- Branch prefix: `$USER/` (default) - all branches created with this prefix
- Sync untracked worktrees: `false` (default) - on startup, detect git worktrees not in config and add them automatically
- Session type: `tmux` (default) - type of session manager to use
- Session panes: Configure initial panes when creating tmux sessions
  - Format: `<window>.<pane>:<command with args>`
  - Pane can be 0 or 1 (left and right panes)
  - Example: `["0.0:nvim", "0.1:git status"]`
- Agents: Configure AI coding agents
  - Claude: Claude agent configuration
    - Enabled: `true` (default) - Enable Claude agent
    - Trust by default: `false` (default) - Trust Claude operations by default
  - Codex: `false` (default) - Enable Codex agent

# Data Storage
Location: `~/.mitra/`
- `config.toml` - Application configuration
- `repo.toml` - Repository list
- `worktree.toml` - Worktree list (linked to repos)

# Repositories
- Format: `host/owner/repo` (e.g., `github.com/owner/repo`)
- Supports: GitHub, GitLab, Bitbucket, custom git hosts
- URL formats: `https://host/owner/repo`, `host/owner/repo`, `git@host:owner/repo`
- Delete safety: Cannot delete if non-main worktrees exist or main worktree is not clean
- CLI: `mitra repo add <url>`, `mitra repo list`, `mitra repo delete <repo-id>`

# Worktrees
- Automatically creates main/master worktree when repo is added
- Create new worktrees with `mitra worktree add [repo-id] [branch[:parent-branch]]`
- Interactive mode: `mitra worktree add` (no arguments) shows TUI to select repo
- Delete worktrees with `mitra worktree delete [worktree-id]`
- Interactive delete: `mitra worktree delete` (no arguments) shows TUI to select worktree
- Protection: Cannot delete main worktrees or worktrees that are not clean; if not clean, prompts `[y/N]` to force delete
- Clean checks: No uncommitted changes, stashes, merge/rebase in progress, unpushed commits, etc.
- Stale worktrees: If git reports "is not a working tree", mitra removes it from config anyway
- Claude files: When a new worktree is created, `.claude/settings.local.json` and `CLAUDE.local.md` are symlinked from the main worktree if they exist
- Branch prefix: All branches created with configured prefix (default: `$USER/`)
- Branch syntax:
  - Omit branch: uses generated ID as branch name, main as parent
  - `feature`: creates branch `$USER/feature` from main
  - `feature:develop`: creates branch `$USER/feature` from develop
  - `:develop`: uses generated ID as branch name from develop
- All worktrees stored in repo directory: `~/code/work/owner/repo/branch`
- Only main worktree is monitored and synced automatically
- CLI: `mitra worktree add [repo-id] [branch[:parent-branch]]`, `mitra worktree list [repo-id]`, `mitra worktree delete [worktree-id]`

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
  session     Manage sessions
  tui         TUI utilities
  worktree    Manage worktrees

Flags:
  -h, --help   help for mitra

Use "mitra [command] --help" for more information about a command.
```

# Dashboard
- Running `mitra` without any command launches the TUI dashboard
- Shows ASCII art logo and project tagline
- Press `q` or `Ctrl+C` to quit the dashboard
- All subcommands work as before

# Commands
- `mitra` - Launch TUI dashboard (default command)
- `mitra serve` - Start HTTP + gRPC servers
- `mitra config show` (aliases: `c s`) - Show current config
- `mitra config generate` (aliases: `c g`, `c gen`) - Generate default config file
- `mitra repo add <url>` (aliases: `r a`) - Add a repository
- `mitra repo list` (aliases: `r l`, `r ls`) - List repos (TOML format)
- `mitra repo delete [repo-id]` (aliases: `r d`, `r del`, `r rm`) - Delete a repository (interactive TUI if no args, requires all non-main worktrees deleted and main worktree clean)
- `mitra worktree add [repo-id] [branch[:parent-branch]]` (aliases: `w a`, `wt a`) - Create a new worktree (interactive TUI if no args)
- `mitra worktree list [repo-id]` (aliases: `w l`, `w ls`) - List worktrees (TOML format)
- `mitra worktree delete [worktree-id]` (aliases: `w d`, `w del`, `w rm`) - Delete a worktree (interactive TUI if no args; prompts to force delete if not clean)
- `mitra session list` (aliases: `s l`, `s ls`, `sess l`) - List all tmux sessions grouped by repo
- `mitra session attach [session-id]` (aliases: `s a`, `s at`, `sess a`) - Attach to a tmux session (interactive TUI if no args)
- `mitra session delete [session-id]` (aliases: `s d`, `s del`, `sess d`) - Kill a tmux session (interactive TUI if no args)
- `mitra tui test` - Run a full-screen TUI test (for UI development)

# Development
- Proto regeneration: `make protogen` (includes gRPC code generation)
- Tests include validation for git URL parsing (GitHub, GitLab, etc.)
