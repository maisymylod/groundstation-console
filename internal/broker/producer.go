package broker

import (
	"context"
	"encoding/json"
	"time"

	"github.com/maisymylod/groundstation-console/internal/telemetry"
	"github.com/twmb/franz-go/pkg/kgo"
)

// Produce publishes the bundled fixture to the topic at a steady cadence so the
// docker-compose demo shows live frames arriving. It blocks until ctx is done or
// the fixture is exhausted.
func Produce(ctx context.Context, seeds []string, topic string, seed int64, perSat int, interval time.Duration) error {
	cl, err := kgo.NewClient(kgo.SeedBrokers(seeds...), kgo.AllowAutoTopicCreation())
	if err != nil {
		return err
	}
	defer cl.Close()

	frames := telemetry.GenerateFixture(seed, perSat)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for _, f := range frames {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
		val, err := json.Marshal(f)
		if err != nil {
			continue
		}
		cl.Produce(ctx, &kgo.Record{Topic: topic, Key: []byte(f.SatID), Value: val}, nil)
	}
	return cl.Flush(ctx)
}
