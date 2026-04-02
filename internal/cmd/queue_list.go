package cmd

import (
	"fmt"

	spotifyclient "github.com/orjan/spotify/internal/spotify"
	"github.com/spf13/cobra"
)

var queueListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show the current play queue",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := spotifyClient.GetQueue(cmd.Context())
		if err != nil {
			return err
		}
		if resp.CurrentlyPlaying != nil {
			fmt.Printf("▶ %s — %s (id: %s)\n", resp.CurrentlyPlaying.Name, spotifyclient.ArtistNames(*resp.CurrentlyPlaying), resp.CurrentlyPlaying.ID)
		} else {
			fmt.Println("▶ (nothing)")
		}
		fmt.Printf("Queue items: %d\n", len(resp.Queue))
		for i, t := range resp.Queue {
			if i >= 10 {
				fmt.Printf("  ... and %d more\n", len(resp.Queue)-10)
				break
			}
			fmt.Printf("%2d. %s — %s\n", i+1, t.Name, spotifyclient.ArtistNames(t))
		}
		return nil
	},
}

func init() {
	queueCmd.AddCommand(queueListCmd)
}
