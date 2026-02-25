package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os/exec"
)

// GenerateState returns a random hex string suitable for use as an OAuth state parameter.
func GenerateState() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// WaitForCode starts a local server on :8888, opens authURL in the browser,
// and blocks until the OAuth callback delivers the authorization code.
// state must match the value embedded in authURL.
func WaitForCode(authURL, state string) (string, error) {
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	srv := &http.Server{Addr: "127.0.0.1:8888", Handler: mux}

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			http.Error(w, "state mismatch", http.StatusBadRequest)
			errCh <- fmt.Errorf("state mismatch in OAuth callback")
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			msg := r.URL.Query().Get("error")
			http.Error(w, "no code received", http.StatusBadRequest)
			errCh <- fmt.Errorf("no code in callback: %s", msg)
			return
		}
		fmt.Fprintln(w, "Authentication successful! You can close this tab.")
		codeCh <- code
		go func() { _ = srv.Shutdown(context.Background()) }()
	})

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	if err := exec.Command("open", authURL).Start(); err != nil {
		return "", fmt.Errorf("could not open browser: %w", err)
	}

	select {
	case code := <-codeCh:
		return code, nil
	case err := <-errCh:
		return "", err
	}
}
