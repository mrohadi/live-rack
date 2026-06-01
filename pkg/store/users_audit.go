package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AuditRow is one audit-trail entry for display.
type AuditRow struct {
	TS           time.Time      `json:"ts"`
	ActorUserID  *uuid.UUID     `json:"actor_user_id"`
	Action       string         `json:"action"`
	ResourceType string         `json:"resource_type"`
	ResourceID   string         `json:"resource_id"`
	Metadata     map[string]any `json:"metadata"`
}

const listAudit = `
SELECT ts, actor_user_id, action, resource_type, resource_id, metadata
FROM audit_log
WHERE org_id = $1
  AND ($2::uuid IS NULL OR actor_user_id = $2)
ORDER BY ts DESC
LIMIT $3
`

// ListAudit returns recent audit entries for an org, optionally filtered to one
// actor. A nil actor returns the whole org trail.
func (q *Queries) ListAudit(ctx context.Context, orgID uuid.UUID, actor *uuid.UUID, limit int) ([]AuditRow, error) {
	rows, err := q.db.Query(ctx, listAudit, orgID, actor, limit)
	if err != nil {
		return nil, fmt.Errorf("store: list audit: %w", err)
	}
	defer rows.Close()

	var out []AuditRow
	for rows.Next() {
		var a AuditRow
		var meta []byte
		if err := rows.Scan(&a.TS, &a.ActorUserID, &a.Action, &a.ResourceType, &a.ResourceID, &meta); err != nil {
			return nil, fmt.Errorf("store: scan audit row: %w", err)
		}
		if len(meta) > 0 {
			_ = json.Unmarshal(meta, &a.Metadata)
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// SetUserRole repoints a user's single role binding to the named built-in role.
func (q *Queries) SetUserRole(ctx context.Context, orgID, userID uuid.UUID, role string) error {
	if _, err := q.db.Exec(ctx,
		`DELETE FROM role_bindings WHERE org_id = $1 AND user_id = $2`, orgID, userID); err != nil {
		return fmt.Errorf("store: clear role bindings: %w", err)
	}
	if _, err := q.db.Exec(ctx,
		`INSERT INTO role_bindings (org_id, user_id, role_id)
		 SELECT $1, $2, id FROM roles WHERE org_id = $1 AND name = $3`,
		orgID, userID, role); err != nil {
		return fmt.Errorf("store: bind role: %w", err)
	}
	return nil
}
