package analytics_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/onexdemo/iamv2/internal/authz/pkg/analytics"
)

func TestMarshalRecord(t *testing.T) {
	record := &analytics.Record{
		TimeStamp:  time.Now().Unix(),
		Username:   "alice",
		Effect:     "allow",
		Conclusion: "policy permitted access",
		Request:    `{"resource":"articles"}`,
		Policies:   "read-articles",
		Deciders:   "read-articles",
	}
	record.SetExpiry(time.Hour)

	data, err := analytics.MarshalRecord(record)
	require.NoError(t, err)

	var decoded analytics.Record
	require.NoError(t, msgpack.Unmarshal(data, &decoded))

	assert.Equal(t, record.Username, decoded.Username)
	assert.Equal(t, record.Effect, decoded.Effect)
	assert.WithinDuration(t, record.ExpireAt, decoded.ExpireAt, time.Second)
}

func TestRecordSetExpiry(t *testing.T) {
	record := &analytics.Record{}
	record.SetExpiry(0)

	assert.True(t, record.ExpireAt.After(time.Now().Add(99*365*24*time.Hour)))
}
