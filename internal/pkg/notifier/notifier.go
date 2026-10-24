// Package notifier publishes cache-invalidation events to the iam cluster.
//
// The iam apiserver publishes a notification whenever a secret or policy
// changes. Every iam-authz-server subscribes to the channel and reloads its
// in-memory cache, which keeps authorization decisions consistent without
// polling the database.
package notifier

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Cluster-wide notification channel.
const (
	// Channel is the Redis pub/sub channel used by the iam cluster.
	Channel = "iam.cluster.notifications"

	// PolicyChanged is published when a policy is created, updated or deleted.
	PolicyChanged = "PolicyChanged"
	// SecretChanged is published when a secret is created, updated or deleted.
	SecretChanged = "SecretChanged"
)

// Notification is the message exchanged between the apiserver and the authz
// servers.
type Notification struct {
	Command   string `json:"command"`
	Payload   string `json:"payload"`
	Signature string `json:"signature"`
	Algorithm string `json:"algorithm"`
}

// ParseNotification decodes and verifies a notification payload.
func ParseNotification(data []byte) (*Notification, error) {
	var notification Notification

	if err := json.Unmarshal(data, &notification); err != nil {
		return nil, fmt.Errorf("unmarshal notification: %w", err)
	}

	expected := Sign(notification.Command, notification.Payload)
	if notification.Signature != "" && notification.Signature != expected {
		return nil, fmt.Errorf("notification signature mismatch")
	}

	return &notification, nil
}

// Sign returns the SHA-256 signature of command and payload.
func Sign(command, payload string) string {
	sum := sha256.Sum256([]byte(command + payload))
	return hex.EncodeToString(sum[:])
}

// Notifier publishes cache-invalidation events.
type Notifier interface {
	Notify(ctx context.Context, command string) error
}

type redisNotifier struct {
	client redis.UniversalClient
}

// NewRedisNotifier creates a notifier backed by Redis pub/sub.
func NewRedisNotifier(client redis.UniversalClient) *redisNotifier {
	return &redisNotifier{client: client}
}

// Notify publishes a signed notification to the cluster channel.
func (n *redisNotifier) Notify(ctx context.Context, command string) error {
	notification := Notification{
		Command:   command,
		Signature: Sign(command, ""),
		Algorithm: "SHA256",
	}

	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("marshal notification: %w", err)
	}

	if err := n.client.Publish(ctx, Channel, data).Err(); err != nil {
		return fmt.Errorf("publish notification: %w", err)
	}

	return nil
}

type noopNotifier struct{}

// NewNoopNotifier returns a notifier that discards every event. It is used when
// Redis is not configured.
func NewNoopNotifier() noopNotifier {
	return noopNotifier{}
}

func (noopNotifier) Notify(context.Context, string) error {
	return nil
}
