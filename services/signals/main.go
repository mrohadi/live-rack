// Package main is the signals worker: it polls external providers (weather,
// transit) per store on a schedule and publishes canonical signal events for
// the insight engine to consume.
package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"

	"github.com/live-rack/pkg/events"
	"github.com/live-rack/services/signals/internal/poller"
)

// pgStoreLister reads store coordinates for geo-scoped polling.
type pgStoreLister struct {
	pool *pgxpool.Pool
}

func (l *pgStoreLister) ListStoreGeo(ctx context.Context) ([]poller.StoreGeo, error) {
	rows, err := l.pool.Query(ctx,
		"SELECT org_id, id, lat, lon FROM stores WHERE lat IS NOT NULL AND lon IS NOT NULL")
	if err != nil {
		return nil, fmt.Errorf("query stores: %w", err)
	}
	defer rows.Close()

	var out []poller.StoreGeo
	for rows.Next() {
		var s poller.StoreGeo
		if err := rows.Scan(&s.OrgID, &s.StoreID, &s.Lat, &s.Lon); err != nil {
			return nil, fmt.Errorf("scan store: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// httpFetcher performs outbound GETs with a timeout.
type httpFetcher struct {
	client *http.Client
}

func (f *httpFetcher) Get(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	return body, nil
}

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(log)
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, mustEnv("DATABASE_URL"))
	if err != nil {
		log.Error("connect postgres", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	nc, err := nats.Connect(envOr("NATS_URL", "nats://localhost:4222"),
		nats.Name("signals"), nats.MaxReconnects(-1), nats.ReconnectWait(2*time.Second))
	if err != nil {
		log.Error("connect nats", "err", err)
		os.Exit(1)
	}
	defer func() { _ = nc.Drain() }()
	js, err := nc.JetStream()
	if err != nil {
		log.Error("jetstream", "err", err)
		os.Exit(1)
	}

	fetch := &httpFetcher{client: &http.Client{Timeout: 10 * time.Second}}
	lister := &pgStoreLister{pool: pool}
	pub := events.NewNATSPublisher(js)
	weather := poller.NewWeatherPoller(lister, fetch, pub, mustEnv("OPENWEATHER_API_KEY"))
	transit := poller.NewTransitPoller(lister, fetch, pub, envOr("TRANSIT_API_KEY", ""))

	interval := envDuration("SIGNALS_INTERVAL", 15*time.Minute)
	runOnce := func() {
		if err := weather.Poll(ctx); err != nil {
			log.Error("weather poll", "err", err)
		}
		if err := transit.Poll(ctx); err != nil {
			log.Error("transit poll", "err", err)
		}
	}
	runOnce()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	log.Info("signals worker started", "interval", interval.String())

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-ticker.C:
			runOnce()
		case <-quit:
			log.Info("signals worker shutdown")
			return
		}
	}
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

func envDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
