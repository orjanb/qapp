package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/orjan/spotify/internal/auth"
	"github.com/orjan/spotify/internal/config"
	spotifyclient "github.com/orjan/spotify/internal/spotify"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var (
	spotifyClient *spotifyclient.Client
	oauthConfig   *oauth2.Config
)

var rootCmd = &cobra.Command{
	Use:   "spotify",
	Short: "A CLI for interacting with Spotify",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == "auth" {
			return nil
		}
		return initClient(cmd.Context())
	},
	RunE: runTUI,
}

func Execute() error {
	return rootCmd.ExecuteContext(context.Background())
}

func buildOAuthConfig(clientID string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:    clientID,
		RedirectURL: "http://127.0.0.1:8888/callback",
		Scopes: []string{
			"user-read-playback-state",
			"user-modify-playback-state",
			"user-read-currently-playing",
		},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.spotify.com/authorize",
			TokenURL: "https://accounts.spotify.com/api/token",
		},
	}
}

func initClient(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	oauthConfig = buildOAuthConfig(cfg.ClientID)

	token, err := auth.LoadToken()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("not authenticated — run `spotify auth` first (see https://developer.spotify.com/documentation/web-api/concepts/apps)")
		}
		return fmt.Errorf("loading token: %w", err)
	}

	ts := &savingTokenSource{
		base: oauthConfig.TokenSource(ctx, token),
	}
	httpClient := oauth2.NewClient(ctx, ts)
	spotifyClient = spotifyclient.NewClient(httpClient)
	return nil
}

// savingTokenSource wraps a TokenSource and persists refreshed tokens to disk.
type savingTokenSource struct {
	base  oauth2.TokenSource
	saved string
}

func (s *savingTokenSource) Token() (*oauth2.Token, error) {
	token, err := s.base.Token()
	if err != nil {
		return nil, err
	}
	if token.AccessToken != s.saved {
		s.saved = token.AccessToken
		if saveErr := auth.SaveToken(token); saveErr != nil {
			fmt.Fprintf(os.Stderr, "warning: could not save refreshed token: %v\n", saveErr)
		}
	}
	return token, nil
}

func init() {
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(queueCmd)
	rootCmd.AddCommand(tuiCmd)
	rootCmd.AddCommand(nowCmd)
}

