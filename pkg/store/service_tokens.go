package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// ResolveServiceTokenRow is a resolved, non-revoked service token.
type ResolveServiceTokenRow struct {
	ID    uuid.UUID
	OrgID uuid.UUID
}

const resolveServiceToken = `
UPDATE service_tokens
SET last_used_at = NOW()
WHERE token_hash = $1 AND revoked_at IS NULL
RETURNING id, org_id
`

// ResolveServiceToken looks up a live token by hash and stamps last_used_at.
// Runs on the pool connection (RLS-exempt owner) since it executes before the
// tenant context is established. Hand-written.
func (q *Queries) ResolveServiceToken(ctx context.Context, hash string) (ResolveServiceTokenRow, error) {
	var r ResolveServiceTokenRow
	err := q.db.QueryRow(ctx, resolveServiceToken, hash).Scan(&r.ID, &r.OrgID)
	if err != nil {
		return ResolveServiceTokenRow{}, fmt.Errorf("store: resolve service token: %w", err)
	}
	return r, nil
}

//nolint:gosec // G101: SQL statement text, not a hardcoded credential.
const createServiceToken = `
INSERT INTO service_tokens (org_id, name, token_hash)
VALUES ($1, $2, $3)
RETURNING id
`

// CreateServiceToken stores a new token's hash and returns its id. Hand-written.
func (q *Queries) CreateServiceToken(ctx context.Context, orgID uuid.UUID, name, hash string) (uuid.UUID, error) {
	var id uuid.UUID
	if err := q.db.QueryRow(ctx, createServiceToken, orgID, name, hash).Scan(&id); err != nil {
		return uuid.Nil, fmt.Errorf("store: create service token: %w", err)
	}
	return id, nil
}
