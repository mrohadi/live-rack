package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	pkgauth "github.com/live-rack/pkg/auth"
	apimw "github.com/live-rack/services/api/internal/middleware"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slogLevel(os.Getenv("LOG_LEVEL")),
	}))
	slog.SetDefault(log)

	dbURL := mustEnv("DATABASE_URL")
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Error("connect postgres", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	// setOrgID executes SET LOCAL app.org_id = '<id>' on the acquired connection.
	// Called inside Auth middleware before every request hits a handler.
	setOrgID := func(orgID string) error {
		conn, err := pool.Acquire(context.Background())
		if err != nil {
			return fmt.Errorf("acquire conn: %w", err)
		}
		defer conn.Release()
		_, err = conn.Exec(context.Background(), fmt.Sprintf("SET LOCAL app.org_id = '%s'", orgID))
		return err
	}

	// TODO(LR-005): wire real DBResolver once sqlc is generated.
	// For now verifier is constructed; resolver stubbed until pkg/store is generated.
	_ = pkgauth.NewClerkVerifier(mustEnv("CLERK_SECRET_KEY"), nil)

	e := echo.New()
	e.HideBanner = true
	e.Use(echomw.Recover())
	e.Use(echomw.RequestID())
	e.Use(echomw.Logger())
	e.Use(echomw.CORS())

	// Health — no auth required.
	e.GET("/healthz", func(c echo.Context) error {
		if err := pool.Ping(context.Background()); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"db": "down"})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	// Clerk webhook — signed by Svix, no JWT auth.
	clerkWebhookSecret := os.Getenv("CLERK_WEBHOOK_SECRET")
	if clerkWebhookSecret != "" {
		whHandler, err := apimw.NewClerkWebhookHandler(clerkWebhookSecret, nil /* provisioner wired post-sqlc */)
		if err != nil {
			log.Error("init clerk webhook handler", "err", err)
			os.Exit(1)
		}
		e.POST("/webhooks/clerk", echo.WrapHandler(http.HandlerFunc(whHandler.ServeHTTP)))
	}

	// Authenticated API group — RLS applied to every handler via setOrgID.
	api := e.Group("/api/v1", apimw.Auth(
		pkgauth.NewClerkVerifier(mustEnv("CLERK_SECRET_KEY"), nil),
		setOrgID,
	))
	_ = api // routes added per phase

	port := envOr("PORT", "8080")
	srv := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Info("api listening", "port", port)
		if err := e.StartServer(srv); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_ = e.Shutdown(ctx)
	log.Info("server shutdown")
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		slog.Error("missing required env var", "key", key)
		os.Exit(1)
	}
	return v
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func slogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
