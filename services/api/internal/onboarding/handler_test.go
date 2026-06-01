package onboarding_test

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

	pkgauth "github.com/live-rack/pkg/auth"
	"github.com/live-rack/services/api/internal/onboarding"
)

type fakeCompleter struct {
	verifiedUser, verifiedCode string
	pwUser, pwOrg, pw          string
	totpVerifiedUser           string
	verifyErr, pwErr           error
}

func (f *fakeCompleter) VerifyEmail(_ context.Context, userID, code string) error {
	if f.verifyErr != nil {
		return f.verifyErr
	}
	f.verifiedUser, f.verifiedCode = userID, code
	return nil
}

func (f *fakeCompleter) SetPassword(_ context.Context, orgID, userID, password string) error {
	if f.pwErr != nil {
		return f.pwErr
	}
	f.pwOrg, f.pwUser, f.pw = orgID, userID, password
	return nil
}

func (f *fakeCompleter) GetLoginName(_ context.Context, _ string) (string, error) {
	return "ada@acme.test", nil
}

func (f *fakeCompleter) RegisterTOTP(_ context.Context, _ string) (string, string, error) {
	return "otpauth://totp/x?secret=ABC", "ABC", nil
}

func (f *fakeCompleter) VerifyTOTP(_ context.Context, userID, _ string) error {
	f.totpVerifiedUser = userID
	return nil
}

// fakeChecker accepts the password "right" and rejects everything else.
type fakeChecker struct{}

func (fakeChecker) StartSession(_ context.Context, _ string) (pkgauth.Session, bool, error) {
	return pkgauth.Session{SessionID: "s1", SessionToken: "t1"}, false, nil
}

func (fakeChecker) CheckPassword(_ context.Context, _, _, password string) (pkgauth.Session, error) {
	if password != "Sup3rSecret!" {
		return pkgauth.Session{}, errors.New("wrong password")
	}
	return pkgauth.Session{SessionID: "s1", SessionToken: "t2"}, nil
}

func serve(t *testing.T, f *fakeCompleter, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	onboarding.New(f, fakeChecker{}).Register(e)
	req := httptest.NewRequestWithContext(context.Background(),
		http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

const completePath = "/api/v1/onboard/complete"

func TestComplete_VerifiesThenSetsPassword(t *testing.T) {
	f := &fakeCompleter{}
	rec := serve(t, f, completePath,
		`{"user_id":"u1","org_id":"o1","code":"ABC123","password":"Sup3rSecret!"}`)
	require.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "u1", f.verifiedUser)
	assert.Equal(t, "ABC123", f.verifiedCode)
	assert.Equal(t, "u1", f.pwUser)
	assert.Equal(t, "o1", f.pwOrg)
	assert.Equal(t, "Sup3rSecret!", f.pw)
}

func TestComplete_RejectsShortPassword(t *testing.T) {
	rec := serve(t, &fakeCompleter{}, completePath, `{"user_id":"u1","org_id":"o1","code":"ABC","password":"short"}`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestComplete_RejectsMissingFields(t *testing.T) {
	rec := serve(t, &fakeCompleter{}, completePath, `{"password":"Sup3rSecret!"}`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestComplete_InvalidCode(t *testing.T) {
	f := &fakeCompleter{verifyErr: errors.New("bad")}
	rec := serve(t, f, completePath, `{"user_id":"u1","org_id":"o1","code":"BAD","password":"Sup3rSecret!"}`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTOTPStart_GatedByPassword(t *testing.T) {
	rec := serve(t, &fakeCompleter{}, "/api/v1/onboard/totp/start",
		`{"user_id":"u1","password":"Sup3rSecret!"}`)
	require.Equal(t, http.StatusOK, rec.Code)
	var out onboarding.TOTPStartResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	assert.Contains(t, out.URI, "otpauth://")
	assert.Equal(t, "ABC", out.Secret)
}

func TestTOTPStart_WrongPassword(t *testing.T) {
	rec := serve(t, &fakeCompleter{}, "/api/v1/onboard/totp/start",
		`{"user_id":"u1","password":"nope"}`)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestTOTPVerify_GatedThenRecords(t *testing.T) {
	f := &fakeCompleter{}
	rec := serve(t, f, "/api/v1/onboard/totp/verify",
		`{"user_id":"u1","password":"Sup3rSecret!","code":"123456"}`)
	require.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "u1", f.totpVerifiedUser)
}

func TestTOTPVerify_WrongPassword(t *testing.T) {
	rec := serve(t, &fakeCompleter{}, "/api/v1/onboard/totp/verify",
		`{"user_id":"u1","password":"nope","code":"123456"}`)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
