package auth

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/live-rack/pkg/domain"
)

// Querier is the DB-agnostic surface pkg/auth needs.
// The concrete implementation lives at the services/api layer (authadapter)
// to keep pkg/auth free of a pkg/store dependency.
type Querier interface {
	GetOrgByIdpID(ctx context.Context, idpOrgID string) (OrgRow, error)
	GetUserByIdpID(ctx context.Context, idpUserID string) (UserRow, error)
	GetUserRole(ctx context.Context, userID, orgID uuid.UUID) (string, error)
	GetUserStoreIDs(ctx context.Context, userID, orgID uuid.UUID) ([]uuid.UUID, error)

	// JIT provisioning (replaces the Clerk Svix webhook).
	UpsertOrg(ctx context.Context, idpOrgID, name string) (OrgRow, error)
	UpsertUser(ctx context.Context, orgID uuid.UUID, idpUserID, email, displayName, avatarURL string) (UserRow, error)
	BindUserRole(ctx context.Context, orgID, userID uuid.UUID, role string) error
}

// OrgRow / UserRow are minimal projections — the concrete types live in pkg/store.
type OrgRow struct {
	ID       uuid.UUID
	IDPOrgID string
	Name     string
	Plan     string
}

type UserRow struct {
	ID          uuid.UUID
	OrgID       uuid.UUID
	IDPUserID   string
	Email       string
	DisplayName string
	AvatarURL   string
}

type DBResolver struct {
	q Querier
}

func NewDBResolver(q Querier) *DBResolver { return &DBResolver{q: q} }

func (r *DBResolver) ResolveOrg(ctx context.Context, idpOrgID string) (domain.Org, error) {
	row, err := r.q.GetOrgByIdpID(ctx, idpOrgID)
	if err != nil {
		return domain.Org{}, fmt.Errorf("resolver: get org %s: %w", idpOrgID, err)
	}
	return domain.Org{
		ID:       row.ID,
		IDPOrgID: row.IDPOrgID,
		Name:     row.Name,
		Plan:     domain.Plan(row.Plan),
	}, nil
}

func (r *DBResolver) ResolveUser(ctx context.Context, idpUserID string, orgID uuid.UUID) (domain.User, error) {
	row, err := r.q.GetUserByIdpID(ctx, idpUserID)
	if err != nil {
		return domain.User{}, fmt.Errorf("resolver: get user %s: %w", idpUserID, err)
	}
	if row.OrgID != orgID {
		return domain.User{}, fmt.Errorf("resolver: user org mismatch")
	}
	return domain.User{
		ID:          row.ID,
		OrgID:       row.OrgID,
		IDPUserID:   row.IDPUserID,
		Email:       row.Email,
		DisplayName: row.DisplayName,
		AvatarURL:   row.AvatarURL,
	}, nil
}

func (r *DBResolver) UserRole(ctx context.Context, userID, orgID uuid.UUID) (domain.RoleName, error) {
	role, err := r.q.GetUserRole(ctx, userID, orgID)
	if err != nil {
		return "", fmt.Errorf("resolver: get role: %w", err)
	}
	return domain.RoleName(role), nil
}

func (r *DBResolver) UserStoreIDs(ctx context.Context, userID, orgID uuid.UUID) ([]uuid.UUID, error) {
	ids, err := r.q.GetUserStoreIDs(ctx, userID, orgID)
	if err != nil {
		return nil, fmt.Errorf("resolver: get store ids: %w", err)
	}
	return ids, nil
}

// Provision upserts the org + user + role binding from token claims.
// Called on every verified request so first-login users are created lazily.
func (r *DBResolver) Provision(ctx context.Context, c Claims) error {
	org, err := r.q.UpsertOrg(ctx, c.IDPOrgID, c.OrgName)
	if err != nil {
		return fmt.Errorf("resolver: upsert org: %w", err)
	}
	user, err := r.q.UpsertUser(ctx, org.ID, c.Subject, c.Email, c.DisplayName, c.AvatarURL)
	if err != nil {
		return fmt.Errorf("resolver: upsert user: %w", err)
	}
	if err := r.q.BindUserRole(ctx, org.ID, user.ID, string(c.Role)); err != nil {
		return fmt.Errorf("resolver: bind role: %w", err)
	}
	return nil
}
