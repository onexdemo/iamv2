package store

import (
	"context"

	storelogger "github.com/onexstack/onexstack/pkg/logger/slog/store"
	genericstore "github.com/onexstack/onexstack/pkg/store"
	"github.com/onexstack/onexstack/pkg/store/where"

	"github.com/onexdemo/iamv2/internal/apiserver/model"
)

// SecretStore defines the persistence operations for access keys.
type SecretStore interface {
	Create(ctx context.Context, obj *model.SecretM) error
	Update(ctx context.Context, obj *model.SecretM) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.SecretM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.SecretM, error)

	SecretExpansion
}

// SecretExpansion groups non-CRUD secret store operations.
type SecretExpansion any

type secretStore struct {
	*genericstore.Store[model.SecretM]
}

var _ SecretStore = (*secretStore)(nil)

func newSecretStore(s *store) *secretStore {
	return &secretStore{
		Store: genericstore.NewStore[model.SecretM](s, storelogger.NewLogger()),
	}
}
