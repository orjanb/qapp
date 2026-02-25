package cmd

import (
	"fmt"
	"strings"

	spotifyclient "github.com/orjan/spotify/internal/spotify"
	"github.com/spf13/cobra"
)

var searchLimit int

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for tracks on Spotify",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.Join(args, " ")
		ctx := cmd.Context()

		tracks, err := spotifyClient.SearchTracks(ctx, query, searchLimit)
		if err != nil {
			return err
		}
		if len(tracks) == 0 {
			fmt.Println("No tracks found.")
			return nil
		}
		for i, t := range tracks {
			fmt.Printf("%d. %s - %s  [%s]\n", i+1, t.Name, spotifyclient.ArtistNames(t), t.ID)
		}
		return nil
	},
}

func init() {
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "l", 5, "Number of results to return (max 10)")
}
