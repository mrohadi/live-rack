package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// UserListRow is one row of the org user roster with role, presence, and scope.
type UserListRow struct {
	ID          uuid.UUID  `json:"id"`
	IDPUserID   string     `json:"idp_user_id"`
	Email       string     `json:"email"`
	DisplayName string     `json:"display_name"`
	AvatarURL   string     `json:"avatar_url"`
	Role        string     `json:"role"`
	Title       string     `json:"title"`
	Shift       string     `json:"shift"`
	Status      string     `json:"status"`
	MFAEnabled  bool       `json:"mfa_enabled"`
	LastSeenAt  *time.Time `json:"last_seen_at"`
	Zones       []string   `json:"zones"`
}

const listUsersByOrg = `
SELECT u.id, u.idp_user_id, u.email, u.display_name, COALESCE(u.avatar_url, ''), COALESCE(r.name, ''),
       u.title, u.shift, u.status, u.mfa_enabled, u.last_seen_at,
       COALESCE(array_remove(array_agg(z.name ORDER BY z.name), NULL), '{}') AS zones
FROM users u
LEFT JOIN role_bindings rb ON rb.user_id = u.id AND rb.org_id = u.org_id
LEFT JOIN roles r          ON r.id = rb.role_id
LEFT JOIN user_zones uz    ON uz.user_id = u.id
LEFT JOIN zones z          ON z.id = uz.zone_id
WHERE u.org_id = $1
GROUP BY u.id, r.name
ORDER BY u.display_name
`

// ListUsersByOrg returns the org's users with role, presence, and zone scope.
func (q *Queries) ListUsersByOrg(ctx context.Context, orgID uuid.UUID) ([]UserListRow, error) {
	rows, err := q.db.Query(ctx, listUsersByOrg, orgID)
	if err != nil {
		return nil, fmt.Errorf("store: list users by org: %w", err)
	}
	defer rows.Close()

	var out []UserListRow
	for rows.Next() {
		var u UserListRow
		if err := rows.Scan(&u.ID, &u.IDPUserID, &u.Email, &u.DisplayName, &u.AvatarURL, &u.Role,
			&u.Title, &u.Shift, &u.Status, &u.MFAEnabled, &u.LastSeenAt, &u.Zones); err != nil {
			return nil, fmt.Errorf("store: scan user row: %w", err)
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// RosterStats aggregates the Users & Access header metrics. Members counts
// joined users; Pending counts invited users awaiting first login.
type RosterStats struct {
	Members   int `json:"members"`
	ActiveNow int `json:"active_now"`
	MFAUsers  int `json:"mfa_users"`
	Roles     int `json:"roles"`
	Pending   int `json:"pending"`
}

const rosterStats = `
SELECT
  count(*) FILTER (WHERE u.status <> 'pending')::int   AS members,
  count(*) FILTER (WHERE u.status = 'active')::int     AS active_now,
  count(*) FILTER (WHERE u.mfa_enabled)::int           AS mfa_users,
  count(DISTINCT r.name)::int                          AS roles,
  count(*) FILTER (WHERE u.status = 'pending')::int    AS pending
FROM users u
LEFT JOIN role_bindings rb ON rb.user_id = u.id AND rb.org_id = u.org_id
LEFT JOIN roles r          ON r.id = rb.role_id
WHERE u.org_id = $1
`

// RosterStatsByOrg returns aggregate roster metrics for the org.
func (q *Queries) RosterStatsByOrg(ctx context.Context, orgID uuid.UUID) (RosterStats, error) {
	var s RosterStats
	if err := q.db.QueryRow(ctx, rosterStats, orgID).
		Scan(&s.Members, &s.ActiveNow, &s.MFAUsers, &s.Roles, &s.Pending); err != nil {
		return RosterStats{}, fmt.Errorf("store: roster stats: %w", err)
	}
	return s, nil
}

// TouchLastSeen stamps a user's last-seen time. Best-effort presence tracking.
func (q *Queries) TouchLastSeen(ctx context.Context, userID, orgID uuid.UUID) error {
	_, err := q.db.Exec(ctx,
		`UPDATE users SET last_seen_at = now() WHERE id = $1 AND org_id = $2`, userID, orgID)
	if err != nil {
		return fmt.Errorf("store: touch last seen: %w", err)
	}
	return nil
}

// SetUserMFA records whether a user has a verified second factor (synced from
// the caller's OIDC amr claim, which is only present in the ID token).
func (q *Queries) SetUserMFA(ctx context.Context, userID, orgID uuid.UUID, enabled bool) error {
	_, err := q.db.Exec(ctx,
		`UPDATE users SET mfa_enabled = $3 WHERE id = $1 AND org_id = $2`, userID, orgID, enabled)
	if err != nil {
		return fmt.Errorf("store: set user mfa: %w", err)
	}
	return nil
}
