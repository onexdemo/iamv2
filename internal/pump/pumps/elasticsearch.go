package pumps

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	elastic "github.com/olivere/elastic/v7"

	"github.com/onexdemo/iamv2/internal/pump/analytics"
)

// ElasticsearchBulkConfig configures the bulk processor.
type ElasticsearchBulkConfig struct {
	Workers       int `mapstructure:"workers"`
	FlushInterval int `mapstructure:"flush_interval"`
	BulkActions   int `mapstructure:"bulk_actions"`
	BulkSize      int `mapstructure:"bulk_size"`
}

// ElasticsearchConf configures the Elasticsearch pump.
type ElasticsearchConf struct {
	BulkConfig       ElasticsearchBulkConfig `mapstructure:"bulk_config"`
	IndexName        string                  `mapstructure:"index_name"`
	ElasticsearchURL string                  `mapstructure:"elasticsearch_url"`
	DocumentType     string                  `mapstructure:"document_type"`
	AuthAPIKeyID     string                  `mapstructure:"auth_api_key_id"`
	AuthAPIKey       string                  `mapstructure:"auth_api_key"`
	Username         string                  `mapstructure:"auth_basic_username"`
	Password         string                  `mapstructure:"auth_basic_password"`
	EnableSniffing   bool                    `mapstructure:"use_sniffing"`
	RollingIndex     bool                    `mapstructure:"rolling_index"`
	DisableBulk      bool                    `mapstructure:"disable_bulk"`
}

// apiKeyTransport adds an Elasticsearch API key to every request.
type apiKeyTransport struct {
	apiKey   string
	apiKeyID string
}

// RoundTrip implements http.RoundTripper.
func (t *apiKeyTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	auth := t.apiKeyID + ":" + t.apiKey
	key := base64.StdEncoding.EncodeToString([]byte(auth))
	r.Header.Set("Authorization", "ApiKey "+key)

	return http.DefaultTransport.RoundTrip(r)
}

// ElasticsearchPump writes records to Elasticsearch.
type ElasticsearchPump struct {
	client        *elastic.Client
	bulkProcessor *elastic.BulkProcessor
	conf          *ElasticsearchConf
	CommonPumpConfig
}

// New returns a fresh Elasticsearch pump.
func (e *ElasticsearchPump) New() Pump {
	return &ElasticsearchPump{}
}

// GetName returns the pump name.
func (*ElasticsearchPump) GetName() string {
	return "Elasticsearch Pump"
}

// Init connects to Elasticsearch and starts the bulk processor.
func (e *ElasticsearchPump) Init(config any) error {
	e.conf = &ElasticsearchConf{}
	if err := mapstructure.Decode(config, e.conf); err != nil {
		return fmt.Errorf("decode elasticsearch configuration: %w", err)
	}

	if e.conf.IndexName == "" {
		e.conf.IndexName = "iam_analytics"
	}

	if e.conf.ElasticsearchURL == "" {
		e.conf.ElasticsearchURL = "http://localhost:9200"
	}

	if e.conf.DocumentType == "" {
		e.conf.DocumentType = "iam_analytics"
	}

	urls := strings.Split(e.conf.ElasticsearchURL, ",")
	httpClient := http.DefaultClient
	if e.conf.AuthAPIKey != "" && e.conf.AuthAPIKeyID != "" {
		e.conf.Username = ""
		e.conf.Password = ""
		httpClient = &http.Client{
			Transport: &apiKeyTransport{apiKey: e.conf.AuthAPIKey, apiKeyID: e.conf.AuthAPIKeyID},
		}
	}

	client, err := elastic.NewClient(
		elastic.SetURL(urls...),
		elastic.SetSniff(e.conf.EnableSniffing),
		elastic.SetBasicAuth(e.conf.Username, e.conf.Password),
		elastic.SetHttpClient(httpClient),
	)
	if err != nil {
		return fmt.Errorf("create elasticsearch client: %w", err)
	}

	e.client = client

	processor := client.BulkProcessor().Name("iam-pump-bulk-processor")
	if e.conf.BulkConfig.Workers != 0 {
		processor = processor.Workers(e.conf.BulkConfig.Workers)
	}

	if e.conf.BulkConfig.FlushInterval != 0 {
		processor = processor.FlushInterval(time.Duration(e.conf.BulkConfig.FlushInterval) * time.Second)
	}

	if e.conf.BulkConfig.BulkActions != 0 {
		processor = processor.BulkActions(e.conf.BulkConfig.BulkActions)
	}

	if e.conf.BulkConfig.BulkSize != 0 {
		processor = processor.BulkSize(e.conf.BulkConfig.BulkSize)
	}

	e.bulkProcessor, err = processor.Do(context.Background())
	if err != nil {
		return fmt.Errorf("start elasticsearch bulk processor: %w", err)
	}

	slog.Info("initialized elasticsearch pump", "index", e.conf.IndexName)

	return nil
}

// WriteData indexes a batch of records.
func (e *ElasticsearchPump) WriteData(ctx context.Context, data []any) error {
	if e.client == nil {
		return fmt.Errorf("elasticsearch pump is not initialized")
	}

	index := e.client.Index().Index(e.indexName())
	for _, item := range data {
		if err := ctx.Err(); err != nil {
			return err
		}

		record, ok := item.(analytics.Record)
		if !ok {
			continue
		}

		doc := map[string]any{
			"@timestamp": record.TimeStamp,
			"username":   record.Username,
			"effect":     record.Effect,
			"conclusion": record.Conclusion,
			"request":    record.Request,
			"policies":   record.Policies,
			"deciders":   record.Deciders,
			"expireAt":   record.ExpireAt,
		}

		if !e.conf.DisableBulk {
			e.bulkProcessor.Add(elastic.NewBulkIndexRequest().Index(e.indexName()).Doc(doc))

			continue
		}

		if _, err := index.BodyJson(doc).Do(ctx); err != nil {
			slog.WarnContext(ctx, "failed to index elasticsearch document", "error", err)
		}
	}

	return nil
}

func (e *ElasticsearchPump) indexName() string {
	if !e.conf.RollingIndex {
		return e.conf.IndexName
	}

	return e.conf.IndexName + "-" + time.Now().Format("2006.01.02")
}
