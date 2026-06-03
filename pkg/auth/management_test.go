package auth_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/live-rack/pkg/auth"
)

func TestZitadelManagement_CreateOrgUserGrant(t *testing.T) {
	var seen []string // method+path log
	var lastOrgHeader, lastAuth string

	var grantBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Method+" "+r.URL.Path)
		lastOrgHeader = r.Header.Get("x-zitadel-orgid")
		lastAuth = r.Header.Get("Authorization")

		switch r.URL.Path {
		case "/management/v1/orgs":
			_ = json.NewEncoder(w).Encode(map[string]string{"id": "org-123"})
		case "/v2/users/human":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			// org id must ride in the body for v2 user creation.
			org := body["organization"].(map[string]any)
			assert.Equal(t, "org-123", org["orgId"])
			_ = json.NewEncoder(w).Encode(map[string]string{"userId": "user-456"})
		case "/management/v1/projects/proj-1":
			// The project is owned by a different org than the new tenant.
			_ = json.NewEncoder(w).Encode(map[string]any{
				"project": map[string]any{"details": map[string]any{"resourceOwner": "owner-org"}},
			})
		case "/management/v1/projects/proj-1/grants/_search":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": []any{}})
		case "/management/v1/projects/proj-1/grants":
			_ = json.NewEncoder(w).Encode(map[string]string{"grantId": "grant-1"})
		default: // user grant
			_ = json.NewDecoder(r.Body).Decode(&grantBody)
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer srv.Close()

	m := auth.NewZitadelManagement(srv.URL, "proj-1", "http://localhost:5173", auth.StaticToken("svc-token"))
	ctx := context.Background()

	orgID, err := m.CreateOrg(ctx, "Acme Co")
	require.NoError(t, err)
	assert.Equal(t, "org-123", orgID)

	userID, err := m.CreateHumanUser(ctx, orgID, "ada@acme.test", "Ada Lovelace")
	require.NoError(t, err)
	assert.Equal(t, "user-456", userID)

	require.NoError(t, m.GrantProjectRole(ctx, orgID, userID, "admin"))

	// A tenant org first gets the project granted to it, then the user grant
	// references that project grant.
	assert.Equal(t, []string{
		"POST /management/v1/orgs",
		"POST /v2/users/human",
		"GET /management/v1/projects/proj-1",
		"POST /management/v1/projects/proj-1/grants/_search",
		"POST /management/v1/projects/proj-1/grants",
		"POST /management/v1/users/user-456/grants",
	}, seen)
	assert.Equal(t, "grant-1", grantBody["projectGrantId"])
	assert.Equal(t, "org-123", lastOrgHeader)
	assert.Equal(t, "Bearer svc-token", lastAuth)
}

func TestZitadelManagement_PropagatesError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"message":"email already exists"}`))
	}))
	defer srv.Close()

	m := auth.NewZitadelManagement(srv.URL, "proj-1", "http://localhost:5173", auth.StaticToken("t"))
	_, err := m.CreateHumanUser(context.Background(), "org-1", "dup@acme.test", "Dup User")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status 409")
}

func TestStaticToken_RejectsEmpty(t *testing.T) {
	_, err := auth.StaticToken("")(context.Background())
	require.Error(t, err)
}
