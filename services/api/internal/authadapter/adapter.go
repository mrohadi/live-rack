package authadapter

import (
	"context"

	"github.com/google/uuid"

	pkgauth "github.com/live-rack/pkg/auth"
	"github.com/live-rack/pkg/store"
)

// Adapter bridges *store.Queries to the pkgauth.Querier interface.
type Adapter struct {
	q *store.Queries
}

func New(q *store.Queries) *Adapter { return &Adapter{q: q} }

func (a *Adapter) GetOrgByClerkID(ctx context.Context, clerkOrgID string) (pkgauth.OrgRow, error) {
	org, err := a.q.GetOrgByClerkID(ctx, clerkOrgID)
	if err != nil {
		return pkgauth.OrgRow{}, err
	}
	return pkgauth.OrgRow{
		ID:         org.ID,
		ClerkOrgID: org.ClerkOrgID,
		Name:       org.Name,
		Plan:       org.Plan,
	}, nil
}

func (a *Adapter) GetUserByClerkID(ctx context.Context, clerkUserID string) (pkgauth.UserRow, error) {
	user, err := a.q.GetUserByClerkID(ctx, clerkUserID)
	if err != nil {
		return pkgauth.UserRow{}, err
	}
	return pkgauth.UserRow{
		ID:          user.ID,
		OrgID:       user.OrgID,
		ClerkUserID: user.ClerkUserID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarUrl.String,
	}, nil
}

func (a *Adapter) GetUserRole(ctx context.Context, userID, orgID uuid.UUID) (string, error) {
	return a.q.GetUserRole(ctx, store.GetUserRoleParams{UserID: userID, OrgID: orgID})
}

func (a *Adapter) GetUserStoreIDs(ctx context.Context, userID, orgID uuid.UUID) ([]uuid.UUID, error) {
	return a.q.GetUserStoreIDs(ctx, store.GetUserStoreIDsParams{UserID: userID, OrgID: orgID})
}
