// Package zones provides the Echo HTTP handler for zone CRUD operations.
package zones

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	pkgauth "github.com/live-rack/pkg/auth"
	"github.com/live-rack/pkg/store"
)

// ZoneStore is the narrow interface the handler requires.
type ZoneStore interface {
	CreateZone(ctx context.Context, arg store.CreateZoneParams) (store.Zone, error)
	GetZone(ctx context.Context, arg store.GetZoneParams) (store.Zone, error)
	ListZonesByStore(ctx context.Context, arg store.ListZonesByStoreParams) ([]store.Zone, error)
	UpdateZone(ctx context.Context, arg store.UpdateZoneParams) (store.Zone, error)
	DeleteZone(ctx context.Context, arg store.DeleteZoneParams) error
}

// Handler handles zone endpoints.
type Handler struct {
	q ZoneStore
}

// New creates a Handler backed by q.
func New(q ZoneStore) *Handler {
	return &Handler{q: q}
}

// Register mounts zone routes onto g (expected: /api/v1/stores).
func (h *Handler) Register(g *echo.Group) {
	g.GET("/:storeID/zones", h.List)
	g.POST("/:storeID/zones", h.Create)
	g.GET("/:storeID/zones/:id", h.Get)
	g.PUT("/:storeID/zones/:id", h.Update)
	g.DELETE("/:storeID/zones/:id", h.Delete)
}

// zoneRequest is the JSON body for create/update.
type zoneRequest struct {
	Name        string          `json:"name"`
	Type        store.ZoneType  `json:"type"`
	X           float64         `json:"x"`
	Y           float64         `json:"y"`
	Width       float64         `json:"width"`
	Height      float64         `json:"height"`
	Color       string          `json:"color"`
	Capacity    int32           `json:"capacity"`
	Constraints json.RawMessage `json:"constraints"`
}

func (r *zoneRequest) validate() error {
	if r.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}
	if r.Width <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "width must be positive")
	}
	if r.Height <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "height must be positive")
	}
	if r.Capacity < 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "capacity must be non-negative")
	}
	return nil
}

func parseStoreID(c echo.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param("storeID"))
	if err != nil {
		return uuid.Nil, echo.NewHTTPError(http.StatusBadRequest, "invalid storeID")
	}
	return id, nil
}

func parseZoneID(c echo.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return uuid.Nil, echo.NewHTTPError(http.StatusBadRequest, "invalid zone id")
	}
	return id, nil
}

func orgIDFrom(c echo.Context) (uuid.UUID, error) {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return uuid.Nil, echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	return p.OrgID, nil
}

func constraintsOrDefault(raw json.RawMessage) []byte {
	if len(raw) == 0 {
		return []byte(`{}`)
	}
	return []byte(raw)
}

