//go:build wireinject
// +build wireinject

package apiserver

import (
	"context"

	"github.com/google/wire"

	"github.com/onexdemo/iamv2/internal/apiserver/biz"
	"github.com/onexdemo/iamv2/internal/apiserver/handler"
	"github.com/onexdemo/iamv2/internal/apiserver/pkg/validation"
	"github.com/onexdemo/iamv2/internal/apiserver/store"
)

// infrastructureSet groups all infrastructure-related providers.
var infrastructureSet = wire.NewSet(
	ProvideDB,
	ProvideNotifier,
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
		store.ProviderSet,
		biz.ProviderSet,
		validation.ProviderSet,
		handler.NewHandler,

		// Infrastructure dependencies.
		infrastructureSet,
	)

	return nil, nil
}
