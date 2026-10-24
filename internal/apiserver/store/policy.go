package store

import (
	"context"

	storelogger "github.com/onexstack/onexstack/pkg/logger/slog/store"
	genericstore "github.com/onexstack/onexstack/pkg/store"
	"github.com/onexstack/onexstack/pkg/store/where"

	"github.com/onexdemo/iamv2/internal/apiserver/model"
)

// PolicyStore defines the persistence operations for ladon policies.
type PolicyStore interface {
	Create(ctx context.Context, obj *model.PolicyM) error
	Update(ctx context.Context, obj *model.PolicyM) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.PolicyM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.PolicyM, error)

	PolicyExpansion
}

// PolicyExpansion groups non-CRUD policy store operations.
type PolicyExpansion interface{}

type policyStore struct {
	*genericstore.Store[model.PolicyM]
}

var _ PolicyStore = (*policyStore)(nil)

func newPolicyStore(s *store) *policyStore {
	return &policyStore{
		Store: genericstore.NewStore[model.PolicyM](s, storelogger.NewLogger()),
	}
}
