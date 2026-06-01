package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	"golang.org/x/oauth2"

	"github.com/live-rack/pkg/domain"
)

// Verifier authenticates an HTTP request and returns the resolved Principal.
type Verifier interface {
	VerifyRequest(r *http.Request) (*domain.Principal, error)
}

// Claims is the IdP-neutral identity extracted from a verified OIDC token.
type Claims struct {
	Subject     string
	IDPOrgID    string
	OrgName     string
	Email       string
	DisplayName string
	AvatarURL   string
	Role        domain.RoleName
	MFA         bool
}

// mfaMethods are amr (Authentication Methods References) values that indicate a
// second factor was used. "mfa" is the aggregate marker; the rest are concrete
// second factors Zitadel may report.
var mfaMethods = map[string]bool{
	"mfa": true, "otp": true, "totp": true, "hwk": true, "webauthn": true, "u2f": true,
}

// amrIndicatesMFA reports whether an amr claim array contains a second factor. Pure.
func amrIndicatesMFA(amr []any) bool {
	for _, v := range amr {
		if s, ok := v.(string); ok && mfaMethods[strings.ToLower(s)] {
			return true
		}
	}
	return false
}

// OrgResolver looks up internal org + user records and provisions them on first login.
type OrgResolver interface {
	ResolveOrg(ctx context.Context, idpOrgID string) (domain.Org, error)
	ResolveUser(ctx context.Context, idpUserID string, orgID uuid.UUID) (domain.User, error)
	UserRole(ctx context.Context, userID, orgID uuid.UUID) (domain.RoleName, error)
	UserStoreIDs(ctx context.Context, userID, orgID uuid.UUID) ([]uuid.UUID, error)
	Provision(ctx context.Context, c Claims) error
}

// Zitadel claim URNs.
const (
	claimOrgID    = "urn:zitadel:iam:user:resourceowner:id"
	claimOrgName  = "urn:zitadel:iam:user:resourceowner:name"
	claimRolesFmt = "urn:zitadel:iam:org:project:%s:roles"
)

// rolePrecedence picks the strongest role when a user holds several.
var rolePrecedence = []domain.RoleName{
	domain.RoleAdmin,
	domain.RoleManager,
	domain.RoleStaff,
	domain.RoleService,
	domain.RoleReadonly,
}

// ZitadelVerifier validates OIDC JWTs against Zitadel's JWKS and maps them to a Principal.
type ZitadelVerifier struct {
	verifier  *oidc.IDTokenVerifier
	provider  *oidc.Provider
	resolver  OrgResolver
	projectID string
}

// NewZitadelVerifier discovers the issuer's OIDC config (JWKS, etc.) once at startup.
func NewZitadelVerifier(ctx context.Context, issuer, projectID string, resolver OrgResolver) (*ZitadelVerifier, error) {
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("auth: oidc discovery %s: %w", issuer, err)
	}
	return &ZitadelVerifier{
		verifier:  provider.Verifier(&oidc.Config{SkipClientIDCheck: true}),
		provider:  provider,
		resolver:  resolver,
		projectID: projectID,
	}, nil
}

// VerifyRequest extracts + validates the Bearer JWT, JIT-provisions, returns Principal.
func (v *ZitadelVerifier) VerifyRequest(r *http.Request) (*domain.Principal, error) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, fmt.Errorf("auth: missing bearer token")
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")

	ctx := r.Context()
	idToken, err := v.verifier.Verify(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("auth: invalid token: %w", err)
	}

	var raw map[string]any
	if err := idToken.Claims(&raw); err != nil {
		return nil, fmt.Errorf("auth: decode claims: %w", err)
	}

	claims := v.parseClaims(idToken.Subject, raw)
	if claims.IDPOrgID == "" {
		return nil, fmt.Errorf("auth: token missing org claim")
	}

	// Zitadel JWT access tokens carry roles but not profile/email; fetch those
	// from the userinfo endpoint so the roster shows real names, not blanks.
	v.enrichFromUserInfo(ctx, token, &claims)

	if err := v.resolver.Provision(ctx, claims); err != nil {
		return nil, fmt.Errorf("auth: provision: %w", err)
	}

	org, err := v.resolver.ResolveOrg(ctx, claims.IDPOrgID)
	if err != nil {
		return nil, fmt.Errorf("auth: resolve org: %w", err)
	}
	user, err := v.resolver.ResolveUser(ctx, claims.Subject, org.ID)
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
		UserID:      user.ID,
		OrgID:       org.ID,
		IDPOrgID:    claims.IDPOrgID,
		Role:        role,
		StoreIDs:    storeIDs,
		MFAVerified: claims.MFA,
	}, nil
}

