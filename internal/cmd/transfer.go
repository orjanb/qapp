package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var transferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "Transfer playback to the current active device (unstick queue)",
	RunE: func(cmd *cobra.Command, args []string) error {
		state, err := spotifyClient.GetPlaybackState(cmd.Context())
		if err != nil {
			return err
		}
		if state == nil {
			fmt.Println("Nothing is currently playing.")
			return nil
		}
		fmt.Printf("Current device: %s (%s)\n", state.Device.Name, state.Device.Type)
		fmt.Println("Transferring playback to this device…")
		if err := spotifyClient.TransferPlayback(cmd.Context(), state.Device.ID); err != nil {
			return err
		}
		fmt.Println("Done. Checking queue…")
		resp, err := spotifyClient.GetQueue(cmd.Context())
		if err != nil {
			return err
		}
		fmt.Printf("Queue items: %d\n", len(resp.Queue))
		for i, t := range resp.Queue {
			if i >= 5 {
				break
			}
			fmt.Printf("  %d. %s\n", i+1, t.Name)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(transferCmd)
}
