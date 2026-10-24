// Package cache keeps an in-memory snapshot of secrets and policies.
package cache

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/ory/ladon"
	"github.com/redis/go-redis/v9"

	"github.com/onexdemo/iamv2/internal/authz/store"
	"github.com/onexdemo/iamv2/internal/pkg/notifier"
	v1 "github.com/onexdemo/iamv2/pkg/api/apiserver/v1"
)

var (
	// ErrSecretNotFound is returned when a secret ID is absent from the cache.
	ErrSecretNotFound = errors.New("secret not found")
	// ErrPolicyNotFound is returned when a user has no policies in the cache.
	ErrPolicyNotFound = errors.New("policy not found")
)

// Cache stores secrets and policies in memory.
type Cache struct {
	lock     sync.RWMutex
	store    store.Factory
	secrets  map[string]*v1.SecretInfo
	policies map[string][]*ladon.DefaultPolicy
}

// New creates an empty cache backed by store.
func New(factory store.Factory) *Cache {
	return &Cache{
		store:    factory,
		secrets:  make(map[string]*v1.SecretInfo),
		policies: make(map[string][]*ladon.DefaultPolicy),
	}
}

// Reload replaces the cached snapshot with a fresh one from the apiserver.
func (c *Cache) Reload(ctx context.Context) error {
	secrets, err := c.store.Secrets().List(ctx)
	if err != nil {
		return err
	}

	policies, err := c.store.Policies().List(ctx)
	if err != nil {
		return err
	}

	c.lock.Lock()

	c.secrets = secrets
	c.policies = policies
	c.lock.Unlock()

	slog.InfoContext(ctx, "authorization cache reloaded", "secrets", len(secrets), "users", len(policies))

	return nil
}

// GetSecret returns a secret by its access-key identifier.
func (c *Cache) GetSecret(key string) (*v1.SecretInfo, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	secret, ok := c.secrets[key]
	if !ok {
		return nil, ErrSecretNotFound
	}

	return secret, nil
}

// GetPolicy returns the ladon policies owned by the given user.
func (c *Cache) GetPolicy(key string) ([]*ladon.DefaultPolicy, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	policies, ok := c.policies[key]
	if !ok {
		return nil, ErrPolicyNotFound
	}

	return policies, nil
}

// Run keeps the cache warm. It blocks until the first snapshot is loaded,
// retrying when the apiserver is not reachable yet, then refreshes on every
// cluster notification and on a periodic fallback tick.
func (c *Cache) Run(ctx context.Context, client *redis.Client, interval time.Duration) {
	for {
		if err := c.Reload(ctx); err == nil {
			break
		} else {
			slog.WarnContext(ctx, "initial cache reload failed, retrying", "error", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
		}
	}

	go c.subscribe(ctx, client)

	if interval <= 0 {
		interval = time.Minute
	}

	go c.tick(ctx, interval)
}

func (c *Cache) tick(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := c.Reload(ctx); err != nil {
				slog.WarnContext(ctx, "periodic cache reload failed", "error", err)
			}
		}
	}
}

func (c *Cache) subscribe(ctx context.Context, client *redis.Client) {
	pubsub := client.Subscribe(ctx, notifier.Channel)
	defer pubsub.Close()

	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}

			slog.WarnContext(ctx, "cache notification subscription failed, retrying", "error", err)

			select {
			case <-ctx.Done():
				return
			case <-time.After(10 * time.Second):
			}

			continue
		}

		notification, err := notifier.ParseNotification([]byte(msg.Payload))
		if err != nil {
			slog.WarnContext(ctx, "discarding malformed cache notification", "error", err)

			continue
		}

		switch notification.Command {
		case notifier.PolicyChanged, notifier.SecretChanged:
			if err := c.Reload(ctx); err != nil {
				slog.WarnContext(ctx, "cache reload after notification failed", "error", err)
			}

		default:
			slog.WarnContext(ctx, "unknown cache notification command", "command", notification.Command)
		}
	}
}
