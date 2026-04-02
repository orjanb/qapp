package spotify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

type QueueResponse struct {
	CurrentlyPlaying *Track  `json:"currently_playing"`
	Queue            []Track `json:"queue"`
}

func (c *Client) GetQueue(ctx context.Context) (*QueueResponse, error) {
	var result QueueResponse
	if err := c.get(ctx, "/me/player/queue", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

type PlaybackStateResponse struct {
	IsPlaying  bool    `json:"is_playing"`
	ProgressMs int     `json:"progress_ms"`
	Item       *Track  `json:"item"`
	Device     Device  `json:"device"`
}

func (c *Client) SkipToNext(ctx context.Context) error {
	resp, err := c.postEmpty(ctx, "/me/player/next")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 204 && resp.StatusCode != 200 {
		return fmt.Errorf("spotify API error: %s", resp.Status)
	}
	return nil
}

// TransferPlayback transfers playback to the given device.
func (c *Client) TransferPlayback(ctx context.Context, deviceID string) error {
	body, _ := json.Marshal(map[string]any{
		"device_ids": []string{deviceID},
	})
	resp, err := c.put(ctx, "/me/player", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 204 && resp.StatusCode != 200 {
		return fmt.Errorf("spotify API error: %s", resp.Status)
	}
	return nil
}

// GetPlaybackState returns the full playback state including device info, or nil if nothing is playing.
func (c *Client) GetPlaybackState(ctx context.Context) (*PlaybackStateResponse, error) {
	var result PlaybackStateResponse
	if err := c.get(ctx, "/me/player", &result); err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, nil
	}
	return &result, nil
}
