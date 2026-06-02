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
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"

	echoSwagger "github.com/swaggo/echo-swagger"

	"github.com/live-rack/pkg/audit"
	pkgauth "github.com/live-rack/pkg/auth"
	"github.com/live-rack/pkg/chstore"
	"github.com/live-rack/pkg/domain"
	"github.com/live-rack/pkg/events"
	"github.com/live-rack/pkg/integrations"
	obs "github.com/live-rack/pkg/observability"
	"github.com/live-rack/pkg/store"
	_ "github.com/live-rack/services/api/docs" // swaggo generated
	"github.com/live-rack/services/api/internal/analytics"
	"github.com/live-rack/services/api/internal/authadapter"
	"github.com/live-rack/services/api/internal/billing"
	"github.com/live-rack/services/api/internal/counts"
	integrationsapi "github.com/live-rack/services/api/internal/integrations"
	"github.com/live-rack/services/api/internal/inventory"
	"github.com/live-rack/services/api/internal/login"
	apimw "github.com/live-rack/services/api/internal/middleware"
	"github.com/live-rack/services/api/internal/onboarding"
	"github.com/live-rack/services/api/internal/passwordreset"
	"github.com/live-rack/services/api/internal/picking"
	"github.com/live-rack/services/api/internal/pipelines"
	"github.com/live-rack/services/api/internal/recommendations"
	"github.com/live-rack/services/api/internal/sales"
	"github.com/live-rack/services/api/internal/scans"
	"github.com/live-rack/services/api/internal/search"
	"github.com/live-rack/services/api/internal/servicetokens"
	"github.com/live-rack/services/api/internal/signup"
	"github.com/live-rack/services/api/internal/tasks"
	"github.com/live-rack/services/api/internal/users"
	"github.com/live-rack/services/api/internal/webhooks"
	"github.com/live-rack/services/api/internal/ws"
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

	natsURL := envOr("NATS_URL", "nats://localhost:4222")
	nc, err := nats.Connect(natsURL,
		nats.Name("api"),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
	)
	if err != nil {
		log.Error("connect nats", "err", err)
		os.Exit(1)
	}
	defer func() {
		if err := nc.Drain(); err != nil {
			log.Error("drain nats connection", "err", err)
		}
	}()

	js, err := nc.JetStream()
	if err != nil {
		log.Error("jetstream", "err", err)
		os.Exit(1)
	}
	if _, err := js.AddStream(&nats.StreamConfig{
		Name:      "LIVE_RACK",
		Subjects:  []string{"lr.>"},
		MaxAge:    24 * time.Hour,
		Storage:   nats.FileStorage,
		Retention: nats.LimitsPolicy,
	}); err != nil && !errors.Is(err, nats.ErrStreamNameAlreadyInUse) {
		log.Error("add stream", "err", err)
		os.Exit(1)
	}

	publisher := events.NewNATSPublisher(js)

	// setSession sets app.org_id + app.user_id on the acquired connection; RLS
	// policies read both to enforce tenant and zone scope.
	setSession := func(orgID, userID string) error {
		conn, err := pool.Acquire(context.Background())
		if err != nil {
			return fmt.Errorf("acquire conn: %w", err)
		}
		defer conn.Release()
		_, err = conn.Exec(context.Background(),
			fmt.Sprintf("SET LOCAL app.org_id = '%s'; SET LOCAL app.user_id = '%s'", orgID, userID))
		return err
	}

	q := store.New(pool)

	// Zitadel OIDC verifier — discovers JWKS at startup, JIT-provisions on first login.
	issuer := mustEnv("OIDC_ISSUER")
	projectID := mustEnv("OIDC_PROJECT_ID")
	adapter := authadapter.New(q)
	resolver := pkgauth.NewDBResolver(adapter)
	oidcVerifier, err := pkgauth.NewZitadelVerifier(ctx, issuer, projectID, resolver)
	if err != nil {
		log.Error("init oidc verifier", "err", err)
		os.Exit(1)
	}

	// Zitadel management client drives onboarding (signup + invites). The
	// service-account token authorises org/user creation and role grants.
	appBaseURL := envOr("APP_BASE_URL", "http://localhost:5173")
	mgmt := pkgauth.NewZitadelManagement(issuer, projectID, appBaseURL,
		pkgauth.StaticToken(os.Getenv("ZITADEL_MGMT_TOKEN")))
	auditWriter := audit.NewWriter(pool)
	// Login client drives the custom sign-in UI via Zitadel's Session API. Needs
	// an IAM_LOGIN_CLIENT token; falls back to the management token (IAM_OWNER
	// is a superset) when a dedicated one is not configured.
	loginTok := os.Getenv("ZITADEL_LOGIN_CLIENT_TOKEN")
	if loginTok == "" {
		loginTok = os.Getenv("ZITADEL_MGMT_TOKEN")
	}
	loginClient := pkgauth.NewZitadelLogin(issuer, pkgauth.StaticToken(loginTok))
	// Composite verifier: opaque service tokens ("lrk_...") resolve to service
	// principals; everything else goes through OIDC.
	verifier := pkgauth.NewCompositeVerifier(pkgauth.NewServiceVerifier(adapter), oidcVerifier)

	e := echo.New()
	e.HideBanner = true
	e.Use(echomw.Recover())
	e.Use(echomw.RequestID())
	e.Use(echomw.Logger())
	e.Use(echomw.CORS())
	e.Use(otelecho.Middleware("api"))
	// Per-IP throttle on the public auth endpoints (brute-force / enumeration).
	e.Use(apimw.AuthRateLimiter(
		envFloat("AUTH_RATE_LIMIT_RPS", 5), envInt("AUTH_RATE_LIMIT_BURST", 10)))

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

	// Inbound POS webhooks — unauthenticated, verified by per-vendor signature.
	webhooks.New(q, publisher, integrations.NewShopify(), integrations.NewSquare(), integrations.NewStripe()).Register(e)

	// Stripe billing webhook → org plan changes. Price→plan map from env (JSON).
	billing.New(q, envOr("STRIPE_BILLING_SECRET", ""), parsePricePlans(os.Getenv("STRIPE_PRICE_PLANS"))).Register(e)

	// Public self-service signup — provisions a tenant org + admin in Zitadel.
	signup.New(mgmt).Register(e)

	// Public custom-login proxy — drives Zitadel's Session API for our own sign-in UI.
	login.New(loginClient).Register(e)

	// Public invite acceptance — verify email + set password + enroll TOTP.
	onboarding.New(mgmt, loginClient, q, auditWriter).Register(e)

	// Public forgot/reset-password flow.
	passwordreset.New(mgmt, q, auditWriter).Register(e)

	// Authenticated API group.
	api := e.Group("/api/v1", apimw.Auth(verifier, setSession))

	zones.New(q).Register(api.Group("/stores"))
	scans.New(q, q, q, q, publisher).Register(api.Group("/stores"))
	inventory.New(q, auditWriter).Register(api.Group("/stores"))
	counts.New(q, auditWriter).Register(api.Group("/stores"))
	tasks.New(q, publisher).Register(api.Group("/stores"))
	pipelines.New(q, publisher).Register(api.Group("/stores"))
	picking.New(q, publisher, auditWriter).Register(api.Group("/stores"))
	sales.New(q).Register(api)
	integrationsapi.New(q).Register(api)
	search.New(q).Register(api)

	// ClickHouse-backed analytics reads.
	chCfg, err := chstore.ParseConfig(mustEnv("CLICKHOUSE_URL"), envOr("CLICKHOUSE_DB", "liverack"))
	if err != nil {
		log.Error("parse clickhouse url", "err", err)
		os.Exit(1)
	}
	analytics.New(chstore.New(chCfg)).Register(api)
	recommendations.New(q).Register(api)
	users.New(q).Register(api)
	users.NewMetrics(q, mgmt).Register(api)
	users.NewInvite(mgmt, q, auditWriter).Register(api)
	users.NewMFA(mgmt, q, auditWriter).Register(api)
	users.NewAccess(q, mgmt, auditWriter).Register(api)
	servicetokens.New(q).Register(api)

	hub := ws.NewHub(log)
	if _, err := nc.Subscribe("lr.*.>", func(m *nats.Msg) {
		hub.Broadcast(events.ExtractOrgID(m.Subject), m.Data)
	}); err != nil {
		log.Error("ws nats subscribe", "err", err)
		os.Exit(1)
	}
	ws.NewHandler(hub, verifier).Register(e)

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

// envFloat reads a float env var, falling back to def on missing/invalid input.
func envFloat(key string, def float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}

// envInt reads an int env var, falling back to def on missing/invalid input.
func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

// parsePricePlans decodes a JSON map of Stripe price id -> plan name into a
// domain plan map. Invalid/empty input yields an empty map (all unknown prices
// fall back to free).
func parsePricePlans(raw string) map[string]domain.Plan {
	out := map[string]domain.Plan{}
	if raw == "" {
		return out
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		slog.Error("parse STRIPE_PRICE_PLANS", "err", err)
		return out
	}
	for price, plan := range m {
		out[price] = domain.PlanFromString(plan)
	}
	return out
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
