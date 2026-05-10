package auth

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/live-rack/pkg/domain"
)

// DBResolver implements OrgResolver against a store.Querier.
// Imported at the services/api layer to avoid a circular pkg dependency.
// Defined here as an interface so pkg/auth stays DB-agnostic.
type Querier interface {
	GetOrgByClerkID(ctx context.Context, clerkOrgID string) (OrgRow, error)
	GetUserByClerkID(ctx context.Context, clerkUserID string) (UserRow, error)
	GetUserRole(ctx context.Context, userID, orgID uuid.UUID) (string, error)
	GetUserStoreIDs(ctx context.Context, userID, orgID uuid.UUID) ([]uuid.UUID, error)
}

// OrgRow / UserRow are minimal projections — the concrete types live in pkg/store.
type OrgRow struct {
	ID         uuid.UUID
	ClerkOrgID string
	Name       string
	Plan       string
}

type UserRow struct {
	ID          uuid.UUID
	OrgID       uuid.UUID
	ClerkUserID string
	Email       string
	DisplayName string
	AvatarURL   string
}

type DBResolver struct {
	q Querier
}

func NewDBResolver(q Querier) *DBResolver { return &DBResolver{q: q} }

func (r *DBResolver) ResolveOrg(ctx context.Context, clerkOrgID string) (domain.Org, error) {
	row, err := r.q.GetOrgByClerkID(ctx, clerkOrgID)
	if err != nil {
		return domain.Org{}, fmt.Errorf("resolver: get org %s: %w", clerkOrgID, err)
	}
	return domain.Org{
		ID:         row.ID,
		ClerkOrgID: row.ClerkOrgID,
		Name:       row.Name,
		Plan:       domain.Plan(row.Plan),
	}, nil
}

func (r *DBResolver) ResolveUser(ctx context.Context, clerkUserID string, orgID uuid.UUID) (domain.User, error) {
	row, err := r.q.GetUserByClerkID(ctx, clerkUserID)
	if err != nil {
		return domain.User{}, fmt.Errorf("resolver: get user %s: %w", clerkUserID, err)
	}
	if row.OrgID != orgID {
		return domain.User{}, fmt.Errorf("resolver: user org mismatch")
	}
	return domain.User{
		ID:          row.ID,
		OrgID:       row.OrgID,
		ClerkUserID: row.ClerkUserID,
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
