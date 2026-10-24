//go:build wireinject
// +build wireinject

package authz

import (
	"context"

	"github.com/google/wire"

	"github.com/onexdemo/iamv2/internal/authz/biz"
	"github.com/onexdemo/iamv2/internal/authz/handler"
)

// infrastructureSet groups all infrastructure-related providers.
var infrastructureSet = wire.NewSet(
	ProvideStore,
	ProvideRedisClient,
	ProvideCache,
	ProvideRecorder,
	ProvideAuthorization,
	ProvideAuthorizer,
)

// NewServer initializes and creates the web server with all dependencies.
func NewServer(context.Context, *Config) (*Server, error) {
	wire.Build(
		// Server infrastructure.
		NewWebServer,
		NewDependencies,
		NewAuthenticator,
		wire.Struct(new(ServerConfig), "*"),
		wire.Struct(new(Server), "*"),

		// Domain layers.
		biz.ProviderSet,
		handler.NewHandler,

		// Infrastructure dependencies.
		infrastructureSet,
	)

	return nil, nil
}
