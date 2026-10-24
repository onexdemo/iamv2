package biz

import (
	"github.com/google/wire"

	cachev1 "github.com/onexdemo/iamv2/internal/apiserver/biz/v1/cache"
	policyv1 "github.com/onexdemo/iamv2/internal/apiserver/biz/v1/policy"
	secretv1 "github.com/onexdemo/iamv2/internal/apiserver/biz/v1/secret"
	userv1 "github.com/onexdemo/iamv2/internal/apiserver/biz/v1/user"
	"github.com/onexdemo/iamv2/internal/apiserver/store"
	"github.com/onexdemo/iamv2/internal/pkg/notifier"
)

// ProviderSet declares the Wire providers of the business layer.
var ProviderSet = wire.NewSet(NewBiz, wire.Bind(new(IBiz), new(*biz)))

// IBiz exposes the business modules of the apiserver.
type IBiz interface {
	Store() store.IStore
	UserV1() userv1.UserBiz
	SecretV1() secretv1.SecretBiz
	PolicyV1() policyv1.PolicyBiz
	CacheV1() cachev1.CacheBiz
}

type biz struct {
	store    store.IStore
	notifier notifier.Notifier
}

var _ IBiz = (*biz)(nil)

// NewBiz creates the business layer facade.
func NewBiz(store store.IStore, notifier notifier.Notifier) *biz {
	return &biz{store: store, notifier: notifier}
}

// Store returns the shared store layer.
func (b *biz) Store() store.IStore {
	return b.store
}

// UserV1 returns the user business module.
func (b *biz) UserV1() userv1.UserBiz {
	return userv1.New(b.store, b.notifier)
}

// SecretV1 returns the secret business module.
func (b *biz) SecretV1() secretv1.SecretBiz {
	return secretv1.New(b.store, b.notifier)
}

// PolicyV1 returns the policy business module.
func (b *biz) PolicyV1() policyv1.PolicyBiz {
	return policyv1.New(b.store, b.notifier)
}

// CacheV1 returns the cache business module.
func (b *biz) CacheV1() cachev1.CacheBiz {
	return cachev1.New(b.store)
}
