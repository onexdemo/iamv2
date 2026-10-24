// Package cache exposes the read-only snapshot endpoints used by the authz
// server to warm its in-memory policy and secret cache.
package cache

import (
	"context"

	bizutil "github.com/onexdemo/iamv2/internal/apiserver/biz/v1/bizutil"
	"github.com/onexdemo/iamv2/internal/apiserver/pkg/conversion"
	"github.com/onexdemo/iamv2/internal/apiserver/store"
	"github.com/onexdemo/iamv2/internal/pkg/errno"
	v1 "github.com/onexdemo/iamv2/pkg/api/apiserver/v1"
)

// CacheBiz defines the cache snapshot operations.
type CacheBiz interface {
	ListSecrets(ctx context.Context, rq *v1.ListSecretInfoRequest) (*v1.ListSecretInfoResponse, error)
	ListPolicies(ctx context.Context, rq *v1.ListPolicyInfoRequest) (*v1.ListPolicyInfoResponse, error)
}

type cacheBiz struct {
	store store.IStore
}

var _ CacheBiz = (*cacheBiz)(nil)

// New creates a cache business module.
func New(store store.IStore) *cacheBiz {
	return &cacheBiz{store: store}
}

// ListSecrets returns a snapshot of every access key.
func (b *cacheBiz) ListSecrets(
	ctx context.Context,
	rq *v1.ListSecretInfoRequest,
) (*v1.ListSecretInfoResponse, error) {

	total, secrets, err := b.store.Secret().List(ctx, bizutil.ListOptions("", rq.Offset, rq.Limit))
	if err != nil {
		return nil, errno.ErrDBRead.WithMessage("%s", err.Error())
	}

	items := make([]*v1.SecretInfo, 0, len(secrets))
	for _, secretM := range secrets {
		items = append(items, conversion.SecretMToSecretInfo(secretM))
	}

	return &v1.ListSecretInfoResponse{TotalCount: total, Items: items}, nil
}

// ListPolicies returns a snapshot of every authorization policy.
func (b *cacheBiz) ListPolicies(
	ctx context.Context,
	rq *v1.ListPolicyInfoRequest,
) (*v1.ListPolicyInfoResponse, error) {

	total, policies, err := b.store.Policy().List(ctx, bizutil.ListOptions("", rq.Offset, rq.Limit))
	if err != nil {
		return nil, errno.ErrDBRead.WithMessage("%s", err.Error())
	}

	items := make([]*v1.PolicyInfo, 0, len(policies))
	for _, policyM := range policies {
		items = append(items, conversion.PolicyMToPolicyInfo(policyM))
	}

	return &v1.ListPolicyInfoResponse{TotalCount: total, Items: items}, nil
}
