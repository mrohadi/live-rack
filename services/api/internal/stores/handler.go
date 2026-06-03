package stores

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"

	pkgauth "github.com/live-rack/pkg/auth"
	"github.com/live-rack/pkg/domain"
	"github.com/live-rack/pkg/store"
)

// StoreQuerier is the narrow DB interface the handler needs.
type StoreQuerier interface {
	ListStoresByOrg(ctx context.Context, orgID uuid.UUID) ([]store.Store, error)
	CreateStore(ctx context.Context, arg store.CreateStoreParams) (store.Store, error)
}

// Handler serves store management endpoints.
type Handler struct {
	q StoreQuerier
}

func New(q StoreQuerier) *Handler { return &Handler{q: q} }

// Register mounts authenticated store routes onto g (/api/v1).
func (h *Handler) Register(g *echo.Group) {
	g.GET("/stores", h.List)
	g.POST("/stores", h.Create)
}

// StoreResponse is the public representation of a store.
type StoreResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Address  string `json:"address,omitempty"`
	Timezone string `json:"timezone"`
}

// List godoc
//
//	@Summary	List stores for the authenticated org
//	@Tags		stores
//	@Produce	json
//	@Success	200	{array}	StoreResponse
//	@Router		/stores [get]
func (h *Handler) List(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	rows, err := h.q.ListStoresByOrg(c.Request().Context(), p.OrgID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list stores")
	}
	out := make([]StoreResponse, len(rows))
	for i, s := range rows {
		out[i] = toResponse(s)
	}
	return c.JSON(http.StatusOK, out)
}

// CreateRequest is the payload for creating a new store.
type CreateRequest struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Timezone string `json:"timezone"`
}

// Create godoc
//
//	@Summary	Create a new store (admin only)
//	@Tags		stores
//	@Accept		json
//	@Produce	json
//	@Param		body	body		CreateRequest	true	"Store details"
//	@Success	201		{object}	StoreResponse
//	@Router		/stores [post]
func (h *Handler) Create(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	if !domain.Can(p.Role, domain.PermEditUsers) {
		return echo.NewHTTPError(http.StatusForbidden, "requires admin")
	}

	var req CreateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name required")
	}
	tz := strings.TrimSpace(req.Timezone)
	if tz == "" {
		tz = "UTC"
	}

	s, err := h.q.CreateStore(c.Request().Context(), store.CreateStoreParams{
		OrgID:    p.OrgID,
		Name:     name,
		Address:  pgtype.Text{String: req.Address, Valid: req.Address != ""},
		Timezone: tz,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "create store")
	}
	return c.JSON(http.StatusCreated, toResponse(s))
}

func toResponse(s store.Store) StoreResponse {
	return StoreResponse{
		ID:       s.ID.String(),
		Name:     s.Name,
		Address:  s.Address.String,
		Timezone: s.Timezone,
	}
}
