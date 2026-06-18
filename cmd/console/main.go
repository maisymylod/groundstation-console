// Command console runs the Heliosnet ops-console backend: an HTTP/JSON API over
// telemetry and incident state, optionally fed live by a Kafka/Redpanda consumer.
// With no broker configured it serves the bundled deterministic fixture, so the
// service is runnable from a clean clone with one command.
//
// A "produce" subcommand publishes the fixture to the broker for the compose demo.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/maisymylod/groundstation-console/internal/api"
	"github.com/maisymylod/groundstation-console/internal/broker"
	"github.com/maisymylod/groundstation-console/internal/config"
	"github.com/maisymylod/groundstation-console/internal/store"
	"github.com/maisymylod/groundstation-console/internal/telemetry"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg := config.Load()

	if len(os.Args) > 1 && os.Args[1] == "produce" {
		runProducer(cfg, log)
		return
	}

	runServer(cfg, log)
}

func runServer(cfg config.Config, log *slog.Logger) {
	st := store.New(cfg.CapPerSat)

	// Preload the bundled fixture so the console renders data immediately.
	for _, f := range telemetry.GenerateFixture(cfg.FixtureSeed, cfg.FixtureSize) {
		st.Ingest(f)
	}
	log.Info("fixture loaded", "satellites", len(telemetry.Fleet), "frames_per_sat", cfg.FixtureSize)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if len(cfg.BrokerSeeds) > 0 {
		c, err := broker.NewConsumer(cfg.BrokerSeeds, cfg.Topic, cfg.Group, st, log)
		if err != nil {
			log.Error("broker connect failed; continuing on fixture only", "err", err)
		} else {
			go func() {
				if err := c.Run(ctx); err != nil && ctx.Err() == nil {
					log.Error("consumer stopped", "err", err)
				}
			}()
			log.Info("live telemetry consumer started", "seeds", cfg.BrokerSeeds, "topic", cfg.Topic)
		}
	} else {
		log.Info("no BROKER_SEEDS set; serving bundled fixture only")
	}

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           api.New(st, log).Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		log.Info("console api listening", "addr", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	log.Info("shutting down")
	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutCtx)
}

func runProducer(cfg config.Config, log *slog.Logger) {
	if len(cfg.BrokerSeeds) == 0 {
		log.Error("produce requires BROKER_SEEDS")
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	log.Info("producing fixture to broker", "seeds", cfg.BrokerSeeds, "topic", cfg.Topic)
	if err := broker.Produce(ctx, cfg.BrokerSeeds, cfg.Topic, cfg.FixtureSeed, cfg.FixtureSize, 200*time.Millisecond); err != nil && ctx.Err() == nil {
		log.Error("producer error", "err", err)
		os.Exit(1)
	}
	log.Info("fixture produced")
}
