package gin

import (
	"context"
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/onexstack/onexstack/pkg/core"

	"github.com/onexdemo/iamv2/internal/pkg/contextx"
	"github.com/onexdemo/iamv2/internal/pkg/errno"
)

// UsernameKey is the gin.Context key that stores the authenticated username.
// It is kept for compatibility with handlers that predate contextx.
const UsernameKey = "username"

// Identity is the authenticated caller attached to a request.
type Identity struct {
	// UserID is the stable identifier of the caller.
	UserID string
	// Username is the login name of the caller.
	Username string
	// IsAdmin reports whether the caller is an administrator.
	IsAdmin bool
}

// Authenticator validates an Authorization header and returns the caller
// identity. The scheme is either "Basic" or "Bearer".
type Authenticator interface {
	Authenticate(ctx context.Context, scheme, credential string) (*Identity, error)
}

// Authn authenticates the request. It accepts both HTTP Basic and Bearer
// credentials so that interactive clients and machine clients share the same
// routes.
func Authn(authenticator Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.Request.Header.Get("Authorization")
		if header == "" {
			// Fall back to the cookie used by the legacy iam web console.
			if cookie, err := c.Cookie("jwt"); err == nil && cookie != "" {
				header = "Bearer " + cookie
			}
		}

		if header == "" {
			core.WriteResponse(c, nil, errno.ErrMissingHeader.WithMessage("Authorization header cannot be empty."))
			c.Abort()

			return
		}

		scheme, credential, ok := strings.Cut(header, " ")
		if !ok || credential == "" {
			core.WriteResponse(c, nil, errno.ErrInvalidAuthHeader)
			c.Abort()

			return
		}

		identity, err := authenticator.Authenticate(c.Request.Context(), scheme, credential)
		if err != nil {
			slog.WarnContext(c.Request.Context(), "authentication failed", "scheme", scheme, "error", err)
			core.WriteResponse(c, nil, err)
			c.Abort()

			return
		}

		ctx := contextx.WithUserID(c.Request.Context(), identity.UserID)
		ctx = contextx.WithUsername(ctx, identity.Username)
		ctx = contextx.WithIsAdmin(ctx, identity.IsAdmin)
		c.Request = c.Request.WithContext(ctx)
		c.Set(UsernameKey, identity.Username)

		c.Next()
	}
}
