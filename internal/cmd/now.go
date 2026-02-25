package cmd

import (
	"fmt"

	spotifyclient "github.com/orjan/spotify/internal/spotify"
	"github.com/spf13/cobra"
)

var nowCmd = &cobra.Command{
	Use:   "now",
	Short: "Show the currently playing track",
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := spotifyClient.GetCurrentlyPlaying(cmd.Context())
		if err != nil {
			return err
		}
		if result == nil {
			fmt.Println("Nothing is currently playing.")
			return nil
		}
		status := "⏸"
		if result.IsPlaying {
			status = "▶"
		}
		fmt.Printf("%s  %s\n   %s · %s\n",
			status,
			result.Item.Name,
			spotifyclient.ArtistNames(*result.Item),
			result.Item.Album.Name,
		)
		return nil
	},
}
