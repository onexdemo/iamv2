// Package client implements a small typed REST client for the iam apiserver.
package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	v1 "github.com/onexdemo/iamv2/pkg/api/apiserver/v1"
)

// Config holds the connection settings for iamctl.
type Config struct {
	// Server is the base URL of iam-apiserver, for example
	// https://127.0.0.1:8443.
	Server string
	// Username and Password authenticate with HTTP Basic.
	Username string
	Password string
	// Token authenticates with a Bearer JWT.
	Token string
	// SecretID and SecretKey authenticate with an access key.
	SecretID  string
	SecretKey string

	Timeout               time.Duration
	InsecureSkipTLSVerify bool
	CertificateAuthority  string
}

// Client talks to the iam apiserver.
type Client struct {
	config Config
	http   *http.Client
}

// New creates a client from config.
func New(config Config) (*Client, error) {
	if config.Server == "" {
		return nil, fmt.Errorf("server address is required")
	}

	if _, err := url.Parse(config.Server); err != nil {
		return nil, fmt.Errorf("invalid server address: %w", err)
	}

	if config.Timeout <= 0 {
		config.Timeout = 10 * time.Second
	}

	tlsConfig := &tls.Config{InsecureSkipVerify: config.InsecureSkipTLSVerify} //nolint:gosec // controlled by an explicit CLI flag
	if config.CertificateAuthority != "" {
		ca, err := os.ReadFile(config.CertificateAuthority)
		if err != nil {
			return nil, fmt.Errorf("read certificate authority: %w", err)
		}

		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(ca) {
			return nil, fmt.Errorf("certificate authority contains no valid certificates")
		}

		tlsConfig.RootCAs = pool
	}

	return &Client{
		config: config,
		http: &http.Client{
			Timeout: config.Timeout,
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
		},
	}, nil
}

// Healthz checks whether the apiserver is reachable.
func (c *Client) Healthz(ctx context.Context) error {
	return c.do(ctx, http.MethodGet, "/healthz", nil, nil)
}

