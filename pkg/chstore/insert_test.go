package chstore

import (
	"strings"
	"testing"
)

func TestEncodeJSONEachRow(t *testing.T) {
	rows := []map[string]any{
		{"sku": "LR-1", "qty": 2},
		{"sku": "LR-2", "qty": 5},
	}
	b, err := EncodeJSONEachRow(rows)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	lines := strings.Split(strings.TrimRight(string(b), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2: %q", len(lines), string(b))
	}
	if !strings.Contains(lines[0], `"sku":"LR-1"`) {
		t.Errorf("line 0 missing sku: %s", lines[0])
	}
}

func TestEncodeJSONEachRow_Empty(t *testing.T) {
	b, err := EncodeJSONEachRow(nil)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if len(b) != 0 {
		t.Errorf("expected empty output, got %q", string(b))
	}
}
