// Package engine is a thin client for the gospel-engine admin API.
//
// Used by ibeco.me to mint, list, and revoke per-user gospel-engine tokens
// on behalf of authenticated ibeco.me users. Engine tokens are tagged with
// `external_user = "ibeco:<user_id>"` so we can scope listings/revokes to
// the right ibeco user without changing the engine's data model.
package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client talks to a gospel-engine instance using a service token.
type Client struct {
	BaseURL      string // e.g. https://engine.ibeco.me
	ServiceToken string // stdy_… token with admin scope
	HTTP         *http.Client
}

// New returns a new engine client. baseURL is trimmed of trailing slashes.
func New(baseURL, serviceToken string) *Client {
	return &Client{
		BaseURL:      strings.TrimRight(baseURL, "/"),
		ServiceToken: serviceToken,
		HTTP:         &http.Client{Timeout: 15 * time.Second},
	}
}

// Configured returns true if the client has both a base URL and a service token.
func (c *Client) Configured() bool {
	return c != nil && c.BaseURL != "" && c.ServiceToken != ""
}

// Token mirrors the engine's APIToken JSON shape.
type Token struct {
	ID           int64      `json:"id"`
	ExternalUser string     `json:"external_user,omitempty"`
	Name         string     `json:"name"`
	Prefix       string     `json:"prefix"`
	CreatedAt    time.Time  `json:"created_at"`
	LastUsed     *time.Time `json:"last_used,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	RateLimit    int        `json:"rate_limit"`
	Revoked      bool       `json:"revoked"`
}

// CreateTokenRequest matches the engine's createTokenReq.
type CreateTokenRequest struct {
	ExternalUser  string `json:"external_user"`
	Name          string `json:"name"`
	RateLimit     int    `json:"rate_limit"`
	ExpiresInDays int    `json:"expires_in_days"`
}

// CreateTokenResponse matches the engine's response.
type CreateTokenResponse struct {
	Token Token  `json:"token"`
	Raw   string `json:"raw"` // full secret, returned exactly once
}

// CreateToken mints a new engine token via /api/admin/tokens.
func (c *Client) CreateToken(req CreateTokenRequest) (*CreateTokenResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	resp, err := c.do("POST", "/api/admin/tokens", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return nil, errFromResp(resp)
	}
	var out CreateTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode create response: %w", err)
	}
	return &out, nil
}

// ListTokens returns ALL tokens on the engine. Callers should filter by
// external_user to scope to a specific ibeco user.
func (c *Client) ListTokens() ([]Token, error) {
	resp, err := c.do("GET", "/api/admin/tokens", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return nil, errFromResp(resp)
	}
	var out struct {
		Tokens []Token `json:"tokens"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode list response: %w", err)
	}
	return out.Tokens, nil
}

// RevokeToken revokes a token by its engine ID.
func (c *Client) RevokeToken(id int64) error {
	resp, err := c.do("DELETE", fmt.Sprintf("/api/admin/tokens/%d", id), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return errFromResp(resp)
	}
	return nil
}

func (c *Client) do(method, path string, body []byte) (*http.Response, error) {
	var rdr io.Reader
	if body != nil {
		rdr = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, c.BaseURL+path, rdr)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.ServiceToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.HTTP.Do(req)
}

func errFromResp(resp *http.Response) error {
	b, _ := io.ReadAll(resp.Body)
	msg := strings.TrimSpace(string(b))
	if msg == "" {
		msg = resp.Status
	}
	return fmt.Errorf("engine %s: %s", resp.Status, msg)
}
