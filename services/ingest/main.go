// Package main is the ingest worker: it consumes pos.sale events from NATS and
// writes them into the sales_events Timescale hypertable.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"

	"github.com/live-rack/pkg/chstore"
	"github.com/live-rack/pkg/store"
	"github.com/live-rack/services/ingest/internal/chsink"
	"github.com/live-rack/services/ingest/internal/consumer"
)

// pgRecorder writes a sale inside its org's RLS context via a scoped transaction.
type pgRecorder struct {
	pool *pgxpool.Pool
}

func (r *pgRecorder) RecordSale(ctx context.Context, orgID uuid.UUID, arg store.CreateSaleEventParams) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, "SELECT set_config('app.org_id', $1, true)", orgID.String()); err != nil {
		return fmt.Errorf("set org_id: %w", err)
	}
	if _, err := store.New(tx).CreateSaleEvent(ctx, arg); err != nil {
		return fmt.Errorf("insert sale: %w", err)
	}
	return tx.Commit(ctx)
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

	natsURL := envOr("NATS_URL", "nats://localhost:4222")
	nc, err := nats.Connect(natsURL, nats.Name("ingest"), nats.MaxReconnects(-1), nats.ReconnectWait(2*time.Second))
	if err != nil {
		log.Error("connect nats", "err", err)
		os.Exit(1)
	}
	defer func() { _ = nc.Drain() }()

	cons := consumer.New(&pgRecorder{pool: pool})

	// ClickHouse analytics sink: mirror scan + sale events into the raw tables
	// the rollups and materialized views aggregate from.
	chCfg, err := chstore.ParseConfig(mustEnv("CLICKHOUSE_URL"), envOr("CLICKHOUSE_DB", "liverack"))
	if err != nil {
		log.Error("parse clickhouse url", "err", err)
		os.Exit(1)
	}
	ch := chstore.New(chCfg)
	if err := ch.Migrate(ctx); err != nil {
		log.Error("clickhouse migrate", "err", err)
		os.Exit(1)
	}
	sink := chsink.New(ch)

	saleSub, err := nc.QueueSubscribe("lr.*.pos.sale", "ingest", func(m *nats.Msg) {
		if err := cons.Handle(context.Background(), m.Data); err != nil {
			log.Error("handle pos.sale", "err", err, "subject", m.Subject)
		}
		if err := sink.HandleSale(context.Background(), m.Data); err != nil {
			log.Error("sink pos.sale", "err", err, "subject", m.Subject)
		}
	})
	if err != nil {
		log.Error("subscribe pos.sale", "err", err)
		os.Exit(1)
	}
	defer func() { _ = saleSub.Unsubscribe() }()

	scanSub, err := nc.QueueSubscribe("lr.*.scan.recorded", "ingest", func(m *nats.Msg) {
		if err := sink.HandleScan(context.Background(), m.Data); err != nil {
			log.Error("sink scan.recorded", "err", err, "subject", m.Subject)
		}
	})
	if err != nil {
		log.Error("subscribe scan.recorded", "err", err)
		os.Exit(1)
	}
	defer func() { _ = scanSub.Unsubscribe() }()

	log.Info("ingest worker listening", "subjects", "lr.*.pos.sale, lr.*.scan.recorded")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("ingest worker shutdown")
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
