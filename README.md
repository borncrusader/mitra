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
./bin/mitra server

# Show config
./bin/mitra config show

# Generate default config
./bin/mitra config generate
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

# Clean build artifacts
make clean
```
