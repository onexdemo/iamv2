//go:build !windows

package pumps

import (
	"context"
	"fmt"
	"log/syslog"

	"github.com/mitchellh/mapstructure"

	"github.com/onexdemo/iamv2/internal/pump/analytics"
)

// SyslogConf configures the syslog pump.
type SyslogConf struct {
	Transport   string `mapstructure:"transport"`
	NetworkAddr string `mapstructure:"network_addr"`
	LogLevel    int    `mapstructure:"log_level"`
	Tag         string `mapstructure:"tag"`
}

// SyslogPump writes records to a syslog daemon.
type SyslogPump struct {
	conf   *SyslogConf
	writer *syslog.Writer
	CommonPumpConfig
}

// New returns a fresh syslog pump.
func (s *SyslogPump) New() Pump {
	return &SyslogPump{}
}

// GetName returns the pump name.
func (*SyslogPump) GetName() string {
	return "Syslog Pump"
}

// Init connects to the syslog daemon.
func (s *SyslogPump) Init(config any) error {
	s.conf = &SyslogConf{}
	if err := mapstructure.Decode(config, s.conf); err != nil {
		return fmt.Errorf("decode syslog configuration: %w", err)
	}

	if s.conf.Transport == "" {
		s.conf.Transport = "udp"
	}

	if s.conf.Transport != "udp" && s.conf.Transport != "tcp" && s.conf.Transport != "tls" {
		return fmt.Errorf("unsupported syslog transport %q", s.conf.Transport)
	}

	if s.conf.NetworkAddr == "" {
		s.conf.NetworkAddr = "localhost:5140"
	}

	tag := s.conf.Tag
	if tag == "" {
		tag = "iam-pump"
	}

	writer, err := syslog.Dial(
		s.conf.Transport,
		s.conf.NetworkAddr,
		syslog.Priority(s.conf.LogLevel),
		tag,
	)
	if err != nil {
		return fmt.Errorf("connect to syslog daemon: %w", err)
	}

	s.writer = writer

	return nil
}

// WriteData writes each record to syslog.
func (s *SyslogPump) WriteData(ctx context.Context, data []any) error {
	if s.writer == nil {
		return fmt.Errorf("syslog pump is not initialized")
	}

	for _, item := range data {
		if err := ctx.Err(); err != nil {
			return err
		}

		record, ok := item.(analytics.Record)
		if !ok {
			continue
		}

		message := fmt.Sprintf(
			"timestamp=%d username=%s effect=%s conclusion=%s request=%s",
			record.TimeStamp,
			record.Username,
			record.Effect,
			record.Conclusion,
			record.Request,
		)
		if _, err := fmt.Fprintf(s.writer, "%s", message); err != nil {
			return fmt.Errorf("write syslog message: %w", err)
		}
	}

	return nil
}
