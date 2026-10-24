package pumps

import (
	"context"
	"log/slog"
)

// DummyPump discards every record. It is useful for development and for
// validating the pump pipeline without a back-end.
type DummyPump struct {
	CommonPumpConfig
}

// New returns a fresh dummy pump.
func (p *DummyPump) New() Pump {
	return &DummyPump{}
}

// GetName returns the pump name.
func (*DummyPump) GetName() string {
	return "Dummy Pump"
}

// Init initializes the dummy pump.
func (*DummyPump) Init(any) error {
	return nil
}

// WriteData logs the batch size and discards the records.
func (*DummyPump) WriteData(_ context.Context, data []any) error {
	slog.Info("dummy pump discarded records", "count", len(data))

	return nil
}
