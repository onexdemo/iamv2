package pumps

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/onexdemo/iamv2/internal/pump/analytics"
)

// Config describes a single configured pump.
type Config struct {
	// Type is the registered pump name. It defaults to the map key.
	Type string `json:"type" mapstructure:"type"`
	// Meta is the pump-specific configuration.
	Meta map[string]any `json:"meta" mapstructure:"meta"`
	// Filters selects which records the pump receives.
	Filters analytics.Filters `json:"filters" mapstructure:"filters"`
	// Timeout is the per-write timeout in seconds. Zero means no timeout.
	Timeout int `json:"timeout" mapstructure:"timeout"`
	// OmitDetailedRecording strips request policy details before writing.
	OmitDetailedRecording bool `json:"omit_detailed_recording" mapstructure:"omit_detailed_recording"`
}

// Manager owns every initialized pump.
type Manager struct {
	pumps []Pump
}

// NewManager initializes every configured pump.
func NewManager(configs map[string]Config) (*Manager, error) {
	manager := &Manager{pumps: make([]Pump, 0, len(configs))}

	errs := []error{}
	for key, cfg := range configs {
		typeName := cfg.Type
		if typeName == "" {
			typeName = key
		}

		prototype, err := GetPumpByName(typeName)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		pump := prototype.New()
		if err := pump.Init(cfg.Meta); err != nil {
			errs = append(errs, fmt.Errorf("init pump: %w", err))
			continue
		}

		pump.SetFilters(cfg.Filters)
		pump.SetTimeout(cfg.Timeout)
		pump.SetOmitDetailedRecording(cfg.OmitDetailedRecording)

		slog.Info("initialized pump", "name", pump.GetName())

		manager.pumps = append(manager.pumps, pump)
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return manager, nil
}

// Pump writes the records to every configured pump.
func (m *Manager) Pump(ctx context.Context, records []analytics.Record, purgeDelay int) error {
	if len(m.pumps) == 0 {
		slog.WarnContext(ctx, "no pumps configured")

		return nil
	}

	var (
		wg   sync.WaitGroup
		lock sync.Mutex
		errs []error
	)

	for _, pump := range m.pumps {
		wg.Add(1)

		go func(pump Pump) {
			defer wg.Done()

			writeCtx := ctx
			if timeout := pump.GetTimeout(); timeout > 0 {
				var cancel context.CancelFunc

				writeCtx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
				defer cancel()
			}

			if err := pump.WriteData(writeCtx, filterData(pump, records)); err != nil {
				lock.Lock()

				errs = append(errs, fmt.Errorf("pump write failed: %w", err))
				lock.Unlock()

				return
			}

			if purgeDelay > 0 && pump.GetTimeout() > purgeDelay {
				slog.WarnContext(
					ctx,
					"pump timeout exceeds purge delay",
					"pump", pump.GetName(),
					"timeout", pump.GetTimeout(),
					"purge_delay", purgeDelay,
				)
			}
		}(pump)
	}

	wg.Wait()

	return errors.Join(errs...)
}

func filterData(pump Pump, records []analytics.Record) []any {
	filters := pump.GetFilters()
	omit := pump.GetOmitDetailedRecording()

	if !filters.HasFilter() && !omit {
		out := make([]any, 0, len(records))
		for _, record := range records {
			out = append(out, record)
		}

		return out
	}

	out := make([]any, 0, len(records))
	for _, record := range records {
		if omit {
			record.Policies = ""
			record.Deciders = ""
		}

		if filters.ShouldFilter(record) {
			continue
		}

		out = append(out, record)
	}

	return out
}