func (v *ZitadelVerifier) parseClaims(subject string, raw map[string]any) Claims {
	roles := v.rolesMap(raw)
	orgID, orgName := orgFromRoles(roles)
	// Prefer the dedicated resourceowner claim when present; fall back to roles map.
	if id := stringClaim(raw, claimOrgID); id != "" {
		orgID = id
	}
	if name := stringClaim(raw, claimOrgName); name != "" {
		orgName = name
	}
	return Claims{
		Subject:     subject,
		IDPOrgID:    orgID,
		OrgName:     orgName,
		Email:       stringClaim(raw, "email"),
		DisplayName: stringClaim(raw, "name"),
		AvatarURL:   stringClaim(raw, "picture"),
		Role:        strongestRole(roles),
		MFA:         mfaFromClaims(raw),
	}
}

// enrichFromUserInfo fills missing profile claims from the OIDC userinfo
// endpoint. Best-effort: the token already authenticated, so a userinfo
// failure must not block the request. Only fills empty fields.
func (v *ZitadelVerifier) enrichFromUserInfo(ctx context.Context, token string, c *Claims) {
	if c.Email != "" && c.DisplayName != "" && c.AvatarURL != "" {
		return
	}
	ui, err := v.provider.UserInfo(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}))
	if err != nil {
		return
	}
	var raw map[string]any
	if err := ui.Claims(&raw); err != nil {
		return
	}
	if c.Email == "" {
		c.Email = firstNonEmpty(ui.Email, stringClaim(raw, "email"))
	}
	if c.DisplayName == "" {
		c.DisplayName = firstNonEmpty(stringClaim(raw, "name"), stringClaim(raw, "preferred_username"))
	}
	if c.AvatarURL == "" {
		c.AvatarURL = stringClaim(raw, "picture")
	}
}

// firstNonEmpty returns the first non-empty string. Pure.
func firstNonEmpty(vals ...string) string {
	for _, s := range vals {
		if s != "" {
			return s
		}
	}
	return ""
}

// mfaFromClaims reads the amr claim. Pure.
func mfaFromClaims(raw map[string]any) bool {
	amr, ok := raw["amr"].([]any)
	if !ok {
		return false
	}
	return amrIndicatesMFA(amr)
}

// rolesMap returns the Zitadel project-roles claim: { role: { orgID: orgDomain } }.
func (v *ZitadelVerifier) rolesMap(raw map[string]any) map[string]any {
	if m, ok := raw[fmt.Sprintf(claimRolesFmt, v.projectID)].(map[string]any); ok {
		return m
	}
	return nil
}

// orgFromRoles derives org id + domain from the inner key of any role grant.
func orgFromRoles(roles map[string]any) (id, name string) {
	for _, grant := range roles {
		orgs, ok := grant.(map[string]any)
		if !ok {
			continue
		}
		for orgID, domain := range orgs {
			d, _ := domain.(string)
			return orgID, d
		}
	}
	return "", ""
}

func strongestRole(roles map[string]any) domain.RoleName {
	for _, want := range rolePrecedence {
		if _, ok := roles[string(want)]; ok {
			return want
		}
	}
	return domain.RoleReadonly
}

func stringClaim(raw map[string]any, key string) string {
	if s, ok := raw[key].(string); ok {
		return s
	}
	return ""
}
