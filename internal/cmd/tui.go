package cmd

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/orjan/spotify/internal/tui"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Interactive TUI",
	RunE:  runTUI,
}

func runTUI(cmd *cobra.Command, args []string) error {
	p := tea.NewProgram(tui.New(spotifyClient, cmd.Context()), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
