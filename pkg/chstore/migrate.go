package chstore

import (
	"context"
	"embed"
	"fmt"
	"sort"
	"strings"
)

//go:embed schema/*.sql
var schemaFS embed.FS

// SplitStatements breaks a SQL file into individual statements on top-level
// semicolons. Line comments (`-- ...`) are stripped first so that a semicolon
// appearing inside a comment does not split a statement. ClickHouse's HTTP
// interface executes one statement per request, so multi-statement files must
// be split first. Pure.
func SplitStatements(sql string) []string {
	var clean []string
	for _, line := range strings.Split(sql, "\n") {
		if i := strings.Index(line, "--"); i >= 0 {
			line = line[:i]
		}
		clean = append(clean, line)
	}

	var out []string
	for _, raw := range strings.Split(strings.Join(clean, "\n"), ";") {
		if stmt := strings.TrimSpace(raw); stmt != "" {
			out = append(out, stmt)
		}
	}
	return out
}

// schemaStatements returns every embedded DDL statement in filename order. Pure
// apart from reading the embedded FS.
func schemaStatements() ([]string, error) {
	entries, err := schemaFS.ReadDir("schema")
	if err != nil {
		return nil, fmt.Errorf("chstore: read schema dir: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	var stmts []string
	for _, name := range names {
		b, err := schemaFS.ReadFile("schema/" + name)
		if err != nil {
			return nil, fmt.Errorf("chstore: read %s: %w", name, err)
		}
		stmts = append(stmts, SplitStatements(string(b))...)
	}
	return stmts, nil
}

// Migrate applies every embedded schema statement. All DDL is idempotent
// (CREATE ... IF NOT EXISTS), so Migrate is safe to run on every boot.
func (c *Client) Migrate(ctx context.Context) error {
	stmts, err := schemaStatements()
	if err != nil {
		return err
	}
	for i, stmt := range stmts {
		if err := c.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("chstore: migrate stmt %d: %w", i, err)
		}
	}
	return nil
}
