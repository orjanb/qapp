package spotify

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const baseURL = "https://api.spotify.com/v1"

type Client struct {
	http *http.Client
}

func NewClient(httpClient *http.Client) *Client {
	return &Client{http: httpClient}
}

func (c *Client) get(ctx context.Context, path string, dst any) error {
	fullURL := baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent {
		return nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("spotify API error: %s (GET %s): %s", resp.Status, path, body)
	}
	return json.NewDecoder(resp.Body).Decode(dst)
}

func (c *Client) postEmpty(ctx context.Context, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Length", "0")
	return c.http.Do(req)
}
