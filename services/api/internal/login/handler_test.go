package login_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pkgauth "github.com/live-rack/pkg/auth"
	"github.com/live-rack/services/api/internal/login"
)

type fakeLogin struct {
	startName, pwd, totp           string
	finalizeAuthReq, finalizeSessn string
	mfaRequired                    bool
	startErr, pwdErr, finalizeErr  error
}

func (f *fakeLogin) StartSession(_ context.Context, loginName string) (pkgauth.Session, bool, error) {
	f.startName = loginName
	if f.startErr != nil {
		return pkgauth.Session{}, false, f.startErr
	}
	return pkgauth.Session{SessionID: "s1", SessionToken: "t1"}, f.mfaRequired, nil
}

func (f *fakeLogin) CheckPassword(_ context.Context, _, _, password string) (pkgauth.Session, error) {
	f.pwd = password
	if f.pwdErr != nil {
		return pkgauth.Session{}, f.pwdErr
	}
	return pkgauth.Session{SessionID: "s1", SessionToken: "t2"}, nil
}

func (f *fakeLogin) CheckTOTP(_ context.Context, _, _, code string) (pkgauth.Session, error) {
	f.totp = code
	return pkgauth.Session{SessionID: "s1", SessionToken: "t3"}, nil
}

func (f *fakeLogin) Finalize(_ context.Context, authReq, sessID, _ string) (string, error) {
	f.finalizeAuthReq, f.finalizeSessn = authReq, sessID
	if f.finalizeErr != nil {
		return "", f.finalizeErr
	}
	return "http://localhost:5173/callback?code=abc&state=xyz", nil
}

func serve(t *testing.T, f *fakeLogin, method, target, body string) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	login.New(f).Register(e)
	req := httptest.NewRequestWithContext(context.Background(), method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestStart_OpensSession(t *testing.T) {
	f := &fakeLogin{}
	rec := serve(t, f, http.MethodPost, "/api/v1/login/start", `{"login_name":"ADA@x.io"}`)
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "ada@x.io", f.startName) // normalised
	var out struct {
		SessionID    string `json:"session_id"`
		SessionToken string `json:"session_token"`
		MFARequired  bool   `json:"mfa_required"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	assert.Equal(t, "s1", out.SessionID)
	assert.Equal(t, "t1", out.SessionToken)
	assert.False(t, out.MFARequired)
}

func TestStart_ReportsMFARequired(t *testing.T) {
	f := &fakeLogin{mfaRequired: true}
	rec := serve(t, f, http.MethodPost, "/api/v1/login/start", `{"login_name":"ada@x.io"}`)
	require.Equal(t, http.StatusOK, rec.Code)
	var out struct {
		MFARequired bool `json:"mfa_required"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	assert.True(t, out.MFARequired)
}

func TestStart_RejectsEmpty(t *testing.T) {
	rec := serve(t, &fakeLogin{}, http.MethodPost, "/api/v1/login/start", `{"login_name":"  "}`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPassword_ChecksAndRotatesToken(t *testing.T) {
	f := &fakeLogin{}
	rec := serve(t, f, http.MethodPost, "/api/v1/login/password",
		`{"session_id":"s1","session_token":"t1","password":"hunter2"}`)
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "hunter2", f.pwd)
	var out map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	assert.Equal(t, "t2", out["session_token"])
}

func TestPassword_InvalidCredentials401(t *testing.T) {
	f := &fakeLogin{pwdErr: assert.AnError}
	rec := serve(t, f, http.MethodPost, "/api/v1/login/password",
		`{"session_id":"s1","session_token":"t1","password":"wrong"}`)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestFinalize_ReturnsCallbackURL(t *testing.T) {
	f := &fakeLogin{}
	rec := serve(t, f, http.MethodPost, "/api/v1/login/finalize",
		`{"auth_request_id":"ar1","session_id":"s1","session_token":"t2"}`)
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "ar1", f.finalizeAuthReq)
	var out login.FinalizeResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	assert.Contains(t, out.CallbackURL, "/callback?code=")
}

func TestFinalize_RejectsMissingFields(t *testing.T) {
	rec := serve(t, &fakeLogin{}, http.MethodPost, "/api/v1/login/finalize",
		`{"auth_request_id":"ar1"}`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
