package spotify

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

var ErrNoActiveDevice = errors.New("no active Spotify device found. Open Spotify and start playing something first")

func (c *Client) AddToQueue(ctx context.Context, trackURI string) error {
	path := "/me/player/queue?uri=" + url.QueryEscape(trackURI)
	resp, err := c.postEmpty(ctx, path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200, 204:
		return nil
	case 404:
		return ErrNoActiveDevice
	default:
		return fmt.Errorf("spotify API error: %s", resp.Status)
	}
}
