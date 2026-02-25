package spotify

import "context"

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

type CurrentlyPlayingResponse struct {
	IsPlaying bool   `json:"is_playing"`
	Item      *Track `json:"item"`
}

// GetCurrentlyPlaying returns what's playing now, or nil if nothing is playing.
func (c *Client) GetCurrentlyPlaying(ctx context.Context) (*CurrentlyPlayingResponse, error) {
	var result CurrentlyPlayingResponse
	if err := c.get(ctx, "/me/player/currently-playing", &result); err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, nil
	}
	return &result, nil
}
