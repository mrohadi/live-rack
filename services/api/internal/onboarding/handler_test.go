package onboarding_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/live-rack/services/api/internal/onboarding"
)

type fakeCompleter struct {
	verifiedUser, verifiedCode string
	pwUser, pwOrg, pw          string
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

func serve(t *testing.T, f *fakeCompleter, body string) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	onboarding.New(f).Register(e)
	req := httptest.NewRequestWithContext(context.Background(),
		http.MethodPost, "/api/v1/onboard/complete", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestComplete_VerifiesThenSetsPassword(t *testing.T) {
	f := &fakeCompleter{}
	rec := serve(t, f,
		`{"user_id":"u1","org_id":"o1","code":"ABC123","password":"Sup3rSecret!"}`)
	require.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "u1", f.verifiedUser)
	assert.Equal(t, "ABC123", f.verifiedCode)
	assert.Equal(t, "u1", f.pwUser)
	assert.Equal(t, "o1", f.pwOrg)
	assert.Equal(t, "Sup3rSecret!", f.pw)
}

func TestComplete_RejectsShortPassword(t *testing.T) {
	rec := serve(t, &fakeCompleter{}, `{"user_id":"u1","org_id":"o1","code":"ABC","password":"short"}`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestComplete_RejectsMissingFields(t *testing.T) {
	rec := serve(t, &fakeCompleter{}, `{"password":"Sup3rSecret!"}`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestComplete_InvalidCode(t *testing.T) {
	f := &fakeCompleter{verifyErr: errors.New("bad")}
	rec := serve(t, f, `{"user_id":"u1","org_id":"o1","code":"BAD","password":"Sup3rSecret!"}`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
