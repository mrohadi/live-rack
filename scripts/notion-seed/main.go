// notion-seed creates the live-rack · Backlog database in Notion and seeds
// all phase tickets (P0–P11). Idempotent: skips tickets whose Ticket ID
// already exists in the DB.
//
// Usage:
//
//	NOTION_API_KEY=secret_xxx NOTION_PARENT_PAGE_ID=<page-id> go run ./scripts/notion-seed
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

const notionVersion = "2022-06-28"
const baseURL = "https://api.notion.com/v1"

func main() {
	apiKey := mustEnv("NOTION_API_KEY")
	parentPageID := mustEnv("NOTION_PARENT_PAGE_ID")

	c := &client{key: apiKey, http: &http.Client{Timeout: 30 * time.Second}}

	slog.Info("creating database", "parent", parentPageID)
	dbID, err := c.createDatabase(parentPageID)
	if err != nil {
		slog.Error("create database", "err", err)
		os.Exit(1)
	}
	slog.Info("database created", "id", dbID)

	existing, err := c.existingTicketIDs(dbID)
	if err != nil {
		slog.Error("query existing tickets", "err", err)
		os.Exit(1)
	}
	slog.Info("existing tickets", "count", len(existing))

	created, skipped := 0, 0
	for _, t := range allTickets() {
		if existing[t.ID] {
			skipped++
			continue
		}
		if err := c.createPage(dbID, t); err != nil {
			slog.Error("create page", "ticket", t.ID, "err", err)
			os.Exit(1)
		}
		slog.Info("created", "ticket", t.ID, "title", t.Title)
		created++
		time.Sleep(334 * time.Millisecond) // Notion rate limit: 3 req/s
	}

	slog.Info("done", "created", created, "skipped", skipped)
}

// ── Notion client ────────────────────────────────────────────────────────────

type client struct {
	key  string
	http *http.Client
}

func (c *client) do(method, path string, body any) (map[string]any, error) {
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, baseURL+path, r)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.key)
	req.Header.Set("Notion-Version", notionVersion)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("notion %s %s → %d: %s", method, path, resp.StatusCode, raw)
	}
	var out map[string]any
	return out, json.Unmarshal(raw, &out)
}

func (c *client) createDatabase(parentPageID string) (string, error) {
	payload := map[string]any{
		"parent": map[string]any{"type": "page_id", "page_id": parentPageID},
		"title": []map[string]any{
			{"type": "text", "text": map[string]any{"content": "live-rack · Backlog"}},
		},
		"properties": dbSchema(),
	}
	res, err := c.do("POST", "/databases", payload)
	if err != nil {
		return "", err
	}
	id, _ := res["id"].(string)
	return id, nil
}

func (c *client) existingTicketIDs(dbID string) (map[string]bool, error) {
	out := map[string]bool{}
	cursor := ""
	for {
		payload := map[string]any{"page_size": 100}
		if cursor != "" {
			payload["start_cursor"] = cursor
		}
		res, err := c.do("POST", "/databases/"+dbID+"/query", payload)
		if err != nil {
			return nil, err
		}
		results, _ := res["results"].([]any)
		for _, r := range results {
			page, _ := r.(map[string]any)
			props, _ := page["properties"].(map[string]any)
			ticketProp, _ := props["Ticket ID"].(map[string]any)
			richTexts, _ := ticketProp["rich_text"].([]any)
			if len(richTexts) > 0 {
				rt, _ := richTexts[0].(map[string]any)
				text, _ := rt["plain_text"].(string)
				if text != "" {
					out[text] = true
				}
			}
		}
		hasMore, _ := res["has_more"].(bool)
		if !hasMore {
			break
		}
		cursor, _ = res["next_cursor"].(string)
	}
	return out, nil
}

func (c *client) createPage(dbID string, t ticket) error {
	props := map[string]any{
		"Name":      titleProp(t.Title),
		"Ticket ID": richTextProp(t.ID),
		"Phase":     selectProp(t.Phase),
		"Epic":      selectProp(t.Epic),
		"Status":    statusProp("Backlog"),
		"Priority":  selectProp(t.Priority),
		"Type":      selectProp(t.Type),
		"Estimate":  numberProp(t.Estimate),
	}
	_, err := c.do("POST", "/pages", map[string]any{
		"parent":     map[string]any{"database_id": dbID},
		"properties": props,
	})
	return err
}

// ── Property helpers ─────────────────────────────────────────────────────────

