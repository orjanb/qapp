package cmd

import (
	"fmt"

	"github.com/orjan/spotify/internal/auth"
	"github.com/orjan/spotify/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with Spotify via browser",
	RunE: func(cmd *cobra.Command, args []string) error {
		clientID, _ := cmd.Flags().GetString("client-id")

		var cfg *config.Config
		if clientID != "" {
			cfg = &config.Config{ClientID: clientID}
			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}
		} else {
			var err error
			cfg, err = config.Load()
			if err != nil {
				return err
			}
		}

		verifier, err := auth.GenerateVerifier()
		if err != nil {
			return fmt.Errorf("generating PKCE verifier: %w", err)
		}
		challenge := auth.GenerateChallenge(verifier)

		state, err := auth.GenerateState()
		if err != nil {
			return fmt.Errorf("generating state: %w", err)
		}

		oc := buildOAuthConfig(cfg.ClientID)
		authURL := oc.AuthCodeURL(state,
			oauth2.SetAuthURLParam("code_challenge", challenge),
			oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		)

		fmt.Println("Opening browser for Spotify authentication...")
		code, err := auth.WaitForCode(authURL, state)
		if err != nil {
			return fmt.Errorf("waiting for auth code: %w", err)
		}

		ctx := cmd.Context()
		token, err := oc.Exchange(ctx, code,
			oauth2.SetAuthURLParam("code_verifier", verifier),
		)
		if err != nil {
			return fmt.Errorf("exchanging code for token: %w", err)
		}

		if err := auth.SaveToken(token); err != nil {
			return fmt.Errorf("saving token: %w", err)
		}

		fmt.Println("Authentication successful! Token saved.")
		return nil
	},
}

func init() {
	authCmd.Flags().String("client-id", "", "Spotify app client ID (only needed on first run)")
}
