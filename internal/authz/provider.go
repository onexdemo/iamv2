package authz

import (
	"context"

	"github.com/onexstack/onexstack/pkg/server"
	"github.com/redis/go-redis/v9"

	"github.com/onexdemo/iamv2/internal/authz/pkg/analytics"
	"github.com/onexdemo/iamv2/internal/authz/pkg/authorization"
	"github.com/onexdemo/iamv2/internal/authz/pkg/cache"
	"github.com/onexdemo/iamv2/internal/authz/store"
	apiserverstore "github.com/onexdemo/iamv2/internal/authz/store/apiserver"
)

// ProvideStore provides the read-only apiserver store.
func ProvideStore(cfg *Config) store.Factory {
	return apiserverstore.New(
		cfg.APIServerAddr,
		cfg.APIServerUsername,
		cfg.APIServerPassword,
		cfg.APIServerTimeout,
	)
}

// ProvideRedisClient provides the Redis client used for pub/sub and analytics.
func ProvideRedisClient(cfg *Config) (*redis.Client, error) {
	return cfg.RedisOptions.NewClient()
}

// ProvideCache provides the in-memory secret/policy cache.
func ProvideCache(factory store.Factory) *cache.Cache {
	return cache.New(factory)
}

// ProvideRecorder provides the asynchronous analytics recorder.
func ProvideRecorder(client *redis.Client, cfg *Config) *analytics.Recorder {
	return analytics.NewRecorder(client, cfg.AnalyticsOptions)
}

// ProvideAuthorization provides the ladon policy source and audit sink.
func ProvideAuthorization(
	c *cache.Cache,
	recorder *analytics.Recorder,
	cfg *Config,
) authorization.Interface {

	return authorization.NewAuthorization(c, recorder, cfg.AnalyticsOptions.EnableDetailedRecording)
}

// ProvideAuthorizer provides the ladon authorizer.
func ProvideAuthorizer(client authorization.Interface) *authorization.Authorizer {
	return authorization.NewAuthorizer(client)
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
