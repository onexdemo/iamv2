package pumps

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/mitchellh/mapstructure"

	"github.com/onexdemo/iamv2/internal/pump/analytics"
)

const influxTable = "analytics"

// InfluxConf configures the InfluxDB pump.
type InfluxConf struct {
	DatabaseName string   `mapstructure:"database_name"`
	Addr         string   `mapstructure:"address"`
	Username     string   `mapstructure:"username"`
	Password     string   `mapstructure:"password"`
	Fields       []string `mapstructure:"fields"`
	Tags         []string `mapstructure:"tags"`
}

// InfluxPump writes records to InfluxDB.
type InfluxPump struct {
	conf *InfluxConf
	CommonPumpConfig
}

// New returns a fresh InfluxDB pump.
func (i *InfluxPump) New() Pump {
	return &InfluxPump{}
}

// GetName returns the pump name.
func (*InfluxPump) GetName() string {
	return "InfluxDB Pump"
}

// Init validates the InfluxDB configuration.
func (i *InfluxPump) Init(config any) error {
	i.conf = &InfluxConf{}
	if err := mapstructure.Decode(config, i.conf); err != nil {
		return fmt.Errorf("decode influx configuration: %w", err)
	}

	if i.conf.Addr == "" {
		return fmt.Errorf("influx address is required")
	}

	if i.conf.DatabaseName == "" {
		return fmt.Errorf("influx database_name is required")
	}

	slog.Info("initialized influx pump", "address", i.conf.Addr, "database", i.conf.DatabaseName)

	return nil
}

// WriteData writes a batch of points.
func (i *InfluxPump) WriteData(ctx context.Context, data []any) error {
	httpClient, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     i.conf.Addr,
		Username: i.conf.Username,
		Password: i.conf.Password,
	})
	if err != nil {
		return fmt.Errorf("create influx client: %w", err)
	}
	defer httpClient.Close()

	batchPoints, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  i.conf.DatabaseName,
		Precision: "us",
	})
	if err != nil {
		return fmt.Errorf("create influx batch: %w", err)
	}

	for _, item := range data {
		record, ok := item.(analytics.Record)
		if !ok {
			continue
		}

		mapping := map[string]any{
			"timestamp":  record.TimeStamp,
			"username":   record.Username,
			"effect":     record.Effect,
			"conclusion": record.Conclusion,
			"request":    record.Request,
			"policies":   record.Policies,
			"deciders":   record.Deciders,
			"expireAt":   record.ExpireAt,
		}

		tags := make(map[string]string)
		for _, tag := range i.conf.Tags {
			if encoded, err := json.Marshal(mapping[tag]); err == nil {
				tags[tag] = strings.Trim(string(encoded), "\"")
			}
		}

		fields := make(map[string]any)
		for _, field := range i.conf.Fields {
			fields[field] = mapping[field]
		}

		point, err := client.NewPoint(influxTable, tags, fields, time.Now())
		if err != nil {
			slog.WarnContext(ctx, "failed to create influx point", "error", err)

			continue
		}

		batchPoints.AddPoint(point)
	}

	if err := httpClient.Write(batchPoints); err != nil {
		return fmt.Errorf("write influx batch: %w", err)
	}

	return nil
}
