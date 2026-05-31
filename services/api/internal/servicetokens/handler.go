// Package servicetokens issues opaque service tokens (first-class service
// principals). Creating one is a high-impact action: it requires the edit-users
// permission and a verified second factor.
package servicetokens

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	pkgauth "github.com/live-rack/pkg/auth"
	"github.com/live-rack/pkg/domain"
)

// Creator persists a service token's hash. *store.Queries satisfies it.
type Creator interface {
	CreateServiceToken(ctx context.Context, orgID uuid.UUID, name, hash string) (uuid.UUID, error)
}

// Handler serves service-token endpoints.
type Handler struct {
	q Creator
}

// New creates a Handler.
func New(q Creator) *Handler {
	return &Handler{q: q}
}

// Register mounts routes on the authenticated API group.
func (h *Handler) Register(g *echo.Group) {
	g.POST("/service-tokens", h.Create)
}

// generateToken returns a fresh opaque token and its storage hash.
func generateToken() (token, hash string, err error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", "", err
	}
	token = pkgauth.ServiceTokenPrefix + hex.EncodeToString(b[:])
	return token, pkgauth.HashToken(token), nil
}

// CreateRequest names a new service token.
type CreateRequest struct {
	Name string `json:"name"`
}

// CreateResponse returns the plaintext token exactly once.
type CreateResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Token string `json:"token"`
}

// Create godoc
//
//	@Summary	Issue a service token (shown once); admin + MFA required
//	@Tags		service-tokens
//	@Accept		json
//	@Produce	json
//	@Param		body	body		CreateRequest	true	"Token name"
//	@Success	201		{object}	CreateResponse
//	@Router		/service-tokens [post]
func (h *Handler) Create(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	if !p.CanWithMFA(domain.PermEditUsers) {
		return echo.NewHTTPError(http.StatusForbidden, "requires admin + 2FA")
	}

	var req CreateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name required")
	}

	token, hash, err := generateToken()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "generate token")
	}
	id, err := h.q.CreateServiceToken(c.Request().Context(), p.OrgID, name, hash)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "create token")
	}
	return c.JSON(http.StatusCreated, CreateResponse{ID: id.String(), Name: name, Token: token})
}
