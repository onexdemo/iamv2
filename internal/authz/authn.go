package authz

import (
	"context"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"

	"github.com/onexdemo/iamv2/internal/authz/pkg/cache"
	"github.com/onexdemo/iamv2/internal/pkg/errno"
	mw "github.com/onexdemo/iamv2/internal/pkg/middleware/gin"
)

// authenticator implements secret-based authentication for the authz server.
//
// Clients either present a JWT signed with an access-key secret (the key ID is
// carried in the JWT "kid" header) or HTTP Basic credentials in the form
// "secretID:secretKey".
type authenticator struct {
	cache *cache.Cache
}

var _ mw.Authenticator = (*authenticator)(nil)

// NewAuthenticator creates an authenticator backed by the secret cache.
func NewAuthenticator(c *cache.Cache) mw.Authenticator {
	return &authenticator{cache: c}
}

// Authenticate validates a Bearer token or HTTP Basic credential.
func (a *authenticator) Authenticate(
	_ context.Context,
	scheme, credential string,
) (*mw.Identity, error) {

	switch {
	case strings.EqualFold(scheme, "Bearer"):
		return a.authenticateToken(credential)
	case strings.EqualFold(scheme, "Basic"):
		return a.authenticateBasic(credential)
	default:
		return nil, errno.ErrInvalidAuthHeader.WithMessage("unsupported authorization scheme %q", scheme)
	}
}

func (a *authenticator) authenticateToken(tokenString string) (*mw.Identity, error) {
	var secretUsername string
	var secretExpires int64

	parsed, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		kid, ok := token.Header["kid"].(string)
		if !ok || kid == "" {
			return nil, errno.ErrSignatureInvalid.WithMessage("the token does not carry a key id")
		}

		secret, err := a.cache.GetSecret(kid)
		if err != nil {
			return nil, errno.ErrSignatureInvalid.WithMessage("the access key is unknown")
		}

		secretUsername = secret.Username
		secretExpires = secret.Expires

		return []byte(secret.SecretKey), nil
	})
	if err != nil || !parsed.Valid {
		return nil, errno.ErrSignatureInvalid.WithMessage("%v", err)
	}

	if keyExpired(secretExpires) {
		return nil, errno.ErrExpired.WithMessage("the access key expired at %s", time.Unix(secretExpires, 0).Format(time.RFC3339))
	}

	return &mw.Identity{Username: secretUsername}, nil
}

func (a *authenticator) authenticateBasic(credential string) (*mw.Identity, error) {
	payload, err := base64.StdEncoding.DecodeString(credential)
	if err != nil {
		return nil, errno.ErrInvalidAuthHeader.WithMessage("failed to decode basic credentials")
	}

	secretID, secretKey, ok := strings.Cut(string(payload), ":")
	if !ok || secretID == "" {
		return nil, errno.ErrInvalidAuthHeader.WithMessage("malformed basic credentials")
	}

	secret, err := a.cache.GetSecret(secretID)
	if err != nil {
		return nil, errno.ErrSignatureInvalid.WithMessage("the access key is unknown")
	}

	if subtle.ConstantTimeCompare([]byte(secret.SecretKey), []byte(secretKey)) != 1 {
		return nil, errno.ErrSignatureInvalid.WithMessage("the access key secret is invalid")
	}

	if keyExpired(secret.Expires) {
		return nil, errno.ErrExpired.WithMessage("the access key expired at %s", time.Unix(secret.Expires, 0).Format(time.RFC3339))
	}

	return &mw.Identity{Username: secret.Username}, nil
}

// keyExpired reports whether an access key has expired. A non-positive expiry
// means the key never expires.
func keyExpired(expires int64) bool {
	return expires > 0 && time.Now().After(time.Unix(expires, 0))
}
