package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwks"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/google/uuid"
	"github.com/live-rack/pkg/domain"
)

// OrgResolver looks up internal org + user records from Clerk IDs.
type OrgResolver interface {
	ResolveOrg(ctx context.Context, clerkOrgID string) (domain.Org, error)
	ResolveUser(ctx context.Context, clerkUserID string, orgID uuid.UUID) (domain.User, error)
	UserRole(ctx context.Context, userID, orgID uuid.UUID) (domain.RoleName, error)
	UserStoreIDs(ctx context.Context, userID, orgID uuid.UUID) ([]uuid.UUID, error)
}

type ClerkVerifier struct {
	jwksClient *jwks.Client
	resolver   OrgResolver
}

func NewClerkVerifier(secretKey string, resolver OrgResolver) *ClerkVerifier {
	clerk.SetKey(secretKey)
	return &ClerkVerifier{
		jwksClient: jwks.NewClient(&clerk.ClientConfig{}),
		resolver:   resolver,
	}
}

// VerifyRequest extracts + validates the Bearer JWT, returns Principal.
func (v *ClerkVerifier) VerifyRequest(r *http.Request) (*domain.Principal, error) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, fmt.Errorf("auth: missing bearer token")
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")

	claims, err := jwt.Verify(r.Context(), &jwt.VerifyParams{
		Token:      token,
		JWKSClient: v.jwksClient,
	})
	if err != nil {
		return nil, fmt.Errorf("auth: invalid token: %w", err)
	}

	clerkOrgID := claims.ActiveOrganizationID
	clerkUserID := claims.Subject
	if clerkOrgID == "" {
		return nil, fmt.Errorf("auth: token missing org_id claim")
	}

	ctx := r.Context()

	org, err := v.resolver.ResolveOrg(ctx, clerkOrgID)
	if err != nil {
		return nil, fmt.Errorf("auth: resolve org: %w", err)
	}

	user, err := v.resolver.ResolveUser(ctx, clerkUserID, org.ID)
	if err != nil {
		return nil, fmt.Errorf("auth: resolve user: %w", err)
	}

	role, err := v.resolver.UserRole(ctx, user.ID, org.ID)
	if err != nil {
		return nil, fmt.Errorf("auth: resolve role: %w", err)
	}

	storeIDs, err := v.resolver.UserStoreIDs(ctx, user.ID, org.ID)
	if err != nil {
		return nil, fmt.Errorf("auth: resolve store scope: %w", err)
	}

	return &domain.Principal{
		UserID:     user.ID,
		OrgID:      org.ID,
		ClerkOrgID: clerkOrgID,
		Role:       role,
		StoreIDs:   storeIDs,
	}, nil
}
