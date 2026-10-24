package pump

import (
	"context"

	"github.com/onexstack/onexstack/pkg/server"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/onexdemo/iamv2/internal/pump/pumps"
	"github.com/onexdemo/iamv2/internal/pump/storage"
)

// ProvideDB provides the database used for watch locking.
func ProvideDB(cfg *Config) (*gorm.DB, error) {
	return cfg.NewDB()
}

// ProvideRedisClient provides the Redis client used to drain analytics data.
func ProvideRedisClient(cfg *Config) (*redis.Client, error) {
	return cfg.RedisOptions.NewClient()
}

// ProvideStorage provides the analytics storage.
func ProvideStorage(client *redis.Client) *storage.RedisStorage {
	return storage.NewRedisStorage(client)
}

// ProvideManager initializes every configured pump.
func ProvideManager(cfg *Config) (*pumps.Manager, error) {
	return pumps.NewManager(cfg.PumpOptions.Pumps)
}

// NewDependencies initializes components that need to be started but are not
// directly stored on the server struct.
func NewDependencies(_ context.Context) *Dependencies {
	return &Dependencies{}
}

// NewJobServer creates the watch server.
func NewJobServer(serverConfig *ServerConfig) (server.Server, error) {
	return serverConfig.NewJobServer()
}
