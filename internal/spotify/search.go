package spotify

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

type searchResponse struct {
	Tracks struct {
		Items []Track `json:"items"`
	} `json:"tracks"`
}

func (c *Client) SearchTracks(ctx context.Context, query string, limit int) ([]Track, error) {
	path := fmt.Sprintf("/search?q=%s&type=track&limit=%d",
		url.QueryEscape(query), limit)
	var result searchResponse
	if err := c.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return result.Tracks.Items, nil
}

func ArtistNames(track Track) string {
	names := make([]string, len(track.Artists))
	for i, a := range track.Artists {
		names[i] = a.Name
	}
	return strings.Join(names, ", ")
}
