package cmd

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/orjan/spotify/internal/lastfm"
	"github.com/orjan/spotify/internal/tui"
	"github.com/spf13/cobra"
)

var lastfmUser, lastfmAPIKey string

func init() {
	for _, cmd := range []*cobra.Command{tuiCmd, rootCmd} {
		cmd.Flags().StringVar(&lastfmUser, "lastfm-user", "", "Last.fm username")
		cmd.Flags().StringVar(&lastfmAPIKey, "lastfm-api-key", "", "Last.fm API key")
	}
}

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Interactive TUI",
	RunE:  runTUI,
}

func runTUI(cmd *cobra.Command, args []string) error {
	var lfm *lastfm.Client
	if lastfmUser != "" && lastfmAPIKey != "" {
		lfm = lastfm.NewClient(lastfmAPIKey, lastfmUser)
	}
	p := tea.NewProgram(tui.New(spotifyClient, lfm, cmd.Context()), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
