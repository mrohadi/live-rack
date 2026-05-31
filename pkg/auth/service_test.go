package auth_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/live-rack/pkg/auth"
	"github.com/live-rack/pkg/domain"
)

func TestHashToken_StableAndDistinct(t *testing.T) {
	h := auth.HashToken("lrk_abc")
	assert.Len(t, h, 64)
	assert.Equal(t, h, auth.HashToken("lrk_abc"))
	assert.NotEqual(t, h, auth.HashToken("lrk_xyz"))
}

func TestIsServiceToken(t *testing.T) {
	assert.True(t, auth.IsServiceToken("lrk_deadbeef"))
	assert.False(t, auth.IsServiceToken("eyJhbGci..."))
}

type fakeLookup struct {
	org     uuid.UUID
	gotHash string
	err     error
}

func (f *fakeLookup) ResolveServiceToken(_ context.Context, hash string) (*domain.Principal, error) {
	f.gotHash = hash
	if f.err != nil {
		return nil, f.err
	}
	return &domain.Principal{UserID: uuid.New(), OrgID: f.org}, nil
}

func req(token string) *http.Request {
	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	return r
}

func TestServiceVerifier_ResolvesServicePrincipal(t *testing.T) {
	org := uuid.New()
	lk := &fakeLookup{org: org}
	p, err := auth.NewServiceVerifier(lk).VerifyRequest(req("lrk_secret"))
	require.NoError(t, err)
	assert.Equal(t, org, p.OrgID)
	assert.Equal(t, domain.RoleService, p.Role)
	assert.Equal(t, auth.HashToken("lrk_secret"), lk.gotHash)
}

func TestServiceVerifier_RejectsNonService(t *testing.T) {
	_, err := auth.NewServiceVerifier(&fakeLookup{}).VerifyRequest(req("jwt-token"))
	assert.Error(t, err)
}

func TestServiceVerifier_PropagatesLookupError(t *testing.T) {
	_, err := auth.NewServiceVerifier(&fakeLookup{err: errors.New("revoked")}).VerifyRequest(req("lrk_x"))
	assert.Error(t, err)
}

type fakeOIDC struct{ called bool }

func (f *fakeOIDC) VerifyRequest(*http.Request) (*domain.Principal, error) {
	f.called = true
	return &domain.Principal{Role: domain.RoleAdmin}, nil
}

func TestCompositeVerifier_Routes(t *testing.T) {
	oidc := &fakeOIDC{}
	cv := auth.NewCompositeVerifier(auth.NewServiceVerifier(&fakeLookup{org: uuid.New()}), oidc)

	p, err := cv.VerifyRequest(req("lrk_tok"))
	require.NoError(t, err)
	assert.Equal(t, domain.RoleService, p.Role)
	assert.False(t, oidc.called, "service token should not hit OIDC")

	p2, err := cv.VerifyRequest(req("jwt"))
	require.NoError(t, err)
	assert.Equal(t, domain.RoleAdmin, p2.Role)
	assert.True(t, oidc.called)
}
