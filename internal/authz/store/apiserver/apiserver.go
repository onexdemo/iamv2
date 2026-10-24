// Package apiserver implements the authz store against the iam-apiserver REST
// cache endpoints.
package apiserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/onexdemo/iamv2/internal/authz/store"
)

// datastore talks to iam-apiserver over HTTP.
type datastore struct {
	baseURL  string
	username string
	password string
	client   *http.Client
}

// New creates a store backed by iam-apiserver.
func New(baseURL, username, password string, timeout time.Duration) store.Factory {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	return &datastore{
		baseURL:  strings.TrimRight(baseURL, "/"),
		username: username,
		password: password,
		client:   &http.Client{Timeout: timeout},
	}
}

// Secrets returns the access-key store.
func (d *datastore) Secrets() store.SecretStore {
	return &secrets{datastore: d}
}

// Policies returns the policy store.
func (d *datastore) Policies() store.PolicyStore {
	return &policies{datastore: d}
}

// get performs an authenticated GET and decodes the JSON response into out.
func (d *datastore) get(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.SetBasicAuth(d.username, d.password)
	req.Header.Set("Accept", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("send snapshot request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("snapshot request failed with status %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode snapshot response: %w", err)
	}

	return nil
}
