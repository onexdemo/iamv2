// Package task implements the inactive-account maintenance watcher.
package task

import (
	"context"
	"log/slog"
	"time"

	"github.com/onexstack/onexstack/pkg/store/where"
	"github.com/onexstack/onexstack/pkg/watch/registry"

	watcherpkg "github.com/onexdemo/iamv2/internal/watcher/watcher"
)

// watcherSpec runs once a day.
const watcherSpec = "@daily"

// maxUsers bounds a single maintenance pass.
const maxUsers = 1 << 20

type taskWatcher struct {
	config *watcherpkg.AggregateConfig
}

var (
	_ registry.Watcher                = (*taskWatcher)(nil)
	_ registry.ISpec                  = (*taskWatcher)(nil)
	_ watcherpkg.WantsAggregateConfig = (*taskWatcher)(nil)
)

// SetAggregateConfig injects the shared dependencies.
func (w *taskWatcher) SetAggregateConfig(config *watcherpkg.AggregateConfig) {
	w.config = config
}

// Run disables accounts that have been inactive for too long.
func (w *taskWatcher) Run() {
	if w.config == nil || w.config.Store == nil {
		slog.Warn("task watcher is not initialized")

		return
	}

	maxInactiveDays := w.config.Options.Task.MaxInactiveDays
	if maxInactiveDays <= 0 {
		return
	}

	ctx := context.Background()
	_, users, err := w.config.Store.User().List(ctx, where.NewWhere().L(maxUsers))
	if err != nil {
		slog.ErrorContext(ctx, "failed to list users", "error", err)

		return
	}

	inactiveBefore := time.Now().AddDate(0, 0, -maxInactiveDays)
	for _, user := range users {
		if user.LoginedAt == nil || user.LoginedAt.After(inactiveBefore) {
			continue
		}

		slog.InfoContext(ctx, "disabling inactive account", "username", user.Username, "max_inactive_days", maxInactiveDays)

		user.Status = 0
		if err := w.config.Store.User().Update(ctx, user); err != nil {
			slog.WarnContext(ctx, "failed to disable inactive account", "username", user.Username, "error", err)
		}
	}
}

// Spec returns the cron schedule.
func (*taskWatcher) Spec() string {
	return watcherSpec
}

func init() {
	registry.Register("task", &taskWatcher{})
}
