package apiserver

import (
	"context"
	"log/slog"

	"github.com/onexstack/onexstack/pkg/server"
	"gorm.io/gorm"

	"github.com/onexdemo/iamv2/internal/pkg/notifier"
)

// ProvideDB provides a database instance based on the configuration.
func ProvideDB(cfg *Config) (*gorm.DB, error) {
	return cfg.NewDB()
}

// ProvideNotifier provides a Redis-backed notifier. When Redis is not
// configured the notifier becomes a no-op so that a single-node deployment
// still works.
func ProvideNotifier(cfg *Config) (notifier.Notifier, error) {
	if cfg.RedisOptions == nil || cfg.RedisOptions.Addr == "" {
		slog.Warn("redis is not configured; cache notifications are disabled")

		return notifier.NewNoopNotifier(), nil
	}

	client, err := cfg.RedisOptions.NewClient()
	if err != nil {
		return nil, err
	}

	return notifier.NewRedisNotifier(client), nil
}

// NewDependencies initializes components that need to be started but are not
// directly stored on the server struct.
func NewDependencies(_ context.Context) *Dependencies {
	return &Dependencies{}
}

// NewWebServer creates the HTTP server from the server configuration.
func NewWebServer(serverConfig *ServerConfig) (server.Server, error) {
	return serverConfig.NewGinServer()
}
