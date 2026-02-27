# qapp

A terminal client for Spotify. Includes an interactive TUI and a handful of CLI commands.

## Setup

### 1. Create a Spotify app

1. Go to [developer.spotify.com/dashboard](https://developer.spotify.com/dashboard) and create an app
2. In the app settings, add `http://127.0.0.1:8888/callback` as a Redirect URI
3. Copy your Client ID

### 2. Build

Requires [Go](https://go.dev) 1.21+.

```sh
go build -o qapp ./cmd/qapp/
```

**Cross-compile for Windows:**

```sh
GOOS=windows GOARCH=amd64 go build -o qapp.exe ./cmd/qapp/
```

On Windows, set the environment and run from [Windows Terminal](https://aka.ms/terminal).

### 3. Authenticate

```sh
qapp auth <client-id>
```

This opens a browser for the OAuth flow and saves the token to `~/.config/qapp-cli/token.json`. The client ID is saved to `~/.config/qapp-cli/config.json` — you only need to pass it once.

Re-run `qapp auth` (without arguments) any time you need to refresh your token.

## Usage

### TUI (default)

```sh
qapp
```

Launches the interactive TUI.

| View | Key | Action |
|------|-----|--------|
| Search | type | fill search input |
| Search | enter | search |
| Search | tab | open queue |
| Search | esc / ctrl+c | quit |
| Results | ↑ ↓ | navigate |
| Results | enter | add to queue |
| Results | / | back to search |
| Results | tab | open queue |
| Results | esc / ctrl+c | quit |
| Queue | ↑ ↓ | navigate |
| Queue | n | skip to next track |
| Queue | r | refresh |
| Queue | tab | back |
| Queue | esc / ctrl+c | quit |

### CLI commands

```sh
qapp now                   # show currently playing track
qapp search <query>        # search for tracks (prints track IDs)
qapp queue <track-id>      # add a track to the queue
qapp auth                  # re-authenticate
qapp auth <client-id>      # first-time setup
```

## Notes

- **Spotify Premium** is required for playback control (skip, add to queue). Search and `now` work on free accounts.
- The TUI requires a terminal with ANSI support. On Windows, use Windows Terminal.
