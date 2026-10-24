// Package analytics buffers authorization audit records and writes them to the
// Redis list consumed by iam-pump.
package analytics

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
)

// AnalyticsKey is the Redis key consumed by iam-pump. It intentionally keeps
// the original iam name so that an iam-pump deployment can be replaced
// without losing in-flight records.
const AnalyticsKey = "analytics-iam-system-analytics"

// Record encodes the details of an authorization request.
type Record struct {
	TimeStamp  int64     `json:"timestamp" msgpack:"timestamp"`
	Username   string    `json:"username" msgpack:"username"`
	Effect     string    `json:"effect" msgpack:"effect"`
	Conclusion string    `json:"conclusion" msgpack:"conclusion"`
	Request    string    `json:"request" msgpack:"request"`
	Policies   string    `json:"policies" msgpack:"policies"`
	Deciders   string    `json:"deciders" msgpack:"deciders"`
	ExpireAt   time.Time `json:"expireAt" bson:"expireAt" msgpack:"expireAt"`
}

// SetExpiry sets the record expiry. A zero duration means "keep for 100
// years".
func (r *Record) SetExpiry(expiresIn time.Duration) {
	if expiresIn <= 0 {
		expiresIn = 24 * 365 * 100 * time.Hour
	}

	r.ExpireAt = time.Now().Add(expiresIn)
}

// Recorder asynchronously writes audit records to Redis.
type Recorder struct {
	client   *redis.Client
	options  *Options
	records  chan *Record
	wg       sync.WaitGroup
	stopOnce sync.Once
}

// NewRecorder creates a recorder. Call Start before using Record and Stop on
// shutdown.
func NewRecorder(client *redis.Client, options *Options) *Recorder {
	return &Recorder{
		client:  client,
		options: options,
		records: make(chan *Record, options.RecordsBufferSize),
	}
}

// Start launches the worker pool.
func (r *Recorder) Start(ctx context.Context) {
	for range r.options.PoolSize {
		r.wg.Add(1)

		go r.worker(ctx)
	}
}

// Stop drains the buffer and waits for the workers to finish.
func (r *Recorder) Stop() {
	r.stopOnce.Do(func() {
		close(r.records)
		r.wg.Wait()
	})
}

// Record enqueues an audit record. It never blocks the authorization path for
// longer than the buffer takes to fill.
func (r *Recorder) Record(ctx context.Context, record *Record) error {
	if r == nil || r.options == nil || !r.options.Enable {
		return nil
	}

	select {
	case r.records <- record:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *Recorder) worker(ctx context.Context) {
	defer r.wg.Done()

	buffer := make([][]byte, 0, r.options.RecordsBufferSize/uint64(r.options.PoolSize))
	flushInterval := time.Duration(r.options.FlushInterval) * time.Millisecond
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	for {
		select {
		case record, ok := <-r.records:
			if !ok {
				r.flush(ctx, buffer)

				return
			}

			encoded, err := msgpack.Marshal(record)
			if err != nil {
				slog.ErrorContext(ctx, "failed to encode analytics record", "error", err)

				continue
			}

			buffer = append(buffer, encoded)

		case <-ticker.C:
			if len(buffer) == 0 {
				continue
			}

			r.flush(ctx, buffer)

			buffer = buffer[:0]

		case <-ctx.Done():
			r.flush(context.WithoutCancel(ctx), buffer)

			return
		}
	}
}

func (r *Recorder) flush(ctx context.Context, buffer [][]byte) {
	if len(buffer) == 0 {
		return
	}

	values := make([]any, 0, len(buffer))
	for _, item := range buffer {
		values = append(values, item)
	}

	pipe := r.client.Pipeline()
	pipe.RPush(ctx, AnalyticsKey, values...)
	pipe.Expire(ctx, AnalyticsKey, r.options.StorageExpirationTime)

	if _, err := pipe.Exec(ctx); err != nil {
		slog.ErrorContext(ctx, "failed to write analytics records", "error", err)
	}
}

// ErrDisabled is returned when analytics is disabled.
var ErrDisabled = errors.New("analytics is disabled")

// MarshalRecord exposes the wire encoding for tests and tooling.
func MarshalRecord(record *Record) ([]byte, error) {
	data, err := msgpack.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("marshal analytics record: %w", err)
	}

	return data, nil
}