func titleProp(s string) map[string]any {
	return map[string]any{"title": []map[string]any{{"text": map[string]any{"content": s}}}}
}

func richTextProp(s string) map[string]any {
	return map[string]any{"rich_text": []map[string]any{{"text": map[string]any{"content": s}}}}
}

func selectProp(s string) map[string]any {
	return map[string]any{"select": map[string]any{"name": s}}
}

func statusProp(s string) map[string]any {
	return map[string]any{"status": map[string]any{"name": s}}
}

func numberProp(n int) map[string]any {
	return map[string]any{"number": n}
}

// ── Database schema ───────────────────────────────────────────────────────────

func dbSchema() map[string]any {
	phases := selectOptions("P0", "P1", "P2", "P3", "P4", "P5", "P6", "P7", "P8", "P9", "P10", "P11")
	epics := selectOptions("Foundations", "Map", "Scanner", "Inventory", "Tasks", "Pipelines", "POS", "Analytics", "Signals", "RBAC", "Marketplace", "Hardening")
	priorities := selectOptions("P0-blocker", "P1-high", "P2-med", "P3-low")
	types := selectOptions("feat", "chore", "spike", "bug", "docs")
	tests := selectOptions("unit", "repo", "service", "contract", "e2e", "load")

	return map[string]any{
		"Name":                map[string]any{"title": map[string]any{}},
		"Ticket ID":           map[string]any{"rich_text": map[string]any{}},
		"Phase":               map[string]any{"select": map[string]any{"options": phases}},
		"Epic":                map[string]any{"select": map[string]any{"options": epics}},
		"Status":              map[string]any{"status": map[string]any{"options": statusOptions()}},
		"Priority":            map[string]any{"select": map[string]any{"options": priorities}},
		"Type":                map[string]any{"select": map[string]any{"options": types}},
		"Estimate":            map[string]any{"number": map[string]any{"format": "number"}},
		"Owner":               map[string]any{"people": map[string]any{}},
		"Branch":              map[string]any{"url": map[string]any{}},
		"PR":                  map[string]any{"url": map[string]any{}},
		"Tests":               map[string]any{"multi_select": map[string]any{"options": tests}},
		"Acceptance criteria": map[string]any{"rich_text": map[string]any{}},
		"Linked design":       map[string]any{"url": map[string]any{}},
	}
}

func selectOptions(names ...string) []map[string]any {
	opts := make([]map[string]any, len(names))
	for i, n := range names {
		opts[i] = map[string]any{"name": n}
	}
	return opts
}

func statusOptions() []map[string]any {
	return []map[string]any{
		{"name": "Backlog"},
		{"name": "Todo"},
		{"name": "In progress"},
		{"name": "In review"},
		{"name": "Blocked"},
		{"name": "Done"},
	}
}

// ── Ticket data ───────────────────────────────────────────────────────────────

type ticket struct {
	ID       string
	Title    string
	Phase    string
	Epic     string
	Priority string
	Type     string
	Estimate int
}

