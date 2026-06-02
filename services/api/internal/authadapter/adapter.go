package authadapter

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	pkgauth "github.com/live-rack/pkg/auth"
	"github.com/live-rack/pkg/domain"
	"github.com/live-rack/pkg/store"
)

// Adapter bridges *store.Queries to the pkgauth.Querier interface.
type Adapter struct {
	q *store.Queries
}

func New(q *store.Queries) *Adapter { return &Adapter{q: q} }

// ResolveServiceToken implements pkgauth.ServiceTokenLookup: it maps a token
// hash to a service principal (the token id doubles as the principal user id).
func (a *Adapter) ResolveServiceToken(ctx context.Context, hash string) (*domain.Principal, error) {
	row, err := a.q.ResolveServiceToken(ctx, hash)
	if err != nil {
		return nil, err
	}
	return &domain.Principal{UserID: row.ID, OrgID: row.OrgID, Role: domain.RoleService}, nil
}

func (a *Adapter) GetOrgByIdpID(ctx context.Context, idpOrgID string) (pkgauth.OrgRow, error) {
	org, err := a.q.GetOrgByIdpID(ctx, idpOrgID)
	if err != nil {
		return pkgauth.OrgRow{}, err
	}
	return orgRow(org), nil
}

func (a *Adapter) GetUserByIdpID(ctx context.Context, idpUserID string) (pkgauth.UserRow, error) {
	user, err := a.q.GetUserByIdpID(ctx, idpUserID)
	if err != nil {
		return pkgauth.UserRow{}, err
	}
	return userRow(user), nil
}

func (a *Adapter) GetUserRole(ctx context.Context, userID, orgID uuid.UUID) (string, error) {
	return a.q.GetUserRole(ctx, store.GetUserRoleParams{UserID: userID, OrgID: orgID})
}

func (a *Adapter) GetUserStoreIDs(ctx context.Context, userID, orgID uuid.UUID) ([]uuid.UUID, error) {
	return a.q.GetUserStoreIDs(ctx, store.GetUserStoreIDsParams{UserID: userID, OrgID: orgID})
}

func (a *Adapter) UpsertOrg(ctx context.Context, idpOrgID, name string) (pkgauth.OrgRow, error) {
	org, err := a.q.UpsertOrg(ctx, store.UpsertOrgParams{
		IdpOrgID: idpOrgID,
		Name:     name,
		Plan:     "free",
	})
	if err != nil {
		return pkgauth.OrgRow{}, err
	}
	return orgRow(org), nil
}

func (a *Adapter) UpsertUser(ctx context.Context, orgID uuid.UUID, idpUserID, email, displayName, avatarURL string) (pkgauth.UserRow, error) {
	user, err := a.q.UpsertUser(ctx, store.UpsertUserParams{
		OrgID:       orgID,
		IdpUserID:   idpUserID,
		Email:       email,
		DisplayName: displayName,
		AvatarUrl:   pgtype.Text{String: avatarURL, Valid: avatarURL != ""},
	})
	if err != nil {
		return pkgauth.UserRow{}, err
	}
	return userRow(user), nil
}

func (a *Adapter) BindUserRole(ctx context.Context, orgID, userID uuid.UUID, role string) error {
	return a.q.BindUserRole(ctx, store.BindUserRoleParams{OrgID: orgID, UserID: userID, Name: role})
}

func orgRow(o store.Org) pkgauth.OrgRow {
	return pkgauth.OrgRow{ID: o.ID, IDPOrgID: o.IdpOrgID, Name: o.Name, Plan: o.Plan}
}

func userRow(u store.User) pkgauth.UserRow {
	return pkgauth.UserRow{
		ID:          u.ID,
		OrgID:       u.OrgID,
		IDPUserID:   u.IdpUserID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		AvatarURL:   u.AvatarUrl.String,
	}
}
