// Package policy implements the ladon authorization-policy business logic.
package policy

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/onexstack/onexstack/pkg/store/where"
	"gorm.io/gorm"

	bizutil "github.com/onexdemo/iamv2/internal/apiserver/biz/v1/bizutil"
	model "github.com/onexdemo/iamv2/internal/apiserver/model"
	"github.com/onexdemo/iamv2/internal/apiserver/pkg/conversion"
	"github.com/onexdemo/iamv2/internal/apiserver/store"
	"github.com/onexdemo/iamv2/internal/pkg/contextx"
	"github.com/onexdemo/iamv2/internal/pkg/errno"
	"github.com/onexdemo/iamv2/internal/pkg/notifier"
	v1 "github.com/onexdemo/iamv2/pkg/api/apiserver/v1"
)

// PolicyBiz defines the policy business operations.
type PolicyBiz interface {
	Create(ctx context.Context, rq *v1.CreatePolicyRequest) (*v1.CreatePolicyResponse, error)
	Update(ctx context.Context, rq *v1.UpdatePolicyRequest) (*v1.UpdatePolicyResponse, error)
	Delete(ctx context.Context, rq *v1.DeletePolicyRequest) (*v1.DeletePolicyResponse, error)
	DeleteCollection(ctx context.Context, rq *v1.DeletePoliciesRequest) (*v1.DeletePoliciesResponse, error)
	Get(ctx context.Context, rq *v1.GetPolicyRequest) (*v1.GetPolicyResponse, error)
	List(ctx context.Context, rq *v1.ListPoliciesRequest) (*v1.ListPoliciesResponse, error)

	// ListAllPolicies returns every policy for the authz server cache.
	ListAllPolicies(ctx context.Context, rq *v1.ListPolicyInfoRequest) (*v1.ListPolicyInfoResponse, error)
}

type policyBiz struct {
	store    store.IStore
	notifier notifier.Notifier
}

var _ PolicyBiz = (*policyBiz)(nil)

// New creates a policy business module.
func New(store store.IStore, n notifier.Notifier) *policyBiz {
	return &policyBiz{store: store, notifier: n}
}

// Create stores a new ladon policy for the authenticated user.
func (b *policyBiz) Create(ctx context.Context, rq *v1.CreatePolicyRequest) (*v1.CreatePolicyResponse, error) {
	username := contextx.Username(ctx)
	if username == "" {
		return nil, errno.ErrUnauthenticated
	}

	name := strings.TrimSpace(rq.Name)
	if _, err := b.store.Policy().Get(ctx, where.NewWhere().F("username", username).F("name", name)); err == nil {
		return nil, errno.ErrPolicyAlreadyExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errno.ErrDBRead.WithMessage("%s", err.Error())
	}

	policyM := &model.PolicyM{
		Name:     name,
		Username: username,
		Policy:   rq.Policy,
		Extend:   rq.Extend,
	}

	if err := b.store.Policy().Create(ctx, policyM); err != nil {
		return nil, errno.ErrDBWrite.WithMessage("%s", err.Error())
	}

	b.notify(ctx)

	return &v1.CreatePolicyResponse{Policy: conversion.PolicyMToPolicyV1(policyM)}, nil
}

// Update replaces the ladon rules of an existing policy.
func (b *policyBiz) Update(ctx context.Context, rq *v1.UpdatePolicyRequest) (*v1.UpdatePolicyResponse, error) {
	policyM, err := b.getOwned(ctx, rq.Name)
	if err != nil {
		return nil, err
	}

	policyM.Policy = rq.Policy
	policyM.Extend = rq.Extend

	if err := b.store.Policy().Update(ctx, policyM); err != nil {
		return nil, errno.ErrDBWrite.WithMessage("%s", err.Error())
	}

	b.notify(ctx)

	return &v1.UpdatePolicyResponse{Policy: conversion.PolicyMToPolicyV1(policyM)}, nil
}