// List godoc
//
//	@Summary		List zones by store
//	@Tags			zones
//	@Produce		json
//	@Param			storeID	path		string		true	"Store UUID"
//	@Success		200		{array}		store.Zone
//	@Failure		400		{object}	echo.HTTPError
//	@Failure		401		{object}	echo.HTTPError
//	@Router			/stores/{storeID}/zones [get]
func (h *Handler) List(c echo.Context) error {
	storeID, err := parseStoreID(c)
	if err != nil {
		return err
	}
	orgID, err := orgIDFrom(c)
	if err != nil {
		return err
	}

	list, err := h.q.ListZonesByStore(c.Request().Context(), store.ListZonesByStoreParams{
		StoreID: storeID,
		OrgID:   orgID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if list == nil {
		list = []store.Zone{}
	}
	return c.JSON(http.StatusOK, list)
}

// Create godoc
//
//	@Summary		Create a zone
//	@Tags			zones
//	@Accept			json
//	@Produce		json
//	@Param			storeID	path		string		true	"Store UUID"
//	@Param			body	body		zoneRequest	true	"Zone body"
//	@Success		201		{object}	store.Zone
//	@Failure		400		{object}	echo.HTTPError
//	@Failure		401		{object}	echo.HTTPError
//	@Router			/stores/{storeID}/zones [post]
func (h *Handler) Create(c echo.Context) error {
	storeID, err := parseStoreID(c)
	if err != nil {
		return err
	}
	orgID, err := orgIDFrom(c)
	if err != nil {
		return err
	}

	var req zoneRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := req.validate(); err != nil {
		return err
	}

	z, err := h.q.CreateZone(c.Request().Context(), store.CreateZoneParams{
		OrgID:       orgID,
		StoreID:     storeID,
		Name:        req.Name,
		Type:        req.Type,
		X:           req.X,
		Y:           req.Y,
		Width:       req.Width,
		Height:      req.Height,
		Color:       req.Color,
		Capacity:    req.Capacity,
		Constraints: constraintsOrDefault(req.Constraints),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, z)
}

// Get godoc
//
//	@Summary		Get a zone by ID
//	@Tags			zones
//	@Produce		json
//	@Param			storeID	path		string	true	"Store UUID"
//	@Param			id		path		string	true	"Zone UUID"
//	@Success		200		{object}	store.Zone
//	@Failure		404		{object}	echo.HTTPError
//	@Router			/stores/{storeID}/zones/{id} [get]
func (h *Handler) Get(c echo.Context) error {
	zoneID, err := parseZoneID(c)
	if err != nil {
		return err
	}
	orgID, err := orgIDFrom(c)
	if err != nil {
		return err
	}

	z, err := h.q.GetZone(c.Request().Context(), store.GetZoneParams{
		ID:    zoneID,
		OrgID: orgID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "zone not found")
	}
	return c.JSON(http.StatusOK, z)
}

// Update godoc
//
//	@Summary		Update a zone
//	@Tags			zones
//	@Accept			json
//	@Produce		json
//	@Param			storeID	path		string		true	"Store UUID"
//	@Param			id		path		string		true	"Zone UUID"
//	@Param			body	body		zoneRequest	true	"Zone body"
//	@Success		200		{object}	store.Zone
//	@Failure		400		{object}	echo.HTTPError
//	@Failure		404		{object}	echo.HTTPError
//	@Router			/stores/{storeID}/zones/{id} [put]
func (h *Handler) Update(c echo.Context) error {
	zoneID, err := parseZoneID(c)
	if err != nil {
		return err
	}
	orgID, err := orgIDFrom(c)
	if err != nil {
		return err
	}

	var req zoneRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := req.validate(); err != nil {
		return err
	}

	z, err := h.q.UpdateZone(c.Request().Context(), store.UpdateZoneParams{
		ID:          zoneID,
		OrgID:       orgID,
		Name:        req.Name,
		Type:        req.Type,
		X:           req.X,
		Y:           req.Y,
		Width:       req.Width,
		Height:      req.Height,
		Color:       req.Color,
		Capacity:    req.Capacity,
		Constraints: constraintsOrDefault(req.Constraints),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "zone not found")
	}
	return c.JSON(http.StatusOK, z)
}

// Delete godoc
//
//	@Summary		Delete a zone
//	@Tags			zones
//	@Param			storeID	path	string	true	"Store UUID"
//	@Param			id		path	string	true	"Zone UUID"
//	@Success		204
//	@Failure		404	{object}	echo.HTTPError
//	@Router			/stores/{storeID}/zones/{id} [delete]
func (h *Handler) Delete(c echo.Context) error {
	zoneID, err := parseZoneID(c)
	if err != nil {
		return err
	}
	orgID, err := orgIDFrom(c)
	if err != nil {
		return err
	}

	if err := h.q.DeleteZone(c.Request().Context(), store.DeleteZoneParams{
		ID:    zoneID,
		OrgID: orgID,
	}); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "zone not found")
	}
	return c.NoContent(http.StatusNoContent)
}
