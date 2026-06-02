// Package chstore is a thin HTTP client for the ClickHouse analytics warehouse.
// It exposes Exec for DDL/inserts and Query for reads, plus a schema migrator.
// Business logic (DSN parsing, statement splitting, row insertion encoding) is
// pure so it can be unit-tested without a running ClickHouse.
package chstore

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Config holds resolved ClickHouse connection settings.
type Config struct {
	Endpoint string // scheme://host:port (no credentials, no path)
	User     string
	Password string
	Database string
}

// ParseConfig derives a Config from a CLICKHOUSE_URL (creds may be embedded as
// userinfo) and an explicit database name. Pure.
func ParseConfig(rawURL, database string) (Config, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return Config{}, fmt.Errorf("chstore: parse url: %w", err)
	}
	if u.Host == "" {
		return Config{}, fmt.Errorf("chstore: url missing host: %q", rawURL)
	}
	cfg := Config{
		Endpoint: u.Scheme + "://" + u.Host,
		Database: database,
	}
	if u.User != nil {
		cfg.User = u.User.Username()
		cfg.Password, _ = u.User.Password()
	}
	return cfg, nil
}

// Client talks to ClickHouse over HTTP.
type Client struct {
	cfg  Config
	http *http.Client
}

// New builds a Client with a default 30s timeout.
func New(cfg Config) *Client {
	return &Client{cfg: cfg, http: &http.Client{Timeout: 30 * time.Second}}
}

// reqURL builds the request URL, optionally selecting the configured database.
func (c *Client) reqURL(withDB bool) string {
	q := url.Values{}
	if withDB && c.cfg.Database != "" {
		q.Set("database", c.cfg.Database)
	}
	if len(q) == 0 {
		return c.cfg.Endpoint + "/"
	}
	return c.cfg.Endpoint + "/?" + q.Encode()
}

func (c *Client) do(ctx context.Context, sql string, withDB bool) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.reqURL(withDB), strings.NewReader(sql))
	if err != nil {
		return nil, fmt.Errorf("chstore: build request: %w", err)
	}
	if c.cfg.User != "" {
		req.SetBasicAuth(c.cfg.User, c.cfg.Password)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("chstore: do request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("chstore: status %d: %s", resp.StatusCode, bytes.TrimSpace(body))
	}
	return body, nil
}

// Exec runs a statement (DDL or insert) that returns no rows.
func (c *Client) Exec(ctx context.Context, sql string) error {
	_, err := c.do(ctx, sql, true)
	return err
}

// Query runs a read and returns the raw response body. Callers append a FORMAT
// clause (e.g. "FORMAT JSON") to control encoding.
func (c *Client) Query(ctx context.Context, sql string) ([]byte, error) {
	return c.do(ctx, sql, true)
}

// Ping verifies connectivity.
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.do(ctx, "SELECT 1", false)
	return err
}
