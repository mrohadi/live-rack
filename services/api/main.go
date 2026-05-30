// Package main is the entry point for the live-rack API service.
//
//	@title			live-rack API
//	@version		0.1.0
//	@description	Warehouse zoning, scanning, and analytics API.
//	@host			localhost:8080
//	@BasePath		/api/v1
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
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
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"

	echoSwagger "github.com/swaggo/echo-swagger"

	pkgauth "github.com/live-rack/pkg/auth"
	obs "github.com/live-rack/pkg/observability"
	"github.com/live-rack/pkg/store"
	_ "github.com/live-rack/services/api/docs" // swaggo generated
	"github.com/live-rack/services/api/internal/authadapter"
	apimw "github.com/live-rack/services/api/internal/middleware"
	"github.com/live-rack/services/api/internal/zones"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slogLevel(os.Getenv("LOG_LEVEL")),
	}))
	slog.SetDefault(log)

	ctx := context.Background()

	otel := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	log.Info("starting api service", "otel_endpoint", otel)

	shutdown, err := obs.Setup(ctx, obs.Config{
		ServiceName:    "api",
		ServiceVersion: envOr("SERVICE_VERSION", "dev"),
		OTLPEndpoint:   os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
	})
	if err != nil {
		log.Error("otel setup", "err", err)
		os.Exit(1)
	}
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			log.Error("otel shutdown", "err", err)
		}
	}()

	dbURL := mustEnv("DATABASE_URL")
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Error("connect postgres", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	// setOrgID executes SET LOCAL app.org_id = '<id>' on the acquired connection.
	setOrgID := func(orgID string) error {
		conn, err := pool.Acquire(context.Background())
		if err != nil {
			return fmt.Errorf("acquire conn: %w", err)
		}
		defer conn.Release()
		_, err = conn.Exec(context.Background(), fmt.Sprintf("SET LOCAL app.org_id = '%s'", orgID))
		return err
	}

	// _ = pkgauth.NewClerkVerifier(mustEnv("CLERK_SECRET_KEY"), nil)

	e := echo.New()
	e.HideBanner = true
	e.Use(echomw.Recover())
	e.Use(echomw.RequestID())
	e.Use(echomw.Logger())
	e.Use(echomw.CORS())
	e.Use(otelecho.Middleware("api"))

	// Swagger UI — no auth, dev/staging only.
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Health — no auth.
	e.GET("/healthz", func(c echo.Context) error {
		if err := pool.Ping(c.Request().Context()); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"db": "down"})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	// OpenMetrics endpoint (scraped by Elastic Metricbeat) — no auth.
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	// Clerk webhook — signed by Svix, no JWT auth.
	clerkWebhookSecret := os.Getenv("CLERK_WEBHOOK_SECRET")
	if clerkWebhookSecret != "" {
		whHandler, err := apimw.NewClerkWebhookHandler(clerkWebhookSecret, nil)
		if err != nil {
			log.Error("init clerk webhook handler", "err", err)
			os.Exit(1)
		}
		e.POST("/webhooks/clerk", echo.WrapHandler(http.HandlerFunc(whHandler.ServeHTTP)))
	}

	q := store.New(pool)

	// Authenticated API group.
	api := e.Group("/api/v1", apimw.Auth(
		pkgauth.NewClerkVerifier(mustEnv("CLERK_SECRET_KEY"), pkgauth.NewDBResolver(authadapter.New(q))),
		setOrgID,
	))

	zones.New(q).Register(api.Group("/stores"))

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
