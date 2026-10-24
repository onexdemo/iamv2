package apiserver

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/onexstack/onexstack/pkg/token"

	"github.com/onexdemo/iamv2/internal/apiserver/biz"
	"github.com/onexdemo/iamv2/internal/pkg/errno"
	mw "github.com/onexdemo/iamv2/internal/pkg/middleware/gin"
)

// identityClaim is the JWT claim that carries the user ID.
const identityClaim = "userID"

// authenticator implements mw.Authenticator for the apiserver.
type authenticator struct {
	biz biz.IBiz
}

var _ mw.Authenticator = (*authenticator)(nil)

// NewAuthenticator creates an authenticator backed by the user business layer.
func NewAuthenticator(biz biz.IBiz) mw.Authenticator {
	return &authenticator{biz: biz}
}

// Authenticate validates a Bearer token or HTTP Basic credential.
func (a *authenticator) Authenticate(
	ctx context.Context,
	scheme, credential string,
) (*mw.Identity, error) {

	switch {
	case strings.EqualFold(scheme, "Bearer"):
		return a.authenticateToken(ctx, credential)
	case strings.EqualFold(scheme, "Basic"):
		return a.authenticateBasic(ctx, credential)
	default:
		return nil, errno.ErrInvalidAuthHeader.WithMessage("unsupported authorization scheme %q", scheme)
	}
}

func (a *authenticator) authenticateToken(ctx context.Context, tokenString string) (*mw.Identity, error) {
	if err := token.Parse(tokenString); err != nil {
		return nil, errno.ErrTokenInvalid.WithMessage("%s", err.Error())
	}

	claims, err := token.GetClaims(tokenString)
	if err != nil {
		return nil, errno.ErrTokenInvalid.WithMessage("%s", err.Error())
	}

	userID, ok := claims[identityClaim].(string)
	if !ok || userID == "" {
		return nil, errno.ErrTokenInvalid.WithMessage("token does not carry a user identity")
	}

	userM, err := a.biz.UserV1().GetByUserID(ctx, userID)
	if err != nil {
		return nil, errno.ErrUnauthenticated.WithMessage("%s", err.Error())
	}

	if userM.Status != 1 {
		return nil, errno.ErrUnauthenticated.WithMessage("the account has been disabled")
	}

	return &mw.Identity{
		UserID:   userM.UserID,
		Username: userM.Username,
		IsAdmin:  userM.IsAdmin == 1,
	}, nil
}

func (a *authenticator) authenticateBasic(ctx context.Context, credential string) (*mw.Identity, error) {
	payload, err := base64.StdEncoding.DecodeString(credential)
	if err != nil {
		return nil, errno.ErrInvalidAuthHeader.WithMessage("failed to decode basic credentials")
	}

	username, password, ok := strings.Cut(string(payload), ":")
	if !ok {
		return nil, errno.ErrInvalidAuthHeader.WithMessage("malformed basic credentials")
	}

	userM, err := a.biz.UserV1().Authenticate(ctx, username, password)
	if err != nil {
		return nil, err
	}

	return &mw.Identity{
		UserID:   userM.UserID,
		Username: userM.Username,
		IsAdmin:  userM.IsAdmin == 1,
	}, nil
}
