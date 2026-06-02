package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

const updateOrgPlan = `UPDATE orgs SET plan = $2, updated_at = NOW() WHERE id = $1`

// UpdateOrgPlan sets an org's billing plan. Hand-written. RLS-exempt: billing
// webhooks run before tenant context, on the pool connection.
func (q *Queries) UpdateOrgPlan(ctx context.Context, orgID uuid.UUID, plan string) error {
	if _, err := q.db.Exec(ctx, updateOrgPlan, orgID, plan); err != nil {
		return fmt.Errorf("store: update org plan: %w", err)
	}
	return nil
}
