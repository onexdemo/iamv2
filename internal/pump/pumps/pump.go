// Package pumps contains the analytics back-ends supported by iam-pump.
package pumps

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/onexdemo/iamv2/internal/pump/analytics"
)

// Pump is the interface implemented by every analytics back-end.
type Pump interface {
	// New returns a fresh, uninitialized pump instance.
	New() Pump
	// GetName returns the human-readable pump name.
	GetName() string
	// Init configures the pump. The config is the pump's "meta" map.
	Init(config any) error
	// WriteData writes a batch of records.
	WriteData(ctx context.Context, data []any) error
	// SetFilters configures record filtering.
	SetFilters(filters analytics.Filters)
	// GetFilters returns the configured filters.
	GetFilters() analytics.Filters
	// SetTimeout configures the per-write timeout in seconds.
	SetTimeout(timeout int)
	// GetTimeout returns the per-write timeout in seconds.
	GetTimeout() int
	// SetOmitDetailedRecording controls whether policy details are stripped.
	SetOmitDetailedRecording(omit bool)
	// GetOmitDetailedRecording reports whether details are stripped.
	GetOmitDetailedRecording() bool
}

// CommonPumpConfig holds the options shared by every pump.
type CommonPumpConfig struct {
	filters               analytics.Filters
	timeout               int
	omitDetailedRecording bool
}

// SetFilters configures record filtering.
func (p *CommonPumpConfig) SetFilters(filters analytics.Filters) {
	p.filters = filters
}

// GetFilters returns the configured filters.
func (p *CommonPumpConfig) GetFilters() analytics.Filters {
	return p.filters
}

// SetTimeout configures the per-write timeout in seconds.
func (p *CommonPumpConfig) SetTimeout(timeout int) {
	p.timeout = timeout
}

// GetTimeout returns the per-write timeout in seconds.
func (p *CommonPumpConfig) GetTimeout() int {
	return p.timeout
}

// SetOmitDetailedRecording controls whether policy details are stripped.
func (p *CommonPumpConfig) SetOmitDetailedRecording(omit bool) {
	p.omitDetailedRecording = omit
}

// GetOmitDetailedRecording reports whether details are stripped.
func (p *CommonPumpConfig) GetOmitDetailedRecording() bool {
	return p.omitDetailedRecording
}

var availablePumps = map[string]Pump{}

// GetPumpByName returns a prototype for the named pump.
func GetPumpByName(name string) (Pump, error) {
	pump, ok := availablePumps[name]
	if !ok || pump == nil {
		return nil, fmt.Errorf("pump %q is not registered", name)
	}

	return pump, nil
}

func register(name string, pump Pump) {
	if _, ok := availablePumps[name]; ok {
		slog.Warn("duplicate pump registration ignored", "pump", name)

		return
	}

	availablePumps[name] = pump
}

func init() {
	register("dummy", &DummyPump{})
	register("csv", &CSVPump{})
	register("mongo", &MongoPump{})
	register("kafka", &KafkaPump{})
	register("elasticsearch", &ElasticsearchPump{})
	register("influx", &InfluxPump{})
	register("prometheus", &PrometheusPump{})
	register("syslog", &SyslogPump{})
}
