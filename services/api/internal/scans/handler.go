// Package scans provides the scan validation HTTP handler.
package scans

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	pkgauth "github.com/live-rack/pkg/auth"
	"github.com/live-rack/pkg/domain"
	"github.com/live-rack/pkg/store"
)

// ZoneGetter is the narrow store dependency the handler needs.
type ZoneGetter interface {
	GetZone(ctx context.Context, arg store.GetZoneParams) (store.Zone, error)
}

// Handler handles scan validation endpoints.
type Handler struct{ q ZoneGetter }

// New creates a Handler backed by q.
func New(q ZoneGetter) *Handler { return &Handler{q: q} }

// Register mounts scan routes onto g (expected: /api/v1/stores).
func (h *Handler) Register(g *echo.Group) {
	g.POST("/:storeID/scan/validate", h.Validate)
}

type validateRequest struct {
	ZoneID            uuid.UUID `json:"zone_id"`
	Category          string    `json:"category"`
	CurrentQty        int       `json:"current_qty"`
	ScanQty           int       `json:"scan_qty"`
	LastScanAt        time.Time `json:"last_scan_at"` // TODO LR-205: source from scan_events server-side
	DualScanConfirmed bool      `json:"dual_scan_confirmed"`
}

// ValidateResponse is the scan decision returned to the client.
type ValidateResponse struct {
	Valid            bool   `json:"valid"`
	Code             string `json:"code,omitempty"`
	Reason           string `json:"reason,omitempty"`
	RequiresDualScan bool   `json:"requires_dual_scan,omitempty"`
}

// Validate godoc
//
//	@Summary		Validate a scan against zone rules
//	@Tags			scans
//	@Accept			json
//	@Produce		json
//	@Param			storeID	path		string			true	"Store UUID"
//	@Param			body	body		validateRequest	true	"Scan body"
//	@Success		200		{object}	ValidateResponse
//	@Failure		400		{object}	ValidateResponse
//	@Failure		404		{object}	ValidateResponse
//	@Router			/stores/{storeID}/scan/validate [post]
func (h *Handler) Validate(c echo.Context) error {
	orgID, err := orgIDFrom(c)
	if err != nil {
		return err
	}

	var req validateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.ScanQty <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "scan_qty must be positive")
	}

	z, err := h.q.GetZone(c.Request().Context(), store.GetZoneParams{ID: req.ZoneID, OrgID: orgID})
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "zone not found")
	}

	zone := domain.Zone{Capacity: int(z.Capacity), Constraints: z.Constraints}
	verr := zone.ValidateScan(domain.ScanRequest{
		Category:          req.Category,
		CurrentQty:        req.CurrentQty,
		ScanQty:           req.ScanQty,
		LastScanAt:        req.LastScanAt,
		Now:               time.Now(),
		DualScanConfirmed: req.DualScanConfirmed,
	})
	if verr == nil {
		return c.JSON(http.StatusOK, ValidateResponse{Valid: true})
	}
	return c.JSON(http.StatusOK, decision(verr))
}

func decision(err error) ValidateResponse {
	switch {
	case errors.Is(err, domain.ErrDualScanRequired):
		return ValidateResponse{Code: "dual_scan_required", Reason: err.Error(), RequiresDualScan: true}
	case errors.Is(err, domain.ErrCategoryNotAllowed):
		return ValidateResponse{Code: "category_not_allowed", Reason: err.Error()}
	case errors.Is(err, domain.ErrCategoryDenied):
		return ValidateResponse{Code: "category_denied", Reason: err.Error()}
	case errors.Is(err, domain.ErrCapacityExceeded):
		return ValidateResponse{Code: "capacity_exceeded", Reason: err.Error()}
	case errors.Is(err, domain.ErrMaxPerSKUExceeded):
		return ValidateResponse{Code: "max_per_sku_exceeded", Reason: err.Error()}
	case errors.Is(err, domain.ErrDwellViolation):
		return ValidateResponse{Code: "dwell_violation", Reason: err.Error()}
	default:
		return ValidateResponse{Code: "invalid", Reason: err.Error()}
	}
}

func orgIDFrom(c echo.Context) (uuid.UUID, error) {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return uuid.Nil, echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	return p.OrgID, nil
}
