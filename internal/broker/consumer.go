// Package broker connects the console to a Kafka/Redpanda telemetry topic for
// live frame push. The consumer decodes JSON telemetry frames and ingests them
// into the store, deriving incidents on the way in. When no seed brokers are
// configured the console runs entirely off the bundled fixture and this package
// is not started.
package broker

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/maisymylod/groundstation-console/internal/store"
	"github.com/maisymylod/groundstation-console/internal/telemetry"
	"github.com/twmb/franz-go/pkg/kgo"
)

// Consumer drains telemetry frames from the broker into the store.
type Consumer struct {
	client *kgo.Client
	store  *store.Store
	log    *slog.Logger
}

// NewConsumer dials the seed brokers and joins the consumer group on the topic.
func NewConsumer(seeds []string, topic, group string, st *store.Store, log *slog.Logger) (*Consumer, error) {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(seeds...),
		kgo.ConsumerGroup(group),
		kgo.ConsumeTopics(topic),
		kgo.DisableAutoCommit(),
	)
	if err != nil {
		return nil, err
	}
	return &Consumer{client: cl, store: st, log: log}, nil
}

// Run consumes until the context is cancelled. Frames that fail to decode are
// logged and skipped rather than stalling the partition.
func (c *Consumer) Run(ctx context.Context) error {
	defer c.client.Close()
	for {
		fetches := c.client.PollFetches(ctx)
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if errs := fetches.Errors(); len(errs) > 0 {
			for _, e := range errs {
				c.log.Error("fetch error", "topic", e.Topic, "partition", e.Partition, "err", e.Err)
			}
			continue
		}
		fetches.EachRecord(func(rec *kgo.Record) {
			var f telemetry.Frame
			if err := json.Unmarshal(rec.Value, &f); err != nil {
				c.log.Warn("skipping malformed frame", "err", err)
				return
			}
			if inc, ok := c.store.Ingest(f); ok {
				c.log.Info("incident", "id", inc.ID, "sat", inc.SatID, "type", inc.Type, "severity", inc.Severity.String())
			}
		})
		if err := c.client.CommitUncommittedOffsets(ctx); err != nil {
			c.log.Error("commit failed", "err", err)
		}
	}
}
