package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsAuthPath(t *testing.T) {
	assert.True(t, isAuthPath("/api/v1/login/start"))
	assert.True(t, isAuthPath("/api/v1/onboard/complete"))
	assert.True(t, isAuthPath("/api/v1/password/forgot"))
	assert.True(t, isAuthPath("/api/v1/signup"))
	assert.False(t, isAuthPath("/api/v1/users"))
	assert.False(t, isAuthPath("/healthz"))
}

func send(e *echo.Echo, path string) int {
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, path, nil)
	req.RemoteAddr = "203.0.113.7:1234"
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec.Code
}

func TestAuthRateLimiter_ThrottlesAuthPaths(t *testing.T) {
	e := echo.New()
	e.Use(AuthRateLimiter(1, 1)) // 1 rps, burst 1
	e.POST("/api/v1/login/start", func(c echo.Context) error { return c.NoContent(http.StatusOK) })

	require.Equal(t, http.StatusOK, send(e, "/api/v1/login/start"))
	assert.Equal(t, http.StatusTooManyRequests, send(e, "/api/v1/login/start"))
}

func TestAuthRateLimiter_SkipsNonAuthPaths(t *testing.T) {
	e := echo.New()
	e.Use(AuthRateLimiter(1, 1))
	e.GET("/api/v1/users", func(c echo.Context) error { return c.NoContent(http.StatusOK) })

	for i := 0; i < 5; i++ {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/users", nil)
		req.RemoteAddr = "203.0.113.7:1234"
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
	}
}
