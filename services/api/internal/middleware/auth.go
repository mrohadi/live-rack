// Package middleware provides Echo middleware for the live-rack API.
package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"

	pkgauth "github.com/live-rack/pkg/auth"
)

// Auth verifies the OIDC JWT, sets Principal + app.org_id on DB connection.
func Auth(verifier pkgauth.Verifier, setOrgID func(orgID string) error) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			principal, err := verifier.VerifyRequest(c.Request())
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
			}

			// Set Postgres session parameter — RLS reads this per-query.
			if err := setOrgID(principal.OrgID.String()); err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to set tenant context")
			}

			ctx := pkgauth.WithPrincipal(c.Request().Context(), principal)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

// RequireRole rejects requests whose principal lacks one of the given roles.
func RequireRole(roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			p, err := pkgauth.PrincipalFrom(c.Request().Context())
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
			}
			for _, r := range roles {
				if string(p.Role) == r {
					return next(c)
				}
			}
			return echo.NewHTTPError(http.StatusForbidden, "insufficient role")
		}
	}
}