// Delete removes one of the caller's policies.
func (b *policyBiz) Delete(ctx context.Context, rq *v1.DeletePolicyRequest) (*v1.DeletePolicyResponse, error) {
	username := contextx.Username(ctx)
	if username == "" {
		return nil, errno.ErrUnauthenticated
	}

	if err := b.store.TX(ctx, func(txCtx context.Context) error {
		if policyM, err := b.store.Policy().Get(txCtx, where.NewWhere().F("username", username).F("name", rq.Name)); err == nil {
			if err := b.store.PolicyAudit().CreateBatch(txCtx, []*model.PolicyAuditM{policyM.ToAudit()}); err != nil {
				return err
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		return b.store.Policy().Delete(txCtx, where.NewWhere().F("username", username).F("name", rq.Name))
	}); err != nil {
		return nil, errno.ErrDBWrite.WithMessage("%s", err.Error())
	}

	b.notify(ctx)

	return &v1.DeletePolicyResponse{}, nil
}

// DeleteCollection removes several of the caller's policies.
func (b *policyBiz) DeleteCollection(
	ctx context.Context,
	rq *v1.DeletePoliciesRequest,
) (*v1.DeletePoliciesResponse, error) {

	username := contextx.Username(ctx)
	if username == "" {
		return nil, errno.ErrUnauthenticated
	}

	opts := where.NewWhere().F("username", username).Q("name IN (?)", rq.Names)
	if err := b.store.TX(ctx, func(txCtx context.Context) error {
		_, policies, err := b.store.Policy().List(txCtx, opts)
		if err != nil {
			return err
		}

		audits := make([]*model.PolicyAuditM, 0, len(policies))
		for _, policyM := range policies {
			audits = append(audits, policyM.ToAudit())
		}

		if err := b.store.PolicyAudit().CreateBatch(txCtx, audits); err != nil {
			return err
		}

		return b.store.Policy().Delete(txCtx, where.NewWhere().F("username", username).Q("name IN (?)", rq.Names))
	}); err != nil {
		return nil, errno.ErrDBWrite.WithMessage("%s", err.Error())
	}

	b.notify(ctx)

	return &v1.DeletePoliciesResponse{}, nil
}

// Get returns one of the caller's policies.
func (b *policyBiz) Get(ctx context.Context, rq *v1.GetPolicyRequest) (*v1.GetPolicyResponse, error) {
	policyM, err := b.getOwned(ctx, rq.Name)
	if err != nil {
		return nil, err
	}

	return &v1.GetPolicyResponse{Policy: conversion.PolicyMToPolicyV1(policyM)}, nil
}

// List returns the caller's policies.
func (b *policyBiz) List(ctx context.Context, rq *v1.ListPoliciesRequest) (*v1.ListPoliciesResponse, error) {
	username := contextx.Username(ctx)
	if username == "" {
		return nil, errno.ErrUnauthenticated
	}

	opts := bizutil.ListOptions(rq.Name, rq.Offset, rq.Limit)
	opts.F("username", username)

	total, policies, err := b.store.Policy().List(ctx, opts)
	if err != nil {
		return nil, errno.ErrDBRead.WithMessage("%s", err.Error())
	}

	return &v1.ListPoliciesResponse{
		TotalCount: total,
		Items:      conversion.PolicyMListToPolicyV1List(policies),
	}, nil
}

// ListAllPolicies returns every policy for the authz server cache.
func (b *policyBiz) ListAllPolicies(
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

func (b *policyBiz) getOwned(ctx context.Context, name string) (*model.PolicyM, error) {
	username := contextx.Username(ctx)
	if username == "" {
		return nil, errno.ErrUnauthenticated
	}

	policyM, err := b.store.Policy().Get(ctx, where.NewWhere().F("username", username).F("name", name))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.ErrPolicyNotFound
		}

		return nil, errno.ErrDBRead.WithMessage("%s", err.Error())
	}

	return policyM, nil
}

func (b *policyBiz) notify(ctx context.Context) {
	if err := b.notifier.Notify(ctx, notifier.PolicyChanged); err != nil {
		slog.WarnContext(ctx, "failed to publish policy change notification", "error", err)
	}
}
