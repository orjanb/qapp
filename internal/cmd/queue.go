package cmd

import (
	"errors"
	"fmt"
	"strings"

	spotifyclient "github.com/orjan/spotify/internal/spotify"
	"github.com/spf13/cobra"
)

var queueCmd = &cobra.Command{
	Use:   "queue <track-id>",
	Short: "Add a track to the Spotify play queue",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		uri := id
		if !strings.HasPrefix(id, "spotify:track:") {
			uri = "spotify:track:" + id
		}

		err := spotifyClient.AddToQueue(cmd.Context(), uri)
		if err != nil {
			if errors.Is(err, spotifyclient.ErrNoActiveDevice) {
				fmt.Println("No active Spotify device found. Open Spotify and start playing something first.")
				return nil
			}
			return err
		}
		fmt.Println("Added to queue.")
		return nil
	},
}
