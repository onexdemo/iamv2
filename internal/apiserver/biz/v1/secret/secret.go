// Package secret implements the access-key business logic.
package secret

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/onexstack/onexstack/pkg/aksk"
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

// maxSecretCount is the maximum number of access keys a user may own.
const maxSecretCount = 10

// SecretBiz defines the access-key business operations.
type SecretBiz interface {
	Create(ctx context.Context, rq *v1.CreateSecretRequest) (*v1.CreateSecretResponse, error)
	Update(ctx context.Context, rq *v1.UpdateSecretRequest) (*v1.UpdateSecretResponse, error)
	Delete(ctx context.Context, rq *v1.DeleteSecretRequest) (*v1.DeleteSecretResponse, error)
	DeleteCollection(ctx context.Context, rq *v1.DeleteSecretsRequest) (*v1.DeleteSecretsResponse, error)
	Get(ctx context.Context, rq *v1.GetSecretRequest) (*v1.GetSecretResponse, error)
	List(ctx context.Context, rq *v1.ListSecretsRequest) (*v1.ListSecretsResponse, error)

	// ListAllSecrets returns every secret for the authz server cache.
	ListAllSecrets(ctx context.Context, rq *v1.ListSecretInfoRequest) (*v1.ListSecretInfoResponse, error)
	// GetBySecretID returns a secret by its access key identifier.
	GetBySecretID(ctx context.Context, secretID string) (*model.SecretM, error)
}

type secretBiz struct {
	store    store.IStore
	notifier notifier.Notifier
}

var _ SecretBiz = (*secretBiz)(nil)

// New creates a secret business module.
func New(store store.IStore, n notifier.Notifier) *secretBiz {
	return &secretBiz{store: store, notifier: n}
}

// Create generates a new access key pair for the authenticated user.
func (b *secretBiz) Create(ctx context.Context, rq *v1.CreateSecretRequest) (*v1.CreateSecretResponse, error) {
	username := contextx.Username(ctx)
	if username == "" {
		return nil, errno.ErrUnauthenticated
	}

	total, _, err := b.store.Secret().List(ctx, where.F("username", username))
	if err != nil {
		return nil, errno.ErrDBRead.WithMessage("%s", err.Error())
	}

	if total >= maxSecretCount {
		return nil, errno.ErrSecretReachMaxCount.WithMessage("secret count: %d", total)
	}

	secretKey, err := aksk.GenerateSecretKey()
	if err != nil {
		return nil, errno.ErrInternal.WithMessage("failed to generate secret key")
	}

	secretM := &model.SecretM{
		Name:        strings.TrimSpace(rq.Name),
		Username:    username,
		SecretKey:   secretKey,
		Expires:     rq.Expires,
		Description: rq.Description,
		Extend:      rq.Extend,
	}

	if err := b.store.Secret().Create(ctx, secretM); err != nil {
		return nil, errno.ErrDBWrite.WithMessage("%s", err.Error())
	}

	b.notify(ctx)

	return &v1.CreateSecretResponse{Secret: conversion.SecretMToSecretV1(secretM)}, nil
}

// Update changes the expiry and description of a secret.
func (b *secretBiz) Update(ctx context.Context, rq *v1.UpdateSecretRequest) (*v1.UpdateSecretResponse, error) {
	secretM, err := b.getOwned(ctx, rq.Name)
	if err != nil {
		return nil, err
	}

	secretM.Expires = rq.Expires
	secretM.Description = rq.Description
	secretM.Extend = rq.Extend

	if err := b.store.Secret().Update(ctx, secretM); err != nil {
		return nil, errno.ErrDBWrite.WithMessage("%s", err.Error())
	}

	b.notify(ctx)

	return &v1.UpdateSecretResponse{Secret: conversion.SecretMToSecretV1(secretM)}, nil
}

// Delete removes one of the caller's secrets.
func (b *secretBiz) Delete(ctx context.Context, rq *v1.DeleteSecretRequest) (*v1.DeleteSecretResponse, error) {
	username := contextx.Username(ctx)
	if username == "" {
		return nil, errno.ErrUnauthenticated
	}

	if err := b.store.Secret().Delete(ctx, where.NewWhere().F("username", username).F("name", rq.Name)); err != nil {
		return nil, errno.ErrDBWrite.WithMessage("%s", err.Error())
	}

	b.notify(ctx)

	return &v1.DeleteSecretResponse{}, nil
}

// DeleteCollection removes several of the caller's secrets.
func (b *secretBiz) DeleteCollection(
	ctx context.Context,
	rq *v1.DeleteSecretsRequest,
) (*v1.DeleteSecretsResponse, error) {

	username := contextx.Username(ctx)
	if username == "" {
		return nil, errno.ErrUnauthenticated
	}

	opts := where.NewWhere().F("username", username).Q("name IN (?)", rq.Names)
	if err := b.store.Secret().Delete(ctx, opts); err != nil {
		return nil, errno.ErrDBWrite.WithMessage("%s", err.Error())
	}

	b.notify(ctx)

	return &v1.DeleteSecretsResponse{}, nil
}

// Get returns one of the caller's secrets.
func (b *secretBiz) Get(ctx context.Context, rq *v1.GetSecretRequest) (*v1.GetSecretResponse, error) {
	secretM, err := b.getOwned(ctx, rq.Name)
	if err != nil {
		return nil, err
	}

	return &v1.GetSecretResponse{Secret: conversion.SecretMToSecretV1(secretM)}, nil
}

// List returns the caller's secrets.
func (b *secretBiz) List(ctx context.Context, rq *v1.ListSecretsRequest) (*v1.ListSecretsResponse, error) {
	username := contextx.Username(ctx)
	if username == "" {
		return nil, errno.ErrUnauthenticated
	}

	opts := bizutil.ListOptions(rq.Name, rq.Offset, rq.Limit)
	opts.F("username", username)

	total, secrets, err := b.store.Secret().List(ctx, opts)
	if err != nil {
		return nil, errno.ErrDBRead.WithMessage("%s", err.Error())
	}

	items := conversion.SecretMListToSecretV1List(secrets)
	for _, item := range items {
		// Listing never exposes the secret material; callers must fetch an
		// individual secret instead.
		item.SecretKey = ""
	}

	return &v1.ListSecretsResponse{
		TotalCount: total,
		Items:      items,
	}, nil
}

// ListAllSecrets returns every secret for the authz server cache.
func (b *secretBiz) ListAllSecrets(
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

// GetBySecretID returns a secret by its access key identifier.
func (b *secretBiz) GetBySecretID(ctx context.Context, secretID string) (*model.SecretM, error) {
	secretM, err := b.store.Secret().Get(ctx, where.F("secret_id", secretID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.ErrSecretNotFound
		}

		return nil, errno.ErrDBRead.WithMessage("%s", err.Error())
	}

	return secretM, nil
}

func (b *secretBiz) getOwned(ctx context.Context, name string) (*model.SecretM, error) {
	username := contextx.Username(ctx)
	if username == "" {
		return nil, errno.ErrUnauthenticated
	}

	secretM, err := b.store.Secret().Get(ctx, where.NewWhere().F("username", username).F("name", name))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.ErrSecretNotFound
		}

		return nil, errno.ErrDBRead.WithMessage("%s", err.Error())
	}

	return secretM, nil
}

func (b *secretBiz) notify(ctx context.Context) {
	if err := b.notifier.Notify(ctx, notifier.SecretChanged); err != nil {
		// A failed notification only delays cache convergence; log it and keep
		// the mutation successful.
		slog.WarnContext(ctx, "failed to publish secret change notification", "error", err)
	}
}
