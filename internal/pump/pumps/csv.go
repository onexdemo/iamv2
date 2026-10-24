package pumps

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mitchellh/mapstructure"

	"github.com/onexdemo/iamv2/internal/pump/analytics"
)

// CSVConf configures the CSV pump.
type CSVConf struct {
	// CSVDir is the directory in which hourly CSV files are written.
	CSVDir string `mapstructure:"csv_dir"`
}

// CSVPump writes records to hourly CSV files.
type CSVPump struct {
	csvConf *CSVConf
	CommonPumpConfig
}

// New returns a fresh CSV pump.
func (c *CSVPump) New() Pump {
	return &CSVPump{}
}

// GetName returns the pump name.
func (*CSVPump) GetName() string {
	return "CSV Pump"
}

// Init initializes the CSV pump.
func (c *CSVPump) Init(config any) error {
	c.csvConf = &CSVConf{}
	if err := mapstructure.Decode(config, c.csvConf); err != nil {
		return fmt.Errorf("decode csv configuration: %w", err)
	}

	if c.csvConf.CSVDir == "" {
		c.csvConf.CSVDir = "./_output/analytics"
	}

	if err := os.MkdirAll(c.csvConf.CSVDir, 0o755); err != nil {
		return fmt.Errorf("create csv directory: %w", err)
	}

	return nil
}

// WriteData appends a batch to the current hourly CSV file.
func (c *CSVPump) WriteData(_ context.Context, data []any) error {
	now := time.Now()
	fname := filepath.Join(
		c.csvConf.CSVDir,
		fmt.Sprintf("%d-%s-%d-%d.csv", now.Year(), now.Month().String(), now.Day(), now.Hour()),
	)

	appendHeader := false
	if _, err := os.Stat(fname); os.IsNotExist(err) {
		appendHeader = true
	}

	flags := os.O_APPEND | os.O_WRONLY
	if appendHeader {
		flags = os.O_CREATE | os.O_WRONLY
	}

	outfile, err := os.OpenFile(fname, flags, 0o600)
	if err != nil {
		return fmt.Errorf("open csv file: %w", err)
	}
	defer outfile.Close()

	writer := csv.NewWriter(outfile)
	if appendHeader {
		start := analytics.Record{}
		if err := writer.Write(start.GetFieldNames()); err != nil {
			return fmt.Errorf("write csv header: %w", err)
		}
	}

	for _, item := range data {
		record, ok := item.(analytics.Record)
		if !ok {
			continue
		}

		if err := writer.Write(record.GetLineValues()); err != nil {
			return fmt.Errorf("write csv row: %w", err)
		}
	}

	writer.Flush()

	return writer.Error()
}
