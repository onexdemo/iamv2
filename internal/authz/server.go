package authz

import (
	"context"
	"log/slog"
	"time"

	genericoptions "github.com/onexstack/onexstack/pkg/options"
	"github.com/onexstack/onexstack/pkg/server"
	"github.com/redis/go-redis/v9"

	"github.com/onexdemo/iamv2/internal/authz/handler"
	"github.com/onexdemo/iamv2/internal/authz/pkg/analytics"
	"github.com/onexdemo/iamv2/internal/authz/pkg/cache"
	"github.com/onexdemo/iamv2/internal/authz/pkg/metrics"
	mw "github.com/onexdemo/iamv2/internal/pkg/middleware/gin"
)

const serviceName = "iam-authz-server"

// Dependencies collects components that need initialization but are not used
// directly by the server struct.
type Dependencies struct{}

// Config contains application-level configuration for the authz server.
type Config struct {
	SecureServingOptions   *genericoptions.SecureServingOptions
	InsecureServingOptions *genericoptions.InsecureServingOptions
	RedisOptions           *genericoptions.RedisOptions

	// APIServerAddr is the base URL of iam-apiserver, for example
	// http://127.0.0.1:8080.
	APIServerAddr string
	// APIServerUsername is an administrator account used to fetch cache
	// snapshots.
	APIServerUsername string
	// APIServerPassword is the administrator password.
	APIServerPassword string
	// APIServerTimeout bounds each cache snapshot request.
	APIServerTimeout time.Duration

	// CacheReloadInterval is the fallback reload interval when no cluster
	// notification is received.
	CacheReloadInterval time.Duration

	// AnalyticsOptions controls authorization audit recording.
	AnalyticsOptions *analytics.Options
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
	Cache         *cache.Cache
	RedisClient   *redis.Client
	Recorder      *analytics.Recorder
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

// Prepare performs post-initialization tasks.
func (s *Server) Prepare(_ context.Context) (*Server, error) {
	_ = metrics.Init(serviceName)
	return s, nil
}

// Run starts the HTTP server and the cache/analytics background workers.
func (s *Server) Run(ctx context.Context) error {
	// Load the initial snapshot before accepting traffic, then keep it fresh
	// through Redis notifications and a periodic fallback.
	s.cfg.Cache.Run(ctx, s.cfg.RedisClient, s.cfg.CacheReloadInterval)

	if s.cfg.AnalyticsOptions.Enable {
		s.cfg.Recorder.Start(ctx)
	}

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

	if s.cfg.AnalyticsOptions.Enable {
		s.cfg.Recorder.Stop()
	}

	slog.Info("server exited successfully")

	return <-errCh
}
