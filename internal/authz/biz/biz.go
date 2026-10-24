// Package biz implements the authorization business logic.
package biz

import (
	"context"

	"github.com/google/wire"
	"github.com/ory/ladon"

	"github.com/onexdemo/iamv2/internal/authz/pkg/authorization"
	authzv1 "github.com/onexdemo/iamv2/pkg/api/authz/v1"
)

// ProviderSet declares the Wire providers of the business layer.
var ProviderSet = wire.NewSet(New, wire.Bind(new(IBiz), new(*biz)))

// IBiz exposes the authorization business operation.
type IBiz interface {
	Authorize(ctx context.Context, request *ladon.Request) (*authzv1.AuthzResponse, error)
}

type biz struct {
	authorizer *authorization.Authorizer
}

var _ IBiz = (*biz)(nil)

// New creates the authorization business module.
func New(authorizer *authorization.Authorizer) *biz {
	return &biz{authorizer: authorizer}
}

// Authorize evaluates a subject access review.
func (b *biz) Authorize(_ context.Context, request *ladon.Request) (*authzv1.AuthzResponse, error) {
	return b.authorizer.Authorize(request), nil
}
