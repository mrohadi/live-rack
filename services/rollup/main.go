// Package main is the rollup worker: it runs daily ClickHouse aggregation jobs
// (time-to-sell, sell-through, co-purchase lift) on a schedule. The 5-minute
// and hourly rollups are maintained by materialized views on insert, so they
// need no job here.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/live-rack/pkg/chstore"
	"github.com/live-rack/services/rollup/internal/jobs"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(log)
	ctx := context.Background()

	cfg, err := chstore.ParseConfig(mustEnv("CLICKHOUSE_URL"), envOr("CLICKHOUSE_DB", "liverack"))
	if err != nil {
		log.Error("parse clickhouse url", "err", err)
		os.Exit(1)
	}
	ch := chstore.New(cfg)
	if err := ch.Migrate(ctx); err != nil {
		log.Error("clickhouse migrate", "err", err)
		os.Exit(1)
	}

	// Jobs are registered by their owning tickets (LR-706, LR-707).
	runner := jobs.NewRunner(ch)

	interval := envDuration("ROLLUP_INTERVAL", 24*time.Hour)
	runOnce := func() {
		day := time.Now().UTC()
		if err := runner.Run(ctx, day); err != nil {
			log.Error("rollup run", "err", err, "day", jobs.DayString(day))
			return
		}
		log.Info("rollup run complete", "day", jobs.DayString(day))
	}

	runOnce() // catch up on boot
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Info("rollup worker started", "interval", interval.String())

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-ticker.C:
			runOnce()
		case <-quit:
			log.Info("rollup worker shutdown")
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
