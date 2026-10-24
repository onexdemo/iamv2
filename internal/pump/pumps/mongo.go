package pumps

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/onexdemo/iamv2/internal/pump/analytics"
)

// Size units.
const (
	KiB = 1 << (10 * (iota + 1))
	MiB
	GiB
)

// MongoConf configures the MongoDB pump.
type MongoConf struct {
	MongoURL                  string `mapstructure:"mongo_url"`
	CollectionName            string `mapstructure:"collection_name"`
	MaxInsertBatchSizeBytes   int    `mapstructure:"max_insert_batch_size_bytes"`
	MaxDocumentSizeBytes      int    `mapstructure:"max_document_size_bytes"`
	CollectionCapMaxSizeBytes int    `mapstructure:"collection_cap_max_size_bytes"`
	CollectionCapEnable       bool   `mapstructure:"collection_cap_enable"`
}

// MongoPump writes records to MongoDB.
type MongoPump struct {
	client     *mongo.Client
	collection *mongo.Collection
	conf       *MongoConf
	CommonPumpConfig
}

// New returns a fresh MongoDB pump.
func (m *MongoPump) New() Pump {
	return &MongoPump{}
}

// GetName returns the pump name.
func (*MongoPump) GetName() string {
	return "MongoDB Pump"
}

// Init connects to MongoDB and prepares the collection.
func (m *MongoPump) Init(config any) error {
	m.conf = &MongoConf{}
	if err := mapstructure.Decode(config, m.conf); err != nil {
		return fmt.Errorf("decode mongo configuration: %w", err)
	}

	if m.conf.MongoURL == "" {
		return errors.New("mongo_url is required")
	}

	if m.conf.CollectionName == "" {
		return errors.New("collection_name is required")
	}

	if m.conf.MaxInsertBatchSizeBytes == 0 {
		m.conf.MaxInsertBatchSizeBytes = 10 * MiB
	}

	if m.conf.MaxDocumentSizeBytes == 0 {
		m.conf.MaxDocumentSizeBytes = 10 * MiB
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(m.conf.MongoURL))
	if err != nil {
		return fmt.Errorf("connect to mongo: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return fmt.Errorf("ping mongo: %w", err)
	}

	m.client = client
	m.collection = client.Database("").Collection(m.conf.CollectionName)

	m.capCollection(ctx)

	if err := m.ensureIndexes(ctx); err != nil {
		slog.Warn("failed to ensure mongo indexes", "error", err)
	}

	slog.Info("initialized mongo pump", "collection", m.conf.CollectionName)

	return nil
}

func (m *MongoPump) capCollection(ctx context.Context) {
	if !m.conf.CollectionCapEnable {
		return
	}

	names, err := m.client.Database("").ListCollectionNames(ctx, bson.M{"name": m.conf.CollectionName})
	if err != nil {
		slog.Warn("failed to list mongo collections", "error", err)

		return
	}

	if len(names) > 0 {
		slog.Warn("mongo collection already exists, not capping it", "collection", m.conf.CollectionName)

		return
	}

	size := int64(m.conf.CollectionCapMaxSizeBytes)
	if size == 0 {
		size = 5 * GiB
	}

	err = m.client.Database("").CreateCollection(ctx, m.conf.CollectionName, options.CreateCollection().SetCapped(true).SetSizeInBytes(size))
	if err != nil {
		slog.Warn("failed to create capped mongo collection", "error", err)

		return
	}

	slog.Info("created capped mongo collection", "collection", m.conf.CollectionName, "bytes", size)
}

func (m *MongoPump) ensureIndexes(ctx context.Context) error {
	_, err := m.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "timestamp", Value: -1}, {Key: "username", Value: 1}},
	})

	return err
}

// WriteData inserts records in size-bounded batches.
func (m *MongoPump) WriteData(ctx context.Context, data []any) error {
	if m.collection == nil {
		return errors.New("mongo pump is not initialized")
	}

	errs := []error{}
	for _, batch := range m.accumulateSet(data) {
		documents := make([]any, 0, len(batch))
		documents = append(documents, batch...)

		if _, err := m.collection.InsertMany(ctx, documents); err != nil {
			errs = append(errs, fmt.Errorf("insert mongo batch: %w", err))
		}
	}

	return errors.Join(errs...)
}

func (m *MongoPump) accumulateSet(data []any) [][]any {
	accumulatorTotal := 0
	result := make([][]any, 0)
	current := make([]any, 0)

	for _, item := range data {
		record, _ := item.(analytics.Record)
		sizeBytes := len(record.Policies) + len(record.Deciders) + 1024

		if sizeBytes > m.conf.MaxDocumentSizeBytes {
			slog.Warn("analytics document is too large, stripping details")

			record.Policies = ""
			record.Deciders = ""
			sizeBytes = 1024
		}

		if accumulatorTotal+sizeBytes > m.conf.MaxInsertBatchSizeBytes && len(current) > 0 {
			result = append(result, current)
			current = make([]any, 0)
			accumulatorTotal = 0
		}

		accumulatorTotal += sizeBytes
		current = append(current, record)
	}

	if len(current) > 0 {
		result = append(result, current)
	}

	return result
}
