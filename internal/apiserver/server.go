package apiserver

import (
	"context"
	"log/slog"
	"time"

	genericoptions "github.com/onexstack/onexstack/pkg/options"
	"github.com/onexstack/onexstack/pkg/server"
	"github.com/onexstack/onexstack/pkg/store/registry"
	"github.com/onexstack/onexstack/pkg/token"
	"gorm.io/gorm"

	"github.com/onexdemo/iamv2/internal/apiserver/handler"
	"github.com/onexdemo/iamv2/internal/apiserver/pkg/metrics"
	"github.com/onexdemo/iamv2/internal/apiserver/seed"
	apiserverstore "github.com/onexdemo/iamv2/internal/apiserver/store"
	mw "github.com/onexdemo/iamv2/internal/pkg/middleware/gin"
)

const serviceName = "iam-apiserver"

// Dependencies collects components that need initialization but are not used
// directly by the server struct.
type Dependencies struct{}

// Config contains application-level configuration for the apiserver.
type Config struct {
	SecureServingOptions   *genericoptions.SecureServingOptions
	InsecureServingOptions *genericoptions.InsecureServingOptions
	MySQLOptions           *genericoptions.MySQLOptions
	RedisOptions           *genericoptions.RedisOptions
	JWTOptions             *genericoptions.JWTOptions

	// Bootstrap administrator. Created when the database has no administrator.
	AdminUsername string
	AdminPassword string
	AdminEmail    string
}

// Server represents the HTTP server and its background workers.
type Server struct {
	cfg *ServerConfig
	srv server.Server
}

// ServerConfig contains the core dependencies of the server.
type ServerConfig struct {
	*Config

	Dependencies  *Dependencies
	Handler       *handler.Handler
	Authenticator mw.Authenticator
}

// New creates and returns a new Server instance.
func (cfg *Config) New(ctx context.Context) (*Server, error) {
	s, err := NewServer(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return s.Prepare(ctx)
}

// Prepare performs post-initialization tasks such as configuring the token
// package.
func (s *Server) Prepare(_ context.Context) (*Server, error) {
	token.Init(
		s.cfg.JWTOptions.Key,
		token.WithIdentityKey(identityClaim),
		token.WithExpiration(s.cfg.JWTOptions.Expired),
		token.WithCommonSkipPaths(),
	)

	_ = metrics.Init(serviceName)

	return s, nil
}

// Run starts the server and blocks until the context is canceled.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.srv.Run(ctx)
	}()

	<-ctx.Done()

	slog.Info("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.srv.GracefulStop(shutdownCtx); err != nil {
		slog.Error("failed to gracefully stop server", "error", err)
	}

	slog.Info("server exited successfully")

	return <-errCh
}

// NewDB creates the MySQL connection and runs schema migration.
func (cfg *Config) NewDB() (*gorm.DB, error) {
	slog.Info("initializing database connection", "type", "mysql")

	dbInstance, err := cfg.MySQLOptions.NewDB()
	if err != nil {
		slog.Error("failed to create database connection", "error", err)

		return nil, err
	}

	if err := registry.Migrate(dbInstance); err != nil {
		slog.Error("failed to migrate database schema", "error", err)

		return nil, err
	}

	seedCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := seed.EnsureAdmin(seedCtx, apiserverstore.NewStore(dbInstance), seed.AdminOptions{
		Username: cfg.AdminUsername,
		Password: cfg.AdminPassword,
		Email:    cfg.AdminEmail,
	}); err != nil {
		slog.Error("failed to seed the bootstrap administrator", "error", err)

		return nil, err
	}

	return dbInstance, nil
}