// GetUser returns a single user.
func (c *Client) GetUser(ctx context.Context, username string) (*v1.GetUserResponse, error) {
	var out v1.GetUserResponse

	if err := c.do(ctx, http.MethodGet, "/v1/users/"+url.PathEscape(username), nil, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

// ListUsers returns a page of users.
func (c *Client) ListUsers(ctx context.Context, offset, limit int64) (*v1.ListUsersResponse, error) {
	var out v1.ListUsersResponse

	path := fmt.Sprintf("/v1/users?offset=%d&limit=%d", offset, limit)
	if err := c.do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

// CreateUser registers a user.
func (c *Client) CreateUser(ctx context.Context, rq *v1.CreateUserRequest) (*v1.CreateUserResponse, error) {
	var out v1.CreateUserResponse

	if err := c.do(ctx, http.MethodPost, "/v1/users", rq, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

// UpdateUser updates a user profile.
func (c *Client) UpdateUser(ctx context.Context, rq *v1.UpdateUserRequest) (*v1.UpdateUserResponse, error) {
	var out v1.UpdateUserResponse

	if err := c.do(ctx, http.MethodPut, "/v1/users/"+url.PathEscape(rq.Name), rq, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

// DeleteUser deletes a user.
func (c *Client) DeleteUser(ctx context.Context, username string) error {
	return c.do(ctx, http.MethodDelete, "/v1/users/"+url.PathEscape(username), nil, nil)
}

// ChangePassword changes the authenticated user's password.
func (c *Client) ChangePassword(ctx context.Context, username string, rq *v1.ChangePasswordRequest) error {
	return c.do(ctx, http.MethodPut, "/v1/users/"+url.PathEscape(username)+"/change-password", rq, nil)
}

// GetSecret returns a single secret.
func (c *Client) GetSecret(ctx context.Context, name string) (*v1.GetSecretResponse, error) {
	var out v1.GetSecretResponse

	if err := c.do(ctx, http.MethodGet, "/v1/secrets/"+url.PathEscape(name), nil, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

// ListSecrets returns a page of secrets.
func (c *Client) ListSecrets(ctx context.Context, offset, limit int64) (*v1.ListSecretsResponse, error) {
	var out v1.ListSecretsResponse

	path := fmt.Sprintf("/v1/secrets?offset=%d&limit=%d", offset, limit)
	if err := c.do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

// CreateSecret creates an access key pair.
func (c *Client) CreateSecret(ctx context.Context, rq *v1.CreateSecretRequest) (*v1.CreateSecretResponse, error) {
	var out v1.CreateSecretResponse

	if err := c.do(ctx, http.MethodPost, "/v1/secrets", rq, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

// UpdateSecret updates a secret.
func (c *Client) UpdateSecret(ctx context.Context, rq *v1.UpdateSecretRequest) (*v1.UpdateSecretResponse, error) {
	var out v1.UpdateSecretResponse

	if err := c.do(ctx, http.MethodPut, "/v1/secrets/"+url.PathEscape(rq.Name), rq, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

// DeleteSecret deletes a secret.
func (c *Client) DeleteSecret(ctx context.Context, name string) error {
	return c.do(ctx, http.MethodDelete, "/v1/secrets/"+url.PathEscape(name), nil, nil)
}

// GetPolicy returns a single policy.
func (c *Client) GetPolicy(ctx context.Context, name string) (*v1.GetPolicyResponse, error) {
	var out v1.GetPolicyResponse

	if err := c.do(ctx, http.MethodGet, "/v1/policies/"+url.PathEscape(name), nil, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

// ListPolicies returns a page of policies.
func (c *Client) ListPolicies(ctx context.Context, offset, limit int64) (*v1.ListPoliciesResponse, error) {
	var out v1.ListPoliciesResponse

	path := fmt.Sprintf("/v1/policies?offset=%d&limit=%d", offset, limit)
	if err := c.do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

// CreatePolicy creates a policy.
func (c *Client) CreatePolicy(ctx context.Context, rq *v1.CreatePolicyRequest) (*v1.CreatePolicyResponse, error) {
	var out v1.CreatePolicyResponse

	if err := c.do(ctx, http.MethodPost, "/v1/policies", rq, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

// UpdatePolicy updates a policy.
func (c *Client) UpdatePolicy(ctx context.Context, rq *v1.UpdatePolicyRequest) (*v1.UpdatePolicyResponse, error) {
	var out v1.UpdatePolicyResponse

	if err := c.do(ctx, http.MethodPut, "/v1/policies/"+url.PathEscape(rq.Name), rq, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

// DeletePolicy deletes a policy.
func (c *Client) DeletePolicy(ctx context.Context, name string) error {
	return c.do(ctx, http.MethodDelete, "/v1/policies/"+url.PathEscape(name), nil, nil)
}

func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	var reader io.Reader

	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode request: %w", err)
		}

		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(c.config.Server, "/")+path, reader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	c.authorize(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return decodeError(resp)
	}

	if out == nil || resp.StatusCode == http.StatusNoContent {
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

func (c *Client) authorize(req *http.Request) {
	switch {
	case c.config.Token != "":
		req.Header.Set("Authorization", "Bearer "+c.config.Token)
	case c.config.SecretID != "":
		req.SetBasicAuth(c.config.SecretID, c.config.SecretKey)
	case c.config.Username != "":
		req.SetBasicAuth(c.config.Username, c.config.Password)
	}
}

func decodeError(resp *http.Response) error {
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))

	var payload struct {
		Reason  string `json:"reason"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(data, &payload); err == nil && payload.Message != "" {
		if payload.Reason != "" {
			return fmt.Errorf("%s: %s", payload.Reason, payload.Message)
		}

		return fmt.Errorf("%s", payload.Message)
	}

	return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
}
