package watcher

import (
	"context"

	"github.com/onexstack/onexstack/pkg/server"
	"gorm.io/gorm"
)

// ProvideDB provides a database instance based on the configuration.
func ProvideDB(cfg *Config) (*gorm.DB, error) {
	return cfg.NewDB()
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
