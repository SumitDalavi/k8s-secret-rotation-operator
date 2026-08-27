package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestNewClient_Defaults(t *testing.T) {
	os.Unsetenv("VAULT_ADDR")
	os.Unsetenv("VAULT_TOKEN")

	c := NewClient()
	if c.addr != "http://vault:8200" {
		t.Errorf("expected default addr, got %s", c.addr)
	}
	if c.token != "" {
		t.Errorf("expected empty default token")
	}
}

func TestNewClient_FromEnv(t *testing.T) {
	os.Setenv("VAULT_ADDR", "http://my-vault:8200")
	os.Setenv("VAULT_TOKEN", "my-token")
	defer os.Unsetenv("VAULT_ADDR")
	defer os.Unsetenv("VAULT_TOKEN")

	c := NewClient()
	if c.addr != "http://my-vault:8200" {
		t.Errorf("expected http://my-vault:8200, got %s", c.addr)
	}
	if c.token != "my-token" {
		t.Errorf("expected my-token, got %s", c.token)
	}
}

func TestRenewLease_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/sys/leases/renew" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("X-Vault-Token") != "test-token" {
			t.Errorf("unexpected token: %s", r.Header.Get("X-Vault-Token"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := &Client{addr: srv.URL, token: "test-token", http: &http.Client{}}
	err := c.RenewLease(context.Background(), "lease-123", 3600)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRenewLease_Failure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("permission denied"))
	}))
	defer srv.Close()

	c := &Client{addr: srv.URL, token: "bad-token", http: &http.Client{}}
	err := c.RenewLease(context.Background(), "lease-123", 3600)
	if err == nil {
		t.Fatal("expected error for non-200 status")
	}
}

func TestRenewLease_NetworkError(t *testing.T) {
	c := &Client{addr: "http://localhost:19999", token: "tok", http: &http.Client{}}
	err := c.RenewLease(context.Background(), "lease-123", 3600)
	if err == nil {
		t.Fatal("expected error for connection failure")
	}
}

func TestGetSecret_Success(t *testing.T) {
	expected := map[string]string{"username": "admin", "password": "secret123"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		resp := map[string]interface{}{
			"data": map[string]interface{}{
				"data": expected,
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := &Client{addr: srv.URL, token: "tok", http: &http.Client{}}
	data, err := c.GetSecret(context.Background(), "secret/myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data["username"] != "admin" {
		t.Errorf("expected username admin, got %s", data["username"])
	}
}

func TestGetSecret_NetworkError(t *testing.T) {
	c := &Client{addr: "http://localhost:19999", token: "tok", http: &http.Client{}}
	_, err := c.GetSecret(context.Background(), "secret/myapp")
	if err == nil {
		t.Fatal("expected error for connection failure")
	}
}

func TestGetSecret_DecodeError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not-json"))
	}))
	defer srv.Close()

	c := &Client{addr: srv.URL, token: "tok", http: &http.Client{}}
	_, err := c.GetSecret(context.Background(), "secret/myapp")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_KEY", "test-value")
	defer os.Unsetenv("TEST_KEY")

	val := getEnv("TEST_KEY", "default")
	if val != "test-value" {
		t.Errorf("expected test-value, got %s", val)
	}

	val = getEnv("NONEXISTENT_KEY", "default")
	if val != "default" {
		t.Errorf("expected default, got %s", val)
	}
}
