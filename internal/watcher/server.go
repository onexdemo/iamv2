package watcher

import (
	"context"
	"log/slog"
	"time"

	genericoptions "github.com/onexstack/onexstack/pkg/options"
	"github.com/onexstack/onexstack/pkg/server"
	"github.com/onexstack/onexstack/pkg/store/registry"
	"github.com/onexstack/onexstack/pkg/watch"
	"gorm.io/gorm"

	"github.com/onexdemo/iamv2/internal/watcher/options"
)

// Dependencies collects components that need initialization but are not used
// directly by the server struct.
type Dependencies struct{}

// Config contains application-level configuration for the watcher.
type Config struct {
	WatchOptions   *watch.Options
	MySQLOptions   *genericoptions.MySQLOptions
	WatcherOptions *options.WatcherOptions
}

// Server represents the watch scheduler.
type Server struct {
	cfg *ServerConfig
	srv server.Server
}

// ServerConfig contains the core dependencies of the server.
type ServerConfig struct {
	*Config

	DB           *gorm.DB
	Dependencies *Dependencies
}

// New creates and returns a new Server instance.
func (cfg *Config) New(ctx context.Context) (*Server, error) {
	s, err := NewServer(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return s.Prepare(ctx)
}

// Prepare performs post-initialization tasks.
func (s *Server) Prepare(_ context.Context) (*Server, error) {
	return s, nil
}

// Run starts the watch scheduler and blocks until the context is canceled.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.srv.Run(ctx)
	}()

	<-ctx.Done()

	slog.Info("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
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

	return dbInstance, nil
}
