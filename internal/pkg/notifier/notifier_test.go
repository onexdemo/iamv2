package notifier_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onexdemo/iamv2/internal/pkg/notifier"
)

func TestSign(t *testing.T) {
	first := notifier.Sign(notifier.SecretChanged, "payload")
	second := notifier.Sign(notifier.SecretChanged, "payload")

	assert.Equal(t, first, second)
	assert.NotEqual(t, first, notifier.Sign(notifier.PolicyChanged, "payload"))
	assert.NotEmpty(t, first)
}

func TestParseNotification(t *testing.T) {
	notification := notifier.Notification{
		Command:   notifier.SecretChanged,
		Payload:   "payload",
		Signature: notifier.Sign(notifier.SecretChanged, "payload"),
		Algorithm: "SHA256",
	}

	data, err := json.Marshal(notification)
	require.NoError(t, err)

	parsed, err := notifier.ParseNotification(data)
	require.NoError(t, err)
	assert.Equal(t, notifier.SecretChanged, parsed.Command)

	notification.Signature = "tampered"
	data, err = json.Marshal(notification)
	require.NoError(t, err)

	_, err = notifier.ParseNotification(data)
	require.Error(t, err)
}
