package store

import (
	"context"

	storelogger "github.com/onexstack/onexstack/pkg/logger/slog/store"
	genericstore "github.com/onexstack/onexstack/pkg/store"
	"github.com/onexstack/onexstack/pkg/store/where"

	"github.com/onexdemo/iamv2/internal/apiserver/model"
)

// UserStore defines the persistence operations for users.
type UserStore interface {
	Create(ctx context.Context, obj *model.UserM) error
	Update(ctx context.Context, obj *model.UserM) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.UserM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.UserM, error)

	UserExpansion
}

// UserExpansion groups non-CRUD user store operations.
type UserExpansion any

type userStore struct {
	*genericstore.Store[model.UserM]
}

var _ UserStore = (*userStore)(nil)

func newUserStore(s *store) *userStore {
	return &userStore{
		Store: genericstore.NewStore[model.UserM](s, storelogger.NewLogger()),
	}
}
