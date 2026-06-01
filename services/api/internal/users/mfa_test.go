package users_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pkgauth "github.com/live-rack/pkg/auth"
	"github.com/live-rack/pkg/domain"
	"github.com/live-rack/services/api/internal/users"
)

type fakeEnroller struct {
	verifiedUserID, verifiedCode string
	verifyErr                    error
}

func (f *fakeEnroller) RegisterTOTP(_ context.Context, _ string) (string, string, error) {
	return "otpauth://totp/live-rack:ada?secret=ABC", "ABC", nil
}

func (f *fakeEnroller) VerifyTOTP(_ context.Context, userID, code string) error {
	if f.verifyErr != nil {
		return f.verifyErr
	}
	f.verifiedUserID, f.verifiedCode = userID, code
	return nil
}

type fakeMFAStore struct{ enabledFor uuid.UUID }

func (f *fakeMFAStore) SetUserMFA(_ context.Context, userID, _ uuid.UUID, _ bool) error {
	f.enabledFor = userID
	return nil
}

func serveMFA(t *testing.T, p *domain.Principal, e *fakeEnroller, s *fakeMFAStore, method, target, body string) *httptest.ResponseRecorder {
	t.Helper()
	srv := echo.New()
	users.NewMFA(e, s, &fakeAuditor{}).Register(srv.Group("/api/v1"))
	req := httptest.NewRequestWithContext(
		pkgauth.WithPrincipal(context.Background(), p), method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	return rec
}

func enrollee() *domain.Principal {
	return &domain.Principal{
		UserID:    uuid.New(),
		IDPUserID: "zid-user-1",
		OrgID:     uuid.New(),
		Role:      domain.RoleStaff,
	}
}

func TestMFAStart_ReturnsProvisioning(t *testing.T) {
	rec := serveMFA(t, enrollee(), &fakeEnroller{}, &fakeMFAStore{},
		http.MethodPost, "/api/v1/me/2fa/totp", "")
	require.Equal(t, http.StatusOK, rec.Code)

	var out users.StartTOTPResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	assert.Contains(t, out.URI, "otpauth://")
	assert.Equal(t, "ABC", out.Secret)
}

func TestMFAVerify_RecordsCoverage(t *testing.T) {
	p := enrollee()
	enr := &fakeEnroller{}
	st := &fakeMFAStore{}
	rec := serveMFA(t, p, enr, st, http.MethodPost, "/api/v1/me/2fa/totp/verify",
		`{"code":"123456"}`)
	require.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "zid-user-1", enr.verifiedUserID)
	assert.Equal(t, "123456", enr.verifiedCode)
	assert.Equal(t, p.UserID, st.enabledFor)
}

func TestMFAVerify_InvalidCode(t *testing.T) {
	rec := serveMFA(t, enrollee(), &fakeEnroller{verifyErr: errors.New("bad")}, &fakeMFAStore{},
		http.MethodPost, "/api/v1/me/2fa/totp/verify", `{"code":"000000"}`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestMFAVerify_MissingCode(t *testing.T) {
	rec := serveMFA(t, enrollee(), &fakeEnroller{}, &fakeMFAStore{},
		http.MethodPost, "/api/v1/me/2fa/totp/verify", `{}`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
