package passwordreset_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/live-rack/pkg/audit"
	"github.com/live-rack/pkg/store"
	"github.com/live-rack/services/api/internal/passwordreset"
)

type fakeResolver struct{}

func (fakeResolver) GetUserByIdpID(_ context.Context, _ string) (store.User, error) {
	return store.User{ID: uuid.New(), OrgID: uuid.New()}, nil
}

type fakeAuditor struct{ actions []string }

func (f *fakeAuditor) Write(_ context.Context, e audit.Entry) error {
	f.actions = append(f.actions, e.Action)
	return nil
}

type fakeResetter struct {
	foundID                  string
	findErr                  error
	sentForUser              string
	resetUser, resetCode, pw string
	resetErr                 error
}

func (f *fakeResetter) FindUserByEmail(_ context.Context, _ string) (string, error) {
	return f.foundID, f.findErr
}

func (f *fakeResetter) SendPasswordResetCode(_ context.Context, userID string) error {
	f.sentForUser = userID
	return nil
}

func (f *fakeResetter) ResetPassword(_ context.Context, userID, code, password string) error {
	if f.resetErr != nil {
		return f.resetErr
	}
	f.resetUser, f.resetCode, f.pw = userID, code, password
	return nil
}

func serve(t *testing.T, f *fakeResetter, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	passwordreset.New(f, fakeResolver{}, &fakeAuditor{}).Register(e)
	req := httptest.NewRequestWithContext(context.Background(),
		http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestForgot_SendsWhenFound(t *testing.T) {
	f := &fakeResetter{foundID: "u1"}
	rec := serve(t, f, "/api/v1/password/forgot", `{"email":"ADA@acme.test"}`)
	require.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "u1", f.sentForUser)
}

func TestForgot_NoEnumerationWhenMissing(t *testing.T) {
	f := &fakeResetter{foundID: ""}
	rec := serve(t, f, "/api/v1/password/forgot", `{"email":"nobody@acme.test"}`)
	require.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "", f.sentForUser)
}

func TestForgot_NoEnumerationOnError(t *testing.T) {
	f := &fakeResetter{findErr: errors.New("boom")}
	rec := serve(t, f, "/api/v1/password/forgot", `{"email":"ada@acme.test"}`)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestForgot_RejectsBadEmail(t *testing.T) {
	rec := serve(t, &fakeResetter{}, "/api/v1/password/forgot", `{"email":"not-an-email"}`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestReset_SetsPassword(t *testing.T) {
	f := &fakeResetter{}
	rec := serve(t, f, "/api/v1/password/reset",
		`{"user_id":"u1","code":"ABC123","password":"Brandnew123!"}`)
	require.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "u1", f.resetUser)
	assert.Equal(t, "ABC123", f.resetCode)
	assert.Equal(t, "Brandnew123!", f.pw)
}

func TestReset_RejectsShortPassword(t *testing.T) {
	rec := serve(t, &fakeResetter{}, "/api/v1/password/reset",
		`{"user_id":"u1","code":"ABC","password":"short"}`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestReset_InvalidCode(t *testing.T) {
	f := &fakeResetter{resetErr: errors.New("bad")}
	rec := serve(t, f, "/api/v1/password/reset",
		`{"user_id":"u1","code":"BAD","password":"Brandnew123!"}`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
