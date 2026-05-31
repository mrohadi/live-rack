// Package search serves the ⌘K global search endpoint backed by PG trigram.
package search

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	pkgauth "github.com/live-rack/pkg/auth"
	"github.com/live-rack/pkg/store"
)

const (
	defaultLimit = 20
	maxLimit     = 50
)

// Searcher is the narrow store dependency the handler needs.
type Searcher interface {
	SearchEntities(ctx context.Context, arg store.SearchEntitiesParams) ([]store.SearchEntitiesRow, error)
}

// Handler serves the global search endpoint.
type Handler struct {
	q Searcher
}

// New creates a Handler.
func New(q Searcher) *Handler {
	return &Handler{q: q}
}

// Register mounts the search route onto g (expected: /api/v1).
func (h *Handler) Register(g *echo.Group) {
	g.GET("/search", h.Search)
}

// Result is one search hit returned to the client.
type Result struct {
	Kind     string    `json:"kind"`
	ID       uuid.UUID `json:"id"`
	Label    string    `json:"label"`
	Sublabel string    `json:"sublabel"`
	Score    float32   `json:"score"`
}

// Search godoc
//
//	@Summary		Fuzzy ⌘K search over items and zones
//	@Tags			search
//	@Produce		json
//	@Param			q		query		string	true	"Search query (min 2 chars)"
//	@Param			limit	query		int		false	"Max results (default 20, max 50)"
//	@Success		200		{array}		Result
//	@Failure		400		{object}	map[string]string
//	@Router			/search [get]
func (h *Handler) Search(c echo.Context) error {
	p, err := pkgauth.PrincipalFrom(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	q := strings.TrimSpace(c.QueryParam("q"))
	if len(q) < 2 {
		return echo.NewHTTPError(http.StatusBadRequest, "query must be at least 2 characters")
	}

	rows, err := h.q.SearchEntities(c.Request().Context(), store.SearchEntitiesParams{
		Query:      q,
		OrgID:      p.OrgID,
		MaxResults: parseLimit(c.QueryParam("limit")),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "search")
	}

	out := make([]Result, 0, len(rows))
	for _, r := range rows {
		out = append(out, Result{
			Kind:     r.Kind,
			ID:       r.ID,
			Label:    r.Label,
			Sublabel: r.Sublabel,
			Score:    r.Score,
		})
	}
	return c.JSON(http.StatusOK, out)
}

// parseLimit clamps a raw limit param to [1, maxLimit], falling back to default.
func parseLimit(raw string) int32 {
	n, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || n < 1 {
		return defaultLimit
	}
	if n > maxLimit {
		return maxLimit
	}
	return int32(n)
}
