package config

import (
	"errors"
	"os"
)

type Config struct {
	ClientID string
}

func Load() (*Config, error) {
	id := os.Getenv("SPOTIFY_CLIENT_ID")
	if id == "" {
		return nil, errors.New("SPOTIFY_CLIENT_ID environment variable is not set")
	}
	return &Config{ClientID: id}, nil
}
