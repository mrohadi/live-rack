// Package signup serves public self-service registration: it provisions a new
// tenant org in Zitadel, creates the first user as its admin, and triggers a
// verification email. The internal org/user rows are created lazily by JIT
// provisioning on the new admin's first authenticated request.
package signup

import (
	"context"
	"crypto/rand"
	"log/slog"
	"math/big"
	"net/http"
	"net/mail"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/live-rack/pkg/domain"
)

// Provisioner is the Zitadel management surface signup needs.
// *auth.ZitadelManagement satisfies it.
type Provisioner interface {
	CreateOrg(ctx context.Context, name string) (string, error)
	CreateHumanUser(ctx context.Context, orgID, email, displayName string) (string, error)
	CreateHumanUserReturnCode(ctx context.Context, orgID, email, displayName string) (userID, code string, err error)
	GrantProjectRole(ctx context.Context, orgID, userID, role string) error
}

// Handler serves the public signup endpoint.
type Handler struct {
	zit    Provisioner
	dev    bool   // when true, return email verification code directly (no SMTP needed)
	appURL string // SPA origin for constructing verify-email URLs in dev mode
}

// New builds a signup Handler. Pass dev=true when ENV=development so the
// verification code is returned in the response instead of emailed.
func New(zit Provisioner, dev bool, appURL string) *Handler {
	return &Handler{zit: zit, dev: dev, appURL: appURL}
}

// Register mounts the public route on the root router (no auth middleware).
func (h *Handler) Register(e *echo.Echo) {
	e.POST("/api/v1/signup", h.Signup)
}

// Request is a self-service signup payload.
type Request struct {
	Company     string `json:"company"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

// Response confirms the tenant was provisioned and a verification email sent.
// In development mode, VerifyURL is populated so the caller can skip email.
type Response struct {
	OrgID     string `json:"org_id"`
	UserID    string `json:"user_id"`
	Status    string `json:"status"`
	VerifyURL string `json:"verify_url,omitempty"`
}

// Signup godoc
//
//	@Summary	Self-service signup: create a tenant org + admin user (public)
//	@Tags		signup
//	@Accept		json
//	@Produce	json
//	@Param		body	body		Request	true	"Signup details"
//	@Success	201		{object}	Response
//	@Router		/signup [post]
func (h *Handler) Signup(c echo.Context) error {
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	company := strings.TrimSpace(req.Company)
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if company == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "company required")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid email")
	}
	display := strings.TrimSpace(req.DisplayName)
	if display == "" {
		display = email
	}

	ctx := c.Request().Context()
	orgID, err := h.zit.CreateOrg(ctx, company+"-"+randSuffix(6))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, "create org")
	}

	var userID, verifyURL string
	if h.dev {
		var code string
		userID, code, err = h.zit.CreateHumanUserReturnCode(ctx, orgID, email, display)
		if err != nil {
			if strings.Contains(err.Error(), "User already exists") {
				return echo.NewHTTPError(http.StatusConflict, "email already registered")
			}
			return echo.NewHTTPError(http.StatusBadGateway, "create user")
		}
		verifyURL = h.appURL + "/verify-email?code=" + code + "&userID=" + userID + "&orgID=" + orgID
	} else {
		userID, err = h.zit.CreateHumanUser(ctx, orgID, email, display)
		if err != nil {
			if strings.Contains(err.Error(), "User already exists") {
				return echo.NewHTTPError(http.StatusConflict, "email already registered")
			}
			return echo.NewHTTPError(http.StatusBadGateway, "create user")
		}
	}

	if err := h.zit.GrantProjectRole(ctx, orgID, userID, string(domain.RoleAdmin)); err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, "grant role")
	}

	slog.Info("self-service signup provisioned",
		"org_id", orgID, "user_id", userID, "company", company, "dev_mode", h.dev)

	return c.JSON(http.StatusCreated, Response{
		OrgID: orgID, UserID: userID, Status: "pending_verification", VerifyURL: verifyURL,
	})
}

// randSuffix returns n random lowercase alphanumeric characters using crypto/rand
// so gosec is satisfied; the suffix is not a secret but collision avoidance only.
func randSuffix(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		b[i] = chars[idx.Int64()]
	}
	return string(b)
}
