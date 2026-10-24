// Package bizutil contains helpers shared by the apiserver business modules.
package bizutil

import (
	"context"

	"github.com/onexstack/onexstack/pkg/store/where"

	"github.com/onexdemo/iamv2/internal/pkg/contextx"
	"github.com/onexdemo/iamv2/internal/pkg/errno"
)

// maxListLimit is used when the caller asks for every record. The generic store
// maps non-positive limits to its small default, so an explicit large bound is
// the portable way to express "all".
const maxListLimit = 1 << 30

// ListOptions builds the where options shared by the resource list endpoints.
func ListOptions(name string, offset, limit int64) *where.Options {
	opts := where.NewWhere()

	if name != "" {
		opts.Q("name LIKE ?", "%"+name+"%")
	}

	if offset > 0 {
		opts.O(int(offset))
	}

	if limit > 0 {
		opts.L(int(limit))
	} else {
		opts.L(maxListLimit)
	}

	return opts
}

// AuthorizeSelf rejects the request unless the caller is the target user or an
// administrator.
func AuthorizeSelf(ctx context.Context, username string) error {
	if contextx.IsAdmin(ctx) || contextx.Username(ctx) == username {
		return nil
	}

	return errno.ErrPermissionDenied
}
