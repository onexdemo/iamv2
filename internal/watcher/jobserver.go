package watcher

import (
	"context"

	"github.com/onexstack/onexstack/pkg/server"
	"github.com/onexstack/onexstack/pkg/watch"
	sloglogger "github.com/onexstack/onexstack/pkg/watch/logger/slog"

	"github.com/onexdemo/iamv2/internal/apiserver/store"
	"github.com/onexdemo/iamv2/internal/watcher/watcher"
	_ "github.com/onexdemo/iamv2/internal/watcher/watcher/all"
)

// jobServer implements server.Server on top of the onex watch scheduler.
type jobServer struct {
	watch *watch.Watch
}

var _ server.Server = (*jobServer)(nil)

// NewJobServer creates the watch scheduler and registers every watcher.
func (c *ServerConfig) NewJobServer() (server.Server, error) {
	aggregate := &watcher.AggregateConfig{
		Store:   store.NewStore(c.DB),
		Options: c.WatcherOptions,
	}

	opts := []watch.Option{
		watch.WithInitialize(watcher.NewInitializer(aggregate)),
		watch.WithLogger(sloglogger.NewLogger()),
		watch.WithAutoMigrate(true),
		watch.WithHealthzAddr(c.WatchOptions.HealthzAddr),
		watch.WithMetricsAddr(c.WatchOptions.MetricsAddr),
	}

	w, err := watch.NewWatch(c.WatchOptions, c.DB, opts...)
	if err != nil {
		return nil, err
	}

	return &jobServer{watch: w}, nil
}

// Run starts the watch scheduler and blocks until ctx is canceled.
func (s *jobServer) Run(ctx context.Context) error {
	s.watch.Start(ctx.Done())

	return nil
}

// GracefulStop stops the scheduler and releases the distributed lock.
func (s *jobServer) GracefulStop(context.Context) error {
	s.watch.Stop()

	return nil
}
