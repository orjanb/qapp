# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build

```sh
go build -o qapp ./cmd/qapp/
```

Run directly without building:

```sh
go run ./cmd/qapp/ [args]
```

There are no tests in this project.

## Architecture

The app is a Spotify terminal client with two modes: a Bubbletea TUI (default) and individual CLI commands via Cobra.

**Entry point:** `cmd/qapp/main.go` → `internal/cmd.Execute()`

**`internal/cmd/`** — Cobra command tree. `root.go` wires everything together:
- A `PersistentPreRunE` hook initializes the Spotify client before any command runs (except `auth`)
- A `savingTokenSource` wraps the oauth2 token source so refreshed tokens are automatically persisted to disk
- The root command's `RunE` launches the TUI

**`internal/spotify/`** — Spotify Web API client. `client.go` provides two low-level helpers (`get`, `postEmpty`); the other files add domain methods (search, queue, player, models).

**`internal/auth/`** — OAuth2 PKCE flow. `pkce.go` generates verifier/challenge, `callback.go` runs a local HTTP server on port 8888 to receive the redirect, `token.go` saves/loads tokens from `~/.config/qapp-cli/token.json`.

**`internal/config/`** — Reads/writes `~/.config/qapp-cli/config.json` (stores `client_id`).

**`internal/tui/`** — Single Bubbletea `Model` with three views (`viewSearch` → `viewResults` → `viewQueue`). All async Spotify calls are issued as `tea.Cmd` functions that return typed messages (e.g. `searchResultsMsg`, `queueLoadedMsg`).

## Runtime config

Credentials are stored outside the repo:
- `~/.config/qapp-cli/config.json` — Spotify client ID
- `~/.config/qapp-cli/token.json` — OAuth2 token

Required OAuth scopes: `user-read-playback-state`, `user-modify-playback-state`, `user-read-currently-playing`. Spotify Premium is required for queue/skip operations.
