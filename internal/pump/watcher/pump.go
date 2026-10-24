// Package watcher implements the iam-pump cron plugin.
package watcher

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/onexstack/onexstack/pkg/watch/registry"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/onexdemo/iamv2/internal/pump/analytics"
	"github.com/onexdemo/iamv2/internal/pump/pumps"
	"github.com/onexdemo/iamv2/internal/pump/storage"
)

// Config is injected into the pump watcher.
type Config struct {
	Storage               *storage.RedisStorage
	Manager               *pumps.Manager
	PurgeDelay            int
	OmitDetailedRecording bool
}

// WantsConfig is implemented by watchers that need the pump dependencies.
type WantsConfig interface {
	registry.Watcher
	SetPumpConfig(config *Config)
}

type pumpWatcher struct {
	config *Config
}

var (
	_ registry.Watcher = (*pumpWatcher)(nil)
	_ registry.ISpec   = (*pumpWatcher)(nil)
	_ WantsConfig      = (*pumpWatcher)(nil)
)

// SetPumpConfig injects the shared pump dependencies.
func (w *pumpWatcher) SetPumpConfig(config *Config) {
	w.config = config
}

// Run drains the analytics list and writes the records to every pump.
func (w *pumpWatcher) Run() {
	if w.config == nil {
		slog.Warn("pump watcher is not initialized")

		return
	}

	ctx := context.Background()

	values, err := w.config.Storage.GetAndDeleteSet(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to read analytics batch", "error", err)

		return
	}

	if len(values) == 0 {
		return
	}

	records := make([]analytics.Record, 0, len(values))
	for _, value := range values {
		var record analytics.Record

		if err := msgpack.Unmarshal([]byte(value), &record); err != nil {
			slog.WarnContext(ctx, "failed to decode analytics record", "error", err)

			continue
		}

		if w.config.OmitDetailedRecording {
			record.Policies = ""
			record.Deciders = ""
		}

		records = append(records, record)
	}

	if err := w.config.Manager.Pump(ctx, records, w.config.PurgeDelay); err != nil {
		slog.ErrorContext(ctx, "failed to write analytics batch", "error", err)
	}
}

// Spec returns the cron schedule derived from the purge delay.
func (w *pumpWatcher) Spec() string {
	delay := 10
	if w.config != nil && w.config.PurgeDelay > 0 {
		delay = w.config.PurgeDelay
	}

	return fmt.Sprintf("@every %ds", delay)
}

func init() {
	registry.Register("pump", &pumpWatcher{})
}
