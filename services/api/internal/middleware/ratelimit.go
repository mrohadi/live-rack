package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

// authPrefixes are the public, pre-authentication endpoints worth throttling
// per IP: sign-in, invite acceptance, password reset, signup. Brute-force and
// user-enumeration attempts target these.
var authPrefixes = []string{
	"/api/v1/login",
	"/api/v1/onboard",
	"/api/v1/password",
	"/api/v1/signup",
}

// isAuthPath reports whether a request path is an auth-sensitive endpoint. Pure.
func isAuthPath(path string) bool {
	for _, p := range authPrefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

// AuthRateLimiter throttles the public auth endpoints per client IP. Non-auth
// paths are skipped, so the org-level gateway limit still governs the rest of
// the API. rps is the sustained per-IP rate; burst is the short-term allowance.
func AuthRateLimiter(rps float64, burst int) echo.MiddlewareFunc {
	store := middleware.NewRateLimiterMemoryStoreWithConfig(
		middleware.RateLimiterMemoryStoreConfig{
			Rate:      rate.Limit(rps),
			Burst:     burst,
			ExpiresIn: 10 * time.Minute,
		},
	)
	return middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Skipper: func(c echo.Context) bool {
			return !isAuthPath(c.Request().URL.Path)
		},
		Store: store,
		IdentifierExtractor: func(c echo.Context) (string, error) {
			return c.RealIP(), nil
		},
		DenyHandler: func(c echo.Context, _ string, _ error) error {
			return echo.NewHTTPError(http.StatusTooManyRequests, "too many requests")
		},
	})
}
