package users_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/live-rack/pkg/audit"
	pkgauth "github.com/live-rack/pkg/auth"
	"github.com/live-rack/pkg/domain"
	"github.com/live-rack/pkg/store"
	"github.com/live-rack/services/api/internal/users"
)

type fakeInviter struct {
	createdEmail, grantedRole, resentID string
	createErr                           error
}

func (f *fakeInviter) CreateHumanUser(_ context.Context, _, email, _ string) (string, error) {
	if f.createErr != nil {
		return "", f.createErr
	}
	f.createdEmail = email
	return "zid-user-1", nil
}

func (f *fakeInviter) GrantProjectRole(_ context.Context, _, _, role string) error {
	f.grantedRole = role
	return nil
}

func (f *fakeInviter) ResendInvite(_ context.Context, _, userID string) error {
	f.resentID = userID
	return nil
}

type fakeInviteStore struct {
	createdEmail, boundRole string
}

func (f *fakeInviteStore) CreateInvitedUser(_ context.Context, arg store.CreateInvitedUserParams) (store.User, error) {
	f.createdEmail = arg.Email
	return store.User{ID: uuid.New(), OrgID: arg.OrgID, IdpUserID: arg.IdpUserID, Email: arg.Email}, nil
}

func (f *fakeInviteStore) BindUserRole(_ context.Context, arg store.BindUserRoleParams) error {
	f.boundRole = arg.Name
	return nil
}

type fakeAuditor struct{ actions []string }

func (f *fakeAuditor) Write(_ context.Context, e audit.Entry) error {
	f.actions = append(f.actions, e.Action)
	return nil
}

func serveInvite(t *testing.T, p *domain.Principal, method, target, body string) (*httptest.ResponseRecorder, *fakeInviter, *fakeAuditor) {
	t.Helper()
	zit := &fakeInviter{}
	aud := &fakeAuditor{}
	st := &fakeInviteStore{}
	e := echo.New()
	users.NewInvite(zit, st, aud).Register(e.Group("/api/v1"))
	req := httptest.NewRequestWithContext(
		pkgauth.WithPrincipal(context.Background(), p), method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec, zit, aud
}

func adminMFA() *domain.Principal {
	return &domain.Principal{
		UserID:      uuid.New(),
		OrgID:       uuid.New(),
		IDPOrgID:    "zid-org-1",
		Role:        domain.RoleAdmin,
		MFAVerified: true,
	}
}

func TestInvite_AdminCreatesAndGrants(t *testing.T) {
	rec, zit, aud := serveInvite(t, adminMFA(), http.MethodPost, "/api/v1/users/invite",
		`{"email":"NEW@acme.test","display_name":"New Person","role":"manager"}`)
	require.Equal(t, http.StatusCreated, rec.Code)

	var out users.InviteResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	assert.Equal(t, "zid-user-1", out.UserID)
	assert.Equal(t, "new@acme.test", out.Email) // normalised
	assert.Equal(t, "manager", out.Role)
	assert.Equal(t, "invited", out.Status)

	assert.Equal(t, "new@acme.test", zit.createdEmail)
	assert.Equal(t, "manager", zit.grantedRole)
	assert.Equal(t, []string{"user.invited"}, aud.actions)
}

func TestInvite_RejectsInvalidRole(t *testing.T) {
	rec, _, _ := serveInvite(t, adminMFA(), http.MethodPost, "/api/v1/users/invite",
		`{"email":"x@acme.test","role":"service"}`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestInvite_RejectsBadEmail(t *testing.T) {
	rec, _, _ := serveInvite(t, adminMFA(), http.MethodPost, "/api/v1/users/invite",
		`{"email":"not-an-email","role":"staff"}`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestInvite_NonAdminForbidden(t *testing.T) {
	p := &domain.Principal{UserID: uuid.New(), OrgID: uuid.New(), Role: domain.RoleStaff}
	rec, _, _ := serveInvite(t, p, http.MethodPost, "/api/v1/users/invite",
		`{"email":"x@acme.test","role":"staff"}`)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestInvite_AdminWithoutTokenMFAStillAllowed(t *testing.T) {
	// Access tokens carry no amr; MFA is enforced at the IdP login policy, so the
	// gateway authorizes on the admin role alone.
	p := &domain.Principal{UserID: uuid.New(), OrgID: uuid.New(), IDPOrgID: "zid-org-1", Role: domain.RoleAdmin, MFAVerified: false}
	rec, _, _ := serveInvite(t, p, http.MethodPost, "/api/v1/users/invite",
		`{"email":"x@acme.test","role":"staff"}`)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestResend_AdminResends(t *testing.T) {
	rec, zit, aud := serveInvite(t, adminMFA(), http.MethodPost, "/api/v1/users/zid-user-9/resend", "")
	require.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "zid-user-9", zit.resentID)
	assert.Equal(t, []string{"user.invite_resent"}, aud.actions)
}
