package store

import (
	"context"

	"github.com/google/wire"
	"github.com/onexstack/onexstack/pkg/store/where"
	"gorm.io/gorm"

	"github.com/onexdemo/iamv2/internal/apiserver/model"
)

// ProviderSet binds the abstract store interface to its gorm implementation.
var ProviderSet = wire.NewSet(NewStore, wire.Bind(new(IStore), new(*store)))

// IStore defines the persistence operations exposed to the business layer.
type IStore interface {
	// DB returns the underlying *gorm.DB instance with optional filters applied.
	DB(ctx context.Context, wheres ...where.Where) *gorm.DB
	// TX executes the given function inside a transaction.
	TX(ctx context.Context, fn func(ctx context.Context) error) error

	User() UserStore
	Secret() SecretStore
	Policy() PolicyStore
	PolicyAudit() PolicyAuditStore
}

// txKey is the context key used to carry an in-flight transaction.
type txKey struct{}

// store is the concrete gorm-backed implementation of IStore.
type store struct {
	db *gorm.DB
}

var _ IStore = (*store)(nil)

// NewStore creates a store backed by the given database handle.
func NewStore(db *gorm.DB) *store {
	return &store{db: db}
}

// DB returns the database handle, preferring an in-flight transaction when one
// is present in the context.
func (s *store) DB(ctx context.Context, wheres ...where.Where) *gorm.DB {
	db := s.db.WithContext(ctx)
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		db = tx.WithContext(ctx)
	}

	for _, w := range wheres {
		db = w.Where(db)
	}

	return db
}

// TX runs fn inside a database transaction.
func (s *store) TX(ctx context.Context, fn func(ctx context.Context) error) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, txKey{}, tx))
	})
}

// User returns the user store.
func (s *store) User() UserStore {
	return newUserStore(s)
}

// Secret returns the secret store.
func (s *store) Secret() SecretStore {
	return newSecretStore(s)
}

// Policy returns the policy store.
func (s *store) Policy() PolicyStore {
	return newPolicyStore(s)
}

// PolicyAudit returns the policy audit store.
func (s *store) PolicyAudit() PolicyAuditStore {
	return newPolicyAuditStore(s)
}

// Ensure model types stay referenced from this package so that registry
// migration is triggered for consumers that import only the store layer.
var _ = []any{
	&model.UserM{},
	&model.SecretM{},
	&model.PolicyM{},
	&model.PolicyAuditM{},
}
