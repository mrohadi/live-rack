package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// UserListRow is one row of the org user roster with its bound role.
type UserListRow struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
	Role        string    `json:"role"`
}

const listUsersByOrg = `
SELECT u.id, u.email, u.display_name, COALESCE(u.avatar_url, ''), COALESCE(r.name, '')
FROM users u
LEFT JOIN role_bindings rb ON rb.user_id = u.id AND rb.org_id = u.org_id
LEFT JOIN roles r ON r.id = rb.role_id
WHERE u.org_id = $1
ORDER BY u.display_name
`

// ListUsersByOrg returns the org's users with their role. Hand-written (sqlc
// query lives outside the generated set) but follows the same shape.
func (q *Queries) ListUsersByOrg(ctx context.Context, orgID uuid.UUID) ([]UserListRow, error) {
	rows, err := q.db.Query(ctx, listUsersByOrg, orgID)
	if err != nil {
		return nil, fmt.Errorf("store: list users by org: %w", err)
	}
	defer rows.Close()

	var out []UserListRow
	for rows.Next() {
		var u UserListRow
		if err := rows.Scan(&u.ID, &u.Email, &u.DisplayName, &u.AvatarURL, &u.Role); err != nil {
			return nil, fmt.Errorf("store: scan user row: %w", err)
		}
		out = append(out, u)
	}
	return out, rows.Err()
}
