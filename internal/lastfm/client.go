package lastfm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type TrackStats struct {
	UserPlayCount int
	Listeners     int
}

type Client struct {
	apiKey   string
	username string
	http     *http.Client
}

func NewClient(apiKey, username string) *Client {
	return &Client{
		apiKey:   apiKey,
		username: username,
		http:     &http.Client{},
	}
}

func (c *Client) GetTrackInfo(ctx context.Context, artist, track string) (*TrackStats, error) {
	endpoint := "https://ws.audioscrobbler.com/2.0/"
	params := url.Values{
		"method":   {"track.getInfo"},
		"api_key":  {c.apiKey},
		"artist":   {artist},
		"track":    {track},
		"username": {c.username},
		"format":   {"json"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Track struct {
			Listeners    string `json:"listeners"`
			UserPlayCount string `json:"userplaycount"`
		} `json:"track"`
		Error   int    `json:"error"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if result.Error == 6 {
		return nil, fmt.Errorf("track not found: %s", result.Message)
	}
	if result.Error != 0 {
		return nil, fmt.Errorf("last.fm error %d: %s", result.Error, result.Message)
	}

	listeners, _ := strconv.Atoi(result.Track.Listeners)
	userPlayCount, _ := strconv.Atoi(result.Track.UserPlayCount)
	return &TrackStats{
		UserPlayCount: userPlayCount,
		Listeners:     listeners,
	}, nil
}
