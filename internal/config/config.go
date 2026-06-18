// Package config loads console settings from the environment. Defaults make the
// service runnable from a clean clone against the bundled fixture, with no broker
// required; setting BROKER_SEEDS switches on the live telemetry consumer.
package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds runtime settings.
type Config struct {
	Addr        string   // HTTP listen address
	BrokerSeeds []string // Kafka/Redpanda seed brokers; empty disables live consume
	Topic       string   // telemetry topic
	Group       string   // consumer group
	CapPerSat   int      // retained frames per satellite (0 = unbounded)
	FixtureSeed int64    // seed for the bundled fixture
	FixtureSize int      // frames per satellite to preload from the fixture
}

// Load reads configuration from the environment, applying defaults.
func Load() Config {
	return Config{
		Addr:        env("CONSOLE_ADDR", ":8080"),
		BrokerSeeds: splitNonEmpty(env("BROKER_SEEDS", "")),
		Topic:       env("TELEMETRY_TOPIC", "heliosnet.telemetry"),
		Group:       env("CONSUMER_GROUP", "groundstation-console"),
		CapPerSat:   envInt("CAP_PER_SAT", 5000),
		FixtureSeed: int64(envInt("FIXTURE_SEED", 42)),
		FixtureSize: envInt("FIXTURE_SIZE", 240),
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func splitNonEmpty(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
