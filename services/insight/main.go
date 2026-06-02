// Package main is the insight worker: it consumes external signal events and
// publishes operator recommendations via the rule-based insight engine.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/live-rack/pkg/events"
	"github.com/live-rack/services/insight/internal/engine"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(log)

	nc, err := nats.Connect(envOr("NATS_URL", "nats://localhost:4222"),
		nats.Name("insight"), nats.MaxReconnects(-1), nats.ReconnectWait(2*time.Second))
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

	eng := engine.New(events.NewNATSPublisher(js), nil, nil)

	weatherSub, err := nc.QueueSubscribe("lr.*.signal.weather", "insight", func(m *nats.Msg) {
		if err := eng.HandleWeather(context.Background(), m.Data); err != nil {
			log.Error("handle weather signal", "err", err, "subject", m.Subject)
		}
	})
	if err != nil {
		log.Error("subscribe signal.weather", "err", err)
		os.Exit(1)
	}
	defer func() { _ = weatherSub.Unsubscribe() }()

	transitSub, err := nc.QueueSubscribe("lr.*.signal.transit", "insight", func(m *nats.Msg) {
		if err := eng.HandleTransit(context.Background(), m.Data); err != nil {
			log.Error("handle transit signal", "err", err, "subject", m.Subject)
		}
	})
	if err != nil {
		log.Error("subscribe signal.transit", "err", err)
		os.Exit(1)
	}
	defer func() { _ = transitSub.Unsubscribe() }()

	log.Info("insight worker listening", "subjects", "lr.*.signal.weather, lr.*.signal.transit")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("insight worker shutdown")
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
