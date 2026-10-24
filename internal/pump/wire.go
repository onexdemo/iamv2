//go:build wireinject
// +build wireinject

package pump

import (
	"context"

	"github.com/google/wire"
)

// infrastructureSet groups all infrastructure-related providers.
var infrastructureSet = wire.NewSet(
	ProvideDB,
	ProvideRedisClient,
	ProvideStorage,
	ProvideManager,
)

// NewServer initializes and creates the watch server with all dependencies.
func NewServer(context.Context, *Config) (*Server, error) {
	wire.Build(
		// Server infrastructure.
		NewJobServer,
		NewDependencies,
		wire.Struct(new(ServerConfig), "*"),
		wire.Struct(new(Server), "*"),

		// Infrastructure dependencies.
		infrastructureSet,
	)

	return nil, nil
}
