package chstore

import "testing"

func TestParseConfig_WithCredentials(t *testing.T) {
	cfg, err := ParseConfig("http://liverack:secret@localhost:8123", "liverack")
	if err != nil {
		t.Fatalf("ParseConfig: %v", err)
	}
	if cfg.Endpoint != "http://localhost:8123" {
		t.Errorf("endpoint = %q, want http://localhost:8123", cfg.Endpoint)
	}
	if cfg.User != "liverack" || cfg.Password != "secret" {
		t.Errorf("creds = %q/%q, want liverack/secret", cfg.User, cfg.Password)
	}
	if cfg.Database != "liverack" {
		t.Errorf("database = %q, want liverack", cfg.Database)
	}
}

func TestParseConfig_NoCredentials(t *testing.T) {
	cfg, err := ParseConfig("http://localhost:8123", "analytics")
	if err != nil {
		t.Fatalf("ParseConfig: %v", err)
	}
	if cfg.User != "" || cfg.Password != "" {
		t.Errorf("expected empty creds, got %q/%q", cfg.User, cfg.Password)
	}
}

func TestParseConfig_MissingHost(t *testing.T) {
	if _, err := ParseConfig("not-a-url", "db"); err == nil {
		t.Fatal("expected error for url missing host")
	}
}

func TestSplitStatements(t *testing.T) {
	sql := `-- a comment
CREATE TABLE a (x Int32) ENGINE = MergeTree ORDER BY x;

-- another comment
CREATE TABLE b (y Int32) ENGINE = MergeTree ORDER BY y;
`
	stmts := SplitStatements(sql)
	if len(stmts) != 2 {
		t.Fatalf("got %d statements, want 2: %#v", len(stmts), stmts)
	}
	for _, s := range stmts {
		if got := s[:12]; got != "CREATE TABLE" {
			t.Errorf("statement should start with CREATE TABLE, got %q", got)
		}
	}
}

func TestSplitStatements_TrailingEmpty(t *testing.T) {
	if got := SplitStatements(";\n\n  ;  "); got != nil {
		t.Errorf("expected no statements, got %#v", got)
	}
}

func TestSchemaStatements_Embedded(t *testing.T) {
	stmts, err := schemaStatements()
	if err != nil {
		t.Fatalf("schemaStatements: %v", err)
	}
	if len(stmts) == 0 {
		t.Fatal("expected embedded schema statements, got none")
	}
}
