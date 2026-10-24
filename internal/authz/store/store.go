// Package store defines the read-only data source used by the authz server.
package store

import (
	"context"

	"github.com/ory/ladon"

	v1 "github.com/onexdemo/iamv2/pkg/api/apiserver/v1"
)

// Factory provides the stores needed by the authz server cache.
type Factory interface {
	Secrets() SecretStore
	Policies() PolicyStore
}

// SecretStore lists every access key managed by the apiserver.
type SecretStore interface {
	List(ctx context.Context) (map[string]*v1.SecretInfo, error)
}

// PolicyStore lists every authorization policy managed by the apiserver.
type PolicyStore interface {
	List(ctx context.Context) (map[string][]*ladon.DefaultPolicy, error)
}
