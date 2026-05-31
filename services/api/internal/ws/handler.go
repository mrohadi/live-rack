package ws

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"

	pkgauth "github.com/live-rack/pkg/auth"
)

// TODO LR-206: tighten CheckOrigin to allowed web origins before GA.
var upgrader = websocket.Upgrader{CheckOrigin: func(_ *http.Request) bool { return true }}

// Handler upgrades authenticated requests and registers clients on the hub.
type Handler struct {
	hub      *Hub
	verifier pkgauth.Verifier
}

// NewHandler wires the hub + token verifier.
func NewHandler(hub *Hub, v pkgauth.Verifier) *Handler {
	return &Handler{hub: hub, verifier: v}
}

// Register mounts GET /api/v1/ws (auth via ?token= query param).
func (h *Handler) Register(e *echo.Echo) {
	e.GET("/api/v1/ws", h.serve)
}

func (h *Handler) serve(c echo.Context) error {
	token := c.QueryParam("token")
	if token == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "missing token")
	}
	r := c.Request().Clone(c.Request().Context())
	r.Header.Set("Authorization", "Bearer "+token)
	p, err := h.verifier.VerifyRequest(r)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err // upgrader already wrote the error
	}

	client := &Client{orgID: p.OrgID.String(), send: make(chan []byte, 64)}
	h.hub.register(client)
	go client.writePump(conn)
	client.readPump(conn, client.hubOf(h)) // blocks until disconnect
	return nil
}

func (c *Client) hubOf(h *Handler) *Hub { return h.hub }
