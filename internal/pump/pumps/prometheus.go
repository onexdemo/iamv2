package pumps

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/ory/ladon"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/onexdemo/iamv2/internal/pump/analytics"
)

// PrometheusConf configures the Prometheus pump.
type PrometheusConf struct {
	Addr string `mapstructure:"listen_address"`
	Path string `mapstructure:"path"`
}

// PrometheusPump exposes authorization counters over HTTP.
type PrometheusPump struct {
	conf           *PrometheusConf
	registry       *prometheus.Registry
	totalStatusVec *prometheus.CounterVec
	server         *http.Server
	CommonPumpConfig
}

// New returns a fresh Prometheus pump.
func (p *PrometheusPump) New() Pump {
	pump := &PrometheusPump{
		registry: prometheus.NewRegistry(),
		totalStatusVec: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "iam_user_authorization_status_total",
				Help: "Authorization effect per user.",
			},
			[]string{"code", "username"},
		),
	}
	pump.registry.MustRegister(pump.totalStatusVec)

	return pump
}

// GetName returns the pump name.
func (*PrometheusPump) GetName() string {
	return "Prometheus Pump"
}

// Init starts the metrics listener.
func (p *PrometheusPump) Init(config any) error {
	p.conf = &PrometheusConf{}
	if err := mapstructure.Decode(config, p.conf); err != nil {
		return fmt.Errorf("decode prometheus configuration: %w", err)
	}

	if p.conf.Addr == "" {
		return errors.New("prometheus listen_address is required")
	}

	if p.conf.Path == "" {
		p.conf.Path = "/metrics"
	}

	mux := http.NewServeMux()
	mux.Handle(p.conf.Path, promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{}))

	p.server = &http.Server{Addr: p.conf.Addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}

	go func() {
		slog.Info("starting prometheus pump listener", "addr", p.conf.Addr, "path", p.conf.Path)

		if err := p.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("prometheus pump listener failed", "error", err)
		}
	}()

	return nil
}

// WriteData updates the authorization counters.
func (p *PrometheusPump) WriteData(_ context.Context, data []any) error {
	for _, item := range data {
		record, ok := item.(analytics.Record)
		if !ok {
			continue
		}

		code := "0"
		if record.Effect != ladon.AllowAccess {
			code = "1"
		}

		p.totalStatusVec.WithLabelValues(code, record.Username).Inc()
	}

	return nil
}
