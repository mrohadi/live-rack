package servicetokens_test

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

	pkgauth "github.com/live-rack/pkg/auth"
	"github.com/live-rack/pkg/domain"
	"github.com/live-rack/services/api/internal/servicetokens"
)

type fakeCreator struct {
	gotOrg  uuid.UUID
	gotHash string
}

func (f *fakeCreator) CreateServiceToken(_ context.Context, orgID uuid.UUID, _ string, hash string) (uuid.UUID, error) {
	f.gotOrg = orgID
	f.gotHash = hash
	return uuid.New(), nil
}

func post(t *testing.T, fc *fakeCreator, p *domain.Principal, body string) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	servicetokens.New(fc).Register(e.Group("/api/v1"))
	req := httptest.NewRequestWithContext(
		pkgauth.WithPrincipal(context.Background(), p), http.MethodPost, "/api/v1/service-tokens", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestCreate_AdminWithMFA(t *testing.T) {
	org := uuid.New()
	fc := &fakeCreator{}
	p := &domain.Principal{UserID: uuid.New(), OrgID: org, Role: domain.RoleAdmin, MFAVerified: true}
	rec := post(t, fc, p, `{"name":"Shopify sync"}`)

	require.Equal(t, http.StatusCreated, rec.Code)
	var out servicetokens.CreateResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	assert.True(t, strings.HasPrefix(out.Token, pkgauth.ServiceTokenPrefix))
	assert.Equal(t, org, fc.gotOrg)
	// stored hash matches the returned plaintext
	assert.Equal(t, pkgauth.HashToken(out.Token), fc.gotHash)
}

func TestCreate_AdminWithoutMFA_Forbidden(t *testing.T) {
	p := &domain.Principal{UserID: uuid.New(), OrgID: uuid.New(), Role: domain.RoleAdmin, MFAVerified: false}
	rec := post(t, &fakeCreator{}, p, `{"name":"x"}`)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestCreate_Manager_Forbidden(t *testing.T) {
	p := &domain.Principal{UserID: uuid.New(), OrgID: uuid.New(), Role: domain.RoleManager, MFAVerified: true}
	rec := post(t, &fakeCreator{}, p, `{"name":"x"}`)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestCreate_EmptyName(t *testing.T) {
	p := &domain.Principal{UserID: uuid.New(), OrgID: uuid.New(), Role: domain.RoleAdmin, MFAVerified: true}
	rec := post(t, &fakeCreator{}, p, `{"name":"  "}`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