func allTickets() []ticket {
	return []ticket{
		// P0 · Foundations
		{ID: "LR-001", Title: "monorepo + tooling (Go, pnpm, Turbo, Make targets)", Phase: "P0", Epic: "Foundations", Priority: "P0-blocker", Type: "feat", Estimate: 3},
		{ID: "LR-002", Title: "GitHub repo + branch protection + PR template", Phase: "P0", Epic: "Foundations", Priority: "P0-blocker", Type: "chore", Estimate: 1},
		{ID: "LR-003", Title: "docker-compose dev stack (PG+TS, NATS, ClickHouse, Redis, MinIO)", Phase: "P0", Epic: "Foundations", Priority: "P0-blocker", Type: "feat", Estimate: 2},
		{ID: "LR-004", Title: "Terraform skeleton (modules: network, db, cache, bus, observability)", Phase: "P0", Epic: "Foundations", Priority: "P1-high", Type: "feat", Estimate: 5},
		{ID: "LR-005", Title: "Clerk integration + tenant model (org → store hierarchy)", Phase: "P0", Epic: "Foundations", Priority: "P0-blocker", Type: "feat", Estimate: 5},
		{ID: "LR-006", Title: "Tailwind tokens extracted from styles.css → tokens.css", Phase: "P0", Epic: "Foundations", Priority: "P1-high", Type: "feat", Estimate: 2},
		{ID: "LR-007", Title: "app shell (sidebar, topbar, mobile tabbar) from shell.jsx", Phase: "P0", Epic: "Foundations", Priority: "P1-high", Type: "feat", Estimate: 3},
		{ID: "LR-008", Title: "OTel + Loki + Tempo + Prom + Grafana wiring", Phase: "P0", Epic: "Foundations", Priority: "P1-high", Type: "feat", Estimate: 3},
		{ID: "LR-009", Title: "CI gates (lint, type, test, coverage, security)", Phase: "P0", Epic: "Foundations", Priority: "P0-blocker", Type: "chore", Estimate: 3},
		{ID: "LR-010", Title: "pre-commit hooks + conventional commits enforcement", Phase: "P0", Epic: "Foundations", Priority: "P1-high", Type: "chore", Estimate: 2},
		{ID: "LR-011", Title: "create Notion DB and seed tickets", Phase: "P0", Epic: "Foundations", Priority: "P2-med", Type: "chore", Estimate: 2},

		// P1 · Zones & Map
		{ID: "LR-101", Title: "spike: Konva vs SVG perf for 500+ zones", Phase: "P1", Epic: "Map", Priority: "P0-blocker", Type: "spike", Estimate: 2},
		{ID: "LR-102", Title: "zones table + RLS + sqlc queries", Phase: "P1", Epic: "Map", Priority: "P0-blocker", Type: "feat", Estimate: 3},
		{ID: "LR-103", Title: "zones REST API + OpenAPI", Phase: "P1", Epic: "Map", Priority: "P0-blocker", Type: "feat", Estimate: 3},
		{ID: "LR-104", Title: "Konva editor (drag/resize/snap, grid, multi-select)", Phase: "P1", Epic: "Map", Priority: "P0-blocker", Type: "feat", Estimate: 8},
		{ID: "LR-105", Title: "zone constraints (allowed item types, capacity)", Phase: "P1", Epic: "Map", Priority: "P1-high", Type: "feat", Estimate: 3},
		{ID: "LR-106", Title: "heat / items / zones view toggles", Phase: "P1", Epic: "Map", Priority: "P1-high", Type: "feat", Estimate: 2},
		{ID: "LR-107", Title: "zone detail sidebar", Phase: "P1", Epic: "Map", Priority: "P1-high", Type: "feat", Estimate: 2},
		{ID: "LR-108", Title: "e2e: zone create → drag → save", Phase: "P1", Epic: "Map", Priority: "P2-med", Type: "chore", Estimate: 2},

		// P2 · Scanner
		{ID: "LR-201", Title: "spike: WebHID Zebra DataWedge profile", Phase: "P2", Epic: "Scanner", Priority: "P0-blocker", Type: "spike", Estimate: 3},
		{ID: "LR-202", Title: "scanner PWA shell + camera (zxing)", Phase: "P2", Epic: "Scanner", Priority: "P0-blocker", Type: "feat", Estimate: 5},
		{ID: "LR-203", Title: "validation engine (category, capacity, dwell, dual-scan)", Phase: "P2", Epic: "Scanner", Priority: "P0-blocker", Type: "feat", Estimate: 5},
		{ID: "LR-204", Title: "offline IndexedDB queue + sync on reconnect", Phase: "P2", Epic: "Scanner", Priority: "P1-high", Type: "feat", Estimate: 5},
		{ID: "LR-205", Title: "scan_events Timescale hypertable + repo", Phase: "P2", Epic: "Scanner", Priority: "P0-blocker", Type: "feat", Estimate: 3},
		{ID: "LR-206", Title: "scan WS push to Map/Inventory clients", Phase: "P2", Epic: "Scanner", Priority: "P1-high", Type: "feat", Estimate: 3},
		{ID: "LR-207", Title: "e2e: mis-scan blocking with reason", Phase: "P2", Epic: "Scanner", Priority: "P2-med", Type: "chore", Estimate: 2},

		// P3 · Inventory
		{ID: "LR-301", Title: "items master + item_locations realtime", Phase: "P3", Epic: "Inventory", Priority: "P0-blocker", Type: "feat", Estimate: 5},
		{ID: "LR-302", Title: "inventory table + filters (zone, status, velocity)", Phase: "P3", Epic: "Inventory", Priority: "P1-high", Type: "feat", Estimate: 3},
		{ID: "LR-303", Title: "⌘K search via PG trigram", Phase: "P3", Epic: "Inventory", Priority: "P1-high", Type: "feat", Estimate: 3},
		{ID: "LR-304", Title: "velocity calc (rolling 7d sales)", Phase: "P3", Epic: "Inventory", Priority: "P2-med", Type: "feat", Estimate: 3},

		// P4 · Tasks
		{ID: "LR-401", Title: "tasks table + repo + RBAC", Phase: "P4", Epic: "Tasks", Priority: "P0-blocker", Type: "feat", Estimate: 3},
		{ID: "LR-402", Title: "kanban board UI + drag-drop (dnd-kit)", Phase: "P4", Epic: "Tasks", Priority: "P0-blocker", Type: "feat", Estimate: 5},
		{ID: "LR-403", Title: "deadline + assignee notifications via NATS", Phase: "P4", Epic: "Tasks", Priority: "P1-high", Type: "feat", Estimate: 3},
		{ID: "LR-404", Title: "e2e: task lifecycle", Phase: "P4", Epic: "Tasks", Priority: "P2-med", Type: "chore", Estimate: 2},

		// P5 · Pipelines
		{ID: "LR-501", Title: "spike: pipeline DSL (stage definitions per org)", Phase: "P5", Epic: "Pipelines", Priority: "P0-blocker", Type: "spike", Estimate: 3},
		{ID: "LR-502", Title: "pipelines + stages + cards model", Phase: "P5", Epic: "Pipelines", Priority: "P0-blocker", Type: "feat", Estimate: 5},
		{ID: "LR-503", Title: "pipeline UI (column board, ageing alerts)", Phase: "P5", Epic: "Pipelines", Priority: "P0-blocker", Type: "feat", Estimate: 5},
		{ID: "LR-504", Title: "bottleneck detection + alerts", Phase: "P5", Epic: "Pipelines", Priority: "P1-high", Type: "feat", Estimate: 3},
		{ID: "LR-505", Title: "Restoration pipeline template", Phase: "P5", Epic: "Pipelines", Priority: "P1-high", Type: "feat", Estimate: 2},

		// P6 · POS & Sales Integrations
		{ID: "LR-601", Title: "integrations hub adapter interface", Phase: "P6", Epic: "POS", Priority: "P0-blocker", Type: "feat", Estimate: 3},
		{ID: "LR-602", Title: "Shopify OAuth + webhook handler + idempotency", Phase: "P6", Epic: "POS", Priority: "P0-blocker", Type: "feat", Estimate: 5},
		{ID: "LR-603", Title: "Square OAuth + Catalog/Inventory webhooks", Phase: "P6", Epic: "POS", Priority: "P0-blocker", Type: "feat", Estimate: 5},
		{ID: "LR-604", Title: "sales_events Timescale hypertable + ingest", Phase: "P6", Epic: "POS", Priority: "P0-blocker", Type: "feat", Estimate: 3},
		{ID: "LR-605", Title: "dashboard sales widgets live", Phase: "P6", Epic: "POS", Priority: "P1-high", Type: "feat", Estimate: 3},
		{ID: "LR-606", Title: "webhook event log UI", Phase: "P6", Epic: "POS", Priority: "P2-med", Type: "feat", Estimate: 2},

		// P7 · Analytics
		{ID: "LR-701", Title: "ClickHouse cluster + schema (zone_perf, heatmap, tts, combos)", Phase: "P7", Epic: "Analytics", Priority: "P0-blocker", Type: "feat", Estimate: 5},
		{ID: "LR-702", Title: "NATS → ClickHouse ingest worker", Phase: "P7", Epic: "Analytics", Priority: "P0-blocker", Type: "feat", Estimate: 5},
		{ID: "LR-703", Title: "rollup jobs (5m, hourly, daily)", Phase: "P7", Epic: "Analytics", Priority: "P1-high", Type: "feat", Estimate: 5},
		{ID: "LR-704", Title: "heatmap component (visx) + 7×24 view", Phase: "P7", Epic: "Analytics", Priority: "P1-high", Type: "feat", Estimate: 5},
		{ID: "LR-705", Title: "zone perf bars + sparklines", Phase: "P7", Epic: "Analytics", Priority: "P1-high", Type: "feat", Estimate: 3},
		{ID: "LR-706", Title: "time-to-sell + sell-through queries", Phase: "P7", Epic: "Analytics", Priority: "P1-high", Type: "feat", Estimate: 3},
		{ID: "LR-707", Title: "co-purchase lift (combos)", Phase: "P7", Epic: "Analytics", Priority: "P2-med", Type: "feat", Estimate: 5},
		{ID: "LR-708", Title: "demographics ingest", Phase: "P7", Epic: "Analytics", Priority: "P2-med", Type: "feat", Estimate: 3},

		// P8 · External Signals
		{ID: "LR-801", Title: "OpenWeather adapter (cron, per-store geo)", Phase: "P8", Epic: "Signals", Priority: "P1-high", Type: "feat", Estimate: 3},
		{ID: "LR-802", Title: "Transit adapter (cron)", Phase: "P8", Epic: "Signals", Priority: "P1-high", Type: "feat", Estimate: 3},
		{ID: "LR-803", Title: "insight engine (rule-based recommender)", Phase: "P8", Epic: "Signals", Priority: "P0-blocker", Type: "feat", Estimate: 8},
		{ID: "LR-804", Title: "Apply action wires to task creation", Phase: "P8", Epic: "Signals", Priority: "P1-high", Type: "feat", Estimate: 3},
		{ID: "LR-805", Title: "signal cards on Analytics screen", Phase: "P8", Epic: "Signals", Priority: "P1-high", Type: "feat", Estimate: 2},

		// P9 · Users, RBAC, Audit
		{ID: "LR-901", Title: "role matrix (Admin/Manager/Staff/Read-only/Service)", Phase: "P9", Epic: "RBAC", Priority: "P0-blocker", Type: "feat", Estimate: 5},
		{ID: "LR-902", Title: "zone-scoped policies (RLS predicates)", Phase: "P9", Epic: "RBAC", Priority: "P0-blocker", Type: "feat", Estimate: 5},
		{ID: "LR-903", Title: "2FA via Clerk (TOTP + WebAuthn)", Phase: "P9", Epic: "RBAC", Priority: "P1-high", Type: "feat", Estimate: 3},
		{ID: "LR-904", Title: "audit_log table + monthly partition", Phase: "P9", Epic: "RBAC", Priority: "P1-high", Type: "feat", Estimate: 3},
		{ID: "LR-905", Title: "Users & Access screen pixel-match", Phase: "P9", Epic: "RBAC", Priority: "P1-high", Type: "feat", Estimate: 5},
		{ID: "LR-906", Title: "service token first-class users", Phase: "P9", Epic: "RBAC", Priority: "P2-med", Type: "feat", Estimate: 3},

		// P10 · Integrations Marketplace
		{ID: "LR-A01", Title: "Stripe Connect adapter", Phase: "P10", Epic: "Marketplace", Priority: "P1-high", Type: "feat", Estimate: 5},
		{ID: "LR-A02", Title: "Shippo adapter", Phase: "P10", Epic: "Marketplace", Priority: "P2-med", Type: "feat", Estimate: 3},
		{ID: "LR-A03", Title: "Klaviyo adapter", Phase: "P10", Epic: "Marketplace", Priority: "P2-med", Type: "feat", Estimate: 3},
		{ID: "LR-A04", Title: "NetSuite adapter (skeleton)", Phase: "P10", Epic: "Marketplace", Priority: "P3-low", Type: "feat", Estimate: 3},
		{ID: "LR-A05", Title: "outbound visibility toggles + webhook config UI", Phase: "P10", Epic: "Marketplace", Priority: "P2-med", Type: "feat", Estimate: 3},

		// P11 · Hardening & Launch
		{ID: "LR-B01", Title: "k6 load suite (10k scans/min, 1k WS)", Phase: "P11", Epic: "Hardening", Priority: "P0-blocker", Type: "chore", Estimate: 5},
		{ID: "LR-B02", Title: "Grafana dashboards as code", Phase: "P11", Epic: "Hardening", Priority: "P1-high", Type: "chore", Estimate: 3},
		{ID: "LR-B03", Title: "runbooks + on-call rotation", Phase: "P11", Epic: "Hardening", Priority: "P1-high", Type: "docs", Estimate: 3},
		{ID: "LR-B04", Title: "Stripe billing + plans", Phase: "P11", Epic: "Hardening", Priority: "P0-blocker", Type: "feat", Estimate: 8},
		{ID: "LR-B05", Title: "docs site (Mintlify or Astro)", Phase: "P11", Epic: "Hardening", Priority: "P2-med", Type: "docs", Estimate: 5},
		{ID: "LR-B06", Title: "pen test + ZAP baseline + remediation", Phase: "P11", Epic: "Hardening", Priority: "P0-blocker", Type: "chore", Estimate: 5},
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		fmt.Fprintf(os.Stderr, "missing required env var: %s\n", key)
		os.Exit(1)
	}
	return v
}
