//go:build windows

package pumps

import (
	"context"
	"errors"
)

// SyslogPump is unavailable on Windows. The type exists so that the pump
// registry remains compilable on every platform.
type SyslogPump struct {
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

// Init always fails on Windows.
func (*SyslogPump) Init(any) error {
	return errors.New("the syslog pump is not supported on windows")
}

// WriteData always fails on Windows.
func (*SyslogPump) WriteData(context.Context, []any) error {
	return errors.New("the syslog pump is not supported on windows")
}
