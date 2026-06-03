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
	createUserErr                      error
}

func (f *fakeProvisioner) CreateOrg(_ context.Context, name string) (string, error) {
	if f.createOrgErr != nil {
		return "", f.createOrgErr
	}
	f.orgName = name
	return "zid-org-1", nil
}

func (f *fakeProvisioner) CreateHumanUser(_ context.Context, _, email, _ string) (string, error) {
	if f.createUserErr != nil {
		return "", f.createUserErr
	}
	f.createdEmail = email
	return "zid-user-1", nil
}

func (f *fakeProvisioner) CreateHumanUserReturnCode(_ context.Context, _, email, _ string) (string, string, error) {
	if f.createUserErr != nil {
		return "", "", f.createUserErr
	}
	f.createdEmail = email
	return "zid-user-1", "TEST-CODE", nil
}

func (f *fakeProvisioner) GrantProjectRole(_ context.Context, _, _, role string) error {
	f.grantedRole = role
	return nil
}

func serve(t *testing.T, body string) (*httptest.ResponseRecorder, *fakeProvisioner) {
	t.Helper()
	fp := &fakeProvisioner{}
	e := echo.New()
	signup.New(fp, false, "http://localhost:5173").Register(e)
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

	assert.True(t, strings.HasPrefix(fp.orgName, "Acme Co-"), "org name must start with company + suffix: %s", fp.orgName)
	assert.Len(t, strings.TrimPrefix(fp.orgName, "Acme Co-"), 6, "suffix must be 6 chars")
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
	signup.New(fp, false, "http://localhost:5173").Register(e)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/signup",
		strings.NewReader(`{"company":"Acme","email":"a@acme.test"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadGateway, rec.Code)
}

func TestSignup_DevModeReturnsVerifyURL(t *testing.T) {
	fp := &fakeProvisioner{}
	e := echo.New()
	signup.New(fp, true, "http://localhost:5173").Register(e)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/signup",
		strings.NewReader(`{"company":"DevCo","email":"dev@dev.test","display_name":"Dev User"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var out signup.Response
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	assert.NotEmpty(t, out.VerifyURL, "dev mode must return verify_url")
	assert.Contains(t, out.VerifyURL, "TEST-CODE")
	assert.Contains(t, out.VerifyURL, "zid-user-1")
}

func TestSignup_DuplicateEmailReturns409(t *testing.T) {
	fp := &fakeProvisioner{createUserErr: errors.New("zitadel: POST /v2/users/human: status 409: User already exists")}
	e := echo.New()
	signup.New(fp, false, "http://localhost:5173").Register(e)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/signup",
		strings.NewReader(`{"company":"Acme","email":"dup@acme.test"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusConflict, rec.Code)
}
