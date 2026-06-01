package signup_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/live-rack/services/api/internal/signup"
)

type fakeProvisioner struct {
	orgName, createdEmail, grantedRole string
	createOrgErr                       error
}

func (f *fakeProvisioner) CreateOrg(_ context.Context, name string) (string, error) {
	if f.createOrgErr != nil {
		return "", f.createOrgErr
	}
	f.orgName = name
	return "zid-org-1", nil
}

func (f *fakeProvisioner) CreateHumanUser(_ context.Context, _, email, _ string) (string, error) {
	f.createdEmail = email
	return "zid-user-1", nil
}

func (f *fakeProvisioner) GrantProjectRole(_ context.Context, _, _, role string) error {
	f.grantedRole = role
	return nil
}

func serve(t *testing.T, body string) (*httptest.ResponseRecorder, *fakeProvisioner) {
	t.Helper()
	fp := &fakeProvisioner{}
	e := echo.New()
	signup.New(fp).Register(e)
	req := httptest.NewRequestWithContext(
		context.Background(), http.MethodPost, "/api/v1/signup", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec, fp
}

func TestSignup_ProvisionsOrgAdmin(t *testing.T) {
	rec, fp := serve(t, `{"company":"Acme Co","email":"Founder@acme.test","display_name":"Ada Founder"}`)
	require.Equal(t, http.StatusCreated, rec.Code)

	var out signup.Response
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	assert.Equal(t, "zid-org-1", out.OrgID)
	assert.Equal(t, "zid-user-1", out.UserID)
	assert.Equal(t, "pending_verification", out.Status)

	assert.Equal(t, "Acme Co", fp.orgName)
	assert.Equal(t, "founder@acme.test", fp.createdEmail) // normalised
	assert.Equal(t, "admin", fp.grantedRole)
}

func TestSignup_RejectsMissingCompany(t *testing.T) {
	rec, _ := serve(t, `{"email":"a@acme.test"}`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSignup_RejectsBadEmail(t *testing.T) {
	rec, _ := serve(t, `{"company":"Acme","email":"nope"}`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSignup_PropagatesProviderError(t *testing.T) {
	fp := &fakeProvisioner{createOrgErr: errors.New("zitadel down")}
	e := echo.New()
	signup.New(fp).Register(e)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/signup",
		strings.NewReader(`{"company":"Acme","email":"a@acme.test"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadGateway, rec.Code)
}
