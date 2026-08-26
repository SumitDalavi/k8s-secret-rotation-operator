package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Client is a minimal Vault HTTP client for secret renewal.
type Client struct {
	addr  string
	token string
	http  *http.Client
}

// NewClient creates a new Vault client from environment variables.
func NewClient() *Client {
	return &Client{
		addr:  getEnv("VAULT_ADDR", "http://vault:8200"),
		token: getEnv("VAULT_TOKEN", ""),
		http:  &http.Client{Timeout: 10 * time.Second},
	}
}

// RenewLease renews a Vault lease, extending its TTL.
func (c *Client) RenewLease(ctx context.Context, leaseID string, increment int) error {
	body := fmt.Sprintf(`{"lease_id":%q,"increment":%d}`, leaseID, increment)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPut, c.addr+"/v1/sys/leases/renew",
		strings.NewReader(body))
	req.Header.Set("X-Vault-Token", c.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("vault renew lease: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("vault renew lease: status %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

// GetSecret reads a KV v2 secret and returns its data map.
func (c *Client) GetSecret(ctx context.Context, path string) (map[string]string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet,
		c.addr+"/v1/"+path, nil)
	req.Header.Set("X-Vault-Token", c.token)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("vault get secret: %w", err)
	}
	defer resp.Body.Close()
	var result struct {
		Data struct {
			Data map[string]string `json:"data"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("vault decode: %w", err)
	}
	return result.Data.Data, nil
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
